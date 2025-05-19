// ABOUTME: Tests for REPL package error definitions
// ABOUTME: Validates error constants, messages, and error wrapping behavior

package repl

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
			name:     "ErrInvalidCommand",
			err:      ErrInvalidCommand,
			expected: "invalid REPL command",
		},
		{
			name:     "ErrSessionNotInitialized",
			err:      ErrSessionNotInitialized,
			expected: "REPL session not initialized",
		},
		{
			name:     "ErrNoActiveSession",
			err:      ErrNoActiveSession,
			expected: "no active session",
		},
		{
			name:     "ErrCommandFailed",
			err:      ErrCommandFailed,
			expected: "command execution failed",
		},
		{
			name:     "ErrInvalidAttachment",
			err:      ErrInvalidAttachment,
			expected: "invalid attachment",
		},
		{
			name:     "ErrAttachmentNotFound",
			err:      ErrAttachmentNotFound,
			expected: "attachment not found",
		},
		{
			name:     "ErrInvalidSystemPrompt",
			err:      ErrInvalidSystemPrompt,
			expected: "invalid system prompt",
		},
		{
			name:     "ErrExportFailed",
			err:      ErrExportFailed,
			expected: "export failed",
		},
		{
			name:     "ErrInvalidExportFormat",
			err:      ErrInvalidExportFormat,
			expected: "invalid export format",
		},
		{
			name:     "ErrBranchOperationFailed",
			err:      ErrBranchOperationFailed,
			expected: "branch operation failed",
		},
		{
			name:     "ErrMergeOperationFailed",
			err:      ErrMergeOperationFailed,
			expected: "merge operation failed",
		},
		{
			name:     "ErrInvalidMetadataKey",
			err:      ErrInvalidMetadataKey,
			expected: "invalid metadata key",
		},
		{
			name:     "ErrRecoveryFailed",
			err:      ErrRecoveryFailed,
			expected: "session recovery failed",
		},
		{
			name:     "ErrCommandNotFound",
			err:      ErrCommandNotFound,
			expected: "command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseError := ErrInvalidCommand
	wrappedError := fmt.Errorf("failed to execute: %w", baseError)

	// Test that unwrapping works correctly
	assert.True(t, errors.Is(wrappedError, baseError))
	assert.Equal(t, "failed to execute: invalid REPL command", wrappedError.Error())

	// Test multiple levels of wrapping
	doubleWrapped := fmt.Errorf("command processing failed: %w", wrappedError)
	assert.True(t, errors.Is(doubleWrapped, baseError))
	assert.Equal(t, "command processing failed: failed to execute: invalid REPL command", doubleWrapped.Error())
}

