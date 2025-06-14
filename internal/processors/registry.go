package processors

import (
	"fmt"
	"sync"
)

// Registry manages available image processors
type Registry struct {
	processors map[string]Processor
	mu         sync.RWMutex
}

// NewRegistry creates a new processor registry
func NewRegistry() *Registry {
	return &Registry{
		processors: make(map[string]Processor),
	}
}

// Register adds a processor to the registry
func (r *Registry) Register(name string, processor Processor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.processors[name]; exists {
		return fmt.Errorf("processor %s already registered", name)
	}

	r.processors[name] = processor
	return nil
}

// Get retrieves a processor by name
func (r *Registry) Get(name string) (Processor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	processor, exists := r.processors[name]
	if !exists {
		return nil, fmt.Errorf("processor %s not found", name)
	}

	return processor, nil
}

// List returns all registered processor names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.processors))
	for name := range r.processors {
		names = append(names, name)
	}

	return names
}

// DefaultRegistry creates a registry with default processors
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	registry.Register("crop", NewCropProcessor())
	registry.Register("resize", NewResizeProcessor())
	return registry
}
