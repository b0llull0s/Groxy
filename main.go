package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"

	"Groxy/logger"
	"Groxy/proxy"
	"Groxy/servers" 
)

var (
	targetURLStr  string
	transparent   bool
	customHeader  string
	enableServer   bool 
)

func main() {
	// Parse command-line flags
	flag.StringVar(&targetURLStr, "t", "", "Target URL for target-specific mode (e.g., http://10.10.10.80)")
	flag.BoolVar(&transparent, "transparent", false, "Run in transparent mode")
	flag.StringVar(&customHeader, "H", "", "Add a custom header (e.g., \"X-Request-ID: 12345\")")
	flag.BoolVar(&enableServer, "server", false, "Enable the HTTPS server") 
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

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.TransparentProxyHandler(w, r, customHeader)
		})
	} else {
		// Target-specific mode
		targetURL, err := url.Parse(targetURLStr)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.TargetSpecificProxyHandler(targetURL, w, r, customHeader)
		})
	}

	// Start HTTP server
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start HTTP proxy server: %v", err)
		}
	}()

		// Start HTTPS server if enabled
		if enableServer {
			log.Println("Starting HTTPS server on :8443")
			go servers.StartHTTPSServer(":8443", "certs/server-cert.pem", "certs/server-key.pem")
		}

	// Keep the program running
	log.Println("Proxy server is running")
	select {}
}