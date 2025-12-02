package manager

import (
	"testing"
	"time"
)

func TestCircuitBreaker_IsOpen(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 5, 2*time.Second)

	if cb.IsOpen() {
		t.Error("Circuit breaker should be closed initially")
	}
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 5, 2*time.Second)

	for i := 0; i < 10; i++ {
		cb.RecordSuccess()
	}

	if cb.IsOpen() {
		t.Error("Circuit breaker should not open on successes")
	}

	total, failures, failureRate := cb.GetStats()
	if total != 10 {
		t.Errorf("Expected 10 total requests, got %d", total)
	}
	if failures != 0 {
		t.Errorf("Expected 0 failures, got %d", failures)
	}
	if failureRate != 0 {
		t.Errorf("Expected 0%% failure rate, got %.2f%%", failureRate)
	}
}

func TestCircuitBreaker_RecordFailure(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 5, 500*time.Millisecond)

	// Record enough failures to open the circuit
	for i := 0; i < 3; i++ {
		cb.RecordSuccess()
	}
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if !cb.IsOpen() {
		t.Error("Circuit breaker should be open after high failure rate")
	}

	// Wait for recovery
	time.Sleep(600 * time.Millisecond)
	if cb.IsOpen() {
		t.Error("Circuit breaker should transition to half-open after timeout")
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 5, 500*time.Millisecond)

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.RecordSuccess()
	}
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if !cb.IsOpen() {
		t.Fatal("Circuit breaker should be open")
	}

	// Wait for transition to half-open
	time.Sleep(600 * time.Millisecond)

	// Record success in half-open state
	for i := 0; i < 3; i++ {
		cb.RecordSuccess()
	}

	state := cb.GetState()
	if state != StateClosed {
		t.Errorf("Circuit breaker should be closed after successes in half-open, got %s", state.String())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 5, 500*time.Millisecond)

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.RecordSuccess()
	}
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// Wait for transition to half-open
	time.Sleep(600 * time.Millisecond)

	// Record failure in half-open state
	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Error("Circuit breaker should go back to open after failure in half-open")
	}
}

func TestCircuitBreaker_GetState(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 5, 1*time.Second)

	if cb.GetState() != StateClosed {
		t.Error("Circuit breaker should be closed initially")
	}

	// Open the circuit
	for i := 0; i < 3; i++ {
		cb.RecordSuccess()
	}
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if cb.GetState() != StateOpen {
		t.Error("Circuit breaker should be open")
	}
}

func TestCircuitBreaker_MinRequests(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 10, 1*time.Second)

	// Record failures but below min requests
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if cb.IsOpen() {
		t.Error("Circuit breaker should not open below min requests")
	}
}

func TestCircuitBreaker_Call(t *testing.T) {
	cb := NewCircuitBreaker(50, 1*time.Second, 5, 500*time.Millisecond)

	// Successful calls
	for i := 0; i < 5; i++ {
		err := cb.Call(func() error { return nil })
		if err != nil {
			t.Errorf("Call should succeed: %v", err)
		}
	}

	// Failed calls to open circuit
	for i := 0; i < 5; i++ {
		err := cb.Call(func() error { return &CircuitBreakerError{} })
		if err == nil {
			t.Error("Call should return error")
		}
	}

	// Circuit should be open
	err := cb.Call(func() error { return nil })
	if err != ErrCircuitBreakerOpen {
		t.Errorf("Expected ErrCircuitBreakerOpen, got %v", err)
	}
}

// Benchmark tests
func BenchmarkCircuitBreaker_RecordSuccess(b *testing.B) {
	cb := NewCircuitBreaker(50, 1*time.Second, 10, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordSuccess()
	}
}

func BenchmarkCircuitBreaker_RecordFailure(b *testing.B) {
	cb := NewCircuitBreaker(50, 1*time.Second, 10, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordFailure()
	}
}

func BenchmarkCircuitBreaker_IsOpen(b *testing.B) {
	cb := NewCircuitBreaker(50, 1*time.Second, 10, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.IsOpen()
	}
}

func BenchmarkCircuitBreaker_GetStats(b *testing.B) {
	cb := NewCircuitBreaker(50, 1*time.Second, 10, 1*time.Second)
	for i := 0; i < 50; i++ {
		cb.RecordSuccess()
	}
	for i := 0; i < 50; i++ {
		cb.RecordFailure()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.GetStats()
	}
}
