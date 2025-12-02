package middleware

import (
	"fmt"
	"net"
)

// AuthMiddleware handles proxy authentication
type AuthMiddleware struct {
	enabled     bool
	credentials map[string]string // username -> password
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(enabled bool, credentials map[string]string) *AuthMiddleware {
	return &AuthMiddleware{
		enabled:     enabled,
		credentials: credentials,
	}
}

// Authenticate verifies the provided credentials
func (a *AuthMiddleware) Authenticate(username, password string) bool {
	if !a.enabled {
		return true // Authentication disabled
	}

	expectedPassword, exists := a.credentials[username]
	if !exists {
		return false
	}

	return expectedPassword == password
}

// IsEnabled returns whether authentication is enabled
func (a *AuthMiddleware) IsEnabled() bool {
	return a.enabled
}

// GetClientIP extracts the IP address from a network connection
func GetClientIP(conn net.Conn) string {
	if conn == nil {
		return ""
	}

	addr := conn.RemoteAddr()
	if addr == nil {
		return ""
	}

	// Extract IP from address (remove port)
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}

	return host
}

// ProxyAuthError represents an authentication error
type ProxyAuthError struct {
	Message string
}

func (e *ProxyAuthError) Error() string {
	return fmt.Sprintf("proxy authentication failed: %s", e.Message)
}

// NewProxyAuthError creates a new proxy authentication error
func NewProxyAuthError(message string) *ProxyAuthError {
	return &ProxyAuthError{Message: message}
}
