// ABOUTME: Tests for LLM error definitions
// ABOUTME: Verifies error constants and error handling

package llm

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "ErrProviderNotFound",
			err:  ErrProviderNotFound,
			msg:  "LLM provider not found",
		},
		{
			name: "ErrModelNotFound",
			err:  ErrModelNotFound,
			msg:  "LLM model not found",
		},
		{
			name: "ErrInvalidProvider",
			err:  ErrInvalidProvider,
			msg:  "invalid LLM provider",
		},
		{
			name: "ErrInvalidModel",
			err:  ErrInvalidModel,
			msg:  "invalid LLM model",
		},
		{
			name: "ErrAPIKeyMissing",
			err:  ErrAPIKeyMissing,
			msg:  "API key missing",
		},
		{
			name: "ErrInvalidAPIKey",
			err:  ErrInvalidAPIKey,
			msg:  "invalid API key",
		},
		{
			name: "ErrProviderUnavailable",
			err:  ErrProviderUnavailable,
			msg:  "LLM provider unavailable",
		},
		{
			name: "ErrContextLengthExceeded",
			err:  ErrContextLengthExceeded,
			msg:  "context length exceeded",
		},
		{
			name: "ErrRateLimitExceeded",
			err:  ErrRateLimitExceeded,
			msg:  "rate limit exceeded",
		},
		{
			name: "ErrTokenLimitExceeded",
			err:  ErrTokenLimitExceeded,
			msg:  "token limit exceeded",
		},
		{
			name: "ErrStreamingNotSupported",
			err:  ErrStreamingNotSupported,
			msg:  "streaming not supported",
		},
		{
			name: "ErrInvalidResponse",
			err:  ErrInvalidResponse,
			msg:  "invalid provider response",
		},
		{
			name: "ErrPartialResponse",
			err:  ErrPartialResponse,
			msg:  "partial response received",
		},
		{
			name: "ErrProviderTimeout",
			err:  ErrProviderTimeout,
			msg:  "provider timeout",
		},
		{
			name: "ErrProviderError",
			err:  ErrProviderError,
			msg:  "provider error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.msg, tt.err.Error())
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all errors are properly defined as error types
	errs := []error{
		ErrProviderNotFound,
		ErrModelNotFound,
		ErrInvalidProvider,
		ErrInvalidModel,
		ErrAPIKeyMissing,
		ErrInvalidAPIKey,
		ErrProviderUnavailable,
		ErrContextLengthExceeded,
		ErrRateLimitExceeded,
		ErrTokenLimitExceeded,
		ErrStreamingNotSupported,
		ErrInvalidResponse,
		ErrPartialResponse,
		ErrProviderTimeout,
		ErrProviderError,
	}

	for i, err := range errs {
		require.NotNil(t, err, "Error at index %d should not be nil", i)
		require.Error(t, err, "Value at index %d should be an error", i)
	}
}

func TestErrorComparison(t *testing.T) {
	// Test using errors.Is for error comparison
	tests := []struct {
		name        string
		err1        error
		err2        error
		shouldMatch bool
	}{
		{
			name:        "same error",
			err1:        ErrProviderNotFound,
			err2:        ErrProviderNotFound,
			shouldMatch: true,
		},
		{
			name:        "different errors",
			err1:        ErrProviderNotFound,
			err2:        ErrModelNotFound,
			shouldMatch: false,
		},
		{
			name:        "nil error",
			err1:        nil,
			err2:        ErrProviderNotFound,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err1, tt.err2)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

func TestErrorMessages(t *testing.T) {
	// Test that error messages are unique
	errorMessages := make(map[string]error)
	allErrors := []error{
		ErrProviderNotFound,
		ErrModelNotFound,
		ErrInvalidProvider,
		ErrInvalidModel,
		ErrAPIKeyMissing,
		ErrInvalidAPIKey,
		ErrProviderUnavailable,
		ErrContextLengthExceeded,
		ErrRateLimitExceeded,
		ErrTokenLimitExceeded,
		ErrStreamingNotSupported,
		ErrInvalidResponse,
		ErrPartialResponse,
		ErrProviderTimeout,
		ErrProviderError,
	}

	for _, err := range allErrors {
		msg := err.Error()
		if existing, found := errorMessages[msg]; found {
			t.Errorf("Duplicate error message '%s' for %v and %v", msg, err, existing)
		}
		errorMessages[msg] = err
	}
}

// Test error string constants are not empty
func TestErrorStringsNotEmpty(t *testing.T) {
	allErrors := []error{
		ErrProviderNotFound,
		ErrModelNotFound,
		ErrInvalidProvider,
		ErrInvalidModel,
		ErrAPIKeyMissing,
		ErrInvalidAPIKey,
		ErrProviderUnavailable,
		ErrContextLengthExceeded,
		ErrRateLimitExceeded,
		ErrTokenLimitExceeded,
		ErrStreamingNotSupported,
		ErrInvalidResponse,
		ErrPartialResponse,
		ErrProviderTimeout,
		ErrProviderError,
	}

	for _, err := range allErrors {
		assert.NotEmpty(t, err.Error(), "Error message should not be empty for %v", err)
	}
}

// Benchmark error comparison
func BenchmarkErrorComparison(b *testing.B) {
	err1 := ErrProviderNotFound
	err2 := ErrProviderNotFound

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.Is(err1, err2)
	}
}

func BenchmarkErrorCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = errors.New("test error")
	}
}

