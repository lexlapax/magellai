// ABOUTME: Tests for the storage backend interface
// ABOUTME: Ensures the interface is properly defined and constants are correct

package storage

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestBackendType_Constants(t *testing.T) {
	// Verify backend type constants
	assert.Equal(t, BackendType("filesystem"), FileSystemBackend)
	assert.Equal(t, BackendType("sqlite"), SQLiteBackend)
	assert.Equal(t, BackendType("postgresql"), PostgreSQLBackend)
	assert.Equal(t, BackendType("memory"), MemoryBackend)
}

func TestConfig_Type(t *testing.T) {
	// Test Config type
	config := Config{
		"base_dir": "/tmp/test",
		"db_path":  "/tmp/test.db",
		"user_id":  "test-user",
	}

	// Verify type assertions work
	baseDir, ok := config["base_dir"].(string)
	assert.True(t, ok)
	assert.Equal(t, "/tmp/test", baseDir)

	dbPath, ok := config["db_path"].(string)
	assert.True(t, ok)
	assert.Equal(t, "/tmp/test.db", dbPath)
}

// MockBackend implements the Backend interface for testing
type MockBackend struct {
	sessions map[string]*domain.Session
	closed   bool
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		sessions: make(map[string]*domain.Session),
	}
}

func (mb *MockBackend) NewSession(name string) *domain.Session {
	session := domain.NewSession(GenerateSessionID())
	session.Name = name
	return session
}

func (mb *MockBackend) SaveSession(session *domain.Session) error {
	mb.sessions[session.ID] = session
	return nil
}

func (mb *MockBackend) LoadSession(id string) (*domain.Session, error) {
	if session, ok := mb.sessions[id]; ok {
		return session, nil
	}
	return nil, nil
}

func (mb *MockBackend) ListSessions() ([]*domain.SessionInfo, error) {
	var sessions []*domain.SessionInfo
	for _, session := range mb.sessions {
		sessions = append(sessions, session.ToSessionInfo())
	}
	return sessions, nil
}

func (mb *MockBackend) DeleteSession(id string) error {
	delete(mb.sessions, id)
	return nil
}

func (mb *MockBackend) SearchSessions(query string) ([]*domain.SearchResult, error) {
	return []*domain.SearchResult{}, nil
}

func (mb *MockBackend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	return nil
}

func (mb *MockBackend) Close() error {
	mb.closed = true
	return nil
}

func (mb *MockBackend) GetChildren(sessionID string) ([]*domain.SessionInfo, error) {
	var children []*domain.SessionInfo
	if session, exists := mb.sessions[sessionID]; exists {
		for _, childID := range session.ChildIDs {
			if child, exists := mb.sessions[childID]; exists {
				children = append(children, child.ToSessionInfo())
			}
		}
	}
	return children, nil
}

func (mb *MockBackend) GetBranchTree(sessionID string) (*domain.BranchTree, error) {
	session, exists := mb.sessions[sessionID]
	if !exists {
		return nil, nil
	}

	tree := &domain.BranchTree{
		Session:  session.ToSessionInfo(),
		Children: make([]*domain.BranchTree, 0),
	}

	for _, childID := range session.ChildIDs {
		if childTree, err := mb.GetBranchTree(childID); err == nil && childTree != nil {
			tree.Children = append(tree.Children, childTree)
		}
	}

	return tree, nil
}

func (mb *MockBackend) MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error) {
	targetSession, exists := mb.sessions[targetID]
	if !exists {
		return nil, nil
	}

	sourceSession, exists := mb.sessions[sourceID]
	if !exists {
		return nil, nil
	}

	// Execute the merge
	mergedSession, result, err := targetSession.ExecuteMerge(sourceSession, options)
	if err != nil {
		return nil, err
	}

	// Save the merged session
	mb.sessions[mergedSession.ID] = mergedSession

	// Update parent if needed
	if options.CreateBranch {
		mb.sessions[targetID] = targetSession
	}

	return result, nil
}

