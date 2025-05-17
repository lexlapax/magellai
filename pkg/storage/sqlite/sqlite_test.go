// ABOUTME: Tests for the SQLite storage backend implementation
// ABOUTME: Ensures proper database operations, schema creation, and FTS search

//go:build sqlite || db

package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test with default db_path
	tempDir := t.TempDir()
	config := storage.Config{}

	// Temporarily override home directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	backend, err := New(config)
	require.NoError(t, err)
	assert.NotNil(t, backend)

	// Type assert to get the concrete type
	sqliteBackend, ok := backend.(*Backend)
	require.True(t, ok, "Backend should be of type *sqlite.Backend")
	assert.NotEmpty(t, sqliteBackend.userID)

	backend.Close()

	// Test with custom db_path
	dbPath := filepath.Join(tempDir, "custom.db")
	config = storage.Config{
		"db_path": dbPath,
	}

	backend, err = New(config)
	require.NoError(t, err)
	assert.NotNil(t, backend)
	assert.FileExists(t, dbPath)
	backend.Close()
}

func TestBackend_InitSchema(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Check that tables exist
	tables := []string{"sessions", "messages", "tags"}
	for _, table := range tables {
		var name string
		err := backend.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, table, name)
	}

	// Check indexes exist
	indexes := []string{"idx_sessions_user", "idx_messages_session", "idx_tags_session"}
	for _, idx := range indexes {
		var name string
		err := backend.db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&name)
		require.NoError(t, err)
		assert.Equal(t, idx, name)
	}
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
	assert.Empty(t, session.Messages)
	assert.Empty(t, session.Tags)
}

func TestBackend_SaveAndLoadSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create a session with all fields
	session := &storage.Session{
		ID:   "test-session-001",
		Name: "Test Session",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Hello",
				Timestamp: time.Now(),
				Attachments: []storage.Attachment{
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
				Role:      "assistant",
				Content:   "Hi there!",
				Timestamp: time.Now(),
			},
		},
		Config: map[string]interface{}{
			"setting": "value",
		},
		Created:      time.Now(),
		Updated:      time.Now(),
		Tags:         []string{"test", "example"},
		Metadata:     map[string]interface{}{"meta": "data"},
		Model:        "gpt-4",
		Provider:     "openai",
		Temperature:  0.7,
		MaxTokens:    1000,
		SystemPrompt: "You are helpful",
	}

	// Save session
	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Load session
	loaded, err := backend.LoadSession(session.ID)
	require.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.Name, loaded.Name)
	assert.Len(t, loaded.Messages, 2)
	assert.Equal(t, session.Model, loaded.Model)
	assert.Equal(t, session.Provider, loaded.Provider)
	assert.Equal(t, session.Temperature, loaded.Temperature)
	assert.Equal(t, session.MaxTokens, loaded.MaxTokens)
	assert.Equal(t, session.SystemPrompt, loaded.SystemPrompt)
	assert.ElementsMatch(t, session.Tags, loaded.Tags) // Use ElementsMatch instead of Equal since order is not guaranteed

	// Check message details
	assert.Equal(t, session.Messages[0].ID, loaded.Messages[0].ID)
	assert.Equal(t, session.Messages[0].Role, loaded.Messages[0].Role)
	assert.Len(t, loaded.Messages[0].Attachments, 1)
}

func TestBackend_UpdateSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create and save initial session
	session := backend.NewSession("Update Test")
	session.Tags = []string{"initial"}
	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Update session
	session.Name = "Updated Session"
	session.Tags = []string{"updated", "modified"}
	session.Messages = append(session.Messages, storage.Message{
		ID:        "msg-new",
		Role:      "user",
		Content:   "New message",
		Timestamp: time.Now(),
	})

	err = backend.SaveSession(session)
	require.NoError(t, err)

	// Load and verify updates
	loaded, err := backend.LoadSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Session", loaded.Name)
	assert.ElementsMatch(t, []string{"updated", "modified"}, loaded.Tags) // Use ElementsMatch since order is not guaranteed
	assert.Len(t, loaded.Messages, 1)
}

