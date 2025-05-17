package repl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	repl.session.Conversation.AddMessage("user", "Hello", nil)
	err := repl.manager.SaveSession(repl.session)
	require.NoError(t, err)
	sessionID := repl.session.ID

	// Create a new session
	newSession := repl.manager.NewSession("New Session")
	repl.session = newSession

	// Load the original session
	err = repl.loadSession([]string{sessionID})
	require.NoError(t, err)

	assert.Equal(t, sessionID, repl.session.ID)
	assert.Equal(t, "Original Session", repl.session.Name)
	assert.Len(t, repl.session.Conversation.Messages, 1)
	assert.Contains(t, output.String(), "Loaded session: Original Session")
}

func TestREPL_exportSession(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	var err error

	// Add some conversation to export
	repl.session.Name = "Export Test Session"
	repl.session.Conversation.SetSystemPrompt("You are a helpful assistant.")
	repl.session.Conversation.AddMessage("user", "Hello!", nil)
	repl.session.Conversation.AddMessage("assistant", "Hi there! How can I help you?", nil)

	// Add a message with attachment
	attachment := llm.Attachment{
		Type:     llm.AttachmentTypeText,
		FilePath: "test.txt",
		MimeType: "text/plain",
		Content:  "Test content",
	}
	repl.session.Conversation.AddMessage("user", "Check this file", []llm.Attachment{attachment})
	repl.session.Conversation.AddMessage("assistant", "I've reviewed the file.", nil)

	// Save the session so it can be exported
	err = repl.manager.SaveSession(repl.session)
	require.NoError(t, err)

	// Test export to stdout (JSON)
	err = repl.exportSession([]string{"json"})
	require.NoError(t, err)
	exported := output.String()
	assert.Contains(t, exported, "Export Test Session")
	assert.Contains(t, exported, "Hello!")
	assert.Contains(t, exported, "Hi there!")

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(exported), &jsonData)
	if err != nil {
		t.Logf("Invalid JSON output: %s", exported)
	}
	require.NoError(t, err)

	output.Reset()

	// Test export to stdout (Markdown)
	err = repl.exportSession([]string{"markdown"})
	require.NoError(t, err)
	exported = output.String()
	assert.Contains(t, exported, "# Session: Export Test Session")
	assert.Contains(t, exported, "### User")
	assert.Contains(t, exported, "### Assistant")
	assert.Contains(t, exported, "Attachments:")

	output.Reset()

	// Test export to file
	tempFile := filepath.Join(repl.manager.StorageDir, "export_test.json")
	err = repl.exportSession([]string{"json", tempFile})
	require.NoError(t, err)
	assert.Contains(t, output.String(), "Session exported to:")

	// Verify file was created and contains valid JSON
	data, err := os.ReadFile(tempFile)
	require.NoError(t, err)
	err = json.Unmarshal(data, &jsonData)
	require.NoError(t, err)

	// Test invalid format
	err = repl.exportSession([]string{"invalid"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")

	// Test no arguments
	err = repl.exportSession([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usage:")
}

func TestREPL_loadSession_NoID(t *testing.T) {
	repl, _, cleanup := setupTestREPL(t)
	defer cleanup()

	err := repl.loadSession([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session ID required")
}

func TestREPL_resetConversation(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Add messages
	repl.session.Conversation.AddMessage("user", "Hello", nil)
	repl.session.Conversation.AddMessage("assistant", "Hi!", nil)
	assert.Len(t, repl.session.Conversation.Messages, 2)

	// Reset
	err := repl.resetConversation()
	require.NoError(t, err)

	assert.Len(t, repl.session.Conversation.Messages, 0)
	assert.Contains(t, output.String(), "Conversation history cleared")
}

func TestREPL_setSystemPrompt(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Show prompt when none set
	err := repl.setSystemPrompt([]string{})
	require.NoError(t, err)
	assert.Contains(t, output.String(), "No system prompt set")

	output.Reset()

	// Set prompt
	err = repl.setSystemPrompt([]string{"You", "are", "helpful"})
	require.NoError(t, err)
	assert.Equal(t, "You are helpful", repl.session.Conversation.SystemPrompt)
	assert.Contains(t, output.String(), "System prompt updated")

	output.Reset()

	// Show current prompt
	err = repl.setSystemPrompt([]string{})
	require.NoError(t, err)
	assert.Contains(t, output.String(), "System prompt: You are helpful")
}

func TestREPL_showHistory(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// No history
	err := repl.showHistory()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "No conversation history")

	output.Reset()

	// Add messages
	repl.session.Conversation.AddMessage("user", "Hello", nil)
	repl.session.Conversation.AddMessage("assistant", "Hi there!", nil)

	// With attachments
	attachments := []llm.Attachment{
		{Type: llm.AttachmentTypeImage, FilePath: "test.jpg"},
	}
	repl.session.Conversation.AddMessage("user", "Check this", attachments)

	// Show history
	err = repl.showHistory()
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Conversation history (3 messages)")
	assert.Contains(t, outputStr, "[1] User:")
	assert.Contains(t, outputStr, "Hello")
	assert.Contains(t, outputStr, "[2] Assistant:")
	assert.Contains(t, outputStr, "Hi there!")
	assert.Contains(t, outputStr, "[3] User:")
	assert.Contains(t, outputStr, "Check this")
	assert.Contains(t, outputStr, "Attachments: 1")
}

func TestREPL_listSessions(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Save current session
	repl.session.Name = "Current Session"
	err := repl.manager.SaveSession(repl.session)
	require.NoError(t, err)
	currentID := repl.session.ID

	// Create and save another session
	session2 := repl.manager.NewSession("Another Session")
	session2.Tags = []string{"test", "demo"}
	err = repl.manager.SaveSession(session2)
	require.NoError(t, err)

	// List sessions
	err = repl.listSessions()
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Available sessions (2)")
	assert.Contains(t, outputStr, currentID)
	assert.Contains(t, outputStr, "Current Session (current)")
	assert.Contains(t, outputStr, "Another Session")
	assert.Contains(t, outputStr, "Tags: test, demo")
}

func TestREPL_attachFile(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Create test file
	testFile := filepath.Join(t.TempDir(), "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Attach file
	err = repl.attachFile([]string{testFile})
	require.NoError(t, err)

	// Check pending attachments
	attachments, ok := repl.session.Metadata["pending_attachments"].([]llm.Attachment)
	require.True(t, ok)
	require.Len(t, attachments, 1)

	att := attachments[0]
	assert.Equal(t, llm.AttachmentTypeText, att.Type)
	assert.Equal(t, testFile, att.FilePath)
	assert.Contains(t, att.MimeType, "text")
	assert.Equal(t, "test content", att.Content)

	assert.Contains(t, output.String(), "Attached: test.txt")
}

func TestREPL_attachFile_NoPath(t *testing.T) {
	repl, _, cleanup := setupTestREPL(t)
	defer cleanup()

	err := repl.attachFile([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file path required")
}

func TestREPL_listAttachments(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// No attachments
	err := repl.listAttachments()
	require.NoError(t, err)
	assert.Contains(t, output.String(), "No pending attachments")

	output.Reset()

	// Add attachments
	attachments := []llm.Attachment{
		{Type: llm.AttachmentTypeImage, FilePath: "/path/to/image.jpg", MimeType: "image/jpeg"},
		{Type: llm.AttachmentTypeText, FilePath: "/path/to/doc.txt", MimeType: "text/plain"},
	}
	repl.session.Metadata = map[string]interface{}{
		"pending_attachments": attachments,
	}

	err = repl.listAttachments()
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Pending attachments (2)")
	assert.Contains(t, outputStr, "1. image.jpg (image/jpeg)")
	assert.Contains(t, outputStr, "2. doc.txt (text/plain)")
}

func TestREPL_switchModel(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Missing model name
	err := repl.switchModel([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model name required")

	// Valid model switch
	err = repl.switchModel([]string{"mock/new-model"})
	require.NoError(t, err)

	assert.Equal(t, "mock/new-model", repl.session.Conversation.Model)
	assert.Equal(t, "mock", repl.session.Conversation.Provider)
	assert.Contains(t, output.String(), "Switched to model: mock/new-model")
}

func TestREPL_toggleStreaming(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Turn on
	err := repl.toggleStreaming([]string{"on"})
	require.NoError(t, err)
	assert.True(t, repl.config.GetBool("stream"))
	assert.Contains(t, output.String(), "Streaming enabled")

	output.Reset()

	// Turn off
	err = repl.toggleStreaming([]string{"off"})
	require.NoError(t, err)
	assert.False(t, repl.config.GetBool("stream"))
	assert.Contains(t, output.String(), "Streaming disabled")

	output.Reset()

	// Toggle (should turn on)
	err = repl.toggleStreaming([]string{})
	require.NoError(t, err)
	assert.True(t, repl.config.GetBool("stream"))
	assert.Contains(t, output.String(), "Streaming enabled")
}

func TestREPL_setTemperature(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Missing value
	err := repl.setTemperature([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "temperature value required")

	// Invalid value
	err = repl.setTemperature([]string{"invalid"})
	assert.Error(t, err)

	// Out of range
	err = repl.setTemperature([]string{"2.5"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "temperature must be between")

	// Valid value
	err = repl.setTemperature([]string{"0.8"})
	require.NoError(t, err)
	assert.Equal(t, 0.8, repl.session.Conversation.Temperature)
	assert.Contains(t, output.String(), "Temperature set to: 0.80")
}

func TestREPL_setMaxTokens(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Missing value
	err := repl.setMaxTokens([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max tokens value required")

	// Invalid value
	err = repl.setMaxTokens([]string{"invalid"})
	assert.Error(t, err)

	// Non-positive value
	err = repl.setMaxTokens([]string{"0"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max tokens must be positive")

	// Valid value
	err = repl.setMaxTokens([]string{"500"})
	require.NoError(t, err)
	assert.Equal(t, 500, repl.session.Conversation.MaxTokens)
	assert.Contains(t, output.String(), "Max tokens set to: 500")
}

func TestREPL_toggleMultiline(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Initially off
	assert.False(t, repl.multiline)

	// Turn on
	err := repl.toggleMultiline()
	require.NoError(t, err)
	assert.True(t, repl.multiline)
	assert.Contains(t, output.String(), "Multi-line mode enabled")

	output.Reset()

	// Turn off
	err = repl.toggleMultiline()
	require.NoError(t, err)
	assert.False(t, repl.multiline)
	assert.Contains(t, output.String(), "Multi-line mode disabled")
}
