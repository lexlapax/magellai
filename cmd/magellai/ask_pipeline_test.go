package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/internal/logging"
)

func TestAskCmd_PipelineSupport(t *testing.T) {
	// Create temporary configuration directory
	tmpDir, err := os.MkdirTemp("", "magellai-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Set up mock configuration
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)
	
	tests := []struct {
		name        string
		prompt      string
		stdin       string
		expectError bool
		expectedPrompt string
	}{
		{
			name:        "no prompt no stdin",
			prompt:      "",
			stdin:       "",
			expectError: true,
		},
		{
			name:        "prompt only",
			prompt:      "test prompt",
			stdin:       "",
			expectError: false,
			expectedPrompt: "test prompt",
		},
		{
			name:        "stdin only",
			prompt:      "",
			stdin:       "stdin content",
			expectError: false,
			expectedPrompt: "stdin content",
		},
		{
			name:        "both prompt and stdin",
			prompt:      "test prompt",
			stdin:       "stdin content",
			expectError: false,
			expectedPrompt: "stdin content\n\ntest prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize config
			err := config.Init()
			require.NoError(t, err)
			
			// Create command registry
			registry := command.NewRegistry()
			
			// Mock the ask command to capture the prompt
			var capturedPrompt string
			testCmd := &mockAskCommand{
				capturePrompt: func(prompt string) {
					capturedPrompt = prompt
				},
			}
			err = registry.Register(testCmd)
			require.NoError(t, err)
			
			// Create test context
			var buf bytes.Buffer
			ctx := &Context{
				Context:  nil, // Kong context not needed for this test
				Registry: registry,
				Config:   config.Manager,
				Logger:   logging.GetLogger(),
				Stdout:   &buf,
				Stderr:   io.Discard,
				Ctx:      context.Background(),
			}

			// Create ask command
			cmd := &AskCmd{
				Prompt: tt.prompt,
			}

			// Mock stdin if needed
			if tt.stdin != "" {
				oldStdin := os.Stdin
				defer func() { os.Stdin = oldStdin }()

				r, w, err := os.Pipe()
				require.NoError(t, err)
				os.Stdin = r

				// Write to pipe in a goroutine
				go func() {
					defer w.Close()
					w.WriteString(tt.stdin)
				}()
			}

			// Run the command
			err = cmd.Run(ctx)

			// Check error expectation
			if tt.expectError {
				assert.Error(t, err)
			} else {
				// The mock command doesn't have a real implementation, so we check the captured prompt
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPrompt, capturedPrompt)
			}
		})
	}
}

// mockAskCommand is a mock implementation of the ask command for testing
type mockAskCommand struct {
	capturePrompt func(string)
}

func (c *mockAskCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if c.capturePrompt != nil && len(exec.Args) > 0 {
		c.capturePrompt(exec.Args[0])
	}
	return nil
}

func (c *mockAskCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "ask",
		Category:    command.CategoryCLI,
		Description: "Test ask command",
	}
}

func (c *mockAskCommand) Validate() error {
	return nil
}