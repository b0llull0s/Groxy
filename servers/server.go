package servers

import (
    "net/http"
    "fmt"
    "Groxy/tls"
    "time"

)

type Server struct {
    handler     http.Handler
    tlsManager  *tls.Manager
    certFile    string
    keyFile     string
    httpPort    string
    httpsPort   string
}

func NewServer(handler http.Handler, tlsManager *tls.Manager, certFile, keyFile string, httpPort, httpsPort string) *Server {
    return &Server{
        handler:    handler,
        tlsManager: tlsManager,
        certFile:   certFile,
        keyFile:    keyFile,
        httpPort:   httpPort,
        httpsPort:  httpsPort,
    }
}

func (s *Server) StartHTTP() error {
    addr := ":" + s.httpPort
    server := &http.Server{
        Addr:    addr,
        Handler: s.handler,
    }
    return server.ListenAndServe()
}

func (s *Server) StartHTTPS() error {
    tlsConfig, err := s.tlsManager.LoadServerConfig()
    if err != nil {
        return fmt.Errorf("failed to load TLS config: %v", err)
    }

    // Start certificate rotation (e.g., every 30 days)
    s.tlsManager.StartRotation(30 * 24 * time.Hour)
    
    addr := ":" + s.httpsPort
    server := &http.Server{
        Addr:      addr,
        Handler:   s.handler,
        TLSConfig: tlsConfig,
    }
    
    return server.ListenAndServeTLS(s.certFile, s.keyFile)
}