package proxy

import (
	"Groxy/logger"
	"Groxy/tls"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Proxy struct {
	targetURL       *url.URL
	tlsConfig       *tls.Config
	customHeader    string
	workerPool      *WorkerPool
	useWorkers      bool
	ctx             context.Context
	cancel          context.CancelFunc
	timeout         time.Duration
	obfuscator      *TrafficObfuscator
	obfuscationMode ObfuscationMode
}

func NewProxy(targetURL *url.URL, tlsConfig *tls.Config, customHeader string, obfuscationMode ObfuscationMode) *Proxy {
	ctx, cancel := context.WithCancel(context.Background())
	return &Proxy{
		targetURL:       targetURL,
		tlsConfig:       tlsConfig,
		customHeader:    customHeader,
		useWorkers:      false,
		ctx:             ctx,
		cancel:          cancel,
		timeout:         30 * time.Second, // Default timeout
		obfuscator:      NewTrafficObfuscator(obfuscationMode),
		obfuscationMode: obfuscationMode,
	}
}

func (p *Proxy) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}

func (p *Proxy) EnableWorkerPool(workerCount, queueSize int) {
	p.workerPool = NewWorkerPool(workerCount, queueSize)
	
	p.workerPool.Start()
	
	for _, worker := range p.workerPool.workers {
		worker.SetProxy(p)
	}
	
	p.useWorkers = true
	logger.LogHTTPServerStart(fmt.Sprintf("with %d workers", workerCount))
}

func (p *Proxy) StopWorkerPool() {
	if p.workerPool != nil {
		p.workerPool.Stop()
		p.useWorkers = false
	}
}

func (p *Proxy) createReverseProxy(targetURL *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	transport := &http.Transport{
		ResponseHeaderTimeout: p.timeout,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}
	
	if targetURL.Scheme == "https" {
		transport.TLSClientConfig = p.tlsConfig.LoadClientConfig()
	}
	
	proxy.Transport = transport
	
	ModifyRequest(proxy, p.customHeader, p.obfuscator)
	ModifyResponse(proxy, p.obfuscator)
	return proxy
}

func (p *Proxy) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(p.ctx, p.timeout)
		defer cancel()
		
		r = r.WithContext(ctx)
		
		if p.useWorkers {
			p.workerPool.Submit(w, r)
			return
		}
		
		doneCh := make(chan struct{})
		
		go func() {
			if p.targetURL == nil {
				p.handleTransparentProxy(w, r)
			} else {
				proxy := p.createReverseProxy(p.targetURL)
				proxy.ServeHTTP(w, r)
			}
			close(doneCh)
		}()
		
		select {
		case <-doneCh:
		case <-ctx.Done():
			logger.LogRequestTimeout(r)
			// Note: cannot write to response here as the goroutine might have already written to it
		}
	})
}

func (p *Proxy) handleTransparentProxy(w http.ResponseWriter, r *http.Request) {
	if r.Host == "localhost:8080" || r.Host == "localhost:8443" {
		http.Error(w, "Cannot proxy to self", http.StatusBadRequest)
		return
	}

	select {
	case <-r.Context().Done():
		http.Error(w, "Request cancelled or timed out", http.StatusGatewayTimeout)
		return
	default:
	}

	destinationHost := r.Host
	if destinationHost == "" {
		logger.LogTransparentProxyHandlerUnableToDetermineDestinationHost(w)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	destinationURL := &url.URL{
		Scheme:   scheme,
		Host:     destinationHost,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	proxy := p.createReverseProxy(destinationURL)
	proxy.ServeHTTP(w, r)
}

func (p *Proxy) Shutdown() {
	p.cancel()
	if p.workerPool != nil {
		p.StopWorkerPool()
	}
}