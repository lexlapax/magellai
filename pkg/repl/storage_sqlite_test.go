//go:build sqlite || db
// +build sqlite db

// ABOUTME: Tests for SQLite storage backend implementation
// ABOUTME: Validates database operations and multi-tenant support

package repl

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteStorage(t *testing.T) {
	// Create temporary database file
	tmpDir, err := os.MkdirTemp("", "magellai-sqlite-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Test configuration with custom user
	config := map[string]interface{}{
		"path":    dbPath,
		"user_id": "test_user",
	}

	t.Run("CreateStorage", func(t *testing.T) {
		storage, err := NewSQLiteStorage(config)
		require.NoError(t, err)
		require.NotNil(t, storage)
		defer storage.Close()

		assert.Equal(t, "test_user", storage.userID)
	})

	t.Run("SessionOperations", func(t *testing.T) {
		storage, err := NewSQLiteStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		// Create a new session
		session := storage.NewSession("test-session")
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, "test-session", session.Name)

		// Add messages
		session.Conversation.AddMessage("user", "Hello", nil)
		session.Conversation.AddMessage("assistant", "Hi there!", nil)

		// Add attachments to test JSON serialization
		attachments := []llm.Attachment{
			{Type: llm.AttachmentTypeFile, FilePath: "test.txt", Content: "test content"},
		}
		session.Conversation.AddMessage("user", "Here's a file", attachments)

		// Save session
		err = storage.SaveSession(session)
		require.NoError(t, err)

		// Load session
		loaded, err := storage.LoadSession(session.ID)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		assert.Equal(t, session.ID, loaded.ID)
		assert.Equal(t, session.Name, loaded.Name)
		assert.Len(t, loaded.Conversation.Messages, 3)

		// Check messages
		assert.Equal(t, "user", loaded.Conversation.Messages[0].Role)
		assert.Equal(t, "Hello", loaded.Conversation.Messages[0].Content)
		assert.Nil(t, loaded.Conversation.Messages[0].Attachments)

		assert.Equal(t, "assistant", loaded.Conversation.Messages[1].Role)
		assert.Equal(t, "Hi there!", loaded.Conversation.Messages[1].Content)

		assert.Equal(t, "user", loaded.Conversation.Messages[2].Role)
		assert.Equal(t, "Here's a file", loaded.Conversation.Messages[2].Content)
		assert.Len(t, loaded.Conversation.Messages[2].Attachments, 1)
		assert.Equal(t, "test.txt", loaded.Conversation.Messages[2].Attachments[0].FilePath)
		assert.Equal(t, "test content", loaded.Conversation.Messages[2].Attachments[0].Content)
	})

	t.Run("ListSessions", func(t *testing.T) {
		// Use a unique database for this test
		tmpDir2, err := os.MkdirTemp("", "magellai-sqlite-list-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir2)

		config2 := map[string]interface{}{
			"path":    filepath.Join(tmpDir2, "list-test.db"),
			"user_id": "test_user",
		}

		storage, err := NewSQLiteStorage(config2)
		require.NoError(t, err)
		defer storage.Close()

		// Create multiple sessions
		session1 := storage.NewSession("session-1")
		session1.Conversation.AddMessage("user", "Message 1", nil)
		err = storage.SaveSession(session1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // Ensure different timestamps

		session2 := storage.NewSession("session-2")
		session2.Conversation.AddMessage("user", "Message 2", nil)
		session2.Conversation.AddMessage("assistant", "Reply 2", nil)
		err = storage.SaveSession(session2)
		require.NoError(t, err)

		// List sessions
		sessions, err := storage.ListSessions()
		require.NoError(t, err)
		assert.Len(t, sessions, 2)

		// Should be ordered by updated time (newest first)
		assert.Equal(t, session2.ID, sessions[0].ID)
		assert.Equal(t, "session-2", sessions[0].Name)
		assert.Equal(t, 2, sessions[0].MessageCount)

		assert.Equal(t, session1.ID, sessions[1].ID)
		assert.Equal(t, "session-1", sessions[1].Name)
		assert.Equal(t, 1, sessions[1].MessageCount)
	})

	t.Run("DeleteSession", func(t *testing.T) {
		storage, err := NewSQLiteStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		// Create and save a session
		session := storage.NewSession("to-delete")
		err = storage.SaveSession(session)
		require.NoError(t, err)

		// Verify it exists
		loaded, err := storage.LoadSession(session.ID)
		require.NoError(t, err)
		assert.NotNil(t, loaded)

		// Delete it
		err = storage.DeleteSession(session.ID)
		require.NoError(t, err)

		// Verify it's gone
		_, err = storage.LoadSession(session.ID)
		assert.Error(t, err)
	})

	t.Run("SearchSessions", func(t *testing.T) {
		storage, err := NewSQLiteStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		// Create sessions with specific content
		session1 := storage.NewSession("search-1")
		session1.Conversation.AddMessage("user", "Tell me about golang", nil)
		session1.Conversation.AddMessage("assistant", "Golang is a statically typed programming language", nil)
		err = storage.SaveSession(session1)
		require.NoError(t, err)

		session2 := storage.NewSession("search-2")
		session2.Conversation.AddMessage("user", "What is Python?", nil)
		session2.Conversation.AddMessage("assistant", "Python is a dynamically typed programming language", nil)
		err = storage.SaveSession(session2)
		require.NoError(t, err)

		// Search for "golang"
		results, err := storage.SearchSessions("golang")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, session1.ID, results[0].Session.ID)
		assert.Equal(t, "search-1", results[0].Session.Name)
		// With LIKE search, we get multiple matches for the same session
		assert.GreaterOrEqual(t, len(results[0].Matches), 1)

		// Search for "programming language"
		results, err = storage.SearchSessions("programming language")
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("MultiTenantSupport", func(t *testing.T) {
		// Create storage for user1
		config1 := map[string]interface{}{
			"path":    dbPath,
			"user_id": "user1",
		}
		storage1, err := NewSQLiteStorage(config1)
		require.NoError(t, err)
		defer storage1.Close()

		// Create storage for user2
		config2 := map[string]interface{}{
			"path":    dbPath,
			"user_id": "user2",
		}
		storage2, err := NewSQLiteStorage(config2)
		require.NoError(t, err)
		defer storage2.Close()

		// Create sessions for each user
		session1 := storage1.NewSession("user1-session")
		err = storage1.SaveSession(session1)
		require.NoError(t, err)

		session2 := storage2.NewSession("user2-session")
		err = storage2.SaveSession(session2)
		require.NoError(t, err)

		// Each user should only see their own sessions
		sessions1, err := storage1.ListSessions()
		require.NoError(t, err)
		assert.Len(t, sessions1, 1)
		assert.Equal(t, "user1-session", sessions1[0].Name)

		sessions2, err := storage2.ListSessions()
		require.NoError(t, err)
		assert.Len(t, sessions2, 1)
		assert.Equal(t, "user2-session", sessions2[0].Name)

		// User1 should not be able to load user2's session
		_, err = storage1.LoadSession(session2.ID)
		assert.Error(t, err)

		// User2 should not be able to load user1's session
		_, err = storage2.LoadSession(session1.ID)
		assert.Error(t, err)
	})
}

func TestSQLiteStorage_ErrorCases(t *testing.T) {
	t.Run("MissingPath", func(t *testing.T) {
		config := map[string]interface{}{}
		_, err := NewSQLiteStorage(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database path not specified")
	})

	t.Run("LoadNonExistentSession", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "magellai-sqlite-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		config := map[string]interface{}{
			"path": filepath.Join(tmpDir, "test.db"),
		}
		storage, err := NewSQLiteStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		_, err = storage.LoadSession("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("DeleteNonExistentSession", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "magellai-sqlite-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		config := map[string]interface{}{
			"path": filepath.Join(tmpDir, "test.db"),
		}
		storage, err := NewSQLiteStorage(config)
		require.NoError(t, err)
		defer storage.Close()

		err = storage.DeleteSession("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}
