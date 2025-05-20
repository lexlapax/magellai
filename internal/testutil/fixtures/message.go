// ABOUTME: Test fixtures for message and attachment objects
// ABOUTME: Provides reusable test data for message-related tests

package fixtures

import (
	"fmt"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
)

// CreateTestMessage creates a test message
func CreateTestMessage(role domain.MessageRole, content string) domain.Message {
	return domain.Message{
		ID:          fmt.Sprintf("msg-%s-%d", role, time.Now().UnixNano()),
		Role:        role,
		Content:     content,
		Timestamp:   time.Now(),
		Attachments: []domain.Attachment{},
		Metadata:    make(map[string]interface{}),
	}
}

// CreateTestUserMessage creates a test user message
func CreateTestUserMessage(content string) domain.Message {
	if content == "" {
		content = "Test user message"
	}
	return CreateTestMessage(domain.MessageRoleUser, content)
}

// CreateTestAssistantMessage creates a test assistant message
func CreateTestAssistantMessage(content string) domain.Message {
	if content == "" {
		content = "Test assistant response"
	}
	return CreateTestMessage(domain.MessageRoleAssistant, content)
}

// CreateTestSystemMessage creates a test system message
func CreateTestSystemMessage(content string) domain.Message {
	if content == "" {
		content = "System information message"
	}
	return CreateTestMessage(domain.MessageRoleSystem, content)
}

// CreateTestMessageWithAttachment creates a message with a text attachment
func CreateTestMessageWithAttachment(role domain.MessageRole, content, filename, attachmentContent string) domain.Message {
	msg := CreateTestMessage(role, content)

	attachment := domain.Attachment{
		ID:       fmt.Sprintf("att-%d", time.Now().UnixNano()),
		Type:     domain.AttachmentTypeText,
		Name:     filename,
		MimeType: "text/plain",
		Content:  []byte(attachmentContent),
		Size:     int64(len(attachmentContent)),
		Metadata: make(map[string]interface{}),
	}

	msg.Attachments = append(msg.Attachments, attachment)
	return msg
}

// CreateTestImageAttachment creates an image attachment (fake content)
func CreateTestImageAttachment(filename string) domain.Attachment {
	// Create a minimal JPEG header (not a real image, just for testing)
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}

	return domain.Attachment{
		ID:       fmt.Sprintf("img-%d", time.Now().UnixNano()),
		Type:     domain.AttachmentTypeImage,
		Name:     filename,
		MimeType: "image/jpeg",
		Content:  jpegHeader,
		Size:     int64(len(jpegHeader)),
		Metadata: make(map[string]interface{}),
	}
}

// CreateTestConversation creates a test conversation with messages
func CreateTestConversation(id string, messages []domain.Message) *domain.Conversation {
	if id == "" {
		id = fmt.Sprintf("conv-%d", time.Now().UnixNano())
	}

	now := time.Now()
	conversation := &domain.Conversation{
		ID:       id,
		Created:  now,
		Updated:  now,
		Messages: messages,
		Metadata: make(map[string]interface{}),
	}

	return conversation
}
