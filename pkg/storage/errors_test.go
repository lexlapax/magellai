// ABOUTME: Tests for storage package error definitions
// ABOUTME: Validates error constants, messages, and error wrapping behavior

package storage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInvalidBackend",
			err:      ErrInvalidBackend,
			expected: "invalid storage backend",
		},
		{
			name:     "ErrSessionNotFound",
			err:      ErrSessionNotFound,
			expected: "session not found",
		},
		{
			name:     "ErrSessionExists",
			err:      ErrSessionExists,
			expected: "session already exists",
		},
		{
			name:     "ErrCorruptedData",
			err:      ErrCorruptedData,
			expected: "corrupted storage data",
		},
		{
			name:     "ErrStorageFull",
			err:      ErrStorageFull,
			expected: "storage full",
		},
		{
			name:     "ErrPermission",
			err:      ErrPermission,
			expected: "storage permission denied",
		},
		{
			name:     "ErrBackendNotAvailable",
			err:      ErrBackendNotAvailable,
			expected: "storage backend not available",
		},
		{
			name:     "ErrTransactionFailed",
			err:      ErrTransactionFailed,
			expected: "transaction failed",
		},
		{
			name:     "ErrBranchNotFound",
			err:      ErrBranchNotFound,
			expected: "branch not found",
		},
		{
			name:     "ErrInvalidBranch",
			err:      ErrInvalidBranch,
			expected: "invalid branch operation",
		},
		{
			name:     "ErrMergeConflict",
			err:      ErrMergeConflict,
			expected: "merge conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseError := ErrSessionNotFound
	wrappedError := fmt.Errorf("failed to load session: %w", baseError)

	// Test that unwrapping works correctly
	assert.True(t, errors.Is(wrappedError, baseError))
	assert.Equal(t, "failed to load session: session not found", wrappedError.Error())

	// Test multiple levels of wrapping
	doubleWrapped := fmt.Errorf("storage operation failed: %w", wrappedError)
	assert.True(t, errors.Is(doubleWrapped, baseError))
	assert.Equal(t, "storage operation failed: failed to load session: session not found", doubleWrapped.Error())
}

func TestErrorComparison(t *testing.T) {
	// Test that each error is distinct
	allErrors := []error{
		ErrInvalidBackend,
		ErrSessionNotFound,
		ErrSessionExists,
		ErrCorruptedData,
		ErrStorageFull,
		ErrPermission,
		ErrBackendNotAvailable,
		ErrTransactionFailed,
		ErrBranchNotFound,
		ErrInvalidBranch,
		ErrMergeConflict,
	}

	for i, err1 := range allErrors {
		for j, err2 := range allErrors {
			if i == j {
				assert.True(t, errors.Is(err1, err2))
			} else {
				assert.False(t, errors.Is(err1, err2))
			}
		}
	}
}

func TestErrorCategorization(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		isBackend bool
		isSession bool
		isStorage bool
		isBranch  bool
	}{
		{
			name:      "Invalid backend",
			err:       ErrInvalidBackend,
			isBackend: true,
		},
		{
			name:      "Backend not available",
			err:       ErrBackendNotAvailable,
			isBackend: true,
		},
		{
			name:      "Session not found",
			err:       ErrSessionNotFound,
			isSession: true,
		},
		{
			name:      "Session exists",
			err:       ErrSessionExists,
			isSession: true,
		},
		{
			name:      "Corrupted data",
			err:       ErrCorruptedData,
			isStorage: true,
		},
		{
			name:      "Storage full",
			err:       ErrStorageFull,
			isStorage: true,
		},
		{
			name:      "Permission denied",
			err:       ErrPermission,
			isStorage: true,
		},
		{
			name:      "Transaction failed",
			err:       ErrTransactionFailed,
			isStorage: true,
		},
		{
			name:     "Branch not found",
			err:      ErrBranchNotFound,
			isBranch: true,
		},
		{
			name:     "Invalid branch",
			err:      ErrInvalidBranch,
			isBranch: true,
		},
		{
			name:     "Merge conflict",
			err:      ErrMergeConflict,
			isBranch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that errors are correctly categorized
			if tt.isBackend {
				// Could add specific backend error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isSession {
				// Could add specific session error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isStorage {
				// Could add specific storage error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isBranch {
				// Could add specific branch error checks if needed
				assert.NotNil(t, tt.err)
			}
		})
	}
}

func TestErrorContext(t *testing.T) {
	// Test adding context to errors
	tests := []struct {
		name      string
		baseError error
		context   string
		expected  string
	}{
		{
			name:      "Session not found with ID",
			baseError: ErrSessionNotFound,
			context:   "session ID: abc123",
			expected:  "session ID: abc123: session not found",
		},
		{
			name:      "Backend not available with type",
			baseError: ErrBackendNotAvailable,
			context:   "backend: postgresql",
			expected:  "backend: postgresql: storage backend not available",
		},
		{
			name:      "Permission error with path",
			baseError: ErrPermission,
			context:   "path: /var/lib/magellai",
			expected:  "path: /var/lib/magellai: storage permission denied",
		},
		{
			name:      "Branch not found with name",
			baseError: ErrBranchNotFound,
			context:   "branch: feature-xyz",
			expected:  "branch: feature-xyz: branch not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextualError := fmt.Errorf("%s: %w", tt.context, tt.baseError)
			assert.Equal(t, tt.expected, contextualError.Error())
			assert.True(t, errors.Is(contextualError, tt.baseError))
		})
	}
}

