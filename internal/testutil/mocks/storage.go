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
	mu                 sync.RWMutex
	sessions           map[string]*domain.Session
	closed             bool
	errOnSave          error
	errOnLoad          error
	errOnList          error
	errOnDelete        error
	errOnSearch        error
	errOnMerge         error
	errOnBranch        error
	errOnExport        error
	errOnGetChildren   error
	errOnGetBranchTree error
	errOnNewSession    error
	currUserID         string
	searchResults      []*domain.Session
	branchFunc         func(sessionID, branchName string) (*domain.Session, error)
}

// NewMockBackend creates a new mock backend
func NewMockBackend() *MockBackend {
	return &MockBackend{
		sessions:   make(map[string]*domain.Session),
		currUserID: "default-user",
	}
}

// WithUserID sets the user ID for the mock backend
func (m *MockBackend) WithUserID(userID string) *MockBackend {
	m.currUserID = userID
	return m
}

// WithSaveError sets an error to be returned on save operations
func (m *MockBackend) WithSaveError(err error) *MockBackend {
	m.errOnSave = err
	return m
}

// WithLoadError sets an error to be returned on load operations
func (m *MockBackend) WithLoadError(err error) *MockBackend {
	m.errOnLoad = err
	return m
}

// WithListError sets an error to be returned on list operations
func (m *MockBackend) WithListError(err error) *MockBackend {
	m.errOnList = err
	return m
}

// WithDeleteError sets an error to be returned on delete operations
func (m *MockBackend) WithDeleteError(err error) *MockBackend {
	m.errOnDelete = err
	return m
}

// WithSearchError sets an error to be returned on search operations
func (m *MockBackend) WithSearchError(err error) *MockBackend {
	m.errOnSearch = err
	return m
}

// WithMergeError sets an error to be returned on merge operations
func (m *MockBackend) WithMergeError(err error) *MockBackend {
	m.errOnMerge = err
	return m
}

// WithBranchError sets an error to be returned on branch operations
func (m *MockBackend) WithBranchError(err error) *MockBackend {
	m.errOnBranch = err
	return m
}

// WithExportError sets an error to be returned on export operations
func (m *MockBackend) WithExportError(err error) *MockBackend {
	m.errOnExport = err
	return m
}

// WithGetChildrenError sets an error to be returned on get children operations
func (m *MockBackend) WithGetChildrenError(err error) *MockBackend {
	m.errOnGetChildren = err
	return m
}

// WithGetBranchTreeError sets an error to be returned on get branch tree operations
func (m *MockBackend) WithGetBranchTreeError(err error) *MockBackend {
	m.errOnGetBranchTree = err
	return m
}

// WithSearchResults sets specific search results to be returned
func (m *MockBackend) WithSearchResults(results []*domain.Session) *MockBackend {
	m.searchResults = results
	return m
}

// WithBranchFunc sets a custom branch function
func (m *MockBackend) WithBranchFunc(f func(sessionID, branchName string) (*domain.Session, error)) *MockBackend {
	m.branchFunc = f
	return m
}

// NewSession creates a new session with the given name
func (m *MockBackend) NewSession(name string) *domain.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errOnNewSession != nil {
		return nil
	}

	id := fmt.Sprintf("session-%d", time.Now().UnixNano())
	now := time.Now()

	session := &domain.Session{
		ID:       id,
		Name:     name,
		Created:  now,
		Updated:  now,
		Config:   make(map[string]interface{}),
		Tags:     []string{},
		Metadata: make(map[string]interface{}),
		ChildIDs: []string{},
	}

	// Create conversation
	session.Conversation = &domain.Conversation{
		ID:       id,
		Created:  now,
		Updated:  now,
		Messages: []domain.Message{},
		Metadata: make(map[string]interface{}),
	}

	m.sessions[id] = session
	return session
}

