// ABOUTME: Tests for configuration schema types and parsing
// ABOUTME: Validates struct definitions and schema type behaviors

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseProviderModelExtended(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantProvider string
		wantModel    string
		description  string
	}{
		{
			name:         "Full provider/model format",
			input:        "openai/gpt-4",
			wantProvider: "openai",
			wantModel:    "gpt-4",
			description:  "Should split on first slash",
		},
		{
			name:         "Model only format",
			input:        "gpt-4",
			wantProvider: "",
			wantModel:    "gpt-4",
			description:  "Should treat as model-only when no slash",
		},
		{
			name:         "Provider with model containing slash",
			input:        "anthropic/claude-3.5-sonnet",
			wantProvider: "anthropic",
			wantModel:    "claude-3.5-sonnet",
			description:  "Should only split on first slash",
		},
		{
			name:         "Empty string",
			input:        "",
			wantProvider: "",
			wantModel:    "",
			description:  "Should handle empty input",
		},
		{
			name:         "Only slash",
			input:        "/",
			wantProvider: "",
			wantModel:    "",
			description:  "Should handle slash-only input",
		},
		{
			name:         "Multiple slashes",
			input:        "provider/model/variant",
			wantProvider: "provider",
			wantModel:    "model/variant",
			description:  "Should preserve slashes after first split",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseProviderModel(tt.input)
			assert.Equal(t, tt.wantProvider, got.Provider, "Provider mismatch: %s", tt.description)
			assert.Equal(t, tt.wantModel, got.Model, "Model mismatch: %s", tt.description)
		})
	}
}

func TestSchemaStructures(t *testing.T) {
	t.Run("Full Schema creation", func(t *testing.T) {
		schema := Schema{
			Log: LogConfig{
				Level:  "info",
				Format: "json",
			},
			Provider: ProviderConfig{
				Default: "openai",
				OpenAI: &OpenAIConfig{
					APIKey:       "test-key",
					BaseURL:      "https://api.openai.com",
					DefaultModel: "gpt-4",
					Timeout:      30 * time.Second,
				},
			},
			Model: ModelConfig{
				Default: "openai/gpt-4",
				Settings: map[string]ModelSettings{
					"openai/gpt-4": {
						Temperature: pointerFloat64(0.7),
						MaxTokens:   pointerInt(2048),
					},
				},
			},
			Output: OutputConfig{
				Format: "text",
				Color:  true,
				Pretty: true,
			},
			Session: SessionConfig{
				Directory:   "/tmp/sessions",
				AutoSave:    true,
				MaxAge:      24 * time.Hour,
				Compression: true,
			},
			Profiles: map[string]ProfileConfig{
				"default": {
					Description: "Default profile",
					Provider:    "openai",
					Model:       "gpt-4",
				},
			},
			Aliases: map[string]string{
				"gpt": "openai/gpt-4",
			},
		}

		// Verify structure
		assert.Equal(t, "info", schema.Log.Level)
		assert.Equal(t, "openai", schema.Provider.Default)
		assert.NotNil(t, schema.Provider.OpenAI)
		assert.Equal(t, "test-key", schema.Provider.OpenAI.APIKey)
		assert.Equal(t, "openai/gpt-4", schema.Model.Default)
		assert.Contains(t, schema.Model.Settings, "openai/gpt-4")
		assert.Equal(t, float64(0.7), *schema.Model.Settings["openai/gpt-4"].Temperature)
		assert.True(t, schema.Output.Color)
		assert.True(t, schema.Session.AutoSave)
		assert.Contains(t, schema.Profiles, "default")
		assert.Equal(t, "gpt", "gpt")
		assert.Equal(t, "openai/gpt-4", schema.Aliases["gpt"])
	})
}

func TestModelSettings(t *testing.T) {
	t.Run("ModelSettings with nil values", func(t *testing.T) {
		settings := ModelSettings{
			Temperature: nil,
			MaxTokens:   pointerInt(1000),
			TopP:        pointerFloat64(0.9),
		}

		assert.Nil(t, settings.Temperature)
		assert.NotNil(t, settings.MaxTokens)
		assert.Equal(t, 1000, *settings.MaxTokens)
		assert.NotNil(t, settings.TopP)
		assert.Equal(t, 0.9, *settings.TopP)
	})

	t.Run("ModelSettings with extra parameters", func(t *testing.T) {
		settings := ModelSettings{
			Temperature: pointerFloat64(0.8),
			MaxTokens:   pointerInt(2000),
			Extra: map[string]interface{}{
				"custom_param": "value",
				"another":      123,
			},
		}

		assert.Equal(t, 0.8, *settings.Temperature)
		assert.Equal(t, 2000, *settings.MaxTokens)
		assert.Equal(t, "value", settings.Extra["custom_param"])
		assert.Equal(t, 123, settings.Extra["another"])
	})
}

