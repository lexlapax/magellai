// ABOUTME: Tests for color integration in help formatter
// ABOUTME: Verifies that color functionality is properly shared between CLI and REPL

package core

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/pkg/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextAwareHelpFormatterColor(t *testing.T) {
	t.Run("color enabled in config", func(t *testing.T) {
		// Initialize config manager
		err := config.Init()
		require.NoError(t, err)

		// Load defaults and set colors enabled
		err = config.Manager.Load(nil)
		require.NoError(t, err)
		require.NoError(t, config.Manager.SetValue("repl.colors.enabled", true))

		cfg := config.Manager

		formatter := NewContextAwareHelpFormatter(cfg)

		// Verify color formatter is created
		assert.NotNil(t, formatter.colorFormatter)

		// Test formatting a simple command
		cmd := &mockCommand{
			meta: &command.Metadata{
				Name:        "test",
				Description: "Test command",
				Aliases:     []string{"t"},
			},
		}

		output := formatter.FormatCommand(cmd)

		// Should contain formatted text (with color codes if terminal is available)
		assert.Contains(t, output, "Command:")
		assert.Contains(t, output, "test")
		assert.Contains(t, output, "aliases:")
		assert.Contains(t, output, "t")
	})

	t.Run("color disabled in config", func(t *testing.T) {
		// Initialize config manager
		err := config.Init()
		require.NoError(t, err)

		// Load defaults and set colors disabled
		err = config.Manager.Load(nil)
		require.NoError(t, err)
		require.NoError(t, config.Manager.SetValue("repl.colors.enabled", false))

		cfg := config.Manager

		formatter := NewContextAwareHelpFormatter(cfg)

		// Verify color formatter exists but is disabled
		assert.NotNil(t, formatter.colorFormatter)

		// Test formatting - should not contain ANSI codes
		cmd := &mockCommand{
			meta: &command.Metadata{
				Name:        "test",
				Description: "Test command",
			},
		}

		output := formatter.FormatCommand(cmd)

		// Should not contain ANSI escape codes
		assert.NotContains(t, output, "\033[")
	})

	t.Run("error formatting with color", func(t *testing.T) {
		// Initialize config manager
		err := config.Init()
		require.NoError(t, err)

		// Load defaults and set colors enabled
		err = config.Manager.Load(nil)
		require.NoError(t, err)
		require.NoError(t, config.Manager.SetValue("repl.colors.enabled", true))

		cfg := config.Manager

		formatter := NewContextAwareHelpFormatter(cfg)
		formatter.Category = command.CategoryREPL

		err = fmt.Errorf("unknown command")
		output := formatter.FormatError(err, "help")

		// Should contain error and suggestion
		assert.Contains(t, output, "Error:")
		assert.Contains(t, output, "unknown command")
		assert.Contains(t, output, "Did you mean:")
		assert.Contains(t, output, "help")
		assert.Contains(t, output, "/help")
	})

	t.Run("flag formatting with color", func(t *testing.T) {
		// Initialize config manager
		err := config.Init()
		require.NoError(t, err)

		// Load defaults and set colors enabled
		err = config.Manager.Load(nil)
		require.NoError(t, err)
		require.NoError(t, config.Manager.SetValue("repl.colors.enabled", true))

		cfg := config.Manager

		formatter := NewContextAwareHelpFormatter(cfg)

		flags := []command.Flag{
			{
				Name:        "verbose",
				Short:       "v",
				Description: "Enable verbose output",
				Type:        command.FlagTypeBool,
				Default:     false,
			},
			{
				Name:        "output",
				Short:       "o",
				Description: "Output file",
				Type:        command.FlagTypeString,
				Required:    true,
			},
		}

		output := formatter.formatFlags(flags)

		// Should contain flag information
		assert.Contains(t, output, "--verbose")
		assert.Contains(t, output, "-v")
		assert.Contains(t, output, "Enable verbose output")
		assert.Contains(t, output, "(default: false)")
		assert.Contains(t, output, "--output")
		assert.Contains(t, output, "(required)")
	})

	t.Run("command list formatting with color", func(t *testing.T) {
		// Initialize config manager
		err := config.Init()
		require.NoError(t, err)

		// Load defaults and set colors enabled
		err = config.Manager.Load(nil)
		require.NoError(t, err)
		require.NoError(t, config.Manager.SetValue("repl.colors.enabled", true))

		cfg := config.Manager

		formatter := NewContextAwareHelpFormatter(cfg)
		formatter.Category = command.CategoryCLI

		commands := []command.Interface{
			&mockCommand{
				meta: &command.Metadata{
					Name:        "ask",
					Description: "Ask a question",
					Category:    command.CategoryShared,
				},
			},
			&mockCommand{
				meta: &command.Metadata{
					Name:        "chat",
					Description: "Start a chat",
					Category:    command.CategoryCLI,
				},
			},
		}

		output := formatter.FormatCommandList(commands, command.CategoryCLI)

		// Should contain proper headers and commands
		assert.Contains(t, output, "CLI Commands:")
		assert.Contains(t, output, "magellai <command> --help")
		assert.Contains(t, output, "ask")
		assert.Contains(t, output, "chat")
	})
}

