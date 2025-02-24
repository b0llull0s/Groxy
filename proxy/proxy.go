package proxy

import (
	"Groxy/logger"
	"Groxy/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	targetURL    *url.URL
	tlsConfig    *tls.Config
	customHeader string
	workerPool   *WorkerPool
	useWorkers   bool
}

func NewProxy(targetURL *url.URL, tlsConfig *tls.Config, customHeader string) *Proxy {
	return &Proxy{
		targetURL:    targetURL,
		tlsConfig:    tlsConfig,
		customHeader: customHeader,
		useWorkers:   false,
	}
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
	
	if targetURL.Scheme == "https" {
		proxy.Transport = &http.Transport{
			TLSClientConfig: p.tlsConfig.LoadClientConfig(),
		}
	}
	
	ModifyRequest(proxy, p.customHeader)
	ModifyResponse(proxy)
	return proxy
}

func (p *Proxy) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.useWorkers {
			p.workerPool.Submit(w, r)
			return
		}
		
		if p.targetURL == nil {
			p.handleTransparentProxy(w, r)
		} else {
			proxy := p.createReverseProxy(p.targetURL)
			proxy.ServeHTTP(w, r)
		}
	})
}

func (p *Proxy) handleTransparentProxy(w http.ResponseWriter, r *http.Request) {
	if r.Host == "localhost:8080" || r.Host == "localhost:8443" {
		http.Error(w, "Cannot proxy to self", http.StatusBadRequest)
		return
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
		Scheme: scheme,
		Host:   destinationHost,
		Path:   r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	proxy := p.createReverseProxy(destinationURL)
	proxy.ServeHTTP(w, r)
}