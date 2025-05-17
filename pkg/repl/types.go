// ABOUTME: Core types for the REPL package including session and search structures
// ABOUTME: Defines the types used for session management and conversation handling

package repl

import (
	"time"
)

// Session represents a complete REPL session with conversation and metadata
type Session struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name,omitempty"`
	Conversation *Conversation          `json:"conversation"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Created      time.Time              `json:"created"`
	Updated      time.Time              `json:"updated"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SessionInfo represents basic session information for listing
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

// SearchResult represents a search result with context
type SearchResult struct {
	Session *SessionInfo
	Matches []SearchMatch
}

// SearchMatch represents a single match with context
type SearchMatch struct {
	Type     string // "message", "system_prompt", "name", "tag"
	Role     string // for messages: "user", "assistant", "system"
	Content  string // the actual matched content snippet
	Context  string // surrounding context
	Position int    // message index if applicable
}
