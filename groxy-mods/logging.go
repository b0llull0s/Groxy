package main

import (
    "log"
    "net/http"
)

func logRequest(req *http.Request) {
    log.Printf("Request: %s %s\n", req.Method, req.URL.String())
    for name, values := range req.Header {
        for _, value := range values {
            log.Printf("Header: %s: %s\n", name, value)
        }
    }
}

func logResponse(res *http.Response) {
    log.Printf("Response: %s\n", res.Status)
    for name, values := range res.Header {
        for _, value := range values {
            log.Printf("Header: %s: %s\n", name, value)
        }
    }
}