func TestProviderConfigs(t *testing.T) {
	t.Run("OpenAIConfig", func(t *testing.T) {
		config := OpenAIConfig{
			APIKey:       "sk-test",
			BaseURL:      "https://api.openai.com",
			Organization: "org-123",
			APIVersion:   "v1",
			DefaultModel: "gpt-4",
			Timeout:      30 * time.Second,
			MaxRetries:   3,
		}

		assert.Equal(t, "sk-test", config.APIKey)
		assert.Equal(t, "https://api.openai.com", config.BaseURL)
		assert.Equal(t, "org-123", config.Organization)
		assert.Equal(t, "v1", config.APIVersion)
		assert.Equal(t, "gpt-4", config.DefaultModel)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, 3, config.MaxRetries)
	})

	t.Run("AnthropicConfig", func(t *testing.T) {
		config := AnthropicConfig{
			APIKey:       "sk-ant-test",
			BaseURL:      "https://api.anthropic.com",
			APIVersion:   "2023-06-01",
			DefaultModel: "claude-3-sonnet",
			Timeout:      60 * time.Second,
			MaxRetries:   5,
		}

		assert.Equal(t, "sk-ant-test", config.APIKey)
		assert.Equal(t, "https://api.anthropic.com", config.BaseURL)
		assert.Equal(t, "2023-06-01", config.APIVersion)
		assert.Equal(t, "claude-3-sonnet", config.DefaultModel)
		assert.Equal(t, 60*time.Second, config.Timeout)
		assert.Equal(t, 5, config.MaxRetries)
	})

	t.Run("GeminiConfig", func(t *testing.T) {
		config := GeminiConfig{
			APIKey:       "gemini-key",
			BaseURL:      "https://generativelanguage.googleapis.com",
			ProjectID:    "project-123",
			Location:     "us-central1",
			DefaultModel: "gemini-pro",
			Timeout:      45 * time.Second,
			MaxRetries:   4,
		}

		assert.Equal(t, "gemini-key", config.APIKey)
		assert.Equal(t, "https://generativelanguage.googleapis.com", config.BaseURL)
		assert.Equal(t, "project-123", config.ProjectID)
		assert.Equal(t, "us-central1", config.Location)
		assert.Equal(t, "gemini-pro", config.DefaultModel)
		assert.Equal(t, 45*time.Second, config.Timeout)
		assert.Equal(t, 4, config.MaxRetries)
	})
}

func TestSessionAndStorageConfig(t *testing.T) {
	t.Run("SessionConfig with StorageConfig", func(t *testing.T) {
		config := SessionConfig{
			Directory:   "/var/lib/magellai/sessions",
			AutoSave:    true,
			MaxAge:      7 * 24 * time.Hour,
			Compression: false,
			Storage: StorageConfig{
				Type: "sqlite",
				Settings: map[string]interface{}{
					"path":     "/var/lib/magellai/sessions.db",
					"timeout":  "30s",
					"max_conn": 10,
				},
			},
		}

		assert.Equal(t, "/var/lib/magellai/sessions", config.Directory)
		assert.True(t, config.AutoSave)
		assert.Equal(t, 7*24*time.Hour, config.MaxAge)
		assert.False(t, config.Compression)
		assert.Equal(t, "sqlite", config.Storage.Type)
		assert.Equal(t, "/var/lib/magellai/sessions.db", config.Storage.Settings["path"])
		assert.Equal(t, "30s", config.Storage.Settings["timeout"])
		assert.Equal(t, 10, config.Storage.Settings["max_conn"])
	})
}

func TestProfileConfig(t *testing.T) {
	profiles := map[string]ProfileConfig{
		"development": {
			Description: "Development environment settings",
			Provider:    "openai",
			Model:       "gpt-4o",
			Settings: map[string]interface{}{
				"temperature": 0.9,
				"max_tokens":  1000,
				"stream":      true,
			},
		},
		"production": {
			Description: "Production environment settings",
			Provider:    "anthropic",
			Model:       "claude-3-opus",
			Settings: map[string]interface{}{
				"temperature": 0.3,
				"max_tokens":  4000,
				"safe_mode":   true,
			},
		},
	}

	assert.Len(t, profiles, 2)

	dev := profiles["development"]
	assert.Equal(t, "Development environment settings", dev.Description)
	assert.Equal(t, "openai", dev.Provider)
	assert.Equal(t, "gpt-4o", dev.Model)
	assert.Equal(t, 0.9, dev.Settings["temperature"])
	assert.Equal(t, 1000, dev.Settings["max_tokens"])
	assert.Equal(t, true, dev.Settings["stream"])

	prod := profiles["production"]
	assert.Equal(t, "Production environment settings", prod.Description)
	assert.Equal(t, "anthropic", prod.Provider)
	assert.Equal(t, "claude-3-opus", prod.Model)
	assert.Equal(t, 0.3, prod.Settings["temperature"])
	assert.Equal(t, 4000, prod.Settings["max_tokens"])
	assert.Equal(t, true, prod.Settings["safe_mode"])
}

func TestPluginConfig(t *testing.T) {
	config := PluginConfig{
		Directory: "/usr/local/share/magellai/plugins",
		Path: []string{
			"/home/user/.magellai/plugins",
			"/opt/magellai/plugins",
		},
		Enabled: []string{
			"code-completion",
			"syntax-highlighting",
			"git-integration",
		},
		Disabled: []string{
			"experimental-plugin",
			"deprecated-feature",
		},
	}

	assert.Equal(t, "/usr/local/share/magellai/plugins", config.Directory)
	assert.Len(t, config.Path, 2)
	assert.Contains(t, config.Path, "/home/user/.magellai/plugins")
	assert.Contains(t, config.Path, "/opt/magellai/plugins")
	assert.Len(t, config.Enabled, 3)
	assert.Contains(t, config.Enabled, "code-completion")
	assert.Len(t, config.Disabled, 2)
	assert.Contains(t, config.Disabled, "experimental-plugin")
}

// Helper functions
func pointerInt(i int) *int {
	return &i
}

func pointerFloat64(f float64) *float64 {
	return &f
}
