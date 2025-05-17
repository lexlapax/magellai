// ABOUTME: Factory pattern implementation for creating storage backends
// ABOUTME: Provides a centralized way to register and instantiate different storage implementations

package storage

import (
	"fmt"
	"sync"
)

// Factory creates storage backends based on configuration
type Factory struct {
	mu       sync.RWMutex
	backends map[BackendType]func(config Config) (Backend, error)
}

// defaultFactory is the global factory instance
var defaultFactory = NewFactory()

// NewFactory creates a new storage factory
func NewFactory() *Factory {
	factory := &Factory{
		backends: make(map[BackendType]func(config Config) (Backend, error)),
	}
	return factory
}

// RegisterBackend registers a new storage backend factory function
func (f *Factory) RegisterBackend(backendType BackendType, factory func(config Config) (Backend, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.backends[backendType] = factory
}

// CreateBackend creates a storage backend of the specified type
func (f *Factory) CreateBackend(backendType BackendType, config Config) (Backend, error) {
	f.mu.RLock()
	factory, exists := f.backends[backendType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown storage backend type: %s", backendType)
	}

	return factory(config)
}

// IsBackendAvailable checks if a storage backend type is registered
func (f *Factory) IsBackendAvailable(backendType BackendType) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, exists := f.backends[backendType]
	return exists
}

// GetAvailableBackends returns a list of registered backend types
func (f *Factory) GetAvailableBackends() []BackendType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	backends := make([]BackendType, 0, len(f.backends))
	for backendType := range f.backends {
		backends = append(backends, backendType)
	}
	return backends
}

// Package-level convenience functions that use the default factory

// RegisterBackend registers a storage backend with the default factory
func RegisterBackend(backendType BackendType, factory func(config Config) (Backend, error)) {
	defaultFactory.RegisterBackend(backendType, factory)
}

// CreateBackend creates a storage backend using the default factory
func CreateBackend(backendType BackendType, config Config) (Backend, error) {
	return defaultFactory.CreateBackend(backendType, config)
}

// IsBackendAvailable checks if a backend is available in the default factory
func IsBackendAvailable(backendType BackendType) bool {
	return defaultFactory.IsBackendAvailable(backendType)
}

// GetAvailableBackends returns available backends from the default factory
func GetAvailableBackends() []BackendType {
	return defaultFactory.GetAvailableBackends()
}
