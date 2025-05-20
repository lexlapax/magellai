// ABOUTME: CLI integration tests for the ask command
// ABOUTME: Tests various options and behaviors of the ask command

//go:build integration
// +build integration

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI_AskBasic tests basic ask command functionality
func TestCLI_AskBasic(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Test basic ask command with a simple prompt
		output, err := env.RunCommand("ask", "Hello, what is your name?")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response should not be empty")
	})
}

// TestCLI_AskWithModel tests ask command with a specific model
func TestCLI_AskWithModel(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test ask command with a specific model
		output, err := env.RunCommand("ask", "--model", "mock/default", "Hello, what is your name?")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response should not be empty")
	})
}

// TestCLI_AskWithProfile tests ask command with a specific profile
func TestCLI_AskWithProfile(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test ask command with fallback profile
		output, err := env.RunCommand("ask", "--profile", "fallback", "Hello, what is your name?")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response should not be empty")

		// Test ask with error profile (should fallback to default provider)
		output, err = env.RunCommand("ask", "--profile", "error", "Hello, what is your name?")
		// This might fail or succeed depending on fallback behavior
		if err == nil {
			assert.NotEmpty(t, output, "Response should not be empty if fallback worked")
		}
	})
}

// TestCLI_AskWithFlags tests ask command with various flags
func TestCLI_AskWithFlags(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test ask with --json flag
		jsonOutput, err := env.RunCommand("ask", "--json", "Tell me a joke")
		require.NoError(t, err)
		assert.Contains(t, jsonOutput, "\"response\":")

		// Test ask with --raw flag (no formatting)
		rawOutput, err := env.RunCommand("ask", "--raw", "Tell me a joke")
		require.NoError(t, err)
		assert.NotEmpty(t, rawOutput, "Raw response should not be empty")
		
		// Test ask with --stream flag
		streamOutput, err := env.RunCommand("ask", "--stream", "Tell me a joke")
		require.NoError(t, err)
		assert.NotEmpty(t, streamOutput, "Streaming response should not be empty")
	})
}

// TestCLI_AskWithAttachment tests ask command with file attachments
func TestCLI_AskWithAttachment(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Create a test file to attach
		testFilePath := filepath.Join(env.TempDir, "test_attachment.txt")
		testContent := "This is a test file for attachment testing."
		err := os.WriteFile(testFilePath, []byte(testContent), 0644)
		require.NoError(t, err)

		// Test ask with file attachment
		output, err := env.RunCommand("ask", "--file", testFilePath, "Summarize the attached file")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response with attachment should not be empty")
	})
}

// TestCLI_AskWithSystemPrompt tests ask command with system prompts
func TestCLI_AskWithSystemPrompt(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test ask with system prompt
		output, err := env.RunCommand("ask", 
			"--system", "You are a helpful assistant that speaks like a pirate.",
			"Tell me about the weather")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response with system prompt should not be empty")
	})
}

// TestCLI_AskWithTemperature tests ask command with temperature setting
func TestCLI_AskWithTemperature(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test ask with temperature setting
		output, err := env.RunCommand("ask", "--temperature", "0.7", "Generate a creative story")
		require.NoError(t, err)
		assert.NotEmpty(t, output, "Response with temperature should not be empty")
	})
}

// TestCLI_AskWithPipe tests ask command with piped input
func TestCLI_AskWithPipe(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Create a command that uses pipe input instead of command-line argument
		cmd := "echo 'Summarize this text' | " + env.BinaryPath + " --config-file " + env.ConfigPath + " ask"
		output, err := runBashCommand(cmd)
		assert.NoError(t, err)
		assert.NotEmpty(t, output, "Response from piped input should not be empty")
	})
}

// runBashCommand is a helper to run bash commands for pipe testing
func runBashCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}