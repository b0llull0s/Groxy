package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// TransparentProxyHandler handles requests in transparent mode.
func TransparentProxyHandler(w http.ResponseWriter, r *http.Request, customHeader string) {
	log.Printf("Transparent Proxy Request: %s %s", r.Method, r.URL.String())
	targetHost := r.Host
	if targetHost == "" {
		http.Error(w, "Unable to determine target host", http.StatusBadRequest)
		return
	}

	targetURL, err := url.Parse("http://" + targetHost)
	if err != nil {
		log.Printf("Failed to parse target host: %v", err)
		http.Error(w, "Invalid target host", http.StatusBadRequest)
		return
	}

	proxy := CreateProxy(targetURL, customHeader)
	proxy.ServeHTTP(w, r)
}

// TargetSpecificProxyHandler handles requests in target-specific mode.
func TargetSpecificProxyHandler(targetURL *url.URL, w http.ResponseWriter, r *http.Request, customHeader string) {
	log.Printf("Target-Specific Proxy Request: %s %s", r.Method, r.URL.String())
	proxy := CreateProxy(targetURL, customHeader)
	proxy.ServeHTTP(w, r)
}

// CreateProxy creates a reverse proxy for the given target URL.
func CreateProxy(targetURL *url.URL, customHeader string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	ModifyRequest(proxy, customHeader) // Pass custom header
	ModifyResponse(proxy)
	return proxy
}