func TestBackend_MultiTenant(t *testing.T) {
	// Create two backends with different user IDs
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "multi-tenant.db")

	// Backend 1
	backend1 := &Backend{
		userID: "user1",
	}
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	backend1.db = db
	err = backend1.initSchema()
	require.NoError(t, err)

	// Backend 2
	backend2 := &Backend{
		userID: "user2",
		db:     db,
	}

	// Create sessions for each user
	session1 := backend1.NewSession("User1 Session")
	err = backend1.SaveSession(session1)
	require.NoError(t, err)

	session2 := backend2.NewSession("User2 Session")
	err = backend2.SaveSession(session2)
	require.NoError(t, err)

	// Verify isolation
	sessions1, err := backend1.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessions1, 1)
	assert.Equal(t, "User1 Session", sessions1[0].Name)

	sessions2, err := backend2.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessions2, 1)
	assert.Equal(t, "User2 Session", sessions2[0].Name)

	// Cleanup
	backend1.Close()
}

func TestBackend_ListSessions(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		session := backend.NewSession(string(rune('A' + i)))
		session.Tags = []string{"tag" + string(rune('1'+i))}
		// Add messages to get message count
		for j := 0; j <= i; j++ {
			session.Messages = append(session.Messages, storage.Message{
				ID:        fmt.Sprintf("msg-%d-%d", i, j),
				Role:      "user",
				Content:   "Message",
				Timestamp: time.Now(),
			})
		}
		err := backend.SaveSession(session)
		require.NoError(t, err)
	}

	// List sessions
	sessions, err := backend.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessions, 3)

	// Check ordering (by updated DESC)
	assert.Equal(t, "C", sessions[0].Name)
	assert.Equal(t, "B", sessions[1].Name)
	assert.Equal(t, "A", sessions[2].Name)

	// Check message counts
	assert.Equal(t, 3, sessions[0].MessageCount)
	assert.Equal(t, 2, sessions[1].MessageCount)
	assert.Equal(t, 1, sessions[2].MessageCount)

	// Check tags
	assert.Equal(t, []string{"tag3"}, sessions[0].Tags)
}

