package main

import (
    "net/http"
    "net/http/httputil"
    "net/url"
)

func transparentProxyHandler(w http.ResponseWriter, r *http.Request) {
    // Extract the target host from the request
    targetHost := r.Host
    if targetHost == "" {
        http.Error(w, "Unable to determine target host", http.StatusBadRequest)
        return
    }

    // Parse the target URL
    targetURL, err := url.Parse("http://" + targetHost)
    if err != nil {
        http.Error(w, "Invalid target host", http.StatusBadRequest)
        return
    }

    // Create a reverse proxy for the target
    proxy := httputil.NewSingleHostReverseProxy(targetURL)

    // Apply modules
    modifyRequest(proxy)
    modifyResponse(proxy)

    // Serve the request
    proxy.ServeHTTP(w, r)
}