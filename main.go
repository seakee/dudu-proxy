package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/seakee/dudu-proxy/internal/config"
	"github.com/seakee/dudu-proxy/internal/server"
	"github.com/seakee/dudu-proxy/pkg/logger"
)

var (
	configFile = flag.String("config", "configs/config.example.json", "Path to configuration file")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	// Print banner
	printBanner()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Log.Level, cfg.Log.Format)

	logger.Info("Starting DuDu Proxy",
		"version", version,
		"config_file", *configFile)

	// Log configuration summary
	logConfigSummary(cfg)

	// Create and run server
	srv := server.NewServer(cfg)
	if err := srv.Run(); err != nil {
		logger.Fatal("Server failed", "error", err)
	}
}

func printBanner() {
	banner := `
 ____        ____        ____                      
|  _ \  _   |  _ \  _   |  _ \ _ __ _____  ___   _
| | | || | || | | || | || |_) | '__/ _ \ \/ / | | |
| |_| || |_|| |_| || |_|| |  _| | | (_) >  <| |_| |
|____/  \__||____/  \__||_| |_|_|  \___/_/\_\\__, |
                                             |___/ 
DuDu Proxy - High-Performance Proxy Server
Version: %s
`
	fmt.Printf(banner, version)
	fmt.Println()
}

func logConfigSummary(cfg *config.Config) {
	logger.Info("Server configuration",
		"http_port", cfg.Server.HTTPPort,
		"socks5_port", cfg.Server.SOCKS5Port,
		"auth_enabled", cfg.Auth.Enabled,
		"auth_users", len(cfg.Auth.Users))

	logger.Info("IP ban configuration",
		"ip_ban_enabled", cfg.IPBan.Enabled,
		"max_failures", cfg.IPBan.MaxFailures,
		"ban_duration_seconds", cfg.IPBan.BanDurationSeconds,
		"whitelist_count", len(cfg.IPBan.Whitelist))

	logger.Info("Rate limit configuration",
		"rate_limit_enabled", cfg.RateLimit.Enabled,
		"global_rps", cfg.RateLimit.GlobalRequestsPerSecond,
		"per_ip_rps", cfg.RateLimit.PerIPRequestsPerSecond)

	logger.Info("Circuit breaker configuration",
		"circuit_breaker_enabled", cfg.CircuitBreaker.Enabled,
		"failure_threshold_percent", cfg.CircuitBreaker.FailureThresholdPercent,
		"window_size_seconds", cfg.CircuitBreaker.WindowSizeSeconds,
		"min_requests", cfg.CircuitBreaker.MinRequests,
		"break_duration_seconds", cfg.CircuitBreaker.BreakDurationSeconds)
}
