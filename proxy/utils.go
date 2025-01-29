package proxy

import (
	"net/url"
	"net/http" 
)

// Checks if a URL is valid.
func ValidateURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// Limits the number of requests per second.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implement rate limiting logic here
		next.ServeHTTP(w, r)
	})
}