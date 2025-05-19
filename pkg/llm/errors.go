// ABOUTME: Error definitions for the LLM package
// ABOUTME: Provides standard errors for LLM operations

package llm

import "errors"

// LLM-specific errors
var (
	// ErrProviderNotFound indicates the provider was not found
	ErrProviderNotFound = errors.New("LLM provider not found")

	// ErrModelNotFound indicates the model was not found
	ErrModelNotFound = errors.New("LLM model not found")

	// ErrInvalidProvider indicates an invalid provider
	ErrInvalidProvider = errors.New("invalid LLM provider")

	// ErrInvalidModel indicates an invalid model
	ErrInvalidModel = errors.New("invalid LLM model")

	// ErrAPIKeyMissing indicates the API key is missing
	ErrAPIKeyMissing = errors.New("API key missing")

	// ErrInvalidAPIKey indicates an invalid API key
	ErrInvalidAPIKey = errors.New("invalid API key")

	// ErrProviderUnavailable indicates the provider is temporarily unavailable
	ErrProviderUnavailable = errors.New("LLM provider unavailable")

	// ErrContextLengthExceeded indicates the context length was exceeded
	ErrContextLengthExceeded = errors.New("context length exceeded")

	// ErrRateLimitExceeded indicates the rate limit was exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrTokenLimitExceeded indicates the token limit was exceeded
	ErrTokenLimitExceeded = errors.New("token limit exceeded")

	// ErrStreamingNotSupported indicates streaming is not supported
	ErrStreamingNotSupported = errors.New("streaming not supported")

	// ErrInvalidResponse indicates an invalid response from the provider
	ErrInvalidResponse = errors.New("invalid provider response")

	// ErrPartialResponse indicates a partial response was received
	ErrPartialResponse = errors.New("partial response received")

	// ErrProviderTimeout indicates a provider timeout
	ErrProviderTimeout = errors.New("provider timeout")

	// ErrProviderError indicates a generic provider error
	ErrProviderError = errors.New("provider error")
)
