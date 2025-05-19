// ABOUTME: Tests for ANSI color output functionality
// ABOUTME: Ensures color formatting works correctly in TTY environments

package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorFormatter(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		text     string
		color    string
		expected string
	}{
		{
			name:     "color enabled",
			enabled:  true,
			text:     "Hello",
			color:    ColorRed,
			expected: ColorRed + "Hello" + ColorReset,
		},
		{
			name:     "color disabled",
			enabled:  false,
			text:     "Hello",
			color:    ColorRed,
			expected: "Hello",
		},
		{
			name:     "empty text",
			enabled:  true,
			text:     "",
			color:    ColorBlue,
			expected: ColorBlue + ColorReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewColorFormatter(tt.enabled, nil)
			result := formatter.Format(tt.color, tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestThemeFormatting(t *testing.T) {
	formatter := NewColorFormatter(true, nil)

	tests := []struct {
		name   string
		format func(string) string
		text   string
	}{
		{"command", formatter.FormatCommand, "help"},
		{"error", formatter.FormatError, "error message"},
		{"warning", formatter.FormatWarning, "warning"},
		{"success", formatter.FormatSuccess, "success"},
		{"info", formatter.FormatInfo, "info"},
		{"user", formatter.FormatUserMessage, "user text"},
		{"assistant", formatter.FormatAssistantMessage, "assistant response"},
		{"system", formatter.FormatSystemMessage, "system message"},
		{"code block", formatter.FormatCodeBlock, "print('hello')"},
		{"highlight", formatter.FormatHighlight, "important"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.format(tt.text)
			// Check that color codes are added
			assert.Contains(t, result, ColorReset)
			assert.Contains(t, result, tt.text)
			assert.NotEqual(t, tt.text, result)
		})
	}
}

func TestCommandHelp(t *testing.T) {
	formatter := NewColorFormatter(true, nil)

	result := formatter.FormatCommandHelp("help", "Show help")

	// Should contain both the command and description
	assert.Contains(t, result, "help")
	assert.Contains(t, result, "Show help")
	assert.Contains(t, result, ColorReset)
}

func TestPromptFormatting(t *testing.T) {
	formatter := NewColorFormatter(true, nil)

	prompt := "> "
	result := formatter.FormatPrompt(prompt)

	// Check bold is applied
	assert.Contains(t, result, ColorBold)
	assert.Contains(t, result, prompt)
	assert.Contains(t, result, ColorReset)
}

func TestStripColors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple color",
			input:    ColorRed + "Hello" + ColorReset,
			expected: "Hello",
		},
		{
			name:     "multiple colors",
			input:    ColorRed + "Error: " + ColorReset + ColorBlue + "Info" + ColorReset,
			expected: "Error: Info",
		},
		{
			name:     "no colors",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "incomplete escape",
			input:    "Text\033[incomplete",
			expected: "Text\033[incomplete",
		},
		{
			name:     "complex formatting",
			input:    ColorBold + ColorRed + "Bold Red" + ColorReset,
			expected: "Bold Red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripColors(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorToggle(t *testing.T) {
	formatter := NewColorFormatter(true, nil)

	// Initially enabled
	assert.True(t, formatter.Enabled())
	result := formatter.FormatError("error")
	assert.Contains(t, result, ColorReset)

	// Disable colors
	formatter.SetEnabled(false)
	assert.False(t, formatter.Enabled())
	result = formatter.FormatError("error")
	assert.Equal(t, "error", result)

	// Re-enable colors
	formatter.SetEnabled(true)
	assert.True(t, formatter.Enabled())
	result = formatter.FormatError("error")
	assert.Contains(t, result, ColorReset)
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultColorTheme()

	// Check some key theme settings
	assert.Equal(t, ColorBrightCyan, theme.Command)
	assert.Equal(t, ColorRed, theme.Error)
	assert.Equal(t, ColorGreen, theme.Success)
	assert.Equal(t, ColorYellow, theme.Warning)
}

func TestColorConstants(t *testing.T) {
	// Verify ANSI codes are correct
	assert.Equal(t, "\033[0m", ColorReset)
	assert.Equal(t, "\033[31m", ColorRed)
	assert.Equal(t, "\033[32m", ColorGreen)
	assert.Equal(t, "\033[33m", ColorYellow)
	assert.Equal(t, "\033[34m", ColorBlue)
}

func TestFormatString(t *testing.T) {
	formatter := NewColorFormatter(true, nil)

	// Test color formatting works with special characters
	tests := []string{
		"Hello\nWorld",
		"With\ttabs",
		"Special © ® chars",
		"Unicode: こんにちは",
		"Emoji: 😊",
	}

	for _, text := range tests {
		colored := formatter.FormatInfo(text)
		assert.Contains(t, colored, text)
		assert.True(t, strings.HasPrefix(colored, ColorBlue))
		assert.True(t, strings.HasSuffix(colored, ColorReset))

		// Strip colors should restore original
		stripped := StripColors(colored)
		assert.Equal(t, text, stripped)
	}
}
