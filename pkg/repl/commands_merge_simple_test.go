// ABOUTME: Simple test for session merge functionality
// ABOUTME: Basic validation of merge operations without full REPL setup

package repl

import (
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleMerge(t *testing.T) {
	// Create a mock backend
	backend := NewMockStorageBackend()

	// Create two test sessions
	targetSession := backend.NewSession("target")
	AddMessageToConversation(targetSession.Conversation, "user", "Target message", nil)
	err := backend.SaveSession(targetSession)
	require.NoError(t, err)

	sourceSession := backend.NewSession("source")
	AddMessageToConversation(sourceSession.Conversation, "user", "Source message", nil)
	err = backend.SaveSession(sourceSession)
	require.NoError(t, err)

	// Debug print IDs
	t.Logf("Target ID: %s", targetSession.ID)
	t.Logf("Source ID: %s", sourceSession.ID)

	// Execute merge
	options := domain.MergeOptions{
		Type:         domain.MergeTypeContinuation,
		SourceID:     sourceSession.ID,
		TargetID:     targetSession.ID,
		CreateBranch: false,
	}

	result, err := backend.MergeSessions(targetSession.ID, sourceSession.ID, options)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.MergedCount)

	// Load target session to verify merge
	updatedTarget, err := backend.LoadSession(targetSession.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(updatedTarget.Conversation.Messages))
	assert.Equal(t, "Target message", updatedTarget.Conversation.Messages[0].Content)
	assert.Equal(t, "Source message", updatedTarget.Conversation.Messages[1].Content)
}
