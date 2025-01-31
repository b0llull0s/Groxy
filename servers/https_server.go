package servers

import (
	"crypto/tls"
	"log"
	"net/http"
)

// StartMinimalHTTPSServer starts a minimal HTTPS server on the given address.
func StartHTTPSServer(addr, certFile, keyFile string) {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	// Create TLS config
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Enforce TLS 1.2 or higher
	}

	// Create HTTP server with logging middleware
	server := &http.Server{
		Addr:      addr,
		Handler:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request: %s %s", r.Method, r.URL.Path)
			w.Write([]byte("Hello from the HTTPS server!"))
		}),
		TLSConfig: config,
	}

	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}
}