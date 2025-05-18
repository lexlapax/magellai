// ABOUTME: Shared types and constants used across the domain layer
// ABOUTME: Common enums, interfaces, and utility types for domain entities

package domain

import (
	"errors"
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

// DefaultConversationSettings defines default settings for new conversations.
const (
	DefaultTemperature  float64 = 0.7
	DefaultMaxTokens    int     = 0  // Use model's default
	DefaultContextLimit int     = 10 // Number of messages to keep in context
)

// Repository interfaces define the contract for data persistence.
// These interfaces should be implemented by the infrastructure layer.

// SessionRepository defines the contract for session persistence.
type SessionRepository interface {
	Create(session *Session) error
	Get(id string) (*Session, error)
	Update(session *Session) error
	Delete(id string) error
	List() ([]*SessionInfo, error)
	Search(query string) ([]*SearchResult, error)
}

// ProviderRepository defines the contract for provider/model configuration.
type ProviderRepository interface {
	GetProvider(name string) (*Provider, error)
	ListProviders() ([]*Provider, error)
	GetModel(providerName, modelName string) (*Model, error)
	ListModels() ([]*Model, error)
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
