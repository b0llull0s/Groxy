package main

import (
	"flag"
	"net/http"
	"net/url"
	"fmt"
	"os"

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
	if !transparent && targetURLStr == "" {
		fmt.Println("Error: You must specify either -t <target> or --transparent")
		flag.Usage()
		os.Exit(1)
	}

	if transparent && targetURLStr != "" {
		fmt.Println("Error: You cannot specify both -t <target> and --transparent")
		flag.Usage()
		os.Exit(1)
	}

	if !enableHTTP && !enableHTTPS {
		fmt.Println("Error: You must specify either -http or -https to enable the server")
		flag.Usage()
		os.Exit(1)
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
		if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
			fmt.Println("Failed to parse target URL: Invalid URL format")
			os.Exit(1)
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
		fmt.Println("HTTP server is running on port 8080")
	}

	// Load certificates for HTTPS server
	certFile := "certs/server-cert.pem"
	keyFile := "certs/server-key.pem"

	// Start HTTPS server if enabled	
	if enableHTTPS {
		go servers.StartHTTPSServer(certFile, keyFile, proxyHandler)
		fmt.Println("HTTP server is running on port 8443")
	}

	logger.KeepServerRunning()
}