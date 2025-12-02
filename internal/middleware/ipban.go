package middleware

import (
	"github.com/seakee/dudu-proxy/internal/manager"
)

// IPBanMiddleware handles IP banning
type IPBanMiddleware struct {
	enabled bool
	manager *manager.IPBanManager
}

// NewIPBanMiddleware creates a new IP ban middleware
func NewIPBanMiddleware(enabled bool, manager *manager.IPBanManager) *IPBanMiddleware {
	return &IPBanMiddleware{
		enabled: enabled,
		manager: manager,
	}
}

// IsBlocked checks if an IP is banned
func (i *IPBanMiddleware) IsBlocked(ip string) bool {
	if !i.enabled {
		return false
	}

	return i.manager.IsBanned(ip)
}

// RecordAuthFailure records an authentication failure for an IP
func (i *IPBanMiddleware) RecordAuthFailure(ip string) {
	if !i.enabled {
		return
	}

	i.manager.RecordFailure(ip)
}

// RecordAuthSuccess records a successful authentication for an IP
func (i *IPBanMiddleware) RecordAuthSuccess(ip string) {
	if !i.enabled {
		return
	}

	i.manager.RecordSuccess(ip)
}

// IsEnabled returns whether IP banning is enabled
func (i *IPBanMiddleware) IsEnabled() bool {
	return i.enabled
}
