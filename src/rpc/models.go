package rpc

import (
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Config struct {
	Addr              string
	Timeout           time.Duration
	MaxConnections    int
	RateLimitPerSec   int
	RateLimitBurst    int
	RetryAttempts     int
	RetryDelay        time.Duration
	HeartbeatInterval time.Duration
	HeartbeatTimeout  time.Duration
}

type Server struct {
	registry       *Registry
	config         *Config
	mu             sync.Mutex
	listener       net.Listener
	wg             sync.WaitGroup
	activeConns    map[net.Conn]struct{}
	shutdownChan   chan struct{}
	connMu         sync.RWMutex
	limiter        *rate.Limiter
	metrics        *Metrics
	isShuttingDown bool
}

type Metrics struct {
	RequestsTotal   uint64
	ErrorsTotal     uint64
	ActiveConns     uint64
	ProcessingTime  time.Duration
	HeartbeatsTotal uint64
	HeartbeatFails  uint64
	mu              sync.Mutex
}