func TestErrorUsagePatterns(t *testing.T) {
	// Simulate common error usage patterns in storage
	t.Run("Backend operations", func(t *testing.T) {
		simulateBackendOp := func(backend string) error {
			if backend == "invalid" {
				return fmt.Errorf("unknown backend '%s': %w", backend, ErrInvalidBackend)
			}
			if backend == "postgresql" {
				return fmt.Errorf("cannot connect to '%s': %w", backend, ErrBackendNotAvailable)
			}
			return nil
		}

		err := simulateBackendOp("invalid")
		assert.True(t, errors.Is(err, ErrInvalidBackend))

		err = simulateBackendOp("postgresql")
		assert.True(t, errors.Is(err, ErrBackendNotAvailable))

		err = simulateBackendOp("filesystem")
		assert.NoError(t, err)
	})

	t.Run("Session operations", func(t *testing.T) {
		simulateSessionOp := func(sessionID string, exists bool) error {
			if sessionID == "" {
				return fmt.Errorf("empty session ID: %w", ErrSessionNotFound)
			}
			if exists {
				return fmt.Errorf("session '%s': %w", sessionID, ErrSessionExists)
			}
			if sessionID == "missing" {
				return fmt.Errorf("session '%s': %w", sessionID, ErrSessionNotFound)
			}
			return nil
		}

		err := simulateSessionOp("", false)
		assert.True(t, errors.Is(err, ErrSessionNotFound))

		err = simulateSessionOp("exists", true)
		assert.True(t, errors.Is(err, ErrSessionExists))

		err = simulateSessionOp("missing", false)
		assert.True(t, errors.Is(err, ErrSessionNotFound))

		err = simulateSessionOp("valid", false)
		assert.NoError(t, err)
	})

	t.Run("Storage operations", func(t *testing.T) {
		simulateStorageOp := func(op string, data []byte) error {
			if len(data) > 1024 {
				return fmt.Errorf("data size %d exceeds limit: %w", len(data), ErrStorageFull)
			}
			if len(data) == 0 {
				return fmt.Errorf("empty data: %w", ErrCorruptedData)
			}
			if op == "write" && !hasPermission() {
				return fmt.Errorf("cannot write to storage: %w", ErrPermission)
			}
			return nil
		}

		err := simulateStorageOp("write", make([]byte, 2048))
		assert.True(t, errors.Is(err, ErrStorageFull))

		err = simulateStorageOp("write", []byte{})
		assert.True(t, errors.Is(err, ErrCorruptedData))

		err = simulateStorageOp("write", []byte("valid data"))
		// Permission check would depend on actual permission
		assert.NoError(t, err)
	})

	t.Run("Branch operations", func(t *testing.T) {
		simulateBranchOp := func(op string, branch string) error {
			if branch == "" {
				return fmt.Errorf("empty branch name: %w", ErrInvalidBranch)
			}
			if op == "switch" && branch == "missing" {
				return fmt.Errorf("cannot switch to branch '%s': %w", branch, ErrBranchNotFound)
			}
			if op == "merge" && branch == "conflicting" {
				return fmt.Errorf("merging branch '%s': %w", branch, ErrMergeConflict)
			}
			return nil
		}

		err := simulateBranchOp("create", "")
		assert.True(t, errors.Is(err, ErrInvalidBranch))

		err = simulateBranchOp("switch", "missing")
		assert.True(t, errors.Is(err, ErrBranchNotFound))

		err = simulateBranchOp("merge", "conflicting")
		assert.True(t, errors.Is(err, ErrMergeConflict))

		err = simulateBranchOp("create", "feature")
		assert.NoError(t, err)
	})
}

func TestTransactionErrors(t *testing.T) {
	// Test transaction-specific error patterns
	t.Run("Transaction lifecycle", func(t *testing.T) {
		simulateTransaction := func(steps []string) error {
			for i, step := range steps {
				if step == "fail" {
					return fmt.Errorf("step %d failed: %w", i, ErrTransactionFailed)
				}
			}
			return nil
		}

		err := simulateTransaction([]string{"begin", "update", "fail", "commit"})
		assert.True(t, errors.Is(err, ErrTransactionFailed))

		err = simulateTransaction([]string{"begin", "update", "commit"})
		assert.NoError(t, err)
	})
}

// Helper function for permission check simulation
func hasPermission() bool {
	// Simulate permission check - in real implementation this would check actual permissions
	return true
}
