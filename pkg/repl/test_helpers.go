package repl

import (
	"testing"
)

// createTestSessionManager creates a SessionManager with a filesystem backend for testing
func createTestSessionManager(t *testing.T, baseDir string) *SessionManager {
	storage, err := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
		"base_dir": baseDir,
	})
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}

	manager, err := NewSessionManager(storage)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	return manager
}
