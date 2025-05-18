// ABOUTME: Tests for LLM adapter functions that convert between domain and LLM types
// ABOUTME: Ensures proper type conversion, field mapping, and edge case handling

package repl

import (
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
)

func TestLLMAttachmentToDomain(t *testing.T) {
	tests := []struct {
		name     string
		llm      llm.Attachment
		expected domain.Attachment
	}{
		{
			name: "file attachment",
			llm: llm.Attachment{
				Type:     llm.AttachmentTypeFile,
				FilePath: "/path/to/document.pdf",
				MimeType: "application/pdf",
			},
			expected: domain.Attachment{
				Type:     domain.AttachmentTypeFile,
				URL:      "/path/to/document.pdf",
				MimeType: "application/pdf",
				Name:     "/path/to/document.pdf",
				Content:  []byte{},
				Metadata: map[string]interface{}{},
			},
		},
		{
			name: "image attachment with content",
			llm: llm.Attachment{
				Type:     llm.AttachmentTypeImage,
				FilePath: "/path/to/image.png",
				MimeType: "image/png",
				Content:  "base64encodedcontent",
			},
			expected: domain.Attachment{
				Type:     domain.AttachmentTypeImage,
				URL:      "/path/to/image.png",
				MimeType: "image/png",
				Name:     "/path/to/image.png",
				Content:  []byte("base64encodedcontent"),
				Metadata: map[string]interface{}{},
			},
		},
		{
			name: "text attachment",
			llm: llm.Attachment{
				Type:     llm.AttachmentTypeText,
				MimeType: "text/plain",
				Content:  "Plain text content",
			},
			expected: domain.Attachment{
				Type:     domain.AttachmentTypeText,
				URL:      "",
				MimeType: "text/plain",
				Name:     "",
				Content:  []byte("Plain text content"),
				Metadata: map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := llmAttachmentToDomain(tt.llm)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.URL, result.URL)
			assert.Equal(t, tt.expected.MimeType, result.MimeType)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Content, result.Content)
			assert.Equal(t, tt.expected.Metadata, result.Metadata)
		})
	}
}

func TestDomainAttachmentToLLM(t *testing.T) {
	tests := []struct {
		name     string
		domain   domain.Attachment
		expected llm.Attachment
	}{
		{
			name: "file attachment",
			domain: domain.Attachment{
				Type:     domain.AttachmentTypeFile,
				URL:      "/path/to/document.pdf",
				MimeType: "application/pdf",
				Name:     "document.pdf",
			},
			expected: llm.Attachment{
				Type:     llm.AttachmentTypeFile,
				FilePath: "document.pdf",
				MimeType: "application/pdf",
			},
		},
		{
			name: "image attachment with content",
			domain: domain.Attachment{
				Type:     domain.AttachmentTypeImage,
				URL:      "/path/to/image.png",
				MimeType: "image/png",
				Name:     "image.png",
				Content:  []byte("base64encodedcontent"),
			},
			expected: llm.Attachment{
				Type:     llm.AttachmentTypeImage,
				FilePath: "image.png",
				MimeType: "image/png",
				Content:  "base64encodedcontent",
			},
		},
		{
			name: "text attachment",
			domain: domain.Attachment{
				Type:     domain.AttachmentTypeText,
				MimeType: "text/plain",
				Content:  []byte("Plain text content"),
			},
			expected: llm.Attachment{
				Type:     llm.AttachmentTypeText,
				MimeType: "text/plain",
				Content:  "Plain text content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domainAttachmentToLLM(tt.domain)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.FilePath, result.FilePath)
			assert.Equal(t, tt.expected.MimeType, result.MimeType)
			assert.Equal(t, tt.expected.Content, result.Content)
		})
	}
}

