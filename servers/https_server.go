package servers

import (
	"Groxy/logger"
	"crypto/tls"
	"net/http"
)

// Starts an HTTPS server
func StartHTTPSServer(addr, certFile, keyFile string, handler http.Handler) error {
	logger.LogHTTPSServerStart(addr)

	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.LogCertificateError(err)
		return nil 
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
	if err := server.ListenAndServeTLS("", ""); err != nil {

		logger.LogServerError(err)
		return nil 
	}

	return nil
}


