# DuDu Proxy

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**DuDu Proxy** is a high-performance proxy server supporting both HTTP and SOCKS5 protocols, featuring authentication, IP banning, rate limiting, and circuit breaker capabilities.

[中文文档](README_CN.md)

## ✨ Features

- **Multi-Protocol Support**
  - HTTP/HTTPS proxy (including CONNECT tunneling)
  - SOCKS5 proxy with full protocol support

- **Security & Authentication**
  - Multi-user authentication with username/password
  - IP ban mechanism with configurable thresholds
  - IP whitelist support

- **Traffic Control**
  - Global and per-IP rate limiting using Token Bucket algorithm
  - Circuit breaker with sliding window for protection against burst failures
  - Automatic recovery mechanisms

- **Operations**
  - JSON-based configuration
  - Structured logging with multiple formats (JSON/Console)
  - Graceful shutdown
  - Docker support

## 📦 Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/seakee/dudu-proxy/releases):

**Linux:**
```bash
# AMD64
wget https://github.com/seakee/dudu-proxy/releases/latest/download/dudu-proxy-v*-linux-amd64.zip
unzip dudu-proxy-v*-linux-amd64.zip
chmod +x dudu-proxy
./dudu-proxy -config config.json

# ARM64
wget https://github.com/seakee/dudu-proxy/releases/latest/download/dudu-proxy-v*-linux-arm64.zip
unzip dudu-proxy-v*-linux-arm64.zip
chmod +x dudu-proxy
./dudu-proxy -config config.json
```

**macOS:**
```bash
# Intel (AMD64)
curl -LO https://github.com/seakee/dudu-proxy/releases/latest/download/dudu-proxy-v*-darwin-amd64.zip
unzip dudu-proxy-v*-darwin-amd64.zip
chmod +x dudu-proxy
./dudu-proxy -config config.json

# Apple Silicon (ARM64)
curl -LO https://github.com/seakee/dudu-proxy/releases/latest/download/dudu-proxy-v*-darwin-arm64.zip
unzip dudu-proxy-v*-darwin-arm64.zip
chmod +x dudu-proxy
./dudu-proxy -config config.json
```

**Windows:**
```powershell
# Download and extract ZIP from releases page
# The ZIP contains dudu-proxy.exe and config.json
dudu-proxy.exe -config config.json
```

**Verify checksums:**
```bash
# Download checksums file
wget https://github.com/seakee/dudu-proxy/releases/latest/download/checksums.txt

# Verify ZIP files (Linux/macOS)
sha256sum -c checksums.txt

# Or on macOS
shasum -a 256 -c checksums.txt
```

### From Source

```bash
# Clone the repository
git clone https://github.com/seakee/dudu-proxy.git
cd dudu-proxy

# Build
make build

# Or build for all platforms
make build-all
```

### Using Docker

```bash
# Build Docker image
make docker

# Or use docker-compose
docker-compose up -d
```

## 🚀 Quick Start

1. **Configure the proxy**

   Copy the example config and modify as needed:
   ```bash
   cp configs/config.example.json configs/config.json
   ```

2. **Run the proxy**

   ```bash
   # Using make
   make run

   # Or run directly
   ./build/dudu-proxy -config configs/config.json
   ```

3. **Test the proxy**

   ```bash
   # HTTP proxy
   curl -x http://user1:pass1@localhost:8080 http://www.google.com

   # HTTPS proxy
   curl -x http://user1:pass1@localhost:8080 https://www.google.com

   # SOCKS5 proxy
   curl --socks5 user1:pass1@localhost:1080 http://www.google.com
   ```

## ⚙️ Configuration

Configuration is managed through a JSON file. Here's a complete example:

