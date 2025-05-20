// ABOUTME: Mock implementations for storage backend interface
// ABOUTME: Provides reusable mock storage backends for testing

package mocks

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
)

// MockBackend implements the storage.Backend interface for testing
type MockBackend struct {
	mu            sync.RWMutex
	sessions      map[string]*domain.Session
	closed        bool
	errorToReturn error
	callCounts    map[string]int
}

// NewMockBackend creates a new mock backend
func NewMockBackend() *MockBackend {
	return &MockBackend{
		sessions:   make(map[string]*domain.Session),
		callCounts: make(map[string]int),
	}
}

// SetError sets the error to return for all operations
func (mb *MockBackend) SetError(err error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.errorToReturn = err
}

// GetCallCount returns the call count for a specific method
func (mb *MockBackend) GetCallCount(method string) int {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.callCounts[method]
}

// NewSession creates a new session
func (mb *MockBackend) NewSession(name string) *domain.Session {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.callCounts["NewSession"]++

	session := domain.NewSession(fmt.Sprintf("test-session-%d", time.Now().Unix()))
	session.Name = name

	mb.sessions[session.ID] = session
	return session
}

// SaveSession saves a session
func (mb *MockBackend) SaveSession(session *domain.Session) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.callCounts["SaveSession"]++
	if mb.errorToReturn != nil {
		return mb.errorToReturn
	}

	// Clone session to avoid reference issues
	cloned := cloneSession(session)
	mb.sessions[session.ID] = cloned

	return nil
}

// LoadSession loads a session
func (mb *MockBackend) LoadSession(id string) (*domain.Session, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	mb.callCounts["LoadSession"]++
	if mb.errorToReturn != nil {
		return nil, mb.errorToReturn
	}

	session, exists := mb.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// Return a clone to avoid reference issues
	return cloneSession(session), nil
}

// ListSessions lists all sessions
func (mb *MockBackend) ListSessions() ([]*domain.SessionInfo, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	mb.callCounts["ListSessions"]++
	if mb.errorToReturn != nil {
		return nil, mb.errorToReturn
	}

	var infos []*domain.SessionInfo
	for _, session := range mb.sessions {
		infos = append(infos, session.ToSessionInfo())
	}

	// Sort for consistency
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ID < infos[j].ID
	})

	return infos, nil
}

// DeleteSession deletes a session
func (mb *MockBackend) DeleteSession(id string) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.callCounts["DeleteSession"]++
	if mb.errorToReturn != nil {
		return mb.errorToReturn
	}

	if _, exists := mb.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(mb.sessions, id)
	return nil
}

// SearchSessions searches for sessions
func (mb *MockBackend) SearchSessions(query string) ([]*domain.SearchResult, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	mb.callCounts["SearchSessions"]++
	if mb.errorToReturn != nil {
		return nil, mb.errorToReturn
	}

	var results []*domain.SearchResult

	// Create a simple mock result for each session
	for _, session := range mb.sessions {
		info := session.ToSessionInfo()
		result := domain.NewSearchResult(info)

		// Add a match based on the session name
		match := domain.NewSearchMatch(
			domain.SearchMatchTypeName,
			"",
			info.Name,
			info.Name,
			0,
		)
		result.AddMatch(match)

		results = append(results, result)
	}

	return results, nil
}

// ExportSession exports a session
func (mb *MockBackend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	mb.callCounts["ExportSession"]++
	if mb.errorToReturn != nil {
		return mb.errorToReturn
	}

	session, exists := mb.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	// Create a simple export format
	content := fmt.Sprintf("Export of session %s (%s) in %s format\n\n",
		session.ID, session.Name, format)

	if session.Conversation != nil {
		content += fmt.Sprintf("Messages: %d\n\n", len(session.Conversation.Messages))

		for i, msg := range session.Conversation.Messages {
			content += fmt.Sprintf("%d. %s: %s\n", i+1, msg.Role, msg.Content)
		}
	}

	_, err := w.Write([]byte(content))
	return err
}

