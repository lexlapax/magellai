package repl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to add messages to conversations in tests
func addTestMessage(conv *domain.Conversation, role, content string, attachments []llm.Attachment) {
	msg := NewMessage(role, content, attachments)
	conv.AddMessage(msg)
}

func TestREPL_saveSession(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Save with no name
	err := repl.saveSession([]string{})
	require.NoError(t, err)
	assert.Contains(t, output.String(), "Session saved:")

	output.Reset()

	// Save with custom name
	err = repl.saveSession([]string{"My", "Test", "Session"})
	require.NoError(t, err)
	assert.Equal(t, "My Test Session", repl.session.Name)
	assert.Contains(t, output.String(), "Session saved:")
}

func TestREPL_loadSession(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Create and save a session
	repl.session.Name = "Original Session"
	addTestMessage(repl.session.Conversation, "user", "Hello", nil)
	err := repl.manager.SaveSession(repl.session)
	require.NoError(t, err)
	sessionID := repl.session.ID

	// Create a new session
	newSession, err := repl.manager.NewSession("New Session")
	require.NoError(t, err)
	repl.session = newSession

	// Load the original session
	err = repl.loadSession([]string{sessionID})
	require.NoError(t, err)
	assert.Equal(t, sessionID, repl.session.ID)
	assert.Equal(t, "Original Session", repl.session.Name)
	assert.Len(t, repl.session.Conversation.Messages, 1)
	assert.Contains(t, output.String(), "Loaded session: Original Session")
}

func TestREPL_listSessions(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Create multiple sessions
	session1 := repl.session
	session1.Name = "Session 1"
	addTestMessage(session1.Conversation, "user", "Hello", nil)
	addTestMessage(session1.Conversation, "assistant", "Hi there", nil)

	// Create attachment for session 2
	attachment := llm.Attachment{Type: "file", FilePath: "test.txt"}
	addTestMessage(session1.Conversation, "user", "With attachment", []llm.Attachment{attachment})
	addTestMessage(session1.Conversation, "assistant", "Got it", nil)

	err := repl.manager.SaveSession(session1)
	require.NoError(t, err)

	session2, err := repl.manager.NewSession("Session 2")
	require.NoError(t, err)
	err = repl.manager.SaveSession(session2)
	require.NoError(t, err)

	// List sessions
	err = repl.listSessions()
	require.NoError(t, err)

	output_str := output.String()
	assert.Contains(t, output_str, "Session 1")
	assert.Contains(t, output_str, "Session 2")
	assert.Contains(t, output_str, "Messages: 4")
	assert.Contains(t, output_str, "Messages: 0")
}

func TestREPL_showHistory(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Add messages with different types of content
	addTestMessage(repl.session.Conversation, "user", "Hello", nil)
	addTestMessage(repl.session.Conversation, "assistant", "Hi there! How can I help?", nil)

	// Test showing history
	err := repl.showHistory()
	require.NoError(t, err)

	output_str := output.String()
	assert.Contains(t, output_str, "Hello")
	assert.Contains(t, output_str, "Hi there! How can I help?")
	assert.Contains(t, output_str, "[1]")
	assert.Contains(t, output_str, "[2]")
}

func TestREPL_showHistory_EmptyConversation(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Test showing empty history
	err := repl.showHistory()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "No conversation history.")
}

func TestREPL_showHistory_WithSystemPrompt(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Set system prompt
	repl.session.Conversation.SystemPrompt = "You are a helpful assistant."

	// Add messages
	addTestMessage(repl.session.Conversation, "user", "Hello", nil)
	addTestMessage(repl.session.Conversation, "assistant", "Hi there!", nil)

	// Test showing history
	err := repl.showHistory()
	require.NoError(t, err)

	output_str := output.String()
	assert.Contains(t, output_str, "You are a helpful assistant.")
	assert.Contains(t, output_str, "Hello")
	assert.Contains(t, output_str, "Hi there!")
}

func TestREPL_resetConversation(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Add messages
	addTestMessage(repl.session.Conversation, "user", "Message 1", nil)
	addTestMessage(repl.session.Conversation, "assistant", "Response 1", nil)
	assert.Len(t, repl.session.Conversation.Messages, 2)

	// Reset conversation
	err := repl.resetConversation()
	require.NoError(t, err)
	assert.Len(t, repl.session.Conversation.Messages, 0)
	assert.Contains(t, output.String(), "Conversation history cleared.")
}

