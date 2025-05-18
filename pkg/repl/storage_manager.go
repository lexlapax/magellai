// ABOUTME: Session storage manager that uses the storage backend abstraction
// ABOUTME: Provides REPL-specific session management with pluggable storage

package repl

import (
	"fmt"
	"io"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

// StorageManager manages sessions using the storage backend abstraction
type StorageManager struct {
	backend storage.Backend
}

// NewStorageManager creates a new storage manager with the specified backend
func NewStorageManager(backend storage.Backend) (*StorageManager, error) {
	if backend == nil {
		return nil, fmt.Errorf("storage backend cannot be nil")
	}

	return &StorageManager{
		backend: backend,
	}, nil
}

// NewSession creates a new session
func (sm *StorageManager) NewSession(name string) *Session {
	return sm.backend.NewSession(name)
}

// SaveSession saves a session
func (sm *StorageManager) SaveSession(session *Session) error {
	return sm.backend.SaveSession(session)
}

// LoadSession loads a session by ID
func (sm *StorageManager) LoadSession(id string) (*Session, error) {
	return sm.backend.LoadSession(id)
}

// ListSessions lists all available sessions
func (sm *StorageManager) ListSessions() ([]*SessionInfo, error) {
	return sm.backend.ListSessions()
}

// DeleteSession removes a session
func (sm *StorageManager) DeleteSession(id string) error {
	return sm.backend.DeleteSession(id)
}

// SearchSessions searches for sessions by query
func (sm *StorageManager) SearchSessions(query string) ([]*SearchResult, error) {
	return sm.backend.SearchSessions(query)
}

// ExportSession exports a session in the specified format
func (sm *StorageManager) ExportSession(id string, format string, w io.Writer) error {
	// Convert string format to domain.ExportFormat
	var exportFormat domain.ExportFormat
	switch format {
	case "json":
		exportFormat = domain.ExportFormatJSON
	case "markdown":
		exportFormat = domain.ExportFormatMarkdown
	case "text":
		exportFormat = domain.ExportFormatText
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	return sm.backend.ExportSession(id, exportFormat, w)
}

// Close closes the storage backend
func (sm *StorageManager) Close() error {
	return sm.backend.Close()
}

// CreateStorageManager creates a storage manager with the specified backend type
func CreateStorageManager(backendType storage.BackendType, config storage.Config) (*StorageManager, error) {
	backend, err := storage.CreateBackend(backendType, config)
	if err != nil {
		return nil, err
	}

	return NewStorageManager(backend)
}

// IsBackendAvailable checks if a storage backend is available
func IsBackendAvailable(backendType storage.BackendType) bool {
	return storage.IsBackendAvailable(backendType)
}

// GetAvailableBackends returns the list of available storage backends
func GetAvailableBackends() []storage.BackendType {
	return storage.GetAvailableBackends()
}
