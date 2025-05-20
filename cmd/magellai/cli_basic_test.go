// ABOUTME: Basic CLI integration tests for core functionality
// ABOUTME: Tests basic commands like version, help, and config

//go:build integration
// +build integration

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI_Version tests the version command
func TestCLI_Version(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Test with default arguments
		output, err := env.RunCommand("version")
		require.NoError(t, err)
		assert.Contains(t, output, "magellai version")
		
		// Test with --json flag
		outputJSON, err := env.RunCommand("version", "--json")
		require.NoError(t, err)
		assert.Contains(t, outputJSON, "\"version\":")
		assert.Contains(t, outputJSON, "\"build_date\":")
	})
}

// TestCLI_Help tests the help command
func TestCLI_Help(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Test global help
		output, err := env.RunCommand("--help")
		require.NoError(t, err)
		assert.Contains(t, output, "Usage:")
		assert.Contains(t, output, "Commands:")
		assert.Contains(t, output, "ask")
		assert.Contains(t, output, "chat")
		assert.Contains(t, output, "config")
		
		// Test command-specific help
		askHelp, err := env.RunCommand("ask", "--help")
		require.NoError(t, err)
		assert.Contains(t, askHelp, "Usage:")
		assert.Contains(t, askHelp, "ask [prompt]")
		
		// Test help command explicitly
		helpOutput, err := env.RunCommand("help")
		require.NoError(t, err)
		assert.Contains(t, helpOutput, "Usage:")
		assert.Contains(t, helpOutput, "Commands:")
	})
}

// TestCLI_Config tests the config command
func TestCLI_Config(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Test config show
		output, err := env.RunCommand("config", "show")
		require.NoError(t, err)
		assert.Contains(t, output, "Configuration:")
		assert.Contains(t, output, "provider:")
		
		// Test config show with --json flag
		jsonOutput, err := env.RunCommand("config", "show", "--json")
		require.NoError(t, err)
		assert.Contains(t, jsonOutput, "\"provider\":")
		
		// Test config generate
		configGenPath := filepath.Join(env.TempDir, "generated_config.yaml")
		_, err = env.RunCommand("config", "generate", configGenPath)
		require.NoError(t, err)
		
		// Verify the generated config exists
		_, err = os.Stat(configGenPath)
		assert.NoError(t, err, "Generated config file should exist")
		
		// Test with nonexistent config file
		_, err = env.RunCommand("--config", "/nonexistent/config.yaml", "version")
		assert.Error(t, err, "Should error with nonexistent config")
	})
}

// TestCLI_InvalidFlags tests behavior with invalid flags
func TestCLI_InvalidFlags(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test with invalid flag
		_, err := env.RunCommand("--nonexistent-flag")
		assert.Error(t, err, "Should error with nonexistent flag")
		
		// Test with invalid command
		_, err = env.RunCommand("nonexistent-command")
		assert.Error(t, err, "Should error with nonexistent command")
		
		// Test with missing required arguments
		_, err = env.RunCommand("ask")
		assert.Error(t, err, "Should error with missing required arguments")
	})
}

// TestCLI_Models tests the model command
func TestCLI_Models(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test models list command
		output, err := env.RunCommand("model", "list")
		require.NoError(t, err)
		assert.Contains(t, output, "Available models:")
		
		// Test models list with --json flag
		jsonOutput, err := env.RunCommand("model", "list", "--json")
		require.NoError(t, err)
		assert.Contains(t, jsonOutput, "[")
		assert.Contains(t, jsonOutput, "]")
	})
}

// TestCLI_AliasCommands tests that command aliases work
func TestCLI_AliasCommands(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// The config has aliases: h=help, v=version
		
		// Test 'v' alias for version
		output, err := env.RunCommand("v")
		require.NoError(t, err)
		assert.Contains(t, output, "magellai version")
		
		// Test 'h' alias for help
		helpOutput, err := env.RunCommand("h")
		require.NoError(t, err)
		assert.Contains(t, helpOutput, "Usage:")
		assert.Contains(t, helpOutput, "Commands:")
	})
}