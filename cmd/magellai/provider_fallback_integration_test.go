// ABOUTME: Integration tests for provider fallback scenarios
// ABOUTME: Tests the behavior of the system when providers fail and fallback to alternates

package main

import (
	"context"
	"errors"
	"testing"
	"time"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/stretchr/testify/assert"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// mockProvider is a test provider that can simulate various conditions
type mockProvider struct {
	responses  []string
	callCount  int
	shouldFail bool
	delay      time.Duration
}

func (m *mockProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	m.callCount++
	if m.shouldFail {
		return "", errors.New("provider failed")
	}

	// Simulate delay if configured
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(m.delay):
			// Delay completed
		}
	}

	if len(m.responses) > 0 {
		idx := (m.callCount - 1) % len(m.responses)
		return m.responses[idx], nil
	}
	return "mock response", nil
}

func (m *mockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	m.callCount++
	if m.shouldFail {
		return nil, errors.New("provider failed")
	}
	return &llm.Response{Content: "mock response", Model: "mock-model"}, nil
}

func (m *mockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	m.callCount++
	if m.shouldFail {
		return nil, errors.New("provider failed")
	}
	return map[string]string{"result": "mock"}, nil
}

func (m *mockProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	m.callCount++
	if m.shouldFail {
		return nil, errors.New("provider failed")
	}
	ch := make(chan llm.StreamChunk, 2)
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{Content: "mock "}
		ch <- llm.StreamChunk{Content: "stream"}
	}()
	return ch, nil
}

func (m *mockProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	return m.Stream(ctx, "", options...)
}

func (m *mockProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{
		Provider:    "mock",
		Model:       "mock-model",
		DisplayName: "Mock Model",
		Description: "A mock model for testing",
		Capabilities: llm.ModelCapabilities{
			Text: true,
		},
	}
}

// MockFailingProvider is a provider that fails after a certain number of calls
type MockFailingProvider struct {
	failAfter int
	calls     int
}

func NewMockFailingProvider(failAfter int) *MockFailingProvider {
	return &MockFailingProvider{
		failAfter: failAfter,
	}
}

func (m *MockFailingProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	m.calls++
	if m.calls > m.failAfter {
		return "", errors.New("provider failed")
	}
	return "Mock response", nil
}

func (m *MockFailingProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	m.calls++
	if m.calls > m.failAfter {
		return nil, errors.New("provider failed")
	}
	return &llm.Response{
		Content: "Mock response",
		Model:   "mock-model",
	}, nil
}

func (m *MockFailingProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	m.calls++
	if m.calls > m.failAfter {
		return nil, errors.New("provider failed")
	}
	return "Mock schema response", nil
}

func (m *MockFailingProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	m.calls++
	if m.calls > m.failAfter {
		return nil, errors.New("provider failed")
	}
	ch := make(chan llm.StreamChunk, 2)
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{Content: "Mock "}
		ch <- llm.StreamChunk{Content: "streaming response"}
	}()
	return ch, nil
}

func (m *MockFailingProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	return m.Stream(ctx, "", options...)
}

func (m *MockFailingProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{
		Provider:    "mock",
		Model:       "mock-model",
		DisplayName: "Mock Model",
	}
}

func TestProviderFallback_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// These tests use mock providers so no real config is needed

	t.Run("SingleProviderFailure", func(t *testing.T) {
		// Test fallback from primary to secondary provider

		// Create mock providers
		primary := NewMockFailingProvider(1) // Fail after 1 call
		secondary := &mockProvider{}

		// Create resilient provider with chain
		resilientConfig := llm.ResilientProviderConfig{
			Primary:        primary,
			Fallbacks:      []llm.Provider{secondary},
			EnableFallback: true,
		}
		resilient := llm.NewResilientProvider(resilientConfig)

		// First call should succeed
		ctx := context.Background()
		resp1, err := resilient.Generate(ctx, "Test prompt 1")
		assert.NoError(t, err)
		assert.NotEmpty(t, resp1)

		// Second call should trigger fallback to secondary
		resp2, err := resilient.Generate(ctx, "Test prompt 2")
		assert.NoError(t, err)
		assert.NotEmpty(t, resp2)

		// Verify the primary provider failed
		assert.Greater(t, primary.calls, primary.failAfter)
	})

	t.Run("MultipleProviderFailures", func(t *testing.T) {
		// Test multiple providers failing in sequence

		primary := NewMockFailingProvider(0)   // Fail immediately
		secondary := NewMockFailingProvider(0) // Also fail immediately
		tertiary := &mockProvider{}            // This one succeeds

		// Create resilient provider with full chain
		resilientConfig := llm.ResilientProviderConfig{
			Primary:        primary,
			Fallbacks:      []llm.Provider{secondary, tertiary},
			EnableFallback: true,
			RetryConfig: llm.RetryConfig{
				MaxRetries: 0, // Disable retries for predictable test behavior
			},
		}
		resilient := llm.NewResilientProvider(resilientConfig)

		ctx := context.Background()

		// Should fall back to tertiary
		resp, err := resilient.Generate(ctx, "Test prompt")
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)

		// Verify providers were called
		assert.Equal(t, 1, primary.calls)
		assert.Equal(t, 1, secondary.calls)
		assert.Equal(t, 1, tertiary.callCount)
	})

	t.Run("StreamingFallback", func(t *testing.T) {
		// Test that streaming doesn't use fallback (current implementation limitation)

		primary := NewMockFailingProvider(0) // Fail immediately
		secondary := &mockProvider{}

		resilientConfig := llm.ResilientProviderConfig{
			Primary:        primary,
			Fallbacks:      []llm.Provider{secondary},
			EnableFallback: true,
		}
		resilient := llm.NewResilientProvider(resilientConfig)

		ctx := context.Background()

		// Streaming call should fail since primary fails and fallback isn't implemented for streaming
		stream, err := resilient.Stream(ctx, "Test prompt")
		assert.Error(t, err) // Expect error since streaming doesn't support fallback
		assert.Nil(t, stream)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// Test proper handling of context cancellation

		// Create a slow provider that respects context
		slowProvider := &mockProvider{
			shouldFail: false,
			delay:      100 * time.Millisecond, // Longer than context timeout
		}

		resilientConfig := llm.ResilientProviderConfig{
			Primary:        slowProvider,
			EnableFallback: false,
			Timeout:        50 * time.Millisecond,
		}
		resilient := llm.NewResilientProvider(resilientConfig)

		// Create a context with shorter timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Should fail due to context cancellation
		_, err := resilient.Generate(ctx, "Test prompt")
		assert.Error(t, err)
	})

	t.Run("AllProvidersFail", func(t *testing.T) {
		// Test behavior when all providers fail

		primary := NewMockFailingProvider(0)   // Fail immediately
		secondary := NewMockFailingProvider(0) // Fail immediately

		resilientConfig := llm.ResilientProviderConfig{
			Primary:        primary,
			Fallbacks:      []llm.Provider{secondary},
			EnableFallback: true,
		}
		resilient := llm.NewResilientProvider(resilientConfig)

		ctx := context.Background()

		// Should fail when all providers are exhausted
		_, err := resilient.Generate(ctx, "Test prompt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed")
	})
}
