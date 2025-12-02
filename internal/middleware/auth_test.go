package middleware

import (
	"testing"
)

func TestAuthMiddleware_Authenticate(t *testing.T) {
	credentials := map[string]string{
		"user1": "pass1",
		"user2": "pass2",
	}

	tests := []struct {
		name     string
		enabled  bool
		username string
		password string
		want     bool
	}{
		{
			name:     "correct credentials",
			enabled:  true,
			username: "user1",
			password: "pass1",
			want:     true,
		},
		{
			name:     "wrong password",
			enabled:  true,
			username: "user1",
			password: "wrong",
			want:     false,
		},
		{
			name:     "non-existent user",
			enabled:  true,
			username: "user3",
			password: "pass3",
			want:     false,
		},
		{
			name:     "auth disabled",
			enabled:  false,
			username: "anyone",
			password: "anything",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAuthMiddleware(tt.enabled, credentials)
			if got := auth.Authenticate(tt.username, tt.password); got != tt.want {
				t.Errorf("Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthMiddleware_IsEnabled(t *testing.T) {
	auth1 := NewAuthMiddleware(true, map[string]string{})
	if !auth1.IsEnabled() {
		t.Error("Expected auth to be enabled")
	}

	auth2 := NewAuthMiddleware(false, map[string]string{})
	if auth2.IsEnabled() {
		t.Error("Expected auth to be disabled")
	}
}

// Benchmark tests
func BenchmarkAuthMiddleware_Authenticate(b *testing.B) {
	credentials := map[string]string{
		"user1": "pass1",
		"user2": "pass2",
	}
	auth := NewAuthMiddleware(true, credentials)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.Authenticate("user1", "pass1")
	}
}
