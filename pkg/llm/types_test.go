// ABOUTME: Unit tests for LLM types and conversions
// ABOUTME: Tests wrapper types and conversion methods between Magellai and go-llms
package llm

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestParseModelString(t *testing.T) {
	tests := []struct {
		input    string
		provider string
		model    string
	}{
		{"openai/gpt-4", "openai", "gpt-4"},
		{"anthropic/claude-3-opus", "anthropic", "claude-3-opus"},
		{"gemini/gemini-pro", "gemini", "gemini-pro"},
		{"gpt-3.5-turbo", "openai", "gpt-3.5-turbo"}, // Default to OpenAI
		{"", "openai", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			provider, model := ParseModelString(tt.input)
			if provider != tt.provider {
				t.Errorf("ParseModelString(%q) provider = %q, want %q", tt.input, provider, tt.provider)
			}
			if model != tt.model {
				t.Errorf("ParseModelString(%q) model = %q, want %q", tt.input, model, tt.model)
			}
		})
	}
}

func TestFormatModelString(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		expected string
	}{
		{"openai", "gpt-4", "openai/gpt-4"},
		{"anthropic", "claude-3-opus", "anthropic/claude-3-opus"},
		{"gemini", "gemini-pro", "gemini/gemini-pro"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatModelString(tt.provider, tt.model)
			if result != tt.expected {
				t.Errorf("FormatModelString(%q, %q) = %q, want %q", tt.provider, tt.model, result, tt.expected)
			}
		})
	}
}

func TestMessageConversion(t *testing.T) {
	// Test basic message conversion
	msg := Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	llmMsg := msg.ToLLMMessage()
	if llmMsg.Role != domain.RoleUser {
		t.Errorf("Expected role user, got %v", llmMsg.Role)
	}
	if len(llmMsg.Content) != 1 {
		t.Fatalf("Expected 1 content part, got %d", len(llmMsg.Content))
	}
	if llmMsg.Content[0].Text != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got %q", llmMsg.Content[0].Text)
	}

	// Test conversion back
	converted := FromLLMMessage(llmMsg)
	if converted.Role != msg.Role {
		t.Errorf("Expected role %q, got %q", msg.Role, converted.Role)
	}
	if converted.Content != msg.Content {
		t.Errorf("Expected content %q, got %q", msg.Content, converted.Content)
	}
}

func TestMessageWithAttachments(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Look at this image",
		Attachments: []Attachment{
			{
				Type:    AttachmentTypeImage,
				Content: "https://example.com/image.jpg",
			},
			{
				Type:    AttachmentTypeText,
				Content: "Additional text",
			},
		},
	}

	llmMsg := msg.ToLLMMessage()
	// Should have text content + 2 attachments = 3 parts
	if len(llmMsg.Content) != 3 {
		t.Fatalf("Expected 3 parts, got %d", len(llmMsg.Content))
	}

	// Check main text
	if llmMsg.Content[0].Type != domain.ContentTypeText {
		t.Errorf("Expected text part first, got %v", llmMsg.Content[0].Type)
	}
	if llmMsg.Content[0].Text != "Look at this image" {
		t.Errorf("Expected main text content, got %q", llmMsg.Content[0].Text)
	}

	// Check image part
	if llmMsg.Content[1].Type != domain.ContentTypeImage {
		t.Errorf("Expected image part, got %v", llmMsg.Content[1].Type)
	}
	if llmMsg.Content[1].Image == nil {
		t.Error("Expected image content to be non-nil")
	}

	// Check text attachment part
	if llmMsg.Content[2].Type != domain.ContentTypeText {
		t.Errorf("Expected text part, got %v", llmMsg.Content[2].Type)
	}
	if llmMsg.Content[2].Text != "Additional text" {
		t.Errorf("Expected text attachment content, got %q", llmMsg.Content[2].Text)
	}

	// Test conversion back
	converted := FromLLMMessage(llmMsg)
	if converted.Content != msg.Content {
		t.Errorf("Expected content %q, got %q", msg.Content, converted.Content)
	}
	if len(converted.Attachments) != 2 {
		t.Fatalf("Expected 2 attachments after conversion, got %d", len(converted.Attachments))
	}
}

