package rpc

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Ajinx1/go-message-pattern-server/src/utility"
)

// sendResponse sends a successful response to the client
func (s *Server) sendResponse(conn net.Conn, pattern string, resp Response) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		utility.LogAndPrint(fmt.Sprintf("RPC: Failed to marshal response | Error: %v", err))
		return
	}
	jsonBytes = append(jsonBytes, '\n')
	if _, err := conn.Write(jsonBytes); err != nil {
		s.metrics.mu.Lock()
		s.metrics.ErrorsTotal++
		s.metrics.mu.Unlock()
		utility.LogAndPrint(fmt.Sprintf("RPC: Failed to send response | Pattern: %s | RemoteAddr: %s | Time: %s | Error: %v",
			pattern, conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05"), err))
		return
	}
	utility.LogAndPrint(fmt.Sprintf("RPC: Response sent | Pattern: %s | RemoteAddr: %s | Time: %s",
		pattern, conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05")))
}

// sendError sends an error response to the client
func (s *Server) sendError(conn net.Conn, pattern, msg string) {
	resp := Response{Error: msg}
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		utility.LogAndPrint(fmt.Sprintf("RPC: Failed to marshal error response | Error: %v", err))
		return
	}
	jsonBytes = append(jsonBytes, '\n')
	if _, err := conn.Write(jsonBytes); err != nil {
		s.metrics.mu.Lock()
		s.metrics.ErrorsTotal++
		s.metrics.mu.Unlock()
		utility.LogAndPrint(fmt.Sprintf("RPC: Failed to send error response | Pattern: %s | RemoteAddr: %s | Time: %s | Error: %v",
			pattern, conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05"), err))
		return
	}
	utility.LogAndPrint(fmt.Sprintf("RPC: Error response sent | Pattern: %s | RemoteAddr: %s | Time: %s | Message: %s",
		pattern, conn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05"), msg))
}
