# Groxy
- Groxy is a modular proxy written in `Go`, designed to be integrated with a `C2` server in the future.
## Features

- `Transparent Mode`: Operate in a mode that intercepts all the traffic.
- `Target Mode`: Intercepts traffic for a specific target.
- `Custom Headers`: Add custom headers to requests with the `-H` flag.
- `User-Agent Rotation`: Rotate `User-Agent` strings to mimic different browsers or devices.
- `HTTPS Server`: Deploy an HTTPS server.
- `Logging`: Detailed logging of proxy activity.

## Planned Features

- `TLS` Encryption: Enhance security with TLS encryption.
- `Certificate rotation`: Periodically replacing SSL/TLS certificates with new ones to ensure the ongoing security of the connection.
- `Authorization`: Controls who can access the proxy and what actions they can perform.
- `Polyglot` Features: Add support for `C++` and multiple protocols.
- `Rate Limiting Middleware`: Implement rate limiting to control traffic flow.
- `Validation` for incoming and outgoing requests.

## Certificates
- The certificates provided in the repository are for testing purposes.
- You can generate your own certificates using the following guide:
### Generating Certificates
- Generate `CA` private key
```
openssl genrsa -out ca-key.pem 2048
```
- Generate `CA` certificate
```
openssl req -x509 -new -nodes -key ca-key.pem -sha256 -days 365 -out ca-cert.pem -subj "/CN=My CA"
```
- Create `openssl.cnf` file:
```
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
x509_extensions = v3_ca

[dn]
C = US
ST = California
L = San Francisco
O = Your Organization
CN = localhost

[v3_ca]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
```
- Generate server private key
```
openssl genrsa -out server-key.pem 2048
```
- Create a certificate signing request (`CSR`)
```
openssl req -new -key server-key.pem -out server-req.pem -config openssl.cnf
```
- Sign the `CSR` with the `CA` certificate:
```
openssl x509 -req -in server-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -days 365 -sha256 -extensions v3_ca -extfile openssl.cnf
```
- Check the certificate to ensure it includes `127.0.0.1` in the `SAN` field:
```
openssl x509 -in server-cert.pem -text -noout
```
- Look for the following section:
```
    X509v3 Subject Alternative Name:
        DNS:localhost, IP Address:127.0.0.1
```
- To test the server, use the following command:
```
curl --cacert certs/ca-cert.pem https://127.0.0.1:8443
```