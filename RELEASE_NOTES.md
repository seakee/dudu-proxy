# Release v1.0.0

We are excited to announce the first stable release of **DuDu Proxy**, a high-performance, feature-rich proxy server written in Go.

## 🚀 Key Features

### Core Proxy Services
- **HTTP/HTTPS Proxy**: Full support for HTTP proxying and HTTPS tunneling via the `CONNECT` method.
- **SOCKS5 Proxy**: Complete implementation of the SOCKS5 protocol (RFC 1928), including authentication and UDP association support.
- **Dual Stack**: Simultaneous operation of HTTP and SOCKS5 servers on configurable ports.

### Security & Access Control
- **Authentication**: Configurable username/password authentication for both HTTP and SOCKS5.
- **IP Banning**: Automatic IP banning system that detects and blocks malicious actors after repeated authentication failures.
  - Configurable failure thresholds and ban duration.
  - Persistent ban records (saved to `data/ipban.json`).
  - Whitelist support.
- **Rate Limiting**: Robust Token Bucket based rate limiting to protect against abuse.
  - Global and per-IP limits.
- **Circuit Breaker**: Advanced circuit breaker pattern to protect the service during high load or attack scenarios.
  - Automatic recovery (half-open state).

### Configuration & Observability
- **JSON Configuration**: Simple and flexible JSON-based configuration file.
- **Structured Logging**: High-performance structured logging using `sk-pkg/logger`.
- **Docker Support**: Ready-to-use `Dockerfile` and `docker-compose.yml` for easy deployment.

### Build & Release
- **Cross-Platform Support**: Pre-built binaries for Linux, macOS, and Windows (amd64 and arm64).
- **Automated Releases**: GitHub Actions workflow for automated building and releasing.
- **Build Scripts**: Convenient Makefile targets and shell scripts for local cross-platform builds.
- **Version Management**: Automatic version injection from git tags.
- **Checksum Verification**: SHA256 checksums for all release binaries.

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
git clone https://github.com/seakee/dudu-proxy.git
cd dudu-proxy
make build
```

Or build for all platforms:
```bash
make build-all
```

### Using Docker
```bash
docker-compose up -d
```

## 🛠 Configuration

Copy the example configuration and adjust it to your needs:
```bash
cp configs/config.example.json configs/config.json
# Edit configs/config.json
./dudu-proxy -config configs/config.json
```

## 📝 Changelog

### Core Features
- Initial release of DuDu Proxy.
- Implemented core HTTP and SOCKS5 proxy logic.
- Added middleware chain: Auth -> IPBan -> RateLimit -> CircuitBreaker.
- Integrated `sk-pkg/logger` for structured logging.
- Added automated verification scripts (`scripts/verify.sh`).
- Comprehensive documentation in English and Chinese.

### Build & Release Infrastructure
- Added cross-platform build support with Makefile targets (`build-linux`, `build-darwin`, `build-windows`, `build-all`).
- Added automated release workflow with GitHub Actions (`.github/workflows/release.yml`).
- Added build script (`scripts/build.sh`) for local cross-platform builds.
- Implemented version management with git tags and automatic version injection.
- Added SHA256 checksum generation for all binaries.
- Pre-built binaries for Linux, macOS, and Windows (amd64 and arm64 architectures).
- Updated documentation with installation instructions for pre-built binaries.

---
*Thank you for using DuDu Proxy!*
