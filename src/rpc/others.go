package rpc

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Ajinx1/go-message-pattern-server/src/utility"
)

func (s *Server) StartHeartbeat() {
	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.connMu.RLock()
			for conn := range s.activeConns {
				go s.sendHeartbeat(conn) // send asynchronously per connection
			}
			s.connMu.RUnlock()
		case <-s.shutdownChan:
			return
		}
	}
}

func (s *Server) sendHeartbeat(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			utility.LogAndPrint(fmt.Sprintf("RPC: Panic in sendHeartbeat | RemoteAddr: %s | Recovered: %v", conn.RemoteAddr().String(), r))
		}
	}()

	resp := Response{Data: "ping"}
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		utility.LogAndPrint(fmt.Sprintf("RPC: Failed to marshal heartbeat | Error: %v", err))
		return
	}
	jsonBytes = append(jsonBytes, '\n')
	if _, err := conn.Write(jsonBytes); err != nil {
		s.metrics.mu.Lock()
		s.metrics.HeartbeatFails++
		s.metrics.mu.Unlock()
		utility.LogAndPrint(fmt.Sprintf("RPC: Failed to send heartbeat | RemoteAddr: %s | Time: %s | Error: %v",
			conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05"), err))
		return
	}
	utility.LogAndPrint(fmt.Sprintf("RPC: Heartbeat sent | RemoteAddr: %s | Time: %s",
		conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05")))
}

func isClosedError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}

func (s *Server) GetMetrics() Metrics {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	return Metrics{
		RequestsTotal:   s.metrics.RequestsTotal,
		ErrorsTotal:     s.metrics.ErrorsTotal,
		ActiveConns:     s.metrics.ActiveConns,
		ProcessingTime:  s.metrics.ProcessingTime,
		HeartbeatsTotal: s.metrics.HeartbeatsTotal,
		HeartbeatFails:  s.metrics.HeartbeatFails,
	}
}
