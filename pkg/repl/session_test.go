package repl

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/repl/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestSessionManager(t *testing.T) (*session.SessionManager, func()) {
	tempDir, err := os.MkdirTemp("", "magellai-session-test")
	require.NoError(t, err)

	sm := createTestSessionManager(t, tempDir)

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return sm, cleanup
}

func TestNewSessionManager(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "magellai-test-"+time.Now().Format("20060102150405"))
	defer os.RemoveAll(tempDir)

	sm := createTestSessionManager(t, tempDir)
	assert.NotNil(t, sm)

	// Check directory was created
	info, err := os.Stat(tempDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestSessionManager_NewSession(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	session, err := sm.NewSession("Test Session")
	require.NoError(t, err)

	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "Test Session", session.Name)
	assert.NotNil(t, session.Conversation)
	assert.Equal(t, session.ID, session.Conversation.ID)
	assert.NotZero(t, session.Created)
	assert.NotZero(t, session.Updated)
	// Updated timestamp should be either equal to Created or slightly after
	assert.True(t, session.Updated.Equal(session.Created) || session.Updated.After(session.Created))
	assert.NotNil(t, session.Config)
	assert.NotNil(t, session.Metadata)
}

func TestSessionManager_SaveAndLoadSession(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	// Create and setup a session
	session, err := sm.NewSession("Test Session")
	require.NoError(t, err)
	session.Tags = []string{"test", "demo"}
	session.Config["model"] = "gpt-4"
	session.Conversation.AddMessage(NewMessage("user", "Hello", nil))
	session.Conversation.AddMessage(NewMessage("assistant", "Hi there!", nil))

	// Save session
	err = sm.SaveSession(session)
	require.NoError(t, err)

	// Load session
	loaded, err := sm.StorageManager.LoadSession(session.ID)
	require.NoError(t, err)

	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.Name, loaded.Name)
	assert.Equal(t, session.Tags, loaded.Tags)
	assert.Equal(t, session.Config["model"], loaded.Config["model"])
	assert.Len(t, loaded.Conversation.Messages, 2)
	assert.Equal(t, "Hello", loaded.Conversation.Messages[0].Content)
	assert.Equal(t, "Hi there!", loaded.Conversation.Messages[1].Content)
}

