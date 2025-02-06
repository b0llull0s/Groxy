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
	enableHTTP    bool
	enableHTTPS   bool
)

func main() {
	// Parse command-line flags
	flag.StringVar(&targetURLStr, "t", "", "Target URL for target-specific mode (e.g., http://10.10.10.80)")
	flag.BoolVar(&transparent, "transparent", false, "Run in transparent mode")
	flag.StringVar(&customHeader, "H", "", "Add a custom header (e.g., \"X-Request-ID: 12345\")")
	flag.BoolVar(&enableHTTP, "http", false, "Enable the HTTP server")
	flag.BoolVar(&enableHTTPS, "https", false, "Enable the HTTPS server")
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

	// Create the proxy handler
	var proxyHandler http.Handler
	if transparent {
		// Transparent mode
		proxyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy.TransparentProxyHandler(w, r, customHeader)
		})
	} else {
		// Target-specific mode
		targetURL, err := url.Parse(targetURLStr)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}

		proxyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy.TargetSpecificProxyHandler(targetURL, w, r, customHeader)
		})
	}

	// Register the handler for the root path "/"
	http.Handle("/", proxyHandler)

	// Start HTTP server if enabled
	if enableHTTP {
		go servers.StartHTTPServer(proxyHandler)
	}

	// Load certificates for HTTPS server
	certFile := "certs/server-cert.pem"
	keyFile := "certs/server-key.pem"

	// Start HTTPS server if enabled	
	if enableHTTPS {
		go servers.StartHTTPSServer(certFile, keyFile, proxyHandler)
	}

	logger.KeepServerRunning()
}