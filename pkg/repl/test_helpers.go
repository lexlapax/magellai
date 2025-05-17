package repl

import (
	"testing"

	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem" // Register filesystem backend
)

// createTestSessionManager creates a SessionManager with a filesystem backend for testing
func createTestSessionManager(t *testing.T, baseDir string) *SessionManager {
	storageManager, err := CreateStorageManager(storage.FileSystemBackend, storage.Config{
		"base_dir": baseDir,
	})
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	return &SessionManager{StorageManager: storageManager}
}
