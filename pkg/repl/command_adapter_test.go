// ABOUTME: Tests for the command adapter that bridges REPL commands to the unified command system
// ABOUTME: Verifies proper command registration, execution, and error handling

package repl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock types for testing

// commandAdapterMockConfig implements ConfigInterface for testing
type commandAdapterMockConfig struct {
	values map[string]interface{}
}

func newCommandAdapterMockConfig() *commandAdapterMockConfig {
	return &commandAdapterMockConfig{
		values: make(map[string]interface{}),
	}
}

func (m *commandAdapterMockConfig) GetString(key string) string {
	if v, ok := m.values[key].(string); ok {
		return v
	}
	return ""
}

func (m *commandAdapterMockConfig) GetBool(key string) bool {
	if v, ok := m.values[key].(bool); ok {
		return v
	}
	return false
}

func (m *commandAdapterMockConfig) Get(key string) interface{} {
	return m.values[key]
}

func (m *commandAdapterMockConfig) Exists(key string) bool {
	_, exists := m.values[key]
	return exists
}

func (m *commandAdapterMockConfig) SetValue(key string, value interface{}) error {
	m.values[key] = value
	return nil
}

// mockProvider implements llm.Provider for testing
type mockLLMProvider struct {
	generateFunc        func(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error)
	generateMessageFunc func(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error)
	streamFunc          func(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error)
	streamMessageFunc   func(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error)
}

func (m *mockLLMProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, prompt, options...)
	}
	return "mock response", nil
}

func (m *mockLLMProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	if m.generateMessageFunc != nil {
		return m.generateMessageFunc(ctx, messages, options...)
	}
	return &llm.Response{Content: "mock message response"}, nil
}

func (m *mockLLMProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	return nil, nil
}

func (m *mockLLMProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, prompt, options...)
	}
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

