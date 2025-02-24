package servers

import (
    "context"
    "net/http"
    "fmt"
    "Groxy/tls"
    "Groxy/logger"
    "time"
    "strings"
    "sync"
)

type Server struct {
    handler     http.Handler
    tlsManager  *tls.Manager
    certFile    string
    keyFile     string
    httpPort    string
    httpsPort   string
    enableRedirection bool
    wg          sync.WaitGroup
    httpServer  *http.Server
    httpsServer *http.Server
}

func NewServer(handler http.Handler, tlsManager *tls.Manager, certFile, keyFile string, httpPort, httpsPort string) *Server {
    return &Server{
        handler:    handler,
        tlsManager: tlsManager,
        certFile:   certFile,
        keyFile:    keyFile,
        httpPort:   httpPort,
        httpsPort:  httpsPort,
        enableRedirection: true, // Change to false to disable redirection
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
    
    s.httpServer = &http.Server{
        Addr:    addr,
        Handler: serverHandler,
    }
    
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        logger.LogHTTPServerStart(s.httpPort)
        
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.LogServerError(err)
        }
    }()
    
    return nil
}

func (s *Server) StartHTTPS() error {
    tlsConfig, err := s.tlsManager.LoadServerConfig()
    if err != nil {
        return fmt.Errorf("failed to load TLS config: %v", err)
    }

    s.tlsManager.StartRotation(30 * 24 * time.Hour)
    
    addr := ":" + s.httpsPort
    s.httpsServer = &http.Server{
        Addr:      addr,
        Handler:   s.handler,
        TLSConfig: tlsConfig,
    }
    
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        logger.LogHTTPSServerStart(s.httpsPort)
        
        if err := s.httpsServer.ListenAndServeTLS(s.certFile, s.keyFile); err != nil && err != http.ErrServerClosed {
            logger.LogServerError(err)
        }
    }()
    
    return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
    if ctx == nil {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
    }
    
    if s.httpServer != nil {
        if err := s.httpServer.Shutdown(ctx); err != nil {
            return err
        }
    }
    
    if s.httpsServer != nil {
        if err := s.httpsServer.Shutdown(ctx); err != nil {
            return err
        }
    }
    
    s.tlsManager.StopRotation()
    
    waitCh := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(waitCh)
    }()
    
    select {
    case <-waitCh:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (s *Server) WaitForServers() {
    s.wg.Wait()
}