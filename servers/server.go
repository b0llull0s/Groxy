package servers

import (
    "net/http"
    "Groxy/tls"
    "fmt"
)

type Server struct {
    handler    http.Handler
    tlsConfig  *tls.Config
    httpPort   string
    httpsPort  string
}

func NewServer(handler http.Handler, tlsConfig *tls.Config, httpPort, httpsPort string) *Server {
    return &Server{
        handler:   handler,
        tlsConfig: tlsConfig,
        httpPort:  httpPort,
        httpsPort: httpsPort,
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
    if s.tlsConfig == nil {
        return fmt.Errorf("TLS configuration is required for HTTPS")
    }
    
    addr := ":" + s.httpsPort
    server := &http.Server{
        Addr:      addr,
        Handler:   s.handler,
        TLSConfig: s.tlsConfig,
    }
    
    return server.ListenAndServeTLS(s.tlsConfig.CertFile, s.tlsConfig.KeyFile)
}