func (m *mockLLMProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	if m.streamMessageFunc != nil {
		return m.streamMessageFunc(ctx, messages, options...)
	}
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

func (m *mockLLMProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{Provider: "mock", Model: "test"}
}

// Tests

func TestREPLCommandAdapter_Create(t *testing.T) {
	// Test adapter creation
	repl := createTestREPL(t)
	meta := &command.Metadata{
		Name:        "test",
		Description: "Test command",
		Category:    command.CategoryREPL,
	}
	handler := func(r *REPL, args []string) error {
		return nil
	}

	adapter := NewREPLCommandAdapter(repl, meta, handler)
	
	assert.NotNil(t, adapter)
	assert.Equal(t, repl, adapter.repl)
	assert.Equal(t, meta, adapter.metadata)
	assert.NotNil(t, adapter.handler)
}

func TestREPLCommandAdapter_Execute(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError error
		executeFunc   func(*REPL, []string) error
		verify        func(*testing.T, *REPL)
	}{
		{
			name: "successful execution",
			args: []string{"arg1", "arg2"},
			executeFunc: func(r *REPL, args []string) error {
				assert.Equal(t, []string{"arg1", "arg2"}, args)
				return nil
			},
		},
		{
			name:          "execution with error",
			args:          []string{"fail"},
			expectedError: errors.New("test error"),
			executeFunc: func(r *REPL, args []string) error {
				return errors.New("test error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repl := createTestREPL(t)
			meta := &command.Metadata{
				Name:        "test",
				Description: "Test command",
				Category:    command.CategoryREPL,
			}
			
			adapter := NewREPLCommandAdapter(repl, meta, tt.executeFunc)
			
			ctx := context.Background()
			exec := &command.ExecutionContext{
				Args:   tt.args,
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
			}
			
			err := adapter.Execute(ctx, exec)
			
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			
			if tt.verify != nil {
				tt.verify(t, repl)
			}
		})
	}
}

func TestREPLCommandAdapter_Metadata(t *testing.T) {
	repl := createTestREPL(t)
	
	tests := []struct {
		name     string
		metadata *command.Metadata
	}{
		{
			name: "full metadata",
			metadata: &command.Metadata{
				Name:        "test",
				Aliases:     []string{"t"},
				Description: "Test command",
				Category:    command.CategoryREPL,
			},
		},
		{
			name: "minimal metadata",
			metadata: &command.Metadata{
				Name:     "minimal",
				Category: command.CategoryREPL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(r *REPL, args []string) error { return nil }
			adapter := NewREPLCommandAdapter(repl, tt.metadata, handler)
			
			metadata := adapter.Metadata()
			assert.Equal(t, tt.metadata, metadata)
		})
	}
}

func TestREPLCommandAdapter_Validate(t *testing.T) {
	repl := createTestREPL(t)
	handler := func(r *REPL, args []string) error { return nil }
	
	tests := []struct {
		name      string
		metadata  *command.Metadata
		wantError bool
	}{
		{
			name: "valid metadata",
			metadata: &command.Metadata{
				Name:        "test",
				Description: "Test command",
				Category:    command.CategoryREPL,
			},
			wantError: false,
		},
		{
			name:      "nil metadata",
			metadata:  nil,
			wantError: true,
		},
		{
			name: "empty name",
			metadata: &command.Metadata{
				Name:        "",
				Description: "Test command",
				Category:    command.CategoryREPL,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewREPLCommandAdapter(repl, tt.metadata, handler)
			err := adapter.Validate()
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterREPLCommands(t *testing.T) {
	// Test registering all REPL commands
	repl := createTestREPL(t)
	registry := command.NewRegistry()
	
	err := RegisterREPLCommands(repl, registry)
	assert.NoError(t, err)
	
	// Verify key commands are registered
	testCommands := []struct {
		name    string
		aliases []string
	}{
		{"help", []string{"h", "?"}},
		{"exit", []string{"quit", "q"}},
		{"save", nil},
		{"load", nil},
		{"reset", nil},
		{"model", nil},
		{"system", nil},
		{"history", nil},
		{"sessions", nil},
		{"attach", nil},
		{"attachments", nil},
		{"config", nil},
		{"export", nil},
		{"search", nil},
		{"tags", nil},
		{"tag", nil},
		{"untag", nil},
		{"metadata", nil},
		{"meta", nil},
		{":model", nil},
		{":stream", []string{":streaming"}},
		{":temperature", []string{":temp"}},
		{":max_tokens", []string{":tokens"}},
		{":multiline", []string{":ml"}},
		{":verbosity", []string{":v"}},
		{":output", nil},
		{":profile", nil},
		{":attach", nil},
		{":attach-remove", nil},
		{":attach-list", nil},
		{":system", nil},
		{"branch", nil},
		{"branches", nil},
		{"tree", nil},
		{"switch", nil},
		{"merge", nil},
		{"recover", nil},
	}
	
	for _, tc := range testCommands {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := registry.Get(tc.name)
			assert.NoError(t, err, "command %s should be registered", tc.name)
			assert.NotNil(t, cmd)
			assert.Equal(t, tc.name, cmd.Metadata().Name)
			
			// Check aliases
			for _, alias := range tc.aliases {
				aliasCmd, err := registry.Get(alias)
				assert.NoError(t, err, "alias %s should be registered", alias)
				assert.NotNil(t, aliasCmd)
				assert.Equal(t, tc.name, aliasCmd.Metadata().Name)
			}
		})
	}
}

func TestRegisterREPLCommands_DuplicateError(t *testing.T) {
	repl := createTestREPL(t)
	registry := command.NewRegistry()
	
	// Pre-register a command that conflicts
	existingMeta := &command.Metadata{
		Name:        "help",
		Description: "Existing help command",
		Category:    command.CategoryCLI,
	}
	existingCmd := command.NewSimpleCommand(existingMeta, func(ctx context.Context, exec *command.ExecutionContext) error {
		return nil
	})
	
	err := registry.Register(existingCmd)
	require.NoError(t, err)
	
	// Try to register REPL commands - should fail on duplicate
	err = RegisterREPLCommands(repl, registry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to register command help")
}

func TestCreateCommandContext(t *testing.T) {
	args := []string{"arg1", "arg2"}
	stdin := strings.NewReader("input")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	
	ctx := CreateCommandContext(args, stdin, stdout, stderr)
	
	assert.Equal(t, args, ctx.Args)
	assert.Equal(t, stdin, ctx.Stdin)
	assert.Equal(t, stdout, ctx.Stdout)
	assert.Equal(t, stderr, ctx.Stderr)
	assert.NotNil(t, ctx.Data)
	assert.Empty(t, ctx.Data)
}

func TestCreateCommandContextWithShared(t *testing.T) {
	args := []string{"arg1", "arg2"}
	stdin := strings.NewReader("input")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	shared := command.NewSharedContext()
	
	// Set some shared data
	shared.Set("testkey", "testvalue")
	
	ctx := CreateCommandContextWithShared(args, stdin, stdout, stderr, shared)
	
	assert.Equal(t, args, ctx.Args)
	assert.Equal(t, stdin, ctx.Stdin)
	assert.Equal(t, stdout, ctx.Stdout)
	assert.Equal(t, stderr, ctx.Stderr)
	assert.NotNil(t, ctx.Data)
	assert.Empty(t, ctx.Data)
	assert.Equal(t, shared, ctx.SharedContext)
	
	// Verify shared context
	val, ok := ctx.SharedContext.Get("testkey")
	assert.True(t, ok)
	assert.Equal(t, "testvalue", val)
}

func TestSpecificCommandExecution(t *testing.T) {
	// Test specific command implementations
	tests := []struct {
		name         string
		commandName  string
		setup        func(*REPL)
		args         []string
		expectedErr  error
		verify       func(*testing.T, *REPL, io.Writer)
	}{
		{
			name:        "help command",
			commandName: "help",
			args:        []string{},
			verify: func(t *testing.T, r *REPL, w io.Writer) {
				// Verify help was displayed
				output := w.(*bytes.Buffer).String()
				assert.Contains(t, output, "COMMANDS:")
			},
		},
		{
			name:        "exit command",
			commandName: "exit",
			args:        []string{},
			expectedErr: io.EOF,
			verify: func(t *testing.T, r *REPL, w io.Writer) {
				output := w.(*bytes.Buffer).String()
				assert.Contains(t, output, "Goodbye!")
			},
		},
		{
			name:        "model command",
			commandName: "model",
			args:        []string{},
			setup: func(r *REPL) {
				r.session.Conversation.Model = "test/model"
			},
			verify: func(t *testing.T, r *REPL, w io.Writer) {
				output := w.(*bytes.Buffer).String()
				assert.Contains(t, output, "Current model: test/model")
			},
		},
		{
			name:        "config show",
			commandName: "config",
			args:        []string{"show"},
			verify: func(t *testing.T, r *REPL, w io.Writer) {
				output := w.(*bytes.Buffer).String()
				assert.Contains(t, output, "Current configuration")
			},
		},
		{
			name:        "config invalid subcommand",
			commandName: "config",
			args:        []string{"invalid"},
			expectedErr: fmt.Errorf("unknown config subcommand: invalid"),
		},
		{
			name:        "meta set",
			commandName: "meta",
			args:        []string{"set", "key", "value"},
			verify: func(t *testing.T, r *REPL, w io.Writer) {
				assert.Equal(t, "value", r.session.Metadata["key"])
			},
		},
		{
			name:        "meta del",
			commandName: "meta",
			args:        []string{"del", "key"},
			setup: func(r *REPL) {
				r.session.Metadata["key"] = "value"
			},
			verify: func(t *testing.T, r *REPL, w io.Writer) {
				_, exists := r.session.Metadata["key"]
				assert.False(t, exists)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create REPL with buffer output
			outputBuffer := &bytes.Buffer{}
			repl := createTestREPLWithOutput(t, outputBuffer)
			
			// Setup if needed
			if tt.setup != nil {
				tt.setup(repl)
			}
			
			// Find the command handler
			registry := command.NewRegistry()
			err := RegisterREPLCommands(repl, registry)
			require.NoError(t, err)
			
			// Get the command
			cmd, err := registry.Get(tt.commandName)
			require.NoError(t, err)
			
			// Create execution context
			ctx := context.Background()
			exec := &command.ExecutionContext{
				Args:   tt.args,
				Stdout: outputBuffer,
				Stderr: outputBuffer,
				Data:   make(map[string]interface{}),
			}
			
			// Execute
			err = cmd.Execute(ctx, exec)
			
			// Check error
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			
			// Verify results
			if tt.verify != nil {
				tt.verify(t, repl, outputBuffer)
			}
		})
	}
}

// Helper functions

func createTestREPL(t *testing.T) *REPL {
	return createTestREPLWithOutput(t, &bytes.Buffer{})
}

func createTestREPLWithOutput(t *testing.T, output io.Writer) *REPL {
	// Create temp directory for storage
	tempDir := t.TempDir()
	
	// Create mock config
	cfg := newCommandAdapterMockConfig()
	require.NoError(t, cfg.SetValue("model", "mock/test-model"))
	require.NoError(t, cfg.SetValue("stream", false))
	require.NoError(t, cfg.SetValue("storage.dir", tempDir))
	
	// Create REPL options
	opts := &REPLOptions{
		Config:     cfg,
		StorageDir: tempDir,
		Writer:     output,
		Reader:     strings.NewReader(""),
	}
	
	// Create REPL
	repl, err := NewREPL(opts)
	require.NoError(t, err)
	
	// Set mock provider
	repl.provider = &mockLLMProvider{}
	
	return repl
}