// SaveSession saves a session to the mock backend
func (m *MockBackend) SaveSession(session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return io.ErrClosedPipe
	}

	if m.errOnSave != nil {
		return m.errOnSave
	}

	if session == nil {
		return fmt.Errorf("session is nil")
	}

	// Clone the session to avoid reference issues
	cloned := &domain.Session{
		ID:          session.ID,
		Name:        session.Name,
		Config:      make(map[string]interface{}),
		Created:     session.Created,
		Updated:     session.Updated,
		Tags:        make([]string, len(session.Tags)),
		Metadata:    make(map[string]interface{}),
		ParentID:    session.ParentID,
		BranchPoint: session.BranchPoint,
		BranchName:  session.BranchName,
		ChildIDs:    make([]string, len(session.ChildIDs)),
	}

	// Copy tags and dependencies
	copy(cloned.Tags, session.Tags)
	copy(cloned.ChildIDs, session.ChildIDs)

	// Copy config and metadata
	for k, v := range session.Config {
		cloned.Config[k] = v
	}
	for k, v := range session.Metadata {
		cloned.Metadata[k] = v
	}

	// Clone conversation
	if session.Conversation != nil {
		cloned.Conversation = &domain.Conversation{
			ID:           session.Conversation.ID,
			Model:        session.Conversation.Model,
			Provider:     session.Conversation.Provider,
			Temperature:  session.Conversation.Temperature,
			MaxTokens:    session.Conversation.MaxTokens,
			SystemPrompt: session.Conversation.SystemPrompt,
			Created:      session.Conversation.Created,
			Updated:      session.Conversation.Updated,
			Messages:     make([]domain.Message, len(session.Conversation.Messages)),
			Metadata:     make(map[string]interface{}),
		}

		copy(cloned.Conversation.Messages, session.Conversation.Messages)

		// Copy conversation metadata
		for k, v := range session.Conversation.Metadata {
			cloned.Conversation.Metadata[k] = v
		}
	}

	m.sessions[session.ID] = cloned
	return nil
}

// LoadSession loads a session from the mock backend
func (m *MockBackend) LoadSession(id string) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, io.ErrClosedPipe
	}

	if m.errOnLoad != nil {
		return nil, m.errOnLoad
	}

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// Clone the session to avoid reference issues
	return cloneSession(session), nil
}

// ListSessions lists all sessions from the mock backend
func (m *MockBackend) ListSessions() ([]*domain.SessionInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, io.ErrClosedPipe
	}

	if m.errOnList != nil {
		return nil, m.errOnList
	}

	var sessions []*domain.SessionInfo
	for _, session := range m.sessions {
		sessions = append(sessions, session.ToSessionInfo())
	}

	// Sort by ID for consistent ordering
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ID < sessions[j].ID
	})

	return sessions, nil
}

// DeleteSession deletes a session from the mock backend
func (m *MockBackend) DeleteSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return io.ErrClosedPipe
	}

	if m.errOnDelete != nil {
		return m.errOnDelete
	}

	if _, exists := m.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(m.sessions, id)
	return nil
}

// Close closes the mock backend
func (m *MockBackend) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return io.ErrClosedPipe
	}

	m.closed = true
	return nil
}

// SearchSessions searches for sessions
func (m *MockBackend) SearchSessions(query string) ([]*domain.SearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, io.ErrClosedPipe
	}

	if m.errOnSearch != nil {
		return nil, m.errOnSearch
	}

	// Create mock search results
	results := []*domain.SearchResult{}

	sessions, err := m.ListSessions()
	if err != nil {
		return nil, err
	}

	// Simple implementation - create a result for each session
	for _, session := range sessions {
		result := domain.NewSearchResult(session)

		// Add a mock match
		result.AddMatch(domain.NewSearchMatch(
			domain.SearchMatchTypeName,
			"",
			session.Name,
			session.Name,
			0,
		))

		results = append(results, result)
	}

	return results, nil
}

// MergeSessions merges two sessions
func (m *MockBackend) MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, io.ErrClosedPipe
	}

	if m.errOnMerge != nil {
		return nil, m.errOnMerge
	}

	// Default implementation - just check sessions exist
	if _, exists := m.sessions[targetID]; !exists {
		return nil, fmt.Errorf("session not found: %s", targetID)
	}
	if _, exists := m.sessions[sourceID]; !exists {
		return nil, fmt.Errorf("session not found: %s", sourceID)
	}

	// Return a basic merge result
	result := &domain.MergeResult{
		SessionID:     targetID,
		MergedCount:   1,
		ConflictCount: 0,
		Conflicts:     []domain.MergeConflict{},
	}

	if options.CreateBranch {
		branchID := "mock-branch-" + time.Now().Format("20060102150405")
		result.NewBranchID = branchID
	}

	return result, nil
}

// BranchSession creates a new branch from a session
func (m *MockBackend) BranchSession(sessionID, branchName string) (*domain.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, io.ErrClosedPipe
	}

	if m.errOnBranch != nil {
		return nil, m.errOnBranch
	}

	if m.branchFunc != nil {
		return m.branchFunc(sessionID, branchName)
	}

	// Default implementation
	parent, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Create branch
	branchID := fmt.Sprintf("%s-branch-%s", sessionID, branchName)

	// Use the parent's CreateBranch method if available
	branch, err := parent.CreateBranch(branchID, branchName, len(parent.Conversation.Messages))
	if err != nil {
		// If CreateBranch is not available, clone the parent manually
		branch = cloneSession(parent)
		branch.ID = branchID
		branch.Name = fmt.Sprintf("%s (%s)", parent.Name, branchName)
		branch.ParentID = sessionID
		branch.BranchName = branchName
		branch.BranchPoint = len(parent.Conversation.Messages)
		branch.ChildIDs = []string{}

		// Update parent
		parent.ChildIDs = append(parent.ChildIDs, branchID)
	}

	m.sessions[branchID] = branch
	return branch, nil
}

