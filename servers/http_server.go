package servers

import (
	"log"
	"net/http"
)

// Starts an HTTP server on the given address.
func StartHTTPServer(addr string, proxyHandler http.Handler) {
    go func() {
        if err := http.ListenAndServe(addr, proxyHandler); err != nil {
            log.Fatalf("Failed to start HTTP proxy server: %v", err)
        }
    }()
}

