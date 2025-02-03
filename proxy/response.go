package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"Groxy/logger" 
)

// Modifies incoming responses.
func ModifyResponse(proxy *httputil.ReverseProxy) {
	proxy.ModifyResponse = func(res *http.Response) error {
		logger.LogResponse(res)

		// Modify the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		modifiedBody := []byte("Modified: " + string(body))
		res.Body = io.NopCloser(bytes.NewReader(modifiedBody))
		res.ContentLength = int64(len(modifiedBody))
		res.Header.Set("Content-Length", fmt.Sprint(len(modifiedBody)))

		return nil
	}
}