package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/seakee/dudu-proxy/internal/middleware"
	"github.com/seakee/dudu-proxy/pkg/logger"
)

const (
	socks5Version = 0x05

	// Authentication methods
	authNone     = 0x00
	authPassword = 0x02
	authNoAccept = 0xFF

	// Commands
	cmdConnect = 0x01

	// Address types
	atypIPv4   = 0x01
	atypDomain = 0x03
	atypIPv6   = 0x04

	// Reply codes
	repSuccess              = 0x00
	repServerFailure        = 0x01
	repConnectionNotAllowed = 0x02
	repNetworkUnreachable   = 0x03
	repHostUnreachable      = 0x04
	repConnectionRefused    = 0x05
	repTTLExpired           = 0x06
	repCommandNotSupported  = 0x07
	repAddressNotSupported  = 0x08
)

// SOCKS5Proxy represents a SOCKS5 proxy server
type SOCKS5Proxy struct {
	port           int
	auth           *middleware.AuthMiddleware
	rateLimit      *middleware.RateLimitMiddleware
	ipBan          *middleware.IPBanMiddleware
	circuitBreaker *middleware.CircuitBreakerMiddleware
}

// NewSOCKS5Proxy creates a new SOCKS5 proxy
func NewSOCKS5Proxy(
	port int,
	auth *middleware.AuthMiddleware,
	rateLimit *middleware.RateLimitMiddleware,
	ipBan *middleware.IPBanMiddleware,
	circuitBreaker *middleware.CircuitBreakerMiddleware,
) *SOCKS5Proxy {
	return &SOCKS5Proxy{
		port:           port,
		auth:           auth,
		rateLimit:      rateLimit,
		ipBan:          ipBan,
		circuitBreaker: circuitBreaker,
	}
}

// Start starts the SOCKS5 proxy server
func (s *SOCKS5Proxy) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to start SOCKS5 proxy: %w", err)
	}

	logger.Info("SOCKS5 proxy server started", "port", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Failed to accept connection", "error", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a single SOCKS5 connection
func (s *SOCKS5Proxy) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	clientIP := middleware.GetClientIP(clientConn)

	// Check circuit breaker
	if s.circuitBreaker.IsOpen() {
		logger.Warn("SOCKS5 request rejected: circuit breaker is open",
			"client_ip", clientIP,
			"circuit_state", s.circuitBreaker.GetState().String())
		return
	}

	// Check IP ban
	if s.ipBan.IsBlocked(clientIP) {
		logger.Warn("SOCKS5 request rejected: IP is banned", "client_ip", clientIP)
		return
	}

	// Check rate limit
	if !s.rateLimit.Allow(clientIP) {
		logger.Warn("SOCKS5 request rejected: rate limit exceeded", "client_ip", clientIP)
		return
	}

	// SOCKS5 handshake
	if err := s.handshake(clientConn, clientIP); err != nil {
		logger.Error("SOCKS5 handshake failed", "client_ip", clientIP, "error", err)
		return
	}

	// Handle the request
	if err := s.handleRequest(clientConn, clientIP); err != nil {
		logger.Error("Failed to handle SOCKS5 request", "client_ip", clientIP, "error", err)
		return
	}
}

// handshake performs the SOCKS5 handshake
func (s *SOCKS5Proxy) handshake(conn net.Conn, clientIP string) error {
	// Read version and methods
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	version := buf[0]
	nMethods := buf[1]

	if version != socks5Version {
		return fmt.Errorf("unsupported SOCKS version: %d", version)
	}

	// Read methods
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("failed to read methods: %w", err)
	}

	// Determine authentication method
	selectedMethod := authNoAccept
	if s.auth.IsEnabled() {
		// Check if client supports password authentication
		for _, method := range methods {
			if method == authPassword {
				selectedMethod = authPassword
				break
			}
		}
	} else {
		// Check if client supports no authentication
		for _, method := range methods {
			if method == authNone {
				selectedMethod = authNone
				break
			}
		}
	}

	// Send selected method
	if _, err := conn.Write([]byte{socks5Version, byte(selectedMethod)}); err != nil {
		return fmt.Errorf("failed to send method selection: %w", err)
	}

	if selectedMethod == authNoAccept {
		return fmt.Errorf("no acceptable authentication method")
	}

	// Perform authentication if required
	if selectedMethod == authPassword {
		if err := s.authenticatePassword(conn, clientIP); err != nil {
			return err
		}
	}

	return nil
}