func TestMockBackend_Implementation(t *testing.T) {
	// This test ensures MockBackend properly implements the Backend interface
	var _ Backend = (*MockBackend)(nil)

	mock := NewMockBackend()

	// Test NewSession
	session := mock.NewSession("test")
	assert.NotNil(t, session)
	assert.Equal(t, "test", session.Name)

	// Test SaveSession
	err := mock.SaveSession(session)
	assert.NoError(t, err)

	// Test LoadSession
	loaded, err := mock.LoadSession(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)

	// Test ListSessions
	sessions, err := mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)

	// Test DeleteSession
	err = mock.DeleteSession(session.ID)
	assert.NoError(t, err)

	sessions, err = mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 0)

	// Test Close
	err = mock.Close()
	assert.NoError(t, err)
	assert.True(t, mock.closed)
}

// TestBackend_InterfaceCompliance verifies all Backend implementations comply with the interface
func TestBackend_InterfaceCompliance(t *testing.T) {
	// Ensure various backends implement the interface
	var _ Backend = (*MockBackend)(nil)
	// Add other backends as they're implemented
}

// TestBackend_SearchSessions tests the search functionality
func TestBackend_SearchSessions(t *testing.T) {
	mock := NewMockBackend()
	
	// Create test sessions
	session1 := mock.NewSession("session1")
	session1.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleUser,
		Content: "Find information about golang testing",
	})
	mock.SaveSession(session1)
	
	session2 := mock.NewSession("session2")
	session2.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleUser,
		Content: "Explain python decorators",
	})
	mock.SaveSession(session2)
	
	// Test search
	results, err := mock.SearchSessions("golang")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	
	// Test empty search
	results, err = mock.SearchSessions("")
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

// TestBackend_ExportSession tests session export functionality
func TestBackend_ExportSession(t *testing.T) {
	mock := NewMockBackend()
	
	// Create a test session
	session := mock.NewSession("export-test")
	session.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleUser,
		Content: "Test message",
	})
	mock.SaveSession(session)
	
	// Test export formats
	formats := []domain.ExportFormat{
		domain.ExportFormatJSON,
		domain.ExportFormatMarkdown,
		domain.ExportFormatText,
	}
	
	for _, format := range formats {
		var buf bytes.Buffer
		err := mock.ExportSession(session.ID, format, &buf)
		assert.NoError(t, err)
		// For the mock, we don't expect output, but real implementations should write
	}
	
	// Test export of non-existent session
	var buf bytes.Buffer
	err := mock.ExportSession("non-existent", domain.ExportFormatJSON, &buf)
	assert.NoError(t, err) // Mock doesn't return error, but real impl should
}

// TestBackend_BranchingOperations tests session branching functionality
func TestBackend_BranchingOperations(t *testing.T) {
	mock := NewMockBackend()
	
	// Create parent session
	parent := mock.NewSession("parent")
	mock.SaveSession(parent)
	
	// Create child sessions
	child1, err := parent.CreateBranch("child1-id", "child1", len(parent.Conversation.Messages))
	assert.NoError(t, err)
	mock.SaveSession(child1)
	
	child2, err := parent.CreateBranch("child2-id", "child2", len(parent.Conversation.Messages))
	assert.NoError(t, err)
	mock.SaveSession(child2)
	
	// Update parent with children
	mock.SaveSession(parent)
	
	// Test GetChildren
	children, err := mock.GetChildren(parent.ID)
	assert.NoError(t, err)
	assert.Len(t, children, 2)
	
	// Test GetBranchTree
	tree, err := mock.GetBranchTree(parent.ID)
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	assert.Equal(t, parent.ID, tree.Session.ID)
	assert.Len(t, tree.Children, 2)
	
	// Test GetBranchTree with non-existent session
	tree, err = mock.GetBranchTree("non-existent")
	assert.NoError(t, err)
	assert.Nil(t, tree)
}

// TestBackend_MergeSessions tests session merging functionality
func TestBackend_MergeSessions(t *testing.T) {
	mock := NewMockBackend()
	
	// Create source and target sessions
	target := mock.NewSession("target")
	target.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleUser,
		Content: "Target message",
	})
	mock.SaveSession(target)
	
	source := mock.NewSession("source")
	source.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleUser,
		Content: "Source message",
	})
	mock.SaveSession(source)
	
	// Test merge
	options := domain.MergeOptions{
		Type:         domain.MergeTypeContinuation,
		SourceID:     source.ID,
		TargetID:     target.ID,
		CreateBranch: true,
		BranchName:   "merged-branch",
	}
	
	result, err := mock.MergeSessions(target.ID, source.ID, options)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Test merge with non-existent sessions
	result, err = mock.MergeSessions("non-existent", source.ID, options)
	assert.NoError(t, err)
	assert.Nil(t, result)
	
	result, err = mock.MergeSessions(target.ID, "non-existent", options)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestBackend_EdgeCases tests edge cases and error conditions
