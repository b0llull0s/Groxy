package proxy

import (
	"crypto/tls"
	"log"
)

// CreateTLSConfig creates a TLS configuration with the given certificate.
func CreateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	log.Printf("Loading certificate from %s and key from %s", certFile, keyFile)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Printf("Failed to load certificate: %v", err)
		return nil, err
	}

	log.Printf("Certificate loaded successfully: Subject: %v, Issuer: %v", cert.Leaf.Subject, cert.Leaf.Issuer)

	// Create TLS config
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Enforce TLS 1.2 or higher
	}

	return config, nil
}