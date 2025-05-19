// ABOUTME: Resilient provider implementation with retry and fallback mechanisms
// ABOUTME: Wraps existing providers with error recovery and graceful degradation

package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/domain"
)

// ResilientProviderConfig configures the resilient provider behavior
type ResilientProviderConfig struct {
	Primary        Provider      // Primary provider to use
	Fallbacks      []Provider    // Fallback providers in order of preference
	RetryConfig    RetryConfig   // Retry configuration
	EnableFallback bool          // Whether to use fallback providers
	Timeout        time.Duration // Timeout for each operation
}

// ResilientProvider wraps providers with retry and fallback logic
type ResilientProvider struct {
	config       ResilientProviderConfig
	errorHandler *ErrorHandler
	logger       *logging.Logger
}

// NewResilientProvider creates a provider with resilience features
func NewResilientProvider(config ResilientProviderConfig) *ResilientProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &ResilientProvider{
		config:       config,
		errorHandler: NewErrorHandler(config.RetryConfig),
		logger:       logging.GetLogger(),
	}
}

// Generate produces text with retry and fallback
func (r *ResilientProvider) Generate(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
	// Create timeout context for the operation
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	operation := "Generate"

	// Try primary provider with retry
	var response string
	var lastErr error
	err := r.errorHandler.WithRetry(ctx, operation, func() error {
		resp, err := r.config.Primary.Generate(ctx, prompt, options...)
		if err != nil {
			lastErr = err
			return err
		}
		response = resp
		return nil
	})

	if err == nil && response != "" {
		// Primary succeeded
		return response, nil
	}

	// Log primary failure
	r.logger.Warn("Primary provider failed, attempting fallbacks",
		"provider", r.config.Primary.GetModelInfo().Provider,
		"error", lastErr)

	// Try fallback providers if enabled
	if r.config.EnableFallback && len(r.config.Fallbacks) > 0 {
		for i, fallback := range r.config.Fallbacks {
			r.logger.Info("Attempting fallback provider",
				"fallback", i+1,
				"provider", fallback.GetModelInfo().Provider)

			// Try fallback with retry
			var fallbackResult string
			err := r.errorHandler.WithRetry(ctx, operation, func() error {
				result, err := fallback.Generate(ctx, prompt, options...)
				if err == nil {
					fallbackResult = result
				}
				return err
			})

			if err == nil {
				// Fallback succeeded
				return fallbackResult, nil
			}

			r.logger.Warn("Fallback provider failed",
				"fallback", i+1,
				"provider", fallback.GetModelInfo().Provider,
				"error", err)
		}
	}

	// All providers failed
	return "", fmt.Errorf("all providers failed for %s: %w", operation, lastErr)
}

// GenerateMessage produces a response with retry and fallback
func (r *ResilientProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	operation := "GenerateMessage"

	// Try primary provider with retry
	var response *Response
	var lastErr error

	err := r.errorHandler.WithRetry(ctx, operation, func() error {
		resp, err := r.config.Primary.GenerateMessage(ctx, messages, options...)
		if err != nil {
			lastErr = err
			// Check for rate limits - use special handling
			if llmdomain.IsRateLimitError(err) {
				return r.errorHandler.WithRateLimitRetry(ctx, operation, func() error {
					resp, err = r.config.Primary.GenerateMessage(ctx, messages, options...)
					response = resp
					return err
				})
			}
			return err
		}
		response = resp
		return nil
	})

	if err == nil && response != nil {
		return response, nil
	}

	// Handle context too long error specially
	if IsContextTooLongError(lastErr) {
		// Try to reduce context by removing older messages
		if len(messages) > 2 {
			r.logger.Info("Context too long, reducing message history")
			reducedMessages := messages[len(messages)-2:] // Keep only last 2 messages
			return r.GenerateMessage(ctx, reducedMessages, options...)
		}
	}

	// Try fallback providers
	if r.config.EnableFallback && len(r.config.Fallbacks) > 0 {
		for i, fallback := range r.config.Fallbacks {
			r.logger.Info("Attempting fallback provider",
				"fallback", i+1,
				"provider", fallback.GetModelInfo().Provider)

			err := r.errorHandler.WithRetry(ctx, operation, func() error {
				resp, err := fallback.GenerateMessage(ctx, messages, options...)
				if err != nil {
					return err
				}
				response = resp
				return nil
			})

			if err == nil && response != nil {
				return response, nil
			}
		}
	}

	return nil, fmt.Errorf("all providers failed for %s: %w", operation, lastErr)
}

