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