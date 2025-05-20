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
		ID:        fmt.Sprintf("msg-%s-%d", role, time.Now().UnixNano()),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
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
		content = "You are a helpful assistant"
	}
	return CreateTestMessage(domain.MessageRoleSystem, content)
}

// CreateTestMessageWithAttachments creates a message with attachments
func CreateTestMessageWithAttachments(role domain.MessageRole, attachmentCount int) domain.Message {
	msg := CreateTestMessage(role, "Message with attachments")

	for i := 0; i < attachmentCount; i++ {
		attachment := CreateTestAttachment(domain.AttachmentTypeText, fmt.Sprintf("file%d.txt", i+1))
		msg.Attachments = append(msg.Attachments, attachment)
	}

	return msg
}

// CreateTestAttachment creates a test attachment
func CreateTestAttachment(attachType domain.AttachmentType, name string) domain.Attachment {
	content := []byte(fmt.Sprintf("Test content for %s", name))
	mimeType := "text/plain"

	switch attachType {
	case domain.AttachmentTypeImage:
		mimeType = "image/jpeg"
		content = []byte("fake image data")
	case domain.AttachmentTypeAudio:
		mimeType = "audio/mp3"
		content = []byte("fake audio data")
	case domain.AttachmentTypeVideo:
		mimeType = "video/mp4"
		content = []byte("fake video data")
	case domain.AttachmentTypeFile:
		mimeType = "application/octet-stream"
		content = []byte("fake binary data")
	}

	return domain.Attachment{
		ID:       fmt.Sprintf("attach-%d", time.Now().UnixNano()),
		Type:     attachType,
		Name:     name,
		Content:  content,
		MimeType: mimeType,
		Size:     int64(len(content)),
		Metadata: make(map[string]interface{}),
	}
}

// CreateTextAttachment creates a text attachment
func CreateTextAttachment(name, content string) domain.Attachment {
	if name == "" {
		name = "test.txt"
	}
	if content == "" {
		content = "Test text content"
	}

	return domain.Attachment{
		ID:       fmt.Sprintf("text-attach-%d", time.Now().UnixNano()),
		Type:     domain.AttachmentTypeText,
		Name:     name,
		Content:  []byte(content),
		MimeType: "text/plain",
		Size:     int64(len(content)),
		Metadata: make(map[string]interface{}),
	}
}

// CreateImageAttachment creates an image attachment
func CreateImageAttachment(name string) domain.Attachment {
	if name == "" {
		name = "test.jpg"
	}

	// Create a minimal valid JPEG header
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}

	return domain.Attachment{
		ID:       fmt.Sprintf("img-attach-%d", time.Now().UnixNano()),
		Type:     domain.AttachmentTypeImage,
		Name:     name,
		Content:  jpegHeader,
		MimeType: "image/jpeg",
		Size:     int64(len(jpegHeader)),
		Metadata: map[string]interface{}{
			"width":  100,
			"height": 100,
		},
	}
}

// CreateConversation creates a test conversation
func CreateConversation(messageCount int) *domain.Conversation {
	conv := domain.NewConversation(fmt.Sprintf("test-conv-%d", time.Now().UnixNano()))
	conv.SystemPrompt = "You are a helpful assistant"

	for i := 0; i < messageCount; i++ {
		// Alternate between user and assistant messages
		if i%2 == 0 {
			userMsg := CreateTestUserMessage(fmt.Sprintf("User message %d", i+1))
			conv.AddMessage(userMsg)
		} else {
			assistantMsg := CreateTestAssistantMessage(fmt.Sprintf("Assistant response %d", i+1))
			conv.AddMessage(assistantMsg)
		}
	}

	return conv
}

// CreateStreamingMessages creates a sequence of messages for streaming tests
func CreateStreamingMessages(chunks []string) []domain.Message {
	var messages []domain.Message

	for i, chunk := range chunks {
		msg := domain.Message{
			ID:        fmt.Sprintf("stream-msg-%d", i),
			Role:      domain.MessageRoleAssistant,
			Content:   chunk,
			Timestamp: time.Now().Add(time.Duration(i) * time.Millisecond),
			Metadata: map[string]interface{}{
				"streaming":   true,
				"chunk_index": i,
			},
		}
		messages = append(messages, msg)
	}

	return messages
}
