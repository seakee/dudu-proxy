package middleware

import (
	"github.com/seakee/dudu-proxy/internal/manager"
)

// CircuitBreakerMiddleware handles circuit breaking
type CircuitBreakerMiddleware struct {
	enabled bool
	breaker *manager.CircuitBreaker
}

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware
func NewCircuitBreakerMiddleware(enabled bool, breaker *manager.CircuitBreaker) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		enabled: enabled,
		breaker: breaker,
	}
}

// IsOpen checks if the circuit breaker is open
func (c *CircuitBreakerMiddleware) IsOpen() bool {
	if !c.enabled {
		return false
	}

	return c.breaker.IsOpen()
}

// RecordAuthFailure records an authentication failure
func (c *CircuitBreakerMiddleware) RecordAuthFailure() {
	if !c.enabled {
		return
	}

	c.breaker.RecordFailure()
}

// RecordAuthSuccess records a successful authentication
func (c *CircuitBreakerMiddleware) RecordAuthSuccess() {
	if !c.enabled {
		return
	}

	c.breaker.RecordSuccess()
}

// GetState returns the current state of the circuit breaker
func (c *CircuitBreakerMiddleware) GetState() manager.CircuitBreakerState {
	if !c.enabled {
		return manager.StateClosed
	}

	return c.breaker.GetState()
}

// IsEnabled returns whether circuit breaking is enabled
func (c *CircuitBreakerMiddleware) IsEnabled() bool {
	return c.enabled
}
