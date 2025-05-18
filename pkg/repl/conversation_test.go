package repl

import (
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	// Test creating a message without attachments
	msg := NewMessage("user", "Hello, world!", nil)

	assert.Equal(t, domain.MessageRoleUser, msg.Role)
	assert.Equal(t, "Hello, world!", msg.Content)
	assert.NotZero(t, msg.Timestamp)
	assert.Empty(t, msg.Attachments)
	assert.NotNil(t, msg.Metadata)
	assert.NotEmpty(t, msg.ID)
}

func TestNewMessage_WithAttachments(t *testing.T) {
	// Test creating a message with attachments
	llmAttachments := []llm.Attachment{
		{
			Type:     "image",
			FilePath: "test.png",
			MimeType: "image/png",
			Content:  "image data",
		},
	}

	msg := NewMessage("assistant", "Here's an image", llmAttachments)

	assert.Equal(t, domain.MessageRoleAssistant, msg.Role)
	assert.Equal(t, "Here's an image", msg.Content)
	assert.Len(t, msg.Attachments, 1)

	// Check attachment conversion
	domainAtt := msg.Attachments[0]
	assert.Equal(t, domain.AttachmentTypeImage, domainAtt.Type)
	assert.Equal(t, "test.png", domainAtt.Name)
	assert.Equal(t, "image/png", domainAtt.MimeType)
	assert.Equal(t, []byte("image data"), domainAtt.Content)
}

func TestGetHistory(t *testing.T) {
	// Create a conversation with messages
	conv := domain.NewConversation("test-conv")
	conv.SystemPrompt = "You are a helpful assistant."

	// Add messages
	msg1 := NewMessage("user", "Hello", nil)
	conv.AddMessage(msg1)

	msg2 := NewMessage("assistant", "Hi there!", nil)
	conv.AddMessage(msg2)

	// Get history
	history := GetHistory(conv)

	// Should have system prompt + 2 messages
	require.Len(t, history, 3)

	// Check system prompt
	assert.Equal(t, "system", string(history[0].Role))
	assert.Equal(t, "You are a helpful assistant.", history[0].Content)

	// Check user message
	assert.Equal(t, "user", string(history[1].Role))
	assert.Equal(t, "Hello", history[1].Content)

	// Check assistant message
	assert.Equal(t, "assistant", string(history[2].Role))
	assert.Equal(t, "Hi there!", history[2].Content)
}

func TestGetHistory_NoSystemPrompt(t *testing.T) {
	// Create a conversation without system prompt
	conv := domain.NewConversation("test-conv")

	// Add messages
	msg1 := NewMessage("user", "Hello", nil)
	conv.AddMessage(msg1)

	// Get history
	history := GetHistory(conv)

	// Should have only 1 message
	require.Len(t, history, 1)
	assert.Equal(t, "user", string(history[0].Role))
}

func TestAddMessageToConversation(t *testing.T) {
	conv := domain.NewConversation("test-conv")
	originalUpdated := conv.Updated

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Add message
	AddMessageToConversation(conv, "user", "Test message", nil)

	// Check message was added
	require.Len(t, conv.Messages, 1)
	msg := conv.Messages[0]
	assert.Equal(t, domain.MessageRoleUser, msg.Role)
	assert.Equal(t, "Test message", msg.Content)
	assert.True(t, conv.Updated.After(originalUpdated))
}

func TestResetConversation(t *testing.T) {
	conv := domain.NewConversation("test-conv")

	// Add some messages
	AddMessageToConversation(conv, "user", "Message 1", nil)
	AddMessageToConversation(conv, "assistant", "Response 1", nil)

	require.Len(t, conv.Messages, 2)
	originalUpdated := conv.Updated

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Reset
	ResetConversation(conv)

	// Check messages are cleared
	assert.Empty(t, conv.Messages)
	assert.True(t, conv.Updated.After(originalUpdated))
}

func TestGetLastUserMessage(t *testing.T) {
	conv := domain.NewConversation("test-conv")

	// Add messages
	AddMessageToConversation(conv, "user", "First user message", nil)
	AddMessageToConversation(conv, "assistant", "Assistant response", nil)
	AddMessageToConversation(conv, "user", "Second user message", nil)

	// Get last user message
	lastUserMsg := GetLastUserMessage(conv)
	require.NotNil(t, lastUserMsg)
	assert.Equal(t, "Second user message", lastUserMsg.Content)
}

func TestGetLastUserMessage_NoMessages(t *testing.T) {
	conv := domain.NewConversation("test-conv")

	// No messages
	lastUserMsg := GetLastUserMessage(conv)
	assert.Nil(t, lastUserMsg)
}

func TestGetLastAssistantMessage(t *testing.T) {
	conv := domain.NewConversation("test-conv")

	// Add messages
	AddMessageToConversation(conv, "user", "User message", nil)
	AddMessageToConversation(conv, "assistant", "First assistant response", nil)
	AddMessageToConversation(conv, "user", "Another user message", nil)
	AddMessageToConversation(conv, "assistant", "Second assistant response", nil)

	// Get last assistant message
	lastAssistantMsg := GetLastAssistantMessage(conv)
	require.NotNil(t, lastAssistantMsg)
	assert.Equal(t, "Second assistant response", lastAssistantMsg.Content)
}

func TestTruncateHistory(t *testing.T) {
	conv := domain.NewConversation("test-conv")

	// Add 5 messages
	for i := 1; i <= 5; i++ {
		AddMessageToConversation(conv, "user", fmt.Sprintf("Message %d", i), nil)
	}

	require.Len(t, conv.Messages, 5)
	originalUpdated := conv.Updated

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Truncate to last 3 messages
	TruncateHistory(conv, 3)

	// Check only last 3 messages remain
	require.Len(t, conv.Messages, 3)
	assert.Equal(t, "Message 3", conv.Messages[0].Content)
	assert.Equal(t, "Message 4", conv.Messages[1].Content)
	assert.Equal(t, "Message 5", conv.Messages[2].Content)
	assert.True(t, conv.Updated.After(originalUpdated))
}

func TestTruncateHistory_NoTruncation(t *testing.T) {
	conv := domain.NewConversation("test-conv")

	// Add 3 messages
	for i := 1; i <= 3; i++ {
		AddMessageToConversation(conv, "user", fmt.Sprintf("Message %d", i), nil)
	}

	originalUpdated := conv.Updated

	// Try to truncate to 5 (no truncation should occur)
	TruncateHistory(conv, 5)

	// All messages should remain
	require.Len(t, conv.Messages, 3)
	assert.Equal(t, originalUpdated, conv.Updated)
}

// End of tests
