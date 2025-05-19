package llm

import (
	"context"
	"strings"
	"testing"
	"time"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/pkg/domain"
)

// mockProvider is a test provider that can simulate various error conditions
type mockProvider struct {
	generateFunc        func(context.Context, string, ...ProviderOption) (string, error)
	generateMessageFunc func(context.Context, []domain.Message, ...ProviderOption) (*Response, error)
	streamFunc          func(context.Context, string, ...ProviderOption) (<-chan StreamChunk, error)
	modelInfo           ModelInfo
	callCount           int
}

func (m *mockProvider) Generate(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
	m.callCount++
	if m.generateFunc != nil {
		return m.generateFunc(ctx, prompt, options...)
	}
	return "mock response", nil
}

func (m *mockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (*Response, error) {
	m.callCount++
	if m.generateMessageFunc != nil {
		return m.generateMessageFunc(ctx, messages, options...)
	}
	return &Response{Content: "mock response", Model: "mock-model"}, nil
}

func (m *mockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...ProviderOption) (interface{}, error) {
	m.callCount++
	return map[string]string{"result": "mock"}, nil
}

func (m *mockProvider) Stream(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
	m.callCount++
	if m.streamFunc != nil {
		return m.streamFunc(ctx, prompt, options...)
	}
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: "mock"}
	close(ch)
	return ch, nil
}

func (m *mockProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (<-chan StreamChunk, error) {
	m.callCount++
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: "mock"}
	close(ch)
	return ch, nil
}

func (m *mockProvider) GetModelInfo() ModelInfo {
	return m.modelInfo
}

func TestResilientProvider_Generate(t *testing.T) {
	tests := []struct {
		name         string
		primaryFunc  func(context.Context, string, ...ProviderOption) (string, error)
		fallbackFunc func(context.Context, string, ...ProviderOption) (string, error)
		expectError  bool
		expectResult string
	}{
		{
			name: "primary succeeds",
			primaryFunc: func(ctx context.Context, prompt string, opts ...ProviderOption) (string, error) {
				return "primary response", nil
			},
			expectError:  false,
			expectResult: "primary response",
		},
		{
			name: "primary fails with retryable error, succeeds on retry",
			primaryFunc: func() func(context.Context, string, ...ProviderOption) (string, error) {
				attempts := 0
				return func(ctx context.Context, prompt string, opts ...ProviderOption) (string, error) {
					attempts++
					if attempts == 1 {
						return "", llmdomain.ErrNetworkConnectivity
					}
					return "primary response after retry", nil
				}
			}(),
			expectError:  false,
			expectResult: "primary response after retry",
		},
		{
			name: "primary fails, fallback succeeds",
			primaryFunc: func(ctx context.Context, prompt string, opts ...ProviderOption) (string, error) {
				return "", llmdomain.ErrProviderUnavailable
			},
			fallbackFunc: func(ctx context.Context, prompt string, opts ...ProviderOption) (string, error) {
				return "fallback response", nil
			},
			expectError:  false,
			expectResult: "fallback response",
		},
		{
			name: "all providers fail",
			primaryFunc: func(ctx context.Context, prompt string, opts ...ProviderOption) (string, error) {
				return "", llmdomain.ErrProviderUnavailable
			},
			fallbackFunc: func(ctx context.Context, prompt string, opts ...ProviderOption) (string, error) {
				return "", llmdomain.ErrProviderUnavailable
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := &mockProvider{
				generateFunc: tt.primaryFunc,
				modelInfo:    ModelInfo{Provider: "primary", Model: "test"},
			}

			var fallbacks []Provider
			if tt.fallbackFunc != nil {
				fallback := &mockProvider{
					generateFunc: tt.fallbackFunc,
					modelInfo:    ModelInfo{Provider: "fallback", Model: "test"},
				}
				fallbacks = append(fallbacks, fallback)
			}

			config := ResilientProviderConfig{
				Primary:        primary,
				Fallbacks:      fallbacks,
				RetryConfig:    DefaultRetryConfig(),
				EnableFallback: true,
				Timeout:        5 * time.Second,
			}
			config.RetryConfig.InitialDelay = 10 * time.Millisecond // Speed up tests
			config.RetryConfig.MaxDelay = 100 * time.Millisecond

			resilient := NewResilientProvider(config)

			ctx := context.Background()
			result, err := resilient.Generate(ctx, "test prompt")

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.expectResult {
				t.Errorf("expected result %q, got %q", tt.expectResult, result)
			}
		})
	}
}

