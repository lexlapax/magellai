// ABOUTME: Adapter functions for converting between domain and LLM types
// ABOUTME: Provides bidirectional conversion for messages and attachments

package llm

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
)

// ToDomainMessage converts an LLM message to a domain message
func ToDomainMessage(msg Message) *domain.Message {
	// Generate a unique ID based on content and timestamp
	id := fmt.Sprintf("msg_%d_%s", time.Now().UnixNano(), msg.Role)

	domainMsg := &domain.Message{
		ID:          id,
		Role:        toDomainRole(msg.Role),
		Content:     msg.Content,
		Timestamp:   time.Now(),
		Attachments: make([]domain.Attachment, 0, len(msg.Attachments)),
		Metadata:    make(map[string]interface{}),
	}

	// Convert attachments
	for i, att := range msg.Attachments {
		domainAtt := toDomainAttachment(att, i)
		domainMsg.Attachments = append(domainMsg.Attachments, domainAtt)
	}

	return domainMsg
}

// FromDomainMessage converts a domain message to an LLM message
func FromDomainMessage(msg *domain.Message) Message {
	llmMsg := Message{
		Role:        fromDomainRole(msg.Role),
		Content:     msg.Content,
		Attachments: make([]Attachment, 0, len(msg.Attachments)),
	}

	// Convert attachments
	for _, att := range msg.Attachments {
		llmAtt := fromDomainAttachment(att)
		llmMsg.Attachments = append(llmMsg.Attachments, llmAtt)
	}

	return llmMsg
}

// ToDomainMessages converts a slice of LLM messages to domain messages
func ToDomainMessages(messages []Message) []*domain.Message {
	domainMessages := make([]*domain.Message, 0, len(messages))
	for _, msg := range messages {
		domainMessages = append(domainMessages, ToDomainMessage(msg))
	}
	return domainMessages
}

// FromDomainMessages converts a slice of domain messages to LLM messages
func FromDomainMessages(messages []*domain.Message) []Message {
	llmMessages := make([]Message, 0, len(messages))
	for _, msg := range messages {
		llmMessages = append(llmMessages, FromDomainMessage(msg))
	}
	return llmMessages
}

// toDomainRole converts LLM role string to domain MessageRole
func toDomainRole(role string) domain.MessageRole {
	switch strings.ToLower(role) {
	case "user":
		return domain.MessageRoleUser
	case "assistant":
		return domain.MessageRoleAssistant
	case "system":
		return domain.MessageRoleSystem
	default:
		// Default to user if unknown
		return domain.MessageRoleUser
	}
}

// fromDomainRole converts domain MessageRole to LLM role string
func fromDomainRole(role domain.MessageRole) string {
	return strings.ToLower(string(role))
}

// toDomainAttachment converts an LLM attachment to a domain attachment
func toDomainAttachment(att Attachment, index int) domain.Attachment {
	// Generate ID based on type and index
	id := fmt.Sprintf("att_%s_%d_%d", att.Type, index, time.Now().UnixNano())

	domainAtt := domain.Attachment{
		ID:       id,
		Type:     toDomainAttachmentType(att.Type),
		FilePath: att.FilePath,
		MimeType: att.MimeType,
		Metadata: make(map[string]interface{}),
	}

	// Handle content - LLM uses string, domain uses []byte
	if att.Content != "" {
		// If it looks like base64, decode it
		if strings.Contains(att.Content, "base64,") {
			parts := strings.Split(att.Content, ",")
			if len(parts) > 1 {
				if data, err := base64.StdEncoding.DecodeString(parts[1]); err == nil {
					domainAtt.Content = data
				} else {
					// Fallback to string bytes
					domainAtt.Content = []byte(att.Content)
				}
			} else {
				domainAtt.Content = []byte(att.Content)
			}
		} else {
			// Regular text content
			domainAtt.Content = []byte(att.Content)
		}
	}

	// Set name if file path is provided
	if att.FilePath != "" {
		parts := strings.Split(att.FilePath, "/")
		domainAtt.Name = parts[len(parts)-1]
	}

	return domainAtt
}

// fromDomainAttachment converts a domain attachment to an LLM attachment
func fromDomainAttachment(att domain.Attachment) Attachment {
	llmAtt := Attachment{
		Type:     fromDomainAttachmentType(att.Type),
		FilePath: att.FilePath,
		MimeType: att.MimeType,
	}

	// Handle content - domain uses []byte, LLM uses string
	if len(att.Content) > 0 {
		// Check if this is binary data that needs base64 encoding
		if att.Type == domain.AttachmentTypeImage || att.Type == domain.AttachmentTypeFile ||
			att.Type == domain.AttachmentTypeAudio || att.Type == domain.AttachmentTypeVideo {
			// Encode binary data as base64
			llmAtt.Content = "data:" + att.MimeType + ";base64," + base64.StdEncoding.EncodeToString(att.Content)
		} else {
			// Text content can be converted directly
			llmAtt.Content = string(att.Content)
		}
	}

	// If we have a URL, use it instead of file path for certain types
	if att.URL != "" && (att.Type == domain.AttachmentTypeImage || att.Type == domain.AttachmentTypeVideo || att.Type == domain.AttachmentTypeAudio) {
		llmAtt.FilePath = att.URL
	}

	return llmAtt
}

// toDomainAttachmentType converts LLM attachment type to domain attachment type
func toDomainAttachmentType(attType AttachmentType) domain.AttachmentType {
	switch attType {
	case AttachmentTypeImage:
		return domain.AttachmentTypeImage
	case AttachmentTypeAudio:
		return domain.AttachmentTypeAudio
	case AttachmentTypeVideo:
		return domain.AttachmentTypeVideo
	case AttachmentTypeFile:
		return domain.AttachmentTypeFile
	case AttachmentTypeText:
		return domain.AttachmentTypeText
	default:
		return domain.AttachmentTypeFile
	}
}

// fromDomainAttachmentType converts domain attachment type to LLM attachment type
func fromDomainAttachmentType(attType domain.AttachmentType) AttachmentType {
	switch attType {
	case domain.AttachmentTypeImage:
		return AttachmentTypeImage
	case domain.AttachmentTypeAudio:
		return AttachmentTypeAudio
	case domain.AttachmentTypeVideo:
		return AttachmentTypeVideo
	case domain.AttachmentTypeFile:
		return AttachmentTypeFile
	case domain.AttachmentTypeText:
		return AttachmentTypeText
	default:
		return AttachmentTypeFile
	}
}
