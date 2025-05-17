// ABOUTME: Manages REPL sessions using the storage backend abstraction
// ABOUTME: Provides high-level session management with pluggable storage implementations

package repl

import (
	"fmt"
	"io"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
)

// SessionManager handles session persistence and lifecycle
type SessionManager struct {
	storage StorageBackend
}

// NewSessionManager creates a new session manager with the given storage backend
func NewSessionManager(storage StorageBackend) (*SessionManager, error) {
	logging.LogDebug("Creating session manager with storage backend")

	if storage == nil {
		return nil, fmt.Errorf("storage backend cannot be nil")
	}

	logging.LogDebug("Session manager created successfully")
	return &SessionManager{
		storage: storage,
	}, nil
}

// NewSession creates a new session with an optional name
func (sm *SessionManager) NewSession(name string) (*Session, error) {
	logging.LogInfo("Creating new session", "name", name)
	session := sm.storage.NewSession(name)

	// Save the initial session
	if err := sm.storage.SaveSession(session); err != nil {
		logging.LogError(err, "Failed to save new session", "id", session.ID)
		return nil, err
	}

	logging.LogInfo("Session created successfully", "id", session.ID, "name", name)
	return session, nil
}

// SaveSession saves a session to storage
func (sm *SessionManager) SaveSession(session *Session) error {
	start := time.Now()
	logging.LogDebug("Saving session", "id", session.ID)

	// Update the updated timestamp
	session.Updated = time.Now()

	if err := sm.storage.SaveSession(session); err != nil {
		logging.LogError(err, "Failed to save session", "id", session.ID)
		return err
	}

	duration := time.Since(start)
	logging.LogDebug("Session saved successfully", "id", session.ID, "duration", duration)
	return nil
}

// LoadSession loads a session from storage
func (sm *SessionManager) LoadSession(id string) (*Session, error) {
	start := time.Now()
	logging.LogInfo("Loading session", "id", id)

	session, err := sm.storage.LoadSession(id)
	if err != nil {
		logging.LogError(err, "Failed to load session", "id", id)
		return nil, err
	}

	duration := time.Since(start)
	logging.LogInfo("Session loaded successfully", "id", id, "duration", duration)
	return session, nil
}

// ListSessions returns a list of all available sessions
func (sm *SessionManager) ListSessions() ([]*SessionInfo, error) {
	start := time.Now()
	logging.LogDebug("Listing sessions")

	sessions, err := sm.storage.ListSessions()
	if err != nil {
		logging.LogError(err, "Failed to list sessions")
		return nil, err
	}

	duration := time.Since(start)
	logging.LogDebug("Sessions listed successfully", "count", len(sessions), "duration", duration)
	return sessions, nil
}

// DeleteSession deletes a session from storage
func (sm *SessionManager) DeleteSession(id string) error {
	logging.LogInfo("Deleting session", "id", id)

	if err := sm.storage.DeleteSession(id); err != nil {
		logging.LogError(err, "Failed to delete session", "id", id)
		return err
	}

	logging.LogInfo("Session deleted successfully", "id", id)
	return nil
}

// SearchSessions searches sessions with the given query
func (sm *SessionManager) SearchSessions(query string) ([]*SearchResult, error) {
	start := time.Now()
	logging.LogDebug("Searching sessions", "query", query)

	results, err := sm.storage.SearchSessions(query)
	if err != nil {
		logging.LogError(err, "Failed to search sessions", "query", query)
		return nil, err
	}

	duration := time.Since(start)
	logging.LogDebug("Sessions searched successfully", "query", query, "results", len(results), "duration", duration)
	return results, nil
}

// ExportSession exports a session in the specified format
func (sm *SessionManager) ExportSession(id string, format string, w io.Writer) error {
	logging.LogInfo("Exporting session", "id", id, "format", format)

	if err := sm.storage.ExportSession(id, format, w); err != nil {
		logging.LogError(err, "Failed to export session", "id", id, "format", format)
		return err
	}

	logging.LogInfo("Session exported successfully", "id", id, "format", format)
	return nil
}