func TestResilientProvider_ContextTooLong(t *testing.T) {
	attemptCount := 0
	primary := &mockProvider{
		generateMessageFunc: func(ctx context.Context, messages []domain.Message, opts ...ProviderOption) (*Response, error) {
			attemptCount++
			if attemptCount == 1 {
				// First attempt fails with context too long
				return nil, llmdomain.ErrContextTooLong
			}
			// Second attempt with reduced messages succeeds
			if len(messages) < 3 {
				return &Response{Content: "success with reduced context"}, nil
			}
			return nil, llmdomain.ErrContextTooLong
		},
	}

	config := ResilientProviderConfig{
		Primary:        primary,
		RetryConfig:    DefaultRetryConfig(),
		EnableFallback: false,
		Timeout:        5 * time.Second,
	}

	resilient := NewResilientProvider(config)

	// Create messages that will trigger context reduction
	messages := []domain.Message{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "Message 1"},
		{Role: "assistant", Content: "Response 1"},
		{Role: "user", Content: "Message 2"},
	}

	ctx := context.Background()
	response, err := resilient.GenerateMessage(ctx, messages)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if response == nil || response.Content != "success with reduced context" {
		t.Error("expected successful response with reduced context")
	}

	if attemptCount != 2 {
		t.Errorf("expected 2 attempts, got %d", attemptCount)
	}
}

func TestResilientProvider_RateLimit(t *testing.T) {
	attemptCount := 0
	primary := &mockProvider{
		generateFunc: func(ctx context.Context, prompt string, opts ...ProviderOption) (string, error) {
			attemptCount++
			if attemptCount < 3 {
				return "", llmdomain.ErrRateLimitExceeded
			}
			return "success after rate limit", nil
		},
		modelInfo: ModelInfo{Provider: "test", Model: "test-model"},
	}

	config := ResilientProviderConfig{
		Primary:        primary,
		RetryConfig:    DefaultRetryConfig(),
		EnableFallback: false,
		Timeout:        30 * time.Second,
	}

	resilient := NewResilientProvider(config)

	ctx := context.Background()

	// Rate limit errors are not retryable by default in our implementation
	// This should fail since we don't retry rate limits
	result, err := resilient.Generate(ctx, "test")

	if err == nil {
		t.Errorf("expected rate limit error but got none")
	}

	if result != "" {
		t.Errorf("expected empty result but got: %s", result)
	}

	// Should fail fast since rate limits aren't retried
	if attemptCount != 1 {
		t.Errorf("expected 1 attempt but got %d", attemptCount)
	}
}

func TestCreateProviderChain(t *testing.T) {
	configs := []ChainProviderConfig{
		{Type: ProviderMock, Model: "mock-1"},
		{Type: ProviderMock, Model: "mock-2"},
	}

	resilient, err := CreateProviderChain(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resilient == nil {
		t.Fatal("expected non-nil resilient provider")
	}

	modelInfo := resilient.GetModelInfo()
	if modelInfo.Model != "mock-1" {
		t.Errorf("expected primary model to be mock-1, got %s", modelInfo.Model)
	}
}

func TestTruncateContext(t *testing.T) {
	tests := []struct {
		name         string
		messages     []domain.Message
		maxTokens    int
		expectLength int
		checkSystem  bool
	}{
		{
			name: "keep all if under limit",
			messages: []domain.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi"},
			},
			maxTokens:    1000,
			expectLength: 2,
		},
		{
			name: "preserve system message",
			messages: []domain.Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "1"},
				{Role: "assistant", Content: "2"},
				{Role: "user", Content: "3"},
				{Role: "assistant", Content: "4"},
				{Role: "user", Content: "5"},
			},
			maxTokens:    100,
			expectLength: 4, // System + last 3
			checkSystem:  true,
		},
		{
			name: "no system message",
			messages: []domain.Message{
				{Role: "user", Content: "1"},
				{Role: "assistant", Content: "2"},
				{Role: "user", Content: "3"},
				{Role: "assistant", Content: "4"},
				{Role: "user", Content: "5"},
			},
			maxTokens:    100,
			expectLength: 3, // Last 3 messages
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateContext(tt.messages, tt.maxTokens)

			if len(result) != tt.expectLength {
				t.Errorf("expected %d messages, got %d", tt.expectLength, len(result))
			}

			if tt.checkSystem && len(result) > 0 {
				if strings.ToLower(string(result[0].Role)) != "system" {
					t.Error("expected first message to be system message")
				}
			}
		})
	}
}
