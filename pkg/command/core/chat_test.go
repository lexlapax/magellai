package core

import (
	"bytes"
	"context"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatCommand(t *testing.T) {
	cfg := createTestConfig(t)
	cmd := NewChatCommand(cfg)

	t.Run("metadata", func(t *testing.T) {
		meta := cmd.Metadata()
		assert.Equal(t, "chat", meta.Name)
		assert.Equal(t, "Start an interactive chat session with the LLM", meta.Description)
		assert.Equal(t, command.CategoryCLI, meta.Category)
		require.Len(t, meta.Flags, 3)

		// Check flags
		flags := meta.Flags
		assert.Equal(t, "resume", flags[0].Name)
		assert.Equal(t, "r", flags[0].Short)
		assert.Equal(t, command.FlagTypeString, flags[0].Type)

		assert.Equal(t, "model", flags[1].Name)
		assert.Equal(t, "m", flags[1].Short)
		assert.Equal(t, command.FlagTypeString, flags[1].Type)

		assert.Equal(t, "attach", flags[2].Name)
		assert.Equal(t, "a", flags[2].Short)
		assert.Equal(t, command.FlagTypeStringSlice, flags[2].Type)
	})

	t.Run("validate", func(t *testing.T) {
		err := cmd.Validate()
		assert.NoError(t, err)
	})

	// We can't easily test Execute since it creates a REPL and runs interactively
	// This would require mocking the REPL or creating special test modes
}

// TestChatCommandAPIKeys verifies that the chat command correctly uses API keys from config
func TestChatCommandAPIKeys(t *testing.T) {
	// Initialize config
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.Manager

	// Configure the model and API key
	if err := cfg.SetValue("model.default", "mock/test-model"); err != nil {
		t.Fatalf("Failed to set model.default: %v", err)
	}

	if err := cfg.SetValue("provider.mock.api_key", "mock-api-key"); err != nil {
		t.Fatalf("Failed to set provider.mock.api_key: %v", err)
	}

	_ = NewChatCommand(cfg)

	// We can't fully test the chat REPL execution, but we can verify
	// the command implementation passes config properly to the REPL

	// Check that the config adapter will return expected values
	adapter := &replConfigAdapter{cfg}

	modelName := adapter.GetString("model.default")
	assert.Equal(t, "mock/test-model", modelName)

	apiKey := adapter.GetString("provider.mock.api_key")
	assert.Equal(t, "mock-api-key", apiKey)
}

// TestChatCommandIntegration would require a more sophisticated test setup
// possibly with mock stdin/stdout and a test REPL implementation
func TestChatCommandIntegration(t *testing.T) {
	// Skip for now as it requires interactive REPL
	t.Skip("Chat command requires interactive REPL - needs special test mode")

	cfg := createTestConfig(t)
	cmd := NewChatCommand(cfg)

	ctx := context.Background()
	exec := &command.ExecutionContext{
		Args:    []string{},
		Flags:   command.NewFlags(nil),
		Stdout:  new(bytes.Buffer),
		Stderr:  new(bytes.Buffer),
		Context: ctx,
	}

	// Would need to mock stdin and handle REPL interaction
	err := cmd.Execute(ctx, exec)
	assert.Error(t, err) // Would fail due to lack of proper stdin
}
