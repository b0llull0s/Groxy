package main

import (
    "net/http"
    "net/http/httputil"
)

func modifyRequest(proxy *httputil.ReverseProxy) {
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        logRequest(req) // Log the request
        req.Header.Add("X-Custom-Header", "MyProxy") // Add a custom header
    }
}