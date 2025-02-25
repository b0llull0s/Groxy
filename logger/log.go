package logger

import (
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	LogFile *os.File
	mu      sync.Mutex /
)

func Init() {
	var err error
	LogFile, err = os.OpenFile("proxy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	
	log.SetOutput(LogFile)
}

func LogRequest(req *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("Request: %s %s\n", req.Method, req.URL.String())
	for name, values := range req.Header {
		for _, value := range values {
			log.Printf("Header: %s: %s\n", name, value)
		}
	}
}

func LogResponse(res *http.Response) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("Response: %s\n", res.Status)
	for name, values := range res.Header {
		for _, value := range values {
			log.Printf("Header: %s: %s\n", name, value)
		}
	}
}

func LogHTTPServerStart(port string) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("Starting HTTP server on port %s", port)
}

func LogHTTPSServerStart(port string) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("Starting HTTPS server on port %s", port)
}

func LogServerError(err error) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("Server error: %v", err)
}

func LogServerShutdown(serverType string) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("%s server shutting down", serverType)
}

func LogServerShutdownComplete(serverType string) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("%s server shutdown complete", serverType)
}

func LogGracefulShutdownStarted() {
	mu.Lock()
	defer mu.Unlock()
	
	log.Println("Graceful shutdown initiated")
}

func LogGracefulShutdownComplete() {
	mu.Lock()
	defer mu.Unlock()
	
	log.Println("Graceful shutdown completed")
}

func LogCertificateError(err error) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("Certificate error: %v", err)
}

func LogCertificateRotation() {
	mu.Lock()
	defer mu.Unlock()
	
	log.Println("TLS certificate rotated")
}

func LogTransparentProxyHandlerUnableToDetermineDestinationHost(w http.ResponseWriter) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("TransparentProxyHandler Error: Unable to determine target host")
	http.Error(w, "Unable to determine target host", http.StatusBadRequest)
}

func LogTransparentProxyHandlerFailedToParseTargetURL(w http.ResponseWriter, err error) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("TransparentProxyHandler Error: Failed to parse target URL - %v", err)
	http.Error(w, "Failed to parse target URL", http.StatusInternalServerError)
}

func LogCustomHeaderError(customHeader string) {
	mu.Lock()
	defer mu.Unlock()
	
    log.Printf("Invalid custom header format: %s\n", customHeader)
}

func LogRequestTimeout(r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	
    log.Printf("Request timeout or cancelled: %s %s\n", r.Method, r.URL.String())
}

func LogContextCancelled(reason string) {
	mu.Lock()
	defer mu.Unlock()
	
    log.Printf("Context cancelled: %s\n", reason)
}

func KeepServerRunning() {
	mu.Lock()
	defer mu.Unlock()
	
    log.Println("Proxy server is running")
    select {}
}