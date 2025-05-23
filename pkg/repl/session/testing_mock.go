// ABOUTME: Mock storage backend for testing
// ABOUTME: Provides a mock implementation of storage.Backend for unit tests

package session

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

// MockStorageBackend implements storage.Backend for testing
type MockStorageBackend struct {
	sessions map[string]*domain.Session
	calls    map[string]int
	err      error
}

// GetCallCount returns the number of times a method was called
func (m *MockStorageBackend) GetCallCount(method string) int {
	return m.calls[method]
}

// ClearCalls resets the call tracking
func (m *MockStorageBackend) ClearCalls() {
	m.calls = make(map[string]int)
}

// NewMockStorageBackend creates a new mock storage backend
func NewMockStorageBackend() *MockStorageBackend {
	return &MockStorageBackend{
		sessions: make(map[string]*domain.Session),
		calls:    make(map[string]int),
	}
}

// AsBackend returns the mock as a storage.Backend interface
func (m *MockStorageBackend) AsBackend() storage.Backend {
	return m
}

func (m *MockStorageBackend) NewSession(name string) *domain.Session {
	m.calls["NewSession"]++
	sessionID := fmt.Sprintf("test-session-%d-%d", time.Now().UnixNano(), m.calls["NewSession"])
	session := domain.NewSession(sessionID)
	session.Name = name
	return session
}

// Create implements storage.Backend.Create
func (m *MockStorageBackend) Create(session *domain.Session) error {
	m.calls["Create"]++
	m.calls["SaveSession"]++ // For backward compatibility with tests
	if m.err != nil {
		return m.err
	}
	// Check if already exists
	if _, exists := m.sessions[session.ID]; exists {
		return fmt.Errorf("session already exists: %s", session.ID)
	}
	m.sessions[session.ID] = session
	return nil
}

// Update implements storage.Backend.Update
func (m *MockStorageBackend) Update(session *domain.Session) error {
	m.calls["Update"]++
	m.calls["SaveSession"]++ // For backward compatibility with tests
	if m.err != nil {
		return m.err
	}
	// Check if exists
	if _, exists := m.sessions[session.ID]; !exists {
		return fmt.Errorf("session not found: %s", session.ID)
	}
	m.sessions[session.ID] = session
	return nil
}

// SaveSession is maintained for backward compatibility
func (m *MockStorageBackend) SaveSession(session *domain.Session) error {
	m.calls["SaveSession"]++
	if m.err != nil {
		return m.err
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *MockStorageBackend) Get(id string) (*domain.Session, error) {
	m.calls["LoadSession"]++
	if m.err != nil {
		return nil, m.err
	}
	session, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

func (m *MockStorageBackend) List() ([]*domain.SessionInfo, error) {
	m.calls["ListSessions"]++
	if m.err != nil {
		return nil, m.err
	}
	var infos []*domain.SessionInfo
	for _, session := range m.sessions {
		infos = append(infos, session.ToSessionInfo())
	}
	return infos, nil
}

func (m *MockStorageBackend) Delete(id string) error {
	m.calls["DeleteSession"]++
	if m.err != nil {
		return m.err
	}
	_, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	delete(m.sessions, id)
	return nil
}

func (m *MockStorageBackend) Search(query string) ([]*domain.SearchResult, error) {
	m.calls["SearchSessions"]++
	if m.err != nil {
		return nil, m.err
	}
	var results []*domain.SearchResult
	// Simple mock implementation
	for _, session := range m.sessions {
		if strings.Contains(strings.ToLower(session.Name), strings.ToLower(query)) {
			sessionInfo := session.ToSessionInfo()
			result := domain.NewSearchResult(sessionInfo)
			result.AddMatch(domain.NewSearchMatch(
				domain.SearchMatchTypeName,
				"",
				session.Name,
				"Session Name",
				-1,
			))
			results = append(results, result)
		}
	}
	return results, nil
}

func (m *MockStorageBackend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	m.calls["ExportSession"]++
	if m.err != nil {
		return m.err
	}
	session, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	_, err := fmt.Fprintf(w, "Exported session %s (%s)", session.ID, format)
	return err
}

func (m *MockStorageBackend) Close() error {
	m.calls["Close"]++
	return m.err
}

// GetChildren returns all direct child branches of a session
func (m *MockStorageBackend) GetChildren(sessionID string) ([]*domain.SessionInfo, error) {
	m.calls["GetChildren"]++
	if m.err != nil {
		return nil, m.err
	}

	// Find the session
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Get info for all child sessions
	children := make([]*domain.SessionInfo, 0, len(session.ChildIDs))
	for _, childID := range session.ChildIDs {
		if child, exists := m.sessions[childID]; exists {
			children = append(children, child.ToSessionInfo())
		}
	}

	return children, nil
}

// GetBranchTree returns the full branch tree starting from a session
func (m *MockStorageBackend) GetBranchTree(sessionID string) (*domain.BranchTree, error) {
	m.calls["GetBranchTree"]++
	if m.err != nil {
		return nil, m.err
	}

	// Find the session
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Create the tree node
	tree := &domain.BranchTree{
		Session:  session.ToSessionInfo(),
		Children: make([]*domain.BranchTree, 0),
	}

	// Recursively build the tree
	for _, childID := range session.ChildIDs {
		if childTree, err := m.GetBranchTree(childID); err == nil {
			tree.Children = append(tree.Children, childTree)
		}
	}

	return tree, nil
}

// MergeSessions merges two sessions according to the specified options
func (m *MockStorageBackend) MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error) {
	m.calls["MergeSessions"]++
	if m.err != nil {
		return nil, m.err
	}

	// Load both sessions
	targetSession, ok := m.sessions[targetID]
	if !ok {
		return nil, fmt.Errorf("target session not found: %s", targetID)
	}

	sourceSession, ok := m.sessions[sourceID]
	if !ok {
		return nil, fmt.Errorf("source session not found: %s", sourceID)
	}

	// Execute the merge
	mergedSession, result, err := targetSession.ExecuteMerge(sourceSession, options)
	if err != nil {
		return nil, err
	}

	// Save the merged session
	m.sessions[mergedSession.ID] = mergedSession

	// Update parent if needed
	if options.CreateBranch {
		m.sessions[targetID] = targetSession
	}

	return result, nil
}
