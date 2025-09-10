package rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Ajinx1/go-message-pattern-server/src/utility"
)

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.connMu.Lock()
		delete(s.activeConns, conn)
		s.metrics.mu.Lock()
		s.metrics.ActiveConns--
		s.metrics.mu.Unlock()
		s.connMu.Unlock()
		conn.Close()
		s.wg.Done()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scanner := bufio.NewScanner(conn)
	heartbeatTimer := time.NewTimer(s.config.HeartbeatTimeout)
	defer heartbeatTimer.Stop()

	// Handle heartbeats
	go func() {
		ticker := time.NewTicker(s.config.HeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.sendHeartbeat(conn)
			}
		}
	}()

	for scanner.Scan() {
		// Reset heartbeat timer on any received message
		if !heartbeatTimer.Stop() {
			select {
			case <-heartbeatTimer.C:
			default:
			}
		}
		heartbeatTimer.Reset(s.config.HeartbeatTimeout)

		select {
		case <-ctx.Done():
			utility.LogAndPrint(fmt.Sprintf("RPC: Connection closed | RemoteAddr: %s", conn.RemoteAddr().String()))
			return
		default:
			jsonBytes := scanner.Bytes()
			if len(jsonBytes) == 0 {
				continue
			}

			var req Request
			if err := json.Unmarshal(jsonBytes, &req); err != nil {
				s.metrics.mu.Lock()
				s.metrics.ErrorsTotal++
				s.metrics.mu.Unlock()
				s.sendError(conn, "unknown", fmt.Sprintf("Invalid JSON: %v", err))
				continue
			}

			// Handle ping requests
			if req.Pattern.Cmd == "ping" {
				s.metrics.mu.Lock()
				s.metrics.HeartbeatsTotal++
				s.metrics.mu.Unlock()
				s.sendResponse(conn, "ping", Response{Data: "pong"})
				continue
			}

			// Basic request validation
			if req.Pattern.Cmd == "" {
				s.metrics.mu.Lock()
				s.metrics.ErrorsTotal++
				s.metrics.mu.Unlock()
				s.sendError(conn, "unknown", "Empty pattern command")
				continue
			}

			s.metrics.mu.Lock()
			s.metrics.RequestsTotal++
			s.metrics.mu.Unlock()

			utility.LogAndPrint(fmt.Sprintf("RPC: Received request | Pattern: %s | RemoteAddr: %s | Time: %s",
				req.Pattern.Cmd, conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05")))

			// Retry logic for handler execution
			var result interface{}
			var handlerErr error
			for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
				if handler, ok := s.registry.Get(req.Pattern.Cmd); ok {
					startTime := time.Now()
					result, handlerErr = handler(req.Data)
					if handlerErr == nil {
						s.metrics.mu.Lock()
						s.metrics.ProcessingTime += time.Since(startTime)
						s.metrics.mu.Unlock()
						break
					}
					if attempt < s.config.RetryAttempts {
						utility.LogAndPrint(fmt.Sprintf("RPC: Handler retry | Pattern: %s | Attempt: %d | Error: %v",
							req.Pattern.Cmd, attempt+1, handlerErr))
						time.Sleep(s.config.RetryDelay)
					}
				} else {
					s.metrics.mu.Lock()
					s.metrics.ErrorsTotal++
					s.metrics.mu.Unlock()
					s.sendError(conn, req.Pattern.Cmd, "Unknown pattern")
					continue
				}
			}

			if handlerErr != nil {
				s.metrics.mu.Lock()
				s.metrics.ErrorsTotal++
				s.metrics.mu.Unlock()
				s.sendError(conn, req.Pattern.Cmd, handlerErr.Error())
				continue
			}

			s.sendResponse(conn, req.Pattern.Cmd, Response{Data: result})
		}
	}

	if err := scanner.Err(); err != nil {
		s.metrics.mu.Lock()
		s.metrics.ErrorsTotal++
		s.metrics.mu.Unlock()
		utility.LogAndPrint(fmt.Sprintf("RPC: Scanner error | RemoteAddr: %s | Error: %v",
			conn.RemoteAddr().String(), err))
	}

	// Check for heartbeat timeout
	select {
	case <-heartbeatTimer.C:
		s.metrics.mu.Lock()
		s.metrics.HeartbeatFails++
		s.metrics.mu.Unlock()
		utility.LogAndPrint(fmt.Sprintf("RPC: Heartbeat timeout | RemoteAddr: %s", conn.RemoteAddr().String()))
	case <-ctx.Done():
	}
}