```json
{
  "server": {
    "http_port": 8080,       // HTTP proxy port
    "socks5_port": 1080,     // SOCKS5 proxy port
    "network": "tcp"         // Network type: tcp (dual-stack), tcp4 (IPv4 only), tcp6 (IPv6 only)
  },
  "auth": {
    "enabled": true,          // Enable authentication
    "users": [
      {"username": "user1", "password": "pass1"},
      {"username": "user2", "password": "pass2"}
    ]
  },
  "ip_ban": {
    "enabled": true,          // Enable IP banning
    "max_failures": 3,        // Max auth failures before ban
    "ban_duration_seconds": 300,  // Ban duration (5 minutes)
    "whitelist": ["127.0.0.1"]  // IPs that are never banned
  },
  "rate_limit": {
    "enabled": true,
    "global_requests_per_second": 1000,  // Global rate limit
    "per_ip_requests_per_second": 10     // Per-IP rate limit
  },
  "circuit_breaker": {
    "enabled": true,
    "failure_threshold_percent": 50,   // Failure rate to trip circuit
    "window_size_seconds": 60,          // Time window for stats
    "min_requests": 20,                  // Min requests before opening
    "break_duration_seconds": 30        // Circuit open duration
  },
  "log": {
    "level": "info",                   // debug, info, warn, error
    "driver": "file",                  // file, stdout
    "path": "logs/"                    // log file path
  }
} 
```

### Configuration Options

| Section | Option | Description | Default |
|---------|--------|-------------|---------|
| `server` | `http_port` | HTTP proxy listening port | 8080 |
| `server` | `socks5_port` | SOCKS5 proxy listening port | 1080 |
| `server` | `network` | Network type (tcp, tcp4, tcp6) | tcp |
| `auth` | `enabled` | Enable user authentication | false |
| `auth` | `users` | List of username/password pairs | [] |
| `ip_ban` | `enabled` | Enable IP ban on auth failures | false |
| `ip_ban` | `max_failures` | Number of failures before ban | 3 |
| `ip_ban` | `ban_duration_seconds` | Ban duration in seconds | 300 |
| `ip_ban` | `whitelist` | IPs exempt from banning | [] |
| `rate_limit` | `enabled` | Enable rate limiting | false |
| `rate_limit` | `global_requests_per_second` | Global RPS limit | 1000 |
| `rate_limit` | `per_ip_requests_per_second` | Per-IP RPS limit | 10 |
| `circuit_breaker` | `enabled` | Enable circuit breaker | false |
| `circuit_breaker` | `failure_threshold_percent` | Failure % to open circuit | 50 |
| `circuit_breaker` | `window_size_seconds` | Stats window size | 60 |
| `circuit_breaker` | `min_requests` | Min requests in window | 20 |
| `circuit_breaker` | `break_duration_seconds` | Circuit open time | 30 |
| `log` | `level` | Logging level | info |
| `log` | `driver` | Logging driver | file |
| `log` | `path` | Log file path | logs/ |

## 🛠️ Development

### Prerequisites

- Go 1.24 or higher
- Make (optional, for using Makefile)

### Building

```bash
# Build binary for current platform
make build

# Build for all platforms
make build-all

# Build for specific platforms
make build-linux    # Linux (amd64 + arm64)
make build-darwin   # macOS (amd64 + arm64)
make build-windows  # Windows (amd64 + arm64)

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Clean build artifacts
make clean
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## 📊 Architecture

```
dudu-proxy/
├── cmd/dudu-proxy/         # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── proxy/              # HTTP and SOCKS5 proxy implementations
│   ├── middleware/         # Auth, rate limit, IP ban, circuit breaker
│   ├── manager/            # State managers (IP ban, circuit breaker)
│   └── server/             # Server orchestration
├── pkg/logger/             # Logging utilities
└── configs/                # Configuration files
```

## 🐳 Docker Deployment

### Using Docker Compose

```bash
# Start the proxy
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the proxy
docker-compose down
```

### Manual Docker Run

```bash
# Build image
docker build -t dudu-proxy .

# Run container
docker run -d \
  -p 8080:8080 \
  -p 1080:1080 \
  -v $(pwd)/configs:/app/configs:ro \
  dudu-proxy
```

## 🔍 Monitoring

DuDu Proxy logs all important events in structured format:

- Authentication attempts (success/failure)
- IP bans and unbans
- Rate limit violations
- Circuit breaker state changes
- Proxy requests and responses

Example log output (JSON format):
```json
{
  "level":"info",
  "ts":"2024-01-01T00:00:00.000Z",
  "msg":"HTTPS tunnel established",
  "client_ip":"10.0.0.1",
  "target":"example.com:443"
}
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## ⚠️ Disclaimer

This proxy server is intended for legitimate use cases only. Users are responsible for ensuring compliance with applicable laws and regulations.

## 📧 Contact

- GitHub: [@seakee](https://github.com/seakee)
- Project Link: [https://github.com/seakee/dudu-proxy](https://github.com/seakee/dudu-proxy)
