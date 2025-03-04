package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"Groxy/logger" 
)

func ModifyResponse(proxy *httputil.ReverseProxy, obfuscator *TrafficObfuscator) {
	proxy.ModifyResponse = func(res *http.Response) error {
		logger.LogResponse(res)

		if obfuscator != nil {
			extractedData, err := obfuscator.ExtractFromResponse(res)
			if err != nil {
				logger.Error("Failed to extract data from obfuscated response: %v", err)
				return nil
			}
			if len(extractedData) > 0 {
				res.Body = io.NopCloser(bytes.NewReader(extractedData))
				res.ContentLength = int64(len(extractedData))
				res.Header.Set("Content-Length", fmt.Sprint(len(extractedData)))
				return nil
			}
		}
		return nil
	}
}