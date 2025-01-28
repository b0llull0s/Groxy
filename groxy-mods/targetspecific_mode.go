package main

import (
    "net/http"
    "net/http/httputil"
    "net/url"
)

func targetSpecificProxyHandler(targetURL *url.URL) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Create a reverse proxy for the target
        proxy := httputil.NewSingleHostReverseProxy(targetURL)

        // Apply modules
        modifyRequest(proxy)
        modifyResponse(proxy)

        // Serve the request
        proxy.ServeHTTP(w, r)
    }
}