func TestBackend_EdgeCases(t *testing.T) {
	mock := NewMockBackend()
	
	// Test LoadSession with non-existent ID
	session, err := mock.LoadSession("non-existent")
	assert.NoError(t, err)
	assert.Nil(t, session)
	
	// Test DeleteSession with non-existent ID
	err = mock.DeleteSession("non-existent")
	assert.NoError(t, err)
	
	// Test ListSessions with empty backend
	sessions, err := mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 0)
	
	// Test multiple Close calls
	err = mock.Close()
	assert.NoError(t, err)
	err = mock.Close()
	assert.NoError(t, err)
}

// TestBackend_SessionLifecycle tests a complete session lifecycle
func TestBackend_SessionLifecycle(t *testing.T) {
	mock := NewMockBackend()
	
	// Create a new session
	session := mock.NewSession("lifecycle-test")
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "lifecycle-test", session.Name)
	assert.NotNil(t, session.Conversation.Messages)
	assert.NotNil(t, session.Metadata)
	
	// Add messages to the session
	session.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleUser,
		Content: "Hello",
	})
	session.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleAssistant,
		Content: "Hi there!",
	})
	
	// Save the session
	err := mock.SaveSession(session)
	assert.NoError(t, err)
	
	// Load the session back
	loaded, err := mock.LoadSession(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)
	assert.Len(t, loaded.Conversation.Messages, 2)
	
	// Update the session
	loaded.Conversation.AddMessage(domain.Message{
		Role:    domain.MessageRoleUser,
		Content: "Another message",
	})
	err = mock.SaveSession(loaded)
	assert.NoError(t, err)
	
	// List sessions should show the session
	sessions, err := mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, session.ID, sessions[0].ID)
	
	// Delete the session
	err = mock.DeleteSession(session.ID)
	assert.NoError(t, err)
	
	// Session should no longer exist
	loaded, err = mock.LoadSession(session.ID)
	assert.NoError(t, err)
	assert.Nil(t, loaded)
	
	// List should be empty
	sessions, err = mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 0)
}

// TestBackend_ConcurrentAccess tests concurrent access patterns
func TestBackend_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}
	
	mock := NewMockBackend()
	sessionCount := 10
	
	// Create multiple sessions concurrently
	done := make(chan bool, sessionCount)
	
	for i := 0; i < sessionCount; i++ {
		go func(n int) {
			session := mock.NewSession(fmt.Sprintf("concurrent-%d", n))
			err := mock.SaveSession(session)
			assert.NoError(t, err)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < sessionCount; i++ {
		<-done
	}
	
	// Verify all sessions were created
	sessions, err := mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, sessionCount)
}

// BenchmarkBackend_SaveSession benchmarks session save performance
func BenchmarkBackend_SaveSession(b *testing.B) {
	mock := NewMockBackend()
	session := mock.NewSession("bench-test")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mock.SaveSession(session)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBackend_LoadSession benchmarks session load performance
func BenchmarkBackend_LoadSession(b *testing.B) {
	mock := NewMockBackend()
	session := mock.NewSession("bench-test")
	mock.SaveSession(session)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mock.LoadSession(session.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBackend_ListSessions benchmarks listing sessions
func BenchmarkBackend_ListSessions(b *testing.B) {
	mock := NewMockBackend()
	
	// Create 100 sessions
	for i := 0; i < 100; i++ {
		session := mock.NewSession(fmt.Sprintf("bench-%d", i))
		mock.SaveSession(session)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mock.ListSessions()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestBackend_GenerateSessionID tests the session ID generation
func TestBackend_GenerateSessionID(t *testing.T) {
	// Test that IDs are unique
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateSessionID()
		assert.NotEmpty(t, id)
		assert.False(t, ids[id], "Duplicate ID generated: %s", id)
		ids[id] = true
	}
	
	// Test ID format
	id := GenerateSessionID()
	assert.NotEmpty(t, id)
	// ID should be a valid string
	assert.IsType(t, "", id)
}
