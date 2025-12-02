package middleware

import (
	"testing"
)

func TestRateLimitMiddleware_Allow(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(true, 100, 10)

	// Test that requests are allowed initially
	for i := 0; i < 5; i++ {
		if !rateLimit.Allow("10.0.0.1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
}

func TestRateLimitMiddleware_Disabled(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(false, 1, 1)

	// All requests should be allowed when disabled
	for i := 0; i < 1000; i++ {
		if !rateLimit.Allow("10.0.0.1") {
			t.Error("Request should be allowed when rate limit is disabled")
		}
	}
}

func TestRateLimitMiddleware_PerIPLimit(t *testing.T) {
	rateLimit := NewRateLimitMiddleware(true, 1000, 5)

	// Each IP should have its own limiter
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	for _, ip := range ips {
		for i := 0; i < 5; i++ {
			if !rateLimit.Allow(ip) {
				t.Errorf("Request %d for IP %s should be allowed", i+1, ip)
			}
		}
	}
}

func TestRateLimitMiddleware_IsEnabled(t *testing.T) {
	enabled := NewRateLimitMiddleware(true, 100, 10)
	if !enabled.IsEnabled() {
		t.Error("Expected rate limit to be enabled")
	}

	disabled := NewRateLimitMiddleware(false, 100, 10)
	if disabled.IsEnabled() {
		t.Error("Expected rate limit to be disabled")
	}
}

// Benchmark tests
func BenchmarkRateLimitMiddleware_Allow(b *testing.B) {
	rateLimit := NewRateLimitMiddleware(true, 1000000, 1000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rateLimit.Allow("10.0.0.1")
	}
}

func BenchmarkRateLimitMiddleware_AllowMultipleIPs(b *testing.B) {
	rateLimit := NewRateLimitMiddleware(true, 1000000, 1000000)
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rateLimit.Allow(ips[i%len(ips)])
	}
}
