package proxy

import (
	"net/http"
)

// Injects a custom payload into the request.
func InjectPolyglotPayload(req *http.Request) {
	req.Header.Add("X-Polyglot-Payload", "CustomPayload")
}

// Processes requests with polyglot payloads.
func HandlePolyglotRequest(req *http.Request) {
	if req.URL.Query().Get("polyglot") == "true" {
		InjectPolyglotPayload(req)
	}
}