func TestREPL_exportSession(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Add data to session
	repl.session.Name = "Export Test"
	addTestMessage(repl.session.Conversation, "user", "Hello world", nil)
	addTestMessage(repl.session.Conversation, "assistant", "Hi there!", nil)

	// Create attachment with multimodal content
	attachment := llm.Attachment{
		Type:     "image",
		FilePath: "test.png",
		MimeType: "image/png",
	}
	addTestMessage(repl.session.Conversation, "user", "Here's an image", []llm.Attachment{attachment})

	err := repl.manager.SaveSession(repl.session)
	require.NoError(t, err)

	// Test JSON export
	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "export.json")
	err = repl.exportSession([]string{repl.session.ID, exportPath})
	require.NoError(t, err)
	assert.Contains(t, output.String(), "Session exported to:")

	// Verify JSON file was created
	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)

	var exported map[string]interface{}
	err = json.Unmarshal(data, &exported)
	require.NoError(t, err)
	assert.Equal(t, "Export Test", exported["name"])

	// Test Markdown export
	output.Reset()
	mdPath := filepath.Join(tempDir, "export.md")
	err = repl.exportSession([]string{repl.session.ID, mdPath, "markdown"})
	require.NoError(t, err)
	assert.Contains(t, output.String(), "Session exported to:")

	// Verify Markdown file was created
	mdData, err := os.ReadFile(mdPath)
	require.NoError(t, err)
	mdContent := string(mdData)
	assert.Contains(t, mdContent, "# Session: Export Test")
	assert.Contains(t, mdContent, "Hello world")
	assert.Contains(t, mdContent, "test.png")
}

func TestREPL_searchSessions(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Create sessions with different content
	session1 := repl.session
	session1.Name = "Python Tutorial"
	addTestMessage(session1.Conversation, "user", "How to use Python decorators?", nil)
	addTestMessage(session1.Conversation, "assistant", "Decorators in Python are functions that modify other functions.", nil)
	err := repl.manager.SaveSession(session1)
	require.NoError(t, err)

	session2, err := repl.manager.NewSession("JavaScript Guide")
	require.NoError(t, err)
	addTestMessage(session2.Conversation, "user", "What are JavaScript promises?", nil)
	addTestMessage(session2.Conversation, "assistant", "Promises are objects representing async operations.", nil)
	err = repl.manager.SaveSession(session2)
	require.NoError(t, err)

	// Search for Python
	err = repl.searchSessions([]string{"Python"})
	require.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Python Tutorial")
	assert.Contains(t, outputStr, "decorators")
	assert.NotContains(t, outputStr, "JavaScript")

	output.Reset()

	// Search for promises
	err = repl.searchSessions([]string{"promises"})
	require.NoError(t, err)
	outputStr = output.String()
	assert.Contains(t, outputStr, "JavaScript Guide")
	assert.Contains(t, outputStr, "promises")
	assert.NotContains(t, outputStr, "Python")
}

func TestREPL_attachFile(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Create a test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Attach file
	err = repl.attachFile([]string{testFile})
	require.NoError(t, err)
	assert.Contains(t, output.String(), "File attached: test.txt")

	// Check pending attachments
	pendingAttachments, ok := repl.session.Metadata["pending_attachments"].([]llm.Attachment)
	require.True(t, ok)
	require.Len(t, pendingAttachments, 1)
	assert.Equal(t, testFile, pendingAttachments[0].FilePath)
}

func TestREPL_showModelInfo(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Set model info
	repl.session.Conversation.Model = "test-model"
	repl.session.Conversation.Provider = "test-provider"
	repl.session.Conversation.Temperature = 0.7
	repl.session.Conversation.MaxTokens = 1000
	repl.session.Conversation.SystemPrompt = "You are helpful."

	// Show model info
	err := repl.showModel()
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "test-provider")
	assert.Contains(t, outputStr, "test-model")
	assert.Contains(t, outputStr, "0.7")
	assert.Contains(t, outputStr, "1000")
	assert.Contains(t, outputStr, "You are helpful.")
}

func TestREPL_listAttachments(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Add attachments
	attachments := []llm.Attachment{
		{FilePath: "file1.txt", Type: "file", MimeType: "text/plain"},
		{FilePath: "image.png", Type: "image", MimeType: "image/png"},
	}
	repl.session.Metadata["pending_attachments"] = attachments

	// List attachments
	err := repl.listAttachments()
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "2 pending attachments")
	assert.Contains(t, outputStr, "file1.txt")
	assert.Contains(t, outputStr, "image.png")
}
