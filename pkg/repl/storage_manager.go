// ABOUTME: Session storage manager that uses the storage backend abstraction
// ABOUTME: Provides REPL-specific session management with pluggable storage

package repl

import (
	"fmt"
	"io"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

// StorageManager manages sessions using the storage backend abstraction
type StorageManager struct {
	backend     storage.Backend
	backendType storage.BackendType
}

// NewStorageManager creates a new storage manager with the specified backend
func NewStorageManager(backend storage.Backend) (*StorageManager, error) {
	if backend == nil {
		return nil, fmt.Errorf("storage backend cannot be nil")
	}

	return &StorageManager{
		backend:     backend,
		backendType: storage.FileSystemBackend, // Default to filesystem
	}, nil
}

// NewSession creates a new session
func (sm *StorageManager) NewSession(name string) *domain.Session {
	return sm.backend.NewSession(name)
}

// SaveSession saves a session
func (sm *StorageManager) SaveSession(session *domain.Session) error {
	return sm.backend.SaveSession(session)
}

// LoadSession loads a session by ID
func (sm *StorageManager) LoadSession(id string) (*domain.Session, error) {
	return sm.backend.LoadSession(id)
}

// ListSessions lists all available sessions
func (sm *StorageManager) ListSessions() ([]*domain.SessionInfo, error) {
	return sm.backend.ListSessions()
}

// DeleteSession removes a session
func (sm *StorageManager) DeleteSession(id string) error {
	return sm.backend.DeleteSession(id)
}

// SearchSessions searches for sessions by query
func (sm *StorageManager) SearchSessions(query string) ([]*domain.SearchResult, error) {
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

	sm, err := NewStorageManager(backend)
	if err != nil {
		return nil, err
	}
	sm.backendType = backendType
	return sm, nil
}

// IsBackendAvailable checks if a storage backend is available
func IsBackendAvailable(backendType storage.BackendType) bool {
	return storage.IsBackendAvailable(backendType)
}

// GetAvailableBackends returns the list of available storage backends
func GetAvailableBackends() []storage.BackendType {
	return storage.GetAvailableBackends()
}

// CurrentSession holds the current active session (temporary implementation)
var currentSession *domain.Session

// CurrentSession returns the current active session
func (sm *StorageManager) CurrentSession() *domain.Session {
	return currentSession
}

// SetCurrentSession sets the current active session
func (sm *StorageManager) SetCurrentSession(session *domain.Session) {
	currentSession = session
}

// GenerateSessionID generates a new unique session ID
func (sm *StorageManager) GenerateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// GetChildren returns all direct child branches of a session
func (sm *StorageManager) GetChildren(sessionID string) ([]*domain.SessionInfo, error) {
	return sm.backend.GetChildren(sessionID)
}

// GetBranchTree returns the full branch tree starting from a session
func (sm *StorageManager) GetBranchTree(sessionID string) (*domain.BranchTree, error) {
	return sm.backend.GetBranchTree(sessionID)
}

// MergeSessions merges two sessions according to the specified options
func (sm *StorageManager) MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error) {
	return sm.backend.MergeSessions(targetID, sourceID, options)
}
