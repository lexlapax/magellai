// ABOUTME: Tests for the filesystem storage backend implementation
// ABOUTME: Ensures proper file operations, JSON persistence, and search functionality

package filesystem

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test with valid base directory
	tempDir := t.TempDir()
	config := storage.Config{
		"base_dir": tempDir,
	}

	backend, err := New(config)
	require.NoError(t, err)
	assert.NotNil(t, backend)

	// Test with missing base_dir
	backend, err = New(storage.Config{})
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "filesystem backend requires 'base_dir'")

	// Test with empty base_dir
	backend, err = New(storage.Config{"base_dir": ""})
	assert.Error(t, err)
	assert.Nil(t, backend)
}

func TestBackend_NewSession(t *testing.T) {
	backend := setupTestBackend(t)

	session := backend.NewSession("Test Session")
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "Test Session", session.Name)
	assert.NotZero(t, session.Created)
	assert.NotZero(t, session.Updated)
	assert.Empty(t, session.Messages)
	assert.NotNil(t, session.Config)
	assert.NotNil(t, session.Metadata)
}

func TestBackend_SaveAndLoadSession(t *testing.T) {
	backend := setupTestBackend(t)

	// Create a session with data
	session := &storage.Session{
		ID:   "test-session-001",
		Name: "Test Session",
		Messages: []storage.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Hello",
				Timestamp: time.Now(),
			},
			{
				ID:        "msg-2",
				Role:      "assistant",
				Content:   "Hi there!",
				Timestamp: time.Now(),
			},
		},
		Config: map[string]interface{}{
			"temperature": 0.7,
		},
		Created:      time.Now(),
		Updated:      time.Now(),
		Tags:         []string{"test", "example"},
		Model:        "gpt-4",
		Provider:     "openai",
		SystemPrompt: "You are helpful",
	}

	// Save session
	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Verify file exists
	sessionFile := filepath.Join(backend.baseDir, session.ID+".json")
	assert.FileExists(t, sessionFile)

	// Load session
	loaded, err := backend.LoadSession(session.ID)
	require.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.Name, loaded.Name)
	assert.Len(t, loaded.Messages, 2)
	assert.Equal(t, session.Model, loaded.Model)
	assert.Equal(t, session.SystemPrompt, loaded.SystemPrompt)
	assert.Equal(t, session.Tags, loaded.Tags)
}

