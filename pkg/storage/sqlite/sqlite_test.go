// ABOUTME: Tests for the SQLite storage backend implementation
// ABOUTME: Ensures proper database operations, schema creation, and FTS search

//go:build (sqlite || db) && integration
// +build sqlite db
// +build integration

package sqlite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test with default db_path
	tmpDir := t.TempDir()
	config := storage.Config{
		"base_dir": tmpDir,
	}

	backend, err := New(config)
	require.NoError(t, err)
	assert.NotNil(t, backend)

	// Verify backend is created successfully
	b := backend.(*Backend)
	assert.NotNil(t, b.db)
	assert.NotEmpty(t, b.userID)

	require.NoError(t, backend.Close())

	// Test with custom db_path
	customPath := filepath.Join(tmpDir, "custom.db")
	config["db_path"] = customPath

	backend2, err := New(config)
	require.NoError(t, err)
	assert.NotNil(t, backend2)

	// Verify backend is created successfully
	b2 := backend2.(*Backend)
	assert.NotNil(t, b2.db)
	assert.NotEmpty(t, b2.userID)

	require.NoError(t, backend2.Close())
}

func TestBackend_Initialize(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Check if tables exist
	rows, err := backend.db.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name IN ('sessions', 'messages', 'attachments', 'tags', 'session_tags')
		ORDER BY name
	`)
	require.NoError(t, err)
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}

	assert.ElementsMatch(t, []string{"messages", "sessions", "tags"}, tables)
}

func TestBackend_NewSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	session := backend.NewSession("Test Session")
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "Test Session", session.Name)
	assert.NotZero(t, session.Created)
	assert.NotZero(t, session.Updated)
	assert.Empty(t, session.Conversation.Messages)
	assert.Empty(t, session.Tags)
}

func TestBackend_SaveAndLoadSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create a session with all fields
	now := time.Now()
	session := &domain.Session{
		ID:   "test-session-001",
		Name: "Test Session",
		Conversation: &domain.Conversation{
			ID: "test-session-001",
			Messages: []domain.Message{
				{
					ID:        "msg-1",
					Role:      domain.MessageRoleUser,
					Content:   "Hello",
					Timestamp: now,
					Attachments: []domain.Attachment{
						{
							Type:     "file",
							Name:     "test.txt",
							MimeType: "text/plain",
						},
					},
					Metadata: map[string]interface{}{
						"source": "test",
					},
				},
				{
					ID:        "msg-2",
					Role:      domain.MessageRoleAssistant,
					Content:   "Hi there!",
					Timestamp: now,
				},
			},
			Model:        "gpt-4",
			Provider:     "openai",
			Temperature:  0.7,
			MaxTokens:    1000,
			SystemPrompt: "You are helpful",
			Created:      now,
			Updated:      now,
		},
		Config: map[string]interface{}{
			"setting": "value",
		},
		Created:  now,
		Updated:  now,
		Tags:     []string{"test", "example"},
		Metadata: map[string]interface{}{"meta": "data"},
	}

	// Save session
	err := backend.Create(session)
	require.NoError(t, err)

	// Load session
	loaded, err := backend.Get(session.ID)
	require.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.Name, loaded.Name)
	assert.Len(t, loaded.Conversation.Messages, 2)
	assert.Equal(t, session.Conversation.Model, loaded.Conversation.Model)
	assert.Equal(t, session.Conversation.Provider, loaded.Conversation.Provider)
	assert.Equal(t, session.Conversation.Temperature, loaded.Conversation.Temperature)
	assert.Equal(t, session.Conversation.MaxTokens, loaded.Conversation.MaxTokens)
	assert.Equal(t, session.Conversation.SystemPrompt, loaded.Conversation.SystemPrompt)
	assert.ElementsMatch(t, session.Tags, loaded.Tags)

	// Check message details
	assert.Equal(t, session.Conversation.Messages[0].ID, loaded.Conversation.Messages[0].ID)
	assert.Equal(t, session.Conversation.Messages[0].Role, loaded.Conversation.Messages[0].Role)
	assert.Len(t, loaded.Conversation.Messages[0].Attachments, 1)
}

func TestBackend_UpdateSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create and save initial session
	session := backend.NewSession("Original Name")
	session.Tags = []string{"original"}
	require.NoError(t, backend.Create(session))

	// Update session
	session.Name = "Updated Name"
	session.Tags = []string{"updated", "modified"}
	session.Conversation.AddMessage(domain.Message{
		ID:        "msg-1",
		Role:      domain.MessageRoleUser,
		Content:   "New message",
		Timestamp: time.Now(),
	})

	// Update session
	require.NoError(t, backend.Update(session))

	// Load and verify
	loaded, err := backend.Get(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", loaded.Name)
	assert.ElementsMatch(t, []string{"updated", "modified"}, loaded.Tags)
	assert.Len(t, loaded.Conversation.Messages, 1)
}

func TestBackend_DeleteSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create and save session
	session := backend.NewSession("To Delete")
	require.NoError(t, backend.Create(session))

	// Verify it exists
	_, err := backend.Get(session.ID)
	require.NoError(t, err)

	// Delete session
	require.NoError(t, backend.Delete(session.ID))

	// Verify it's gone
	_, err = backend.Get(session.ID)
	assert.Error(t, err)
}

func TestBackend_ListSessions(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create multiple sessions
	session1 := backend.NewSession("Session 1")
	session1.Tags = []string{"work", "important"}
	require.NoError(t, backend.Create(session1))

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	session2 := backend.NewSession("Session 2")
	session2.Tags = []string{"personal"}
	require.NoError(t, backend.Create(session2))

	// List sessions
	sessions, err := backend.List()
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Should be ordered by updated time descending
	assert.Equal(t, session2.ID, sessions[0].ID)
	assert.Equal(t, session1.ID, sessions[1].ID)
}

func TestBackend_SearchSessions(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create sessions with searchable content
	session1 := backend.NewSession("Python Tutorial")
	session1.Conversation.SystemPrompt = "You are a Python expert"
	session1.Conversation.AddMessage(domain.Message{
		ID:        "msg-1",
		Role:      domain.MessageRoleUser,
		Content:   "How do I use list comprehensions in Python?",
		Timestamp: time.Now(),
	})
	session1.Conversation.AddMessage(domain.Message{
		ID:        "msg-2",
		Role:      domain.MessageRoleAssistant,
		Content:   "List comprehensions in Python provide a concise way to create lists.",
		Timestamp: time.Now(),
	})
	session1.Tags = []string{"python", "programming"}
	require.NoError(t, backend.Create(session1))

	session2 := backend.NewSession("JavaScript Guide")
	session2.Conversation.AddMessage(domain.Message{
		ID:        "msg-3",
		Role:      domain.MessageRoleUser,
		Content:   "What is closure in JavaScript?",
		Timestamp: time.Now(),
	})
	session2.Tags = []string{"javascript", "web"}
	require.NoError(t, backend.Create(session2))

	// Search for Python
	results, err := backend.Search("python")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, session1.ID, results[0].Session.ID)
	assert.True(t, len(results[0].Matches) > 0)

	// Search for programming
	results, err = backend.Search("programming")
	require.NoError(t, err)
	assert.Len(t, results, 1)

	// Search for something in messages
	results, err = backend.Search("comprehensions")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, session1.ID, results[0].Session.ID)

	// Search for non-existent term
	results, err = backend.Search("nonexistent")
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestBackend_ExportSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create a session
	session := backend.NewSession("Export Test")
	session.Conversation.AddMessage(domain.Message{
		ID:        "msg-1",
		Role:      domain.MessageRoleUser,
		Content:   "Test message",
		Timestamp: time.Now(),
	})
	require.NoError(t, backend.Create(session))

	// Test JSON export
	var jsonBuf bytes.Buffer
	err := backend.ExportSession(session.ID, domain.ExportFormatJSON, &jsonBuf)
	require.NoError(t, err)

	// Verify JSON is valid
	var exported domain.Session
	require.NoError(t, json.Unmarshal(jsonBuf.Bytes(), &exported))
	assert.Equal(t, session.ID, exported.ID)

	// Test Markdown export
	var mdBuf bytes.Buffer
	err = backend.ExportSession(session.ID, domain.ExportFormatMarkdown, &mdBuf)
	require.NoError(t, err)
	assert.Contains(t, mdBuf.String(), "# Session: Export Test")
	assert.Contains(t, mdBuf.String(), "Test message")
}

func TestBackend_ConcurrentAccess(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create a session
	session := backend.NewSession("Concurrent Test")
	require.NoError(t, backend.Create(session))

	// Simulate concurrent access
	done := make(chan bool, 3)

	// Reader 1
	go func() {
		for i := 0; i < 5; i++ {
			_, err := backend.Get(session.ID)
			assert.NoError(t, err)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Reader 2
	go func() {
		for i := 0; i < 5; i++ {
			_, err := backend.List()
			assert.NoError(t, err)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Writer
	go func() {
		for i := 0; i < 5; i++ {
			session.Conversation.AddMessage(domain.Message{
				ID:        fmt.Sprintf("msg-%d", i),
				Role:      domain.MessageRoleUser,
				Content:   fmt.Sprintf("Message %d", i),
				Timestamp: time.Now(),
			})
			assert.NoError(t, backend.Update(session))
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

// Helper function to setup test backend
func setupTestBackend(t *testing.T) *Backend {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := storage.Config{
		"base_dir": tmpDir,
		"db_path":  dbPath,
	}

	backend, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, backend)

	return backend.(*Backend)
}
