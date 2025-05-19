// ABOUTME: Error definitions for the REPL package
// ABOUTME: Provides standard errors for REPL operations

package repl

import "errors"

// REPL-specific errors
var (
	// ErrInvalidCommand indicates an invalid REPL command
	ErrInvalidCommand = errors.New("invalid REPL command")

	// ErrSessionNotInitialized indicates the REPL session is not initialized
	ErrSessionNotInitialized = errors.New("REPL session not initialized")

	// ErrNoActiveSession indicates no session is currently active
	ErrNoActiveSession = errors.New("no active session")

	// ErrCommandFailed indicates a command execution failed
	ErrCommandFailed = errors.New("command execution failed")

	// ErrInvalidAttachment indicates an invalid attachment
	ErrInvalidAttachment = errors.New("invalid attachment")

	// ErrAttachmentNotFound indicates the attachment was not found
	ErrAttachmentNotFound = errors.New("attachment not found")

	// ErrInvalidSystemPrompt indicates an invalid system prompt
	ErrInvalidSystemPrompt = errors.New("invalid system prompt")

	// ErrExportFailed indicates export operation failed
	ErrExportFailed = errors.New("export failed")

	// ErrInvalidExportFormat indicates an invalid export format
	ErrInvalidExportFormat = errors.New("invalid export format")

	// ErrBranchOperationFailed indicates a branch operation failed
	ErrBranchOperationFailed = errors.New("branch operation failed")

	// ErrMergeOperationFailed indicates a merge operation failed
	ErrMergeOperationFailed = errors.New("merge operation failed")

	// ErrInvalidMetadataKey indicates an invalid metadata key
	ErrInvalidMetadataKey = errors.New("invalid metadata key")

	// ErrRecoveryFailed indicates session recovery failed
	ErrRecoveryFailed = errors.New("session recovery failed")

	// ErrCommandNotFound indicates a command was not found
	ErrCommandNotFound = errors.New("command not found")
)
