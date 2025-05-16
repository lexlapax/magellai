// ABOUTME: Unit tests for the context-aware help command
// ABOUTME: Tests help functionality, formatting, and context awareness

package core

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleTestCommand is a minimal command for testing
type SimpleTestCommand struct {
	name        string
	description string
	category    command.Category
	aliases     []string
	flags       []command.Flag
	hidden      bool
}

func (c *SimpleTestCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	return nil
}

func (c *SimpleTestCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        c.name,
		Description: c.description,
		Category:    c.category,
		Aliases:     c.aliases,
		Flags:       c.flags,
		Hidden:      c.hidden,
	}
}

func (c *SimpleTestCommand) Validate() error {
	return nil
}

func TestHelpCommand_Basic(t *testing.T) {
	// Initialize config
	require.NoError(t, config.Init())
	cfg := config.Manager
	require.NoError(t, cfg.Load(nil))

	// Create registry
	registry := command.NewRegistry()

	// Register test commands
	testCmd := &SimpleTestCommand{
		name:        "test",
		description: "Test command",
		category:    command.CategoryShared,
		aliases:     []string{"t"},
	}
	require.NoError(t, registry.Register(testCmd))

	// Create help command
	helpCmd := NewHelpCommand(registry, cfg)

	// Test basic help listing
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exec := &command.ExecutionContext{
		Args:   []string{},
		Flags:  command.NewFlags(map[string]interface{}{}),
		Stdout: stdout,
		Stderr: stderr,
	}

	ctx := context.Background()
	err := helpCmd.Execute(ctx, exec)
	require.NoError(t, err)

	// Check output contains expected content
	output := stdout.String()
	assert.Contains(t, output, "Available Commands")
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "Test command")
}

func TestHelpCommand_SpecificCommand(t *testing.T) {
	// Initialize config
	require.NoError(t, config.Init())
	cfg := config.Manager
	require.NoError(t, cfg.Load(nil))

	// Create registry
	registry := command.NewRegistry()

	// Register test command with flags
	testCmd := &SimpleTestCommand{
		name:        "test",
		description: "Test command",
		category:    command.CategoryShared,
		aliases:     []string{"t"},
		flags: []command.Flag{
			{
				Name:        "verbose",
				Short:       "v",
				Description: "Enable verbose output",
				Type:        command.FlagTypeBool,
				Default:     false,
			},
		},
	}
	require.NoError(t, registry.Register(testCmd))

	// Create help command
	helpCmd := NewHelpCommand(registry, cfg)

	// Test specific command help
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exec := &command.ExecutionContext{
		Args:   []string{"test"},
		Flags:  command.NewFlags(map[string]interface{}{}),
		Stdout: stdout,
		Stderr: stderr,
	}

	ctx := context.Background()
	err := helpCmd.Execute(ctx, exec)
	require.NoError(t, err)

	// Check output contains expected content
	output := stdout.String()
	assert.Contains(t, output, "Command: test")
	assert.Contains(t, output, "Description: Test command")
	assert.Contains(t, output, "aliases: t")
	assert.Contains(t, output, "--verbose")
	assert.Contains(t, output, "Enable verbose output")
}

