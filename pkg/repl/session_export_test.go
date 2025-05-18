// ABOUTME: Unit tests for session export functionality
// ABOUTME: Tests both JSON and Markdown export formats

package repl

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

func TestSessionExport(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "session_export_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a session manager
	manager := createTestSessionManager(t, tempDir)

	// Create a test session with some content
	session, err := manager.NewSession("Test Export Session")
	if err != nil {
		t.Fatal(err)
	}
	session.Conversation.SetSystemPrompt("You are a test assistant.")
	session.Conversation.AddMessage("user", "Test question?", nil)
	session.Conversation.AddMessage("assistant", "Test response.", nil)

	// Add a message with attachment
	attachment := llm.Attachment{
		Type:     llm.AttachmentTypeText,
		FilePath: "test.txt",
		MimeType: "text/plain",
		Content:  "Test content",
	}
	session.Conversation.AddMessage("user", "Analyze this", []llm.Attachment{attachment})
	session.Conversation.AddMessage("assistant", "Analysis complete.", nil)

	// Save the session
	if err := manager.SaveSession(session); err != nil {
		t.Fatal(err)
	}

	// Test JSON export
	t.Run("JSON Export", func(t *testing.T) {
		var buf bytes.Buffer
		err := manager.ExportSession(session.ID, "json", &buf)
		if err != nil {
			t.Fatalf("Failed to export as JSON: %v", err)
		}

		// Verify JSON structure
		var exportedSession domain.Session
		if err := json.Unmarshal(buf.Bytes(), &exportedSession); err != nil {
			t.Fatalf("Failed to unmarshal exported JSON: %v", err)
		}

		// Verify content
		if exportedSession.ID != session.ID {
			t.Errorf("Expected session ID %s, got %s", session.ID, exportedSession.ID)
		}
		if exportedSession.Name != session.Name {
			t.Errorf("Expected session name %s, got %s", session.Name, exportedSession.Name)
		}
		if exportedSession.Conversation == nil || len(exportedSession.Conversation.Messages) != 4 {
			msgCount := 0
			if exportedSession.Conversation != nil {
				msgCount = len(exportedSession.Conversation.Messages)
			}
			t.Errorf("Expected 4 messages, got %d", msgCount)
		}
	})

	// Test Markdown export
	t.Run("Markdown Export", func(t *testing.T) {
		var buf bytes.Buffer
		err := manager.ExportSession(session.ID, "markdown", &buf)
		if err != nil {
			t.Fatalf("Failed to export as Markdown: %v", err)
		}

		markdown := buf.String()

		// Verify Markdown content
		if !strings.Contains(markdown, "# Session: Test Export Session") {
			t.Error("Markdown missing session title")
		}
		if !strings.Contains(markdown, "## Conversation") {
			t.Error("Markdown missing conversation section")
		}
		if !strings.Contains(markdown, "### User") {
			t.Error("Markdown missing user messages")
		}
		if !strings.Contains(markdown, "### Assistant") {
			t.Error("Markdown missing assistant messages")
		}
		if !strings.Contains(markdown, "Attachments:") {
			t.Error("Markdown missing attachment section")
		}
		if !strings.Contains(markdown, "test.txt") {
			t.Error("Markdown missing attachment filename")
		}
	})

	// Test invalid format
	t.Run("Invalid Format", func(t *testing.T) {
		var buf bytes.Buffer
		err := manager.ExportSession(session.ID, "invalid", &buf)
		if err == nil {
			t.Error("Expected error for invalid format, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported export format") {
			t.Errorf("Expected unsupported format error, got: %v", err)
		}
	})

	// Test non-existent session
	t.Run("Non-existent Session", func(t *testing.T) {
		var buf bytes.Buffer
		err := manager.ExportSession("non-existent-id", "json", &buf)
		if err == nil {
			t.Error("Expected error for non-existent session, got nil")
		}
		if !strings.Contains(err.Error(), "session not found") {
			t.Errorf("Expected session not found error, got: %v", err)
		}
	})
}

func TestREPLExportCommand(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "repl_export_cmd_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Mock config interface
	mockConfig := &mockConfig{
		values: map[string]interface{}{
			"model":                  "mock/test-model",
			"stream":                 false,
			"verbosity":              "info",
			"repl.auto_save.enabled": false,
		},
	}

	// Create mock reader and writer
	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}

	// Create REPL options
	opts := &REPLOptions{
		Config:     mockConfig,
		StorageDir: tempDir,
		Model:      "mock/test-model",
		Writer:     writer,
		Reader:     reader,
	}

	// Create mock provider
	mockProvider := &MockProvider{}

	// Create REPL instance
	repl, err := NewREPL(opts)
	if err != nil {
		t.Fatal(err)
	}

	// Replace provider with mock
	repl.provider = mockProvider

	// Add some messages to the session
	repl.session.Conversation.AddMessage("user", "Test message", nil)
	repl.session.Conversation.AddMessage("assistant", "Test response", nil)

	// Save the session so it can be exported
	if err := repl.manager.SaveSession(repl.session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Test export command with no args
	t.Run("Export without args", func(t *testing.T) {
		writer.Reset()
		err := repl.exportSession([]string{})
		if err == nil {
			t.Error("Expected error without arguments")
		}
		if !strings.Contains(err.Error(), "usage") {
			t.Errorf("Expected usage error, got: %v", err)
		}
	})

	// Test export command with JSON format
	t.Run("Export JSON to stdout", func(t *testing.T) {
		writer.Reset()
		err := repl.exportSession([]string{"json"})
		if err != nil {
			t.Fatalf("Failed to export JSON: %v", err)
		}

		output := writer.String()
		if !strings.Contains(output, `"id"`) {
			t.Error("JSON output missing session ID")
		}
		if !strings.Contains(output, `"messages"`) {
			t.Error("JSON output missing messages")
		}
	})

	// Test export command with Markdown format
	t.Run("Export Markdown to stdout", func(t *testing.T) {
		writer.Reset()
		err := repl.exportSession([]string{"markdown"})
		if err != nil {
			t.Fatalf("Failed to export Markdown: %v", err)
		}

		output := writer.String()
		if !strings.Contains(output, "# Session:") {
			t.Error("Markdown output missing session header")
		}
		if !strings.Contains(output, "## Conversation") {
			t.Error("Markdown output missing conversation section")
		}
	})

	// Test export to file
	t.Run("Export to file", func(t *testing.T) {
		writer.Reset()
		exportFile := tempDir + "/export_test.json"
		err := repl.exportSession([]string{"json", exportFile})
		if err != nil {
			t.Fatalf("Failed to export to file: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(exportFile); os.IsNotExist(err) {
			t.Error("Export file was not created")
		}

		// Verify success message
		output := writer.String()
		if !strings.Contains(output, "Session exported to:") {
			t.Error("Missing success message")
		}
		if !strings.Contains(output, exportFile) {
			t.Error("Success message missing filename")
		}
	})

	// Test invalid format
	t.Run("Invalid format", func(t *testing.T) {
		writer.Reset()
		err := repl.exportSession([]string{"invalid"})
		if err == nil {
			t.Error("Expected error for invalid format")
		}
		if !strings.Contains(err.Error(), "unsupported format") {
			t.Errorf("Expected unsupported format error, got: %v", err)
		}
	})
}

// Mock config and provider for testing
type mockConfig struct {
	values map[string]interface{}
}

func (m *mockConfig) GetString(key string) string {
	if v, ok := m.values[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (m *mockConfig) GetBool(key string) bool {
	if v, ok := m.values[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func (m *mockConfig) Get(key string) interface{} {
	return m.values[key]
}

func (m *mockConfig) Exists(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *mockConfig) SetValue(key string, value interface{}) error {
	m.values[key] = value
	return nil
}

type MockProvider struct{}

func (m *MockProvider) Generate(ctx context.Context, prompt string, opts ...llm.ProviderOption) (string, error) {
	return "Test response", nil
}

func (m *MockProvider) GenerateMessage(ctx context.Context, messages []llm.Message, opts ...llm.ProviderOption) (*llm.Response, error) {
	return &llm.Response{
		Content: "Test response",
	}, nil
}

func (m *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, opts ...llm.ProviderOption) (interface{}, error) {
	return nil, nil
}

func (m *MockProvider) Stream(ctx context.Context, prompt string, opts ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		ch <- llm.StreamChunk{Content: "Test stream"}
		close(ch)
	}()
	return ch, nil
}

func (m *MockProvider) StreamMessage(ctx context.Context, messages []llm.Message, opts ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		ch <- llm.StreamChunk{Content: "Test stream"}
		close(ch)
	}()
	return ch, nil
}

func (m *MockProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{
		Provider: "test",
		Model:    "model",
		Capabilities: llm.ModelCapabilities{
			Text: true,
		},
	}
}
