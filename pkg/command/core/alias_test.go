// ABOUTME: Unit tests for the alias command, covering all subcommands and edge cases
// ABOUTME: Tests add, remove, list, show, clear, export/import functionality

package core

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliasCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]interface{}
		setupConfig    func(*config.Config)
		expectedOutput string
		expectedError  string
		outputFormat   string
	}{
		// Basic commands
		{
			name:           "list aliases - with default",
			args:           []string{},
			expectedOutput: "Defined aliases:",
		},
		{
			name: "list aliases - with aliases",
			args: []string{"list"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
				require.NoError(t, c.SetValue("aliases.claude", "model anthropic/claude-3"))
			},
			expectedOutput: "gpt4",
		},
		{
			name:  "list aliases - JSON format",
			args:  []string{"list"},
			flags: map[string]interface{}{"format": "json"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: `"gpt4": "model gpt-4"`,
		},
		{
			name:  "list aliases - REPL scope",
			args:  []string{"list"},
			flags: map[string]interface{}{"scope": "repl"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("repl.aliases.h", "help"))
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: "h (repl)",
		},

		// Add command
		{
			name:           "add alias",
			args:           []string{"add", "gpt4", "model", "gpt-4"},
			expectedOutput: "Alias 'gpt4' created: model gpt-4",
		},
		{
			name:           "add alias with multiple words",
			args:           []string{"add", "fast", "model", "gpt-3.5-turbo", "--temperature", "0.1"},
			expectedOutput: "Alias 'fast' created: model gpt-3.5-turbo --temperature 0.1",
		},
		{
			name:          "add alias - missing command",
			args:          []string{"add", "gpt4"},
			expectedError: "missing argument - name and command required",
		},
		{
			name:          "add alias - reserved name",
			args:          []string{"add", "config", "some command"},
			expectedError: "cannot override reserved command: config",
		},
		{
			name:          "add alias - name with spaces",
			args:          []string{"add", "my alias", "some command"},
			expectedError: "alias name cannot contain spaces",
		},
		{
			name:           "add REPL alias",
			args:           []string{"add", "h", "help"},
			flags:          map[string]interface{}{"scope": "repl"},
			expectedOutput: "Alias 'h' created: help",
		},

		// Show command
		{
			name: "show alias",
			args: []string{"show", "gpt4"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: "gpt4 → model gpt-4",
		},
		{
			name: "show REPL alias",
			args: []string{"show", "r"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("repl.aliases.r", "reload"))
			},
			expectedOutput: "r (repl) → reload",
		},
		{
			name:          "show non-existent alias",
			args:          []string{"show", "nonexistent"},
			expectedError: "alias 'nonexistent' not found",
		},
		{
			name:          "show without name",
			args:          []string{"show"},
			expectedError: "missing argument - name required",
		},

		// Remove command
		{
			name: "remove alias",
			args: []string{"remove", "gpt4"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: "Alias 'gpt4' removed",
		},
		{
			name:  "remove REPL alias",
			args:  []string{"remove", "r"},
			flags: map[string]interface{}{"scope": "repl"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("repl.aliases.r", "reload"))
			},
			expectedOutput: "Alias 'r' removed",
		},
		{
			name:          "remove non-existent alias",
			args:          []string{"remove", "nonexistent"},
			expectedError: "alias 'nonexistent' not found",
		},
		{
			name:          "remove without name",
			args:          []string{"remove"},
			expectedError: "missing argument - name required",
		},

		// Clear command
		{
			name: "clear all aliases",
			args: []string{"clear"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
				require.NoError(t, c.SetValue("aliases.claude", "model claude-3"))
				require.NoError(t, c.SetValue("repl.aliases.h", "help"))
			},
			expectedOutput: "Cleared",
		},
		{
			name:  "clear CLI aliases only",
			args:  []string{"clear"},
			flags: map[string]interface{}{"scope": "cli"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
				require.NoError(t, c.SetValue("repl.aliases.h", "help"))
			},
			expectedOutput: "Cleared",
		},

		// Export command
		{
			name: "export aliases",
			args: []string{"export"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: `"cli": {`,
		},
		{
			name:  "export REPL aliases only",
			args:  []string{"export"},
			flags: map[string]interface{}{"scope": "repl"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("repl.aliases.h", "help"))
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: `"repl": {`,
		},

		// Import command
		{
			name:          "import aliases",
			args:          []string{"import", "aliases.json"},
			expectedError: "alias import not implemented",
		},
		{
			name:          "import without filename",
			args:          []string{"import"},
			expectedError: "missing argument - filename required",
		},

		// Direct alias name (implicit show)
		{
			name: "direct alias name",
			args: []string{"gpt4"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: "gpt4 → model gpt-4",
		},

		// Alternative command names
		{
			name:           "set alias (alternative to add)",
			args:           []string{"set", "gpt4", "model", "gpt-4"},
			expectedOutput: "Alias 'gpt4' created: model gpt-4",
		},
		{
			name: "delete alias (alternative to remove)",
			args: []string{"delete", "gpt4"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: "Alias 'gpt4' removed",
		},
		{
			name: "rm alias (alternative to remove)",
			args: []string{"rm", "gpt4"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: "Alias 'gpt4' removed",
		},
		{
			name: "get alias (alternative to show)",
			args: []string{"get", "gpt4"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("aliases.gpt4", "model gpt-4"))
			},
			expectedOutput: "gpt4 → model gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test config
			cfg := createTestConfig(t)
			if tt.setupConfig != nil {
				tt.setupConfig(cfg)
			}

			cmd := NewAliasCommand(cfg)

			ctx := context.Background()
			var stdout, stderr bytes.Buffer
			exec := &command.ExecutionContext{
				Args:   tt.args,
				Flags:  command.NewFlags(tt.flags),
				Stdout: &stdout,
				Stderr: &stderr,
				Data:   make(map[string]interface{}),
			}

			if tt.outputFormat != "" {
				exec.Data["outputFormat"] = tt.outputFormat
			}

			err := cmd.Execute(ctx, exec)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				output, ok := exec.Data["output"].(string)
				require.True(t, ok)
				if tt.name == "show REPL alias" {
					t.Logf("Actual output: %q", output)
					t.Logf("Expected output: %q", tt.expectedOutput)
				}
				assert.Contains(t, output, tt.expectedOutput)
			}
		})
	}
}

func TestAliasCommand_Metadata(t *testing.T) {
	cmd := NewAliasCommand(nil)
	meta := cmd.Metadata()

	assert.Equal(t, "alias", meta.Name)
	assert.Contains(t, meta.Aliases, "aliases")
	assert.Equal(t, command.CategoryShared, meta.Category)
	assert.NotEmpty(t, meta.Description)
	assert.NotEmpty(t, meta.LongDescription)
	assert.Contains(t, meta.LongDescription, "list")
	assert.Contains(t, meta.LongDescription, "add")
	assert.Contains(t, meta.LongDescription, "remove")
	assert.Contains(t, meta.LongDescription, "show")
	assert.Contains(t, meta.LongDescription, "clear")
	assert.Contains(t, meta.LongDescription, "export")
	assert.Contains(t, meta.LongDescription, "import")

	// Check flags
	assert.Len(t, meta.Flags, 2)

	// Check format flag
	formatFlag := meta.Flags[0]
	assert.Equal(t, "format", formatFlag.Name)
	assert.Equal(t, "f", formatFlag.Short)
	assert.Equal(t, command.FlagTypeString, formatFlag.Type)
	assert.Equal(t, "text", formatFlag.Default)

	// Check scope flag
	scopeFlag := meta.Flags[1]
	assert.Equal(t, "scope", scopeFlag.Name)
	assert.Equal(t, "s", scopeFlag.Short)
	assert.Equal(t, command.FlagTypeString, scopeFlag.Type)
	assert.Equal(t, "all", scopeFlag.Default)
}

func TestAliasCommand_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedError string
	}{
		{
			name:   "valid with config",
			config: createTestConfig(t),
		},
		{
			name:          "invalid without config",
			config:        nil,
			expectedError: "config manager not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewAliasCommand(tt.config)
			err := cmd.Validate()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAliasCommand_EmptyAliases(t *testing.T) {
	t.Skip("Skipping test - defaults now include built-in aliases")
}

func TestAliasCommand_CompleteLifecycle(t *testing.T) {
	cfg := createTestConfig(t)
	cmd := NewAliasCommand(cfg)
	ctx := context.Background()

	// 1. Add some aliases
	exec := &command.ExecutionContext{
		Args:  []string{"add", "gpt4", "model", "gpt-4"},
		Flags: command.NewFlags(nil),
		Data:  make(map[string]interface{}),
	}
	err := cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "Alias 'gpt4' created")

	exec = &command.ExecutionContext{
		Args:  []string{"add", "claude", "model", "anthropic/claude-3"},
		Flags: command.NewFlags(nil),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)

	// 2. Add a REPL alias
	exec = &command.ExecutionContext{
		Args:  []string{"add", "h", "help"},
		Flags: command.NewFlags(map[string]interface{}{"scope": "repl"}),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)

	// 3. List all aliases
	exec = &command.ExecutionContext{
		Args:  []string{"list"},
		Flags: command.NewFlags(nil),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output := exec.Data["output"].(string)
	assert.Contains(t, output, "gpt4")
	assert.Contains(t, output, "claude")
	assert.Contains(t, output, "h (repl)")

	// 4. Show a specific alias
	exec = &command.ExecutionContext{
		Args:  []string{"show", "gpt4"},
		Flags: command.NewFlags(nil),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "gpt4 → model gpt-4")

	// 5. Export aliases
	exec = &command.ExecutionContext{
		Args:  []string{"export"},
		Flags: command.NewFlags(nil),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output = exec.Data["output"].(string)
	assert.Contains(t, output, "cli")
	assert.Contains(t, output, "repl")

	// 6. Remove an alias
	exec = &command.ExecutionContext{
		Args:  []string{"remove", "claude"},
		Flags: command.NewFlags(nil),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "Alias 'claude' removed")

	// 7. Clear REPL aliases
	exec = &command.ExecutionContext{
		Args:  []string{"clear"},
		Flags: command.NewFlags(map[string]interface{}{"scope": "repl"}),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "Cleared")

	// 8. Verify final state
	exec = &command.ExecutionContext{
		Args:  []string{"list"},
		Flags: command.NewFlags(nil),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output = exec.Data["output"].(string)
	assert.Contains(t, output, "gpt4")
	// Check that the specific aliases we created and removed don't exist anymore
	// The format is "aliasname → command" or "aliasname (repl) → command"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Check that our removed aliases don't exist as alias names
		assert.NotRegexp(t, `^\s*claude\s*(→|\(repl\)\s*→)`, line)
		assert.NotRegexp(t, `^\s*h\s*\(repl\)\s*→`, line)
	}
	// Also check that gpt4 still exists
	assert.Regexp(t, `gpt4\s*→\s*model gpt-4`, output)
}

func TestAliasCommand_ScopeHandling(t *testing.T) {
	// Use clean config to avoid pre-existing aliases
	config.Manager = nil
	err := config.Init()
	require.NoError(t, err)

	cfg := config.Manager
	cmd := NewAliasCommand(cfg)
	ctx := context.Background()

	// Add CLI alias
	exec := &command.ExecutionContext{
		Args:  []string{"add", "gpt4", "model", "gpt-4"},
		Flags: command.NewFlags(map[string]interface{}{"scope": "cli"}),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)

	// Add REPL alias with same name but different command
	exec = &command.ExecutionContext{
		Args:  []string{"add", "gpt4", "model", "gpt-4-turbo"},
		Flags: command.NewFlags(map[string]interface{}{"scope": "repl"}),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)

	// List CLI aliases only - should contain the CLI alias we added
	exec = &command.ExecutionContext{
		Args:  []string{"list"},
		Flags: command.NewFlags(map[string]interface{}{"scope": "cli"}),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output := exec.Data["output"].(string)
	assert.Contains(t, output, "gpt4")
	assert.Contains(t, output, "model gpt-4")
	assert.NotContains(t, output, "(repl)")
	assert.NotContains(t, output, "gpt-4-turbo")

	// List REPL aliases only
	exec = &command.ExecutionContext{
		Args:  []string{"list"},
		Flags: command.NewFlags(map[string]interface{}{"scope": "repl"}),
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output = exec.Data["output"].(string)
	assert.Contains(t, output, "gpt4 (repl)")
	assert.Contains(t, output, "model gpt-4-turbo")
	// The CLI alias shouldn't appear when listing REPL only
	assert.NotContains(t, output, "model gpt-4 ") // Add space to avoid matching substring in gpt-4-turbo
}
