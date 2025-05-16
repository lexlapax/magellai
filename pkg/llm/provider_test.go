// ABOUTME: Unit tests for the provider adapter interface
// ABOUTME: Tests provider creation, option handling, and message conversion
package llm

import (
	"context"
	"os"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name         string
		providerType string
		model        string
		apiKey       string
		envKey       string
		envValue     string
		wantErr      bool
	}{
		{
			name:         "OpenAI with API key",
			providerType: ProviderOpenAI,
			model:        "gpt-3.5-turbo",
			apiKey:       "test-key",
			wantErr:      false, // Provider should be created successfully
		},
		{
			name:         "Anthropic with env key",
			providerType: ProviderAnthropic,
			model:        "claude-3-sonnet",
			envKey:       "ANTHROPIC_API_KEY",
			envValue:     "test-env-key",
			wantErr:      false, // Provider should be created successfully
		},
		{
			name:         "Mock provider",
			providerType: ProviderMock,
			model:        "mock-model",
			wantErr:      false, // Mock provider doesn't need API key
		},
		{
			name:         "Invalid provider",
			providerType: "invalid",
			model:        "model",
			apiKey:       "key",
			wantErr:      true,
		},
		{
			name:         "Missing API key",
			providerType: ProviderOpenAI,
			model:        "gpt-4",
			envKey:       "OPENAI_API_KEY",
			envValue:     "", // Clear the env var
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable if needed
			if tt.envKey != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			var p Provider
			var err error

			if tt.apiKey != "" {
				p, err = NewProvider(tt.providerType, tt.model, tt.apiKey)
			} else {
				p, err = NewProvider(tt.providerType, tt.model)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && p != nil {
				// Test that the provider was created successfully
				info := p.GetModelInfo()
				if info.Provider != tt.providerType {
					t.Errorf("Expected provider %s, got %s", tt.providerType, info.Provider)
				}
				if info.Model != tt.model {
					t.Errorf("Expected model %s, got %s", tt.model, info.Model)
				}
			}
		})
	}
}

func TestGetAPIKeyFromEnv(t *testing.T) {
	tests := []struct {
		provider string
		envKey   string
		envValue string
		expected string
	}{
		{
			provider: ProviderOpenAI,
			envKey:   "OPENAI_API_KEY",
			envValue: "test-openai-key",
			expected: "test-openai-key",
		},
		{
			provider: ProviderAnthropic,
			envKey:   "ANTHROPIC_API_KEY",
			envValue: "test-anthropic-key",
			expected: "test-anthropic-key",
		},
		{
			provider: ProviderGemini,
			envKey:   "GEMINI_API_KEY",
			envValue: "test-gemini-key",
			expected: "test-gemini-key",
		},
		{
			provider: "unknown",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			got := getAPIKeyFromEnv(tt.provider)
			if got != tt.expected {
				t.Errorf("getAPIKeyFromEnv(%s) = %q, want %q", tt.provider, got, tt.expected)
			}
		})
	}
}

func TestGetModelCapabilities(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		expected []ModelCapability
	}{
		{
			provider: ProviderOpenAI,
			model:    "gpt-4",
			expected: []ModelCapability{CapabilityText},
		},
		{
			provider: ProviderOpenAI,
			model:    "gpt-4-vision",
			expected: []ModelCapability{CapabilityText, CapabilityImage},
		},
		{
			provider: ProviderGemini,
			model:    "gemini-pro",
			expected: []ModelCapability{CapabilityText, CapabilityImage},
		},
		{
			provider: ProviderGemini,
			model:    "gemini-1.5-pro",
			expected: []ModelCapability{CapabilityText, CapabilityImage, CapabilityVideo, CapabilityAudio},
		},
		{
			provider: ProviderAnthropic,
			model:    "claude-2",
			expected: []ModelCapability{CapabilityText},
		},
		{
			provider: ProviderAnthropic,
			model:    "claude-3-opus",
			expected: []ModelCapability{CapabilityText, CapabilityImage},
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider+"/"+tt.model, func(t *testing.T) {
			got := getModelCapabilities(tt.provider, tt.model)

			if len(got) != len(tt.expected) {
				t.Errorf("getModelCapabilities(%s, %s) returned %d capabilities, want %d",
					tt.provider, tt.model, len(got), len(tt.expected))
			}

			// Check each capability exists
			for _, expectedCap := range tt.expected {
				found := false
				for _, gotCap := range got {
					if gotCap == expectedCap {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected capability %v not found", expectedCap)
				}
			}
		})
	}
}

