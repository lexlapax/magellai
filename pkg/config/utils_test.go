// ABOUTME: Tests for configuration utility functions
// ABOUTME: Validates type-safe access methods and configuration operations

package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigGetters(t *testing.T) {
	// Create a test configuration
	config := createTestConfig(t)

	t.Run("GetString", func(t *testing.T) {
		value := config.GetString("provider.default")
		assert.Equal(t, "openai", value)

		// Test non-existent key
		value = config.GetString("nonexistent.key")
		assert.Empty(t, value)
	})

	t.Run("GetInt", func(t *testing.T) {
		// Set an int value
		err := config.SetValue("test.int", 42)
		require.NoError(t, err)

		value := config.GetInt("test.int")
		assert.Equal(t, 42, value)

		// Test default value for non-existent key
		value = config.GetInt("nonexistent.int")
		assert.Equal(t, 0, value)
	})

	t.Run("GetBool", func(t *testing.T) {
		// Set a bool value
		err := config.SetValue("output.color", true)
		require.NoError(t, err)

		value := config.GetBool("output.color")
		assert.True(t, value)

		// Test default value for non-existent key
		value = config.GetBool("nonexistent.bool")
		assert.False(t, value)
	})

	t.Run("GetFloat64", func(t *testing.T) {
		// Set a float value
		err := config.SetValue("model.temperature", 0.7)
		require.NoError(t, err)

		value := config.GetFloat64("model.temperature")
		assert.Equal(t, 0.7, value)

		// Test default value for non-existent key
		value = config.GetFloat64("nonexistent.float")
		assert.Equal(t, 0.0, value)
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		// Set a string slice
		err := config.SetValue("plugin.enabled", []string{"plugin1", "plugin2"})
		require.NoError(t, err)

		value := config.GetStringSlice("plugin.enabled")
		assert.Len(t, value, 2)
		assert.Contains(t, value, "plugin1")
		assert.Contains(t, value, "plugin2")

		// Test empty slice for non-existent key
		value = config.GetStringSlice("nonexistent.slice")
		assert.Empty(t, value)
	})

	t.Run("GetDuration", func(t *testing.T) {
		// Set a duration value
		err := config.SetValue("session.max_age", "24h")
		require.NoError(t, err)

		value := config.GetDuration("session.max_age")
		assert.Equal(t, 24*time.Hour, value)

		// Test default value for non-existent key
		value = config.GetDuration("nonexistent.duration")
		assert.Equal(t, time.Duration(0), value)
	})

	t.Run("Get", func(t *testing.T) {
		// Test generic get
		err := config.SetValue("test.generic", map[string]interface{}{
			"nested": "value",
		})
		require.NoError(t, err)

		value := config.Get("test.generic")
		assert.NotNil(t, value)
		mapValue, ok := value.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value", mapValue["nested"])
	})

	t.Run("Exists", func(t *testing.T) {
		// Test existing key
		exists := config.Exists("provider.default")
		assert.True(t, exists)

		// Test non-existent key
		exists = config.Exists("nonexistent.key")
		assert.False(t, exists)
	})

	t.Run("All", func(t *testing.T) {
		all := config.All()
		assert.NotEmpty(t, all)
		// Check for the flat key that was set
		assert.Contains(t, all, "provider.default")
	})
}

func TestConfigSetters(t *testing.T) {
	config := createTestConfig(t)

	t.Run("SetValue", func(t *testing.T) {
		// Set various types
		err := config.SetValue("test.string", "test value")
		assert.NoError(t, err)
		assert.Equal(t, "test value", config.GetString("test.string"))

		err = config.SetValue("test.int", 123)
		assert.NoError(t, err)
		assert.Equal(t, 123, config.GetInt("test.int"))

		err = config.SetValue("test.bool", true)
		assert.NoError(t, err)
		assert.True(t, config.GetBool("test.bool"))

		// Test nested value
		err = config.SetValue("test.nested.value", "nested")
		assert.NoError(t, err)
		assert.Equal(t, "nested", config.GetString("test.nested.value"))
	})

	t.Run("SetDefaultProvider", func(t *testing.T) {
		err := config.SetDefaultProvider("anthropic")
		assert.NoError(t, err)
		assert.Equal(t, "anthropic", config.GetDefaultProvider())
	})

	t.Run("SetDefaultModel", func(t *testing.T) {
		err := config.SetDefaultModel("claude-3-sonnet")
		assert.NoError(t, err)
		assert.Equal(t, "claude-3-sonnet", config.GetDefaultModel())
	})
}

