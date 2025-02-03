package servers

import (
	"Groxy/logger"
	"net/http"
)

func StartHTTPServer(addr string, proxyHandler http.Handler) {
	logger.LogHTTPServerStart(addr)
	
	go func() {
		if err := http.ListenAndServe(addr, proxyHandler); err != nil {
			logger.LogServerError(err)
		}
	}()
}
