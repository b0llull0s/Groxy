package servers

import (
	"Groxy/logger"
	"crypto/tls"
	"net/http"
)

const HTTPSPort = "8443"

// Starts an HTTPS server
func StartHTTPSServer(certFile, keyFile string, handler http.Handler) error {
		addr := ":" + HTTPSPort
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.LogCertificateError(err)
		return err 
	}

	// Configure the TLS server
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	logger.LogHTTPSServerStart(HTTPSPort)


	// Start the HTTPS server
	if err := server.ListenAndServeTLS("", ""); err != nil {

		logger.LogServerError(err)
		return err 
	}

	return nil
}


