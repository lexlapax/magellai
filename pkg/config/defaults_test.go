// ABOUTME: Tests for default configuration generator
// ABOUTME: Verifies default values, example generation, and configuration templates

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCompleteDefaultConfig(t *testing.T) {
	// Get the default config
	config := GetCompleteDefaultConfig()

	// Test top-level sections exist
	assert.Contains(t, config, "log")
	assert.Contains(t, config, "provider")
	assert.Contains(t, config, "model")
	assert.Contains(t, config, "output")
	assert.Contains(t, config, "session")
	assert.Contains(t, config, "repl")
	assert.Contains(t, config, "plugin")
	assert.Contains(t, config, "profiles")
	assert.Contains(t, config, "aliases")
	assert.Contains(t, config, "cli")
}

func TestLogConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()

	logConfig, ok := config["log"].(map[string]interface{})
	require.True(t, ok, "log config should be a map")

	assert.Equal(t, "info", logConfig["level"])
	assert.Equal(t, "text", logConfig["format"])
}

func TestProviderConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()

	providerConfig, ok := config["provider"].(map[string]interface{})
	require.True(t, ok, "provider config should be a map")

	assert.Equal(t, "openai", providerConfig["default"])

	// Test OpenAI defaults
	openaiConfig, ok := providerConfig["openai"].(map[string]interface{})
	require.True(t, ok, "openai config should be a map")
	assert.Equal(t, "", openaiConfig["api_key"])
	assert.Equal(t, "https://api.openai.com/v1", openaiConfig["base_url"])
	assert.Equal(t, "gpt-4o", openaiConfig["default_model"])
	assert.Equal(t, "30s", openaiConfig["timeout"])
	assert.Equal(t, 3, openaiConfig["max_retries"])

	// Test Anthropic defaults
	anthropicConfig, ok := providerConfig["anthropic"].(map[string]interface{})
	require.True(t, ok, "anthropic config should be a map")
	assert.Equal(t, "", anthropicConfig["api_key"])
	assert.Equal(t, "https://api.anthropic.com", anthropicConfig["base_url"])
	assert.Equal(t, "claude-3-5-haiku-latest", anthropicConfig["default_model"])

	// Test Gemini defaults
	geminiConfig, ok := providerConfig["gemini"].(map[string]interface{})
	require.True(t, ok, "gemini config should be a map")
	assert.Equal(t, "", geminiConfig["api_key"])
	assert.Equal(t, "https://generativelanguage.googleapis.com/v1beta", geminiConfig["base_url"])
	assert.Equal(t, "gemini-2.0-flash-lite", geminiConfig["default_model"])
}

func TestModelConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()

	modelConfig, ok := config["model"].(map[string]interface{})
	require.True(t, ok, "model config should be a map")

	assert.Equal(t, "openai/gpt-4o", modelConfig["default"])

	settings, ok := modelConfig["settings"].(map[string]interface{})
	require.True(t, ok, "settings should be a map")

	// Test global settings
	globalSettings, ok := settings["*"].(map[string]interface{})
	require.True(t, ok, "global settings should be a map")
	assert.Equal(t, 0.7, globalSettings["temperature"])
	assert.Equal(t, 2048, globalSettings["max_tokens"])
	assert.Equal(t, 1.0, globalSettings["top_p"])

	// Test model-specific settings
	gpt4oSettings, ok := settings["openai/gpt-4o"].(map[string]interface{})
	require.True(t, ok, "gpt-4o settings should be a map")
	assert.Equal(t, 4096, gpt4oSettings["max_tokens"])
}

func TestOutputConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()

	outputConfig, ok := config["output"].(map[string]interface{})
	require.True(t, ok, "output config should be a map")

	assert.Equal(t, "text", outputConfig["format"])
	assert.Equal(t, true, outputConfig["color"])
	assert.Equal(t, true, outputConfig["pretty"])
}

func TestSessionConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()
	homeDir, _ := os.UserHomeDir()
	expectedSessionDir := filepath.Join(homeDir, ".config", "magellai", "sessions")

	sessionConfig, ok := config["session"].(map[string]interface{})
	require.True(t, ok, "session config should be a map")

	assert.Equal(t, expectedSessionDir, sessionConfig["directory"])
	assert.Equal(t, true, sessionConfig["autosave"])
	assert.Equal(t, "0s", sessionConfig["max_age"])
	assert.Equal(t, false, sessionConfig["compression"])

	// Test storage config
	storageConfig, ok := sessionConfig["storage"].(map[string]interface{})
	require.True(t, ok, "storage config should be a map")
	assert.Equal(t, "filesystem", storageConfig["type"])

	storageSettings, ok := storageConfig["settings"].(map[string]interface{})
	require.True(t, ok, "storage settings should be a map")
	assert.Equal(t, expectedSessionDir, storageSettings["base_dir"])

	// Test auto recovery config
	autoRecoveryConfig, ok := sessionConfig["auto_recovery"].(map[string]interface{})
	require.True(t, ok, "auto_recovery config should be a map")
	assert.Equal(t, true, autoRecoveryConfig["enabled"])
	assert.Equal(t, "30s", autoRecoveryConfig["interval"])
	assert.Equal(t, "24h", autoRecoveryConfig["max_age"])
}

func TestReplConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()
	homeDir, _ := os.UserHomeDir()
	expectedHistoryFile := filepath.Join(homeDir, ".config", "magellai", ".repl_history")

	replConfig, ok := config["repl"].(map[string]interface{})
	require.True(t, ok, "repl config should be a map")

	colors, ok := replConfig["colors"].(map[string]interface{})
	require.True(t, ok, "colors config should be a map")
	assert.Equal(t, true, colors["enabled"])

	assert.Equal(t, "> ", replConfig["prompt_style"])
	assert.Equal(t, false, replConfig["multiline"])
	assert.Equal(t, expectedHistoryFile, replConfig["history_file"])

	autoSave, ok := replConfig["auto_save"].(map[string]interface{})
	require.True(t, ok, "auto_save config should be a map")
	assert.Equal(t, true, autoSave["enabled"])
	assert.Equal(t, "5m", autoSave["interval"])
}

func TestPluginConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()
	homeDir, _ := os.UserHomeDir()
	expectedPluginDir := filepath.Join(homeDir, ".config", "magellai", "plugins")

	pluginConfig, ok := config["plugin"].(map[string]interface{})
	require.True(t, ok, "plugin config should be a map")

	assert.Equal(t, expectedPluginDir, pluginConfig["directory"])

	path, ok := pluginConfig["path"].([]string)
	require.True(t, ok, "path should be a string slice")
	assert.Empty(t, path)

	enabled, ok := pluginConfig["enabled"].([]string)
	require.True(t, ok, "enabled should be a string slice")
	assert.Empty(t, enabled)

	disabled, ok := pluginConfig["disabled"].([]string)
	require.True(t, ok, "disabled should be a string slice")
	assert.Empty(t, disabled)
}

func TestProfilesConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()

	profiles, ok := config["profiles"].(map[string]interface{})
	require.True(t, ok, "profiles should be a map")

	// Test fast profile
	fastProfile, ok := profiles["fast"].(map[string]interface{})
	require.True(t, ok, "fast profile should be a map")
	assert.Equal(t, "Fast responses with lower quality", fastProfile["description"])
	assert.Equal(t, "gemini", fastProfile["provider"])
	assert.Equal(t, "gemini-2.0-flash-lite", fastProfile["model"])

	fastSettings, ok := fastProfile["settings"].(map[string]interface{})
	require.True(t, ok, "fast settings should be a map")
	assert.Equal(t, 0.3, fastSettings["temperature"])
	assert.Equal(t, 1024, fastSettings["max_tokens"])

	// Test quality profile
	qualityProfile, ok := profiles["quality"].(map[string]interface{})
	require.True(t, ok, "quality profile should be a map")
	assert.Equal(t, "High-quality responses, slower", qualityProfile["description"])
	assert.Equal(t, "openai", qualityProfile["provider"])
	assert.Equal(t, "o3", qualityProfile["model"])

	// Test creative profile
	creativeProfile, ok := profiles["creative"].(map[string]interface{})
	require.True(t, ok, "creative profile should be a map")
	assert.Equal(t, "Creative and diverse responses", creativeProfile["description"])
	assert.Equal(t, "anthropic", creativeProfile["provider"])
	assert.Equal(t, "claude-3-7-sonnet-latest", creativeProfile["model"])
}

func TestAliasesConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()

	aliases, ok := config["aliases"].(map[string]interface{})
	require.True(t, ok, "aliases should be a map")

	assert.Equal(t, "exit", aliases["q"])
	assert.Equal(t, "exit", aliases["quit"])
	assert.Equal(t, "clear", aliases["cls"])
	assert.Equal(t, "help", aliases["h"])
	assert.Equal(t, "help", aliases["?"])
}

