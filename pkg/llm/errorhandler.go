// ABOUTME: Error handling and recovery mechanisms for LLM provider operations
// ABOUTME: Implements retry logic, fallback mechanisms, and graceful error recovery

package llm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/magellai/internal/logging"
)

// RetryConfig defines retry behavior for network operations
type RetryConfig struct {
	MaxRetries      int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	BackoffFactor   float64       // Exponential backoff factor
	RetryableErrors []error       // Specific errors that should trigger retries
}

// DefaultRetryConfig returns sensible defaults for retry behavior
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []error{
			domain.ErrNetworkConnectivity,
			domain.ErrProviderUnavailable,
			domain.ErrTimeout,
			// Don't retry rate limits here - they need special handling
		},
	}
}

// ErrorHandler provides error handling and recovery mechanisms
type ErrorHandler struct {
	retryConfig RetryConfig
	logger      *logging.Logger
}

// NewErrorHandler creates a new error handler with retry and recovery logic
func NewErrorHandler(config RetryConfig) *ErrorHandler {
	return &ErrorHandler{
		retryConfig: config,
		logger:      logging.GetLogger(),
	}
}

// WithRetry executes a function with retry logic for transient errors
func (h *ErrorHandler) WithRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= h.retryConfig.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !h.isRetryableError(err) {
			h.logger.Debug("Error is not retryable", "operation", operation, "error", err)
			return err
		}

		// Check if context is cancelled
		if ctx.Err() != nil {
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		}

		// Don't retry on last attempt
		if attempt == h.retryConfig.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := h.calculateDelay(attempt)
		h.logger.Info("Retrying operation after error",
			"operation", operation,
			"attempt", attempt+1,
			"delay", delay,
			"error", err)

		// Wait before retry
		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during retry: %w", ctx.Err())
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", h.retryConfig.MaxRetries, lastErr)
}

// WithRateLimitRetry handles rate limit errors with intelligent backoff
func (h *ErrorHandler) WithRateLimitRetry(ctx context.Context, operation string, fn func() error) error {
	const maxRateLimitRetries = 3

	for attempt := 0; attempt < maxRateLimitRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if it's a rate limit error
		if !domain.IsRateLimitError(err) {
			return err
		}

		// Rate limits need longer waits
		delay := time.Duration(math.Pow(2, float64(attempt))) * 10 * time.Second
		if delay > time.Minute {
			delay = time.Minute
		}

		h.logger.Warn("Rate limit exceeded, waiting before retry",
			"operation", operation,
			"attempt", attempt+1,
			"delay", delay)

		select {
		case <-time.After(delay):
			// Continue to next retry
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during rate limit wait: %w", ctx.Err())
		}
	}

	return fmt.Errorf("rate limit persists after retries")
}

// isRetryableError checks if an error should trigger a retry
func (h *ErrorHandler) isRetryableError(err error) bool {
	// Check against configured retryable errors
	for _, retryableErr := range h.retryConfig.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	// Check for specific provider errors that are retryable
	var providerErr *domain.ProviderError
	if errors.As(err, &providerErr) {
		// 5xx errors are generally retryable
		if providerErr.StatusCode >= 500 && providerErr.StatusCode < 600 {
			return true
		}
		// Network timeouts are retryable
		if providerErr.StatusCode == 408 {
			return true
		}
	}

	return false
}

// calculateDelay calculates the retry delay with exponential backoff
func (h *ErrorHandler) calculateDelay(attempt int) time.Duration {
	delay := float64(h.retryConfig.InitialDelay) * math.Pow(h.retryConfig.BackoffFactor, float64(attempt))

	// Add jitter to prevent thundering herd
	jitter := time.Duration(float64(delay) * 0.1)
	delay += float64(jitter)

	// Cap at max delay
	if time.Duration(delay) > h.retryConfig.MaxDelay {
		return h.retryConfig.MaxDelay
	}

	return time.Duration(delay)
}

// HandleError provides intelligent error handling with logging
func (h *ErrorHandler) HandleError(err error, operation string, provider string) error {
	if err == nil {
		return nil
	}

	// Log error with appropriate level based on type
	switch {
	case domain.IsAuthenticationError(err):
		h.logger.Error("Authentication failed",
			"operation", operation,
			"provider", provider,
			"error", err)
		return fmt.Errorf("authentication failed for %s: %w", provider, err)

	case domain.IsRateLimitError(err):
		h.logger.Warn("Rate limit exceeded",
			"operation", operation,
			"provider", provider,
			"error", err)
		return fmt.Errorf("rate limit exceeded for %s: %w", provider, err)

	case domain.IsNetworkConnectivityError(err):
		h.logger.Warn("Network connectivity issue",
			"operation", operation,
			"provider", provider,
			"error", err)
		return fmt.Errorf("network issue with %s: %w", provider, err)

	case domain.IsProviderUnavailableError(err):
		h.logger.Warn("Provider unavailable",
			"operation", operation,
			"provider", provider,
			"error", err)
		return fmt.Errorf("%s is temporarily unavailable: %w", provider, err)

	case errors.Is(err, domain.ErrContextTooLong):
		h.logger.Info("Context too long",
			"operation", operation,
			"provider", provider,
			"error", err)
		return fmt.Errorf("input too long for %s: %w", provider, err)

	default:
		h.logger.Error("Provider operation failed",
			"operation", operation,
			"provider", provider,
			"error", err)
		return fmt.Errorf("%s operation failed: %w", operation, err)
	}
}

// RecoverFromPanic recovers from panics and converts them to errors
func (h *ErrorHandler) RecoverFromPanic(operation string) (err error) {
	if r := recover(); r != nil {
		err = fmt.Errorf("panic during %s: %v", operation, r)
		h.logger.Error("Recovered from panic", "operation", operation, "panic", r)
	}
	return err
}

// IsContextTooLongError checks if the error is due to context length
func IsContextTooLongError(err error) bool {
	return errors.Is(err, domain.ErrContextTooLong)
}

// Domain errors that might not be exported from go-llms
var (
	// Define local reference if needed
	ErrContextTooLong = domain.ErrContextTooLong
)

// ShouldRetry determines if an operation should be retried based on the error
func ShouldRetry(err error) bool {
	// Don't retry on nil error
	if err == nil {
		return false
	}

	// Don't retry authentication errors
	if domain.IsAuthenticationError(err) {
		return false
	}

	// Don't retry invalid parameter errors
	if domain.IsInvalidModelParametersError(err) {
		return false
	}

	// Don't retry content filtered errors
	if domain.IsContentFilteredError(err) {
		return false
	}

	// Retry network and availability issues
	if domain.IsNetworkConnectivityError(err) ||
		domain.IsProviderUnavailableError(err) ||
		domain.IsTimeoutError(err) {
		return true
	}

	// Check status codes for provider errors
	var providerErr *domain.ProviderError
	if errors.As(err, &providerErr) {
		// Retry 5xx errors
		if providerErr.StatusCode >= 500 && providerErr.StatusCode < 600 {
			return true
		}
		// Retry network timeouts (408)
		if providerErr.StatusCode == 408 {
			return true
		}
	}

	return false
}