// authenticatePassword performs username/password authentication
func (s *SOCKS5Proxy) authenticatePassword(conn net.Conn, clientIP string) error {
	// Read authentication request
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("failed to read auth version: %w", err)
	}

	authVersion := buf[0]
	if authVersion != 0x01 {
		return fmt.Errorf("unsupported auth version: %d", authVersion)
	}

	// Read username
	usernameLen := int(buf[1])
	username := make([]byte, usernameLen)
	if _, err := io.ReadFull(conn, username); err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}

	// Read password length
	passwordLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, passwordLenBuf); err != nil {
		return fmt.Errorf("failed to read password length: %w", err)
	}

	// Read password
	passwordLen := int(passwordLenBuf[0])
	password := make([]byte, passwordLen)
	if _, err := io.ReadFull(conn, password); err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Authenticate
	authSuccess := s.auth.Authenticate(string(username), string(password))

	// Send authentication response
	var status byte
	if authSuccess {
		status = 0x00
		s.ipBan.RecordAuthSuccess(clientIP)
		s.circuitBreaker.RecordAuthSuccess()

		logger.Debug("SOCKS5 authentication successful",
			"client_ip", clientIP,
			"username", string(username))
	} else {
		status = 0x01
		s.ipBan.RecordAuthFailure(clientIP)
		s.circuitBreaker.RecordAuthFailure()

		logger.Warn("SOCKS5 authentication failed",
			"client_ip", clientIP,
			"username", string(username))
	}

	if _, err := conn.Write([]byte{0x01, status}); err != nil {
		return fmt.Errorf("failed to send auth response: %w", err)
	}

	if !authSuccess {
		return fmt.Errorf("authentication failed")
	}

	return nil
}

// handleRequest handles the SOCKS5 request
func (s *SOCKS5Proxy) handleRequest(clientConn net.Conn, clientIP string) error {
	// Read request header
	buf := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, buf); err != nil {
		return fmt.Errorf("failed to read request: %w", err)
	}

	version := buf[0]
	cmd := buf[1]
	// buf[2] is reserved
	atyp := buf[3]

	if version != socks5Version {
		s.sendReply(clientConn, repServerFailure, atyp)
		return fmt.Errorf("invalid version: %d", version)
	}

	if cmd != cmdConnect {
		s.sendReply(clientConn, repCommandNotSupported, atyp)
		return fmt.Errorf("unsupported command: %d", cmd)
	}

	// Read target address
	var targetAddr string
	switch atyp {
	case atypIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(clientConn, addr); err != nil {
			s.sendReply(clientConn, repServerFailure, atyp)
			return fmt.Errorf("failed to read IPv4 address: %w", err)
		}
		targetAddr = net.IPv4(addr[0], addr[1], addr[2], addr[3]).String()

	case atypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(clientConn, lenBuf); err != nil {
			s.sendReply(clientConn, repServerFailure, atyp)
			return fmt.Errorf("failed to read domain length: %w", err)
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(clientConn, domain); err != nil {
			s.sendReply(clientConn, repServerFailure, atyp)
			return fmt.Errorf("failed to read domain: %w", err)
		}
		targetAddr = string(domain)

	case atypIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(clientConn, addr); err != nil {
			s.sendReply(clientConn, repServerFailure, atyp)
			return fmt.Errorf("failed to read IPv6 address: %w", err)
		}
		targetAddr = net.IP(addr).String()

	default:
		s.sendReply(clientConn, repAddressNotSupported, atyp)
		return fmt.Errorf("unsupported address type: %d", atyp)
	}

	// Read port
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, portBuf); err != nil {
		s.sendReply(clientConn, repServerFailure, atyp)
		return fmt.Errorf("failed to read port: %w", err)
	}
	targetPort := binary.BigEndian.Uint16(portBuf)

	target := fmt.Sprintf("%s:%d", targetAddr, targetPort)

	// Connect to target
	targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		logger.Error("Failed to connect to target",
			"client_ip", clientIP,
			"target", target,
			"error", err)
		s.sendReply(clientConn, repHostUnreachable, atyp)
		return fmt.Errorf("failed to connect to target: %w", err)
	}
	defer targetConn.Close()

	// Send success reply
	s.sendReply(clientConn, repSuccess, atyp)

	logger.Info("SOCKS5 connection established",
		"client_ip", clientIP,
		"target", target)

	// Bidirectional copy
	s.transfer(clientConn, targetConn)

	return nil
}

// sendReply sends a SOCKS5 reply
func (s *SOCKS5Proxy) sendReply(conn net.Conn, rep byte, atyp byte) {
	reply := []byte{
		socks5Version,
		rep,
		0x00,       // Reserved
		0x01,       // IPv4
		0, 0, 0, 0, // Bind address
		0, 0, // Bind port
	}
	conn.Write(reply)
}

// transfer bidirectionally copies data between two connections
func (s *SOCKS5Proxy) transfer(conn1, conn2 net.Conn) {
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
