// ABOUTME: Domain types for session management including Session and SessionInfo
// ABOUTME: Core business entities for chat session lifecycle and metadata

package domain

import (
	"errors"
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
	
	// Branching support
	ParentID     string                 `json:"parent_id,omitempty"`     // ID of the parent session if this is a branch
	BranchPoint  int                    `json:"branch_point,omitempty"`  // Message index where the branch occurred
	BranchName   string                 `json:"branch_name,omitempty"`   // Optional name for this branch
	ChildIDs     []string               `json:"child_ids,omitempty"`     // IDs of child branches
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
	
	// Branch information
	ParentID     string    `json:"parent_id,omitempty"`
	BranchName   string    `json:"branch_name,omitempty"`
	ChildCount   int       `json:"child_count,omitempty"`
	IsBranch     bool      `json:"is_branch,omitempty"`
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
		ID:         s.ID,
		Name:       s.Name,
		Created:    s.Created,
		Updated:    s.Updated,
		Tags:       s.Tags,
		ParentID:   s.ParentID,
		BranchName: s.BranchName,
		ChildCount: len(s.ChildIDs),
		IsBranch:   s.IsBranch(),
	}

	if s.Conversation != nil {
		info.MessageCount = len(s.Conversation.Messages)
		info.Model = s.Conversation.Model
		info.Provider = s.Conversation.Provider
	}

	return info
}

// CreateBranch creates a new branch from this session at the specified message index.
// Returns the new branched session.
func (s *Session) CreateBranch(branchID string, branchName string, messageIndex int) (*Session, error) {
	if messageIndex < 0 || messageIndex > len(s.Conversation.Messages) {
		return nil, errors.New("invalid message index for branching")
	}
	
	now := time.Now()
	branch := &Session{
		ID:         branchID,
		Name:       branchName,
		Created:    now,
		Updated:    now,
		Tags:       append([]string{}, s.Tags...), // Copy tags
		Config:     copyMap(s.Config),
		ParentID:   s.ID,
		BranchPoint: messageIndex,
		BranchName:  branchName,
		ChildIDs:   []string{},
		Metadata:   make(map[string]interface{}),
	}
	
	// Create the conversation with messages up to the branch point
	branch.Conversation = &Conversation{
		ID:           branchID,
		Model:        s.Conversation.Model,
		Provider:     s.Conversation.Provider,
		Temperature:  s.Conversation.Temperature,
		MaxTokens:    s.Conversation.MaxTokens,
		SystemPrompt: s.Conversation.SystemPrompt,
		Created:      now,
		Updated:      now,
		Messages:     make([]Message, 0, messageIndex),
		Metadata:     copyMap(s.Conversation.Metadata),
	}
	
	// Copy messages up to the branch point
	for i := 0; i < messageIndex && i < len(s.Conversation.Messages); i++ {
		msgCopy := s.Conversation.Messages[i]
		msgCopy.ID = generateMessageID() // Generate new ID for the copy
		branch.Conversation.Messages = append(branch.Conversation.Messages, msgCopy)
	}
	
	// Add this branch to the parent's child list
	s.AddChild(branchID)
	
	return branch, nil
}

// AddChild adds a child branch ID to this session.
func (s *Session) AddChild(childID string) {
	// Check if child ID already exists
	for _, id := range s.ChildIDs {
		if id == childID {
			return
		}
	}
	s.ChildIDs = append(s.ChildIDs, childID)
	s.UpdateTimestamp()
}

// RemoveChild removes a child branch ID from this session.
func (s *Session) RemoveChild(childID string) {
	filtered := make([]string, 0, len(s.ChildIDs))
	for _, id := range s.ChildIDs {
		if id != childID {
			filtered = append(filtered, id)
		}
	}
	s.ChildIDs = filtered
	s.UpdateTimestamp()
}

// IsBranch returns true if this session is a branch of another session.
func (s *Session) IsBranch() bool {
	return s.ParentID != ""
}

// HasBranches returns true if this session has child branches.
func (s *Session) HasBranches() bool {
	return len(s.ChildIDs) > 0
}

// GetBranchDepth returns the depth of this branch in the tree (0 for root).
func (s *Session) GetBranchDepth() int {
	if s.ParentID == "" {
		return 0
	}
	// This is a simple implementation. In practice, you'd need to traverse the parent chain.
	return 1 // Placeholder - would need repository access to compute full depth
}

// Helper function to deep copy a map
func copyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{})
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Helper function to generate a new message ID
func generateMessageID() string {
	// This is a placeholder. In practice, you'd use a proper ID generation method.
	return time.Now().Format("20060102150405.000000")
}
