package llm

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
)

func TestToDomainMessage(t *testing.T) {
	tests := []struct {
		name     string
		llmMsg   Message
		validate func(t *testing.T, domainMsg *domain.Message)
	}{
		{
			name: "simple text message",
			llmMsg: Message{
				Role:    "user",
				Content: "Hello, world!",
			},
			validate: func(t *testing.T, domainMsg *domain.Message) {
				if domainMsg.Role != domain.MessageRoleUser {
					t.Errorf("expected role %s, got %s", domain.MessageRoleUser, domainMsg.Role)
				}
				if domainMsg.Content != "Hello, world!" {
					t.Errorf("expected content 'Hello, world!', got %s", domainMsg.Content)
				}
				if len(domainMsg.Attachments) != 0 {
					t.Errorf("expected no attachments, got %d", len(domainMsg.Attachments))
				}
			},
		},
		{
			name: "message with text attachment",
			llmMsg: Message{
				Role:    "assistant",
				Content: "Here's the file content:",
				Attachments: []Attachment{
					{
						Type:    AttachmentTypeText,
						Content: "File content here",
					},
				},
			},
			validate: func(t *testing.T, domainMsg *domain.Message) {
				if domainMsg.Role != domain.MessageRoleAssistant {
					t.Errorf("expected role %s, got %s", domain.MessageRoleAssistant, domainMsg.Role)
				}
				if len(domainMsg.Attachments) != 1 {
					t.Fatalf("expected 1 attachment, got %d", len(domainMsg.Attachments))
				}
				att := domainMsg.Attachments[0]
				if att.Type != domain.AttachmentTypeText {
					t.Errorf("expected attachment type %s, got %s", domain.AttachmentTypeText, att.Type)
				}
				if string(att.Content) != "File content here" {
					t.Errorf("expected content 'File content here', got %s", string(att.Content))
				}
			},
		},
		{
			name: "message with image attachment",
			llmMsg: Message{
				Role:    "user",
				Content: "Check this image",
				Attachments: []Attachment{
					{
						Type:     AttachmentTypeImage,
						Content:  "data:image/png;base64,aGVsbG8=",
						MimeType: "image/png",
					},
				},
			},
			validate: func(t *testing.T, domainMsg *domain.Message) {
				if len(domainMsg.Attachments) != 1 {
					t.Fatalf("expected 1 attachment, got %d", len(domainMsg.Attachments))
				}
				att := domainMsg.Attachments[0]
				if att.Type != domain.AttachmentTypeImage {
					t.Errorf("expected attachment type %s, got %s", domain.AttachmentTypeImage, att.Type)
				}
				if att.MimeType != "image/png" {
					t.Errorf("expected mime type 'image/png', got %s", att.MimeType)
				}
				// Base64 "aGVsbG8=" decodes to "hello"
				if string(att.Content) != "hello" {
					t.Errorf("expected decoded content 'hello', got %s", string(att.Content))
				}
			},
		},
		{
			name: "system message",
			llmMsg: Message{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
			validate: func(t *testing.T, domainMsg *domain.Message) {
				if domainMsg.Role != domain.MessageRoleSystem {
					t.Errorf("expected role %s, got %s", domain.MessageRoleSystem, domainMsg.Role)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domainMsg := ToDomainMessage(tc.llmMsg)
			tc.validate(t, domainMsg)
		})
	}
}

func TestFromDomainMessage(t *testing.T) {
	tests := []struct {
		name      string
		domainMsg *domain.Message
		validate  func(t *testing.T, llmMsg Message)
	}{
		{
			name: "simple text message",
			domainMsg: &domain.Message{
				ID:      "msg123",
				Role:    domain.MessageRoleUser,
				Content: "Hello, world!",
			},
			validate: func(t *testing.T, llmMsg Message) {
				if llmMsg.Role != "user" {
					t.Errorf("expected role 'user', got %s", llmMsg.Role)
				}
				if llmMsg.Content != "Hello, world!" {
					t.Errorf("expected content 'Hello, world!', got %s", llmMsg.Content)
				}
				if len(llmMsg.Attachments) != 0 {
					t.Errorf("expected no attachments, got %d", len(llmMsg.Attachments))
				}
			},
		},
		{
			name: "message with text attachment",
			domainMsg: &domain.Message{
				ID:      "msg456",
				Role:    domain.MessageRoleAssistant,
				Content: "Here's the content:",
				Attachments: []domain.Attachment{
					{
						ID:      "att1",
						Type:    domain.AttachmentTypeText,
						Content: []byte("Text content"),
					},
				},
			},
			validate: func(t *testing.T, llmMsg Message) {
				if llmMsg.Role != "assistant" {
					t.Errorf("expected role 'assistant', got %s", llmMsg.Role)
				}
				if len(llmMsg.Attachments) != 1 {
					t.Fatalf("expected 1 attachment, got %d", len(llmMsg.Attachments))
				}
				att := llmMsg.Attachments[0]
				if att.Type != AttachmentTypeText {
					t.Errorf("expected attachment type %s, got %s", AttachmentTypeText, att.Type)
				}
				if att.Content != "Text content" {
					t.Errorf("expected content 'Text content', got %s", att.Content)
				}
			},
		},
		{
			name: "message with binary attachment",
			domainMsg: &domain.Message{
				ID:      "msg789",
				Role:    domain.MessageRoleUser,
				Content: "Check this image",
				Attachments: []domain.Attachment{
					{
						ID:       "att2",
						Type:     domain.AttachmentTypeImage,
						Content:  []byte("hello"),
						MimeType: "image/png",
					},
				},
			},
			validate: func(t *testing.T, llmMsg Message) {
				if len(llmMsg.Attachments) != 1 {
					t.Fatalf("expected 1 attachment, got %d", len(llmMsg.Attachments))
				}
				att := llmMsg.Attachments[0]
				if att.Type != AttachmentTypeImage {
					t.Errorf("expected attachment type %s, got %s", AttachmentTypeImage, att.Type)
				}
				// Should be base64 encoded with data URI
				expectedPrefix := "data:image/png;base64,"
				if !strings.HasPrefix(att.Content, expectedPrefix) {
					t.Errorf("expected content to start with %s, got %s", expectedPrefix, att.Content[:30])
				}
				// Extract and decode the base64 part
				base64Part := strings.TrimPrefix(att.Content, expectedPrefix)
				decoded, err := base64.StdEncoding.DecodeString(base64Part)
				if err != nil {
					t.Fatalf("failed to decode base64: %v", err)
				}
				if string(decoded) != "hello" {
					t.Errorf("expected decoded content 'hello', got %s", string(decoded))
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			llmMsg := FromDomainMessage(tc.domainMsg)
			tc.validate(t, llmMsg)
		})
	}
}

func TestRoleConversion(t *testing.T) {
	tests := []struct {
		name       string
		llmRole    string
		domainRole domain.MessageRole
	}{
		{"user role", "user", domain.MessageRoleUser},
		{"assistant role", "assistant", domain.MessageRoleAssistant},
		{"system role", "system", domain.MessageRoleSystem},
		{"uppercase role", "USER", domain.MessageRoleUser},
		{"unknown role defaults to user", "unknown", domain.MessageRoleUser},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test LLM to domain
			domainRole := toDomainRole(tc.llmRole)
			if domainRole != tc.domainRole && tc.llmRole != "unknown" {
				t.Errorf("toDomainRole(%s): expected %s, got %s", tc.llmRole, tc.domainRole, domainRole)
			}

			// Test domain to LLM
			llmRole := fromDomainRole(tc.domainRole)
			expectedLLM := strings.ToLower(string(tc.domainRole))
			if llmRole != expectedLLM {
				t.Errorf("fromDomainRole(%s): expected %s, got %s", tc.domainRole, expectedLLM, llmRole)
			}
		})
	}
}

func TestAttachmentTypeConversion(t *testing.T) {
	tests := []struct {
		name       string
		llmType    AttachmentType
		domainType domain.AttachmentType
	}{
		{"image type", AttachmentTypeImage, domain.AttachmentTypeImage},
		{"audio type", AttachmentTypeAudio, domain.AttachmentTypeAudio},
		{"video type", AttachmentTypeVideo, domain.AttachmentTypeVideo},
		{"file type", AttachmentTypeFile, domain.AttachmentTypeFile},
		{"text type", AttachmentTypeText, domain.AttachmentTypeText},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test LLM to domain
			domainType := toDomainAttachmentType(tc.llmType)
			if domainType != tc.domainType {
				t.Errorf("toDomainAttachmentType(%s): expected %s, got %s", tc.llmType, tc.domainType, domainType)
			}

			// Test domain to LLM
			llmType := fromDomainAttachmentType(tc.domainType)
			if llmType != tc.llmType {
				t.Errorf("fromDomainAttachmentType(%s): expected %s, got %s", tc.domainType, tc.llmType, llmType)
			}
		})
	}
}

