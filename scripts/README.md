# Verification Scripts

This directory contains automated testing and verification scripts for DuDu Proxy.

## verify.sh

Automated verification script that tests all core functionality:

### Features Tested

1. **HTTP Proxy** - Basic HTTP proxy functionality with authentication
2. **HTTPS Proxy** - CONNECT tunnel for HTTPS traffic
3. **SOCKS5 Proxy** - Full SOCKS5 protocol support
4. **Authentication** - User authentication and rejection of invalid credentials
5. **IP Ban** - Automatic IP banning after failed authentication attempts
6. **Rate Limiting** - Request rate limiting functionality
7. **Circuit Breaker** - Circuit breaker triggering on high failure rates
8. **Configuration** - Configuration file validation
9. **Logging** - Log generation and content verification
10. **Graceful Shutdown** - Signal handling and clean shutdown

### Usage

```bash
# Make sure the proxy is built first
make build

# Run the verification script
./scripts/verify.sh
```

### Output

The script provides colored output with:
- 🟦 Test headers
- 🟨 Test descriptions
- 🟢 Pass status
- 🔴 Fail status
- 🔵 Informational messages

### Test Results

At the end, a summary is displayed showing:
- Total number of tests run
- Number of tests passed
- Number of tests failed

Exit code: 
- `0` if all tests pass
- `1` if any test fails

### Requirements

- `curl` command (for HTTP/HTTPS/SOCKS5 testing)
- `python3` (for JSON validation)
- Running on macOS or Linux
