package proxy

import (
    "Groxy/tls"
    "net/http"
    "net/http/httputil"
    "net/url"
    "Groxy/logger"
)

type Proxy struct {
    targetURL    *url.URL
    tlsConfig    *tls.Config
    customHeader string
}

func NewProxy(targetURL *url.URL, tlsConfig *tls.Config, customHeader string) *Proxy {
    return &Proxy{
        targetURL:    targetURL,
        tlsConfig:    tlsConfig,
        customHeader: customHeader,
    }
}

// CreateProxy is needed for backward compatibility with the handlers
func CreateProxy(destinationURL *url.URL, customHeader string) *httputil.ReverseProxy {
    proxy := httputil.NewSingleHostReverseProxy(destinationURL)
    ModifyRequest(proxy, customHeader)
    ModifyResponse(proxy)
    return proxy
}

func (p *Proxy) Handler() http.Handler {
    if p.targetURL == nil {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            TransparentProxyHandler(w, r, p.customHeader)
        })
    }

    proxy := httputil.NewSingleHostReverseProxy(p.targetURL)
    
    if p.targetURL.Scheme == "https" {
        proxy.Transport = &http.Transport{
            TLSClientConfig: p.tlsConfig.LoadClientConfig(),
        }
    }
    
    ModifyRequest(proxy, p.customHeader)
    ModifyResponse(proxy)
    return proxy
}

// Transparent Mode
func TransparentProxyHandler(w http.ResponseWriter, r *http.Request, customHeader string) {
    if r.Host == "localhost:8080" || r.Host == "localhost:8443" {
        http.Error(w, "Cannot proxy to self", http.StatusBadRequest)
        return
    }

    destinationHost := r.Host
    if destinationHost == "" {
        logger.LogTransparentProxyHandlerUnableToDetermineDestinationHost(w)
        http.Error(w, "Unable to determine destination host", http.StatusBadRequest)
        return
    }

    destinationURL := &url.URL{
        Scheme: "http",
        Host:   destinationHost,
        Path:   r.URL.Path,
    }

    if r.TLS != nil {
        destinationURL.Scheme = "https"
    }

    proxy := CreateProxy(destinationURL, customHeader)

    if destinationURL.Scheme == "https" {
        tlsConfig := tls.NewConfig("", "")
        proxy.Transport = &http.Transport{
            TLSClientConfig: tlsConfig.LoadClientConfig(),
        }
    }

    proxy.ServeHTTP(w, r)
}

func TargetSpecificProxyHandler(destinationURL *url.URL, w http.ResponseWriter, r *http.Request, customHeader string) {
    proxy := CreateProxy(destinationURL, customHeader)

    if destinationURL.Scheme == "https" {
        tlsConfig := tls.NewConfig("", "")
        proxy.Transport = &http.Transport{
            TLSClientConfig: tlsConfig.LoadClientConfig(),
        }
    }

    proxy.ServeHTTP(w, r)
}