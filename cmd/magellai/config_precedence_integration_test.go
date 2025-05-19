// ABOUTME: Integration tests for configuration loading precedence
// ABOUTME: Tests the behavior of configuration loading from multiple sources

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/magellai/pkg/config"
)

func TestConfigurationPrecedence_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "magellai-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("EnvironmentOverridesFile", func(t *testing.T) {
		// Test that environment variables override file configuration
		
		// Create config file with some values
		configPath := filepath.Join(tempDir, "test-config.yaml")
		configContent := `
log:
  level: info
provider:
  default: openai
output:
  format: text
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Set environment variable to override log level
		os.Setenv("MAGELLAI_LOG_LEVEL", "debug")
		defer os.Unsetenv("MAGELLAI_LOG_LEVEL")

		// Initialize and load configuration
		err = config.Init()
		require.NoError(t, err)
		
		// Load config file
		err = config.Manager.LoadFile(configPath)
		require.NoError(t, err)
		
		// Get the loaded configuration
		schema, err := config.Manager.GetSchema()
		require.NoError(t, err)

		// Verify environment override took precedence
		assert.Equal(t, "debug", schema.Log.Level)
		// Verify file values are still used for non-overridden settings
		assert.Equal(t, "text", schema.Output.Format)
	})

	t.Run("MultipleSources", func(t *testing.T) {
		// Test loading from multiple configuration sources
		
		// Create global config
		globalDir := filepath.Join(tempDir, ".config", "magellai")
		require.NoError(t, os.MkdirAll(globalDir, 0755))
		globalConfig := filepath.Join(globalDir, "config.yaml")
		globalContent := `
provider:
  default: anthropic
log:
  level: warn
`
		err := os.WriteFile(globalConfig, []byte(globalContent), 0644)
		require.NoError(t, err)

		// Create local config
		localConfig := filepath.Join(tempDir, ".magellai.yaml")
		localContent := `
output:
  format: json
log:
  level: info
`
		err = os.WriteFile(localConfig, []byte(localContent), 0644)
		require.NoError(t, err)

		// Set up environment to use these configs
		os.Setenv("MAGELLAI_CONFIG_DIR", globalDir)
		defer os.Unsetenv("MAGELLAI_CONFIG_DIR")
		
		// Change working directory to temp dir
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		os.Chdir(tempDir)
		defer os.Chdir(originalWd)

		// Initialize and load configuration
		err = config.Init()
		require.NoError(t, err)
		
		schema, err := config.Manager.GetSchema()
		require.NoError(t, err)

		// Local config should override global for log level
		assert.Equal(t, "info", schema.Log.Level)
		// Local config provides output format
		assert.Equal(t, "json", schema.Output.Format)
		// Global config provides provider information
		assert.Equal(t, "anthropic", schema.Provider.Default)
	})

	t.Run("ProfileOverrides", func(t *testing.T) {
		// Test that profile settings override base configuration
		
		configPath := filepath.Join(tempDir, "config-with-profiles.yaml")
		configContent := `
log:
  level: info

provider:
  default: openai

profiles:
  development:
    log:
      level: debug
    provider:
      default: mock
  production:
    log:
      level: error
    output:
      format: json
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Test development profile
		os.Setenv("MAGELLAI_PROFILE", "development")
		err = config.Init()
		require.NoError(t, err)
		err = config.Manager.LoadFile(configPath)
		require.NoError(t, err)
		err = config.Manager.SetProfile("development")
		require.NoError(t, err)
		
		devSchema, err := config.Manager.GetSchema()
		require.NoError(t, err)
		assert.Equal(t, "debug", devSchema.Log.Level)
		assert.Equal(t, "mock", devSchema.Provider.Default)

		// Test production profile  
		os.Setenv("MAGELLAI_PROFILE", "production")
		err = config.Init()
		require.NoError(t, err)
		err = config.Manager.LoadFile(configPath)
		require.NoError(t, err)  
		err = config.Manager.SetProfile("production")
		require.NoError(t, err)
		
		prodSchema, err := config.Manager.GetSchema()
		require.NoError(t, err)
		assert.Equal(t, "error", prodSchema.Log.Level)
		assert.Equal(t, "json", prodSchema.Output.Format)
		
		os.Unsetenv("MAGELLAI_PROFILE")
	})

	t.Run("ConfigMerging", func(t *testing.T) {
		// Test complex configuration merging from multiple sources
		
		// Create base config
		baseConfig := filepath.Join(tempDir, "base.yaml")
		baseContent := `
log:
  level: info
  format: text

provider:
  default: openai

output:
  format: text
  color: true

session:
  storage:
    type: filesystem
    path: ${HOME}/.magellai/sessions
`
		err := os.WriteFile(baseConfig, []byte(baseContent), 0644)
		require.NoError(t, err)

		// Set environment variables
		os.Setenv("MAGELLAI_OUTPUT_COLOR", "false")
		os.Setenv("MAGELLAI_SESSION_STORAGE_TYPE", "sqlite")
		defer func() {
			os.Unsetenv("MAGELLAI_OUTPUT_COLOR")
			os.Unsetenv("MAGELLAI_SESSION_STORAGE_TYPE")
		}()

		// Initialize and load configuration
		err = config.Init()
		require.NoError(t, err)
		err = config.Manager.LoadFile(baseConfig)
		require.NoError(t, err)
		
		schema, err := config.Manager.GetSchema()
		require.NoError(t, err)

		// Environment overrides specific values
		assert.Equal(t, false, schema.Output.Color)
		// Check the storage type
		assert.Equal(t, "sqlite", schema.Session.Storage.Type)
		// File values remain for non-overridden
		assert.Equal(t, "info", schema.Log.Level)
		assert.Equal(t, "text", schema.Output.Format)
	})
}