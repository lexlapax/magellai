// ABOUTME: Session storage manager that uses the storage backend abstraction
// ABOUTME: Provides REPL-specific session management with pluggable storage

package repl

import (
	"fmt"
	"io"

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
	storageSession := sm.backend.NewSession(name)
	return FromStorageSession(storageSession)
}

// SaveSession saves a session
func (sm *StorageManager) SaveSession(session *Session) error {
	storageSession := ToStorageSession(session)
	return sm.backend.SaveSession(storageSession)
}

// LoadSession loads a session by ID
func (sm *StorageManager) LoadSession(id string) (*Session, error) {
	storageSession, err := sm.backend.LoadSession(id)
	if err != nil {
		return nil, err
	}
	return FromStorageSession(storageSession), nil
}

// ListSessions lists all available sessions
func (sm *StorageManager) ListSessions() ([]*SessionInfo, error) {
	storageInfos, err := sm.backend.ListSessions()
	if err != nil {
		return nil, err
	}

	replInfos := make([]*SessionInfo, len(storageInfos))
	for i, info := range storageInfos {
		replInfos[i] = FromStorageSessionInfo(info)
	}

	return replInfos, nil
}

// DeleteSession removes a session
func (sm *StorageManager) DeleteSession(id string) error {
	return sm.backend.DeleteSession(id)
}

// SearchSessions searches for sessions by query
func (sm *StorageManager) SearchSessions(query string) ([]*SearchResult, error) {
	storageResults, err := sm.backend.SearchSessions(query)
	if err != nil {
		return nil, err
	}

	replResults := make([]*SearchResult, len(storageResults))
	for i, result := range storageResults {
		replResults[i] = FromStorageSearchResult(result)
	}

	return replResults, nil
}

// ExportSession exports a session in the specified format
func (sm *StorageManager) ExportSession(id string, format string, w io.Writer) error {
	// Convert string format to storage.ExportFormat
	var exportFormat storage.ExportFormat
	switch format {
	case "json":
		exportFormat = storage.ExportFormatJSON
	case "markdown":
		exportFormat = storage.ExportFormatMarkdown
	case "text":
		exportFormat = storage.ExportFormatText
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
