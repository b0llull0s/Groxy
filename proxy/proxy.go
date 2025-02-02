package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// CreateProxy creates a reverse proxy for the given target URL.
func CreateProxy(targetURL *url.URL, customHeader string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	ModifyRequest(proxy, customHeader) // Pass custom header
	ModifyResponse(proxy)
	return proxy
}

// TransparentProxyHandler handles requests in transparent mode.
func TransparentProxyHandler(w http.ResponseWriter, r *http.Request, customHeader string) {
	// Extract the target host from the request
	targetHost := r.Host
	if targetHost == "" {
		http.Error(w, "Unable to determine target host", http.StatusBadRequest)
		return
	}

	// Construct the target URL
	targetURL, err := url.Parse("http://" + targetHost + r.URL.Path)
	if err != nil {
		http.Error(w, "Failed to parse target URL", http.StatusInternalServerError)
		return
	}

	// Create a reverse proxy
	proxy := CreateProxy(targetURL, customHeader)

	// Set up TLS configuration for HTTPS targets
	if r.URL.Scheme == "https" {
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Skip cert verification for development
		}
	}

	// Serve the request
	proxy.ServeHTTP(w, r)
}

// TargetSpecificProxyHandler handles requests in target-specific mode.
func TargetSpecificProxyHandler(targetURL *url.URL, w http.ResponseWriter, r *http.Request, customHeader string) {
	// Create a reverse proxy
	proxy := CreateProxy(targetURL, customHeader)

	// Set up TLS configuration for HTTPS targets
	if targetURL.Scheme == "https" {
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Skip cert verification for development
		}
	}

	// Serve the request
	proxy.ServeHTTP(w, r)
}