func TestHelpCommandColorIntegration(t *testing.T) {
	t.Run("REPL help command with colors", func(t *testing.T) {
		// Initialize config manager
		err := config.Init()
		require.NoError(t, err)

		// Load defaults and set colors enabled
		err = config.Manager.Load(nil)
		require.NoError(t, err)
		require.NoError(t, config.Manager.SetValue("repl.colors.enabled", true))

		cfg := config.Manager

		registry := command.NewRegistry()
		require.NoError(t, registry.Register(&mockCommand{
			meta: &command.Metadata{
				Name:        "test",
				Description: "Test command",
				Category:    command.CategoryREPL,
			},
		}))

		helpCmd := NewHelpCommand(registry, cfg)
		helpCmd.category = command.CategoryREPL

		// Execute help command
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		ctx := context.Background()
		execCtx := &command.ExecutionContext{
			Stdout: stdout,
			Stderr: stderr,
			Args:   []string{},
			Flags:  command.NewFlags(nil),
		}

		err = helpCmd.Execute(ctx, execCtx)
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "REPL Commands:")
		assert.Contains(t, output, "/help")
	})

	t.Run("CLI help command with colors", func(t *testing.T) {
		// Initialize config manager
		err := config.Init()
		require.NoError(t, err)

		// Load defaults and set colors enabled
		err = config.Manager.Load(nil)
		require.NoError(t, err)
		require.NoError(t, config.Manager.SetValue("repl.colors.enabled", true))

		cfg := config.Manager

		registry := command.NewRegistry()
		require.NoError(t, registry.Register(&mockCommand{
			meta: &command.Metadata{
				Name:        "version",
				Description: "Show version",
				Category:    command.CategoryCLI,
			},
		}))

		helpCmd := NewHelpCommand(registry, cfg)
		helpCmd.category = command.CategoryCLI

		// Execute help command
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		ctx := context.Background()
		execCtx := &command.ExecutionContext{
			Stdout: stdout,
			Stderr: stderr,
			Args:   []string{},
			Flags:  command.NewFlags(nil),
		}

		err = helpCmd.Execute(ctx, execCtx)
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "CLI Commands:")
		assert.Contains(t, output, "magellai")
	})
}

func TestHelpColorStripping(t *testing.T) {
	t.Run("strip colors when piped", func(t *testing.T) {
		// This test simulates a non-TTY environment where colors should be stripped
		text := "\033[0;32mHello\033[0m World"
		stripped := ui.StripColors(text)
		assert.Equal(t, "Hello World", stripped)
	})
}

// mockCommand is a test implementation of command.Interface
type mockCommand struct {
	meta        *command.Metadata
	executeFunc func(ctx context.Context, exec *command.ExecutionContext) error
}

func (m *mockCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, exec)
	}
	return nil
}

func (m *mockCommand) Metadata() *command.Metadata {
	return m.meta
}

func (m *mockCommand) Validate() error {
	return nil
}
