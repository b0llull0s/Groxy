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
    targetURLStr string
    transparent  bool
)

func main() {
    // Parse command-line flags
    flag.StringVar(&targetURLStr, "t", "", "Target URL for target-specific mode (e.g., http://10.10.10.80)")
    flag.BoolVar(&transparent, "transparent", false, "Run in transparent mode")
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
        log.Println("Starting transparent proxy server on :8080")
        http.HandleFunc("/", proxy.TransparentProxyHandler)
    } else {
        // Target-specific mode
        targetURL, err := url.Parse(targetURLStr)
        if err != nil {
            log.Fatalf("Failed to parse target URL: %v", err)
        }
        log.Printf("Starting target-specific proxy server on :8080 for target: %s", targetURLStr)
        http.HandleFunc("/", proxy.TargetSpecificProxyHandler(targetURL))
    }

    // Start HTTP proxy
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start HTTP proxy server: %v", err)
    }
}