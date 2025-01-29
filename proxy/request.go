package proxy

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"time"
	"Groxy/logger"
)

var (
	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	}
)

// Modifies outgoing requests.
func ModifyRequest(proxy *httputil.ReverseProxy) {
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		logger.LogRequest(req) 

		// Add custom headers
		req.Header.Add("X-Custom-Header", "MyProxy")

		// Rotate user-agent
		req.Header.Set("User-Agent", getRandomUserAgent())
	}
}

// Returns a random user-agent from the list.
func getRandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}