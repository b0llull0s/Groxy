package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"

	"Groxy/logger"   
	"Groxy/proxy"    
)

var (
	targetURLStr  string
	transparent   bool
	customHeader  string 
)

func main() {
	// Parse command-line flags
	flag.StringVar(&targetURLStr, "t", "", "Target URL for target-specific mode (e.g., http://10.10.10.80)")
	flag.BoolVar(&transparent, "transparent", false, "Run in transparent mode")
	flag.StringVar(&customHeader, "H", "", "Add a custom header (e.g., \"X-Request-ID: 12345\")") // New flag
	flag.Parse()

	// Initialize logging
	logger.Init()
	defer logger.LogFile.Close()

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
		log.Println("Starting transparent proxy server on :8080 (HTTP) and :8443 (HTTPS)")
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.TransparentProxyHandler(w, r, customHeader) // Pass custom header
		})
	} else {
		// Target-specific mode
		targetURL, err := url.Parse(targetURLStr)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}
		log.Printf("Starting target-specific proxy server on :8080 (HTTP) and :8443 (HTTPS) for target: %s", targetURLStr)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.TargetSpecificProxyHandler(targetURL, w, r, customHeader) // Pass custom header
		})
	}

	// Start HTTP server
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start HTTP proxy server: %v", err)
		}
	}()

	// Keep the program running
	select {}
}