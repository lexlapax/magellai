// ABOUTME: Domain types for session management including Session and SessionInfo
// ABOUTME: Core business entities for chat session lifecycle and metadata

package domain

import (
	"time"
)

// Session represents a complete chat session with all its data.
// This is the primary aggregate root for session management.
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

// SessionInfo provides summary information about a session.
// Used for listing and searching sessions without loading full conversation history.
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

// NewSession creates a new session with the given ID.
func NewSession(id string) *Session {
	now := time.Now()
	return &Session{
		ID:           id,
		Created:      now,
		Updated:      now,
		Conversation: NewConversation(id),
		Config:       make(map[string]interface{}),
		Tags:         []string{},
		Metadata:     make(map[string]interface{}),
	}
}

// UpdateTimestamp updates the session's Updated field to the current time.
func (s *Session) UpdateTimestamp() {
	s.Updated = time.Now()
}

// AddTag adds a tag to the session if it doesn't already exist.
func (s *Session) AddTag(tag string) {
	for _, t := range s.Tags {
		if t == tag {
			return
		}
	}
	s.Tags = append(s.Tags, tag)
	s.UpdateTimestamp()
}

// RemoveTag removes a tag from the session.
func (s *Session) RemoveTag(tag string) {
	tags := make([]string, 0, len(s.Tags))
	for _, t := range s.Tags {
		if t != tag {
			tags = append(tags, t)
		}
	}
	s.Tags = tags
	s.UpdateTimestamp()
}

// ToSessionInfo creates a SessionInfo summary from the full Session.
func (s *Session) ToSessionInfo() *SessionInfo {
	info := &SessionInfo{
		ID:      s.ID,
		Name:    s.Name,
		Created: s.Created,
		Updated: s.Updated,
		Tags:    s.Tags,
	}
	
	if s.Conversation != nil {
		info.MessageCount = len(s.Conversation.Messages)
		info.Model = s.Conversation.Model
		info.Provider = s.Conversation.Provider
	}
	
	return info
}