func TestSessionManager_LoadSessionNotFound(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	_, err := sm.StorageManager.LoadSession("nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSessionManager_ListSessions(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	// Create multiple sessions
	session1, err := sm.NewSession("Session 1")
	require.NoError(t, err)
	session1.Tags = []string{"work"}
	session1.Conversation.AddMessage(NewMessage("user", "First message", nil))
	err = sm.SaveSession(session1)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	session2, err := sm.NewSession("Session 2")
	require.NoError(t, err)
	session2.Tags = []string{"personal"}
	session2.Conversation.AddMessage(NewMessage("user", "Second message", nil))
	session2.Conversation.AddMessage(NewMessage("assistant", "Response", nil))
	err = sm.SaveSession(session2)
	require.NoError(t, err)

	// List sessions
	sessions, err := sm.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	// Find sessions by ID
	var info1, info2 *domain.SessionInfo
	for _, info := range sessions {
		if info.ID == session1.ID {
			info1 = info
		} else if info.ID == session2.ID {
			info2 = info
		}
	}

	require.NotNil(t, info1)
	require.NotNil(t, info2)

	assert.Equal(t, "Session 1", info1.Name)
	assert.Equal(t, []string{"work"}, info1.Tags)
	assert.Equal(t, 1, info1.MessageCount)

	assert.Equal(t, "Session 2", info2.Name)
	assert.Equal(t, []string{"personal"}, info2.Tags)
	assert.Equal(t, 2, info2.MessageCount)
}

func TestSessionManager_DeleteSession(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	// Create and save a session
	session, err := sm.NewSession("To Delete")
	require.NoError(t, err)
	err = sm.SaveSession(session)
	require.NoError(t, err)

	// Verify it exists
	loaded, err := sm.StorageManager.LoadSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, loaded.ID)

	// Delete it
	err = sm.DeleteSession(session.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = sm.StorageManager.LoadSession(session.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSessionManager_DeleteSessionNotFound(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	err := sm.DeleteSession("nonexistent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestSessionManager_SearchSessions(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	// Create sessions with different content
	session1, err := sm.NewSession("Project Discussion")
	require.NoError(t, err)
	session1.Tags = []string{"work", "project"}
	session1.Conversation.AddMessage(NewMessage("user", "Let's discuss the new project", nil))
	session1.Conversation.AddMessage(NewMessage("assistant", "Sure, what aspects would you like to cover?", nil))
	err = sm.SaveSession(session1)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	session2, err := sm.NewSession("Personal Chat")
	require.NoError(t, err)
	session2.Tags = []string{"personal"}
	session2.Conversation.AddMessage(NewMessage("user", "Tell me a joke", nil))
	session2.Conversation.AddMessage(NewMessage("assistant", "Why did the chicken cross the road?", nil))
	err = sm.SaveSession(session2)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	session3, err := sm.NewSession("Code Review")
	require.NoError(t, err)
	session3.Tags = []string{"work", "code"}
	session3.Conversation.AddMessage(NewMessage("user", "Review this code for me", nil))
	err = sm.SaveSession(session3)
	require.NoError(t, err)

	// First let's list all sessions to see if they were saved properly
	allSessions, err := sm.ListSessions()
	require.NoError(t, err)
	t.Logf("Found %d sessions:", len(allSessions))
	for _, s := range allSessions {
		t.Logf("  - ID: %s, Name: %s", s.ID, s.Name)
	}

	// Search for "project"
	results, err := sm.SearchSessions("project")
	require.NoError(t, err)
	t.Logf("Search for 'project' returned %d results", len(results))
	require.Len(t, results, 1)
	assert.Equal(t, session1.ID, results[0].Session.ID)

	// Search for "code"
	results, err = sm.SearchSessions("code")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, session3.ID, results[0].Session.ID)

	// Search for "chicken"
	results, err = sm.SearchSessions("chicken")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, session2.ID, results[0].Session.ID)

	// Search in session names
	results, err = sm.SearchSessions("personal")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, session2.ID, results[0].Session.ID)
}

func TestSessionManager_ExportSessionJSON(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	// Create and save a session
	session, err := sm.NewSession("Export Test")
	require.NoError(t, err)
	session.Tags = []string{"test"}
	session.Conversation.AddMessage(NewMessage("user", "Hello", nil))
	session.Conversation.AddMessage(NewMessage("assistant", "Hi!", nil))
	err = sm.SaveSession(session)
	require.NoError(t, err)

	// Export as JSON
	var buf bytes.Buffer
	err = sm.ExportSession(session.ID, "json", &buf)
	require.NoError(t, err)

	// Verify JSON content
	output := buf.String()
	assert.Contains(t, output, `"name": "Export Test"`)
	assert.Contains(t, output, `"role": "user"`)
	assert.Contains(t, output, `"content": "Hello"`)
	assert.Contains(t, output, `"role": "assistant"`)
	assert.Contains(t, output, `"content": "Hi!"`)
}

func TestSessionManager_ExportSessionMarkdown(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	// Create and save a session
	session, err := sm.NewSession("Markdown Export")
	require.NoError(t, err)
	session.Tags = []string{"test", "export"}
	session.Conversation.AddMessage(NewMessage("user", "What is Go?", nil))
	session.Conversation.AddMessage(NewMessage("assistant", "Go is a programming language.", nil))
	err = sm.SaveSession(session)
	require.NoError(t, err)

	// Export as Markdown
	var buf bytes.Buffer
	err = sm.ExportSession(session.ID, "markdown", &buf)
	require.NoError(t, err)

	// Verify Markdown content
	output := buf.String()
	assert.Contains(t, output, "# Session: Markdown Export")
	assert.Contains(t, output, "Tags: test, export")
	assert.Contains(t, output, "### User")
	assert.Contains(t, output, "What is Go?")
	assert.Contains(t, output, "### Assistant")
	assert.Contains(t, output, "Go is a programming language.")
}

func TestSessionManager_ExportSessionUnsupportedFormat(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	session, err := sm.NewSession("Test")
	require.NoError(t, err)
	err = sm.SaveSession(session)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = sm.ExportSession(session.ID, "xml", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported export format")
}

func TestSessionManager_ExportSessionWithAttachments(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

	// Create session with attachments
	session, err := sm.NewSession("Attachments Test")
	require.NoError(t, err)
	attachments := []domain.Attachment{
		{Type: domain.AttachmentTypeImage, Content: []byte("base64data"), MimeType: "image/jpeg"},
		{Type: domain.AttachmentTypeText, Content: []byte("base64text"), MimeType: "text/plain"},
	}
	session.Conversation.AddMessage(NewMessage("user", "Check these files", attachments))
	err = sm.SaveSession(session)
	require.NoError(t, err)

	// Export as Markdown
	var buf bytes.Buffer
	err = sm.ExportSession(session.ID, "markdown", &buf)
	require.NoError(t, err)

	// Verify attachments are included
	output := buf.String()
	assert.Contains(t, output, "Attachments:")
	assert.Contains(t, output, "- image_attachment (image/jpeg)")
	assert.Contains(t, output, "- text_attachment (text/plain)")
}
