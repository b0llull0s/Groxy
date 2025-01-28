# Groxy
Fancy Proxy written on GO with DeepSeek just to make you mad

## TLS Support
To add HTTPS support, you’ll need to generate a self-signed certificate or use a trusted one. Here’s how to add TLS:
- Generate a Self-Signed Certificate
```shell
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes
```
- Update the Proxy to Use TLS
```go
log.Println("Starting HTTPS proxy server on :8443")
if err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil); err != nil {
	log.Fatalf("Failed to start HTTPS proxy server: %v", err)
}
```
