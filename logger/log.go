package logger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

const (
	LevelInfo    = "INFO"
	LevelWarning = "WARNING"
	LevelError   = "ERROR"
	LevelDebug   = "DEBUG"
)

var (
	LogFile *os.File
	mu      sync.Mutex 
)

func Init() {
	var err error
	LogFile, err = os.OpenFile("proxy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	
	log.SetOutput(LogFile)
}

func logWithLevel(level, format string, v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	
	msg := fmt.Sprintf(format, v...)
	log.Printf("[%s] %s", level, msg)
}

func Info(format string, v ...interface{}) {
	logWithLevel(LevelInfo, format, v...)
}

func Warning(format string, v ...interface{}) {
	logWithLevel(LevelWarning, format, v...)
}

func Error(format string, v ...interface{}) {
	logWithLevel(LevelError, format, v...)
}

func Debug(format string, v ...interface{}) {
	logWithLevel(LevelDebug, format, v...)
}

func LogRequest(req *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("[INFO] Request: %s %s", req.Method, req.URL.String())
	for name, values := range req.Header {
		for _, value := range values {
			log.Printf("[DEBUG] Header: %s: %s", name, value)
		}
	}
}

func LogResponse(res *http.Response) {
	mu.Lock()
	defer mu.Unlock()
	
	log.Printf("[INFO] Response: %s", res.Status)
	for name, values := range res.Header {
		for _, value := range values {
			log.Printf("[DEBUG] Header: %s: %s", name, value)
		}
	}
}

func ServerEvent(eventType, details string) {
	Info("%s: %s", eventType, details)
}

func RequestError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		Error("%s: %v", message, err)
	} else {
		Error("%s", message)
	}
	
	if w != nil {
		http.Error(w, message, status)
	}
}

// --- Backward compatibility functions ---

func LogHTTPServerStart(port string) {
	Info("Starting HTTP server on port %s", port)
}

func LogHTTPSServerStart(port string) {
	Info("Starting HTTPS server on port %s", port)
}

func LogServerError(err error) {
	Error("Server error: %v", err)
}

func LogServerShutdown(serverType string) {
	Info("%s server shutting down", serverType)
}

func LogServerShutdownComplete(serverType string) {
	Info("%s server shutdown complete", serverType)
}

func LogGracefulShutdownStarted() {
	Info("Graceful shutdown initiated")
}

func LogGracefulShutdownComplete() {
	Info("Graceful shutdown completed")
}

func LogCertificateError(err error) {
	Error("Certificate error: %v", err)
}

func LogCertificateRotation() {
	Info("TLS certificate rotated")
}

func LogTransparentProxyHandlerUnableToDetermineDestinationHost(w http.ResponseWriter) {
	RequestError(w, http.StatusBadRequest, "Unable to determine target host", nil)
}

func LogTransparentProxyHandlerFailedToParseTargetURL(w http.ResponseWriter, err error) {
	RequestError(w, http.StatusInternalServerError, "Failed to parse target URL", err)
}

func LogCustomHeaderError(customHeader string) {
	Warning("Invalid custom header format: %s", customHeader)
}

func LogRequestTimeout(r *http.Request) {
	Warning("Request timeout or cancelled: %s %s", r.Method, r.URL.String())
}

func LogContextCancelled(reason string) {
	Warning("Context cancelled: %s", reason)
}

func KeepServerRunning() {
	Info("Proxy server is running")
	select {}
}