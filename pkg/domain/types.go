// ABOUTME: Shared types and constants used across the domain layer
// ABOUTME: Common enums, interfaces, and utility types for domain entities

package domain

import (
	"errors"
	"time"
)

// Common errors used across the domain layer.
var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidSession     = errors.New("invalid session")
	ErrMessageNotFound    = errors.New("message not found")
	ErrInvalidMessage     = errors.New("invalid message")
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrInvalidAttachment  = errors.New("invalid attachment")
	ErrProviderNotFound   = errors.New("provider not found")
	ErrModelNotFound      = errors.New("model not found")
	ErrInvalidModel       = errors.New("invalid model")
	ErrInvalidRole        = errors.New("invalid message role")
	ErrInvalidCapability  = errors.New("invalid model capability")
	ErrNoContent          = errors.New("message must have content or attachments")
)

// SessionState represents the state of a session.
type SessionState string

// SessionState constants define the possible states.
const (
	SessionStateActive   SessionState = "active"
	SessionStateArchived SessionState = "archived"
	SessionStateDeleted  SessionState = "deleted"
)

// ExportFormat represents the format for exporting sessions.
type ExportFormat string

// ExportFormat constants define the possible export formats.
const (
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatText     ExportFormat = "text"
	ExportFormatHTML     ExportFormat = "html"
)

// BranchTree represents the hierarchical structure of session branches.
type BranchTree struct {
	Session  *SessionInfo  `json:"session"`
	Children []*BranchTree `json:"children,omitempty"`
}

// BranchNode represents a node in a branch visualization.
type BranchNode struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	BranchPoint  int          `json:"branch_point"`
	MessageCount int          `json:"message_count"`
	Created      time.Time    `json:"created"`
	Children     []BranchNode `json:"children,omitempty"`
}

// DefaultConversationSettings defines default settings for new conversations.
const (
	DefaultTemperature  float64 = 0.7
	DefaultMaxTokens    int     = 0  // Use model's default
	DefaultContextLimit int     = 10 // Number of messages to keep in context
)

// Repository interfaces define the contract for data persistence.
// These interfaces should be implemented by the infrastructure layer.

// SessionRepository defines the contract for session persistence.
//
// It provides core CRUD operations for sessions and branch management.
// Implementations should handle appropriate error conditions and be
// consistent with return values - empty collections should be returned
// as empty slices, not nil.
type SessionRepository interface {
	// Create creates a new session in storage.
	//
	// Parameters:
	//   - session: The session to create, must not be nil and should have a valid ID
	//
	// Returns:
	//   - error: nil on success, otherwise an error describing what went wrong
	Create(session *Session) error

	// Get retrieves a session by ID.
	//
	// Parameters:
	//   - id: The unique identifier of the session to retrieve
	//
	// Returns:
	//   - *Session: The session if found, nil otherwise
	//   - error: nil if the session was found or ErrSessionNotFound if not found
	Get(id string) (*Session, error)

	// Update updates an existing session in storage.
	//
	// Parameters:
	//   - session: The session to update, must not be nil and should have a valid ID
	//
	// Returns:
	//   - error: nil on success, ErrSessionNotFound if session doesn't exist
	Update(session *Session) error

	// Delete removes a session from storage.
	//
	// Parameters:
	//   - id: The unique identifier of the session to delete
	//
	// Returns:
	//   - error: nil on success, ErrSessionNotFound if session doesn't exist
	Delete(id string) error

	// List returns a list of all stored sessions.
	//
	// Returns:
	//   - []*SessionInfo: Array of session info objects, empty slice if none exist
	//   - error: nil on success, otherwise an error describing what went wrong
	List() ([]*SessionInfo, error)

	// Search finds sessions matching the given query.
	//
	// Parameters:
	//   - query: The search term to match against session content
	//
	// Returns:
	//   - []*SearchResult: Array of search results, empty slice if no matches
	//   - error: nil on success, otherwise an error describing what went wrong
	Search(query string) ([]*SearchResult, error)

	// Branch-specific operations

	// GetChildren returns all direct child branches of a session.
	//
	// Parameters:
	//   - sessionID: The unique identifier of the parent session
	//
	// Returns:
	//   - []*SessionInfo: Array of child session info objects, empty slice if none exist
	//   - error: nil on success, ErrSessionNotFound if parent doesn't exist
	GetChildren(sessionID string) ([]*SessionInfo, error)

	// GetBranchTree returns the full branch tree starting from a session.
	//
	// Parameters:
	//   - sessionID: The unique identifier of the root session for the tree
	//
	// Returns:
	//   - *BranchTree: The hierarchical tree of branches
	//   - error: nil on success, ErrSessionNotFound if root doesn't exist
	GetBranchTree(sessionID string) (*BranchTree, error)
}

// ProviderRepository defines the contract for provider/model configuration.
//
// It manages the available LLM providers and models, including their capabilities
// and configuration. Implementations should cache results where appropriate
// for performance. Empty collections should be returned as empty slices, not nil.
type ProviderRepository interface {
	// GetProvider retrieves a provider by name.
	//
	// Parameters:
	//   - name: The unique name of the provider to retrieve
	//
	// Returns:
	//   - *Provider: The provider if found, nil otherwise
	//   - error: nil if the provider was found or ErrProviderNotFound if not found
	GetProvider(name string) (*Provider, error)

	// ListProviders returns a list of all available providers.
	//
	// Returns:
	//   - []*Provider: Array of provider objects, empty slice if none exist
	//   - error: nil on success, otherwise an error describing what went wrong
	ListProviders() ([]*Provider, error)

	// GetModel retrieves a model by provider and model name.
	//
	// Parameters:
	//   - providerName: The name of the provider
	//   - modelName: The name of the model
	//
	// Returns:
	//   - *Model: The model if found, nil otherwise
	//   - error: nil if the model was found, ErrProviderNotFound if provider doesn't exist,
	//     or ErrModelNotFound if model doesn't exist
	GetModel(providerName, modelName string) (*Model, error)

	// ListModels returns a list of all available models across all providers.
	//
	// Returns:
	//   - []*Model: Array of model objects, empty slice if none exist
	//   - error: nil on success, otherwise an error describing what went wrong
	ListModels() ([]*Model, error)

	// SearchModels finds models with the specified capability.
	//
	// Parameters:
	//   - capability: The capability to search for
	//
	// Returns:
	//   - []*Model: Array of models with the requested capability, empty slice if none match
	//   - error: nil on success, ErrInvalidCapability if capability is invalid
	SearchModels(capability ModelCapability) ([]*Model, error)
}

// String returns the session state as a string.
func (s SessionState) String() string {
	return string(s)
}

// IsValid checks if the session state is valid.
func (s SessionState) IsValid() bool {
	return s == SessionStateActive || s == SessionStateArchived || s == SessionStateDeleted
}

// String returns the export format as a string.
func (f ExportFormat) String() string {
	return string(f)
}

// IsValid checks if the export format is valid.
func (f ExportFormat) IsValid() bool {
	return f == ExportFormatJSON || f == ExportFormatMarkdown ||
		f == ExportFormatText || f == ExportFormatHTML
}
