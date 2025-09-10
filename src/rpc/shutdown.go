package rpc

import (
	"context"
	"fmt"

	"github.com/Ajinx1/go-message-pattern-server/src/utility"
)

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.isShuttingDown = true
	if s.listener != nil {
		utility.LogAndPrint("RPC: Server shutting down...")
		if err := s.listener.Close(); err != nil && !isClosedError(err) {
			s.mu.Unlock()
			utility.LogAndPrint(fmt.Sprintf("RPC: Failed to close listener | Error: %v", err))
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}
	s.mu.Unlock()

	// Close all active connections
	s.connMu.Lock()
	for conn := range s.activeConns {
		if err := conn.Close(); err != nil && !isClosedError(err) {
			utility.LogAndPrint(fmt.Sprintf("RPC: Failed to close connection | Error: %v", err))
		}
	}
	s.connMu.Unlock()

	// Signal all goroutines (heartbeat, etc.) to exit
	close(s.shutdownChan)

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utility.LogAndPrint("RPC: Server shutdown complete")
		return nil
	case <-ctx.Done():
		utility.LogAndPrint(fmt.Sprintf("RPC: Shutdown timeout | Error: %v", ctx.Err()))
		return ctx.Err()
	}
}
