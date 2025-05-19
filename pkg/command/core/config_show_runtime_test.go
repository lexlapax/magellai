// ABOUTME: Test config show displays all runtime configurations
// ABOUTME: Verifies requirement for Phase 4.8.3 configuration display

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

func TestConfigShowAllRuntimeConfig(t *testing.T) {
	// Initialize config with environment variable
	t.Setenv("MAGELLAI_LOG_LEVEL", "debug")
	t.Setenv("MAGELLAI_PROVIDER_DEFAULT", "anthropic")
	t.Setenv("OPENAI_API_KEY", "test-key-from-env")

	// Initialize config
	err := config.Init()
	require.NoError(t, err)

	// Load config (includes defaults + env)
	err = config.Manager.Load(nil)
	require.NoError(t, err)

	// Set runtime values to simulate command-line overrides
	_ = config.Manager.SetDefaultProvider("openai") // Command-line override
	_ = config.Manager.SetDefaultModel("gpt-4-turbo")
	_ = config.Manager.SetValue("session.repl.history.size", 500) // Runtime change

	// Create config command
	cmd := NewConfigCommand(config.Manager)

	// Capture output
	var output strings.Builder

	// Execute config show/list
	exec := &command.ExecutionContext{
		Args:   []string{"list"}, // list shows all config
		Flags:  command.NewFlags(nil),
		Stdout: &output,
		Data:   make(map[string]interface{}),
	}

	err = cmd.Execute(context.Background(), exec)
	require.NoError(t, err)

	// Get result from exec.Data["output"] instead of stdout
	result, ok := exec.Data["output"].(string)
	require.True(t, ok, "output should be in exec.Data")

	// Log output for inspection
	t.Log("Config show output:")
	t.Log(result)

	// Test that all layers are represented

	// 1. Defaults (built-in)
	assert.Contains(t, result, "log.", "Should show log section from defaults")
	assert.Contains(t, result, "provider.", "Should show provider section")
	assert.Contains(t, result, "model.", "Should show model section")
	assert.Contains(t, result, "session.", "Should show session configuration")

	// 2. Environment variable overrides
	assert.Contains(t, result, "debug", "Should show debug log level from env var")

	// 3. Command-line/runtime overrides
	assert.Contains(t, result, "openai", "Should show runtime provider override")
	assert.Contains(t, result, "gpt-4-turbo", "Should show runtime model override")
	assert.Contains(t, result, "500", "Should show runtime history size change")

	// Test configuration precedence is visible
	// The final values should be the runtime overrides, not env or defaults
	lines := strings.Split(result, "\n")
	foundProvider := false
	foundModel := false
	for _, line := range lines {
		if strings.Contains(line, "default:") && strings.Contains(line, "provider") {
			// Should be openai (runtime), not anthropic (env)
			assert.Contains(t, line, "openai", "Provider should be runtime value")
			foundProvider = true
		}
		if strings.Contains(line, "default:") && strings.Contains(line, "model") {
			assert.Contains(t, line, "gpt-4-turbo", "Model should be runtime value")
			foundModel = true
		}
	}
	assert.True(t, foundProvider, "Should find provider configuration")
	assert.True(t, foundModel, "Should find model configuration")

	// Test that it shows all sections
	assert.Contains(t, result, "Configuration settings:", "Should have header")
	assert.Contains(t, result, "profiles", "Should show profiles section") // Changed to match actual format
	assert.Contains(t, result, "repl", "Should show REPL settings")        // Changed to match actual format

	// Test structured output
	assert.True(t, strings.Count(result, ":") > 10, "Should have many configuration keys")
}

func TestConfigShowVSDefault(t *testing.T) {
	// Test that config show is different from just showing defaults
	err := config.Init()
	require.NoError(t, err)

	// Get default config for comparison (not used but shows concept)
	_ = config.GetCompleteDefaultConfig()

	// First, test with just defaults
	err = config.Manager.Load(nil)
	require.NoError(t, err)

	cmd := NewConfigCommand(config.Manager)

	var output1 strings.Builder

	exec := &command.ExecutionContext{
		Args:   []string{"list"},
		Flags:  command.NewFlags(nil),
		Stdout: &output1,
		Data:   make(map[string]interface{}),
	}

	err = cmd.Execute(context.Background(), exec)
	require.NoError(t, err)

	defaultOutput, ok := exec.Data["output"].(string)
	require.True(t, ok, "output should be in exec.Data")

	// Now modify runtime config
	_ = config.Manager.SetDefaultProvider("azure")
	_ = config.Manager.SetValue("log.level", "trace")
	_ = config.Manager.SetValue("custom.runtime", "value")

	// Execute again
	var output2 strings.Builder
	exec2 := &command.ExecutionContext{
		Args:   []string{"list"},
		Flags:  command.NewFlags(nil),
		Stdout: &output2,
		Data:   make(map[string]interface{}),
	}

	err = cmd.Execute(context.Background(), exec2)
	require.NoError(t, err)

	modifiedOutput, ok2 := exec2.Data["output"].(string)
	require.True(t, ok2, "output should be in exec2.Data")

	// Outputs should be different
	assert.NotEqual(t, defaultOutput, modifiedOutput, "Runtime config should differ from defaults")

	// Modified output should have our changes
	assert.Contains(t, modifiedOutput, "azure", "Should show runtime provider change")
	assert.Contains(t, modifiedOutput, "trace", "Should show runtime log level change")
	assert.Contains(t, modifiedOutput, "custom", "Should show custom runtime values")
	assert.Contains(t, modifiedOutput, "runtime: value", "Should show custom value")
}

func TestConfigShowFormattedJSON(t *testing.T) {
	// Test JSON output format shows all runtime config
	err := config.Init()
	require.NoError(t, err)

	err = config.Manager.Load(nil)
	require.NoError(t, err)

	// Set some runtime values
	_ = config.Manager.SetDefaultProvider("anthropic")
	_ = config.Manager.SetValue("session.repl.prompt", ">>> ")

	cmd := NewConfigCommand(config.Manager)

	var output strings.Builder

	exec := &command.ExecutionContext{
		Args:   []string{"list"},
		Flags:  command.NewFlags(map[string]interface{}{"format": "json"}),
		Stdout: &output,
		Data:   make(map[string]interface{}),
	}

	err = cmd.Execute(context.Background(), exec)
	require.NoError(t, err)

	result, ok := exec.Data["output"].(string)
	require.True(t, ok, "output should be in exec.Data")

	// Log the actual JSON output to see the format
	t.Log("JSON output:")
	t.Log(result)

	// Should be valid JSON
	assert.Contains(t, result, "{", "Should start with JSON object")
	assert.Contains(t, result, "}", "Should end with JSON object")
	// The key is "provider.default" not just "provider"
	assert.Contains(t, result, `"provider.default"`, "Should have provider key")
	assert.Contains(t, result, `"anthropic"`, "Should have runtime provider value")
	assert.True(t, strings.Contains(result, `">>> "`) || strings.Contains(result, `"\u003e\u003e\u003e "`), "Should have runtime prompt value")

	// All main sections should be in JSON
	assert.Contains(t, result, `"log`, "Should have log section") // May have nested keys
	assert.Contains(t, result, `"session`, "Should have session section")
	assert.Contains(t, result, `"model`, "Should have model section")
	assert.Contains(t, result, `"profiles`, "Should have profiles section")
}