// Test error wrapping and unwrapping
func TestErrorWrapping(t *testing.T) {
	originalErr := ErrProviderNotFound
	wrappedErr := fmt.Errorf("failed to connect: %w", originalErr)

	assert.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "failed to connect")
	assert.True(t, errors.Is(wrappedErr, originalErr))
}

// Test error categorization helpers
func TestErrorCategorization(t *testing.T) {
	providerErrors := []error{
		ErrProviderNotFound,
		ErrInvalidProvider,
		ErrProviderUnavailable,
		ErrProviderTimeout,
		ErrProviderError,
	}

	authErrors := []error{
		ErrAPIKeyMissing,
		ErrInvalidAPIKey,
	}

	limitErrors := []error{
		ErrContextLengthExceeded,
		ErrRateLimitExceeded,
		ErrTokenLimitExceeded,
	}

	// Test that we can group errors
	for _, err := range providerErrors {
		assert.Contains(t, err.Error(), "provider", "Error %v should mention 'provider'", err)
	}

	for _, err := range authErrors {
		assert.Contains(t, err.Error(), "API key", "Error %v should mention 'API key'", err)
	}

	for _, err := range limitErrors {
		assert.Contains(t, err.Error(), "exceeded", "Error %v should mention 'exceeded'", err)
	}
}

// Test retryable vs permanent errors
func TestRetryableErrors(t *testing.T) {
	// These errors are typically retryable
	retryableErrors := []error{
		ErrProviderUnavailable,
		ErrProviderTimeout,
		ErrRateLimitExceeded,
	}

	// These errors are typically not retryable
	permanentErrors := []error{
		ErrAPIKeyMissing,
		ErrInvalidAPIKey,
		ErrInvalidModel,
		ErrInvalidProvider,
		ErrContextLengthExceeded,
		ErrTokenLimitExceeded,
	}

	// Just verify that they exist and are different
	for _, err := range retryableErrors {
		assert.Error(t, err)
	}

	for _, err := range permanentErrors {
		assert.Error(t, err)
	}
}

// Test error consistency
func TestErrorConsistency(t *testing.T) {
	// Test that error messages follow consistent patterns
	tests := []struct {
		err      error
		contains string
	}{
		{ErrProviderNotFound, "not found"},
		{ErrModelNotFound, "not found"},
		{ErrInvalidProvider, "invalid"},
		{ErrInvalidModel, "invalid"},
		{ErrAPIKeyMissing, "missing"},
		{ErrInvalidAPIKey, "invalid"},
		{ErrProviderUnavailable, "unavailable"},
		{ErrContextLengthExceeded, "exceeded"},
		{ErrRateLimitExceeded, "exceeded"},
		{ErrTokenLimitExceeded, "exceeded"},
		{ErrStreamingNotSupported, "not supported"},
		{ErrInvalidResponse, "invalid"},
		{ErrPartialResponse, "partial"},
		{ErrProviderTimeout, "timeout"},
		{ErrProviderError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Contains(t, tt.err.Error(), tt.contains)
		})
	}
}
