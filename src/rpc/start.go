package rpc

import (
	"fmt"
	"net"

	"github.com/Ajinx1/go-message-pattern-server/src/utility"
)

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.config.Addr, err)
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	utility.LogAndPrint(fmt.Sprintf("RPC: Server starting | Address: %s | MaxConnections: %d | RateLimitPerSec: %d | HeartbeatInterval: %s",
		s.config.Addr, s.config.MaxConnections, s.config.RateLimitPerSec, s.config.HeartbeatInterval))

	s.wg.Add(1)
	go s.acceptConnections()
	return nil
}