// ExportSession exports a session in the specified format
func (m *MockBackend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return io.ErrClosedPipe
	}

	if m.errOnExport != nil {
		return m.errOnExport
	}

	session, exists := m.sessions[id]
	if !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	// Simple mock implementation - just write some content to the writer
	mockContent := fmt.Sprintf("Exported session %s in %s format\n", id, format)
	mockContent += fmt.Sprintf("Name: %s\n", session.Name)
	mockContent += fmt.Sprintf("Created: %s\n", session.Created)
	mockContent += fmt.Sprintf("Updated: %s\n", session.Updated)

	if session.Conversation != nil {
		mockContent += fmt.Sprintf("Message count: %d\n", len(session.Conversation.Messages))

		if len(session.Conversation.Messages) > 0 {
			mockContent += "Messages:\n"
			for i, msg := range session.Conversation.Messages {
				mockContent += fmt.Sprintf("  %d. %s: %s\n", i+1, msg.Role, msg.Content)
			}
		}
	}

	_, err := w.Write([]byte(mockContent))
	return err
}

// GetChildren returns all direct child branches of a session
func (m *MockBackend) GetChildren(sessionID string) ([]*domain.SessionInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, io.ErrClosedPipe
	}

	if m.errOnGetChildren != nil {
		return nil, m.errOnGetChildren
	}

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	var children []*domain.SessionInfo
	for _, childID := range session.ChildIDs {
		child, exists := m.sessions[childID]
		if exists {
			children = append(children, child.ToSessionInfo())
		}
	}

	return children, nil
}

// GetBranchTree returns the full branch tree starting from a session
func (m *MockBackend) GetBranchTree(sessionID string) (*domain.BranchTree, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, io.ErrClosedPipe
	}

	if m.errOnGetBranchTree != nil {
		return nil, m.errOnGetBranchTree
	}

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Create branch tree
	tree := &domain.BranchTree{
		Session:  session.ToSessionInfo(),
		Children: []*domain.BranchTree{},
	}

	// Recursively build the tree
	for _, childID := range session.ChildIDs {
		_, exists := m.sessions[childID]
		if exists {
			childTree, err := m.GetBranchTree(childID)
			if err == nil {
				tree.Children = append(tree.Children, childTree)
			}
		}
	}

	return tree, nil
}

// GetSessionCount returns the number of sessions
func (m *MockBackend) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// HasSession checks if a session exists
func (m *MockBackend) HasSession(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.sessions[id]
	return exists
}

// cloneSession creates a deep copy of a session
func cloneSession(session *domain.Session) *domain.Session {
	if session == nil {
		return nil
	}

	cloned := &domain.Session{
		ID:          session.ID,
		Name:        session.Name,
		Config:      make(map[string]interface{}),
		Created:     session.Created,
		Updated:     session.Updated,
		Tags:        make([]string, len(session.Tags)),
		Metadata:    make(map[string]interface{}),
		ParentID:    session.ParentID,
		BranchPoint: session.BranchPoint,
		BranchName:  session.BranchName,
		ChildIDs:    make([]string, len(session.ChildIDs)),
	}

	// Copy tags and child IDs
	copy(cloned.Tags, session.Tags)
	copy(cloned.ChildIDs, session.ChildIDs)

	// Copy config and metadata
	for k, v := range session.Config {
		cloned.Config[k] = v
	}
	for k, v := range session.Metadata {
		cloned.Metadata[k] = v
	}

	// Clone conversation
	if session.Conversation != nil {
		cloned.Conversation = &domain.Conversation{
			ID:           session.Conversation.ID,
			Model:        session.Conversation.Model,
			Provider:     session.Conversation.Provider,
			Temperature:  session.Conversation.Temperature,
			MaxTokens:    session.Conversation.MaxTokens,
			SystemPrompt: session.Conversation.SystemPrompt,
			Created:      session.Conversation.Created,
			Updated:      session.Conversation.Updated,
			Messages:     make([]domain.Message, len(session.Conversation.Messages)),
			Metadata:     make(map[string]interface{}),
		}

		copy(cloned.Conversation.Messages, session.Conversation.Messages)

		// Copy conversation metadata
		for k, v := range session.Conversation.Metadata {
			cloned.Conversation.Metadata[k] = v
		}
	}

	return cloned
}
