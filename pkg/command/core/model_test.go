// ABOUTME: Unit tests for the model command implementation
// ABOUTME: Tests list, select, info, and validation operations

package core

import (
	"context"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelCommand_Execute(t *testing.T) {
	// Initialize config manager
	if err := config.Init(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		setupConfig   func(*config.Config)
		exec          *command.ExecutionContext
		expectedError bool
		checkOutput   func(t *testing.T, output interface{})
	}{
		{
			name: "show current model",
			setupConfig: func(cfg *config.Config) {
				_ = cfg.SetDefaultModel("openai/gpt-4")
			},
			exec: &command.ExecutionContext{
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				str, ok := output.(string)
				require.True(t, ok)
				assert.Contains(t, str, "GPT-4")
				assert.Contains(t, str, "openai/gpt-4")
			},
		},
		{
			name: "list all models",
			setupConfig: func(cfg *config.Config) {
				_ = cfg.SetDefaultModel("openai/gpt-4")
			},
			exec: &command.ExecutionContext{
				Args:  []string{"list"},
				Flags: command.NewFlags(nil),
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				str, ok := output.(string)
				require.True(t, ok)
				assert.Contains(t, str, "Available Models")
				assert.Contains(t, str, "Openai")
				assert.Contains(t, str, "Anthropic")
				assert.Contains(t, str, "Gemini")
			},
		},
		{
			name: "list models with provider filter",
			setupConfig: func(cfg *config.Config) {
				_ = cfg.SetDefaultModel("openai/gpt-4")
			},
			exec: &command.ExecutionContext{
				Args: []string{"list"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
				Flags: command.NewFlags(map[string]interface{}{
					"provider": "openai",
				}),
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				str, ok := output.(string)
				require.True(t, ok)
				assert.Contains(t, str, "Openai")
				assert.NotContains(t, str, "anthropic")
				assert.NotContains(t, str, "gemini")
			},
		},
		{
			name: "list models with capability filter",
			setupConfig: func(cfg *config.Config) {
				_ = cfg.SetDefaultModel("openai/gpt-4")
			},
			exec: &command.ExecutionContext{
				Args: []string{"list"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
				Flags: command.NewFlags(map[string]interface{}{
					"capabilities": "text,image",
				}),
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				str, ok := output.(string)
				require.True(t, ok)
				// Should only show models with both text and image capabilities
				assert.Contains(t, str, "gpt-4-vision-preview")
			},
		},
		{
			name:        "show model info",
			setupConfig: func(cfg *config.Config) {},
			exec: &command.ExecutionContext{
				Args: []string{"info", "openai/gpt-4"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				str, ok := output.(string)
				require.True(t, ok)
				assert.Contains(t, str, "Model: openai/gpt-4")
				assert.Contains(t, str, "Provider: Openai")
				assert.Contains(t, str, "Display Name: GPT-4")
				assert.Contains(t, str, "Capabilities:")
				assert.Contains(t, str, "Text processing")
			},
		},
		{
			name:        "show model info - missing argument",
			setupConfig: func(cfg *config.Config) {},
			exec: &command.ExecutionContext{
				Args: []string{"info"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: true,
		},
		{
			name:        "show model info - invalid model",
			setupConfig: func(cfg *config.Config) {},
			exec: &command.ExecutionContext{
				Args: []string{"info", "invalid/model"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: true,
		},
		{
			name: "select model",
			setupConfig: func(cfg *config.Config) {
				_ = cfg.SetDefaultModel("openai/gpt-3.5-turbo")
			},
			exec: &command.ExecutionContext{
				Args: []string{"anthropic/claude-3-opus"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				str, ok := output.(string)
				require.True(t, ok)
				assert.Contains(t, str, "Switched to Claude 3 Opus")
				assert.Contains(t, str, "anthropic/claude-3-opus")
			},
		},
		{
			name:        "select model - invalid format",
			setupConfig: func(cfg *config.Config) {},
			exec: &command.ExecutionContext{
				Args: []string{"invalid-format"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: true,
		},
		{
			name:        "select model - non-existent",
			setupConfig: func(cfg *config.Config) {},
			exec: &command.ExecutionContext{
				Args: []string{"provider/nonexistent"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: true,
		},
		{
			name: "no model selected",
			setupConfig: func(cfg *config.Config) {
				_ = cfg.SetDefaultModel("")
			},
			exec: &command.ExecutionContext{
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				str, ok := output.(string)
				require.True(t, ok)
				assert.Equal(t, "No model selected", str)
			},
		},
		{
			name: "json output - current model",
			setupConfig: func(cfg *config.Config) {
				_ = cfg.SetDefaultModel("openai/gpt-4")
			},
			exec: &command.ExecutionContext{
				Data: map[string]interface{}{
					"outputFormat": OutputFormatJSON,
				},
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				data, ok := output.(map[string]string)
				require.True(t, ok)
				assert.Equal(t, "openai", data["provider"])
				assert.Equal(t, "openai/gpt-4", data["model"])
				assert.Equal(t, "GPT-4", data["display_name"])
			},
		},
		{
			name:        "json output - list models",
			setupConfig: func(cfg *config.Config) {},
			exec: &command.ExecutionContext{
				Args:  []string{"list"},
				Flags: command.NewFlags(nil),
				Data: map[string]interface{}{
					"outputFormat": OutputFormatJSON,
				},
			},
			expectedError: false,
			checkOutput: func(t *testing.T, output interface{}) {
				// Should be a slice of ModelInfo structs
				models, ok := output.([]llm.ModelInfo)
				require.True(t, ok, "Expected []llm.ModelInfo, got %T", output)
				assert.NotEmpty(t, models)
			},
		},
		{
			name:        "unknown subcommand",
			setupConfig: func(cfg *config.Config) {},
			exec: &command.ExecutionContext{
				Args: []string{"unknown"},
				Data: map[string]interface{}{
					"outputFormat": OutputFormatText,
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			if tt.setupConfig != nil {
				tt.setupConfig(config.Manager)
			}

			// Create command
			cmd := NewModelCommand(config.Manager)

			// Execute
			err := cmd.Execute(context.Background(), tt.exec)

			// Check error
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check output
			if tt.checkOutput != nil && !tt.expectedError {
				tt.checkOutput(t, tt.exec.Data["output"])
			}
		})
	}
}

func TestModelCommand_Metadata(t *testing.T) {
	cmd := NewModelCommand(&config.Config{})
	meta := cmd.Metadata()

	assert.Equal(t, "model", meta.Name)
	assert.NotEmpty(t, meta.Description)
	assert.Equal(t, command.CategoryShared, meta.Category)
	assert.Len(t, meta.Flags, 2)
	assert.NotEmpty(t, meta.LongDescription)
}

func TestModelCommand_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedError bool
	}{
		{
			name:          "valid config",
			config:        &config.Config{},
			expectedError: false,
		},
		{
			name:          "nil config",
			config:        nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ModelCommand{config: tt.config}
			err := cmd.Validate()
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
