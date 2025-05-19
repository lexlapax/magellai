package llm

import (
	"testing"
	"time"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestToLLMMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    *domain.Message
		expected llmdomain.Message
	}{
		{
			name: "simple text message",
			input: &domain.Message{
				ID:        "msg1",
				Role:      domain.MessageRoleUser,
				Content:   "Hello, world!",
				Timestamp: time.Now(),
			},
			expected: llmdomain.Message{
				Role: llmdomain.RoleUser,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
		{
			name: "message with attachments",
			input: &domain.Message{
				ID:      "msg2",
				Role:    domain.MessageRoleAssistant,
				Content: "Here's an image",
				Attachments: []domain.Attachment{
					{
						ID:       "att1",
						Type:     domain.AttachmentTypeImage,
						Content:  []byte("base64data"),
						MimeType: "image/png",
					},
				},
			},
			expected: llmdomain.Message{
				Role: llmdomain.RoleAssistant,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "Here's an image",
					},
					{
						Type: llmdomain.ContentTypeImage,
						Image: &llmdomain.ImageContent{
							Source: llmdomain.SourceInfo{
								Type:      llmdomain.SourceTypeBase64,
								Data:      "base64data",
								MediaType: "image/png",
							},
						},
					},
				},
			},
		},
		{
			name: "system message",
			input: &domain.Message{
				ID:      "msg3",
				Role:    domain.MessageRoleSystem,
				Content: "You are a helpful assistant",
			},
			expected: llmdomain.Message{
				Role: llmdomain.RoleSystem,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "You are a helpful assistant",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToLLMMessage(tt.input)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, len(tt.expected.Content), len(result.Content))

			for i := range tt.expected.Content {
				assert.Equal(t, tt.expected.Content[i].Type, result.Content[i].Type)
				if tt.expected.Content[i].Type == llmdomain.ContentTypeText {
					assert.Equal(t, tt.expected.Content[i].Text, result.Content[i].Text)
				}
			}
		})
	}
}

func TestFromLLMMessage(t *testing.T) {
	tests := []struct {
		name  string
		input llmdomain.Message
	}{
		{
			name: "simple text message",
			input: llmdomain.Message{
				Role: llmdomain.RoleUser,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
		{
			name: "message with tool role",
			input: llmdomain.Message{
				Role: llmdomain.RoleTool,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "Tool output",
					},
				},
			},
		},
		{
			name: "multimodal message",
			input: llmdomain.Message{
				Role: llmdomain.RoleAssistant,
				Content: []llmdomain.ContentPart{
					{
						Type: llmdomain.ContentTypeText,
						Text: "Here's an analysis",
					},
					{
						Type: llmdomain.ContentTypeImage,
						Image: &llmdomain.ImageContent{
							Source: llmdomain.SourceInfo{
								Type:      llmdomain.SourceTypeURL,
								URL:       "https://example.com/image.png",
								MediaType: "image/png",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromLLMMessage(tt.input)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.ID)
			assert.NotZero(t, result.Timestamp)

			// Check role conversion
			if tt.input.Role == llmdomain.RoleTool {
				assert.Equal(t, domain.MessageRoleAssistant, result.Role)
			} else {
				assert.Equal(t, domain.MessageRole(tt.input.Role), result.Role)
			}

			// Check content
			if len(tt.input.Content) > 0 && tt.input.Content[0].Type == llmdomain.ContentTypeText {
				assert.Equal(t, tt.input.Content[0].Text, result.Content)
			}
		})
	}
}

func TestRoleConversion(t *testing.T) {
	tests := []struct {
		name     string
		llmRole  llmdomain.Role
		expected domain.MessageRole
	}{
		{
			name:     "user role",
			llmRole:  llmdomain.RoleUser,
			expected: domain.MessageRoleUser,
		},
		{
			name:     "assistant role",
			llmRole:  llmdomain.RoleAssistant,
			expected: domain.MessageRoleAssistant,
		},
		{
			name:     "system role",
			llmRole:  llmdomain.RoleSystem,
			expected: domain.MessageRoleSystem,
		},
		{
			name:     "tool role maps to assistant",
			llmRole:  llmdomain.RoleTool,
			expected: domain.MessageRoleAssistant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDomainRole(tt.llmRole)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAttachmentConversion(t *testing.T) {
	domainAtt := &domain.Attachment{
		ID:       "att1",
		Type:     domain.AttachmentTypeImage,
		Content:  []byte("imagedata"),
		URL:      "https://example.com/image.png",
		MimeType: "image/png",
		Name:     "image.png",
	}

	llmPart := attachmentToLLMContentPart(domainAtt)
	assert.NotNil(t, llmPart)
	assert.Equal(t, llmdomain.ContentTypeImage, llmPart.Type)
	assert.NotNil(t, llmPart.Image)
	assert.Equal(t, "imagedata", llmPart.Image.Source.Data)
	assert.Equal(t, "image/png", llmPart.Image.Source.MediaType)
}
