// ABOUTME: Tests for the filesystem storage backend implementation
// ABOUTME: Ensures proper file operations, JSON persistence, and search functionality

package filesystem

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
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
	assert.NotNil(t, session.Conversation)
	assert.Empty(t, session.Conversation.Messages)
	assert.NotNil(t, session.Config)
	assert.NotNil(t, session.Metadata)
}

func TestBackend_SaveAndLoadSession(t *testing.T) {
	backend := setupTestBackend(t)

	// Create a session with data
	session := domain.NewSession("test-session-001")
	session.Name = "Test Session"
	session.Tags = []string{"test", "example"}

	// Add conversation data
	session.Conversation.SetModel("openai", "gpt-4")
	session.Conversation.SetSystemPrompt("You are helpful")
	session.Conversation.AddMessage(*domain.NewMessage("msg-1", domain.MessageRoleUser, "Hello"))
	session.Conversation.AddMessage(*domain.NewMessage("msg-2", domain.MessageRoleAssistant, "Hi there!"))

	// Add config
	session.Config["temperature"] = 0.7

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
	assert.NotNil(t, loaded.Conversation)
	assert.Len(t, loaded.Conversation.Messages, 2)
	assert.Equal(t, session.Conversation.Model, loaded.Conversation.Model)
	assert.Equal(t, session.Conversation.SystemPrompt, loaded.Conversation.SystemPrompt)
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
		session := domain.NewSession(s.id)
		session.Name = s.name
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
	sessions := []*domain.Session{
		createTestSession("search-1", "Python Programming", "You are a Python expert"),
		createTestSession("search-2", "Go Language", "You are a Go expert"),
		createTestSession("search-3", "JavaScript Tutorial", "You are a JavaScript expert"),
	}

	// Add messages to sessions
	sessions[0].Conversation.AddMessage(*domain.NewMessage("msg-1", domain.MessageRoleUser, "Tell me about Python classes"))
	sessions[0].Conversation.AddMessage(*domain.NewMessage("msg-2", domain.MessageRoleAssistant, "Python classes are object blueprints"))

	sessions[1].Conversation.AddMessage(*domain.NewMessage("msg-3", domain.MessageRoleUser, "Explain Go interfaces"))
	sessions[1].Conversation.AddMessage(*domain.NewMessage("msg-4", domain.MessageRoleAssistant, "Go interfaces define method signatures"))

	sessions[2].Conversation.AddMessage(*domain.NewMessage("msg-5", domain.MessageRoleUser, "What is JavaScript closure?"))
	sessions[2].Conversation.AddMessage(*domain.NewMessage("msg-6", domain.MessageRoleAssistant, "A closure is a function that has access to outer scope"))

	// Add tags
	sessions[0].Tags = []string{"programming", "python", "tutorial"}
	sessions[1].Tags = []string{"programming", "golang", "tutorial"}
	sessions[2].Tags = []string{"programming", "javascript", "tutorial"}

	// Save all sessions
	for _, session := range sessions {
		err := backend.SaveSession(session)
		require.NoError(t, err)
	}

	// Test different search queries
	tests := []struct {
		name     string
		query    string
		expected []string // Expected session IDs
	}{
		{
			name:     "Search for Python",
			query:    "Python",
			expected: []string{"search-1"},
		},
		{
			name:     "Search for programming",
			query:    "programming",
			expected: []string{"search-1", "search-2", "search-3"},
		},
		{
			name:     "Search for interfaces",
			query:    "interfaces",
			expected: []string{"search-2"},
		},
		{
			name:     "Search for tutorial",
			query:    "tutorial",
			expected: []string{"search-1", "search-2", "search-3"},
		},
		{
			name:     "Search for expert",
			query:    "expert",
			expected: []string{"search-1", "search-2", "search-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := backend.SearchSessions(tt.query)
			require.NoError(t, err)

			foundIDs := make(map[string]bool)
			for _, result := range results {
				foundIDs[result.Session.ID] = true
				assert.True(t, result.HasMatches())
			}

			assert.Len(t, foundIDs, len(tt.expected))
			for _, expectedID := range tt.expected {
				assert.True(t, foundIDs[expectedID], "Expected to find session %s", expectedID)
			}
		})
	}
}

func TestBackend_ExportSession(t *testing.T) {
	backend := setupTestBackend(t)

	// Create a session with content
	session := createTestSession("export-test", "Export Test", "You are helpful")
	session.Conversation.AddMessage(*domain.NewMessage("msg-1", domain.MessageRoleUser, "Hello"))
	session.Conversation.AddMessage(*domain.NewMessage("msg-2", domain.MessageRoleAssistant, "Hi there!"))

	err := backend.SaveSession(session)
	require.NoError(t, err)

	// Test JSON export
	var buf bytes.Buffer
	err = backend.ExportSession(session.ID, domain.ExportFormatJSON, &buf)
	require.NoError(t, err)

	// Verify JSON structure
	var exported map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &exported)
	require.NoError(t, err)
	assert.Equal(t, session.ID, exported["id"])
	assert.Equal(t, session.Name, exported["name"])

	// Test Markdown export
	buf.Reset()
	err = backend.ExportSession(session.ID, domain.ExportFormatMarkdown, &buf)
	require.NoError(t, err)

	// Verify Markdown content
	markdown := buf.String()
	assert.Contains(t, markdown, "# Session: Export Test")
	assert.Contains(t, markdown, "Hello")
	assert.Contains(t, markdown, "Hi there!")
	assert.Contains(t, markdown, "User")
	assert.Contains(t, markdown, "Assistant")

	// Test unsupported format
	err = backend.ExportSession(session.ID, domain.ExportFormat("invalid"), &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported export format")
}

func TestBackend_Close(t *testing.T) {
	backend := setupTestBackend(t)
	err := backend.Close()
	assert.NoError(t, err)
}

// Helper functions

func setupTestBackend(t *testing.T) *Backend {
	tempDir := t.TempDir()
	config := storage.Config{
		"base_dir": tempDir,
	}
	backend, err := New(config)
	require.NoError(t, err)
	return backend.(*Backend)
}

func createTestSession(id, name, systemPrompt string) *domain.Session {
	session := domain.NewSession(id)
	session.Name = name
	session.Conversation.SetSystemPrompt(systemPrompt)
	return session
}

func TestExtractSnippet(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		query      string
		contextLen int
		expected   string
	}{
		{
			name:       "Match in middle",
			content:    "This is a test of the search functionality with some context",
			query:      "search",
			contextLen: 10,
			expected:   "...test of the search functionality with...",
		},
		{
			name:       "Match at start",
			content:    "Search at the beginning",
			query:      "Search",
			contextLen: 10,
			expected:   "Search at the beginning",
		},
		{
			name:       "Match at end",
			content:    "The query is at the end search",
			query:      "search",
			contextLen: 10,
			expected:   "...at the end search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSnippet(tt.content, tt.query, tt.contextLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}
