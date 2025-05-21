// ABOUTME: Integration tests for end-to-end session branching and merging
// ABOUTME: Tests the complete flow of session branching and merging operations

//go:build integration
// +build integration

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
	"github.com/lexlapax/magellai/pkg/storage/filesystem"
)

func TestSessionBranchingAndMerging_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "magellai-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Configure the test environment
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
log:
  level: warn

storage:
  type: filesystem
  dir: "` + filepath.Join(tempDir, "sessions") + `"

provider:
  - name: mock
    type: mock
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create storage backend
	storageDir := filepath.Join(tempDir, "sessions")
	require.NoError(t, os.MkdirAll(storageDir, 0755))

	backend, err := filesystem.New(storage.Config{
		"base_dir": storageDir,
		"user_id":  "test-user",
	})
	require.NoError(t, err)
	defer backend.Close()

	t.Run("CompleteSessionBranchingFlow", func(t *testing.T) {
		// Test the complete branching flow

		// Step 1: Create a new session
		session := backend.NewSession("parent-session")
		session.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleUser,
			Content: "Hello, test",
		})
		session.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleAssistant,
			Content: "Hello! I'm ready to assist.",
		})
		err = backend.Update(session)
		assert.NoError(t, err)

		// Step 2: Create a branch from the parent session
		branch1, err := session.CreateBranch("branch1-id", "branch1", 1)
		assert.NoError(t, err)
		assert.NotNil(t, branch1)
		assert.Equal(t, session.ID, branch1.ParentID)
		assert.Equal(t, 1, branch1.BranchPoint)
		err = backend.Update(branch1)
		assert.NoError(t, err)
		err = backend.Update(session) // Save parent to update child list
		assert.NoError(t, err)

		// Step 3: Add messages to the branch
		branch1.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleUser,
			Content: "Branch message",
		})
		branch1.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleAssistant,
			Content: "Response from branch",
		})
		err = backend.Update(branch1)
		assert.NoError(t, err)

		// Step 4: Create another branch
		branch2, err := session.CreateBranch("branch2-id", "branch2", 1)
		assert.NoError(t, err)
		err = backend.Update(branch2)
		assert.NoError(t, err)
		err = backend.Update(session) // Save parent to update child list
		assert.NoError(t, err)

		// Step 5: Merge branches
		options := domain.MergeOptions{
			Type:         domain.MergeTypeContinuation,
			SourceID:     branch2.ID,
			TargetID:     branch1.ID,
			CreateBranch: false,
		}
		result, err := backend.MergeSessions(branch1.ID, branch2.ID, options)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Step 6: Verify the merge result
		merged, err := backend.Get(branch1.ID)
		assert.NoError(t, err)
		assert.NotNil(t, merged)
		assert.Greater(t, len(merged.Conversation.Messages), 2)
	})

	t.Run("BranchTreeOperations", func(t *testing.T) {
		// Test branch tree operations

		// Create a parent session
		parent := backend.NewSession("tree-parent")
		parent.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleUser,
			Content: "Root message",
		})
		err = backend.Update(parent)
		assert.NoError(t, err)

		// Create multiple branches
		for i := 1; i <= 3; i++ {
			branchID := fmt.Sprintf("tree-branch%d-id", i)
			branchName := fmt.Sprintf("tree-branch%d", i)
			branch, err := parent.CreateBranch(branchID, branchName, 0)
			assert.NoError(t, err)
			err = backend.Update(branch)
			assert.NoError(t, err)
		}
		err = backend.Update(parent) // Update parent with children
		assert.NoError(t, err)

		// Get children
		children, err := backend.GetChildren(parent.ID)
		assert.NoError(t, err)
		assert.Len(t, children, 3)

		// Get branch tree
		tree, err := backend.GetBranchTree(parent.ID)
		assert.NoError(t, err)
		assert.NotNil(t, tree)
		assert.Equal(t, parent.ID, tree.Session.ID)
		assert.Len(t, tree.Children, 3)
	})

	t.Run("MergeConflictHandling", func(t *testing.T) {
		// Test merge conflict scenarios

		// Create parent session
		parent := backend.NewSession("conflict-parent")
		parent.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleUser,
			Content: "Initial message",
		})
		err = backend.Update(parent)
		assert.NoError(t, err)

		// Create branches with different content
		branch1, err := parent.CreateBranch("conflict-branch1-id", "conflict-branch1", 0)
		assert.NoError(t, err)
		branch1.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleUser,
			Content: "Branch 1 message",
		})
		err = backend.Update(branch1)
		assert.NoError(t, err)

		branch2, err := parent.CreateBranch("conflict-branch2-id", "conflict-branch2", 0)
		assert.NoError(t, err)
		branch2.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleUser,
			Content: "Branch 2 message",
		})
		err = backend.Update(branch2)
		assert.NoError(t, err)

		// Attempt merge with different strategies
		mergeTypes := []domain.MergeType{
			domain.MergeTypeContinuation,
			domain.MergeTypeRebase,
			domain.MergeTypeCherryPick,
		}

		for _, mergeType := range mergeTypes {
			options := domain.MergeOptions{
				Type:         mergeType,
				SourceID:     branch2.ID,
				TargetID:     branch1.ID,
				CreateBranch: true,
				BranchName:   fmt.Sprintf("merged-%d", mergeType),
			}
			result, err := backend.MergeSessions(branch1.ID, branch2.ID, options)
			// The merge may succeed or fail depending on type, but shouldn't panic
			if err == nil {
				assert.NotNil(t, result)
			}
		}
	})

	t.Run("SessionRecoveryAfterBranching", func(t *testing.T) {
		// Test session recovery with branches

		// Create a session with branches
		parent := backend.NewSession("recovery-parent")
		parent.Conversation.AddMessage(domain.Message{
			Role:    domain.MessageRoleUser,
			Content: "Parent message",
		})
		err = backend.Update(parent)
		assert.NoError(t, err)

		branch, err := parent.CreateBranch("recovery-branch-id", "recovery-branch", 0)
		assert.NoError(t, err)
		err = backend.Update(branch)
		assert.NoError(t, err)
		err = backend.Update(parent) // Update parent's child list
		assert.NoError(t, err)

		// Simulate recovery by loading the session
		recovered, err := backend.Get(parent.ID)
		assert.NoError(t, err)
		assert.NotNil(t, recovered)
		assert.Len(t, recovered.ChildIDs, 1)
		assert.Contains(t, recovered.ChildIDs, branch.ID)

		// Verify branch information is preserved
		info := recovered.ToSessionInfo()
		assert.Equal(t, 1, info.ChildCount)
		assert.False(t, info.IsBranch)

		// Load the branch
		recoveredBranch, err := backend.Get(branch.ID)
		assert.NoError(t, err)
		assert.NotNil(t, recoveredBranch)
		assert.Equal(t, parent.ID, recoveredBranch.ParentID)
		assert.True(t, recoveredBranch.IsBranch())
	})
}
