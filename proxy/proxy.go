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

// CreateProxy is needed for backward compatibility with your handlers
func CreateProxy(destinationURL *url.URL, customHeader string) *httputil.ReverseProxy {
    proxy := httputil.NewSingleHostReverseProxy(destinationURL)
    ModifyRequest(proxy, customHeader)
    ModifyResponse(proxy)
    return proxy
}

func (p *Proxy) Handler() http.Handler {
    proxy := httputil.NewSingleHostReverseProxy(p.targetURL)
    
    if p.targetURL != nil && p.targetURL.Scheme == "https" {
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
	// Extract the target host from the request
	destinationHost := r.Host
	if destinationHost == "" {
		logger.LogTransparentProxyHandlerUnableToDetermineDestinationHost(w)
		return
	}

	// Construct the target URL using the request's Host header
	destinationURL := &url.URL{
		Scheme: "http",
		Host:   destinationHost,
		Path:   r.URL.Path,
	}

	// If the request is HTTPS, update the scheme
	if r.TLS != nil {
		destinationURL.Scheme = "https"
	}

	// Create a reverse proxy
	proxy := CreateProxy(destinationURL, customHeader)

	// Set up TLS configuration for HTTPS targets
	if r.URL.Scheme == "https" {
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