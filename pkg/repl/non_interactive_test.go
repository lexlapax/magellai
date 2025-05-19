// ABOUTME: Tests for non-interactive mode detection and handling
// ABOUTME: Validates detection of pipes, terminals, and CI environments

package repl

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestDetectNonInteractiveMode(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() (reader *os.File, writer *os.File, cleanup func())
		envVars    map[string]string
		wantPiped  bool
		wantCI     bool
		wantNonInt bool
	}{
		// Skip this test in non-terminal environments
		// {
		// 	name: "standard terminal input/output",
		// 	setupFunc: func() (*os.File, *os.File, func()) {
		// 		return os.Stdin, os.Stdout, func() {}
		// 	},
		// 	wantPiped:  false,
		// 	wantCI:     false,
		// 	wantNonInt: false,
		// },
		{
			name: "piped input",
			setupFunc: func() (*os.File, *os.File, func()) {
				r, w, _ := os.Pipe()
				return r, os.Stdout, func() {
					r.Close()
					w.Close()
				}
			},
			wantPiped:  true,
			wantCI:     false,
			wantNonInt: true,
		},
		{
			name: "piped output",
			setupFunc: func() (*os.File, *os.File, func()) {
				r, w, _ := os.Pipe()
				return os.Stdin, w, func() {
					r.Close()
					w.Close()
				}
			},
			wantPiped:  true,
			wantCI:     false,
			wantNonInt: true,
		},
		{
			name: "CI environment",
			setupFunc: func() (*os.File, *os.File, func()) {
				return os.Stdin, os.Stdout, func() {}
			},
			envVars: map[string]string{
				"CI": "true",
			},
			wantPiped:  false,
			wantCI:     true,
			wantNonInt: true,
		},
		{
			name: "GitHub Actions",
			setupFunc: func() (*os.File, *os.File, func()) {
				return os.Stdin, os.Stdout, func() {}
			},
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			wantPiped:  false,
			wantCI:     true,
			wantNonInt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			for k, v := range tt.envVars {
				oldVal := os.Getenv(k)
				os.Setenv(k, v)
				defer os.Setenv(k, oldVal)
			}

			// Set up file descriptors
			reader, writer, cleanup := tt.setupFunc()
			defer cleanup()

			// Run detection
			mode := DetectNonInteractiveMode(reader, writer)

			// Check results
			if tt.wantPiped {
				assert.True(t, mode.IsPipedInput || mode.IsPipedOutput,
					"Expected piped mode but got neither piped input nor output")
			}
			assert.Equal(t, tt.wantCI, mode.IsCIEnvironment,
				"CI environment detection mismatch")
			assert.Equal(t, tt.wantNonInt, mode.IsNonInteractive,
				"Non-interactive detection mismatch")
		})
	}
}

func TestConfigureForNonInteractiveMode(t *testing.T) {
	// Create a REPL with test configuration
	config := &testConfig{
		values: map[string]interface{}{
			"repl.colors.enabled": true,
			"repl.multiline":      true,
			"repl.autosave":       false,
			"model":               "mock/test-model",
		},
	}

	// Create a minimal REPL struct for testing
	repl := &REPL{
		config:         config,
		isTerminal:     true,
		multiline:      true,
		exitOnEOF:      false,
		promptStyle:    "> ",
		colorFormatter: utils.NewColorFormatter(true, nil),
		autoSave:       false,
	}

	// Apply non-interactive mode
	mode := NonInteractiveMode{
		IsNonInteractive: true,
		IsPipedInput:     true,
		IsPipedOutput:    false,
	}

	repl.ConfigureForNonInteractiveMode(mode)

	// Verify configuration changes
	assert.False(t, repl.isTerminal, "Terminal should be disabled")
	assert.False(t, repl.multiline, "Multiline should be disabled")
	assert.True(t, repl.exitOnEOF, "Exit on EOF should be enabled")
	assert.Empty(t, repl.promptStyle, "Prompt should be empty for piped input")
	assert.False(t, repl.colorFormatter.Enabled(), "Colors should be disabled")
}

func TestProcessPipedInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		checkFunc func(t *testing.T, repl *REPL)
	}{
		{
			name:      "simple message",
			input:     "Hello, world!",
			wantError: false,
			checkFunc: func(t *testing.T, repl *REPL) {
				// Check that message was processed (1 user + 1 assistant)
				assert.Equal(t, 2, len(repl.session.Conversation.Messages))
				// First message should be user
				assert.Equal(t, domain.MessageRoleUser, repl.session.Conversation.Messages[0].Role)
				assert.Equal(t, "Hello, world!", repl.session.Conversation.Messages[0].Content)
				// Second message should be assistant response
				assert.Equal(t, domain.MessageRoleAssistant, repl.session.Conversation.Messages[1].Role)
			},
		},
		{
			name:      "command input",
			input:     "/help",
			wantError: false,
			checkFunc: func(t *testing.T, repl *REPL) {
				// Commands don't add messages to conversation
				assert.Equal(t, 0, len(repl.session.Conversation.Messages))
			},
		},
		{
			name:      "multiline input",
			input:     "Line 1\nLine 2\nLine 3",
			wantError: false,
			checkFunc: func(t *testing.T, repl *REPL) {
				// All lines should be processed as one message (user + assistant)
				assert.Equal(t, 2, len(repl.session.Conversation.Messages))
				msg := repl.session.Conversation.Messages[0]
				assert.Contains(t, msg.Content, "Line 1")
				assert.Contains(t, msg.Content, "Line 2")
				assert.Contains(t, msg.Content, "Line 3")
			},
		},
		{
			name:      "empty input",
			input:     "",
			wantError: false,
			checkFunc: func(t *testing.T, repl *REPL) {
				// Empty input should not create any messages
				assert.Equal(t, 0, len(repl.session.Conversation.Messages))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewBufferString(tt.input)
			writer := &bytes.Buffer{}

			config := &testConfig{
				values: map[string]interface{}{
					"repl.colors.enabled": false,
					"llm.provider":        "mock",
					"llm.model":           "test",
					"stream":              false,
				},
			}

			repl := &REPL{
				config:         config,
				reader:         bufio.NewReader(reader),
				writer:         writer,
				provider:       newMockProvider(),
				session:        &Session{Conversation: &domain.Conversation{Messages: []domain.Message{}}},
				registry:       command.NewRegistry(),
				isTerminal:     false,
				colorFormatter: utils.NewColorFormatter(false, nil),
				nonInteractive: NonInteractiveMode{
					IsNonInteractive: true,
					IsPipedInput:     true,
				},
			}

			// For the command test, register commands
			if tt.name == "command input" {
				cmd := &mockCommand{name: "help"}
				regErr := repl.registry.Register(cmd)
				assert.NoError(t, regErr)
			}

			// Process piped input
			err := repl.ProcessPipedInput(repl.nonInteractive)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, repl)
			}
		})
	}
}

func TestShouldAutoExit(t *testing.T) {
	tests := []struct {
		name       string
		mode       NonInteractiveMode
		exitOnEOF  bool
		shouldExit bool
	}{
		{
			name: "piped input with exit on EOF",
			mode: NonInteractiveMode{
				IsPipedInput: true,
			},
			exitOnEOF:  true,
			shouldExit: true,
		},
		{
			name: "piped input without exit on EOF",
			mode: NonInteractiveMode{
				IsPipedInput: true,
			},
			exitOnEOF:  false,
			shouldExit: false,
		},
		{
			name: "non-piped input",
			mode: NonInteractiveMode{
				IsPipedInput: false,
			},
			exitOnEOF:  true,
			shouldExit: false,
		},
		{
			name: "CI environment",
			mode: NonInteractiveMode{
				IsPipedInput:    false,
				IsCIEnvironment: true,
			},
			exitOnEOF:  true,
			shouldExit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repl := &REPL{
				exitOnEOF: tt.exitOnEOF,
			}

			shouldExit := repl.ShouldAutoExit(tt.mode)
			assert.Equal(t, tt.shouldExit, shouldExit)
		})
	}
}

// mockCommand is a mock command for testing
type mockCommand struct {
	name string
}

func (m *mockCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	return nil
}

func (m *mockCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        m.name,
		Description: "Mock command for testing",
		Category:    command.CategoryShared,
	}
}

func (m *mockCommand) Validate() error {
	return nil
}
