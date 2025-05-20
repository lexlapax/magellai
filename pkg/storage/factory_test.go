// ABOUTME: Tests for the storage factory pattern implementation
// ABOUTME: Ensures backends can be registered, created, and checked for availability

package storage

import (
	"errors"
	"testing"

	"github.com/lexlapax/magellai/internal/testutil/storagemock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory_NewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
	assert.NotNil(t, factory.backends)
}

func TestFactory_RegisterBackend(t *testing.T) {
	factory := NewFactory()

	// Register a test backend
	testBackendType := BackendType("test")
	called := false

	factory.RegisterBackend(testBackendType, func(config Config) (Backend, error) {
		called = true
		return storagemock.NewMockBackend(), nil
	})

	// Verify it's registered
	assert.True(t, factory.IsBackendAvailable(testBackendType))

	// Try to create it
	backend, err := factory.CreateBackend(testBackendType, nil)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.True(t, called)
}

func TestFactory_CreateBackend_Unknown(t *testing.T) {
	factory := NewFactory()

	// Try to create unknown backend
	backend, err := factory.CreateBackend("unknown", nil)
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "unknown storage backend type")
}

func TestFactory_IsBackendAvailable(t *testing.T) {
	factory := NewFactory()

	// Register a backend
	testType := BackendType("available")
	factory.RegisterBackend(testType, func(config Config) (Backend, error) {
		return storagemock.NewMockBackend(), nil
	})

	// Test availability
	assert.True(t, factory.IsBackendAvailable(testType))
	assert.False(t, factory.IsBackendAvailable("unavailable"))
}

func TestFactory_GetAvailableBackends(t *testing.T) {
	factory := NewFactory()

	// Initially empty
	backends := factory.GetAvailableBackends()
	assert.Empty(t, backends)

	// Register some backends
	factory.RegisterBackend("backend1", func(config Config) (Backend, error) {
		return storagemock.NewMockBackend(), nil
	})
	factory.RegisterBackend("backend2", func(config Config) (Backend, error) {
		return storagemock.NewMockBackend(), nil
	})

	// Check available backends
	backends = factory.GetAvailableBackends()
	assert.Len(t, backends, 2)

	// Check that both are in the list
	found := make(map[BackendType]bool)
	for _, backend := range backends {
		found[backend] = true
	}
	assert.True(t, found["backend1"])
	assert.True(t, found["backend2"])
}

func TestFactory_CreateBackend_Error(t *testing.T) {
	factory := NewFactory()

	// Register a backend that returns an error
	errorType := BackendType("error")
	expectedError := errors.New("backend creation failed")

	factory.RegisterBackend(errorType, func(config Config) (Backend, error) {
		return nil, expectedError
	})

	// Try to create it
	backend, err := factory.CreateBackend(errorType, nil)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, backend)
}

func TestDefaultFactory(t *testing.T) {
	// Test that defaultFactory is initialized
	assert.NotNil(t, defaultFactory)

	// Test package-level functions
	testType := BackendType("package-test")

	// Register through package function
	RegisterBackend(testType, func(config Config) (Backend, error) {
		return storagemock.NewMockBackend(), nil
	})

	// Check availability through package function
	assert.True(t, IsBackendAvailable(testType))

	// Create through package function
	backend, err := CreateBackend(testType, nil)
	assert.NoError(t, err)
	assert.NotNil(t, backend)

	// Get available backends through package function
	backends := GetAvailableBackends()
	assert.Contains(t, backends, testType)
}

func TestFactory_Concurrency(t *testing.T) {
	factory := NewFactory()

	// Register a backend
	testType := BackendType("concurrent")
	factory.RegisterBackend(testType, func(config Config) (Backend, error) {
		return storagemock.NewMockBackend(), nil
	})

	// Test concurrent access
	done := make(chan bool, 10)

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			assert.True(t, factory.IsBackendAvailable(testType))
			backends := factory.GetAvailableBackends()
			assert.NotEmpty(t, backends)
			done <- true
		}()
	}

	// Concurrent creates
	for i := 0; i < 5; i++ {
		go func() {
			backend, err := factory.CreateBackend(testType, nil)
			assert.NoError(t, err)
			assert.NotNil(t, backend)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestFactory_RegisterExisting(t *testing.T) {
	factory := NewFactory()

	// Register a backend
	testType := BackendType("existing")
	callCount := 0

	factory.RegisterBackend(testType, func(config Config) (Backend, error) {
		callCount++
		return storagemock.NewMockBackend(), nil
	})

	// Re-register the same type (should overwrite)
	factory.RegisterBackend(testType, func(config Config) (Backend, error) {
		callCount += 10
		return storagemock.NewMockBackend(), nil
	})

	// Create backend - should use the second registration
	_, err := factory.CreateBackend(testType, nil)
	require.NoError(t, err)

	assert.Equal(t, 10, callCount) // Should have called the second factory
}
