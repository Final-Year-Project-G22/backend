package service

import (
	"fmt"
	"sync"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai_tool/domain/port"
)

type ToolRegistry struct {
	mu       sync.RWMutex
	handlers map[string]port.ToolHandler
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers: make(map[string]port.ToolHandler),
	}
}

func (r *ToolRegistry) Register(h port.ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := h.Name()
	if _, exists := r.handlers[name]; exists {
		panic(fmt.Sprintf("tool %q already registered", name))
	}
	r.handlers[name] = h
}

func (r *ToolRegistry) Get(name string) (port.ToolHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, ok := r.handlers[name]
	return h, ok
}

func (r *ToolRegistry) List() []port.ToolHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]port.ToolHandler, 0, len(r.handlers))
	for _, h := range r.handlers {
		result = append(result, h)
	}
	return result
}
