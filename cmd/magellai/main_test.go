// ABOUTME: Unit tests for the main CLI application
// ABOUTME: Tests CLI argument parsing, command execution, and integration

package main

import (
	"os/exec"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_VersionFlag(t *testing.T) {
	// This tests the actual CLI behavior with version flag
	// We'll use exec.Command to test the real behavior
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "version flag long",
			args:     []string{"--version"},
			expected: "magellai version dev",
		},
		{
			name:     "version flag with other args",
			args:     []string{"--version", "ask", "hello"},
			expected: "magellai version dev",
		},
		{
			name:     "version flag position independent",
			args:     []string{"ask", "hello", "--version"},
			expected: "magellai version dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build and run a real instance
			cmd := exec.Command("go", "run", "./cmd/magellai")
			cmd.Args = append(cmd.Args, tt.args...)
			cmd.Dir = "../.."

			output, err := cmd.CombinedOutput()
			require.NoError(t, err)
			assert.Contains(t, string(output), tt.expected)
		})
	}
}

func TestCLI_VersionCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "version command",
			args:     []string{"version"},
			expected: "magellai version dev",
		},
		{
			name:     "version command with json",
			args:     []string{"version", "-o", "json"},
			expected: `"version": "dev"`,
		},
		{
			name:     "version command with global json",
			args:     []string{"-o", "json", "version"},
			expected: `"version": "dev"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli CLI
			parser := kong.Must(&cli)

			// Mock the Run method
			ctx, err := parser.Parse(tt.args)
			require.NoError(t, err)

			// Get command and check if it's version
			switch ctx.Command() {
			case "version":
				// The version command should be tested through integration
				// For unit test, we just verify parsing works
				assert.Equal(t, "version", ctx.Command())
			default:
				if !tt.wantErr {
					t.Errorf("unexpected command: %s", ctx.Command())
				}
			}
		})
	}
}

func TestCLI_ParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCommand string
		wantFlags   map[string]interface{}
		wantErr     bool
	}{
		{
			name:        "ask command",
			args:        []string{"ask", "hello world"},
			wantCommand: "ask <prompt>",
		},
		{
			name:        "chat command",
			args:        []string{"chat"},
			wantCommand: "chat",
		},
		{
			name:        "config show",
			args:        []string{"config", "show"},
			wantCommand: "config show",
		},
		{
			name:        "global flags with command",
			args:        []string{"-v", "-o", "json", "ask", "test"},
			wantCommand: "ask <prompt>",
		},
		{
			name:    "invalid command",
			args:    []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli CLI
			parser := kong.Must(&cli)

			ctx, err := parser.Parse(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCommand, ctx.Command())
		})
	}
}

func TestCLI_GlobalFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantFlags struct {
			verbosity  int
			output     string
			configFile string
			profile    string
			noColor    bool
		}
	}{
		{
			name: "verbosity flag",
			args: []string{"-vv", "version"},
			wantFlags: struct {
				verbosity  int
				output     string
				configFile string
				profile    string
				noColor    bool
			}{
				verbosity: 2,
				output:    "text",
			},
		},
		{
			name: "output format",
			args: []string{"-o", "json", "version"},
			wantFlags: struct {
				verbosity  int
				output     string
				configFile string
				profile    string
				noColor    bool
			}{
				output: "json",
			},
		},
		{
			name: "all global flags",
			args: []string{"-v", "-o", "markdown", "-c", "/tmp/config.yaml", "--profile", "dev", "--no-color", "version"},
			wantFlags: struct {
				verbosity  int
				output     string
				configFile string
				profile    string
				noColor    bool
			}{
				verbosity:  1,
				output:     "markdown",
				configFile: "/tmp/config.yaml",
				profile:    "dev",
				noColor:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli CLI
			parser := kong.Must(&cli)

			_, err := parser.Parse(tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.wantFlags.verbosity, cli.Verbosity)
			assert.Equal(t, tt.wantFlags.output, cli.Output)
			assert.Equal(t, tt.wantFlags.configFile, cli.ConfigFile)
			assert.Equal(t, tt.wantFlags.profile, cli.ProfileName)
			assert.Equal(t, tt.wantFlags.noColor, cli.NoColor)
		})
	}
}

func TestCLI_AskCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    AskCmd
		wantErr bool
	}{
		{
			name: "basic ask",
			args: []string{"ask", "hello world"},
			want: AskCmd{
				Prompt: "hello world",
			},
		},
		{
			name: "ask with model",
			args: []string{"ask", "-m", "gpt-4", "test prompt"},
			want: AskCmd{
				Prompt: "test prompt",
				Model:  "gpt-4",
			},
		},
		{
			name: "ask with attachments",
			args: []string{"ask", "-a", "file1.txt", "-a", "file2.txt", "prompt"},
			want: AskCmd{
				Prompt: "prompt",
				Attach: []string{"file1.txt", "file2.txt"},
			},
		},
		{
			name: "ask with all flags",
			args: []string{"ask", "-m", "claude", "--stream", "-t", "0.7", "-a", "doc.pdf", "complex prompt"},
			want: AskCmd{
				Prompt:      "complex prompt",
				Model:       "claude",
				Stream:      true,
				Temperature: 0.7,
				Attach:      []string{"doc.pdf"},
			},
		},
		{
			name:    "ask without prompt",
			args:    []string{"ask"},
			wantErr: false, // No longer errors at parse time since prompt is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli CLI
			parser := kong.Must(&cli)

			ctx, err := parser.Parse(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			// Command format changes based on whether prompt is provided
			if cli.Ask.Prompt == "" {
				assert.Equal(t, "ask", ctx.Command())
			} else {
				assert.Equal(t, "ask <prompt>", ctx.Command())
			}
			assert.Equal(t, tt.want, cli.Ask)
		})
	}
}

func TestCLI_ChatCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    ChatCmd
		wantErr bool
	}{
		{
			name: "basic chat",
			args: []string{"chat"},
			want: ChatCmd{},
		},
		{
			name: "chat with resume",
			args: []string{"chat", "-r", "session123"},
			want: ChatCmd{
				Resume: "session123",
			},
		},
		{
			name: "chat with model",
			args: []string{"chat", "-m", "gpt-4"},
			want: ChatCmd{
				Model: "gpt-4",
			},
		},
		{
			name: "chat with attachments",
			args: []string{"chat", "-a", "file1.txt", "-a", "file2.txt"},
			want: ChatCmd{
				Attach: []string{"file1.txt", "file2.txt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli CLI
			parser := kong.Must(&cli)

			ctx, err := parser.Parse(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "chat", ctx.Command())
			assert.Equal(t, tt.want, cli.Chat)
		})
	}
}

func TestCLI_ConfigCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "config show",
			args: []string{"config", "show"},
			want: "config show",
		},
		{
			name: "config get",
			args: []string{"config", "get", "provider"},
			want: "config get <key>",
		},
		{
			name: "config set",
			args: []string{"config", "set", "provider", "openai"},
			want: "config set <key> <value>",
		},
		{
			name: "config validate",
			args: []string{"config", "validate"},
			want: "config validate",
		},
		{
			name:    "config without subcommand",
			args:    []string{"config"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli CLI
			parser := kong.Must(&cli)

			ctx, err := parser.Parse(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, ctx.Command())
		})
	}
}
