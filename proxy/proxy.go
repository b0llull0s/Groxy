package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"Groxy/logger"

)

// Creates a reverse proxy for the given target URL.
func CreateProxy(targetURL *url.URL, customHeader string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	ModifyRequest(proxy, customHeader)
	ModifyResponse(proxy)
	return proxy
}

// Transparent Mode
func TransparentProxyHandler(w http.ResponseWriter, r *http.Request, customHeader string) {
	// Extract the target host from the request
	targetHost := r.Host
	if targetHost == "" {
		logger.LogTransparentProxyHandlerUnableToDetermineTargetHost(w)
		return
	}

	// Construct the target URL
	targetURL, err := url.Parse("http://" + targetHost + r.URL.Path)
	if err != nil {
		logger.LogTransparentProxyHandlerFailedToParseTargetURL(w, err)
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

	proxy.ServeHTTP(w, r)
}

// Target Mode
func TargetSpecificProxyHandler(targetURL *url.URL, w http.ResponseWriter, r *http.Request, customHeader string) {
	// Create a reverse proxy
	proxy := CreateProxy(targetURL, customHeader)

	// Set up TLS configuration for HTTPS targets
	if targetURL.Scheme == "https" {
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Skip cert verification for development
		}
	}

	proxy.ServeHTTP(w, r)
}