package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	configContent := `{
		"server": {
			"http_port": 8080,
			"socks5_port": 1080
		},
		"auth": {
			"enabled": true,
			"users": [
				{"username": "user1", "password": "pass1"}
			]
		},
		"ip_ban": {
			"enabled": true,
			"max_failures": 3,
			"ban_duration_seconds": 300,
			"whitelist": ["127.0.0.1"]
		},
		"rate_limit": {
			"enabled": true,
			"global_requests_per_second": 1000,
			"per_ip_requests_per_second": 10
		},
		"circuit_breaker": {
			"enabled": true,
			"failure_threshold_percent": 50,
			"window_size_seconds": 60,
			"min_requests": 20,
			"break_duration_seconds": 30
		},
		"log": {
			"level": "info",
			"format": "json"
		}
	}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading the config
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if cfg.Server.HTTPPort != 8080 {
		t.Errorf("Expected HTTPPort 8080, got %d", cfg.Server.HTTPPort)
	}
	if cfg.Server.SOCKS5Port != 1080 {
		t.Errorf("Expected SOCKS5Port 1080, got %d", cfg.Server.SOCKS5Port)
	}
	if !cfg.Auth.Enabled {
		t.Error("Expected auth to be enabled")
	}
	if len(cfg.Auth.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(cfg.Auth.Users))
	}
	if cfg.Auth.Users[0].Username != "user1" {
		t.Errorf("Expected username 'user1', got '%s'", cfg.Auth.Users[0].Username)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{HTTPPort: 8080, SOCKS5Port: 1080},
				Auth:   AuthConfig{Enabled: true, Users: []User{{"user1", "pass1"}}},
				IPBan:  IPBanConfig{Enabled: true, MaxFailures: 3, BanDurationSeconds: 300},
				RateLimit: RateLimitConfig{
					Enabled:                 true,
					GlobalRequestsPerSecond: 1000,
					PerIPRequestsPerSecond:  10,
				},
				CircuitBreaker: CircuitBreakerConfig{
					Enabled:                 true,
					FailureThresholdPercent: 50,
					WindowSizeSeconds:       60,
					MinRequests:             20,
					BreakDurationSeconds:    30,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid http port",
			config: Config{
				Server: ServerConfig{HTTPPort: 0, SOCKS5Port: 1080},
			},
			wantErr: true,
		},
		{
			name: "invalid socks5 port",
			config: Config{
				Server: ServerConfig{HTTPPort: 8080, SOCKS5Port: 70000},
			},
			wantErr: true,
		},
		{
			name: "auth enabled with no users",
			config: Config{
				Server: ServerConfig{HTTPPort: 8080, SOCKS5Port: 1080},
				Auth:   AuthConfig{Enabled: true, Users: []User{}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUserCredentials(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			Enabled: true,
			Users: []User{
				{"user1", "pass1"},
				{"user2", "pass2"},
			},
		},
	}

	creds := cfg.GetUserCredentials()
	if len(creds) != 2 {
		t.Errorf("Expected 2 credentials, got %d", len(creds))
	}
	if creds["user1"] != "pass1" {
		t.Errorf("Expected password 'pass1' for user1, got '%s'", creds["user1"])
	}
	if creds["user2"] != "pass2" {
		t.Errorf("Expected password 'pass2' for user2, got '%s'", creds["user2"])
	}
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	configContent := `{
		"server": {"http_port": 8080, "socks5_port": 1080},
		"auth": {"enabled": true, "users": [{"username": "user1", "password": "pass1"}]},
		"ip_ban": {"enabled": true, "max_failures": 3, "ban_duration_seconds": 300, "whitelist": []},
		"rate_limit": {"enabled": true, "global_requests_per_second": 1000, "per_ip_requests_per_second": 10},
		"circuit_breaker": {"enabled": true, "failure_threshold_percent": 50, "window_size_seconds": 60, "min_requests": 20, "break_duration_seconds": 30},
		"log": {"level": "info", "format": "json"}
	}`

	tmpFile, _ := os.CreateTemp("", "config-*.json")
	tmpFile.Write([]byte(configContent))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load(tmpFile.Name())
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := Config{
		Server: ServerConfig{HTTPPort: 8080, SOCKS5Port: 1080},
		Auth:   AuthConfig{Enabled: true, Users: []User{{"user1", "pass1"}}},
		IPBan:  IPBanConfig{Enabled: true, MaxFailures: 3, BanDurationSeconds: 300},
		RateLimit: RateLimitConfig{
			Enabled:                 true,
			GlobalRequestsPerSecond: 1000,
			PerIPRequestsPerSecond:  10,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:                 true,
			FailureThresholdPercent: 50,
			WindowSizeSeconds:       60,
			MinRequests:             20,
			BreakDurationSeconds:    30,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}
