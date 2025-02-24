package servers

import (
    "net/http"
    "fmt"
    "Groxy/tls"
    "time"
    "strings"
)

type Server struct {
    handler     http.Handler
    tlsManager  *tls.Manager
    certFile    string
    keyFile     string
    httpPort    string
    httpsPort   string
    enableRedirection bool
}

func NewServer(handler http.Handler, tlsManager *tls.Manager, certFile, keyFile string, httpPort, httpsPort string) *Server {
    return &Server{
        handler:    handler,
        tlsManager: tlsManager,
        certFile:   certFile,
        keyFile:    keyFile,
        httpPort:   httpPort,
        httpsPort:  httpsPort,
        enableRedirection: true, // Enable redirection by default
    }
}

func (s *Server) SetRedirection(enable bool) {
    s.enableRedirection = enable
}

func (s *Server) StartHTTP() error {
    addr := ":" + s.httpPort
    
    var serverHandler http.Handler
    if s.enableRedirection {
        serverHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            host := r.Host
            if hostParts := strings.Split(host, ":"); len(hostParts) > 0 {
                host = hostParts[0]
            }
            
            httpsURL := fmt.Sprintf("https://%s:%s%s", host, s.httpsPort, r.URL.Path)
            if r.URL.RawQuery != "" {
                httpsURL += "?" + r.URL.RawQuery
            }
            
            http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
        })
    } else {
        serverHandler = s.handler
    }
    
    server := &http.Server{
        Addr:    addr,
        Handler: serverHandler,
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

