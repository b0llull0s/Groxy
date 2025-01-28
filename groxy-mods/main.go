package main

import (
    "flag"
    "log"
    "net/http"
    "net/url"
    "os"

)

var (
    logFile      *os.File
    targetURLStr string
    transparent  bool
)

func init() {
    var err error
    logFile, err = os.OpenFile("proxy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    log.SetOutput(logFile)
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