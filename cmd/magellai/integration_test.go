// ABOUTME: Integration tests for the main CLI application
// ABOUTME: Tests actual command execution and behavior

//go:build integration
// +build integration

package main

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/command/core"
	"github.com/lexlapax/magellai/pkg/config"
)

func TestIntegration_VersionCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedJSON   bool
	}{
		{
			name:           "version command text",
			args:           []string{"version"},
			expectedOutput: "magellai version dev",
			expectedJSON:   false,
		},
		{
			name:           "version command json",
			args:           []string{"version", "-o", "json"},
			expectedOutput: `"version": "dev"`,
			expectedJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would normally use exec.Command to run the built binary
			// For unit testing, we'll test the integration of components

			// Create a CLI parser
			var cli CLI
			parser := kong.Must(&cli)

			var stdout, stderr bytes.Buffer

			// Initialize logger
			require.NoError(t, logging.Initialize(logging.LogConfig{
				Level:      "error",
				Format:     "text",
				OutputPath: "stderr",
			}))

			// Initialize config
			require.NoError(t, config.Init())

			// Create registry and register commands
			registry := command.NewRegistry()

			// Register version command
			versionCmd := core.NewVersionCommand("dev", "none", "unknown")
			require.NoError(t, registry.Register(versionCmd))

			// Parse arguments
			ctx, err := parser.Parse(tt.args)
			require.NoError(t, err)

			// Create context
			testCtx := &Context{
				Context:  ctx,
				Registry: registry,
				Config:   config.Manager,
				Logger:   logging.GetLogger(),
				Stdout:   &stdout,
				Stderr:   &stderr,
				Ctx:      context.Background(),
			}

			// Run command
			err = ctx.Run(testCtx)
			require.NoError(t, err)

			// Check output
			output := stdout.String()
			if tt.expectedJSON {
				assert.Contains(t, output, tt.expectedOutput)
				assert.Contains(t, output, "\"version\"")
			} else {
				assert.Contains(t, output, tt.expectedOutput)
			}
		})
	}
}

// TestIntegration_BasicE2E tests basic end-to-end functionality
func TestIntegration_BasicE2E(t *testing.T) {
	// Test version command directly
	cmd := exec.Command("go", "run", "./cmd/magellai", "version")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "magellai version")

	// Test help
	cmd = exec.Command("go", "run", "./cmd/magellai", "--help")
	cmd.Dir = "../.."
	output, err = cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Usage:")
}

// TestMain_E2E tests the actual main function
// This is closer to a true integration test
func TestMain_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "test-magellai", "./cmd/magellai")
	cmd.Dir = "../.."
	err := cmd.Run()
	require.NoError(t, err)
	defer func() {
		_ = exec.Command("rm", "test-magellai").Run()
	}()

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  bool
	}{
		{
			name:           "version flag",
			args:           []string{"--version"},
			expectedOutput: "magellai version",
		},
		{
			name:           "version command",
			args:           []string{"version"},
			expectedOutput: "magellai version",
		},
		{
			name:           "help",
			args:           []string{"--help"},
			expectedOutput: "Usage:",
		},
		{
			name:          "invalid command",
			args:          []string{"invalid"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./test-magellai", tt.args...)
			cmd.Dir = "../.."

			output, err := cmd.CombinedOutput()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, string(output), tt.expectedOutput)
			}
		})
	}
}

// TestCLI_CommandRegistration verifies all core commands are registered
func TestCLI_CommandRegistration(t *testing.T) {
	// Initialize environment
	require.NoError(t, logging.Initialize(logging.LogConfig{
		Level:      "error",
		Format:     "text",
		OutputPath: "stderr",
	}))

	require.NoError(t, config.Init())
	cfg := config.Manager

	// Create registry
	registry := command.NewRegistry()

	// Register all core commands (mirrors main.go)
	commands := []command.Interface{
		core.NewConfigCommand(cfg),
		core.NewProfileCommand(cfg),
		core.NewModelCommand(cfg),
		core.NewAliasCommand(cfg),
		core.NewHelpCommand(registry, cfg),
		core.NewVersionCommand("dev", "none", "unknown"),
	}

	for _, cmd := range commands {
		meta := cmd.Metadata()
		t.Run(meta.Name, func(t *testing.T) {
			err := registry.Register(cmd)
			assert.NoError(t, err, "Failed to register %s command", meta.Name)

			// Verify command is registered
			registered, err := registry.Get(meta.Name)
			assert.NoError(t, err)
			assert.NotNil(t, registered)
			assert.Equal(t, meta.Name, registered.Metadata().Name)
		})
	}

	// Verify stub commands can be registered
	t.Run("stub commands", func(t *testing.T) {
		err := RegisterStubCommands(registry)
		assert.NoError(t, err)

		// Check chat command (ask is no longer a stub)
		chat, err := registry.Get("chat")
		assert.NoError(t, err)
		assert.NotNil(t, chat)
	})
}
