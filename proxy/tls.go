package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
)

// LoadCertificate loads the certificate and private key from files.
func LoadCertificate(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	return cert, nil
}

// CreateTLSConfig creates a TLS configuration with the given certificate.
func CreateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
    cert, err := LoadCertificate(certFile, keyFile)
    if err != nil {
        log.Printf("Failed to load certificate: %v", err)
        return nil, err
    }

    log.Printf("Certificate loaded successfully: Subject: %v, Issuer: %v", cert.Leaf.Subject, cert.Leaf.Issuer)
	// Create a certificate pool (optional, for client authentication)
	caCertPool := x509.NewCertPool()

	// Load CA certificate (optional, for client authentication)
	caCert, err := os.ReadFile("ca.crt")
	if err == nil {
		caCertPool.AppendCertsFromPEM(caCert)
	}

	// Create TLS config
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		// Require client certificates for mutual TLS (optional)
		// ClientAuth: tls.RequireAndVerifyClientCert,
	}

	return config, nil
}

// StartHTTPSServer starts an HTTPS server with the given handler and TLS config.
func StartHTTPSServer(addr string, handler http.Handler, config *tls.Config) {
	server := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: config,
	}

	log.Printf("Starting HTTPS server on %s\n", addr)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}
}