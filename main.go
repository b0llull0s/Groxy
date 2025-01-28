package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	logFile      *os.File
	targetURLStr string
	transparent  bool
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
		req.Header.Add("X-Custom-Header", "MyProxy") // Add a custom header
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

// Transparent Proxy Handler
func transparentProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the target host from the request
	targetHost := r.Host
	if targetHost == "" {
		http.Error(w, "Unable to determine target host", http.StatusBadRequest)
		return
	}

	// Parse the target URL
	targetURL, err := url.Parse("http://" + targetHost)
	if err != nil {
		http.Error(w, "Invalid target host", http.StatusBadRequest)
		return
	}

	// Create a reverse proxy for the target
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Apply modules
	modifyRequest(proxy)
	modifyResponse(proxy)

	// Serve the request
	proxy.ServeHTTP(w, r)
}

// Target-Specific Proxy Handler
func targetSpecificProxyHandler(targetURL *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a reverse proxy for the target
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Apply modules
		modifyRequest(proxy)
		modifyResponse(proxy)

		// Serve the request
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	defer logFile.Close()

	// Parse command-line flags
	flag.StringVar(&targetURLStr, "t", "", "Target URL for target-specific mode (e.g., http://10.10.10.80)")
	flag.BoolVar(&transparent, "transparent", false, "Run in transparent mode")
	flag.Parse()

	// Validate flags
	if targetURLStr == "" && !transparent {
		log.Fatalf("You must specify either -t <target> or --transparent")
	}
	if targetURLStr != "" && transparent {
		log.Fatalf("You cannot specify both -t <target> and --transparent")
	}

	// Start the proxy in the appropriate mode
	if transparent {
		// Transparent mode
		log.Println("Starting transparent proxy server on :8080")
		http.HandleFunc("/", transparentProxyHandler)
	} else {
		// Target-specific mode
		targetURL, err := url.Parse(targetURLStr)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}
		log.Printf("Starting target-specific proxy server on :8080 for target: %s", targetURLStr)
		http.HandleFunc("/", targetSpecificProxyHandler(targetURL))
	}

	// Start HTTP proxy
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start HTTP proxy server: %v", err)
	}
}