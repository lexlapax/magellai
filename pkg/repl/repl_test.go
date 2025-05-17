package repl

import (
	"bytes"
	"context"
	"testing"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider is a test implementation of llm.Provider
type mockProvider struct {
	generateFunc func(ctx context.Context, messages []llm.Message) (*llm.Response, error)
	streamFunc   func(ctx context.Context, messages []llm.Message) (<-chan llm.StreamChunk, error)
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		generateFunc: func(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
			if len(messages) == 0 {
				return &llm.Response{Content: "Mock response"}, nil
			}
			lastMsg := messages[len(messages)-1]
			return &llm.Response{
				Content: "Mock response to: " + lastMsg.Content,
			}, nil
		},
		streamFunc: func(ctx context.Context, messages []llm.Message) (<-chan llm.StreamChunk, error) {
			ch := make(chan llm.StreamChunk, 1)
			go func() {
				defer close(ch)
				response := "Mock streaming response"
				if len(messages) > 0 {
					response = "Mock response to: " + messages[len(messages)-1].Content
				}
				for _, char := range response {
					ch <- llm.StreamChunk{Content: string(char)}
				}
			}()
			return ch, nil
		},
	}
}

func (m *mockProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	return "Mock response to: " + prompt, nil
}

func (m *mockProvider) GenerateMessage(ctx context.Context, messages []llm.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, messages)
	}
	return &llm.Response{Content: "Mock response"}, nil
}

func (m *mockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	return map[string]string{"response": "Mock structured response"}, nil
}

func (m *mockProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{Content: "Mock streaming response to: " + prompt}
	}()
	return ch, nil
}

func (m *mockProvider) StreamMessage(ctx context.Context, messages []llm.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, messages)
	}
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{Content: "Mock streaming response"}
	}()
	return ch, nil
}

func (m *mockProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{
		Provider:     "mock",
		Model:        "test-model",
		DisplayName:  "Mock Test Model",
		Capabilities: llm.ModelCapabilities{Text: true},
	}
}

// setupTestConfig creates a minimal test configuration
func setupTestConfig() *testConfig {
	return &testConfig{
		values: map[string]interface{}{
			"model":  "mock/test-model",
			"stream": false,
		},
	}
}

// testConfig is a minimal implementation of config for testing
type testConfig struct {
	values map[string]interface{}
}

func (tc *testConfig) GetString(key string) string {
	if v, ok := tc.values[key].(string); ok {
		return v
	}
	return ""
}

func (tc *testConfig) GetBool(key string) bool {
	if v, ok := tc.values[key].(bool); ok {
		return v
	}
	return false
}

func (tc *testConfig) SetValue(key string, value interface{}) error {
	tc.values[key] = value
	return nil
}

func setupTestREPL(t *testing.T) (*REPL, *bytes.Buffer, func()) {
	// Create temp directory for sessions
	tempDir := t.TempDir()

	// Create test config
	cfg := setupTestConfig()

	// Create buffers for input/output
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	// Create REPL options
	opts := &REPLOptions{
		Config:      cfg,
		StorageDir:  tempDir,
		PromptStyle: "> ",
		Reader:      input,
		Writer:      output,
	}

	// Create REPL
	repl, err := NewREPL(opts)
	require.NoError(t, err)

	// Create mock provider
	mockProvider := newMockProvider()
	repl.provider = mockProvider

	cleanup := func() {
		// Cleanup is handled by t.TempDir()
	}

	return repl, output, cleanup
}

func TestNewREPL(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	assert.NotNil(t, repl)
	assert.NotNil(t, repl.session)
	assert.NotNil(t, repl.manager)
	assert.NotNil(t, repl.provider)
	assert.NotNil(t, repl.config)
	assert.Equal(t, "> ", repl.promptStyle)
	assert.Equal(t, "mock/test-model", repl.session.Conversation.Model)
	assert.Equal(t, "mock", repl.session.Conversation.Provider)

	// Check output buffer has expected content
	assert.NotNil(t, output)
}

func TestREPL_processMessage(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Process a message
	err := repl.processMessage("Hello, world!")
	require.NoError(t, err)

	// Check conversation
	assert.Len(t, repl.session.Conversation.Messages, 2)

	// Check user message
	userMsg := repl.session.Conversation.Messages[0]
	assert.Equal(t, "user", userMsg.Role)
	assert.Equal(t, "Hello, world!", userMsg.Content)

	// Check assistant response
	assistantMsg := repl.session.Conversation.Messages[1]
	assert.Equal(t, "assistant", assistantMsg.Role)
	assert.Equal(t, "Mock response to: Hello, world!", assistantMsg.Content)

	// Check output
	assert.Contains(t, output.String(), "Mock response to: Hello, world!")
}

