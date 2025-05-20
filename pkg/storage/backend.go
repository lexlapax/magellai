// ABOUTME: Defines the core storage interface for session management
// ABOUTME: Provides a pluggable backend system for different storage implementations

package storage

import (
	"io"

	"github.com/lexlapax/magellai/pkg/domain"
)

// Backend defines the interface for session storage implementations.
//
// It implements domain.SessionRepository and adds additional storage-specific
// operations like export, merge and lifecycle management. Backend implementations
// should be thread-safe and handle appropriate error conditions.
//
// All implementations should provide proper cleanup in their Close method to
// prevent resource leaks. Errors should be wrapped using fmt.Errorf with %w
// to maintain error context.
type Backend interface {
	// Core Session Repository operations

	// Create creates a new session in storage.
	//
	// Parameters:
	//   - session: The session to create, must not be nil and should have a valid ID
	//
	// Returns:
	//   - error: nil on success, otherwise an error describing what went wrong
	//
	// Implementation should check for duplicate IDs and handle them appropriately.
	Create(session *domain.Session) error

	// Get retrieves a session by ID.
	//
	// Parameters:
	//   - id: The unique identifier of the session to retrieve
	//
	// Returns:
	//   - *domain.Session: The session if found, nil otherwise
	//   - error: nil if the session was found or domain.ErrSessionNotFound if not found,
	//     other errors indicate retrieval problems
	Get(id string) (*domain.Session, error)

	// Update updates an existing session in storage.
	//
	// Parameters:
	//   - session: The session to update, must not be nil and should have a valid ID
	//
	// Returns:
	//   - error: nil on success, domain.ErrSessionNotFound if session doesn't exist,
	//     other errors describe what went wrong
	//
	// Implementation should update all fields of the session.
	Update(session *domain.Session) error

	// Delete removes a session from storage.
	//
	// Parameters:
	//   - id: The unique identifier of the session to delete
	//
	// Returns:
	//   - error: nil on success, domain.ErrSessionNotFound if session doesn't exist,
	//     other errors describe what went wrong
	//
	// Implementation should also clean up any associated resources.
	Delete(id string) error

	// List returns a list of all stored sessions.
	//
	// Returns:
	//   - []*domain.SessionInfo: Array of session info objects
	//   - error: nil on success, otherwise an error describing what went wrong
	//
	// Implementation should return an empty slice, not nil, when no sessions exist.
	List() ([]*domain.SessionInfo, error)

	// Search finds sessions matching the given query.
	//
	// Parameters:
	//   - query: The search term to match against session content
	//
	// Returns:
	//   - []*domain.SearchResult: Array of search results with matching sessions
	//   - error: nil on success, otherwise an error describing what went wrong
	//
	// Implementation should return an empty slice, not nil, when no matches found.
	Search(query string) ([]*domain.SearchResult, error)

	// Storage-specific extensions

	// NewSession creates and returns a new session with the given name.
	// This is a convenience method that doesn't persist the session to storage.
	//
	// Parameters:
	//   - name: The name for the new session
	//
	// Returns:
	//   - *domain.Session: A newly initialized session with generated ID
	NewSession(name string) *domain.Session

	// ExportSession exports a session in the specified format.
	//
	// Parameters:
	//   - id: The unique identifier of the session to export
	//   - format: The format to export the session in (JSON, Markdown, etc.)
	//   - w: The writer to write the exported content to
	//
	// Returns:
	//   - error: nil on success, domain.ErrSessionNotFound if session doesn't exist,
	//     other errors describe export failures
	ExportSession(id string, format domain.ExportFormat, w io.Writer) error

	// GetChildren returns all direct child branches of a session.
	// This implements the domain.SessionRepository method.
	//
	// Parameters:
	//   - sessionID: The unique identifier of the parent session
	//
	// Returns:
	//   - []*domain.SessionInfo: Array of child session info objects
	//   - error: nil on success, domain.ErrSessionNotFound if parent doesn't exist,
	//     other errors describe what went wrong
	//
	// Implementation should return an empty slice, not nil, when no children exist.
	GetChildren(sessionID string) ([]*domain.SessionInfo, error)

	// GetBranchTree returns the full branch tree starting from a session.
	// This implements the domain.SessionRepository method.
	//
	// Parameters:
	//   - sessionID: The unique identifier of the root session for the tree
	//
	// Returns:
	//   - *domain.BranchTree: The hierarchical tree of branches
	//   - error: nil on success, domain.ErrSessionNotFound if root doesn't exist,
	//     other errors describe what went wrong
	GetBranchTree(sessionID string) (*domain.BranchTree, error)

	// MergeSessions merges two sessions according to the specified options.
	//
	// Parameters:
	//   - targetID: The ID of the target session (destination of merge)
	//   - sourceID: The ID of the source session (to merge from)
	//   - options: Merge configuration options
	//
	// Returns:
	//   - *domain.MergeResult: Results of the merge operation
	//   - error: nil on success, domain.ErrSessionNotFound if either session doesn't exist,
	//     other errors describe merge failures
	//
	// Implementation should handle conflict resolution based on the provided options.
	MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error)

	// Close cleans up any resources used by the backend.
	//
	// Returns:
	//   - error: nil on success, otherwise an error describing cleanup failures
	//
	// Implementation should ensure all resources (file handles, connections) are properly released.
	Close() error
}

// Config represents backend-specific configuration
type Config map[string]interface{}

// BackendType represents the type of storage backend
type BackendType string

const (
	// FileSystemBackend represents filesystem-based storage
	FileSystemBackend BackendType = "filesystem"

	// SQLiteBackend represents SQLite database storage
	SQLiteBackend BackendType = "sqlite"

	// PostgreSQLBackend represents PostgreSQL database storage
	PostgreSQLBackend BackendType = "postgresql"

	// MemoryBackend represents in-memory storage
	MemoryBackend BackendType = "memory"
)

// Ensure Backend extends domain.SessionRepository
// We use a type assertion with nil which will be checked at compile time
var _ domain.SessionRepository = (Backend)(nil)
