// ABOUTME: Tests for storage factory functionality
// ABOUTME: Ensures storage backend registration and availability checking works correctly

package repl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStorageBackendAvailable(t *testing.T) {
	// FileSystem backend should always be available
	assert.True(t, IsStorageBackendAvailable(FileSystemStorage))
	
	// Check if SQLite is available (depends on build tags)
	sqliteAvailable := IsStorageBackendAvailable(SQLiteStorage)
	
	// PostgreSQL should not be available (not implemented yet)
	assert.False(t, IsStorageBackendAvailable(PostgreSQLStorage))
	
	// Non-existent backend should not be available
	assert.False(t, IsStorageBackendAvailable("nonexistent"))
	
	t.Logf("FileSystem: available")
	t.Logf("SQLite: %v", sqliteAvailable)
}

func TestGetAvailableBackends(t *testing.T) {
	backends := GetAvailableBackends()
	
	// FileSystem should always be in the list
	found := false
	for _, b := range backends {
		if b == FileSystemStorage {
			found = true
			break
		}
	}
	assert.True(t, found, "FileSystemStorage should always be available")
	
	t.Logf("Available backends: %v", backends)
}

func TestRegisterBackend(t *testing.T) {
	// Register a mock backend
	mockBackendType := StorageType("mock")
	mockCalled := false
	
	RegisterStorageBackend(mockBackendType, func(config map[string]interface{}) (StorageBackend, error) {
		mockCalled = true
		return nil, nil
	})
	
	// Verify it's now available
	assert.True(t, IsStorageBackendAvailable(mockBackendType))
	
	// Try to create it
	_, _ = CreateStorageBackend(mockBackendType, nil)
	assert.True(t, mockCalled, "Mock backend factory should have been called")
}