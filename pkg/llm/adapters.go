// ABOUTME: Adapter functions for converting between domain types and go-llms types
// ABOUTME: Provides bidirectional conversion for messages and attachments

package llm

import (
	"fmt"
	"time"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/magellai/pkg/domain"
)

// ToLLMMessage converts a domain.Message to go-llms domain.Message
func ToLLMMessage(msg *domain.Message) llmdomain.Message {
	// Convert role - handle tool role not present in domain
	role := llmdomain.Role(msg.Role)

	llmMsg := llmdomain.Message{
		Role: role,
	}

	// Convert simple text content
	if msg.Content != "" && len(msg.Attachments) == 0 {
		llmMsg.Content = []llmdomain.ContentPart{
			{
				Type: llmdomain.ContentTypeText,
				Text: msg.Content,
			},
		}
		return llmMsg
	}

	// Convert with attachments
	llmMsg.Content = make([]llmdomain.ContentPart, 0)

	// Add text content first if present
	if msg.Content != "" {
		llmMsg.Content = append(llmMsg.Content, llmdomain.ContentPart{
			Type: llmdomain.ContentTypeText,
			Text: msg.Content,
		})
	}

	// Add attachments
	for _, att := range msg.Attachments {
		if part := attachmentToLLMContentPart(&att); part != nil {
			llmMsg.Content = append(llmMsg.Content, *part)
		}
	}

	return llmMsg
}

// FromLLMMessage converts a go-llms domain.Message to domain.Message
func FromLLMMessage(msg llmdomain.Message) *domain.Message {
	// Generate a unique ID based on content and timestamp
	id := fmt.Sprintf("msg_%d_%s", time.Now().UnixNano(), msg.Role)

	domainMsg := &domain.Message{
		ID:          id,
		Role:        toDomainRole(msg.Role),
		Timestamp:   time.Now(),
		Attachments: make([]domain.Attachment, 0),
		Metadata:    make(map[string]interface{}),
	}

	// Convert content parts
	for _, part := range msg.Content {
		switch part.Type {
		case llmdomain.ContentTypeText:
			if domainMsg.Content == "" {
				domainMsg.Content = part.Text
			} else {
				// Multiple text parts become text attachments
				att := domain.Attachment{
					ID:       fmt.Sprintf("att_%d", len(domainMsg.Attachments)),
					Type:     domain.AttachmentTypeText,
					Content:  []byte(part.Text),
					Name:     fmt.Sprintf("text_%d", len(domainMsg.Attachments)),
					Metadata: make(map[string]interface{}),
				}
				domainMsg.Attachments = append(domainMsg.Attachments, att)
			}
		default:
			if att := contentPartToDomainAttachment(part, len(domainMsg.Attachments)); att != nil {
				domainMsg.Attachments = append(domainMsg.Attachments, *att)
			}
		}
	}

	return domainMsg
}

// ToLLMMessages converts a slice of domain messages to go-llms messages
func ToLLMMessages(messages []domain.Message) []llmdomain.Message {
	llmMessages := make([]llmdomain.Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = ToLLMMessage(&msg)
	}
	return llmMessages
}

// FromLLMMessages converts a slice of go-llms messages to domain messages
func FromLLMMessages(messages []llmdomain.Message) []domain.Message {
	domainMessages := make([]domain.Message, len(messages))
	for i, msg := range messages {
		domainMessages[i] = *FromLLMMessage(msg)
	}
	return domainMessages
}

// Role conversion helpers

func toDomainRole(role llmdomain.Role) domain.MessageRole {
	switch role {
	case llmdomain.RoleUser:
		return domain.MessageRoleUser
	case llmdomain.RoleAssistant:
		return domain.MessageRoleAssistant
	case llmdomain.RoleSystem:
		return domain.MessageRoleSystem
	case llmdomain.RoleTool:
		// Tool role doesn't exist in domain, map to assistant
		return domain.MessageRoleAssistant
	default:
		return domain.MessageRole(role)
	}
}

// Attachment conversion helpers

func attachmentToLLMContentPart(att *domain.Attachment) *llmdomain.ContentPart {
	switch att.Type {
	case domain.AttachmentTypeImage:
		return &llmdomain.ContentPart{
			Type: llmdomain.ContentTypeImage,
			Image: &llmdomain.ImageContent{
				Source: llmdomain.SourceInfo{
					Type:      llmdomain.SourceTypeBase64,
					Data:      string(att.Content), // Assume base64 encoded
					MediaType: att.MimeType,
				},
			},
		}
	case domain.AttachmentTypeText:
		return &llmdomain.ContentPart{
			Type: llmdomain.ContentTypeText,
			Text: string(att.Content),
		}
	case domain.AttachmentTypeFile:
		return &llmdomain.ContentPart{
			Type: llmdomain.ContentTypeFile,
			File: &llmdomain.FileContent{
				FileName: att.Name,
				FileData: string(att.Content), // Assume base64 encoded
				MimeType: att.MimeType,
			},
		}
	case domain.AttachmentTypeVideo:
		return &llmdomain.ContentPart{
			Type: llmdomain.ContentTypeVideo,
			Video: &llmdomain.VideoContent{
				Source: llmdomain.SourceInfo{
					Type:      llmdomain.SourceTypeURL,
					URL:       att.URL,
					MediaType: att.MimeType,
				},
			},
		}
	case domain.AttachmentTypeAudio:
		return &llmdomain.ContentPart{
			Type: llmdomain.ContentTypeAudio,
			Audio: &llmdomain.AudioContent{
				Source: llmdomain.SourceInfo{
					Type:      llmdomain.SourceTypeURL,
					URL:       att.URL,
					MediaType: att.MimeType,
				},
			},
		}
	}
	return nil
}

func contentPartToDomainAttachment(part llmdomain.ContentPart, index int) *domain.Attachment {
	baseID := fmt.Sprintf("att_%d", index)

	switch part.Type {
	case llmdomain.ContentTypeImage:
		if part.Image != nil {
			return &domain.Attachment{
				ID:       baseID,
				Type:     domain.AttachmentTypeImage,
				Content:  []byte(part.Image.Source.Data),
				URL:      part.Image.Source.URL,
				MimeType: part.Image.Source.MediaType,
				Name:     fmt.Sprintf("image_%d", index),
				Metadata: make(map[string]interface{}),
			}
		}
	case llmdomain.ContentTypeFile:
		if part.File != nil {
			return &domain.Attachment{
				ID:       baseID,
				Type:     domain.AttachmentTypeFile,
				Content:  []byte(part.File.FileData),
				Name:     part.File.FileName,
				MimeType: part.File.MimeType,
				Metadata: make(map[string]interface{}),
			}
		}
	case llmdomain.ContentTypeVideo:
		if part.Video != nil {
			return &domain.Attachment{
				ID:       baseID,
				Type:     domain.AttachmentTypeVideo,
				URL:      part.Video.Source.URL,
				MimeType: part.Video.Source.MediaType,
				Name:     fmt.Sprintf("video_%d", index),
				Metadata: make(map[string]interface{}),
			}
		}
	case llmdomain.ContentTypeAudio:
		if part.Audio != nil {
			return &domain.Attachment{
				ID:       baseID,
				Type:     domain.AttachmentTypeAudio,
				URL:      part.Audio.Source.URL,
				MimeType: part.Audio.Source.MediaType,
				Name:     fmt.Sprintf("audio_%d", index),
				Metadata: make(map[string]interface{}),
			}
		}
	}
	return nil
}
