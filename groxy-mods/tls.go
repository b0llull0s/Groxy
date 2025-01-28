	// Start HTTPS proxy
go func() {
    log.Println("Starting HTTPS proxy server on :8443")
    if err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil); err != nil {
        log.Fatalf("Failed to start HTTPS proxy server: %v", err)
    }
}()