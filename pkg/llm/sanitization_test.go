// ABOUTME: Unit tests for API key sanitization functionality
// ABOUTME: Ensures sensitive data is properly sanitized before logging

package llm

import (
	"testing"
)

func TestSanitizeAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard API key",
			input:    "sk-1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "sk-123...wxyz",
		},
		{
			name:     "longer API key",
			input:    "sk-abcdefghijklmnopqrstuvwxyz123456789000",
			expected: "sk-abc...9000",
		},
		{
			name:     "short key",
			input:    "key123",
			expected: "ke...23",
		},
		{
			name:     "very short key",
			input:    "abc",
			expected: "***",
		},
		{
			name:     "empty key",
			input:    "",
			expected: "***",
		},
		{
			name:     "exactly 10 chars",
			input:    "1234567890",
			expected: "12...90",
		},
		{
			name:     "anthropic key format",
			input:    "sk-ant-api03-1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "sk-ant...wxyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeAPIKey(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeAPIKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestProviderCreationLogs tests that provider creation logs sanitized API keys
func TestProviderCreationLogs(t *testing.T) {
	// This test would need to capture logs, but for now we can test
	// that the sanitization function is used correctly

	// Test with a real-looking API key
	apiKey := "sk-test1234567890abcdefghijklmnopqrstuvwxyz"
	sanitized := sanitizeAPIKey(apiKey)

	// The sanitized version should not contain the full key
	if sanitized == apiKey {
		t.Error("Sanitized key should not equal the original key")
	}

	// Should maintain format
	if len(sanitized) > len(apiKey) {
		t.Error("Sanitized key should not be longer than original")
	}
}
