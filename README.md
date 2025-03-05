# Groxy
- Groxy is a powerful and customizable `HTTP/HTTPS` proxy written in `Go`. It is designed to handle both transparent and target-specific proxying, with support for various authentication methods, custom headers, `User-Agent` rotation, `TLS` certificate management, dynamic certificate rotation traffic obfuscation, and worker pools for handling concurrent requests.
- Groxy is designed to be flexible, allowing you to configure it for different use cases, such as load balancing, traffic monitoring, or secure tunneling.
## Features
- `Transparent Proxy Mode`: Automatically forwards requests to the destination host without requiring explicit configuration.
- `Target-Specific Proxy Mode`: Directs traffic to a specific target URL.
- `Custom Headers`: Add custom headers to outgoing requests.
- `TLS Support`: Built-in support for `HTTPS` with dynamic certificate generation and rotation.
- `Request/Response Modification`: Modify incoming responses and outgoing requests on the fly.
- `Logging`: Comprehensive logging for both requests and responses.
- `Certificate Management`: Automatically generate and rotate TLS certificates for secure communication.
- `User-Agent Rotation`: Rotate `User-Agent` strings to mimic different browsers or devices.
- `HTTP/HTTPS Proxy`: Supports both `HTTP` and `HTTPS` traffic with automatic redirection from `HTTP` to `HTTPS`.
- `Worker Pools`: Specify how many workers should be created to handle incoming requests, and determine the buffer size for pending requests.
- `Authentication`: Supports multiple authentication methods, including token-based and basic authentication.
- `Traffic Obfuscation`: Encrypts and obfuscates traffic to prevent detection and tampering.
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
- `-transparent`: Run in transparent mode.
- `-H <header>`: Add a custom header to outgoing requests (e.g., X-Request-ID: 12345).
- `-http`: Enable the HTTP server (listens on port 8080).
- `-https`: Enable the HTTPS server (listens on port 8443).
- `-workers`: Determine the number of workers. Is set to `0` by default.
- `queue-size`: Detemine the buffer size for pending requests.
- `-timeout`: Timeout for requests in seconds. Is set to `30` seconds by default.
- `-obfuscate`: Enable Traffic obfuscation.
- `-redirect`: Enable `HTTP` to `HTTPS` redirection.
- `-auth-method`: Authentication method to use (`none`, `token`, or `basic`).
- `-auth-tokens`: Comma-separated list of valid tokens (for token-based authentication).
- `-auth-username`: Username for basic authentication.
- `-auth-password`: Password for basic authentication.
### Examples
- Transparent mode with `HTTP/HTTPS` redirection:
```bash   
./groxy -transparent -http -https -redirect
```
- Target mode with custom header and worker pool management:
```bash
./groxy -t http://example.com -http -workers=10 -queue-size=200 -H "X-Request-ID: 12345"
```
- Target mode with basic authentication and obfuscation:
```bash
./groxy -t http://example.com -http -auth-method=basic -auth-username=admin -auth-password=secret -obfuscate
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
- `tls/`: Manages `TLS` certificate generation, rotation, and configuration.
- `servers/`: Handles `HTTP/HTTPS` server initialization and management.
- `logger/`: Provides logging functionality for requests, responses, and errors.
- `certs/`: Stores `TLS` certificates and keys.
- `auth/`: Contains authentication-related code, including token-based and basic authentication.
## Contributing
If you'd like to contribute to Groxy, please follow these steps:
1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.