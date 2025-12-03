# Changelog

All notable changes to this project will be documented in this file.

## [v1.0.1] - 2025-12-03

### 🆕 New Features

- **Network Configuration**: Added a new `network` configuration option in the `server` section.
  - Supported values: `tcp` (default, dual-stack), `tcp4` (IPv4 only), `tcp6` (IPv6 only).
  - This allows users to explicitly restrict the server to listen on IPv4 or IPv6 interfaces, which is useful for environments with partial IPv6 support.

### 🐛 Bug Fixes

- **IPv6 Address Handling**: 
  - Fixed an issue where connecting to IPv6 targets via SOCKS5 proxy would fail with "too many colons in address" error.
  - Enforced configured network type (e.g., `tcp4`) for outbound connections to prevent "network is unreachable" errors on non-IPv6 systems.

## [v1.0.0] - 2025-12-02

### 🚀 Key Features

#### Core Proxy Services
- **HTTP/HTTPS Proxy**: Full support for HTTP proxying and HTTPS tunneling via the `CONNECT` method.
- **SOCKS5 Proxy**: Complete implementation of the SOCKS5 protocol (RFC 1928), including authentication and UDP association support.
- **Dual Stack**: Simultaneous operation of HTTP and SOCKS5 servers on configurable ports.

#### Security & Access Control
- **Authentication**: Configurable username/password authentication for both HTTP and SOCKS5.
- **IP Banning**: Automatic IP banning system that detects and blocks malicious actors after repeated authentication failures.
  - Configurable failure thresholds and ban duration.
  - Persistent ban records (saved to `data/ipban.json`).
  - Whitelist support.
- **Rate Limiting**: Robust Token Bucket based rate limiting to protect against abuse.
  - Global and per-IP limits.
- **Circuit Breaker**: Advanced circuit breaker pattern to protect the service during high load or attack scenarios.
  - Automatic recovery (half-open state).

#### Configuration & Observability
- **JSON Configuration**: Simple and flexible JSON-based configuration file.
- **Structured Logging**: High-performance structured logging using `sk-pkg/logger`.
- **Docker Support**: Ready-to-use `Dockerfile` and `docker-compose.yml` for easy deployment.

#### Build & Release
- **Cross-Platform Support**: Pre-built binaries for Linux, macOS, and Windows (amd64 and arm64).
- **Automated Releases**: GitHub Actions workflow for automated building and releasing.
- **Build Scripts**: Convenient Makefile targets and shell scripts for local cross-platform builds.
- **Version Management**: Automatic version injection from git tags.
- **Checksum Verification**: SHA256 checksums for all release binaries.
