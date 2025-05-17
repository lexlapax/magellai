package repl

import (
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConversation(t *testing.T) {
	conv := NewConversation("test-123")

	assert.Equal(t, "test-123", conv.ID)
	assert.Empty(t, conv.Messages)
	assert.NotZero(t, conv.Created)
	assert.Equal(t, conv.Created, conv.Updated)
	assert.NotNil(t, conv.Metadata)
}

func TestConversation_AddMessage(t *testing.T) {
	conv := NewConversation("test-123")
	originalUpdated := conv.Updated

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	conv.AddMessage("user", "Hello, world!", nil)

	require.Len(t, conv.Messages, 1)
	msg := conv.Messages[0]

	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Hello, world!", msg.Content)
	assert.NotZero(t, msg.Timestamp)
	assert.True(t, conv.Updated.After(originalUpdated))
}

func TestConversation_AddMessageWithAttachments(t *testing.T) {
	conv := NewConversation("test-123")

	attachments := []llm.Attachment{
		{Type: llm.AttachmentTypeText, Content: "dGVzdA==", MimeType: "text/plain"},
		{Type: llm.AttachmentTypeImage, Content: "aW1hZ2U=", MimeType: "image/jpeg"},
	}

	conv.AddMessage("user", "Check these files", attachments)

	require.Len(t, conv.Messages, 1)
	msg := conv.Messages[0]

	assert.Equal(t, "user", msg.Role)
	assert.Equal(t, "Check these files", msg.Content)
	assert.Equal(t, attachments, msg.Attachments)
}

func TestConversation_SetSystemPrompt(t *testing.T) {
	conv := NewConversation("test-123")
	originalUpdated := conv.Updated

	time.Sleep(10 * time.Millisecond)

	conv.SetSystemPrompt("You are a helpful assistant.")

	assert.Equal(t, "You are a helpful assistant.", conv.SystemPrompt)
	assert.True(t, conv.Updated.After(originalUpdated))
}

func TestConversation_GetHistory(t *testing.T) {
	conv := NewConversation("test-123")

	// Set system prompt
	conv.SetSystemPrompt("You are a helpful assistant.")

	// Add messages
	conv.AddMessage("user", "Hello", nil)
	conv.AddMessage("assistant", "Hi there!", nil)

	// Add message with attachments
	attachments := []llm.Attachment{
		{Type: llm.AttachmentTypeImage, Content: "aW1hZ2U=", MimeType: "image/jpeg"},
	}
	conv.AddMessage("user", "What's in this image?", attachments)

	history := conv.GetHistory()

	require.Len(t, history, 4)

	// Check system message
	assert.Equal(t, "system", history[0].Role)
	assert.Equal(t, "You are a helpful assistant.", history[0].Content)

	// Check user message
	assert.Equal(t, "user", history[1].Role)
	assert.Equal(t, "Hello", history[1].Content)

	// Check assistant message
	assert.Equal(t, "assistant", history[2].Role)
	assert.Equal(t, "Hi there!", history[2].Content)

	// Check message with attachments
	assert.Equal(t, "user", history[3].Role)
	assert.Equal(t, "What's in this image?", history[3].Content)
	assert.Equal(t, attachments, history[3].Attachments)
}

func TestConversation_GetHistoryNoSystemPrompt(t *testing.T) {
	conv := NewConversation("test-123")

	conv.AddMessage("user", "Hello", nil)
	conv.AddMessage("assistant", "Hi!", nil)

	history := conv.GetHistory()

	require.Len(t, history, 2)
	assert.Equal(t, "user", history[0].Role)
	assert.Equal(t, "assistant", history[1].Role)
}

func TestConversation_Reset(t *testing.T) {
	conv := NewConversation("test-123")

	// Add some messages
	conv.SetSystemPrompt("System prompt")
	conv.AddMessage("user", "Hello", nil)
	conv.AddMessage("assistant", "Hi!", nil)

	// Reset conversation
	conv.Reset()

	assert.Empty(t, conv.Messages)
	assert.Equal(t, "System prompt", conv.SystemPrompt) // System prompt preserved
	assert.Equal(t, "test-123", conv.ID)                // ID preserved
}

func TestConversation_CountTokens(t *testing.T) {
	conv := NewConversation("test-123")

	// Set system prompt
	conv.SetSystemPrompt("You are a helpful assistant.")

	// Add messages
	conv.AddMessage("user", "Hello, how are you?", nil)
	conv.AddMessage("assistant", "I'm doing well, thank you!", nil)

	tokens := conv.CountTokens()

	// With our simple estimation, this should be around 20 tokens total
	assert.Greater(t, tokens, 10)
	assert.Less(t, tokens, 50)
}

func TestConversation_CountTokensWithAttachments(t *testing.T) {
	conv := NewConversation("test-123")

	attachments := []llm.Attachment{
		{Type: llm.AttachmentTypeImage, Content: "aW1hZ2U=", MimeType: "image/jpeg"},
		{Type: llm.AttachmentTypeText, Content: "dGV4dA==", MimeType: "text/plain"},
	}

	conv.AddMessage("user", "Check these files", attachments)

	tokens := conv.CountTokens()

	// Should include base text tokens plus attachment estimates
	assert.Greater(t, tokens, 200) // 2 attachments * 100 tokens each
}

func TestConversation_TrimToMaxTokens(t *testing.T) {
	conv := NewConversation("test-123")

	// Add several messages
	conv.AddMessage("user", "First message", nil)
	conv.AddMessage("assistant", "First response", nil)
	conv.AddMessage("user", "Second message", nil)
	conv.AddMessage("assistant", "Second response", nil)
	conv.AddMessage("user", "Third message", nil)

	originalCount := len(conv.Messages)

	// Trim to a small token limit
	conv.TrimToMaxTokens(20)

	// Should have fewer messages (or same if already under limit)
	assert.LessOrEqual(t, len(conv.Messages), originalCount)
	assert.Greater(t, len(conv.Messages), 0)

	// Last message should be preserved
	lastMsg := conv.Messages[len(conv.Messages)-1]
	assert.Equal(t, "Third message", lastMsg.Content)
}

func TestConversation_TrimToMaxTokensPreservesLast(t *testing.T) {
	conv := NewConversation("test-123")

	// Add just one message
	conv.AddMessage("user", "Only message", nil)

	// Try to trim to very small limit
	conv.TrimToMaxTokens(1)

	// Should still have the one message
	assert.Len(t, conv.Messages, 1)
	assert.Equal(t, "Only message", conv.Messages[0].Content)
}