// GenerateWithSchema produces structured output with retry and fallback
func (r *ResilientProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...ProviderOption) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	operation := "GenerateWithSchema"

	// Try primary provider
	var result interface{}
	var lastErr error

	err := r.errorHandler.WithRetry(ctx, operation, func() error {
		res, err := r.config.Primary.GenerateWithSchema(ctx, prompt, schema, options...)
		if err != nil {
			lastErr = err
			return err
		}
		result = res
		return nil
	})

	if err == nil && result != nil {
		return result, nil
	}

	// Schema generation is complex - only try fallbacks that support it
	if r.config.EnableFallback && len(r.config.Fallbacks) > 0 {
		for i, fallback := range r.config.Fallbacks {
			// Check if fallback supports structured output
			modelInfo := fallback.GetModelInfo()
			if !modelInfo.Capabilities.StructuredOutput {
				r.logger.Debug("Skipping fallback - no structured output support",
					"provider", modelInfo.Provider)
				continue
			}

			r.logger.Info("Attempting schema fallback",
				"fallback", i+1,
				"provider", modelInfo.Provider)

			err := r.errorHandler.WithRetry(ctx, operation, func() error {
				res, err := fallback.GenerateWithSchema(ctx, prompt, schema, options...)
				if err != nil {
					return err
				}
				result = res
				return nil
			})

			if err == nil && result != nil {
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("all providers failed for %s: %w", operation, lastErr)
}

// Stream streams responses with retry (fallback is tricky for streams)
func (r *ResilientProvider) Stream(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	// Streaming is more complex for fallback - just use primary with retry
	var stream <-chan StreamChunk
	var lastErr error

	err := r.errorHandler.WithRetry(ctx, "Stream", func() error {
		str, err := r.config.Primary.Stream(ctx, prompt, options...)
		if err != nil {
			lastErr = err
			return err
		}
		stream = str
		return nil
	})

	if err == nil && stream != nil {
		// Wrap stream to handle errors during streaming
		return r.wrapStreamWithErrorHandling(stream, ctx), nil
	}

	return nil, fmt.Errorf("streaming failed: %w", lastErr)
}

// StreamMessage streams message responses with retry
func (r *ResilientProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (<-chan StreamChunk, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	var stream <-chan StreamChunk
	var lastErr error

	err := r.errorHandler.WithRetry(ctx, "StreamMessage", func() error {
		str, err := r.config.Primary.StreamMessage(ctx, messages, options...)
		if err != nil {
			lastErr = err
			return err
		}
		stream = str
		return nil
	})

	if err == nil && stream != nil {
		return r.wrapStreamWithErrorHandling(stream, ctx), nil
	}

	return nil, fmt.Errorf("message streaming failed: %w", lastErr)
}

// GetModelInfo returns the primary provider's model info
func (r *ResilientProvider) GetModelInfo() ModelInfo {
	return r.config.Primary.GetModelInfo()
}

// wrapStreamWithErrorHandling adds error recovery to streaming responses
func (r *ResilientProvider) wrapStreamWithErrorHandling(input <-chan StreamChunk, ctx context.Context) <-chan StreamChunk {
	output := make(chan StreamChunk)

	go func() {
		defer close(output)
		defer func() {
			if err := r.errorHandler.RecoverFromPanic("stream wrapper"); err != nil {
				output <- StreamChunk{Error: err}
			}
		}()

		for {
			select {
			case chunk, ok := <-input:
				if !ok {
					return
				}

				// Forward chunk
				select {
				case output <- chunk:
				case <-ctx.Done():
					output <- StreamChunk{Error: ctx.Err()}
					return
				}

			case <-ctx.Done():
				output <- StreamChunk{Error: ctx.Err()}
				return
			}
		}
	}()

	return output
}

// CreateProviderChain creates a chain of providers for fallback
func CreateProviderChain(providers []ChainProviderConfig) (*ResilientProvider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers specified")
	}

	// Create the primary provider
	primary, err := NewProvider(providers[0].Type, providers[0].Model, providers[0].APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary provider: %w", err)
	}

	// Create fallback providers
	var fallbacks []Provider
	for i := 1; i < len(providers); i++ {
		fallback, err := NewProvider(providers[i].Type, providers[i].Model, providers[i].APIKey)
		if err != nil {
			logging.LogWarn("Failed to create fallback provider",
				"provider", providers[i].Type,
				"error", err)
			continue
		}
		fallbacks = append(fallbacks, fallback)
	}

	config := ResilientProviderConfig{
		Primary:        primary,
		Fallbacks:      fallbacks,
		RetryConfig:    DefaultRetryConfig(),
		EnableFallback: len(fallbacks) > 0,
		Timeout:        30 * time.Second,
	}

	return NewResilientProvider(config), nil
}

// ChainProviderConfig represents configuration for a provider in the chain
type ChainProviderConfig struct {
	Type   string // Provider type (openai, anthropic, etc.)
	Model  string // Model name
	APIKey string // API key (optional - can use env vars)
}

// TruncateContext intelligently truncates message context to fit within limits
func TruncateContext(messages []domain.Message, maxTokens int) []domain.Message {
	// This is a simplified implementation
	// In practice, you'd want to count tokens properly

	if len(messages) <= 2 {
		return messages // Keep at least system and last user message
	}

	// Keep system message (if present) and most recent messages
	var result []domain.Message

	// Check for system message
	if len(messages) > 0 && strings.ToLower(string(messages[0].Role)) == "system" {
		result = append(result, messages[0])
	}

	// Add most recent messages
	recentCount := 3 // Keep last 3 messages
	if len(messages) > recentCount {
		result = append(result, messages[len(messages)-recentCount:]...)
	} else {
		result = append(result, messages...)
	}

	return result
}
