// 	// Start HTTPS proxy (optional)
// go func() {
//     log.Println("Starting HTTPS proxy server on :8443")
//     if err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil); err != nil {
//         log.Fatalf("Failed to start HTTPS proxy server: %v", err)
//     }
// }()


// Auth Module (commented out for now)
/*
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "password" { // Replace with your credentials
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
*/