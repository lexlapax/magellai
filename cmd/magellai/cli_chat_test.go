// ABOUTME: CLI integration tests for the chat command and REPL functionality
// ABOUTME: Tests interactive chat features, commands, and session management

//go:build cmdline
// +build cmdline

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

// TestCLI_ChatBasic tests basic chat mode functionality
func TestCLI_ChatBasic(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Test basic chat interaction with a simple message and exit
		input := "Hello, what is your name?\n/exit"
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		assert.Contains(t, output, "Starting new chat")
		// Should contain the prompt and some response
		assert.Contains(t, output, "Hello, what is your name?")
	})
}

// TestCLI_ChatCommands tests REPL commands in chat mode
func TestCLI_ChatCommands(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test various REPL commands
		input := `/help
/version
/model list
/config show
/exit`
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		assert.Contains(t, output, "Available commands")
		assert.Contains(t, output, "magellai version")
		assert.Contains(t, output, "Available models")
		assert.Contains(t, output, "Configuration")
	})
}

// TestCLI_ChatWithSessionManagement tests session commands in chat mode
func TestCLI_ChatWithSessionManagement(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test session management commands
		input := `Hello
/session list
/session save test-session
/session info
/exit`
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		assert.Contains(t, output, "Sessions:")
		assert.Contains(t, output, "Saved session")
		assert.Contains(t, output, "Session info")

		// Now try to load the saved session
		loadInput := `/session load test-session
Hello again
/exit`
		loadOutput, err := env.RunInteractiveCommand(loadInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, loadOutput, "Loaded session")
		assert.Contains(t, loadOutput, "Hello") // Previous message from history
	})
}

// TestCLI_ChatWithModel tests chat with specific models
func TestCLI_ChatWithModel(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test chat with specific model
		input := `Hello
/model set mock/default
Tell me more
/exit`
		output, err := env.RunInteractiveCommand(input, "chat", "--model", "mock/default")
		require.NoError(t, err)
		assert.Contains(t, output, "Model changed")
	})
}

// TestCLI_ChatSessionAttachments tests file attachments in chat mode
func TestCLI_ChatSessionAttachments(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Create a test file to attach
		testFilePath := filepath.Join(env.TempDir, "chat_attachment.txt")
		testContent := "This is a test file for chat attachment testing."
		err := os.WriteFile(testFilePath, []byte(testContent), 0644)
		require.NoError(t, err)

		// Test chat with file attachment
		input := fmt.Sprintf(`Hello
/attach %s
Summarize the attached file
/attachments list
/exit`, testFilePath)
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		assert.Contains(t, output, "Attached file")
		assert.Contains(t, output, "Attachments:")
	})
}

// TestCLI_ChatSessionBranching tests session branching in chat mode
func TestCLI_ChatSessionBranching(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test session branching
		input := `Hello
This is the main branch
/session save main-session
/branch create test-branch
This is in the branch
/session info
/branch list
/exit`
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		assert.Contains(t, output, "Saved session")
		assert.Contains(t, output, "Created branch")
		assert.Contains(t, output, "Branch list")

		// Now switch back to main and verify
		switchInput := `/session load main-session
/session info
/exit`
		switchOutput, err := env.RunInteractiveCommand(switchInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, switchOutput, "Loaded session")
		assert.Contains(t, switchOutput, "main-session")
	})
}

// TestCLI_ChatSessionMerging tests session merging in chat mode
func TestCLI_ChatSessionMerging(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Create two branches and then merge them
		setupInput := `First message
/session save source-session
/exit`
		_, err := env.RunInteractiveCommand(setupInput, "chat")
		require.NoError(t, err)

		// Create target session
		targetInput := `Target session first message
/session save target-session
/exit`
		_, err = env.RunInteractiveCommand(targetInput, "chat")
		require.NoError(t, err)

		// Now merge source into target
		mergeInput := `/session load target-session
/merge source-session
/history
/exit`
		mergeOutput, err := env.RunInteractiveCommand(mergeInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, mergeOutput, "Merged session")
		// Should show messages from both sessions in history
		assert.Contains(t, mergeOutput, "Target session")
		assert.Contains(t, mergeOutput, "First message")
	})
}

// TestCLI_ChatNonInteractiveMode tests chat in non-interactive mode
func TestCLI_ChatNonInteractiveMode(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Create a script file to run commands
		scriptPath := filepath.Join(env.TempDir, "chat_script.txt")
		scriptContent := `Hello
/version
/exit`
		err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
		require.NoError(t, err)

		// Run chat with input from script file
		cmd := fmt.Sprintf("cat %s | %s --config-file %s chat",
			scriptPath, env.BinaryPath, env.ConfigPath)
		output, err := runChatBashCommand(cmd)
		assert.NoError(t, err)
		assert.Contains(t, output, "Hello")
		assert.Contains(t, output, "magellai version")
	})
}

// TestCLI_ChatHistory tests history navigation in chat mode
func TestCLI_ChatHistory(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Test history command and navigation
		input := `First message
Second message
/history
/history 1
/exit`
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		assert.Contains(t, output, "Chat history")
		assert.Contains(t, output, "First message")
		assert.Contains(t, output, "Second message")
	})
}

// TestCLI_ChatSearch tests search functionality in chat mode
func TestCLI_ChatSearch(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Create a session with searchable content
		setupInput := `This is a unique phrase to search for
Another message with different content
/session save search-test-session
/exit`
		_, err := env.RunInteractiveCommand(setupInput, "chat")
		require.NoError(t, err)

		// Now search for the content
		searchInput := `/search "unique phrase"
/exit`
		searchOutput, err := env.RunInteractiveCommand(searchInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, searchOutput, "Search results")
		assert.Contains(t, searchOutput, "search-test-session")
	})
}

// runChatBashCommand is a helper to run bash commands for chat tests
func runChatBashCommand(command string) (string, error) {
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
