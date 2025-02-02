package servers

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

// StartHTTPSServer starts a HTTPS server on the given address.
func StartHTTPSServer(addr, certFile, keyFile string, handler http.Handler) error {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
			return fmt.Errorf("Failed to load server certificate: %v", err)
	}

	// Configure the TLS server
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	// Start the HTTPS server
	log.Printf("Starting HTTPS server on %s\n", addr)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		return fmt.Errorf("failed to start HTTPS server: %v", err)
	}

	// Return nil to indicate success
	return nil
}
