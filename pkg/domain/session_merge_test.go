// ABOUTME: Tests for session merging functionality
// ABOUTME: Validates merge operations across different merge types

package domain_test

import (
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanMerge(t *testing.T) {
	tests := []struct {
		name      string
		session1  *domain.Session
		session2  *domain.Session
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Cannot merge session with itself",
			session1:  &domain.Session{ID: "session1"},
			session2:  &domain.Session{ID: "session1"},
			wantError: true,
			errorMsg:  "cannot merge session with itself",
		},
		{
			name:      "Can merge different sessions",
			session1:  createTestSession("session1"),
			session2:  createTestSession("session2"),
			wantError: false,
		},
		{
			name:      "Can merge child with parent",
			session1:  createTestSession("parent"),
			session2:  &domain.Session{ID: "child", ParentID: "parent", Conversation: &domain.Conversation{}},
			wantError: false,
		},
		{
			name:      "Sessions must have conversations",
			session1:  &domain.Session{ID: "session1"},
			session2:  &domain.Session{ID: "session2"},
			wantError: true,
			errorMsg:  "both sessions must have conversations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session1.CanMerge(tt.session2)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExecuteMergeContinuation(t *testing.T) {
	// Create base session with some messages
	target := createTestSession("target")
	addTestMessage(target, "user", "Hello")
	addTestMessage(target, "assistant", "Hi there")

	// Create source session with more messages
	source := createTestSession("source")
	addTestMessage(source, "user", "How are you?")
	addTestMessage(source, "assistant", "I'm doing well")

	options := domain.MergeOptions{
		Type:         domain.MergeTypeContinuation,
		SourceID:     source.ID,
		TargetID:     target.ID,
		CreateBranch: false,
	}

	mergedSession, result, err := target.ExecuteMerge(source, options)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, mergedSession)

	// Check that all messages are merged
	assert.Equal(t, 4, len(mergedSession.Conversation.Messages))
	assert.Equal(t, 2, result.MergedCount)

	// Verify message order
	assert.Equal(t, "Hello", mergedSession.Conversation.Messages[0].Content)
	assert.Equal(t, "Hi there", mergedSession.Conversation.Messages[1].Content)
	assert.Equal(t, "How are you?", mergedSession.Conversation.Messages[2].Content)
	assert.Equal(t, "I'm doing well", mergedSession.Conversation.Messages[3].Content)
}

func TestExecuteMergeRebase(t *testing.T) {
	// Create base session
	target := createTestSession("target")
	addTestMessage(target, "user", "Original message 1")
	addTestMessage(target, "assistant", "Original response 1")
	addTestMessage(target, "user", "Original message 2")

	// Create source session
	source := createTestSession("source")
	addTestMessage(source, "user", "Source message 1")
	addTestMessage(source, "assistant", "Source response 1")

	options := domain.MergeOptions{
		Type:       domain.MergeTypeRebase,
		SourceID:   source.ID,
		TargetID:   target.ID,
		MergePoint: 2, // Keep first two messages from target
	}

	mergedSession, result, err := target.ExecuteMerge(source, options)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, mergedSession)

	// Check that messages are properly rebased
	assert.Equal(t, 4, len(mergedSession.Conversation.Messages))
	assert.Equal(t, 2, result.MergedCount)

	// Verify rebase - first two from target, then all from source
	assert.Equal(t, "Original message 1", mergedSession.Conversation.Messages[0].Content)
	assert.Equal(t, "Original response 1", mergedSession.Conversation.Messages[1].Content)
	assert.Equal(t, "Source message 1", mergedSession.Conversation.Messages[2].Content)
	assert.Equal(t, "Source response 1", mergedSession.Conversation.Messages[3].Content)
}

func TestExecuteMergeWithBranch(t *testing.T) {
	// Create sessions
	target := createTestSession("target")
	addTestMessage(target, "user", "Base message")
	
	source := createTestSession("source")
	addTestMessage(source, "user", "Branch message")

	options := domain.MergeOptions{
		Type:         domain.MergeTypeContinuation,
		SourceID:     source.ID,
		TargetID:     target.ID,
		CreateBranch: true,
		BranchName:   "Merged branch",
	}

	mergedSession, result, err := target.ExecuteMerge(source, options)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, mergedSession)

	// Verify it created a new branch
	assert.NotEmpty(t, result.NewBranchID)
	assert.Equal(t, result.NewBranchID, mergedSession.ID)
	assert.Equal(t, "Merged branch", mergedSession.Name)
	assert.Equal(t, target.ID, mergedSession.ParentID)

	// Check that target now has this as a child
	assert.Contains(t, target.ChildIDs, mergedSession.ID)

	// Verify merge metadata
	assert.NotNil(t, mergedSession.Metadata)
	assert.Equal(t, source.ID, mergedSession.Metadata["merge_source"])
	assert.Equal(t, target.ID, mergedSession.Metadata["merge_target"])
	assert.Equal(t, domain.MergeTypeContinuation, mergedSession.Metadata["merge_type"])
}

func TestIsAncestorOf(t *testing.T) {
	parent := createTestSession("parent")
	child := &domain.Session{
		ID:       "child",
		ParentID: "parent",
	}
	unrelated := createTestSession("unrelated")

	assert.True(t, parent.IsAncestorOf(child))
	assert.False(t, child.IsAncestorOf(parent))
	assert.False(t, parent.IsAncestorOf(unrelated))
	assert.False(t, unrelated.IsAncestorOf(parent))
}

// Helper functions

func createTestSession(id string) *domain.Session {
	return &domain.Session{
		ID:           id,
		Name:         "Test " + id,
		Created:      time.Now(),
		Updated:      time.Now(),
		Conversation: &domain.Conversation{ID: id, Messages: []domain.Message{}},
		ChildIDs:     []string{},
		Metadata:     make(map[string]interface{}),
	}
}

func addTestMessage(session *domain.Session, role, content string) {
	msg := domain.Message{
		ID:        generateTestMessageID(),
		Role:      domain.MessageRole(role),
		Content:   content,
		Timestamp: time.Now(),
	}
	session.Conversation.Messages = append(session.Conversation.Messages, msg)
}

func generateTestMessageID() string {
	return "msg_" + time.Now().Format("20060102150405.000000")
}