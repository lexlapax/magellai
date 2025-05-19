// ABOUTME: Tests for REPL branch commands including merge functionality
// ABOUTME: Validates command parsing and execution for branch operations

package repl

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/repl/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdMerge(t *testing.T) {
	// Create test sessions
	backend := session.NewMockStorageBackend()
	storageManager, err := session.NewStorageManager(backend)
	require.NoError(t, err)

	manager := &session.SessionManager{StorageManager: storageManager}

	// Create target session
	targetSession, err := manager.NewSession("target")
	require.NoError(t, err)
	AddMessageToConversation(targetSession.Conversation, "user", "Target message", nil)
	err = manager.SaveSession(targetSession)
	require.NoError(t, err)

	// Create source session
	sourceSession, err := manager.NewSession("source")
	require.NoError(t, err)
	AddMessageToConversation(sourceSession.Conversation, "user", "Source message", nil)
	err = manager.SaveSession(sourceSession)
	require.NoError(t, err)

	// Create REPL with target session
	output := new(bytes.Buffer)
	input := strings.NewReader("")

	r := &REPL{
		session: targetSession,
		manager: manager,
		writer:  output,
		reader:  bufio.NewReader(input),
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectedMsg string
		checkResult func(t *testing.T)
	}{
		{
			name:        "No arguments",
			args:        []string{},
			expectError: true,
			expectedMsg: "usage: /merge",
		},
		{
			name:        "Basic merge",
			args:        []string{sourceSession.ID},
			expectError: false,
			expectedMsg: "Successfully merged",
			checkResult: func(t *testing.T) {
				// Verify messages were merged
				updatedTarget, err := manager.StorageManager.LoadSession(targetSession.ID)
				require.NoError(t, err)
				assert.Equal(t, 2, len(updatedTarget.Conversation.Messages))
			},
		},
		{
			name:        "Merge with branch",
			args:        []string{sourceSession.ID, "--create-branch", "--branch-name", "Test Merge"},
			expectError: false,
			expectedMsg: "Created new branch",
			checkResult: func(t *testing.T) {
				// Verify a merge was executed
				assert.Equal(t, 1, backend.GetCallCount("MergeSessions"))
			},
		},
		{
			name:        "Invalid merge type",
			args:        []string{sourceSession.ID, "--type", "invalid"},
			expectError: true,
			expectedMsg: "invalid merge type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			backend.ClearCalls()

			err := r.cmdMerge(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectedMsg != "" {
					assert.Contains(t, output.String(), tt.expectedMsg)
				}
				if tt.checkResult != nil {
					tt.checkResult(t)
				}
			}
		})
	}
}

func TestMergeCommandHelp(t *testing.T) {
	// Create REPL instance
	backend := session.NewMockStorageBackend()
	storageManager, err := session.NewStorageManager(backend)
	require.NoError(t, err)

	manager := &session.SessionManager{StorageManager: storageManager}
	targetSession, err := manager.NewSession("test")
	require.NoError(t, err)

	output := new(bytes.Buffer)
	input := strings.NewReader("")

	r := &REPL{
		session: targetSession,
		manager: manager,
		writer:  output,
		reader:  bufio.NewReader(input),
	}

	// Test help text includes merge command
	err = r.showHelp()
	require.NoError(t, err)

	helpText := output.String()
	assert.Contains(t, helpText, "/merge")
	assert.Contains(t, helpText, "Merge another session")
}
