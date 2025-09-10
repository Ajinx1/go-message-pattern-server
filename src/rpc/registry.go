package rpc

import "sync"

type Registry struct {
	handlers map[string]MessageHandler
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]MessageHandler)}
}

func (r *Registry) Register(pattern string, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[pattern] = handler
}

func (r *Registry) Get(pattern string) (MessageHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[pattern]
	return h, ok
}
