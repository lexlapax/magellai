// ABOUTME: Tests for readline functionality including tab completion
// ABOUTME: Ensures tab completion works correctly for REPL commands

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestREPLCompleter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		position int
		expected []string
	}{
		{
			name:     "complete help command",
			input:    "/h",
			position: 2,
			expected: []string{"/help", "/history"},
		},
		{
			name:     "complete with colon prefix",
			input:    ":s",
			position: 2,
			expected: []string{":save", ":session", ":stats", ":system"},
		},
		{
			name:     "no completion for plain text",
			input:    "hello",
			position: 5,
			expected: nil,
		},
		{
			name:     "complete model command",
			input:    "/mod",
			position: 4,
			expected: []string{"/model"},
		},
		{
			name:     "complete empty command",
			input:    "/",
			position: 1,
			expected: []string{
				"/help", "/exit", "/quit", "/clear", "/history",
				"/save", "/export", "/multiline", "/model", "/config",
				"/attach", "/session", "/context", "/reset", "/undo",
				"/redo", "/theme", "/alias", "/info", "/stats",
				"/logs", "/system",
			},
		},
	}

	// Create a completer with test commands
	completer := &ReplCompleter{
		Commands: getCommandNames(),
		Registry: nil,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := []rune(tt.input)
			newLines, offset := completer.Do(line, tt.position)

			// Convert rune slices to strings for comparison
			var completions []string
			for _, runes := range newLines {
				completions = append(completions, string(runes))
			}

			assert.Equal(t, tt.expected, completions)
			assert.Equal(t, 0, offset)
		})
	}
}

func TestReadlineInterface(t *testing.T) {
	// Create a test config
	config := &ReadlineConfig{
		Prompt:           "> ",
		HistoryFile:      "",
		EnableCompletion: true,
		MultilineMode:    false,
	}

	// Create readline interface
	rl, err := NewReadlineInterface(config)
	require.NoError(t, err)
	require.NotNil(t, rl)
	defer rl.Close()

	// Test setting prompt
	rl.SetPrompt(">>> ")
	// Note: we can't easily test the actual prompt without mocking the terminal

	// Test that the instance was created
	assert.NotNil(t, rl.Instance)
}

func TestGetCommandNames(t *testing.T) {
	commands := getCommandNames()

	// Verify we have a reasonable set of commands
	assert.Greater(t, len(commands), 10)

	// Check for some essential commands
	essentialCommands := []string{"help", "exit", "quit", "save", "model", "config"}
	for _, cmd := range essentialCommands {
		assert.Contains(t, commands, cmd)
	}
}