func TestConvertDomainMessageToLLM(t *testing.T) {
	tests := []struct {
		name     string
		domain   domain.Message
		expected llm.Message
	}{
		{
			name: "simple message",
			domain: domain.Message{
				Role:    domain.MessageRoleUser,
				Content: "Hello, AI!",
			},
			expected: llm.Message{
				Role:    "user",
				Content: "Hello, AI!",
			},
		},
		{
			name: "assistant message with attachments",
			domain: domain.Message{
				Role:    domain.MessageRoleAssistant,
				Content: "Here's my analysis",
				Attachments: []domain.Attachment{
					{
						Type:     domain.AttachmentTypeFile,
						Name:     "analysis.pdf",
						MimeType: "application/pdf",
					},
				},
			},
			expected: llm.Message{
				Role:    "assistant",
				Content: "Here's my analysis",
				Attachments: []llm.Attachment{
					{
						Type:     llm.AttachmentTypeFile,
						FilePath: "analysis.pdf",
						MimeType: "application/pdf",
					},
				},
			},
		},
		{
			name: "system message",
			domain: domain.Message{
				Role:    domain.MessageRoleSystem,
				Content: "You are a helpful assistant",
			},
			expected: llm.Message{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertDomainMessageToLLM(tt.domain)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.Content, result.Content)
			assert.Equal(t, len(tt.expected.Attachments), len(result.Attachments))

			for i, expectedAttachment := range tt.expected.Attachments {
				assert.Equal(t, expectedAttachment.Type, result.Attachments[i].Type)
				assert.Equal(t, expectedAttachment.FilePath, result.Attachments[i].FilePath)
				assert.Equal(t, expectedAttachment.MimeType, result.Attachments[i].MimeType)
			}
		})
	}
}

func TestAdapterEdgeCases(t *testing.T) {
	t.Run("empty domain attachment to LLM", func(t *testing.T) {
		domainAtt := domain.Attachment{}
		llmAtt := domainAttachmentToLLM(domainAtt)

		assert.Equal(t, llm.AttachmentType(""), llmAtt.Type)
		assert.Equal(t, "", llmAtt.FilePath)
		assert.Equal(t, "", llmAtt.MimeType)
		assert.Equal(t, "", llmAtt.Content)
	})

	t.Run("empty LLM attachment to domain", func(t *testing.T) {
		llmAtt := llm.Attachment{}
		domainAtt := llmAttachmentToDomain(llmAtt)

		assert.Equal(t, domain.AttachmentType(""), domainAtt.Type)
		assert.Equal(t, "", domainAtt.URL)
		assert.Equal(t, "", domainAtt.Name)
		assert.Equal(t, "", domainAtt.MimeType)
		assert.Equal(t, []byte{}, domainAtt.Content)
		assert.NotNil(t, domainAtt.Metadata)
	})

	t.Run("domain message with nil attachments", func(t *testing.T) {
		domainMsg := domain.Message{
			Role:        domain.MessageRoleUser,
			Content:     "test",
			Attachments: nil,
		}

		llmMsg := convertDomainMessageToLLM(domainMsg)
		assert.Equal(t, "user", llmMsg.Role)
		assert.Equal(t, "test", llmMsg.Content)
		assert.Nil(t, llmMsg.Attachments)
	})

	t.Run("domain message with empty attachments", func(t *testing.T) {
		domainMsg := domain.Message{
			Role:        domain.MessageRoleUser,
			Content:     "test",
			Attachments: []domain.Attachment{},
		}

		llmMsg := convertDomainMessageToLLM(domainMsg)
		assert.Equal(t, "user", llmMsg.Role)
		assert.Equal(t, "test", llmMsg.Content)
		assert.Nil(t, llmMsg.Attachments)
	})
}

func TestRoundTripConversion(t *testing.T) {
	// Test that converting from LLM -> domain -> LLM preserves data
	originalAttachment := llm.Attachment{
		Type:     llm.AttachmentTypeFile,
		FilePath: "document.pdf",
		MimeType: "application/pdf",
		Content:  "pdf content",
	}

	// Convert to domain
	domainAtt := llmAttachmentToDomain(originalAttachment)

	// Convert back to LLM
	llmAtt := domainAttachmentToLLM(domainAtt)

	// The round trip changes some fields due to mapping differences
	assert.Equal(t, originalAttachment.Type, llmAtt.Type)
	assert.Equal(t, originalAttachment.MimeType, llmAtt.MimeType)
	assert.Equal(t, originalAttachment.Content, llmAtt.Content)
	// FilePath changes because domain uses Name for this
	assert.Equal(t, originalAttachment.FilePath, llmAtt.FilePath)
}
