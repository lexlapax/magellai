// ABOUTME: Unit tests for the high-level Ask function
// ABOUTME: Tests prompt handling, provider selection, and response formatting
package magellai

import (
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/llm"
)

func TestAsk(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		prompt  string
		options *AskOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty prompt",
			prompt:  "",
			options: nil,
			wantErr: true,
			errMsg:  "prompt cannot be empty",
		},
		{
			name:    "Simple prompt with default options",
			prompt:  "Hello, world!",
			options: nil,
			wantErr: false,
		},
		{
			name:   "Prompt with system message",
			prompt: "What is 2+2?",
			options: &AskOptions{
				SystemPrompt: "You are a math tutor.",
			},
			wantErr: false,
		},
		{
			name:   "Prompt with temperature",
			prompt: "Tell me a story",
			options: &AskOptions{
				Temperature: floatPtr(0.8),
			},
			wantErr: false,
		},
		{
			name:   "Prompt with max tokens",
			prompt: "Describe the weather",
			options: &AskOptions{
				MaxTokens: intPtr(100),
			},
			wantErr: false,
		},
		{
			name:   "Invalid provider",
			prompt: "Test",
			options: &AskOptions{
				Model: "invalid/model",
			},
			wantErr: true,
			errMsg:  "API key not provided for invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock provider for testing
			if tt.options == nil {
				tt.options = &AskOptions{}
			}
			if !strings.Contains(tt.options.Model, "invalid") {
				tt.options.Model = "mock/test-model"
			}

			result, err := Ask(ctx, tt.prompt, tt.options)

			if (err != nil) != tt.wantErr {
				t.Errorf("Ask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Ask() error = %v, want error containing %q", err, tt.errMsg)
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("Expected non-nil result")
				} else if result.Content == "" {
					t.Error("Expected non-empty content")
				}
			}
		})
	}
}

func TestAskWithAttachments(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		prompt      string
		attachments []llm.Attachment
		options     *AskOptions
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Empty prompt and attachments",
			prompt:      "",
			attachments: nil,
			options:     nil,
			wantErr:     true,
			errMsg:      "prompt and attachments cannot both be empty",
		},
		{
			name:   "Prompt with image attachment",
			prompt: "What's in this image?",
			attachments: []llm.Attachment{
				{
					Type:    llm.AttachmentTypeImage,
					Content: "https://example.com/image.jpg",
				},
			},
			options: &AskOptions{
				Model: "mock/test-model",
			},
			wantErr: false,
		},
		{
			name:   "Multiple attachments",
			prompt: "Analyze these files",
			attachments: []llm.Attachment{
				{
					Type:    llm.AttachmentTypeText,
					Content: "Document content",
				},
				{
					Type:     llm.AttachmentTypeFile,
					FilePath: "document.pdf",
					MimeType: "application/pdf",
				},
			},
			options: &AskOptions{
				Model: "mock/test-model",
			},
			wantErr: false,
		},
		{
			name:   "Only attachments without prompt",
			prompt: "",
			attachments: []llm.Attachment{
				{
					Type:    llm.AttachmentTypeImage,
					Content: "https://example.com/image.jpg",
				},
			},
			options: &AskOptions{
				Model: "mock/test-model",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AskWithAttachments(ctx, tt.prompt, tt.attachments, tt.options)

			if (err != nil) != tt.wantErr {
				t.Errorf("AskWithAttachments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("AskWithAttachments() error = %v, want error containing %q", err, tt.errMsg)
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("Expected non-nil result")
				} else if result.Content == "" {
					t.Error("Expected non-empty content")
				}
			}
		})
	}
}

func TestParseModel(t *testing.T) {
	tests := []struct {
		modelStr string
		provider string
		model    string
	}{
		{"", llm.ProviderOpenAI, "gpt-3.5-turbo"},
		{"openai/gpt-4", "openai", "gpt-4"},
		{"anthropic/claude-3", "anthropic", "claude-3"},
		{"gemini/gemini-pro", "gemini", "gemini-pro"},
		{"gpt-4", "openai", "gpt-4"}, // Default to OpenAI
	}

	for _, tt := range tests {
		t.Run(tt.modelStr, func(t *testing.T) {
			provider, model := parseModel(tt.modelStr)
			if provider != tt.provider {
				t.Errorf("parseModel(%q) provider = %q, want %q", tt.modelStr, provider, tt.provider)
			}
			if model != tt.model {
				t.Errorf("parseModel(%q) model = %q, want %q", tt.modelStr, model, tt.model)
			}
		})
	}
}

func TestBuildProviderOptions(t *testing.T) {
	temp := 0.7
	maxTokens := 1000

	tests := []struct {
		name     string
		options  *AskOptions
		expected int // Expected number of options
	}{
		{
			name:     "No options",
			options:  &AskOptions{},
			expected: 0,
		},
		{
			name: "Temperature only",
			options: &AskOptions{
				Temperature: &temp,
			},
			expected: 1,
		},
		{
			name: "Multiple options",
			options: &AskOptions{
				Temperature:    &temp,
				MaxTokens:      &maxTokens,
				ResponseFormat: "json_object",
			},
			expected: 3,
		},
		{
			name: "With provider options",
			options: &AskOptions{
				Temperature: &temp,
				ProviderOptions: []llm.ProviderOption{
					llm.WithTopP(0.9),
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildProviderOptions(tt.options)
			if len(opts) != tt.expected {
				t.Errorf("buildProviderOptions() returned %d options, want %d", len(opts), tt.expected)
			}
		})
	}
}

func TestAskStreaming(t *testing.T) {
	ctx := context.Background()

	// Test streaming functionality
	options := &AskOptions{
		Model:  "mock/test-model",
		Stream: true,
	}

	result, err := Ask(ctx, "Test streaming", options)
	if err != nil {
		t.Fatalf("Ask() with streaming failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Content == "" {
		t.Error("Expected non-empty content from streaming")
	}

	if result.FinishReason == "" {
		t.Error("Expected finish reason from streaming")
	}
}

func TestAskWithSystemPrompt(t *testing.T) {
	ctx := context.Background()

	options := &AskOptions{
		Model:        "mock/test-model",
		SystemPrompt: "You are a helpful assistant",
	}

	result, err := Ask(ctx, "Hello", options)
	if err != nil {
		t.Fatalf("Ask() with system prompt failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// The mock provider should return something
	if result.Content == "" {
		t.Error("Expected non-empty content")
	}
}

func TestAskResult(t *testing.T) {
	// Test that AskResult properly contains all expected fields
	ctx := context.Background()

	options := &AskOptions{
		Model: "mock/test-model",
	}

	result, err := Ask(ctx, "Test", options)
	if err != nil {
		t.Fatalf("Ask() failed: %v", err)
	}

	// Check all fields are populated correctly
	if result.Provider != "mock" {
		t.Errorf("Expected provider 'mock', got %q", result.Provider)
	}

	if result.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %q", result.Model)
	}

	if result.Content == "" {
		t.Error("Expected non-empty content")
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
