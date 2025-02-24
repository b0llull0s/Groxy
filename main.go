package main

import (
	"context"
	"flag"
	"net/url"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Groxy/logger"
	"Groxy/proxy"
	"Groxy/servers"
	"Groxy/tls"
	cryptotls "crypto/tls"
)

var (
	targetURLStr  string
	transparent   bool
	customHeader  string
	enableHTTP    bool
	enableHTTPS   bool
)

func main() {
	flag.StringVar(&targetURLStr, "t", "", "Target URL for target-specific mode (e.g., http://10.10.10.80)")
	flag.BoolVar(&transparent, "transparent", false, "Run in transparent mode")
	flag.StringVar(&customHeader, "H", "", "Add a custom header (e.g., \"X-Request-ID: 12345\")")
	flag.BoolVar(&enableHTTP, "http", false, "Enable the HTTP server")
	flag.BoolVar(&enableHTTPS, "https", false, "Enable the HTTPS server")
	flag.Parse()

	logger.Init()
	defer logger.LogFile.Close()

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

	tlsConfig := tls.NewConfig("certs/server-cert.pem", "certs/server-key.pem")
	tlsManager := tls.NewManager(tlsConfig)

	tlsManager.OnRotation = func(cert *cryptotls.Certificate) {
		fmt.Println("Certificate rotated successfully")
	}
	tlsManager.OnError = func(err error) {
		fmt.Printf("Certificate rotation error: %v\n", err)
	}

	var targetURL *url.URL
	if !transparent {
		var err error
		targetURL, err = url.Parse(targetURLStr)
		if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
			fmt.Println("Failed to parse target URL: Invalid URL format")
			os.Exit(1)
		}
	}

	proxy := proxy.NewProxy(targetURL, tlsConfig, customHeader)

	server := servers.NewServer(
		proxy.Handler(),
		tlsManager,
		"certs/server-cert.pem",
		"certs/server-key.pem",
		"8080",
		"8443",
	)

	if enableHTTP {
		if err := server.StartHTTP(); err != nil {
			fmt.Printf("Failed to start HTTP server: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("HTTP server is running on port 8080")
	}
	if enableHTTPS {
		if err := server.StartHTTPS(); err != nil {
			fmt.Printf("Failed to start HTTPS server: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("HTTPS server is running on port 8443")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("Received signal %v, shutting down gracefully...\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server shutdown error: %v\n", err)
	}

	fmt.Println("Server shutdown complete")
}