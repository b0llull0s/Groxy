package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	logFile *os.File
)

// Log Module
func init() {
	var err error
	logFile, err = os.OpenFile("proxy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
}

func logRequest(req *http.Request) {
	log.Printf("Request: %s %s\n", req.Method, req.URL.String())
	for name, values := range req.Header {
		for _, value := range values {
			log.Printf("Header: %s: %s\n", name, value)
		}
	}
}

func logResponse(res *http.Response) {
	log.Printf("Response: %s\n", res.Status)
	for name, values := range res.Header {
		for _, value := range values {
			log.Printf("Header: %s: %s\n", name, value)
		}
	}
}

// Director Module
func modifyRequest(proxy *httputil.ReverseProxy) {
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		logRequest(req) // Log the request
		req.Header.Add("X-Custom-Header", "MyProxy") // Custom header 
	}
}

// Modify Response Module
func modifyResponse(proxy *httputil.ReverseProxy) {
	proxy.ModifyResponse = func(res *http.Response) error {
		logResponse(res) // Log the response

		// Modify the response body (optional)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		modifiedBody := []byte("Modified: " + string(body))
		res.Body = io.NopCloser(bytes.NewReader(modifiedBody))
		res.ContentLength = int64(len(modifiedBody))
		res.Header.Set("Content-Length", fmt.Sprint(len(modifiedBody)))

		return nil
	}
}

// Auth Module (commented out for now)
/*
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "password" { // Replace with your credentials
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
*/

func main() {
	defer logFile.Close()

	// Target URL (replace with your target)
	targetURL, err := url.Parse("http://example")
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Apply modules
	modifyRequest(proxy)  // Add request modification
	modifyResponse(proxy) // Add response modification

	// Handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	// Start HTTP proxy
	log.Println("Starting HTTP proxy server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start HTTP proxy server: %v", err)
	}

// 	// Start HTTPS proxy (optional)
// go func() {
//     log.Println("Starting HTTPS proxy server on :8443")
//     if err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil); err != nil {
//         log.Fatalf("Failed to start HTTPS proxy server: %v", err)
//     }
// }()
}