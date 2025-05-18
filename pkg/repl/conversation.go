// ABOUTME: Conversation helper functions for REPL operations
// ABOUTME: Provides REPL-specific conversation utilities

package repl

import (
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// NewMessage creates a new domain message with the given parameters
func NewMessage(role, content string, attachments []llm.Attachment) domain.Message {
	msg := domain.Message{
		ID:        uuid.New().String(),
		Role:      domain.MessageRole(role),
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Convert LLM attachments to domain attachments
	if len(attachments) > 0 {
		msg.Attachments = make([]domain.Attachment, len(attachments))
		for i, att := range attachments {
			msg.Attachments[i] = llmAttachmentToDomain(att)
		}
	}

	return msg
}

// GetHistory returns messages formatted for LLM context
func GetHistory(conv *domain.Conversation) []llm.Message {
	history := []llm.Message{}

	// Add system prompt if present
	if conv.SystemPrompt != "" {
		history = append(history, llm.Message{
			Role:    "system",
			Content: conv.SystemPrompt,
		})
	}

	// Convert domain messages to LLM messages
	for _, msg := range conv.Messages {
		history = append(history, convertDomainMessageToLLM(msg))
	}

	return history
}

// AddMessageToConversation adds a message to a conversation
func AddMessageToConversation(conv *domain.Conversation, role, content string, attachments []llm.Attachment) {
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
