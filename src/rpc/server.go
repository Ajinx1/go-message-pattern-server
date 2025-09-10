package rpc

import (
	"net"
	"time"

	"golang.org/x/time/rate"
)

func applyDefaults(config *Config) *Config {
	if config == nil {
		config = &Config{}
	}

	if config.Addr == "" {
		config.Addr = ":8080"
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxConnections <= 0 {
		config.MaxConnections = 1000
	}
	if config.RateLimitPerSec <= 0 {
		config.RateLimitPerSec = 100
	}
	if config.RateLimitBurst <= 0 {
		config.RateLimitBurst = 200
	}
	if config.RetryAttempts < 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 500 * time.Millisecond
	}
	if config.HeartbeatInterval <= 0 {
		config.HeartbeatInterval = 15 * time.Second
	}
	if config.HeartbeatTimeout <= 0 {
		config.HeartbeatTimeout = 45 * time.Second
	}

	return config
}

func NewServer(config *Config) *Server {
	config = applyDefaults(config)

	return &Server{
		registry:     NewRegistry(),
		config:       config,
		activeConns:  make(map[net.Conn]struct{}),
		limiter:      rate.NewLimiter(rate.Limit(config.RateLimitPerSec), config.RateLimitBurst),
		metrics:      &Metrics{},
		shutdownChan: make(chan struct{}),
	}
}

func (s *Server) RegisterHandler(pattern string, handler MessageHandler) {
	s.registry.Register(pattern, handler)
}
