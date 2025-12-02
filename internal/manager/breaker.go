package manager

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	// StateClosed means the circuit is closed and requests are allowed
	StateClosed CircuitBreakerState = iota
	// StateOpen means the circuit is open and requests are rejected
	StateOpen
	// StateHalfOpen means the circuit is testing if it can close again
	StateHalfOpen
)

// String returns the string representation of the state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements a sliding window circuit breaker
type CircuitBreaker struct {
	mu                   sync.RWMutex
	state                CircuitBreakerState
	failureThreshold     float64 // Percentage (0-100)
	windowSize           time.Duration
	minRequests          int
	breakDuration        time.Duration
	requests             []requestRecord
	lastStateChange      time.Time
	consecutiveSuccesses int
	halfOpenMaxRequests  int
}

type requestRecord struct {
	timestamp time.Time
	success   bool
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThresholdPercent int, windowSize time.Duration, minRequests int, breakDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:               StateClosed,
		failureThreshold:    float64(failureThresholdPercent),
		windowSize:          windowSize,
		minRequests:         minRequests,
		breakDuration:       breakDuration,
		requests:            make([]requestRecord, 0),
		lastStateChange:     time.Now(),
		halfOpenMaxRequests: 3,
	}
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// If in open state, check if we should transition to half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) >= cb.breakDuration {
			return false // Allow transition to half-open
		}
		return true
	}

	return false
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Check if we should transition from open to half-open
	if cb.state == StateOpen && time.Since(cb.lastStateChange) >= cb.breakDuration {
		return StateHalfOpen
	}

	return cb.state
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.requests = append(cb.requests, requestRecord{timestamp: now, success: true})

	// Handle half-open state
	if cb.state == StateHalfOpen {
		cb.consecutiveSuccesses++
		if cb.consecutiveSuccesses >= cb.halfOpenMaxRequests {
			cb.state = StateClosed
			cb.lastStateChange = now
			cb.consecutiveSuccesses = 0
		}
	}

	cb.cleanup(now)
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.requests = append(cb.requests, requestRecord{timestamp: now, success: false})

	// If in half-open state, immediately go back to open on failure
	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.lastStateChange = now
		cb.consecutiveSuccesses = 0
		cb.cleanup(now)
		return
	}

	// Check if we should open the circuit
	cb.cleanup(now)
	if cb.shouldOpen() {
		cb.state = StateOpen
		cb.lastStateChange = now
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	// Check state with potential transition to half-open
	currentState := cb.GetState()

	if currentState == StateOpen {
		return ErrCircuitBreakerOpen
	}

	// If half-open, transition to that state
	if currentState == StateHalfOpen {
		cb.mu.Lock()
		if cb.state == StateOpen && time.Since(cb.lastStateChange) >= cb.breakDuration {
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
		}
		cb.mu.Unlock()
	}

	err := fn()
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// shouldOpen determines if the circuit should be opened based on recent requests
func (cb *CircuitBreaker) shouldOpen() bool {
	if cb.state != StateClosed {
		return false
	}

	if len(cb.requests) < cb.minRequests {
		return false
	}

	failures := 0
	for _, req := range cb.requests {
		if !req.success {
			failures++
		}
	}

	failurePercent := float64(failures) * 100.0 / float64(len(cb.requests))
	return failurePercent >= cb.failureThreshold
}

// cleanup removes requests outside the time window
func (cb *CircuitBreaker) cleanup(now time.Time) {
	cutoff := now.Add(-cb.windowSize)
	validRequests := make([]requestRecord, 0, len(cb.requests))

	for _, req := range cb.requests {
		if req.timestamp.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}

	cb.requests = validRequests
}

// GetStats returns the current statistics
func (cb *CircuitBreaker) GetStats() (total, failures int, failureRate float64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	total = len(cb.requests)
	if total == 0 {
		return 0, 0, 0
	}

	for _, req := range cb.requests {
		if !req.success {
			failures++
		}
	}

	failureRate = float64(failures) * 100.0 / float64(total)
	return
}

// ErrCircuitBreakerOpen is returned when the circuit breaker is open
var ErrCircuitBreakerOpen = &CircuitBreakerError{}

// CircuitBreakerError represents a circuit breaker error
type CircuitBreakerError struct{}

func (e *CircuitBreakerError) Error() string {
	return "circuit breaker is open"
}
