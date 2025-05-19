// ABOUTME: Tests for adapter functions converting between domain types and go-llms types
// ABOUTME: Comprehensive coverage of message and attachment conversions

package llm

import (
	"fmt"
	"strings"
	"testing"
	"time"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/magellai/pkg/domain"
)

func TestToLLMMessage(t *testing.T) {
	tests := []struct {
		name        string
		domainMsg   *domain.Message
		wantRole    llmdomain.Role
		wantContent int // expected number of content parts
		checkFunc   func(t *testing.T, msg llmdomain.Message)
	}{
		{
			name: "simple text message user role",
			domainMsg: &domain.Message{
				ID:        "msg1",
				Role:      domain.MessageRoleUser,
				Content:   "Hello, world!",
				Timestamp: time.Now(),
			},
			wantRole:    llmdomain.RoleUser,
			wantContent: 1,
			checkFunc: func(t *testing.T, msg llmdomain.Message) {
				if len(msg.Content) != 1 {
					t.Errorf("expected 1 content part, got %d", len(msg.Content))
				}
				if msg.Content[0].Type != llmdomain.ContentTypeText {
					t.Errorf("expected text content type, got %s", msg.Content[0].Type)
				}
				if msg.Content[0].Text != "Hello, world!" {
					t.Errorf("expected 'Hello, world!', got %s", msg.Content[0].Text)
				}
			},
		},
		{
			name: "message with text and image attachment",
			domainMsg: &domain.Message{
				ID:      "msg2",
				Role:    domain.MessageRoleAssistant,
				Content: "Here's an image",
				Attachments: []domain.Attachment{
					{
						ID:       "att1",
						Type:     domain.AttachmentTypeImage,
						Content:  []byte("base64imagedata"),
						MimeType: "image/png",
						Name:     "screenshot.png",
					},
				},
				Timestamp: time.Now(),
			},
			wantRole:    llmdomain.RoleAssistant,
			wantContent: 2,
			checkFunc: func(t *testing.T, msg llmdomain.Message) {
				if len(msg.Content) != 2 {
					t.Errorf("expected 2 content parts, got %d", len(msg.Content))
				}
				// First part should be text
				if msg.Content[0].Type != llmdomain.ContentTypeText {
					t.Errorf("expected first part to be text, got %s", msg.Content[0].Type)
				}
				// Second part should be image
				if msg.Content[1].Type != llmdomain.ContentTypeImage {
					t.Errorf("expected second part to be image, got %s", msg.Content[1].Type)
				}
				if msg.Content[1].Image == nil {
					t.Error("expected image content to be non-nil")
				} else {
					if msg.Content[1].Image.Source.Type != llmdomain.SourceTypeBase64 {
						t.Errorf("expected base64 source type, got %s", msg.Content[1].Image.Source.Type)
					}
					if msg.Content[1].Image.Source.Data != "base64imagedata" {
						t.Errorf("expected 'base64imagedata', got %s", msg.Content[1].Image.Source.Data)
					}
				}
			},
		},
		{
			name: "message with multiple attachments",
			domainMsg: &domain.Message{
				ID:   "msg3",
				Role: domain.MessageRoleUser,
				Attachments: []domain.Attachment{
					{
						ID:       "att1",
						Type:     domain.AttachmentTypeFile,
						Content:  []byte("filedata"),
						MimeType: "application/pdf",
						Name:     "document.pdf",
					},
					{
						ID:       "att2",
						Type:     domain.AttachmentTypeVideo,
						URL:      "https://example.com/video.mp4",
						MimeType: "video/mp4",
						Name:     "video.mp4",
					},
				},
				Timestamp: time.Now(),
			},
			wantRole:    llmdomain.RoleUser,
			wantContent: 2,
			checkFunc: func(t *testing.T, msg llmdomain.Message) {
				if len(msg.Content) != 2 {
					t.Errorf("expected 2 content parts, got %d", len(msg.Content))
				}
				// Check file attachment
				if msg.Content[0].Type != llmdomain.ContentTypeFile {
					t.Errorf("expected file content type, got %s", msg.Content[0].Type)
				}
				if msg.Content[0].File == nil {
					t.Error("expected file content to be non-nil")
				} else {
					if msg.Content[0].File.FileName != "document.pdf" {
						t.Errorf("expected 'document.pdf', got %s", msg.Content[0].File.FileName)
					}
				}
				// Check video attachment
				if msg.Content[1].Type != llmdomain.ContentTypeVideo {
					t.Errorf("expected video content type, got %s", msg.Content[1].Type)
				}
				if msg.Content[1].Video == nil {
					t.Error("expected video content to be non-nil")
				} else {
					if msg.Content[1].Video.Source.URL != "https://example.com/video.mp4" {
						t.Errorf("expected video URL, got %s", msg.Content[1].Video.Source.URL)
					}
				}
			},
		},
		{
			name: "system message",
			domainMsg: &domain.Message{
				ID:        "msg4",
				Role:      domain.MessageRoleSystem,
				Content:   "You are a helpful assistant.",
				Timestamp: time.Now(),
			},
			wantRole:    llmdomain.RoleSystem,
			wantContent: 1,
		},
		{
			name: "empty message with no content",
			domainMsg: &domain.Message{
				ID:        "msg5",
				Role:      domain.MessageRoleUser,
				Content:   "",
				Timestamp: time.Now(),
			},
			wantRole:    llmdomain.RoleUser,
			wantContent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToLLMMessage(tt.domainMsg)
			
			if got.Role != tt.wantRole {
				t.Errorf("expected role %s, got %s", tt.wantRole, got.Role)
			}
			
			if len(got.Content) != tt.wantContent {
				t.Errorf("expected %d content parts, got %d", tt.wantContent, len(got.Content))
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

func TestFromLLMMessage(t *testing.T) {
	tests := []struct {
		name        string
		llmMsg      llmdomain.Message
		wantRole    domain.MessageRole
		checkFunc   func(t *testing.T, msg *domain.Message)
	}{
		{
			name: "simple text message",
			llmMsg: llmdomain.Message{
				Role: llmdomain.RoleUser,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "Hello from LLM",
					},
				},
			},
			wantRole: domain.MessageRoleUser,
			checkFunc: func(t *testing.T, msg *domain.Message) {
				if msg.Content != "Hello from LLM" {
					t.Errorf("expected 'Hello from LLM', got %s", msg.Content)
				}
				if len(msg.Attachments) != 0 {
					t.Errorf("expected no attachments, got %d", len(msg.Attachments))
				}
			},
		},
		{
			name: "message with multiple text parts",
			llmMsg: llmdomain.Message{
				Role: llmdomain.RoleAssistant,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "First part",
					},
					{
						Type: llmdomain.ContentTypeText,
						Text: "Second part",
					},
				},
			},
			wantRole: domain.MessageRoleAssistant,
			checkFunc: func(t *testing.T, msg *domain.Message) {
				if msg.Content != "First part" {
					t.Errorf("expected 'First part', got %s", msg.Content)
				}
				if len(msg.Attachments) != 1 {
					t.Errorf("expected 1 attachment, got %d", len(msg.Attachments))
				}
				if msg.Attachments[0].Type != domain.AttachmentTypeText {
					t.Errorf("expected text attachment, got %s", msg.Attachments[0].Type)
				}
				if string(msg.Attachments[0].Content) != "Second part" {
					t.Errorf("expected 'Second part', got %s", string(msg.Attachments[0].Content))
				}
			},
		},
		{
			name: "tool role mapped to assistant",
			llmMsg: llmdomain.Message{
				Role: llmdomain.RoleTool,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "Tool response",
					},
				},
			},
			wantRole: domain.MessageRoleAssistant,
			checkFunc: func(t *testing.T, msg *domain.Message) {
				if msg.Role != domain.MessageRoleAssistant {
					t.Errorf("expected assistant role, got %s", msg.Role)
				}
			},
		},
		{
			name: "message with image attachment",
			llmMsg: llmdomain.Message{
				Role: llmdomain.RoleUser,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeImage,
						Image: &llmdomain.ImageContent{
							Source: llmdomain.SourceInfo{
								Type:      llmdomain.SourceTypeBase64,
								Data:      "imagedata",
								MediaType: "image/jpeg",
							},
						},
					},
				},
			},
			wantRole: domain.MessageRoleUser,
			checkFunc: func(t *testing.T, msg *domain.Message) {
				if msg.Content != "" {
					t.Errorf("expected empty content, got %s", msg.Content)
				}
				if len(msg.Attachments) != 1 {
					t.Errorf("expected 1 attachment, got %d", len(msg.Attachments))
				}
				if msg.Attachments[0].Type != domain.AttachmentTypeImage {
					t.Errorf("expected image attachment, got %s", msg.Attachments[0].Type)
				}
				if string(msg.Attachments[0].Content) != "imagedata" {
					t.Errorf("expected 'imagedata', got %s", string(msg.Attachments[0].Content))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromLLMMessage(tt.llmMsg)
			
			if got.Role != tt.wantRole {
				t.Errorf("expected role %s, got %s", tt.wantRole, got.Role)
			}
			
			// Check that ID is generated
			if got.ID == "" {
				t.Error("expected ID to be generated")
			}
			
			// Check that timestamp is set
			if got.Timestamp.IsZero() {
				t.Error("expected timestamp to be set")
			}
			
			// Check that metadata is initialized
			if got.Metadata == nil {
				t.Error("expected metadata to be initialized")
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

func TestToLLMMessages(t *testing.T) {
	domainMessages := []domain.Message{
		{
			ID:      "msg1",
			Role:    domain.MessageRoleUser,
			Content: "First message",
		},
		{
			ID:      "msg2",
			Role:    domain.MessageRoleAssistant,
			Content: "Second message",
		},
		{
			ID:      "msg3",
			Role:    domain.MessageRoleSystem,
			Content: "System message",
		},
	}

	llmMessages := ToLLMMessages(domainMessages)

	if len(llmMessages) != len(domainMessages) {
		t.Errorf("expected %d messages, got %d", len(domainMessages), len(llmMessages))
	}

	for i, llmMsg := range llmMessages {
		expectedRole := llmdomain.Role(domainMessages[i].Role)
		if llmMsg.Role != expectedRole {
			t.Errorf("message %d: expected role %s, got %s", i, expectedRole, llmMsg.Role)
		}
		if len(llmMsg.Content) != 1 {
			t.Errorf("message %d: expected 1 content part, got %d", i, len(llmMsg.Content))
		}
		if llmMsg.Content[0].Text != domainMessages[i].Content {
			t.Errorf("message %d: expected content %s, got %s", i, domainMessages[i].Content, llmMsg.Content[0].Text)
		}
	}
}

func TestFromLLMMessages(t *testing.T) {
	llmMessages := []llmdomain.Message{
		{
			Role: llmdomain.RoleUser,
			Content: []llmdomain.ContentPart{
				{Type: llmdomain.ContentTypeText, Text: "User message"},
			},
		},
		{
			Role: llmdomain.RoleAssistant,
			Content: []llmdomain.ContentPart{
				{Type: llmdomain.ContentTypeText, Text: "Assistant message"},
			},
		},
	}

	domainMessages := FromLLMMessages(llmMessages)

	if len(domainMessages) != len(llmMessages) {
		t.Errorf("expected %d messages, got %d", len(llmMessages), len(domainMessages))
	}

	for i, domainMsg := range domainMessages {
		// Check role conversion
		switch llmMessages[i].Role {
		case llmdomain.RoleUser:
			if domainMsg.Role != domain.MessageRoleUser {
				t.Errorf("message %d: expected user role, got %s", i, domainMsg.Role)
			}
		case llmdomain.RoleAssistant:
			if domainMsg.Role != domain.MessageRoleAssistant {
				t.Errorf("message %d: expected assistant role, got %s", i, domainMsg.Role)
			}
		}
		
		// Check content
		if domainMsg.Content != llmMessages[i].Content[0].Text {
			t.Errorf("message %d: expected content %s, got %s", i, llmMessages[i].Content[0].Text, domainMsg.Content)
		}
		
		// Check that ID is generated
		if domainMsg.ID == "" {
			t.Errorf("message %d: expected ID to be generated", i)
		}
	}
}

func TestAttachmentConversion(t *testing.T) {
	tests := []struct {
		name       string
		attachment *domain.Attachment
		wantType   llmdomain.ContentType
		checkFunc  func(t *testing.T, part *llmdomain.ContentPart)
	}{
		{
			name: "image attachment",
			attachment: &domain.Attachment{
				Type:     domain.AttachmentTypeImage,
				Content:  []byte("imagedata"),
				MimeType: "image/png",
			},
			wantType: llmdomain.ContentTypeImage,
			checkFunc: func(t *testing.T, part *llmdomain.ContentPart) {
				if part.Image == nil {
					t.Fatal("expected image content")
				}
				if part.Image.Source.Type != llmdomain.SourceTypeBase64 {
					t.Errorf("expected base64 source type, got %s", part.Image.Source.Type)
				}
				if part.Image.Source.Data != "imagedata" {
					t.Errorf("expected 'imagedata', got %s", part.Image.Source.Data)
				}
			},
		},
		{
			name: "text attachment",
			attachment: &domain.Attachment{
				Type:    domain.AttachmentTypeText,
				Content: []byte("text content"),
			},
			wantType: llmdomain.ContentTypeText,
			checkFunc: func(t *testing.T, part *llmdomain.ContentPart) {
				if part.Text != "text content" {
					t.Errorf("expected 'text content', got %s", part.Text)
				}
			},
		},
		{
			name: "file attachment",
			attachment: &domain.Attachment{
				Type:     domain.AttachmentTypeFile,
				Content:  []byte("filedata"),
				Name:     "document.pdf",
				MimeType: "application/pdf",
			},
			wantType: llmdomain.ContentTypeFile,
			checkFunc: func(t *testing.T, part *llmdomain.ContentPart) {
				if part.File == nil {
					t.Fatal("expected file content")
				}
				if part.File.FileName != "document.pdf" {
					t.Errorf("expected 'document.pdf', got %s", part.File.FileName)
				}
				if part.File.FileData != "filedata" {
					t.Errorf("expected 'filedata', got %s", part.File.FileData)
				}
			},
		},
		{
			name: "video attachment with URL",
			attachment: &domain.Attachment{
				Type:     domain.AttachmentTypeVideo,
				URL:      "https://example.com/video.mp4",
				MimeType: "video/mp4",
			},
			wantType: llmdomain.ContentTypeVideo,
			checkFunc: func(t *testing.T, part *llmdomain.ContentPart) {
				if part.Video == nil {
					t.Fatal("expected video content")
				}
				if part.Video.Source.Type != llmdomain.SourceTypeURL {
					t.Errorf("expected URL source type, got %s", part.Video.Source.Type)
				}
				if part.Video.Source.URL != "https://example.com/video.mp4" {
					t.Errorf("expected URL, got %s", part.Video.Source.URL)
				}
			},
		},
		{
			name: "audio attachment with URL",
			attachment: &domain.Attachment{
				Type:     domain.AttachmentTypeAudio,
				URL:      "https://example.com/audio.mp3",
				MimeType: "audio/mpeg",
			},
			wantType: llmdomain.ContentTypeAudio,
			checkFunc: func(t *testing.T, part *llmdomain.ContentPart) {
				if part.Audio == nil {
					t.Fatal("expected audio content")
				}
				if part.Audio.Source.Type != llmdomain.SourceTypeURL {
					t.Errorf("expected URL source type, got %s", part.Audio.Source.Type)
				}
				if part.Audio.Source.URL != "https://example.com/audio.mp3" {
					t.Errorf("expected URL, got %s", part.Audio.Source.URL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := attachmentToLLMContentPart(tt.attachment)
			if part == nil {
				t.Fatal("expected non-nil content part")
			}
			
			if part.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, part.Type)
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, part)
			}
		})
	}
}

func TestContentPartToDomainAttachment(t *testing.T) {
	tests := []struct {
		name      string
		part      llmdomain.ContentPart
		index     int
		wantType  domain.AttachmentType
		checkFunc func(t *testing.T, att *domain.Attachment)
	}{
		{
			name: "image content part",
			part: llmdomain.ContentPart{
				Type: llmdomain.ContentTypeImage,
				Image: &llmdomain.ImageContent{
					Source: llmdomain.SourceInfo{
						Type:      llmdomain.SourceTypeBase64,
						Data:      "imagedata",
						MediaType: "image/jpeg",
					},
				},
			},
			index:    0,
			wantType: domain.AttachmentTypeImage,
			checkFunc: func(t *testing.T, att *domain.Attachment) {
				if string(att.Content) != "imagedata" {
					t.Errorf("expected 'imagedata', got %s", string(att.Content))
				}
				if att.MimeType != "image/jpeg" {
					t.Errorf("expected 'image/jpeg', got %s", att.MimeType)
				}
				if !strings.HasPrefix(att.Name, "image_") {
					t.Errorf("expected name to start with 'image_', got %s", att.Name)
				}
			},
		},
		{
			name: "file content part",
			part: llmdomain.ContentPart{
				Type: llmdomain.ContentTypeFile,
				File: &llmdomain.FileContent{
					FileName: "report.pdf",
					FileData: "pdfdata",
					MimeType: "application/pdf",
				},
			},
			index:    1,
			wantType: domain.AttachmentTypeFile,
			checkFunc: func(t *testing.T, att *domain.Attachment) {
				if string(att.Content) != "pdfdata" {
					t.Errorf("expected 'pdfdata', got %s", string(att.Content))
				}
				if att.Name != "report.pdf" {
					t.Errorf("expected 'report.pdf', got %s", att.Name)
				}
				if att.MimeType != "application/pdf" {
					t.Errorf("expected 'application/pdf', got %s", att.MimeType)
				}
			},
		},
		{
			name: "video content part with URL",
			part: llmdomain.ContentPart{
				Type: llmdomain.ContentTypeVideo,
				Video: &llmdomain.VideoContent{
					Source: llmdomain.SourceInfo{
						Type:      llmdomain.SourceTypeURL,
						URL:       "https://example.com/video.mp4",
						MediaType: "video/mp4",
					},
				},
			},
			index:    2,
			wantType: domain.AttachmentTypeVideo,
			checkFunc: func(t *testing.T, att *domain.Attachment) {
				if att.URL != "https://example.com/video.mp4" {
					t.Errorf("expected URL, got %s", att.URL)
				}
				if att.MimeType != "video/mp4" {
					t.Errorf("expected 'video/mp4', got %s", att.MimeType)
				}
			},
		},
		{
			name: "audio content part with URL",
			part: llmdomain.ContentPart{
				Type: llmdomain.ContentTypeAudio,
				Audio: &llmdomain.AudioContent{
					Source: llmdomain.SourceInfo{
						Type:      llmdomain.SourceTypeURL,
						URL:       "https://example.com/sound.mp3",
						MediaType: "audio/mpeg",
					},
				},
			},
			index:    3,
			wantType: domain.AttachmentTypeAudio,
			checkFunc: func(t *testing.T, att *domain.Attachment) {
				if att.URL != "https://example.com/sound.mp3" {
					t.Errorf("expected URL, got %s", att.URL)
				}
				if att.MimeType != "audio/mpeg" {
					t.Errorf("expected 'audio/mpeg', got %s", att.MimeType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			att := contentPartToDomainAttachment(tt.part, tt.index)
			if att == nil {
				t.Fatal("expected non-nil attachment")
			}
			
			if att.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, att.Type)
			}
			
			// Check that ID is generated
			expectedID := fmt.Sprintf("att_%d", tt.index)
			if att.ID != expectedID {
				t.Errorf("expected ID %s, got %s", expectedID, att.ID)
			}
			
			// Check that metadata is initialized
			if att.Metadata == nil {
				t.Error("expected metadata to be initialized")
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, att)
			}
		})
	}
}

func TestRoleConversion(t *testing.T) {
	tests := []struct {
		name       string
		llmRole    llmdomain.Role
		domainRole domain.MessageRole
	}{
		{
			name:       "user role",
			llmRole:    llmdomain.RoleUser,
			domainRole: domain.MessageRoleUser,
		},
		{
			name:       "assistant role",
			llmRole:    llmdomain.RoleAssistant,
			domainRole: domain.MessageRoleAssistant,
		},
		{
			name:       "system role",
			llmRole:    llmdomain.RoleSystem,
			domainRole: domain.MessageRoleSystem,
		},
		{
			name:       "tool role maps to assistant",
			llmRole:    llmdomain.RoleTool,
			domainRole: domain.MessageRoleAssistant,
		},
		{
			name:       "unknown role preserved",
			llmRole:    llmdomain.Role("custom"),
			domainRole: domain.MessageRole("custom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDomainRole(tt.llmRole)
			if got != tt.domainRole {
				t.Errorf("expected %s, got %s", tt.domainRole, got)
			}
		})
	}
}

func TestNilHandling(t *testing.T) {
	// Test nil attachment
	part := attachmentToLLMContentPart(nil)
	if part != nil {
		t.Error("expected nil for nil attachment")
	}

	// Test content part with nil content
	nilPart := contentPartToDomainAttachment(llmdomain.ContentPart{
		Type:  llmdomain.ContentTypeImage,
		Image: nil,
	}, 0)
	if nilPart != nil {
		t.Error("expected nil for nil image content")
	}
}