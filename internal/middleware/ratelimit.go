package middleware

import (
	"sync"

	"golang.org/x/time/rate"
)

// RateLimitMiddleware handles request rate limiting
type RateLimitMiddleware struct {
	enabled       bool
	globalLimiter *rate.Limiter
	perIPLimiters map[string]*rate.Limiter
	perIPLimit    rate.Limit
	perIPBurst    int
	mu            sync.RWMutex
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(enabled bool, globalRPS, perIPRPS int) *RateLimitMiddleware {
	var globalLimiter *rate.Limiter
	if enabled && globalRPS > 0 {
		globalLimiter = rate.NewLimiter(rate.Limit(globalRPS), globalRPS*2)
	}

	return &RateLimitMiddleware{
		enabled:       enabled,
		globalLimiter: globalLimiter,
		perIPLimiters: make(map[string]*rate.Limiter),
		perIPLimit:    rate.Limit(perIPRPS),
		perIPBurst:    perIPRPS * 2,
	}
}

// Allow checks if a request from the given IP is allowed
func (r *RateLimitMiddleware) Allow(ip string) bool {
	if !r.enabled {
		return true
	}

	// Check global limit
	if r.globalLimiter != nil && !r.globalLimiter.Allow() {
		return false
	}

	// Check per-IP limit
	limiter := r.getIPLimiter(ip)
	return limiter.Allow()
}

// getIPLimiter returns the rate limiter for a specific IP
func (r *RateLimitMiddleware) getIPLimiter(ip string) *rate.Limiter {
	r.mu.RLock()
	limiter, exists := r.perIPLimiters[ip]
	r.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter for this IP
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	limiter, exists = r.perIPLimiters[ip]
	if exists {
		return limiter
	}

	limiter = rate.NewLimiter(r.perIPLimit, r.perIPBurst)
	r.perIPLimiters[ip] = limiter

	return limiter
}

// IsEnabled returns whether rate limiting is enabled
func (r *RateLimitMiddleware) IsEnabled() bool {
	return r.enabled
}
