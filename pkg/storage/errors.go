// ABOUTME: Error definitions for the storage package
// ABOUTME: Provides standard errors for storage operations

package storage

import "errors"

// Common storage errors used across all backends
var (
	// ErrInvalidBackend indicates an invalid or unsupported storage backend
	ErrInvalidBackend = errors.New("invalid storage backend")

	// ErrSessionNotFound indicates the requested session was not found
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExists indicates a session with the same ID already exists
	ErrSessionExists = errors.New("session already exists")

	// ErrCorruptedData indicates the stored data is corrupted or invalid
	ErrCorruptedData = errors.New("corrupted storage data")

	// ErrStorageFull indicates the storage is full
	ErrStorageFull = errors.New("storage full")

	// ErrPermission indicates a permission error accessing storage
	ErrPermission = errors.New("storage permission denied")

	// ErrBackendNotAvailable indicates the requested backend is not available
	ErrBackendNotAvailable = errors.New("storage backend not available")

	// ErrTransactionFailed indicates a transaction failed
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrBranchNotFound indicates the requested branch was not found
	ErrBranchNotFound = errors.New("branch not found")

	// ErrInvalidBranch indicates an invalid branch operation
	ErrInvalidBranch = errors.New("invalid branch operation")

	// ErrMergeConflict indicates a merge conflict occurred
	ErrMergeConflict = errors.New("merge conflict")
)
