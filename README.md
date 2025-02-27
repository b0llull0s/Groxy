# Groxy
Groxy is a powerful and customizable HTTP/HTTPS proxy written in Go. It is designed to handle both transparent and target-specific proxying, with support for custom headers, User-Agent rotation, TLS certificate management, and dynamic certificate rotation.
## Features
- `Transparent Proxy Mode`: Automatically forwards requests to the destination host without requiring explicit configuration.
- `Target-Specific Proxy Mode`: Directs traffic to a specific target URL.
- `Custom Headers`: Add custom headers to outgoing requests.
- `TLS Support`: Built-in support for HTTPS with dynamic certificate generation and rotation.
- `Request/Response Modification`: Modify incoming responses and outgoing requests on the fly.
- `Logging`: Comprehensive logging for both requests and responses.
- `Certificate Management`: Automatically generate and rotate TLS certificates for secure communication.
- `User-Agent Rotation`: Rotate `User-Agent` strings to mimic different browsers or devices.
- `HTTP to HTTPS redirection`: Redirection is set to `True` by default but this can be change in `server.go`.
- `Worker Pools for Request Handling`: Specify how many workers should be created to handle incoming requests, and determine the buffer size for pending requests.
- `Obfuscation`: The obfuscation occurs during the HTTP request/response processing.
## Installation
1. Clone the Repository:
```bash
git clone https://github.com/yourusername/Groxy.git
cd Groxy
```
2. Build the Project:
```bash
go build -o groxy
```
3. Run the Proxy:
```bash
./groxy -http -https -H "X-Custom-Header: MyValue"
```
## Usage
Command-Line Options
- `-t <target>`: Specify the target URL for target-specific mode (e.g., http://example.com).
- `--transparent`: Run in transparent mode.
- `-H <header>`: Add a custom header to outgoing requests (e.g., X-Request-ID: 12345).
- `-http`: Enable the HTTP server (listens on port 8080).
- `-https`: Enable the HTTPS server (listens on port 8443).
- `-workers`: Determine the number of workers. Is set to `0` by default.
- `queue-size`: Detemine the buffer size for pending requests.
- `timeout`: Timeout for requests in seconds. Is set to `30` seconds by default.
- `obfuscate`: Traffic obfuscation mode: 0=None, 1=HttpHeaders, 2=DomainFronting, 3=CustomJitter.
### Examples
- `Transparent Mode`:
```bash   
./groxy --transparent -http -https
```
- `Target-Specific Mode`:
```bash
./groxy -t http://example.com -http -https -H "X-Request-ID: 12345"
```
- `Custom Header`:
```bash
./groxy -t http://example.com -http -H "Authorization: Bearer token"
```
## Configuration
### TLS Certificates
- Certificates are stored in the `certs` directory:
   - `certs/server-cert.pem`: The server certificate.
   - `certs/server-key.pem`: The server private key.
- You can replace these files with your own certificates if needed.
- The certificates provided in the repository are for testing purposes.
### Logging
- Logs are stored in the `logs` directory.
- You can customize the logging behavior by modifying the `logger/log.go` file.
## Code Structure
- `proxy/`: Contains the core proxy logic, including request/response modification and transparent/target-specific handling.
- `tls/`: Manages TLS certificate generation, rotation, and configuration.
- `servers/`: Handles HTTP/HTTPS server initialization and management.
- `logger/`: Provides logging functionality for requests, responses, and errors.
- `certs/`: Stores TLS certificates and keys.
## Contributing
If you'd like to contribute to Groxy, please follow these steps:
1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.