func TestMessagesSliceConversion(t *testing.T) {
	// Test ToDomainMessages
	llmMessages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
		{Role: "user", Content: "How are you?"},
	}

	domainMessages := ToDomainMessages(llmMessages)
	if len(domainMessages) != len(llmMessages) {
		t.Errorf("expected %d messages, got %d", len(llmMessages), len(domainMessages))
	}

	for i, msg := range domainMessages {
		expectedRole := toDomainRole(llmMessages[i].Role)
		if msg.Role != expectedRole {
			t.Errorf("message %d: expected role %s, got %s", i, expectedRole, msg.Role)
		}
		if msg.Content != llmMessages[i].Content {
			t.Errorf("message %d: expected content %s, got %s", i, llmMessages[i].Content, msg.Content)
		}
	}

	// Test FromDomainMessages
	convertedBack := FromDomainMessages(domainMessages)
	if len(convertedBack) != len(domainMessages) {
		t.Errorf("expected %d messages, got %d", len(domainMessages), len(convertedBack))
	}

	for i, msg := range convertedBack {
		if msg.Role != llmMessages[i].Role {
			t.Errorf("message %d: expected role %s, got %s", i, llmMessages[i].Role, msg.Role)
		}
		if msg.Content != llmMessages[i].Content {
			t.Errorf("message %d: expected content %s, got %s", i, llmMessages[i].Content, msg.Content)
		}
	}
}