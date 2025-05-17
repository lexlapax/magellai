package repl

import (
	"io"
)

// StorageBackend defines the interface for session storage implementations
type StorageBackend interface {
	NewSession(name string) *Session
	SaveSession(session *Session) error
	LoadSession(id string) (*Session, error)
	ListSessions() ([]*SessionInfo, error)
	DeleteSession(id string) error
	SearchSessions(query string) ([]*SearchResult, error)
	ExportSession(id string, format string, w io.Writer) error
}