// GetChildren returns child sessions
func (mb *MockBackend) GetChildren(sessionID string) ([]*domain.SessionInfo, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	mb.callCounts["GetChildren"]++
	if mb.errorToReturn != nil {
		return nil, mb.errorToReturn
	}

	session, exists := mb.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	var children []*domain.SessionInfo
	for _, childID := range session.ChildIDs {
		childSession, exists := mb.sessions[childID]
		if exists {
			children = append(children, childSession.ToSessionInfo())
		}
	}

	return children, nil
}

// GetBranchTree returns a branch tree
func (mb *MockBackend) GetBranchTree(sessionID string) (*domain.BranchTree, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	mb.callCounts["GetBranchTree"]++
	if mb.errorToReturn != nil {
		return nil, mb.errorToReturn
	}

	session, exists := mb.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Build a simple tree
	tree := &domain.BranchTree{
		Session:  session.ToSessionInfo(),
		Children: []*domain.BranchTree{},
	}

	// Add children
	for _, childID := range session.ChildIDs {
		childSession, exists := mb.sessions[childID]
		if exists {
			childTree := &domain.BranchTree{
				Session:  childSession.ToSessionInfo(),
				Children: []*domain.BranchTree{},
			}
			tree.Children = append(tree.Children, childTree)
		}
	}

	return tree, nil
}

// MergeSessions merges two sessions
func (mb *MockBackend) MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.callCounts["MergeSessions"]++
	if mb.errorToReturn != nil {
		return nil, mb.errorToReturn
	}

	// Check if sessions exist
	if _, exists := mb.sessions[targetID]; !exists {
		return nil, fmt.Errorf("target session not found")
	}

	source, exists := mb.sessions[sourceID]
	if !exists {
		return nil, fmt.Errorf("source session not found")
	}

	// Simple mock merge result
	result := &domain.MergeResult{
		SessionID:     targetID,
		MergedCount:   len(source.Conversation.Messages),
		ConflictCount: 0,
		Conflicts:     []domain.MergeConflict{},
	}

	// Add new branch ID if needed
	if options.CreateBranch {
		newBranchID := fmt.Sprintf("merge-branch-%d", time.Now().Unix())
		result.NewBranchID = newBranchID
	}

	return result, nil
}

// Close cleans up resources
func (mb *MockBackend) Close() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.callCounts["Close"]++
	if mb.errorToReturn != nil {
		return mb.errorToReturn
	}

	mb.closed = true
	return nil
}

// GetSessionCount returns the number of sessions
func (mb *MockBackend) GetSessionCount() int {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return len(mb.sessions)
}

// HasSession checks if a session exists
func (mb *MockBackend) HasSession(id string) bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	_, exists := mb.sessions[id]
	return exists
}

// Helper functions

func cloneSession(session *domain.Session) *domain.Session {
	if session == nil {
		return nil
	}

	cloned := &domain.Session{
		ID:          session.ID,
		Name:        session.Name,
		Created:     session.Created,
		Updated:     session.Updated,
		Tags:        make([]string, len(session.Tags)),
		Metadata:    make(map[string]interface{}),
		Config:      make(map[string]interface{}),
		ParentID:    session.ParentID,
		BranchPoint: session.BranchPoint,
		BranchName:  session.BranchName,
		ChildIDs:    make([]string, len(session.ChildIDs)),
	}

	// Copy slices and maps
	copy(cloned.Tags, session.Tags)
	copy(cloned.ChildIDs, session.ChildIDs)

	for k, v := range session.Metadata {
		cloned.Metadata[k] = v
	}

	for k, v := range session.Config {
		cloned.Config[k] = v
	}

	// Clone conversation
	if session.Conversation != nil {
		cloned.Conversation = session.Conversation.Clone()
	}

	return cloned
}