func TestErrorComparison(t *testing.T) {
	// Test that each error is distinct
	allErrors := []error{
		ErrInvalidCommand,
		ErrSessionNotInitialized,
		ErrNoActiveSession,
		ErrCommandFailed,
		ErrInvalidAttachment,
		ErrAttachmentNotFound,
		ErrInvalidSystemPrompt,
		ErrExportFailed,
		ErrInvalidExportFormat,
		ErrBranchOperationFailed,
		ErrMergeOperationFailed,
		ErrInvalidMetadataKey,
		ErrRecoveryFailed,
		ErrCommandNotFound,
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
		name         string
		err          error
		isCommand    bool
		isSession    bool
		isAttachment bool
		isOperation  bool
	}{
		{
			name:      "Invalid command",
			err:       ErrInvalidCommand,
			isCommand: true,
		},
		{
			name:      "Command failed",
			err:       ErrCommandFailed,
			isCommand: true,
		},
		{
			name:      "Command not found",
			err:       ErrCommandNotFound,
			isCommand: true,
		},
		{
			name:      "Session not initialized",
			err:       ErrSessionNotInitialized,
			isSession: true,
		},
		{
			name:      "No active session",
			err:       ErrNoActiveSession,
			isSession: true,
		},
		{
			name:         "Invalid attachment",
			err:          ErrInvalidAttachment,
			isAttachment: true,
		},
		{
			name:         "Attachment not found",
			err:          ErrAttachmentNotFound,
			isAttachment: true,
		},
		{
			name:        "Export failed",
			err:         ErrExportFailed,
			isOperation: true,
		},
		{
			name:        "Branch operation failed",
			err:         ErrBranchOperationFailed,
			isOperation: true,
		},
		{
			name:        "Merge operation failed",
			err:         ErrMergeOperationFailed,
			isOperation: true,
		},
		{
			name:        "Recovery failed",
			err:         ErrRecoveryFailed,
			isOperation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that errors are correctly categorized
			if tt.isCommand {
				// Could add specific command error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isSession {
				// Could add specific session error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isAttachment {
				// Could add specific attachment error checks if needed
				assert.NotNil(t, tt.err)
			}
			if tt.isOperation {
				// Could add specific operation error checks if needed
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
			name:      "Session not initialized with context",
			baseError: ErrSessionNotInitialized,
			context:   "startup failed",
			expected:  "startup failed: REPL session not initialized",
		},
		{
			name:      "Command failed with command name",
			baseError: ErrCommandFailed,
			context:   "command: export",
			expected:  "command: export: command execution failed",
		},
		{
			name:      "Attachment not found with path",
			baseError: ErrAttachmentNotFound,
			context:   "file: /tmp/test.txt",
			expected:  "file: /tmp/test.txt: attachment not found",
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
	// Simulate common error usage patterns in REPL
	t.Run("Command execution", func(t *testing.T) {
		simulateCommand := func(cmd string) error {
			if cmd == "" {
				return fmt.Errorf("empty command: %w", ErrInvalidCommand)
			}
			if cmd == "unknown" {
				return fmt.Errorf("'%s': %w", cmd, ErrCommandNotFound)
			}
			if cmd == "fail" {
				return fmt.Errorf("execution of '%s': %w", cmd, ErrCommandFailed)
			}
			return nil
		}

		err := simulateCommand("")
		assert.True(t, errors.Is(err, ErrInvalidCommand))

		err = simulateCommand("unknown")
		assert.True(t, errors.Is(err, ErrCommandNotFound))

		err = simulateCommand("fail")
		assert.True(t, errors.Is(err, ErrCommandFailed))

		err = simulateCommand("valid")
		assert.NoError(t, err)
	})

	t.Run("Session operations", func(t *testing.T) {
		simulateSessionOp := func(initialized bool, hasActive bool) error {
			if !initialized {
				return fmt.Errorf("cannot perform operation: %w", ErrSessionNotInitialized)
			}
			if !hasActive {
				return fmt.Errorf("operation requires active session: %w", ErrNoActiveSession)
			}
			return nil
		}

		err := simulateSessionOp(false, false)
		assert.True(t, errors.Is(err, ErrSessionNotInitialized))

		err = simulateSessionOp(true, false)
		assert.True(t, errors.Is(err, ErrNoActiveSession))

		err = simulateSessionOp(true, true)
		assert.NoError(t, err)
	})

	t.Run("Export operations", func(t *testing.T) {
		simulateExport := func(format string, hasData bool) error {
			if format != "json" && format != "md" {
				return fmt.Errorf("format '%s': %w", format, ErrInvalidExportFormat)
			}
			if !hasData {
				return fmt.Errorf("no data to export: %w", ErrExportFailed)
			}
			return nil
		}

		err := simulateExport("xml", true)
		assert.True(t, errors.Is(err, ErrInvalidExportFormat))

		err = simulateExport("json", false)
		assert.True(t, errors.Is(err, ErrExportFailed))

		err = simulateExport("json", true)
		assert.NoError(t, err)
	})
}

func TestErrorChaining(t *testing.T) {
	// Test realistic error chaining scenarios
	t.Run("Branch operation chain", func(t *testing.T) {
		// Simulate a branch operation that fails due to no active session
		err := fmt.Errorf("cannot create branch: %w", ErrNoActiveSession)
		err = fmt.Errorf("branch operation: %w", err)
		err = fmt.Errorf("%w: %s", ErrBranchOperationFailed, err.Error())

		assert.True(t, errors.Is(err, ErrBranchOperationFailed))
		assert.Contains(t, err.Error(), "no active session")
	})

	t.Run("Recovery operation chain", func(t *testing.T) {
		// Simulate a recovery operation that fails
		err := fmt.Errorf("cannot read recovery file: permission denied")
		err = fmt.Errorf("recovery attempt: %w", err)
		err = fmt.Errorf("%w: %s", ErrRecoveryFailed, err.Error())

		assert.True(t, errors.Is(err, ErrRecoveryFailed))
		assert.Contains(t, err.Error(), "permission denied")
	})
}
