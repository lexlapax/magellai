// ABOUTME: REPL specific types that extend or wrap domain types
// ABOUTME: Contains only REPL-specific concerns not part of core domain

package repl

import (
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// Use domain types directly
type Session = domain.Session
type SessionInfo = domain.SessionInfo
type Message = domain.Message
type Conversation = domain.Conversation
type SearchResult = domain.SearchResult
type SearchMatch = domain.SearchMatch

// REPL-specific types that don't belong in domain

// ConversationState represents the current state of a REPL conversation
type ConversationState struct {
	ActiveSession *domain.Session
	LastResponse  string
	StreamMode    bool
	Provider      llm.Provider
}

// CommandResult represents the result of a REPL command execution
type CommandResult struct {
	Success bool
	Message string
	Data    interface{}
}

// REPLEvent represents an event in the REPL lifecycle
type REPLEvent struct {
	Type    string // "command", "response", "error", "stream"
	Message string
	Data    interface{}
}