func TestCliConfigDefaults(t *testing.T) {
	config := GetCompleteDefaultConfig()

	cliConfig, ok := config["cli"].(map[string]interface{})
	require.True(t, ok, "cli config should be a map")

	assert.Equal(t, true, cliConfig["stream"])
	assert.Equal(t, false, cliConfig["verbose"])
	assert.Equal(t, true, cliConfig["confirm"])
}

func TestGenerateExampleConfig(t *testing.T) {
	example := GenerateExampleConfig()

	// Test that it's valid YAML-like content
	assert.Contains(t, example, "# Magellai Configuration File")
	assert.Contains(t, example, "# This is an example configuration with all available options")

	// Test major sections are present
	assert.Contains(t, example, "# Logging configuration")
	assert.Contains(t, example, "log:")
	assert.Contains(t, example, "  level: info")

	assert.Contains(t, example, "# Provider configuration")
	assert.Contains(t, example, "provider:")
	assert.Contains(t, example, "  default: openai")

	assert.Contains(t, example, "# Model configuration")
	assert.Contains(t, example, "model:")

	assert.Contains(t, example, "# Output configuration")
	assert.Contains(t, example, "output:")

	assert.Contains(t, example, "# Session configuration")
	assert.Contains(t, example, "session:")

	assert.Contains(t, example, "# REPL configuration")
	assert.Contains(t, example, "repl:")

	assert.Contains(t, example, "# Plugin configuration")
	assert.Contains(t, example, "plugin:")

	assert.Contains(t, example, "# Profiles")
	assert.Contains(t, example, "profiles:")

	assert.Contains(t, example, "# Command aliases")
	assert.Contains(t, example, "aliases:")

	assert.Contains(t, example, "# CLI settings")
	assert.Contains(t, example, "cli:")

	// Test that comments are helpful
	assert.Contains(t, example, "# Options:")
	assert.Contains(t, example, "# Set via environment variable:")
}

func TestGetConfigTemplate(t *testing.T) {
	template := GetConfigTemplate()
	example := GenerateExampleConfig()

	// Currently they should be the same
	assert.Equal(t, example, template)
}

func TestExampleConfigStructure(t *testing.T) {
	example := GenerateExampleConfig()
	lines := strings.Split(example, "\n")

	// Test that the file is properly structured with comments
	commentCount := 0
	configCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			commentCount++
		} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			configCount++
		}
	}

	// Should have a good balance of comments and config
	assert.Greater(t, commentCount, 15, "Should have many helpful comments")
	assert.Greater(t, configCount, 50, "Should have many configuration lines")
}

func TestEnvironmentVariableReferences(t *testing.T) {
	example := GenerateExampleConfig()

	// Test that environment variable references are documented
	assert.Contains(t, example, "OPENAI_API_KEY")
	assert.Contains(t, example, "ANTHROPIC_API_KEY")
	assert.Contains(t, example, "GEMINI_API_KEY")
}

func TestDefaultPaths(t *testing.T) {
	config := GetCompleteDefaultConfig()
	homeDir, _ := os.UserHomeDir()

	// Test that paths use the home directory properly
	sessionConfig := config["session"].(map[string]interface{})
	sessionDir := sessionConfig["directory"].(string)
	assert.Contains(t, sessionDir, homeDir)
	assert.Contains(t, sessionDir, ".config/magellai/sessions")

	pluginConfig := config["plugin"].(map[string]interface{})
	pluginDir := pluginConfig["directory"].(string)
	assert.Contains(t, pluginDir, homeDir)
	assert.Contains(t, pluginDir, ".config/magellai/plugins")

	replConfig := config["repl"].(map[string]interface{})
	historyFile := replConfig["history_file"].(string)
	assert.Contains(t, historyFile, homeDir)
	assert.Contains(t, historyFile, ".config/magellai/.repl_history")
}

func TestDefaultValues(t *testing.T) {
	config := GetCompleteDefaultConfig()

	// Test some specific default values are sensible
	logConfig := config["log"].(map[string]interface{})
	assert.Equal(t, "info", logConfig["level"], "Default log level should be info")

	outputConfig := config["output"].(map[string]interface{})
	assert.Equal(t, true, outputConfig["color"], "Color should be enabled by default")

	sessionConfig := config["session"].(map[string]interface{})
	assert.Equal(t, true, sessionConfig["autosave"], "Autosave should be enabled by default")

	cliConfig := config["cli"].(map[string]interface{})
	assert.Equal(t, true, cliConfig["stream"], "Streaming should be enabled by default")
}

// Benchmark tests
func BenchmarkGetCompleteDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetCompleteDefaultConfig()
	}
}

func BenchmarkGenerateExampleConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateExampleConfig()
	}
}
