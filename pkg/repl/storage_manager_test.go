// ABOUTME: Tests for the storage manager that wraps the storage backend abstraction
// ABOUTME: Ensures proper REPL session management with type conversions and error handling

package repl

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorageManager(t *testing.T) {
	// Test with valid backend
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, backend, manager.backend)

	// Test with nil backend
	manager, err = NewStorageManager(nil)
	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "storage backend cannot be nil")
}

func TestStorageManager_NewSession(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	session := manager.NewSession("Test Session")
	assert.NotNil(t, session)
	assert.Equal(t, "Test Session", session.Name)
	assert.NotEmpty(t, session.ID)
	assert.NotNil(t, session.Conversation)
	assert.Equal(t, 1, backend.calls["NewSession"])
}

func TestStorageManager_SaveSession(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create and save a session
	session := &domain.Session{
		ID:   "test-123",
		Name: "Test Session",
		Conversation: &domain.Conversation{
			ID: "test-123",
			Messages: []domain.Message{
				{
					ID:        "msg-1",
					Role:      "user",
					Content:   "Hello",
					Timestamp: time.Now(),
				},
			},
		},
		Created: time.Now(),
		Updated: time.Now(),
		Tags:    []string{"test"},
	}

	err = manager.SaveSession(session)
	assert.NoError(t, err)
	assert.Equal(t, 1, backend.calls["SaveSession"])

	// Verify session was saved correctly
	savedSession, ok := backend.sessions["test-123"]
	assert.True(t, ok)
	assert.Equal(t, session.ID, savedSession.ID)
	assert.Equal(t, session.Name, savedSession.Name)
	assert.NotNil(t, savedSession.Conversation)
	assert.Len(t, savedSession.Conversation.Messages, 1)

	// Test error case
	backend.err = fmt.Errorf("save error")
	err = manager.SaveSession(session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save error")
}

func TestStorageManager_LoadSession(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create a session in the backend
	domainSession := &domain.Session{
		ID:   "test-123",
		Name: "Test Session",
		Conversation: &domain.Conversation{
			Messages: []domain.Message{
				{
					ID:        "msg-1",
					Role:      domain.MessageRoleUser,
					Content:   "Hello",
					Timestamp: time.Now(),
				},
			},
		},
		Created: time.Now(),
		Updated: time.Now(),
		Tags:    []string{"test"},
	}
	backend.sessions["test-123"] = domainSession

	// Load the session
	session, err := manager.LoadSession("test-123")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "test-123", session.ID)
	assert.Equal(t, "Test Session", session.Name)
	assert.NotNil(t, session.Conversation)
	assert.Len(t, session.Conversation.Messages, 1)
	assert.Equal(t, 1, backend.calls["LoadSession"])

	// Test non-existent session
	_, err = manager.LoadSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")

	// Test error case
	backend.err = fmt.Errorf("load error")
	_, err = manager.LoadSession("test-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load error")
}

func TestStorageManager_ListSessions(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Add some sessions
	for i := 0; i < 3; i++ {
		session := &domain.Session{
			ID:      fmt.Sprintf("session-%d", i),
			Name:    fmt.Sprintf("Session %d", i),
			Created: time.Now(),
			Updated: time.Now(),
			Conversation: &domain.Conversation{
				Messages: make([]domain.Message, i),
			},
			Tags: []string{fmt.Sprintf("tag%d", i)},
		}
		backend.sessions[session.ID] = session
	}

	// List sessions
	infos, err := manager.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, infos, 3)
	assert.Equal(t, 1, backend.calls["ListSessions"])

	// Verify session info conversion
	for _, info := range infos {
		assert.Contains(t, info.ID, "session-")
		assert.Contains(t, info.Name, "Session")
		assert.GreaterOrEqual(t, info.MessageCount, 0)
	}

	// Test error case
	backend.err = fmt.Errorf("list error")
	_, err = manager.ListSessions()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list error")
}

func TestStorageManager_DeleteSession(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Add a session
	session := &domain.Session{
		ID:   "delete-test",
		Name: "Delete Test",
	}
	backend.sessions["delete-test"] = session

	// Delete the session
	err = manager.DeleteSession("delete-test")
	assert.NoError(t, err)
	assert.Equal(t, 1, backend.calls["DeleteSession"])

	// Verify deletion
	_, ok := backend.sessions["delete-test"]
	assert.False(t, ok)

	// Test non-existent session
	err = manager.DeleteSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")

	// Test error case
	backend.err = fmt.Errorf("delete error")
	err = manager.DeleteSession("delete-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestStorageManager_SearchSessions(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Add test sessions
	sessions := []struct {
		id   string
		name string
	}{
		{"search-1", "Python Programming"},
		{"search-2", "JavaScript Tutorial"},
		{"search-3", "Go Programming"},
	}

	for _, s := range sessions {
		session := &domain.Session{
			ID:   s.id,
			Name: s.name,
		}
		backend.sessions[s.id] = session
	}

	// Search for "Programming"
	results, err := manager.SearchSessions("Programming")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, 1, backend.calls["SearchSessions"])

	// Verify results are properly converted
	for _, result := range results {
		assert.NotNil(t, result.Session)
		assert.NotEmpty(t, result.Matches)
		assert.Contains(t, result.Session.Name, "Programming")
	}

	// Test error case
	backend.err = fmt.Errorf("search error")
	_, err = manager.SearchSessions("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search error")
}

func TestStorageManager_ExportSession(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Add a session
	session := &domain.Session{
		ID:   "export-test",
		Name: "Export Test",
	}
	backend.sessions["export-test"] = session

	// Test export
	var buf bytes.Buffer
	err = manager.ExportSession("export-test", "json", &buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, backend.calls["ExportSession"])
	assert.Contains(t, buf.String(), "export-test")

	// Test non-existent session
	buf.Reset()
	err = manager.ExportSession("non-existent", "json", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")

	// Test error case
	backend.err = fmt.Errorf("export error")
	buf.Reset()
	err = manager.ExportSession("export-test", "json", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "export error")
}

func TestCreateStorageManager(t *testing.T) {
	// Skip this test as it requires the storage backends to be registered
	// which happens in their init() functions
	t.Skip("Skipping test that requires filesystem backend initialization")
}

func TestStorageManager_Close(t *testing.T) {
	backend := NewMockStorageBackend()
	manager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Test close
	err = manager.Close()
	assert.NoError(t, err)
	assert.Equal(t, 1, backend.calls["Close"])

	// Test error case
	backend.err = fmt.Errorf("close error")
	err = manager.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close error")
}
