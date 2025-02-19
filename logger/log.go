package logger

import (
	"log"
	"net/http"
	"os"
)

var (
	LogFile *os.File
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
	log.Printf("Request: %s %s\n", req.Method, req.URL.String())
	for name, values := range req.Header {
		for _, value := range values {
			log.Printf("Header: %s: %s\n", name, value)
		}
	}
}

func LogResponse(res *http.Response) {
	log.Printf("Response: %s\n", res.Status)
	for name, values := range res.Header {
		for _, value := range values {
			log.Printf("Header: %s: %s\n", name, value)
		}
	}
}

func LogHTTPServerStart(port string) {
	log.Printf("Starting HTTP server on %s", port)
}

func LogHTTPSServerStart(port string) {
	log.Printf("Starting HTTPS server on %s", port)
}

func LogServerError(err error) {
	log.Fatalf("Server error: %v", err)
}

func LogCertificateError(err error) {
	log.Printf("Certificate error: %v", err)
}

func LogTransparentProxyHandlerUnableToDetermineDestinationHost(w http.ResponseWriter) {
	log.Printf("TransparentProxyHandler Error: Unable to determine target host")
	http.Error(w, "Unable to determine target host", http.StatusBadRequest)
}

func LogTransparentProxyHandlerFailedToParseTargetURL(w http.ResponseWriter, err error) {
	log.Printf("TransparentProxyHandler Error: Failed to parse target URL - %v", err)
	http.Error(w, "Failed to parse target URL", http.StatusInternalServerError)
}

func LogCustomHeaderError(customHeader string) {
    log.Printf("Invalid custom header format: %s\n", customHeader)
}

func KeepServerRunning() {
    log.Println("Proxy server is running")
    select {}
}