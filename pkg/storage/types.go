// ABOUTME: Core types for the storage package that are independent of any specific implementation
// ABOUTME: Provides a clean abstraction layer for session storage without coupling to REPL concerns

package storage

import (
	"time"
)

// Session represents a stored session with all its data
type Session struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name,omitempty"`
	Messages []Message              `json:"messages"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Created  time.Time              `json:"created"`
	Updated  time.Time              `json:"updated"`
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Conversation-specific fields
	Model        string  `json:"model,omitempty"`
	Provider     string  `json:"provider,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
	SystemPrompt string  `json:"system_prompt,omitempty"`
}

// Message represents a single message within a session
type Message struct {
	ID          string                 `json:"id"`
	Role        string                 `json:"role"` // user, assistant, system
	Content     string                 `json:"content"`
	Timestamp   time.Time              `json:"timestamp"`
	Attachments []Attachment           `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Attachment represents an attachment to a message
type Attachment struct {
	Type     string                 `json:"type"` // image, file, text, audio, video
	URL      string                 `json:"url,omitempty"`
	MimeType string                 `json:"mime_type,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Size     int64                  `json:"size,omitempty"`
	Content  string                 `json:"content,omitempty"` // for text attachments
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SessionInfo provides summary information about a session
type SessionInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
	MessageCount int       `json:"message_count"`
	Model        string    `json:"model,omitempty"`
	Provider     string    `json:"provider,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
}

// SearchResult represents the result of a session search
type SearchResult struct {
	Session *SessionInfo
	Matches []SearchMatch
}

// SearchMatch represents a single match within a search result
type SearchMatch struct {
	Type     string // "message", "system_prompt", "name", "tag"
	Role     string // for messages: "user", "assistant", "system"
	Content  string // the actual matched content snippet
	Context  string // surrounding context
	Position int    // message index if applicable
}

// ExportFormat represents the format for exporting sessions
type ExportFormat string

const (
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatText     ExportFormat = "text"
)