func TestProviderOptions(t *testing.T) {
	config := &providerConfig{}

	// Test temperature option
	WithTemperature(0.8)(config)
	if config.temperature == nil || *config.temperature != 0.8 {
		t.Errorf("WithTemperature failed, got %v", config.temperature)
	}

	// Test max tokens option
	WithMaxTokens(2048)(config)
	if config.maxTokens == nil || *config.maxTokens != 2048 {
		t.Errorf("WithMaxTokens failed, got %v", config.maxTokens)
	}

	// Test stop sequences option
	stops := []string{"END", "STOP"}
	WithStopSequences(stops)(config)
	if len(config.stopSequences) != 2 {
		t.Errorf("WithStopSequences failed, got %v", config.stopSequences)
	}

	// Test top-p option
	WithTopP(0.9)(config)
	if config.topP == nil || *config.topP != 0.9 {
		t.Errorf("WithTopP failed, got %v", config.topP)
	}

	// Test top-k option
	WithTopK(40)(config)
	if config.topK == nil || *config.topK != 40 {
		t.Errorf("WithTopK failed, got %v", config.topK)
	}

	// Test presence penalty option
	WithPresencePenalty(0.1)(config)
	if config.presencePenalty == nil || *config.presencePenalty != 0.1 {
		t.Errorf("WithPresencePenalty failed, got %v", config.presencePenalty)
	}

	// Test frequency penalty option
	WithFrequencyPenalty(0.2)(config)
	if config.frequencyPenalty == nil || *config.frequencyPenalty != 0.2 {
		t.Errorf("WithFrequencyPenalty failed, got %v", config.frequencyPenalty)
	}

	// Test seed option
	WithSeed(42)(config)
	if config.seed == nil || *config.seed != 42 {
		t.Errorf("WithSeed failed, got %v", config.seed)
	}

	// Test response format option
	WithResponseFormat("json_object")(config)
	if config.responseFormat != "json_object" {
		t.Errorf("WithResponseFormat failed, got %v", config.responseFormat)
	}
}

func TestProviderAdapter_buildLLMOptions(t *testing.T) {
	p := &providerAdapter{}

	temp := 0.8
	maxTokens := 1024
	topP := 0.9

	opts := []ProviderOption{
		WithTemperature(temp),
		WithMaxTokens(maxTokens),
		WithStopSequences([]string{"DONE"}),
		WithTopP(topP),
	}

	llmOpts := p.buildLLMOptions(opts...)

	// Apply options to a provider options struct to verify
	providerOpts := domain.DefaultOptions()
	for _, opt := range llmOpts {
		opt(providerOpts)
	}

	if providerOpts.Temperature != temp {
		t.Errorf("Expected temperature %f, got %f", temp, providerOpts.Temperature)
	}
	if providerOpts.MaxTokens != maxTokens {
		t.Errorf("Expected max tokens %d, got %d", maxTokens, providerOpts.MaxTokens)
	}
	if len(providerOpts.StopSequences) != 1 || providerOpts.StopSequences[0] != "DONE" {
		t.Errorf("Expected stop sequences [DONE], got %v", providerOpts.StopSequences)
	}
	if providerOpts.TopP != topP {
		t.Errorf("Expected top-p %f, got %f", topP, providerOpts.TopP)
	}
}

// Mock tests for methods that require actual provider connections
func TestProviderAdapterMock(t *testing.T) {
	// Create a mock provider
	p, err := NewProvider(ProviderMock, "mock-model")
	if err != nil {
		t.Fatalf("Failed to create mock provider: %v", err)
	}

	ctx := context.Background()

	// Test Generate
	result, err := p.Generate(ctx, "test prompt")
	if err != nil {
		t.Errorf("Generate failed: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty generate result")
	}

	// Test GenerateMessage
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}
	response, err := p.GenerateMessage(ctx, messages)
	if err != nil {
		t.Errorf("GenerateMessage failed: %v", err)
	}
	if response == nil || response.Content == "" {
		t.Error("Expected non-empty message response")
	}

	// Test Stream
	streamChan, err := p.Stream(ctx, "test prompt")
	if err != nil {
		t.Errorf("Stream failed: %v", err)
	}

	chunks := 0
	for chunk := range streamChan {
		if chunk.Error != nil {
			t.Errorf("Received error chunk: %v", chunk.Error)
		}
		chunks++
	}
	if chunks == 0 {
		t.Error("Expected at least one stream chunk")
	}

	// Test StreamMessage
	streamMsgChan, err := p.StreamMessage(ctx, messages)
	if err != nil {
		t.Errorf("StreamMessage failed: %v", err)
	}

	msgChunks := 0
	for chunk := range streamMsgChan {
		if chunk.Error != nil {
			t.Errorf("Received error chunk: %v", chunk.Error)
		}
		msgChunks++
	}
	if msgChunks == 0 {
		t.Error("Expected at least one stream message chunk")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"test"}, "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.item, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.expected {
				t.Errorf("contains(%v, %s) = %v, want %v", tt.slice, tt.item, got, tt.expected)
			}
		})
	}
}
