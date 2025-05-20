// ABOUTME: Mock implementations for storage backend interface
// ABOUTME: Provides reusable mock storage backends for testing

package storagemock

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
)

// Backend interface defines the basic operations for a mock storage backend
type Backend interface {
	// Core Session Repository operations
	Create(session *domain.Session) error
	Get(id string) (*domain.Session, error)
	Update(session *domain.Session) error
	Delete(id string) error
	List() ([]*domain.SessionInfo, error)
	Search(query string) ([]*domain.SearchResult, error)
	GetChildren(sessionID string) ([]*domain.SessionInfo, error)
	GetBranchTree(sessionID string) (*domain.BranchTree, error)

	// Storage-specific extensions
	NewSession(name string) *domain.Session
	ExportSession(id string, format domain.ExportFormat, w io.Writer) error
	MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error)
	Close() error

	// Mock-specific methods
	GetSessionCount() int
	HasSession(id string) bool

	// Error setters for testing
	WithSaveError(err error) *MockBackend
	WithLoadError(err error) *MockBackend
	WithListError(err error) *MockBackend
	WithDeleteError(err error) *MockBackend
	WithSearchError(err error) *MockBackend
	WithMergeError(err error) *MockBackend
	WithExportError(err error) *MockBackend
	WithGetChildrenError(err error) *MockBackend
	WithGetBranchTreeError(err error) *MockBackend
	WithSearchResults(results []*domain.Session) *MockBackend
	WithUserID(userID string) *MockBackend

	// Legacy methods for backward compatibility
	SaveSession(session *domain.Session) error
	LoadSession(id string) (*domain.Session, error)
	ListSessions() ([]*domain.SessionInfo, error)
	DeleteSession(id string) error
	SearchSessions(query string) ([]*domain.SearchResult, error)
}

// MockBackend implements the Backend interface for testing
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
	errOnExport        error
	errOnGetChildren   error
	errOnGetBranchTree error
	errOnNewSession    error
	currUserID         string
	searchResults      []*domain.Session
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

// NewSession creates a new session with the given name
func (m *MockBackend) NewSession(name string) *domain.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errOnNewSession != nil {
		return nil
	}

	id := fmt.Sprintf("session-%d", time.Now().UnixNano())
	now := time.Now()

	session := domain.NewSession(id)
	session.Name = name
	session.Created = now
	session.Updated = now

	m.sessions[id] = session
	return session
}

// Create creates a new session in the mock backend
func (m *MockBackend) Create(session *domain.Session) error {
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

	// Check if session already exists
	if _, exists := m.sessions[session.ID]; exists {
		return fmt.Errorf("session already exists: %s", session.ID)
	}

	// Clone the session to avoid reference issues
	cloned := cloneSession(session)
	m.sessions[session.ID] = cloned
	return nil
}

// Update updates an existing session in the mock backend
func (m *MockBackend) Update(session *domain.Session) error {
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

	// Check if session exists
	if _, exists := m.sessions[session.ID]; !exists {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	// Clone the session to avoid reference issues
	cloned := cloneSession(session)
	m.sessions[session.ID] = cloned
	return nil
}

// SaveSession saves a session to the mock backend (legacy method)
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
	cloned := cloneSession(session)
	m.sessions[session.ID] = cloned
	return nil
}

// Get loads a session from the mock backend
func (m *MockBackend) Get(id string) (*domain.Session, error) {
	return m.LoadSession(id)
}

// LoadSession loads a session from the mock backend (legacy method)
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

// List lists all sessions from the mock backend
func (m *MockBackend) List() ([]*domain.SessionInfo, error) {
	return m.ListSessions()
}

// ListSessions lists all sessions from the mock backend (legacy method)
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

// Delete deletes a session from the mock backend
func (m *MockBackend) Delete(id string) error {
	return m.DeleteSession(id)
}

// DeleteSession deletes a session from the mock backend (legacy method)
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

// Search searches for sessions
func (m *MockBackend) Search(query string) ([]*domain.SearchResult, error) {
	return m.SearchSessions(query)
}

// SearchSessions searches for sessions (legacy method)
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
		child, exists := m.sessions[childID]
		if exists {
			childTree, err := m.buildBranchTreeNode(child)
			if err == nil {
				tree.Children = append(tree.Children, childTree)
			}
		}
	}

	return tree, nil
}

// buildBranchTreeNode builds a branch tree node for a session
func (m *MockBackend) buildBranchTreeNode(session *domain.Session) (*domain.BranchTree, error) {
	node := &domain.BranchTree{
		Session:  session.ToSessionInfo(),
		Children: []*domain.BranchTree{},
	}

	for _, childID := range session.ChildIDs {
		child, exists := m.sessions[childID]
		if exists {
			childNode, err := m.buildBranchTreeNode(child)
			if err == nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node, nil
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

	// Use the session's built-in Clone method if it becomes available
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
		cloned.Conversation = session.Conversation.Clone()
	}

	return cloned
}
