package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	Server         ServerConfig         `json:"server"`
	Auth           AuthConfig           `json:"auth"`
	IPBan          IPBanConfig          `json:"ip_ban"`
	RateLimit      RateLimitConfig      `json:"rate_limit"`
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`
	Log            LogConfig            `json:"log"`
}

// ServerConfig contains server-related settings
type ServerConfig struct {
	HTTPPort   int `json:"http_port"`
	SOCKS5Port int `json:"socks5_port"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Enabled bool   `json:"enabled"`
	Users   []User `json:"users"`
}

// User represents a proxy user
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// IPBanConfig contains IP ban settings
type IPBanConfig struct {
	Enabled            bool     `json:"enabled"`
	MaxFailures        int      `json:"max_failures"`
	BanDurationSeconds int      `json:"ban_duration_seconds"`
	Whitelist          []string `json:"whitelist"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled                 bool `json:"enabled"`
	GlobalRequestsPerSecond int  `json:"global_requests_per_second"`
	PerIPRequestsPerSecond  int  `json:"per_ip_requests_per_second"`
}

// CircuitBreakerConfig contains circuit breaker settings
type CircuitBreakerConfig struct {
	Enabled                 bool `json:"enabled"`
	FailureThresholdPercent int  `json:"failure_threshold_percent"`
	WindowSizeSeconds       int  `json:"window_size_seconds"`
	MinRequests             int  `json:"min_requests"`
	BreakDurationSeconds    int  `json:"break_duration_seconds"`
}

// LogConfig contains logging settings
type LogConfig struct {
	Level  string `json:"level"`
	Driver string `json:"driver"`
	Path   string `json:"path"`
}

// Load reads and parses the configuration file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.HTTPPort <= 0 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.Server.HTTPPort)
	}
	if c.Server.SOCKS5Port <= 0 || c.Server.SOCKS5Port > 65535 {
		return fmt.Errorf("invalid SOCKS5 port: %d", c.Server.SOCKS5Port)
	}

	if c.Auth.Enabled && len(c.Auth.Users) == 0 {
		return fmt.Errorf("authentication is enabled but no users are configured")
	}

	if c.IPBan.Enabled && c.IPBan.MaxFailures <= 0 {
		return fmt.Errorf("max_failures must be positive when IP ban is enabled")
	}

	if c.IPBan.Enabled && c.IPBan.BanDurationSeconds <= 0 {
		return fmt.Errorf("ban_duration_seconds must be positive when IP ban is enabled")
	}

	if c.RateLimit.Enabled {
		if c.RateLimit.GlobalRequestsPerSecond <= 0 {
			return fmt.Errorf("global_requests_per_second must be positive when rate limit is enabled")
		}
		if c.RateLimit.PerIPRequestsPerSecond <= 0 {
			return fmt.Errorf("per_ip_requests_per_second must be positive when rate limit is enabled")
		}
	}

	if c.CircuitBreaker.Enabled {
		if c.CircuitBreaker.FailureThresholdPercent <= 0 || c.CircuitBreaker.FailureThresholdPercent > 100 {
			return fmt.Errorf("failure_threshold_percent must be between 1 and 100")
		}
		if c.CircuitBreaker.WindowSizeSeconds <= 0 {
			return fmt.Errorf("window_size_seconds must be positive")
		}
		if c.CircuitBreaker.MinRequests <= 0 {
			return fmt.Errorf("min_requests must be positive")
		}
		if c.CircuitBreaker.BreakDurationSeconds <= 0 {
			return fmt.Errorf("break_duration_seconds must be positive")
		}
	}

	return nil
}

// GetUserCredentials returns a map of username to password for quick lookup
func (c *Config) GetUserCredentials() map[string]string {
	credentials := make(map[string]string)
	for _, user := range c.Auth.Users {
		credentials[user.Username] = user.Password
	}
	return credentials
}
