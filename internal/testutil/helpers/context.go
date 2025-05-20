// ABOUTME: Context helper functions for testing
// ABOUTME: Provides utilities for creating test contexts and timeouts

package helpers

import (
	"context"
	"testing"
	"time"
)

// Define custom context key types to avoid collisions
type testContextKeyType string

const (
	testIDKey  testContextKeyType = "test_id"
	testRunKey testContextKeyType = "test_run"
)

// TestContext creates a context with a default timeout for tests
func TestContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestContextWithTimeout creates a context with specified timeout
func TestContextWithTimeout(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

// TestContextWithCancel creates a context that can be cancelled
func TestContextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx, cancel
}

// AssertContextDone checks if context is done within timeout
func AssertContextDone(t *testing.T, ctx context.Context, timeout time.Duration) {
	t.Helper()

	select {
	case <-ctx.Done():
		// Context is done as expected
	case <-time.After(timeout):
		t.Errorf("Context was not done within %v", timeout)
	}
}

// AssertContextNotDone checks if context is not done within timeout
func AssertContextNotDone(t *testing.T, ctx context.Context, timeout time.Duration) {
	t.Helper()

	select {
	case <-ctx.Done():
		t.Errorf("Context was done but expected to be active")
	case <-time.After(timeout):
		// Context is still active as expected
	}
}

// WaitForCondition waits for a condition to be true or times out
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, checkInterval time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(checkInterval)
	}

	t.Errorf("Condition was not met within %v", timeout)
}

// RunWithTimeout runs a function with a timeout
func RunWithTimeout(t *testing.T, timeout time.Duration, f func()) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)
		f()
	}()

	select {
	case <-done:
		// Function completed
	case <-time.After(timeout):
		t.Fatalf("Function did not complete within %v", timeout)
	}
}

// TestDeadline returns a deadline for test operations
func TestDeadline(t *testing.T, duration time.Duration) time.Time {
	t.Helper()
	return time.Now().Add(duration)
}

// ContextWithTestValues creates a context with common test values
func ContextWithTestValues(t *testing.T) context.Context {
	t.Helper()

	ctx := TestContext(t)
	ctx = context.WithValue(ctx, testIDKey, t.Name())
	ctx = context.WithValue(ctx, testRunKey, time.Now().UnixNano())
	return ctx
}
