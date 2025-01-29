package proxy

import (
    "net/http"
    "net/http/httputil"
    "net/url"
)

// Creates a reverse proxy for the given target URL.
func CreateProxy(targetURL *url.URL) *httputil.ReverseProxy {
    proxy := httputil.NewSingleHostReverseProxy(targetURL)
    ModifyRequest(proxy)  
    ModifyResponse(proxy) 
    return proxy
}

// Handles requests in transparent mode.
func TransparentProxyHandler(w http.ResponseWriter, r *http.Request) {
    targetHost := r.Host
    if targetHost == "" {
        http.Error(w, "Unable to determine target host", http.StatusBadRequest)
        return
    }

    targetURL, err := url.Parse("http://" + targetHost)
    if err != nil {
        http.Error(w, "Invalid target host", http.StatusBadRequest)
        return
    }

    proxy := CreateProxy(targetURL)
    proxy.ServeHTTP(w, r)
}

// Handles requests in target-specific mode.
func TargetSpecificProxyHandler(targetURL *url.URL) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        proxy := CreateProxy(targetURL)
        proxy.ServeHTTP(w, r)
    }
}