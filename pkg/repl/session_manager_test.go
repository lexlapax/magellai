// ABOUTME: Tests for the session manager that provides high-level session management
// ABOUTME: Ensures proper session lifecycle management with the storage backend abstraction

package repl

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorageManager is a mock implementation of StorageManager for testing
type MockStorageManager struct {
	newSessionFunc    func(name string) *domain.Session
	saveSessionFunc   func(session *domain.Session) error
	loadSessionFunc   func(id string) (*domain.Session, error)
	listSessionsFunc  func() ([]*domain.SessionInfo, error)
	deleteSessionFunc func(id string) error
	searchSessionFunc func(query string) ([]*domain.SearchResult, error)
	exportSessionFunc func(id string, format string, w io.Writer) error
	closeFunc         func() error
	calls             map[string]int
}

func NewMockStorageManager() *MockStorageManager {
	return &MockStorageManager{
		calls: make(map[string]int),
	}
}

func (m *MockStorageManager) NewSession(name string) *domain.Session {
	m.calls["NewSession"]++
	if m.newSessionFunc != nil {
		return m.newSessionFunc(name)
	}
	return &domain.Session{
		ID:   "test-session-123",
		Name: name,
		Conversation: &domain.Conversation{
			ID:       "test-session-123",
			Messages: []domain.Message{},
		},
		Config:   make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
		Tags:     []string{},
	}
}

func (m *MockStorageManager) SaveSession(session *domain.Session) error {
	m.calls["SaveSession"]++
	if m.saveSessionFunc != nil {
		return m.saveSessionFunc(session)
	}
	return nil
}

func (m *MockStorageManager) LoadSession(id string) (*domain.Session, error) {
	m.calls["LoadSession"]++
	if m.loadSessionFunc != nil {
		return m.loadSessionFunc(id)
	}
	return nil, fmt.Errorf("session not found")
}

func (m *MockStorageManager) ListSessions() ([]*domain.SessionInfo, error) {
	m.calls["ListSessions"]++
	if m.listSessionsFunc != nil {
		return m.listSessionsFunc()
	}
	return []*domain.SessionInfo{}, nil
}

func (m *MockStorageManager) DeleteSession(id string) error {
	m.calls["DeleteSession"]++
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(id)
	}
	return nil
}

func (m *MockStorageManager) SearchSessions(query string) ([]*domain.SearchResult, error) {
	m.calls["SearchSessions"]++
	if m.searchSessionFunc != nil {
		return m.searchSessionFunc(query)
	}
	return []*domain.SearchResult{}, nil
}

func (m *MockStorageManager) ExportSession(id string, format string, w io.Writer) error {
	m.calls["ExportSession"]++
	if m.exportSessionFunc != nil {
		return m.exportSessionFunc(id, format, w)
	}
	return nil
}

func (m *MockStorageManager) Close() error {
	m.calls["Close"]++
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNewSessionManager_StorageAbstraction(t *testing.T) {
	// Test with valid storage manager
	backend := NewMockStorageBackend()
	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	manager, err := NewSessionManager(storageManager)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, storageManager, manager.StorageManager)

	// Test with nil storage manager
	manager, err = NewSessionManager(nil)
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "storage manager cannot be nil")
}

func TestSessionManager_NewSession_StorageAbstraction(t *testing.T) {
	// Test with a real StorageManager and mocked backend
	backend := NewMockStorageBackend()
	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	manager, err := NewSessionManager(storageManager)
	require.NoError(t, err)

	session, err := manager.NewSession("Test Session")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "Test Session", session.Name)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, 1, backend.calls["NewSession"])
	assert.Equal(t, 1, backend.calls["SaveSession"])

	// Test save error
	backend.err = fmt.Errorf("save error")
	session, err = manager.NewSession("Error Session")
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "save error")
}

// Integration test to verify SessionManager works with StorageManager
func TestSessionManager_Integration(t *testing.T) {
	// Create a real storage backend
	backend := NewMockStorageBackend()

	// Create storage manager
	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create session manager
	sessionManager, err := NewSessionManager(storageManager)
	require.NoError(t, err)

	// Test creating a session
	session, err := sessionManager.NewSession("Integration Test")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "Integration Test", session.Name)

	// Verify the session was saved
	assert.Equal(t, 1, backend.calls["NewSession"])
	assert.Equal(t, 1, backend.calls["SaveSession"])

	// Test loading the session
	loaded, err := sessionManager.StorageManager.LoadSession(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.Name, loaded.Name)

	// Test listing sessions
	infos, err := sessionManager.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, infos, 1)
	assert.Equal(t, session.ID, infos[0].ID)

	// Test searching sessions
	results, err := sessionManager.SearchSessions("Integration")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, session.ID, results[0].Session.ID)

	// Test deleting the session
	err = sessionManager.DeleteSession(session.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = sessionManager.StorageManager.LoadSession(session.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// Test that SessionManager properly delegates all methods to StorageManager
func TestSessionManager_Delegation(t *testing.T) {
	backend := NewMockStorageBackend()
	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	sessionManager, err := NewSessionManager(storageManager)
	require.NoError(t, err)

	// Create a session to work with
	session, err := sessionManager.NewSession("Delegation Test")
	require.NoError(t, err)
	backend.sessions[session.ID] = session

	// Test each delegated method
	tests := []struct {
		name   string
		method string
		test   func()
	}{
		{
			name:   "LoadSession",
			method: "LoadSession",
			test: func() {
				_, _ = sessionManager.StorageManager.LoadSession(session.ID)
			},
		},
		{
			name:   "SaveSession",
			method: "SaveSession",
			test: func() {
				_ = sessionManager.SaveSession(session)
			},
		},
		{
			name:   "ListSessions",
			method: "ListSessions",
			test: func() {
				_, _ = sessionManager.ListSessions()
			},
		},
		{
			name:   "DeleteSession",
			method: "DeleteSession",
			test: func() {
				_ = sessionManager.DeleteSession(session.ID)
			},
		},
		{
			name:   "SearchSessions",
			method: "SearchSessions",
			test: func() {
				_, _ = sessionManager.SearchSessions("test")
			},
		},
		{
			name:   "ExportSession",
			method: "ExportSession",
			test: func() {
				var buf bytes.Buffer
				_ = sessionManager.ExportSession(session.ID, "json", &buf)
			},
		},
		{
			name:   "Close",
			method: "Close",
			test: func() {
				_ = sessionManager.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset call counters
			backend.calls = make(map[string]int)

			// Execute the test
			tt.test()

			// Verify the backend method was called
			assert.Greater(t, backend.calls[tt.method], 0,
				"Expected %s to be called through delegation", tt.method)
		})
	}
}
