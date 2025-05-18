// ABOUTME: Tests for the storage backend interface
// ABOUTME: Ensures the interface is properly defined and constants are correct

package storage

import (
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