func TestHelpCommand_ContextAwareness(t *testing.T) {
	// Initialize config
	require.NoError(t, config.Init())
	cfg := config.Manager
	require.NoError(t, cfg.Load(nil))

	// Create registry
	registry := command.NewRegistry()

	// Register commands for different categories
	sharedCmd := &SimpleTestCommand{
		name:        "shared",
		description: "Shared command",
		category:    command.CategoryShared,
	}
	cliCmd := &SimpleTestCommand{
		name:        "cli-only",
		description: "CLI only command",
		category:    command.CategoryCLI,
	}
	replCmd := &SimpleTestCommand{
		name:        "repl-only",
		description: "REPL only command",
		category:    command.CategoryREPL,
	}

	require.NoError(t, registry.Register(sharedCmd))
	require.NoError(t, registry.Register(cliCmd))
	require.NoError(t, registry.Register(replCmd))

	tests := []struct {
		name        string
		category    command.Category
		expected    []string
		notExpected []string
	}{
		{
			name:        "CLI context",
			category:    command.CategoryCLI,
			expected:    []string{"CLI Commands", "shared", "cli-only", "Use 'magellai <command> --help'"},
			notExpected: []string{"repl-only"},
		},
		{
			name:        "REPL context",
			category:    command.CategoryREPL,
			expected:    []string{"REPL Commands", "shared", "repl-only", "Use '/help <command>'"},
			notExpected: []string{"cli-only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create help command with specific category
			helpCmd := NewHelpCommand(registry, cfg)
			helpCmd.category = tt.category

			// Execute
			stdout := &bytes.Buffer{}
			exec := &command.ExecutionContext{
				Args:   []string{},
				Flags:  command.NewFlags(map[string]interface{}{}),
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
			}

			ctx := context.Background()
			require.NoError(t, helpCmd.Execute(ctx, exec))

			// Verify context-specific output
			output := stdout.String()
			for _, expected := range tt.expected {
				assert.Contains(t, output, expected)
			}
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestHelpCommand_CommandNotFound(t *testing.T) {
	// Initialize config
	require.NoError(t, config.Init())
	cfg := config.Manager
	require.NoError(t, cfg.Load(nil))

	// Create registry with one command
	registry := command.NewRegistry()
	testCmd := &SimpleTestCommand{name: "test", description: "Test command", category: command.CategoryShared}
	require.NoError(t, registry.Register(testCmd))

	// Create help command
	helpCmd := NewHelpCommand(registry, cfg)

	// Test non-existent command
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exec := &command.ExecutionContext{
		Args:   []string{"nonexistent"},
		Flags:  command.NewFlags(map[string]interface{}{}),
		Stdout: stdout,
		Stderr: stderr,
	}

	ctx := context.Background()
	err := helpCmd.Execute(ctx, exec)
	assert.Error(t, err)

	// Check error output contains suggestion
	errorOutput := stderr.String()
	assert.Contains(t, errorOutput, "Error:")
	assert.Contains(t, errorOutput, "command not found")
}

func TestHelpCommand_Flags(t *testing.T) {
	// Initialize config
	require.NoError(t, config.Init())
	cfg := config.Manager
	require.NoError(t, cfg.Load(nil))

	// Create registry
	registry := command.NewRegistry()

	// Register test command with aliases
	testCmd := &SimpleTestCommand{
		name:        "test",
		description: "Test command",
		category:    command.CategoryShared,
		aliases:     []string{"t", "tst"},
	}
	require.NoError(t, registry.Register(testCmd))

	// Register hidden command
	hiddenCmd := &SimpleTestCommand{
		name:        "hidden",
		description: "Hidden command",
		category:    command.CategoryShared,
		hidden:      true,
	}
	require.NoError(t, registry.Register(hiddenCmd))

	// Create help command
	helpCmd := NewHelpCommand(registry, cfg)

	tests := []struct {
		name       string
		flags      map[string]interface{}
		shouldShow []string
		shouldHide []string
	}{
		{
			name:       "default settings",
			flags:      map[string]interface{}{},
			shouldShow: []string{"test", "(t, tst)"},
			shouldHide: []string{"hidden"},
		},
		{
			name:       "show all including hidden",
			flags:      map[string]interface{}{"all": true},
			shouldShow: []string{"test", "hidden"},
			shouldHide: []string{},
		},
		{
			name:       "hide aliases",
			flags:      map[string]interface{}{"no-aliases": true},
			shouldShow: []string{"test"},
			shouldHide: []string{"(t, tst)", "hidden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			exec := &command.ExecutionContext{
				Args:   []string{},
				Flags:  command.NewFlags(tt.flags),
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
			}

			ctx := context.Background()
			require.NoError(t, helpCmd.Execute(ctx, exec))

			output := stdout.String()
			for _, should := range tt.shouldShow {
				assert.Contains(t, output, should)
			}
			for _, hide := range tt.shouldHide {
				assert.NotContains(t, output, hide)
			}
		})
	}
}

func TestHelpCommand_Metadata(t *testing.T) {
	cmd := &HelpCommand{}
	meta := cmd.Metadata()

	assert.Equal(t, "help", meta.Name)
	assert.Contains(t, meta.Aliases, "h")
	assert.Contains(t, meta.Aliases, "?")
	assert.Equal(t, command.CategoryShared, meta.Category)
	assert.NotEmpty(t, meta.Description)
	assert.Len(t, meta.Flags, 2)
}

func TestHelpCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *HelpCommand
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing registry",
			cmd:     &HelpCommand{formatter: &ContextAwareHelpFormatter{}, config: &config.Config{}},
			wantErr: true,
			errMsg:  "requires a registry",
		},
		{
			name:    "missing formatter",
			cmd:     &HelpCommand{registry: command.NewRegistry(), config: &config.Config{}},
			wantErr: true,
			errMsg:  "requires a formatter",
		},
		{
			name:    "missing config",
			cmd:     &HelpCommand{registry: command.NewRegistry(), formatter: &ContextAwareHelpFormatter{}},
			wantErr: true,
			errMsg:  "requires a config",
		},
		{
			name: "valid command",
			cmd: &HelpCommand{
				registry:  command.NewRegistry(),
				formatter: &ContextAwareHelpFormatter{},
				config:    &config.Config{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHelpFormatter(t *testing.T) {
	formatter := NewContextAwareHelpFormatter(&config.Config{})

	// Test command formatting
	cmd := &SimpleTestCommand{
		name:        "test",
		description: "Test command",
		category:    command.CategoryShared,
		aliases:     []string{"t"},
		flags: []command.Flag{
			{
				Name:        "verbose",
				Short:       "v",
				Description: "Enable verbose output",
				Type:        command.FlagTypeBool,
				Default:     false,
			},
		},
	}

	output := formatter.FormatCommand(cmd)
	assert.Contains(t, output, "Command: test")
	assert.Contains(t, output, "aliases: t")
	assert.Contains(t, output, "Description: Test command")
	assert.Contains(t, output, "--verbose")

	// Test error formatting
	err := fmt.Errorf("command not found: unknown")
	errorOutput := formatter.FormatError(err, "test")
	assert.Contains(t, errorOutput, "Error: command not found: unknown")
	assert.Contains(t, errorOutput, "Did you mean: test?")
}

func TestHelpCommand_AliasResolution(t *testing.T) {
	// Initialize config
	require.NoError(t, config.Init())
	cfg := config.Manager
	require.NoError(t, cfg.Load(nil))

	// Add an alias
	err := cfg.SetValue("aliases.t", "test")
	require.NoError(t, err)

	// Create registry
	registry := command.NewRegistry()
	testCmd := &SimpleTestCommand{
		name:        "test",
		description: "Test command",
		category:    command.CategoryShared,
	}
	require.NoError(t, registry.Register(testCmd))

	// Create help command
	helpCmd := NewHelpCommand(registry, cfg)

	// Test alias resolution
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exec := &command.ExecutionContext{
		Args:   []string{"t"},
		Flags:  command.NewFlags(map[string]interface{}{}),
		Stdout: stdout,
		Stderr: stderr,
	}

	ctx := context.Background()
	execErr := helpCmd.Execute(ctx, exec)
	require.NoError(t, execErr)

	// Should show help for the test command
	output := stdout.String()
	assert.Contains(t, output, "Command: test")
}
