// ABOUTME: Test for runtime configuration display
// ABOUTME: Verifies that config show displays all runtime values including defaults

package core

import (
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigShowRuntimeValues(t *testing.T) {
	// Initialize config with defaults
	t.Setenv("MAGELLAI_LOG_LEVEL", "error")
	err := config.Init()
	require.NoError(t, err)

	// Load defaults
	err = config.Manager.Load(nil)
	require.NoError(t, err)

	// Config command will be created per test

	tests := []struct {
		name           string
		setupFunc      func()
		expectedValues []string
	}{
		{
			name: "shows default values",
			setupFunc: func() {
				// No additional setup - just defaults
			},
			expectedValues: []string{
				"log.",
				// "logging.", // The config now uses "log" instead of "logging" as the prefix
				"provider.",
				"default",
				"openai",
				"model.",
				"session.",
				"repl.",
				"profiles.",
			},
		},
		{
			name: "shows runtime overrides",
			setupFunc: func() {
				// Set some runtime values
				_ = config.Manager.SetDefaultProvider("anthropic")
				_ = config.Manager.SetDefaultModel("claude-3")
				_ = config.Manager.SetValue("custom.key", "custom-value")
			},
			expectedValues: []string{
				"provider.default: anthropic",
				"model.default: claude-3",
				"custom.key: custom-value",
			},
		},
		{
			name: "shows environment variable overrides",
			setupFunc: func() {
				// Environment variables are already applied during Load()
				t.Setenv("MAGELLAI_PROVIDER_DEFAULT", "gemini")
				// Reload to pick up the env var
				_ = config.Manager.Load(nil)
			},
			expectedValues: []string{
				"provider.default: gemini",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset config for each test
			err := config.Init()
			require.NoError(t, err)
			err = config.Manager.Load(nil)
			require.NoError(t, err)

			// Apply test-specific setup
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Create a new config command with the updated config
			cmd := NewConfigCommand(config.Manager)

			// Execute config show (list)
			exec := &command.ExecutionContext{
				Args:   []string{"list"},
				Flags:  command.NewFlags(nil),
				Stdout: &strings.Builder{},
				Stderr: &strings.Builder{},
				Data:   make(map[string]interface{}),
			}

			err = cmd.Execute(context.TODO(), exec)
			require.NoError(t, err)

			output := exec.Data["output"].(string)

			// Verify all expected values are present
			for _, expected := range tt.expectedValues {
				assert.Contains(t, output, expected, "Expected value '%s' not found in output", expected)
			}

			// Log output for debugging
			t.Logf("Config output:\n%s", output)
		})
	}
}

func TestConfigShowFormats(t *testing.T) {
	// Initialize config
	t.Setenv("MAGELLAI_LOG_LEVEL", "error")
	err := config.Init()
	require.NoError(t, err)
	err = config.Manager.Load(nil)
	require.NoError(t, err)

	cmd := NewConfigCommand(config.Manager)

	formats := []string{"text", "json", "yaml"}

	for _, format := range formats {
		t.Run(format+" format", func(t *testing.T) {
			exec := &command.ExecutionContext{
				Args:  []string{"list"},
				Flags: command.NewFlags(map[string]interface{}{"format": format}),
				Data:  make(map[string]interface{}),
			}

			err := cmd.Execute(context.TODO(), exec)
			require.NoError(t, err)

			output := exec.Data["output"].(string)
			assert.NotEmpty(t, output)

			// Check format-specific markers
			switch format {
			case "json":
				assert.Contains(t, output, "{")
				assert.Contains(t, output, "}")
			case "yaml":
				assert.Contains(t, output, ":")
			case "text":
				assert.Contains(t, output, "Configuration settings:")
			}

			// All formats should show key runtime values
			assert.Contains(t, output, "provider")
			assert.Contains(t, output, "model")
			assert.Contains(t, output, "log")
		})
	}
}
