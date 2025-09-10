package rpc

import (
	"encoding/json"
)

type MessageHandler func(data json.RawMessage) (interface{}, error)

type ServerWrapper struct {
	server *Server
}

func NewServerWrapper(server *Server) *ServerWrapper {
	return &ServerWrapper{server: server}
}

func (w *ServerWrapper) MessagePattern(pattern string, handler MessageHandler) {
	w.server.RegisterHandler(pattern, handler)
}
