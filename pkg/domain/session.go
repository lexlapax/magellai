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

// MergeType defines how sessions should be merged
type MergeType int

const (
	// MergeTypeContinuation appends all messages from source after the merge point
	MergeTypeContinuation MergeType = iota
	// MergeTypeRebase replays source messages on top of target
	MergeTypeRebase
	// MergeTypeCherryPick selectively includes messages
	MergeTypeCherryPick
)

// MergeOptions configures how sessions are merged
type MergeOptions struct {
	Type         MergeType
	SourceID     string
	TargetID     string
	MergePoint   int      // Message index in target where merge begins
	MessageIDs   []string // For cherry-pick mode
	CreateBranch bool     // Whether to create a new branch for the merge
	BranchName   string   // Name for the new branch
}

// MergeResult contains the result of a merge operation
type MergeResult struct {
	SessionID      string
	MergedCount    int
	ConflictCount  int
	Conflicts      []MergeConflict
	NewBranchID    string // If a new branch was created
}

// MergeConflict represents a conflict during merge
type MergeConflict struct {
	Index        int
	SourceMsg    Message
	TargetMsg    Message
	Resolution   string
}

// CanMerge checks if this session can be merged with another
func (s *Session) CanMerge(other *Session) error {
	if s.ID == other.ID {
		return errors.New("cannot merge session with itself")
	}
	
	if s.Conversation == nil || other.Conversation == nil {
		return errors.New("both sessions must have conversations")
	}
	
	// Check for circular dependencies
	if s.IsAncestorOf(other) || other.IsAncestorOf(s) {
		return nil // Valid merge scenarios
	}
	
	// For now, allow any two sessions to merge
	return nil
}

// IsAncestorOf checks if this session is an ancestor of another
func (s *Session) IsAncestorOf(other *Session) bool {
	if other.ParentID == "" {
		return false
	}
	
	if other.ParentID == s.ID {
		return true
	}
	
	// Would need repository access to check full ancestry
	return false
}

// PrepareForMerge validates and prepares a session for merging
func (s *Session) PrepareForMerge(source *Session, options MergeOptions) (*Session, error) {
	if err := s.CanMerge(source); err != nil {
		return nil, err
	}
	
	// Create a new session for the merge if requested
	var mergeSession *Session
	if options.CreateBranch {
		branchID := generateSessionID()
		branchName := options.BranchName
		if branchName == "" {
			branchName = "Merge of " + source.Name + " into " + s.Name
		}
		
		mergeSession = &Session{
			ID:           branchID,
			Name:         branchName,
			Created:      time.Now(),
			Updated:      time.Now(),
			Tags:         append([]string{}, s.Tags...),
			Config:       copyMap(s.Config),
			ParentID:     s.ID,
			BranchPoint:  options.MergePoint,
			BranchName:   branchName,
			ChildIDs:     []string{},
			Metadata:     make(map[string]interface{}),
		}
		
		// Copy the target conversation
		mergeSession.Conversation = s.Conversation.Clone()
		mergeSession.Conversation.ID = branchID
		
		// Add merge metadata
		if mergeSession.Metadata == nil {
			mergeSession.Metadata = make(map[string]interface{})
		}
		mergeSession.Metadata["merge_source"] = source.ID
		mergeSession.Metadata["merge_target"] = s.ID
		mergeSession.Metadata["merge_type"] = options.Type
		mergeSession.Metadata["merge_date"] = time.Now()
		
		s.AddChild(branchID)
	} else {
		mergeSession = s
	}
	
	return mergeSession, nil
}

// Helper function to generate a new session ID
func generateSessionID() string {
	return "session_" + time.Now().Format("20060102150405.000000")
}

// ExecuteMerge performs the actual merge of sessions based on options
func (s *Session) ExecuteMerge(source *Session, options MergeOptions) (*Session, *MergeResult, error) {
	// Prepare the merge session
	mergeSession, err := s.PrepareForMerge(source, options)
	if err != nil {
		return nil, nil, err
	}
	
	result := &MergeResult{
		SessionID:   mergeSession.ID,
		Conflicts:   []MergeConflict{},
	}
	
	if options.CreateBranch {
		result.NewBranchID = mergeSession.ID
	}
	
	switch options.Type {
	case MergeTypeContinuation:
		// Append all messages from source after the merge point
		startIndex := 0
		if source.IsBranch() && source.ParentID == s.ID {
			// If source is a branch of target, start after branch point
			startIndex = source.BranchPoint
		}
		
		for i := startIndex; i < len(source.Conversation.Messages); i++ {
			msg := source.Conversation.Messages[i]
			newMsg := msg.Clone()
			newMsg.ID = generateMessageID()
			mergeSession.Conversation.AddMessage(newMsg)
			result.MergedCount++
		}
		
	case MergeTypeRebase:
		// Replay source messages on top of target
		// First, truncate target at merge point if specified
		if options.MergePoint > 0 && options.MergePoint < len(mergeSession.Conversation.Messages) {
			mergeSession.Conversation.Messages = mergeSession.Conversation.Messages[:options.MergePoint]
		}
		
		// Then add all source messages
		for _, msg := range source.Conversation.Messages {
			newMsg := msg.Clone()
			newMsg.ID = generateMessageID()
			mergeSession.Conversation.AddMessage(newMsg)
			result.MergedCount++
		}
		
	case MergeTypeCherryPick:
		// Selectively include messages by ID
		messageMap := make(map[string]Message)
		for _, msg := range source.Conversation.Messages {
			messageMap[msg.ID] = msg
		}
		
		for _, msgID := range options.MessageIDs {
			if msg, exists := messageMap[msgID]; exists {
				newMsg := msg.Clone()
				newMsg.ID = generateMessageID()
				mergeSession.Conversation.AddMessage(newMsg)
				result.MergedCount++
			}
		}
		
	default:
		return nil, nil, errors.New("unsupported merge type")
	}
	
	// Update merge metadata
	mergeSession.UpdateTimestamp()
	
	return mergeSession, result, nil
}

// GetCommonAncestor finds the most recent common ancestor between two sessions
func (s *Session) GetCommonAncestor(other *Session) *string {
	// Simple implementation - checks direct parent relationships
	if s.ParentID == other.ID {
		return &other.ID
	}
	if other.ParentID == s.ID {
		return &s.ID
	}
	if s.ParentID != "" && s.ParentID == other.ParentID {
		return &s.ParentID
	}
	// Would need repository access to find more distant ancestors
	return nil
}

// GetBranchPath returns the path from root to this session
func (s *Session) GetBranchPath() []string {
	path := []string{s.ID}
	// Would need repository access to build full path
	if s.ParentID != "" {
		path = append([]string{s.ParentID}, path...)
	}
	return path
}