func TestREPL_processMessageWithAttachments(t *testing.T) {
	repl, _, cleanup := setupTestREPL(t)
	defer cleanup()

	// Add pending attachments
	attachments := []llm.Attachment{
		{Type: llm.AttachmentTypeImage, FilePath: "test.jpg", MimeType: "image/jpeg"},
	}
	repl.session.Metadata = map[string]interface{}{
		"pending_attachments": attachments,
	}

	// Process message
	err := repl.processMessage("What's in this image?")
	require.NoError(t, err)

	// Check conversation
	assert.Len(t, repl.session.Conversation.Messages, 2)

	// Check user message has attachments
	userMsg := repl.session.Conversation.Messages[0]
	assert.Equal(t, "user", userMsg.Role)
	assert.Equal(t, "What's in this image?", userMsg.Content)
	assert.Len(t, userMsg.Attachments, 1)
	assert.Equal(t, llm.AttachmentTypeImage, userMsg.Attachments[0].Type)

	// Check pending attachments cleared
	_, exists := repl.session.Metadata["pending_attachments"]
	assert.False(t, exists)
}

func TestREPL_handleCommand_Help(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	err := repl.handleCommand("/help")
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "magellai chat - Interactive LLM chat")
	assert.Contains(t, outputStr, "COMMANDS:")
	assert.Contains(t, outputStr, "/help")
	assert.Contains(t, outputStr, "/exit")
	assert.Contains(t, outputStr, "SPECIAL COMMANDS:")
}

func TestREPL_handleCommand_Reset(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Add some messages
	repl.session.Conversation.AddMessage("user", "Hello", nil)
	repl.session.Conversation.AddMessage("assistant", "Hi!", nil)
	assert.Len(t, repl.session.Conversation.Messages, 2)

	// Reset conversation
	err := repl.handleCommand("/reset")
	require.NoError(t, err)

	assert.Len(t, repl.session.Conversation.Messages, 0)
	assert.Contains(t, output.String(), "Conversation history cleared")
}

func TestREPL_handleCommand_Model(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	err := repl.handleCommand("/model")
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Current model: mock/test-model")
	assert.Contains(t, outputStr, "Provider: mock")
}

func TestREPL_handleCommand_Sessions(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Save current session first
	err := repl.manager.SaveSession(repl.session)
	require.NoError(t, err)

	err = repl.handleCommand("/sessions")
	require.NoError(t, err)

	outputStr := output.String()
	assert.Contains(t, outputStr, "Available sessions")
	assert.Contains(t, outputStr, repl.session.ID)
	assert.Contains(t, outputStr, "(current)")
}

func TestREPL_handleSpecialCommand_Stream(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Initially streaming is off
	assert.False(t, repl.config.GetBool("stream"))

	// Turn on streaming
	err := repl.handleSpecialCommand(":stream on")
	require.NoError(t, err)
	assert.True(t, repl.config.GetBool("stream"))
	assert.Contains(t, output.String(), "Streaming enabled")

	output.Reset()

	// Turn off streaming
	err = repl.handleSpecialCommand(":stream off")
	require.NoError(t, err)
	assert.False(t, repl.config.GetBool("stream"))
	assert.Contains(t, output.String(), "Streaming disabled")
}

func TestREPL_handleSpecialCommand_Temperature(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	err := repl.handleSpecialCommand(":temperature 0.8")
	require.NoError(t, err)

	assert.Equal(t, 0.8, repl.session.Conversation.Temperature)
	assert.Contains(t, output.String(), "Temperature set to: 0.80")
}

func TestREPL_handleSpecialCommand_MaxTokens(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	err := repl.handleSpecialCommand(":max_tokens 500")
	require.NoError(t, err)

	assert.Equal(t, 500, repl.session.Conversation.MaxTokens)
	assert.Contains(t, output.String(), "Max tokens set to: 500")
}

func TestREPL_handleSpecialCommand_Multiline(t *testing.T) {
	repl, output, cleanup := setupTestREPL(t)
	defer cleanup()

	// Initially multiline is off
	assert.False(t, repl.multiline)

	err := repl.handleSpecialCommand(":multiline")
	require.NoError(t, err)

	assert.True(t, repl.multiline)
	assert.Contains(t, output.String(), "Multi-line mode enabled")

	output.Reset()

	// Toggle again
	err = repl.handleSpecialCommand(":multiline")
	require.NoError(t, err)

	assert.False(t, repl.multiline)
	assert.Contains(t, output.String(), "Multi-line mode disabled")
}

func TestREPL_readInput_SingleLine(t *testing.T) {
	// Create input buffer with test data
	input := bytes.NewBufferString("Hello, world!\n")
	output := &bytes.Buffer{}

	cfg := setupTestConfig()

	opts := &REPLOptions{
		Config:     cfg,
		StorageDir: t.TempDir(),
		Reader:     input,
		Writer:     output,
	}

	repl, err := NewREPL(opts)
	require.NoError(t, err)

	// Read input
	text, err := repl.readInput()
	require.NoError(t, err)

	assert.Equal(t, "Hello, world!\n", text)
}

func TestREPL_readInput_MultiLine(t *testing.T) {
	// Create input buffer with multi-line data
	input := bytes.NewBufferString("Line 1\nLine 2\n\n") // Empty line ends input
	output := &bytes.Buffer{}

	cfg := setupTestConfig()

	opts := &REPLOptions{
		Config:     cfg,
		StorageDir: t.TempDir(),
		Reader:     input,
		Writer:     output,
	}

	repl, err := NewREPL(opts)
	require.NoError(t, err)
	repl.multiline = true

	// Read input
	text, err := repl.readInput()
	require.NoError(t, err)

	// Should combine lines
	assert.Equal(t, "Line 1\nLine 2", text)
}