func TestBackend_LoadNonExistentSession(t *testing.T) {
	backend := setupTestBackend(t)

	session, err := backend.LoadSession("non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "session not found")
}

func TestBackend_ListSessions(t *testing.T) {
	backend := setupTestBackend(t)

	// Create multiple sessions
	sessions := []struct {
		id   string
		name string
	}{
		{"session-1", "First Session"},
		{"session-2", "Second Session"},
		{"session-3", "Third Session"},
	}

	for _, s := range sessions {
		session := &storage.Session{
			ID:       s.id,
			Name:     s.name,
			Messages: []storage.Message{},
			Created:  time.Now(),
			Updated:  time.Now(),
		}
		err := backend.SaveSession(session)
		require.NoError(t, err)
	}

	// List sessions
	infos, err := backend.ListSessions()
	require.NoError(t, err)
	assert.Len(t, infos, 3)

	// Verify session info
	foundSessions := make(map[string]bool)
	for _, info := range infos {
		foundSessions[info.ID] = true
		assert.NotEmpty(t, info.Name)
		assert.NotZero(t, info.Created)
		assert.NotZero(t, info.Updated)
	}

	for _, s := range sessions {
		assert.True(t, foundSessions[s.id], "Session %s not found", s.id)
	}
}

func TestBackend_DeleteSession(t *testing.T) {
	backend := setupTestBackend(t)

	// Create and save a session
	session := backend.NewSession("Delete Test")
	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Verify it exists
	sessionFile := filepath.Join(backend.baseDir, session.ID+".json")
	assert.FileExists(t, sessionFile)

	// Delete session
	err = backend.DeleteSession(session.ID)
	require.NoError(t, err)

	// Verify file is gone
	assert.NoFileExists(t, sessionFile)

	// Try to delete non-existent session
	err = backend.DeleteSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestBackend_SearchSessions(t *testing.T) {
	backend := setupTestBackend(t)

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
	assert.NotEmpty(t, results[0].Matches)

	// Check match types
	matchTypes := make(map[string]bool)
	for _, match := range results[0].Matches {
		matchTypes[match.Type] = true
	}
	assert.True(t, matchTypes["message"])
	assert.True(t, matchTypes["system_prompt"])
	assert.True(t, matchTypes["name"])
	assert.True(t, matchTypes["tag"])

	// Search for "JavaScript"
	results, err = backend.SearchSessions("JavaScript")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-2", results[0].Session.ID)

	// Search for non-existent term
	results, err = backend.SearchSessions("Rust")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestBackend_ExportSession(t *testing.T) {
	backend := setupTestBackend(t)

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

		// Verify JSON structure
		var exported storage.Session
		err = json.Unmarshal(buf.Bytes(), &exported)
		require.NoError(t, err)
		assert.Equal(t, session.ID, exported.ID)
		assert.Equal(t, session.Name, exported.Name)
		assert.Len(t, exported.Messages, 1)
	})

	// Test Markdown export
	t.Run("Markdown Export", func(t *testing.T) {
		var buf bytes.Buffer
		err := backend.ExportSession(session.ID, storage.ExportFormatMarkdown, &buf)
		require.NoError(t, err)

		markdown := buf.String()
		assert.Contains(t, markdown, "# Session: Export Test Session")
		assert.Contains(t, markdown, "## Conversation")
		assert.Contains(t, markdown, "### User")
		assert.Contains(t, markdown, "Test message")
		assert.Contains(t, markdown, "Attachments:")
		assert.Contains(t, markdown, "test.txt")
	})

	// Test unsupported format
	t.Run("Unsupported Format", func(t *testing.T) {
		var buf bytes.Buffer
		err := backend.ExportSession(session.ID, "unsupported", &buf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported export format")
	})
}

func TestBackend_Close(t *testing.T) {
	backend := setupTestBackend(t)

	// Close should succeed (no-op for filesystem)
	err := backend.Close()
	assert.NoError(t, err)
}

func TestGenerateSessionID(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := storage.GenerateSessionID()
		assert.NotEmpty(t, id)
		assert.False(t, ids[id], "Duplicate ID generated: %s", id)
		ids[id] = true

		// Verify format: YYYYMMDD-HHMMSS-NNNNNNNNN-RRRRRRRR (8 hex chars)
		assert.True(t, strings.Contains(id, "-"))
		assert.Regexp(t, `^\d{8}-\d{6}-\d{9}-[0-9a-f]{8}$`, id)
	}
}

func TestExtractSnippet(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		query    string
		radius   int
		expected string
	}{
		{
			name:     "simple match",
			content:  "This is a test content with matching word here",
			query:    "matching",
			radius:   10,
			expected: "...content with matching word here...",
		},
		{
			name:     "beginning of content",
			content:  "Query at the start of content",
			query:    "Query",
			radius:   10,
			expected: "Query at the start...",
		},
		{
			name:     "end of content",
			content:  "Content ends with query",
			query:    "query",
			radius:   10,
			expected: "...ends with query",
		},
		{
			name:     "no match",
			content:  "No match here",
			query:    "missing",
			radius:   5,
			expected: "No match h...",
		},
		{
			name:     "short content",
			content:  "Short",
			query:    "test",
			radius:   10,
			expected: "Short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSnippet(tt.content, tt.query, tt.radius)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBackend_CorruptedFile(t *testing.T) {
	backend := setupTestBackend(t)

	// Create a corrupted JSON file
	corruptedFile := filepath.Join(backend.baseDir, "corrupted.json")
	err := os.WriteFile(corruptedFile, []byte("{ invalid json"), 0644)
	require.NoError(t, err)

	// Should be skipped in ListSessions
	sessions, err := backend.ListSessions()
	require.NoError(t, err)
	assert.Empty(t, sessions)

	// Should error on LoadSession
	_, err = backend.LoadSession("corrupted")
	assert.Error(t, err)
}

// Helper function to set up a test backend
func setupTestBackend(t *testing.T) *Backend {
	tempDir := t.TempDir()
	config := storage.Config{
		"base_dir": tempDir,
	}

	backend, err := New(config)
	require.NoError(t, err)

	// Type assert to get the concrete type
	fsBackend, ok := backend.(*Backend)
	require.True(t, ok, "Backend should be of type *filesystem.Backend")

	return fsBackend
}
