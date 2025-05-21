// ABOUTME: Simple test to verify provider fallback mechanism works
// ABOUTME: Basic demonstration that the system can handle provider failures

//go:build integration
// +build integration

package main

import (
	"context"
	"errors"
	"testing"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/stretchr/testify/assert"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// SimpleProvider is a minimal provider implementation
type SimpleProvider struct {
	name       string
	shouldFail bool
	calls      int
}

func (p *SimpleProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	p.calls++
	if p.shouldFail {
		return "", errors.New(p.name + " failed")
	}
	return p.name + " response", nil
}

func (p *SimpleProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	p.calls++
	if p.shouldFail {
		return nil, errors.New(p.name + " failed")
	}
	return &llm.Response{Content: p.name + " response", Model: "test"}, nil
}

func (p *SimpleProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	p.calls++
	if p.shouldFail {
		return nil, errors.New(p.name + " failed")
	}
	return p.name + " schema response", nil
}

func (p *SimpleProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	p.calls++
	if p.shouldFail {
		return nil, errors.New(p.name + " failed")
	}
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{Content: p.name + " stream response", Done: true}
	}()
	return ch, nil
}

func (p *SimpleProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	return p.Stream(ctx, "", options...)
}

func (p *SimpleProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{
		Provider: p.name,
		Model:    "test",
	}
}

func TestSimpleProviderFallback(t *testing.T) {
	t.Run("BasicFallback", func(t *testing.T) {
		// Create two providers - first fails, second succeeds
		primary := &SimpleProvider{name: "primary", shouldFail: true}
		secondary := &SimpleProvider{name: "secondary", shouldFail: false}

		// Configure resilient provider
		config := llm.ResilientProviderConfig{
			Primary:        primary,
			Fallbacks:      []llm.Provider{secondary},
			EnableFallback: true,
		}

		resilient := llm.NewResilientProvider(config)

		// Make a request
		ctx := context.Background()
		response, err := resilient.Generate(ctx, "test prompt")

		// Verify fallback worked
		assert.NoError(t, err)
		assert.Equal(t, "secondary response", response)

		// Verify both providers were called
		assert.Equal(t, 1, primary.calls)
		assert.Equal(t, 1, secondary.calls) // Fixed: now only called once
	})

	t.Run("AllProvidersFail", func(t *testing.T) {
		// Create providers that all fail
		primary := &SimpleProvider{name: "primary", shouldFail: true}
		fallback1 := &SimpleProvider{name: "fallback1", shouldFail: true}
		fallback2 := &SimpleProvider{name: "fallback2", shouldFail: true}

		// Configure resilient provider
		config := llm.ResilientProviderConfig{
			Primary:        primary,
			Fallbacks:      []llm.Provider{fallback1, fallback2},
			EnableFallback: true,
		}

		resilient := llm.NewResilientProvider(config)

		// Make a request
		ctx := context.Background()
		_, err := resilient.Generate(ctx, "test prompt")

		// Verify error is returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "all providers failed")

		// Verify all providers were attempted
		assert.Equal(t, 1, primary.calls)
		assert.Equal(t, 1, fallback1.calls)
		assert.Equal(t, 1, fallback2.calls)
	})

	t.Run("FallbackDisabled", func(t *testing.T) {
		// Create providers where primary fails
		primary := &SimpleProvider{name: "primary", shouldFail: true}
		secondary := &SimpleProvider{name: "secondary", shouldFail: false}

		// Configure resilient provider with fallback disabled
		config := llm.ResilientProviderConfig{
			Primary:        primary,
			Fallbacks:      []llm.Provider{secondary},
			EnableFallback: false,
		}

		resilient := llm.NewResilientProvider(config)

		// Make a request
		ctx := context.Background()
		_, err := resilient.Generate(ctx, "test prompt")

		// Verify error is returned (no fallback attempted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "primary failed")

		// Verify only primary was called
		assert.Equal(t, 1, primary.calls)
		assert.Equal(t, 0, secondary.calls)
	})
}
