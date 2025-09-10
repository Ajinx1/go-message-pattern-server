package rpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Ajinx1/go-message-pattern-server/src/utility"
)

// acceptConnections handles incoming connections with rate limiting and max connection checks
func (s *Server) acceptConnections() {
	defer s.wg.Done()
	for {
		if s.isShuttingDown {
			return
		}

		// Check connection limit
		s.connMu.RLock()
		if len(s.activeConns) >= s.config.MaxConnections {
			s.connMu.RUnlock()
			utility.LogAndPrint(fmt.Sprintf("RPC: Max connections reached | Limit: %d", s.config.MaxConnections))
			time.Sleep(100 * time.Millisecond)
			continue
		}
		s.connMu.RUnlock()

		// Rate limiting
		if err := s.limiter.Wait(context.Background()); err != nil {
			utility.LogAndPrint(fmt.Sprintf("RPC: Rate limiter error | Error: %v", err))
			continue
		}

		// Set accept timeout
		if err := s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			utility.LogAndPrint(fmt.Sprintf("RPC: Failed to set accept deadline | Error: %v", err))
			continue
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if s.isShuttingDown {
				return
			}
			utility.LogAndPrint(fmt.Sprintf("RPC: Accept error | Error: %v", err))
			continue
		}

		s.connMu.Lock()
		s.activeConns[conn] = struct{}{}
		s.metrics.mu.Lock()
		s.metrics.ActiveConns++
		s.metrics.mu.Unlock()
		s.connMu.Unlock()

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}
