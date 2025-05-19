// ABOUTME: Conversation helper functions for REPL operations using domain types
// ABOUTME: Provides REPL-specific conversation utilities without type conversion

package repl

import (
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/magellai/pkg/domain"
)

// NewMessage creates a new domain message with the given parameters
func NewMessage(role, content string, attachments []domain.Attachment) domain.Message {
	msg := domain.Message{
		ID:          uuid.New().String(),
		Role:        domain.MessageRole(role),
		Content:     content,
		Timestamp:   time.Now(),
		Attachments: attachments,
		Metadata:    make(map[string]interface{}),
	}

	return msg
}

// GetHistory returns messages formatted for LLM context
func GetHistory(conv *domain.Conversation) []domain.Message {
	history := []domain.Message{}

	// Add system prompt if present
	if conv.SystemPrompt != "" {
		history = append(history, domain.Message{
			ID:        "system_prompt",
			Role:      domain.MessageRoleSystem,
			Content:   conv.SystemPrompt,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		})
	}

	// Domain messages can be used directly now
	history = append(history, conv.Messages...)

	return history
}

// AddMessageToConversation adds a message to a conversation
func AddMessageToConversation(conv *domain.Conversation, role, content string, attachments []domain.Attachment) {
	msg := NewMessage(role, content, attachments)
	conv.AddMessage(msg)
}

// ResetConversation clears all messages from a conversation
func ResetConversation(conv *domain.Conversation) {
	conv.Messages = []domain.Message{}
	conv.Updated = time.Now()
}

// ContextHelpers - additional REPL-specific conversation utilities

// GetLastUserMessage returns the last user message in the conversation
func GetLastUserMessage(conv *domain.Conversation) *domain.Message {
	for i := len(conv.Messages) - 1; i >= 0; i-- {
		if conv.Messages[i].Role == domain.MessageRoleUser {
			return &conv.Messages[i]
		}
	}
	return nil
}

// GetLastAssistantMessage returns the last assistant message in the conversation
func GetLastAssistantMessage(conv *domain.Conversation) *domain.Message {
	for i := len(conv.Messages) - 1; i >= 0; i-- {
		if conv.Messages[i].Role == domain.MessageRoleAssistant {
			return &conv.Messages[i]
		}
	}
	return nil
}

// TruncateHistory keeps only the last N messages in the conversation
func TruncateHistory(conv *domain.Conversation, maxMessages int) {
	if len(conv.Messages) > maxMessages {
		conv.Messages = conv.Messages[len(conv.Messages)-maxMessages:]
		conv.Updated = time.Now()
	}
}
