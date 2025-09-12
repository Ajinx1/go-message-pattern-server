package rpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
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

	heartbeatTimer := time.NewTimer(s.config.HeartbeatTimeout)
	defer heartbeatTimer.Stop()

	// Handle heartbeats (unchanged)
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

	// Main read loop using length-prefixed framing: "<len>#<json bytes>"
	for {
		// 1) Read length prefix until '#'
		lengthBuf := make([]byte, 0, 16)
		tmp := make([]byte, 1)
		for {
			n, err := conn.Read(tmp)
			if err != nil {
				// Connection closed or read error — treat like scanner end
				if err == io.EOF {
					// graceful close by peer
					return
				}
				// Log & increment error metrics then return
				s.metrics.mu.Lock()
				s.metrics.ErrorsTotal++
				s.metrics.mu.Unlock()
				utility.LogAndPrint(fmt.Sprintf("RPC: Read error (prefix) | RemoteAddr: %s | Error: %v",
					conn.RemoteAddr().String(), err))
				return
			}
			if n == 0 {
				continue
			}
			if tmp[0] == '#' {
				break
			}
			lengthBuf = append(lengthBuf, tmp[0])
			// defensive: avoid runaway length prefix
			if len(lengthBuf) > 32 {
				s.metrics.mu.Lock()
				s.metrics.ErrorsTotal++
				s.metrics.mu.Unlock()
				s.sendError(conn, "unknown", "Invalid length prefix (too long)")
				return
			}
		}

		msgLenStr := string(lengthBuf)
		msgLen, err := strconv.Atoi(msgLenStr)
		if err != nil || msgLen <= 0 {
			s.metrics.mu.Lock()
			s.metrics.ErrorsTotal++
			s.metrics.mu.Unlock()
			s.sendError(conn, "unknown", fmt.Sprintf("Invalid length prefix: %s", msgLenStr))
			// continue to next message (or connection likely unusable)
			continue
		}

		// 2) Read exactly msgLen bytes
		msgBytes := make([]byte, msgLen)
		totalRead := 0
		for totalRead < msgLen {
			n, err := conn.Read(msgBytes[totalRead:])
			if err != nil {
				if err == io.EOF {
					s.metrics.mu.Lock()
					s.metrics.ErrorsTotal++
					s.metrics.mu.Unlock()
					utility.LogAndPrint(fmt.Sprintf("RPC: Unexpected EOF while reading message body | RemoteAddr: %s",
						conn.RemoteAddr().String()))
					return
				}
				s.metrics.mu.Lock()
				s.metrics.ErrorsTotal++
				s.metrics.mu.Unlock()
				s.sendError(conn, "unknown", fmt.Sprintf("Read error: %v", err))
				return
			}
			if n == 0 {
				continue
			}
			totalRead += n
		}

		// Reset heartbeat timer on any received message (similar effect to scanner.Scan())
		if !heartbeatTimer.Stop() {
			select {
			case <-heartbeatTimer.C:
			default:
			}
		}
		heartbeatTimer.Reset(s.config.HeartbeatTimeout)

		// Check if server is shutting down
		select {
		case <-ctx.Done():
			utility.LogAndPrint(fmt.Sprintf("RPC: Connection closed | RemoteAddr: %s", conn.RemoteAddr().String()))
			return
		default:
			// parse request (same as before)
			if len(msgBytes) == 0 {
				continue
			}

			req, err := parseRequest(msgBytes)
			if err != nil {
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

			// Retry logic for handler execution (unchanged)
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
					// continue to next incoming message
					goto NEXT_MESSAGE
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
	NEXT_MESSAGE:
		// continue to next loop iteration
		continue
	}

}
