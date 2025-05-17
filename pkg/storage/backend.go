// ABOUTME: Defines the core storage interface for session management
// ABOUTME: Provides a pluggable backend system for different storage implementations

package storage

import (
	"io"
)

// Backend defines the interface for session storage implementations
type Backend interface {
	// NewSession creates a new session with the given name
	NewSession(name string) *Session

	// SaveSession persists a session to storage
	SaveSession(session *Session) error

	// LoadSession retrieves a session by ID
	LoadSession(id string) (*Session, error)

	// ListSessions returns a list of all stored sessions
	ListSessions() ([]*SessionInfo, error)

	// DeleteSession removes a session from storage
	DeleteSession(id string) error

	// SearchSessions finds sessions matching the given query
	SearchSessions(query string) ([]*SearchResult, error)

	// ExportSession exports a session in the specified format
	ExportSession(id string, format ExportFormat, w io.Writer) error

	// Close cleans up any resources used by the backend
	Close() error
}

// Config represents backend-specific configuration
type Config map[string]interface{}

// BackendType represents the type of storage backend
type BackendType string

const (
	// FileSystemBackend represents filesystem-based storage
	FileSystemBackend BackendType = "filesystem"

	// SQLiteBackend represents SQLite database storage
	SQLiteBackend BackendType = "sqlite"

	// PostgreSQLBackend represents PostgreSQL database storage
	PostgreSQLBackend BackendType = "postgresql"

	// MemoryBackend represents in-memory storage
	MemoryBackend BackendType = "memory"
)