func TestAttachmentConversion(t *testing.T) {
	tests := []struct {
		name         string
		attachment   Attachment
		contentType  domain.ContentType
		checkContent func(*testing.T, *domain.ContentPart)
	}{
		{
			name: "Image attachment",
			attachment: Attachment{
				Type:    AttachmentTypeImage,
				Content: "https://example.com/image.jpg",
			},
			contentType: domain.ContentTypeImage,
			checkContent: func(t *testing.T, part *domain.ContentPart) {
				if part.Image == nil {
					t.Error("Expected image content")
				}
				if part.Image.Source.URL != "https://example.com/image.jpg" {
					t.Errorf("Expected image URL, got %q", part.Image.Source.URL)
				}
			},
		},
		{
			name: "Text attachment",
			attachment: Attachment{
				Type:    AttachmentTypeText,
				Content: "Some text content",
			},
			contentType: domain.ContentTypeText,
			checkContent: func(t *testing.T, part *domain.ContentPart) {
				if part.Text != "Some text content" {
					t.Errorf("Expected text content, got %q", part.Text)
				}
			},
		},
		{
			name: "File attachment",
			attachment: Attachment{
				Type:     AttachmentTypeFile,
				FilePath: "/path/to/file.pdf",
				Content:  "base64content",
				MimeType: "application/pdf",
			},
			contentType: domain.ContentTypeFile,
			checkContent: func(t *testing.T, part *domain.ContentPart) {
				if part.File == nil {
					t.Error("Expected file content")
				}
				if part.File.FileName != "/path/to/file.pdf" {
					t.Errorf("Expected file path, got %q", part.File.FileName)
				}
				if part.File.MimeType != "application/pdf" {
					t.Errorf("Expected mime type, got %q", part.File.MimeType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := tt.attachment.ToLLMContentPart()
			if part == nil {
				t.Fatal("Expected non-nil content part")
			}
			if part.Type != tt.contentType {
				t.Errorf("Expected content type %v, got %v", tt.contentType, part.Type)
			}
			tt.checkContent(t, part)
		})
	}
}

func TestPromptParams(t *testing.T) {
	temp := 0.7
	maxTokens := 100
	topP := 0.9

	params := PromptParams{
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		TopP:        &topP,
		Stop:        []string{".", "!"},
		CustomOptions: map[string]interface{}{
			"custom_key": "custom_value",
		},
	}

	// Test that all fields are set correctly
	if *params.Temperature != temp {
		t.Errorf("Expected temperature %v, got %v", temp, *params.Temperature)
	}
	if *params.MaxTokens != maxTokens {
		t.Errorf("Expected max tokens %v, got %v", maxTokens, *params.MaxTokens)
	}
	if *params.TopP != topP {
		t.Errorf("Expected top_p %v, got %v", topP, *params.TopP)
	}
	if len(params.Stop) != 2 {
		t.Errorf("Expected 2 stop tokens, got %d", len(params.Stop))
	}
	if params.CustomOptions["custom_key"] != "custom_value" {
		t.Errorf("Expected custom option value 'custom_value', got %v", params.CustomOptions["custom_key"])
	}
}

func TestRequest(t *testing.T) {
	temp := 0.8
	maxTokens := 1000

	req := Request{
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
		Model:        "openai/gpt-4",
		Temperature:  &temp,
		MaxTokens:    &maxTokens,
		Stream:       true,
		SystemPrompt: "Be concise.",
		Options: &PromptParams{
			Temperature: &temp,
			MaxTokens:   &maxTokens,
		},
	}

	if len(req.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(req.Messages))
	}
	if req.Model != "openai/gpt-4" {
		t.Errorf("Expected model 'openai/gpt-4', got %q", req.Model)
	}
	if !req.Stream {
		t.Error("Expected stream to be true")
	}
}

func TestResponse(t *testing.T) {
	resp := Response{
		Content: "Hello! How can I help you?",
		Model:   "gpt-4",
		Usage: &Usage{
			InputTokens:  10,
			OutputTokens: 6,
			TotalTokens:  16,
		},
		Metadata: map[string]interface{}{
			"request_id": "123456",
		},
		FinishReason: "stop",
	}

	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("Unexpected content: %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 16 {
		t.Errorf("Expected total tokens 16, got %d", resp.Usage.TotalTokens)
	}
	if resp.Metadata["request_id"] != "123456" {
		t.Errorf("Expected request_id '123456', got %v", resp.Metadata["request_id"])
	}
}

func TestMultimodalMessageRoundtrip(t *testing.T) {
	original := Message{
		Role:    "user",
		Content: "Here's a document with an image",
		Attachments: []Attachment{
			{
				Type:     AttachmentTypeFile,
				FilePath: "document.pdf",
				Content:  "base64pdfcontent",
				MimeType: "application/pdf",
			},
			{
				Type:     AttachmentTypeImage,
				Content:  "https://example.com/chart.png",
				MimeType: "image/png",
			},
		},
	}

	// Convert to go-llms format
	llmMsg := original.ToLLMMessage()

	// Convert back to our format
	converted := FromLLMMessage(llmMsg)

	// Verify the content matches
	if converted.Role != original.Role {
		t.Errorf("Role mismatch: %q != %q", converted.Role, original.Role)
	}
	if converted.Content != original.Content {
		t.Errorf("Content mismatch: %q != %q", converted.Content, original.Content)
	}
	if len(converted.Attachments) != len(original.Attachments) {
		t.Fatalf("Attachment count mismatch: %d != %d", len(converted.Attachments), len(original.Attachments))
	}

	// Check attachments
	for i, att := range converted.Attachments {
		origAtt := original.Attachments[i]
		if att.Type != origAtt.Type {
			t.Errorf("Attachment %d type mismatch: %v != %v", i, att.Type, origAtt.Type)
		}
		if att.MimeType != origAtt.MimeType {
			t.Errorf("Attachment %d mime type mismatch: %q != %q", i, att.MimeType, origAtt.MimeType)
		}
	}
}

// TODO: Fix this test after ModelCapabilities refactoring
func TestModelInfo(t *testing.T) {
	model := ModelInfo{
		Provider: "openai",
		Model:    "gpt-4-vision",
		Capabilities: ModelCapabilities{
			Text:  true,
			Image: true,
		},
		MaxTokens:   128000,
		Description: "GPT-4 with vision capabilities",
	}

	// Test capability check
	hasText := false
	hasImage := false
	hasAudio := false

	hasText = model.Capabilities.Text
	hasImage = model.Capabilities.Image
	hasAudio = model.Capabilities.Audio

	if !hasText {
		t.Error("Expected text capability")
	}
	if !hasImage {
		t.Error("Expected image capability")
	}
	if hasAudio {
		t.Error("Did not expect audio capability")
	}

	if model.MaxTokens != 128000 {
		t.Errorf("Expected max tokens 128000, got %d", model.MaxTokens)
	}
}
