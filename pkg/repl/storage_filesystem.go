package repl

import (
	"io"
)

// FileSystemBackend implements the StorageBackend interface using file system storage
type FileSystemBackend struct {
	manager *FileSessionManager
}

// NewFileSystemBackend creates a new file system storage backend
func NewFileSystemBackend(baseDir string) (*FileSystemBackend, error) {
	// Use the FileSessionManager with the base directory
	manager, err := NewFileSessionManager(baseDir)
	if err != nil {
		return nil, err
	}
	return &FileSystemBackend{
		manager: manager,
	}, nil
}

// NewSession creates a new session
func (b *FileSystemBackend) NewSession(name string) *Session {
	return b.manager.NewSession(name)
}

// SaveSession saves a session to storage
func (b *FileSystemBackend) SaveSession(session *Session) error {
	return b.manager.SaveSession(session)
}

// LoadSession loads a session from storage
func (b *FileSystemBackend) LoadSession(id string) (*Session, error) {
	return b.manager.LoadSession(id)
}

// ListSessions lists all available sessions
func (b *FileSystemBackend) ListSessions() ([]*SessionInfo, error) {
	return b.manager.ListSessions()
}

// DeleteSession deletes a session from storage
func (b *FileSystemBackend) DeleteSession(id string) error {
	return b.manager.DeleteSession(id)
}

// SearchSessions searches sessions with the given query
func (b *FileSystemBackend) SearchSessions(query string) ([]*SearchResult, error) {
	return b.manager.SearchSessions(query)
}

// ExportSession exports a session in the specified format
func (b *FileSystemBackend) ExportSession(id string, format string, w io.Writer) error {
	return b.manager.ExportSession(id, format, w)
}
