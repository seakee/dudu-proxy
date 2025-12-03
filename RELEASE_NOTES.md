# Release v1.0.1

## 🆕 New Features

- **Network Configuration**: Added a new `network` configuration option in the `server` section.
  - Supported values: `tcp` (default, dual-stack), `tcp4` (IPv4 only), `tcp6` (IPv6 only).
  - This allows users to explicitly restrict the server to listen on IPv4 or IPv6 interfaces, which is useful for environments with partial IPv6 support.

## 🐛 Bug Fixes

- **IPv6 Address Handling**: Fixed an issue where connecting to IPv6 targets via SOCKS5 proxy would fail with "too many colons in address" error.
