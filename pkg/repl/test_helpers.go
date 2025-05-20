// ABOUTME: Test helper functions for REPL package
// ABOUTME: Contains utilities for setting up test sessions and mocking REPL components
package repl

import (
	"testing"

	"github.com/lexlapax/magellai/pkg/repl/session"
	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem" // Register filesystem backend
)

// createTestSessionManager creates a SessionManager with a filesystem backend for testing
func createTestSessionManager(t *testing.T, baseDir string) *session.SessionManager {
	storageManager, err := session.CreateStorageManager(storage.FileSystemBackend, storage.Config{
		"base_dir": baseDir,
	})
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	return &session.SessionManager{StorageManager: storageManager}
}
