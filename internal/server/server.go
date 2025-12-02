package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seakee/dudu-proxy/internal/config"
	"github.com/seakee/dudu-proxy/internal/manager"
	"github.com/seakee/dudu-proxy/internal/middleware"
	"github.com/seakee/dudu-proxy/internal/proxy"
	"github.com/seakee/dudu-proxy/pkg/logger"
)

// Server represents the proxy server
type Server struct {
	config      *config.Config
	httpProxy   *proxy.HTTPProxy
	socks5Proxy *proxy.SOCKS5Proxy
	ipBanMgr    *manager.IPBanManager
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	// Create managers
	ipBanMgr := manager.NewIPBanManager(
		cfg.IPBan.MaxFailures,
		time.Duration(cfg.IPBan.BanDurationSeconds)*time.Second,
		cfg.IPBan.Whitelist,
	)

	circuitBreaker := manager.NewCircuitBreaker(
		cfg.CircuitBreaker.FailureThresholdPercent,
		time.Duration(cfg.CircuitBreaker.WindowSizeSeconds)*time.Second,
		cfg.CircuitBreaker.MinRequests,
		time.Duration(cfg.CircuitBreaker.BreakDurationSeconds)*time.Second,
	)

	// Create middlewares
	authMW := middleware.NewAuthMiddleware(
		cfg.Auth.Enabled,
		cfg.GetUserCredentials(),
	)

	rateLimitMW := middleware.NewRateLimitMiddleware(
		cfg.RateLimit.Enabled,
		cfg.RateLimit.GlobalRequestsPerSecond,
		cfg.RateLimit.PerIPRequestsPerSecond,
	)

	ipBanMW := middleware.NewIPBanMiddleware(
		cfg.IPBan.Enabled,
		ipBanMgr,
	)

	circuitBreakerMW := middleware.NewCircuitBreakerMiddleware(
		cfg.CircuitBreaker.Enabled,
		circuitBreaker,
	)

	// Create proxies
	httpProxy := proxy.NewHTTPProxy(
		cfg.Server.HTTPPort,
		authMW,
		rateLimitMW,
		ipBanMW,
		circuitBreakerMW,
	)

	socks5Proxy := proxy.NewSOCKS5Proxy(
		cfg.Server.SOCKS5Port,
		authMW,
		rateLimitMW,
		ipBanMW,
		circuitBreakerMW,
	)

	return &Server{
		config:      cfg,
		httpProxy:   httpProxy,
		socks5Proxy: socks5Proxy,
		ipBanMgr:    ipBanMgr,
	}
}

// Run starts the server
func (s *Server) Run() error {
	// Start HTTP proxy in a goroutine
	go func() {
		if err := s.httpProxy.Start(); err != nil {
			logger.Fatal("HTTP proxy failed to start", "error", err)
		}
	}()

	// Start SOCKS5 proxy in a goroutine
	go func() {
		if err := s.socks5Proxy.Start(); err != nil {
			logger.Fatal("SOCKS5 proxy failed to start", "error", err)
		}
	}()

	logger.Info("DuDu Proxy is running")
	logger.Info(fmt.Sprintf("HTTP Proxy: localhost:%d", s.config.Server.HTTPPort))
	logger.Info(fmt.Sprintf("SOCKS5 Proxy: localhost:%d", s.config.Server.SOCKS5Port))

	// Wait for interrupt signal
	s.waitForShutdown()

	return nil
}

// waitForShutdown waits for interrupt signal and performs graceful shutdown
func (s *Server) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info(fmt.Sprintf("Received signal: %v", sig))
	logger.Info("Shutting down gracefully...")

	// Perform cleanup
	s.shutdown()

	logger.Info("Server stopped")
}

// shutdown performs cleanup operations
func (s *Server) shutdown() {
	// Stop IP ban manager cleanup routine
	if s.ipBanMgr != nil {
		s.ipBanMgr.Stop()
	}

	// Add a small delay to allow ongoing connections to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-ctx.Done()
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *config.Config {
	return s.config
}