func TestBackend_DeleteSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create and save session
	session := backend.NewSession("Delete Test")
	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Verify it exists
	loaded, err := backend.LoadSession(session.ID)
	require.NoError(t, err)
	assert.NotNil(t, loaded)

	// Delete session
	err = backend.DeleteSession(session.ID)
	require.NoError(t, err)

	// Verify it's gone
	loaded, err = backend.LoadSession(session.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")

	// Try to delete non-existent session
	err = backend.DeleteSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestBackend_SearchSessions(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create sessions with searchable content
	sessions := []*storage.Session{
		{
			ID:           "search-1",
			Name:         "Python Programming",
			SystemPrompt: "You are a Python expert",
			Messages: []storage.Message{
				{
					ID:        "msg-1",
					Role:      "user",
					Content:   "How do I use Python decorators?",
					Timestamp: time.Now(),
				},
				{
					ID:        "msg-2",
					Role:      "assistant",
					Content:   "Python decorators are functions that modify other functions",
					Timestamp: time.Now(),
				},
			},
			Tags:    []string{"python", "programming"},
			Created: time.Now(),
			Updated: time.Now(),
		},
		{
			ID:   "search-2",
			Name: "JavaScript Tutorial",
			Messages: []storage.Message{
				{
					ID:        "msg-3",
					Role:      "user",
					Content:   "What is JavaScript closure?",
					Timestamp: time.Now(),
				},
			},
			Tags:    []string{"javascript", "web"},
			Created: time.Now(),
			Updated: time.Now(),
		},
	}

	for _, s := range sessions {
		err := backend.SaveSession(s)
		require.NoError(t, err)
	}

	// Search for "Python"
	results, err := backend.SearchSessions("Python")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-1", results[0].Session.ID)

	// Check match types - at least some matches should be found
	matchTypes := make(map[string]bool)
	for _, match := range results[0].Matches {
		matchTypes[match.Type] = true
	}
	assert.True(t, len(matchTypes) > 0, "Should find at least one match type")

	// Search for "programming" (should find in tags)
	results, err = backend.SearchSessions("programming")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-1", results[0].Session.ID)

	// Search case-insensitive
	results, err = backend.SearchSessions("JAVASCRIPT")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-2", results[0].Session.ID)
}

func TestBackend_ExportSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Create a session
	session := &storage.Session{
		ID:   "export-test",
		Name: "Export Test Session",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Test message",
				Timestamp: time.Now(),
				Attachments: []storage.Attachment{
					{
						Type: "file",
						Name: "test.txt",
					},
				},
			},
		},
		Tags:    []string{"export", "test"},
		Created: time.Now(),
		Updated: time.Now(),
	}

	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Test JSON export
	t.Run("JSON Export", func(t *testing.T) {
		var buf bytes.Buffer
		err := backend.ExportSession(session.ID, storage.ExportFormatJSON, &buf)
		require.NoError(t, err)

		var exported storage.Session
		err = json.Unmarshal(buf.Bytes(), &exported)
		require.NoError(t, err)
		assert.Equal(t, session.ID, exported.ID)
		assert.Equal(t, session.Name, exported.Name)
	})

	// Test Markdown export
	t.Run("Markdown Export", func(t *testing.T) {
		var buf bytes.Buffer
		err := backend.ExportSession(session.ID, storage.ExportFormatMarkdown, &buf)
		require.NoError(t, err)

		markdown := buf.String()
		assert.Contains(t, markdown, "# Session: Export Test Session")
		assert.Contains(t, markdown, "### User")
		assert.Contains(t, markdown, "Test message")
		assert.Contains(t, markdown, "Attachments:")
	})

	// Test unsupported format
	t.Run("Unsupported Format", func(t *testing.T) {
		var buf bytes.Buffer
		err := backend.ExportSession(session.ID, "unsupported", &buf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported export format")
	})
}

func TestBackend_LoadNonExistentSession(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	session, err := backend.LoadSession("non-existent")
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "session not found")
}

func TestBackend_Close(t *testing.T) {
	backend := setupTestBackend(t)

	// Should be able to use the backend
	session := backend.NewSession("Test")
	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Close the backend
	err = backend.Close()
	require.NoError(t, err)

	// Should not be able to use after close
	err = backend.SaveSession(session)
	assert.Error(t, err)
}

func TestBackend_FTS_Fallback(t *testing.T) {
	backend := setupTestBackend(t)
	defer backend.Close()

	// Drop FTS table to simulate system without FTS5
	_, err := backend.db.Exec("DROP TABLE IF EXISTS messages_fts")
	require.NoError(t, err)

	// Create test data
	session := &storage.Session{
		ID:   "fts-test",
		Name: "FTS Test",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Testing without FTS5 support",
				Timestamp: time.Now(),
			},
		},
		Created: time.Now(),
		Updated: time.Now(),
	}

	err = backend.SaveSession(session)
	require.NoError(t, err)

	// Search should still work with LIKE fallback
	results, err := backend.SearchSessions("without")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, session.ID, results[0].Session.ID)
}

// Helper function to set up a test backend
func setupTestBackend(t *testing.T) *Backend {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := storage.Config{
		"db_path": dbPath,
	}

	backend, err := New(config)
	require.NoError(t, err)

	// Type assert to get the concrete type
	sqliteBackend, ok := backend.(*Backend)
	require.True(t, ok, "Backend should be of type *sqlite.Backend")

	return sqliteBackend
}
