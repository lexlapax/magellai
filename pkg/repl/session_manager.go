// ABOUTME: Manages REPL sessions using the storage backend abstraction
// ABOUTME: Provides high-level session management with pluggable storage implementations

package repl

import (
	"fmt"

	"github.com/lexlapax/magellai/internal/logging"
)

// SessionManager handles session persistence and lifecycle
type SessionManager struct {
	*StorageManager
}

// NewSessionManager creates a new session manager with the given storage manager
func NewSessionManager(storageManager *StorageManager) (*SessionManager, error) {
	logging.LogDebug("Creating session manager with storage manager")

	if storageManager == nil {
		return nil, fmt.Errorf("storage manager cannot be nil")
	}

	logging.LogDebug("Session manager created successfully")
	return &SessionManager{
		StorageManager: storageManager,
	}, nil
}

// NewSession creates a new session with an optional name
func (sm *SessionManager) NewSession(name string) (*Session, error) {
	logging.LogInfo("Creating new session", "name", name)
	session := sm.StorageManager.NewSession(name)

	// Save the initial session
	if err := sm.StorageManager.SaveSession(session); err != nil {
		logging.LogError(err, "Failed to save new session", "id", session.ID)
		return nil, err
	}

	logging.LogInfo("Session created successfully", "id", session.ID, "name", name)
	return session, nil
}
