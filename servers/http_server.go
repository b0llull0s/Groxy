package servers

import (
	"Groxy/logger"
	"net/http"
)

const HTTPPort = "8080"

func StartHTTPServer(proxyHandler http.Handler) {
		addr := ":" + HTTPPort
	go func() {
		logger.LogHTTPServerStart(HTTPPort)
		if err := http.ListenAndServe(addr, proxyHandler); err != nil {
			logger.LogServerError(err)
		}
	}()
}
