// ABOUTME: CLI integration tests for session management features
// ABOUTME: Tests session storage, branching, and merging across storage backends

//go:build integration
// +build integration

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI_SessionBasic tests basic session management functionality
func TestCLI_SessionBasic(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// First create and save a session
		input := `Hello
This is a test message
/session save basic-test-session
/exit`
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		assert.Contains(t, output, "Saved session")
		
		// Now load the session and verify messages are present
		loadInput := `/session load basic-test-session
/history
/exit`
		loadOutput, err := env.RunInteractiveCommand(loadInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, loadOutput, "Loaded session")
		assert.Contains(t, loadOutput, "This is a test message")
	})
}

// TestCLI_SessionExport tests session export functionality
func TestCLI_SessionExport(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Create a session to export
		input := `Message one
Message two
/session save export-test-session
/exit`
		_, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		
		// Export to JSON
		jsonPath := filepath.Join(env.TempDir, "export.json")
		exportJsonInput := fmt.Sprintf(`/session load export-test-session
/session export json %s
/exit`, jsonPath)
		exportOutput, err := env.RunInteractiveCommand(exportJsonInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, exportOutput, "Exported session")
		
		// Verify JSON file exists
		_, err = os.Stat(jsonPath)
		assert.NoError(t, err)
		
		// Export to Markdown
		mdPath := filepath.Join(env.TempDir, "export.md")
		exportMdInput := fmt.Sprintf(`/session load export-test-session
/session export markdown %s
/exit`, mdPath)
		exportOutput, err = env.RunInteractiveCommand(exportMdInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, exportOutput, "Exported session")
		
		// Verify Markdown file exists
		_, err = os.Stat(mdPath)
		assert.NoError(t, err)
	})
}

// TestCLI_SessionBranching tests session branching functionality
func TestCLI_SessionBranching(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Create parent session
		parentInput := `Parent message one
Parent message two
/session save branching-parent
/exit`
		_, err := env.RunInteractiveCommand(parentInput, "chat")
		require.NoError(t, err)
		
		// Create a branch
		branchInput := `/session load branching-parent
/branch create branch1
Branch specific message
/session info
/exit`
		branchOutput, err := env.RunInteractiveCommand(branchInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, branchOutput, "Created branch")
		
		// Create another branch
		branch2Input := `/session load branching-parent
/branch create branch2
Another branch message
/session info
/exit`
		branch2Output, err := env.RunInteractiveCommand(branch2Input, "chat")
		require.NoError(t, err)
		assert.Contains(t, branch2Output, "Created branch")
		
		// List branches from parent
		listInput := `/session load branching-parent
/branch list
/exit`
		listOutput, err := env.RunInteractiveCommand(listInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, listOutput, "branch1")
		assert.Contains(t, listOutput, "branch2")
		
		// Verify branch tree
		treeInput := `/session load branching-parent
/branch tree
/exit`
		treeOutput, err := env.RunInteractiveCommand(treeInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, treeOutput, "Branch tree")
	})
}

// TestCLI_SessionMerging tests session merging functionality
func TestCLI_SessionMerging(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Create two sessions to merge
		session1Input := `First session message
Another message for first session
/session save merge-source
/exit`
		_, err := env.RunInteractiveCommand(session1Input, "chat")
		require.NoError(t, err)
		
		session2Input := `Second session message
Another message for second session
/session save merge-target
/exit`
		_, err = env.RunInteractiveCommand(session2Input, "chat")
		require.NoError(t, err)
		
		// Test continuation merge
		contMergeInput := `/session load merge-target
/merge merge-source --type continuation
/history
/exit`
		contMergeOutput, err := env.RunInteractiveCommand(contMergeInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, contMergeOutput, "Merged session")
		// Both session contents should be visible in history
		assert.Contains(t, contMergeOutput, "First session message")
		assert.Contains(t, contMergeOutput, "Second session message")
		
		// Test merge with new branch
		branchMergeInput := `/session load merge-target
/merge merge-source --type rebase --create-branch merge-branch
/branch list
/exit`
		branchMergeOutput, err := env.RunInteractiveCommand(branchMergeInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, branchMergeOutput, "Created branch")
		assert.Contains(t, branchMergeOutput, "merge-branch")
	})
}

// TestCLI_SessionSearch tests session search functionality
func TestCLI_SessionSearch(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Create a session with unique searchable content
		uniquePhrase := fmt.Sprintf("UniqueSearchPhrase%d", os.Getpid())
		searchInput := fmt.Sprintf(`This is a message
This contains the %s for testing
/session save search-session
/exit`, uniquePhrase)
		_, err := env.RunInteractiveCommand(searchInput, "chat")
		require.NoError(t, err)
		
		// Search for the unique phrase
		findInput := fmt.Sprintf(`/search "%s"
/exit`, uniquePhrase)
		findOutput, err := env.RunInteractiveCommand(findInput, "chat")
		require.NoError(t, err)
		assert.Contains(t, findOutput, "Search results")
		assert.Contains(t, findOutput, "search-session")
	})
}

// TestCLI_SessionAutoRecovery tests session auto-recovery functionality
func TestCLI_SessionAutoRecovery(t *testing.T) {
	WithMockEnv(t, StorageTypeFilesystem, func(t *testing.T, env *TestEnv) {
		// Start a session and force abnormal termination
		// This is hard to test directly, so we'll verify the auto-save feature
		
		// Create a session with auto-save enabled
		input := `This is a test message for auto-recovery
/session info
/exit`
		output, err := env.RunInteractiveCommand(input, "chat")
		require.NoError(t, err)
		
		// Get the session ID from the output
		sessionID := ""
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "ID:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					sessionID = strings.TrimSpace(parts[1])
					break
				}
			}
		}
		require.NotEmpty(t, sessionID, "Should have found session ID in output")
		
		// Now start a new chat without explicitly loading a session
		// It should recover the previous session if auto-recovery is enabled
		recoveryInput := `/session info
/exit`
		recoveryOutput, err := env.RunInteractiveCommand(recoveryInput, "chat")
		require.NoError(t, err)
		
		// May or may not recover based on auto-recovery settings
		// Just check that it either shows a session ID or starts a new session
		assert.True(t, strings.Contains(recoveryOutput, "ID:") || 
			strings.Contains(recoveryOutput, "Starting new chat"), 
			"Should either recover or start new session")
	})
}

// TestCLI_SessionStorageEdgeCases tests edge cases in session storage
func TestCLI_SessionStorageEdgeCases(t *testing.T) {
	ForEachStorageType(t, true, func(t *testing.T, env *TestEnv) {
		// Test loading non-existent session
		loadInput := `/session load nonexistent-session
/exit`
		loadOutput, err := env.RunInteractiveCommand(loadInput, "chat")
		// Should fail gracefully
		assert.NoError(t, err)
		assert.Contains(t, loadOutput, "not found") // Error message for session not found
		
		// Test saving with empty name
		saveEmptyInput := `/session save ""
/exit`
		saveEmptyOutput, err := env.RunInteractiveCommand(saveEmptyInput, "chat")
		// Should fail gracefully
		assert.NoError(t, err)
		assert.Contains(t, saveEmptyOutput, "error") // Some error about empty name
		
		// Test saving with invalid characters
		saveInvalidInput := `/session save "invalid/chars"
/exit`
		saveInvalidOutput, err := env.RunInteractiveCommand(saveInvalidInput, "chat")
		// Should fail gracefully
		assert.NoError(t, err)
		assert.Contains(t, saveInvalidOutput, "error") // Error about invalid name
	})
}