# Groxy
Fancy Proxy written on GO with DeepSeek just to make you mad

## TLS Support
To add HTTPS support, you’ll need to generate a self-signed certificate or use a trusted one. Here’s how to add TLS:
- Generate a Self-Signed Certificate
```shell
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes
```