func TestProviderConfig(t *testing.T) {
	config := createTestConfig(t)

	t.Run("GetProviderConfig", func(t *testing.T) {
		// Set up test provider config
		err := config.SetValue("provider.testprovider", map[string]interface{}{
			"api_key": "test-key",
			"timeout": "30s",
		})
		require.NoError(t, err)

		providerConfig, err := config.GetProviderConfig("testprovider")
		assert.NoError(t, err)
		assert.Equal(t, "test-key", providerConfig["api_key"])
		assert.Equal(t, "30s", providerConfig["timeout"])

		// Test non-existent provider
		_, err = config.GetProviderConfig("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("GetProviderAPIKey", func(t *testing.T) {
		// Test from config
		err := config.SetValue("provider.openai.api_key", "config-key")
		require.NoError(t, err)

		apiKey := config.GetProviderAPIKey("openai")
		assert.Equal(t, "config-key", apiKey)

		// Test from environment variable
		os.Setenv("MAGELLAI_ANTHROPIC_API_KEY", "env-key")
		defer os.Unsetenv("MAGELLAI_ANTHROPIC_API_KEY")

		apiKey = config.GetProviderAPIKey("anthropic")
		assert.Equal(t, "env-key", apiKey)

		// Test precedence (env over config)
		err = config.SetValue("provider.anthropic.api_key", "config-key-2")
		require.NoError(t, err)
		apiKey = config.GetProviderAPIKey("anthropic")
		assert.Equal(t, "env-key", apiKey)
	})
}

func TestModelSettingsUtils(t *testing.T) {
	config := createTestConfig(t)

	t.Run("GetModelSettings", func(t *testing.T) {
		// Set up test model settings
		err := config.SetValue("model.settings.openai/gpt-4", map[string]interface{}{
			"temperature": 0.8,
			"max_tokens":  2000,
		})
		require.NoError(t, err)

		settings, err := config.GetModelSettings("openai/gpt-4")
		assert.NoError(t, err)
		assert.NotNil(t, settings)

		// Test non-existent model (should return default settings)
		settings, err = config.GetModelSettings("nonexistent/model")
		assert.NoError(t, err)
		assert.NotNil(t, settings)
		assert.Nil(t, settings.Temperature)
		assert.Nil(t, settings.MaxTokens)
	})
}

func TestProfileOperations(t *testing.T) {
	config := createTestConfig(t)

	t.Run("GetProfile", func(t *testing.T) {
		// Set up test profile
		err := config.SetValue("profiles.test", map[string]interface{}{
			"description": "Test profile",
			"provider":    "openai",
			"model":       "gpt-4",
			"settings": map[string]interface{}{
				"temperature": 0.9,
			},
		})
		require.NoError(t, err)

		profile, err := config.GetProfile("test")
		assert.NoError(t, err)
		assert.Equal(t, "Test profile", profile.Description)
		assert.Equal(t, "openai", profile.Provider)
		assert.Equal(t, "gpt-4", profile.Model)
		assert.Equal(t, 0.9, profile.Settings["temperature"])

		// Test non-existent profile
		_, err = config.GetProfile("nonexistent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrProfileNotFound)
	})

	t.Run("MergeProfile", func(t *testing.T) {
		// Set up profile
		err := config.SetValue("profiles.merge-test", map[string]interface{}{
			"provider": "anthropic",
			"model":    "claude-3-opus",
			"settings": map[string]interface{}{
				"output.color": false,
				"log.level":    "debug",
			},
		})
		require.NoError(t, err)

		// Merge profile
		err = config.MergeProfile("merge-test")
		assert.NoError(t, err)

		// Verify merged settings
		assert.Equal(t, "anthropic", config.GetDefaultProvider())
		assert.Equal(t, "claude-3-opus", config.GetDefaultModel())
		assert.False(t, config.GetBool("output.color"))
		assert.Equal(t, "debug", config.GetString("log.level"))

		// Test non-existent profile
		err = config.MergeProfile("nonexistent")
		assert.Error(t, err)
	})
}

func TestConfigExportImport(t *testing.T) {
	config := createTestConfig(t)

	t.Run("Export", func(t *testing.T) {
		// Add some test data
		err := config.SetValue("test.export", "value")
		require.NoError(t, err)

		data, err := config.Export()
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, string(data), "test:")
		assert.Contains(t, string(data), "export: value")
	})

	t.Run("Import", func(t *testing.T) {
		yamlData := []byte(`
provider:
  default: imported
test:
  imported: true
`)

		err := config.Import(yamlData)
		assert.NoError(t, err)

		// Verify imported data
		assert.Equal(t, "imported", config.GetDefaultProvider())
		assert.True(t, config.GetBool("test.imported"))

		// Test invalid YAML
		invalidData := []byte(`invalid yaml {{`)
		err = config.Import(invalidData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestGetSchema(t *testing.T) {
	config := createTestConfig(t)

	// Set up various config values to test schema marshaling
	err := config.SetValue("log.level", "info")
	require.NoError(t, err)
	err = config.SetValue("provider.default", "anthropic")
	require.NoError(t, err)
	err = config.SetValue("output.color", true)
	require.NoError(t, err)

	schema, err := config.GetSchema()
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "info", schema.Log.Level)
	assert.Equal(t, "anthropic", schema.Provider.Default)
	assert.True(t, schema.Output.Color)
}

func TestConcurrentAccess(t *testing.T) {
	config := createTestConfig(t)

	// Test concurrent reads and writes
	done := make(chan bool)
	errors := make(chan error, 10)

	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func(index int) {
			key := fmt.Sprintf("concurrent.test%d", index)
			if err := config.SetValue(key, index); err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			_ = config.GetString("provider.default")
			_ = config.GetBool("output.color")
			_ = config.All()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}

	// Verify writes succeeded
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("concurrent.test%d", i)
		value := config.GetInt(key)
		assert.Equal(t, i, value)
	}
}

// Helper function to create a test configuration
func createTestConfig(t *testing.T) *Config {
	config := &Config{
		koanf:    koanf.New("."),
		watchers: []func(){},
	}

	// Set some default values
	err := config.SetValue("provider.default", "openai")
	require.NoError(t, err)

	return config
}
