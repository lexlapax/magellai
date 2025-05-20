// ABOUTME: Tests for ANSI color formatting utilities
// ABOUTME: Validates color application and stripping functions

package stringutil

import (
	"strings"
	"testing"
)

func TestColorText(t *testing.T) {
	testCases := []struct {
		color    string
		text     string
		expected string
	}{
		{"red", "error", "\033[31merror\033[0m"},
		{"green", "success", "\033[32msuccess\033[0m"},
		{"blue", "info", "\033[34minfo\033[0m"},
		{"nonexistent", "text", "text"}, // Invalid color should return original text
		{"", "text", "text"},            // Empty color should return original text
	}

	for _, tc := range testCases {
		result := ColorText(tc.color, tc.text)
		if result != tc.expected {
			t.Errorf("ColorText(%s, %s) failed: got %q, want %q",
				tc.color, tc.text, result, tc.expected)
		}
	}
}

func TestColorTextf(t *testing.T) {
	result := ColorTextf("red", "Error: %s", "file not found")
	expected := "\033[31mError: file not found\033[0m"

	if result != expected {
		t.Errorf("ColorTextf failed: got %q, want %q", result, expected)
	}
}

func TestStripColors(t *testing.T) {
	testCases := []struct {
		colored  string
		expected string
	}{
		{"\033[31mError\033[0m", "Error"},
		{"\033[32mSuccess\033[0m", "Success"},
		{"\033[1;34mBold Info\033[0m", "Bold Info"},
		{"Plain text", "Plain text"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := StripColors(tc.colored)
		if result != tc.expected {
			t.Errorf("StripColors(%q) failed: got %q, want %q",
				tc.colored, result, tc.expected)
		}
	}
}

func TestFormatFunctions(t *testing.T) {
	testCases := []struct {
		formatter func(string) string
		text      string
		contains  string
	}{
		{FormatCommand, "help", "\033[36m"},        // Cyan
		{FormatError, "error", "\033[31m"},         // Red
		{FormatWarning, "warning", "\033[33m"},     // Yellow
		{FormatSuccess, "success", "\033[32m"},     // Green
		{FormatHighlight, "highlight", "\033[35m"}, // Magenta
		{FormatBold, "bold", "\033[1m"},            // Bold
	}

	for _, tc := range testCases {
		result := tc.formatter(tc.text)
		if !strings.Contains(result, tc.contains) {
			t.Errorf("%T(%s) failed: result %q doesn't contain %q",
				tc.formatter, tc.text, result, tc.contains)
		}
		if !strings.Contains(result, tc.text) {
			t.Errorf("%T(%s) failed: result %q doesn't contain original text %q",
				tc.formatter, tc.text, result, tc.text)
		}
		if !strings.HasSuffix(result, Reset) {
			t.Errorf("%T(%s) failed: result %q doesn't end with Reset code",
				tc.formatter, tc.text, result)
		}
	}
}

func TestFormatWithColor(t *testing.T) {
	// Test with colors enabled
	EnableColors = true
	result := FormatWithColor("red", "error")
	if !strings.Contains(result, Red) {
		t.Errorf("FormatWithColor with enabled colors should apply color: %q", result)
	}

	// Test with colors disabled
	EnableColors = false
	result = FormatWithColor("red", "error")
	if result != "error" {
		t.Errorf("FormatWithColor with disabled colors should return plain text: %q", result)
	}

	// Reset for other tests
	EnableColors = true
}
