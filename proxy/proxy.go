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

// createReverseProxy creates a configured reverse proxy
func (p *Proxy) createReverseProxy(targetURL *url.URL) *httputil.ReverseProxy {
    proxy := httputil.NewSingleHostReverseProxy(targetURL)
    
    // Configure TLS if needed
    if targetURL.Scheme == "https" {
        proxy.Transport = &http.Transport{
            TLSClientConfig: p.tlsConfig.LoadClientConfig(),
        }
    }
    
    ModifyRequest(proxy, p.customHeader)
    ModifyResponse(proxy)
    return proxy
}

func (p *Proxy) Handler() http.Handler {
    // Transparent mode
    if p.targetURL == nil {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            p.handleTransparentProxy(w, r)
        })
    }
    
    // Target-specific mode
    return p.createReverseProxy(p.targetURL)
}

// handleTransparentProxy handles transparent proxy mode requests
func (p *Proxy) handleTransparentProxy(w http.ResponseWriter, r *http.Request) {
    if r.Host == "localhost:8080" || r.Host == "localhost:8443" {
        http.Error(w, "Cannot proxy to self", http.StatusBadRequest)
        return
    }

    destinationHost := r.Host
    if destinationHost == "" {
        logger.LogTransparentProxyHandlerUnableToDetermineDestinationHost(w)
        return
    }

    scheme := "http"
    if r.TLS != nil {
        scheme = "https"
    }

    destinationURL := &url.URL{
        Scheme: scheme,
        Host:   destinationHost,
        Path:   r.URL.Path,
    }

    proxy := p.createReverseProxy(destinationURL)
    proxy.ServeHTTP(w, r)
}
