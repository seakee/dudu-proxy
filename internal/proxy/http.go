package proxy

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/seakee/dudu-proxy/internal/middleware"
	"github.com/seakee/dudu-proxy/pkg/logger"
)

// HTTPProxy represents an HTTP proxy server
type HTTPProxy struct {
	port           int
	network        string // 网络类型: "tcp", "tcp4", "tcp6"
	auth           *middleware.AuthMiddleware
	rateLimit      *middleware.RateLimitMiddleware
	ipBan          *middleware.IPBanMiddleware
	circuitBreaker *middleware.CircuitBreakerMiddleware
}

// NewHTTPProxy creates a new HTTP proxy
func NewHTTPProxy(
	port int,
	network string,
	auth *middleware.AuthMiddleware,
	rateLimit *middleware.RateLimitMiddleware,
	ipBan *middleware.IPBanMiddleware,
	circuitBreaker *middleware.CircuitBreakerMiddleware,
) *HTTPProxy {
	return &HTTPProxy{
		port:           port,
		network:        network,
		auth:           auth,
		rateLimit:      rateLimit,
		ipBan:          ipBan,
		circuitBreaker: circuitBreaker,
	}
}

// Start starts the HTTP proxy server
func (h *HTTPProxy) Start() error {
	listener, err := net.Listen(h.network, fmt.Sprintf(":%d", h.port))
	if err != nil {
		return fmt.Errorf("failed to start HTTP proxy: %w", err)
	}

	logger.Info("HTTP proxy server started", "port", h.port, "network", h.network)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Failed to accept connection", "error", err)
			continue
		}

		go h.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (h *HTTPProxy) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	clientIP := middleware.GetClientIP(clientConn)

	// Check circuit breaker
	if h.circuitBreaker.IsOpen() {
		logger.Warn("Request rejected: circuit breaker is open",
			"client_ip", clientIP,
			"circuit_state", h.circuitBreaker.GetState().String())
		h.sendError(clientConn, http.StatusServiceUnavailable, "Service temporarily unavailable")
		return
	}

	// Check IP ban
	if h.ipBan.IsBlocked(clientIP) {
		logger.Warn("Request rejected: IP is banned", "client_ip", clientIP)
		h.sendError(clientConn, http.StatusForbidden, "Access denied")
		return
	}

	// Check rate limit
	if !h.rateLimit.Allow(clientIP) {
		logger.Warn("Request rejected: rate limit exceeded", "client_ip", clientIP)
		h.sendError(clientConn, http.StatusTooManyRequests, "Too many requests")
		return
	}

	// Read the request
	reader := bufio.NewReader(clientConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		logger.Error("Failed to read request", "client_ip", clientIP, "error", err)
		return
	}

	// Handle authentication
	if h.auth.IsEnabled() {
		username, password, ok := h.parseProxyAuth(req)
		if !ok || !h.auth.Authenticate(username, password) {
			logger.Warn("Authentication failed",
				"client_ip", clientIP,
				"username", username)

			h.ipBan.RecordAuthFailure(clientIP)
			h.circuitBreaker.RecordAuthFailure()
			h.sendProxyAuthRequired(clientConn)
			return
		}

		logger.Debug("Authentication successful",
			"client_ip", clientIP,
			"username", username)

		h.ipBan.RecordAuthSuccess(clientIP)
		h.circuitBreaker.RecordAuthSuccess()
	}

	// Handle CONNECT method (for HTTPS)
	if req.Method == http.MethodConnect {
		h.handleConnect(clientConn, req, clientIP)
	} else {
		// Handle regular HTTP request
		h.handleHTTP(clientConn, req, clientIP)
	}
}

// handleConnect handles HTTPS CONNECT requests
func (h *HTTPProxy) handleConnect(clientConn net.Conn, req *http.Request, clientIP string) {
	// Connect to the target server
	targetConn, err := net.DialTimeout(h.network, req.Host, 10*time.Second)
	if err != nil {
		logger.Error("Failed to connect to target",
			"client_ip", clientIP,
			"target", req.Host,
			"error", err)
		h.sendError(clientConn, http.StatusBadGateway, "Failed to connect to target")
		return
	}
	defer targetConn.Close()

	// Send 200 Connection Established
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		logger.Error("Failed to send response", "client_ip", clientIP, "error", err)
		return
	}

	logger.Info("HTTPS tunnel established",
		"client_ip", clientIP,
		"target", req.Host)

	// Bidirectional copy
	h.transfer(clientConn, targetConn)
}

// handleHTTP handles regular HTTP requests
func (h *HTTPProxy) handleHTTP(clientConn net.Conn, req *http.Request, clientIP string) {
	// Remove proxy-specific headers
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Proxy-Connection")

	// Determine target address
	// For HTTP requests, req.Host might not include port, we need to add default port 80
	targetAddr := req.Host
	if !strings.Contains(targetAddr, ":") {
		targetAddr = net.JoinHostPort(targetAddr, "80")
	}

	// Connect to the target server
	targetConn, err := net.DialTimeout(h.network, targetAddr, 10*time.Second)
	if err != nil {
		logger.Error("Failed to connect to target",
			"client_ip", clientIP,
			"target", targetAddr,
			"error", err)
		h.sendError(clientConn, http.StatusBadGateway, "Failed to connect to target")
		return
	}
	defer targetConn.Close()

	// Write the request to the target
	if err := req.Write(targetConn); err != nil {
		logger.Error("Failed to send request to target",
			"client_ip", clientIP,
			"target", targetAddr,
			"error", err)
		return
	}

	logger.Info("HTTP request proxied",
		"client_ip", clientIP,
		"method", req.Method,
		"url", req.URL.String())

	// Copy response back to client
	_, err = io.Copy(clientConn, targetConn)
	if err != nil && err != io.EOF {
		logger.Debug("Error copying response",
			"client_ip", clientIP,
			"error", err)
	}
}

// transfer bidirectionally copies data between two connections
func (h *HTTPProxy) transfer(conn1, conn2 net.Conn) {
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(conn1, conn2)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(conn2, conn1)
		done <- struct{}{}
	}()

	<-done
}

// parseProxyAuth parses the Proxy-Authorization header
func (h *HTTPProxy) parseProxyAuth(req *http.Request) (username, password string, ok bool) {
	auth := req.Header.Get("Proxy-Authorization")
	if auth == "" {
		return "", "", false
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return "", "", false
	}

	return credentials[0], credentials[1], true
}

// sendProxyAuthRequired sends a 407 Proxy Authentication Required response
func (h *HTTPProxy) sendProxyAuthRequired(conn net.Conn) {
	response := "HTTP/1.1 407 Proxy Authentication Required\r\n" +
		"Proxy-Authenticate: Basic realm=\"DuDu Proxy\"\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n"
	conn.Write([]byte(response))
}

// sendError sends an error response
func (h *HTTPProxy) sendError(conn net.Conn, statusCode int, message string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n"+
		"Content-Type: text/plain\r\n"+
		"Content-Length: %d\r\n"+
		"\r\n"+
		"%s",
		statusCode, http.StatusText(statusCode), len(message), message)
	conn.Write([]byte(response))
}
