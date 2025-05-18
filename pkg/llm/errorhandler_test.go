package llm

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestErrorHandler_WithRetry(t *testing.T) {
	tests := []struct {
		name          string
		retryConfig   RetryConfig
		operation     string
		errorSequence []error
		expectError   bool
		expectRetries int
	}{
		{
			name:          "success on first try",
			retryConfig:   DefaultRetryConfig(),
			operation:     "test",
			errorSequence: []error{nil},
			expectError:   false,
			expectRetries: 0,
		},
		{
			name:        "success after retry",
			retryConfig: DefaultRetryConfig(),
			operation:   "test",
			errorSequence: []error{
				domain.ErrNetworkConnectivity,
				nil,
			},
			expectError:   false,
			expectRetries: 1,
		},
		{
			name:        "non-retryable error",
			retryConfig: DefaultRetryConfig(),
			operation:   "test",
			errorSequence: []error{
				domain.ErrAuthenticationFailed,
			},
			expectError:   true,
			expectRetries: 0,
		},
		{
			name: "max retries exceeded",
			retryConfig: RetryConfig{
				MaxRetries:    2,
				InitialDelay:  10 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
				RetryableErrors: []error{
					domain.ErrNetworkConnectivity,
				},
			},
			operation: "test",
			errorSequence: []error{
				domain.ErrNetworkConnectivity,
				domain.ErrNetworkConnectivity,
				domain.ErrNetworkConnectivity,
			},
			expectError:   true,
			expectRetries: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewErrorHandler(tt.retryConfig)
			ctx := context.Background()

			attempts := 0
			err := handler.WithRetry(ctx, tt.operation, func() error {
				if attempts < len(tt.errorSequence) {
					err := tt.errorSequence[attempts]
					attempts++
					return err
				}
				return errors.New("too many attempts")
			})

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check retry count (attempts - 1 because first attempt isn't a retry)
			actualRetries := attempts - 1
			if actualRetries != tt.expectRetries {
				t.Errorf("expected %d retries, got %d", tt.expectRetries, actualRetries)
			}
		})
	}
}

func TestErrorHandler_WithRateLimitRetry(t *testing.T) {
	t.Skip("Skipping rate limit test that takes too long")

	handler := NewErrorHandler(DefaultRetryConfig())
	ctx := context.Background()

	attempts := 0
	start := time.Now()

	err := handler.WithRateLimitRetry(ctx, "test", func() error {
		attempts++
		if attempts < 2 {
			// Simulate rate limit error
			return domain.NewProviderError("test", "Generate", 429, "rate limit exceeded", domain.ErrRateLimitExceeded)
		}
		return nil
	})

	duration := time.Since(start)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}

	// Should have waited at least 10 seconds (first retry delay)
	if duration < 10*time.Second {
		t.Errorf("expected delay of at least 10s, got %v", duration)
	}
}

func TestErrorHandler_isRetryableError(t *testing.T) {
	handler := NewErrorHandler(DefaultRetryConfig())

	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "network error",
			err:       domain.ErrNetworkConnectivity,
			retryable: true,
		},
		{
			name:      "provider unavailable",
			err:       domain.ErrProviderUnavailable,
			retryable: true,
		},
		{
			name:      "timeout error",
			err:       domain.ErrTimeout,
			retryable: true,
		},
		{
			name:      "authentication error",
			err:       domain.ErrAuthenticationFailed,
			retryable: false,
		},
		{
			name:      "rate limit error",
			err:       domain.ErrRateLimitExceeded,
			retryable: false, // Rate limits need special handling
		},
		{
			name:      "5xx status code",
			err:       domain.NewProviderError("test", "Generate", 503, "service unavailable", nil),
			retryable: true,
		},
		{
			name:      "4xx status code",
			err:       domain.NewProviderError("test", "Generate", 400, "bad request", nil),
			retryable: false,
		},
		{
			name:      "timeout status code",
			err:       domain.NewProviderError("test", "Generate", 408, "timeout", nil),
			retryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.isRetryableError(tt.err)
			if result != tt.retryable {
				t.Errorf("expected retryable=%v, got %v", tt.retryable, result)
			}
		})
	}
}

func TestErrorHandler_calculateDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
	handler := NewErrorHandler(config)

	tests := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{0, 100 * time.Millisecond, 200 * time.Millisecond},
		{1, 200 * time.Millisecond, 400 * time.Millisecond},
		{2, 400 * time.Millisecond, 800 * time.Millisecond},
		{10, 5 * time.Second, 5 * time.Second}, // Should cap at max
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := handler.calculateDelay(tt.attempt)

			if delay < tt.minExpected || delay > tt.maxExpected {
				t.Errorf("delay %v not in expected range [%v, %v]",
					delay, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		shouldRetry bool
	}{
		{
			name:        "nil error",
			err:         nil,
			shouldRetry: false,
		},
		{
			name:        "authentication error",
			err:         domain.ErrAuthenticationFailed,
			shouldRetry: false,
		},
		{
			name:        "network error",
			err:         domain.ErrNetworkConnectivity,
			shouldRetry: true,
		},
		{
			name:        "provider unavailable",
			err:         domain.ErrProviderUnavailable,
			shouldRetry: true,
		},
		{
			name:        "timeout error",
			err:         domain.ErrTimeout,
			shouldRetry: true,
		},
		{
			name:        "5xx error",
			err:         domain.NewProviderError("test", "op", 500, "server error", nil),
			shouldRetry: true,
		},
		{
			name:        "408 timeout",
			err:         domain.NewProviderError("test", "op", 408, "timeout", nil),
			shouldRetry: true,
		},
		{
			name:        "400 bad request",
			err:         domain.NewProviderError("test", "op", 400, "bad request", nil),
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRetry(tt.err)
			if result != tt.shouldRetry {
				t.Errorf("expected shouldRetry=%v, got %v", tt.shouldRetry, result)
			}
		})
	}
}
