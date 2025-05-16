// ABOUTME: Unit tests for the profile command, covering all subcommands and edge cases
// ABOUTME: Tests profile lifecycle management including create, switch, update, copy operations

package core

import (
	"bytes"
	"context"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileCommand_Execute(t *testing.T) {
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
			name: "show current profile",
			args: []string{},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profile.current", "development"))
			},
			expectedOutput: "Current profile: development",
		},
		{
			name: "show current profile with details",
			args: []string{},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profile.current", "work"))
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{
					"provider":    "openai",
					"model":       "gpt-4",
					"description": "Work profile",
				}))
			},
			expectedOutput: "Current profile: work\n  Provider: openai\n  Model: gpt-4\n  Description: Work profile",
		},
		{
			name:           "show current profile JSON",
			args:           []string{},
			outputFormat:   "json",
			expectedOutput: `Current profile: default`,
		},

		// List command
		{
			name: "list profiles",
			args: []string{"list"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{}))
				require.NoError(t, c.SetValue("profiles.personal", map[string]interface{}{}))
			},
			expectedOutput: "personal",
		},
		{
			name:           "list profiles JSON",
			args:           []string{"list"},
			flags:          map[string]interface{}{"format": "json"},
			expectedOutput: `"current": "default"`,
		},

		// Show command
		{
			name: "show specific profile",
			args: []string{"show", "work"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{
					"provider":    "anthropic",
					"model":       "claude-3",
					"description": "Work profile for coding",
					"settings": map[string]interface{}{
						"temperature": 0.7,
						"max_tokens":  2000,
					},
				}))
			},
			expectedOutput: "temperature: 0.7",
		},
		{
			name: "show specific profile with max_tokens",
			args: []string{"show", "work"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{
					"provider":    "anthropic",
					"model":       "claude-3",
					"description": "Work profile for coding",
					"settings": map[string]interface{}{
						"temperature": 0.7,
						"max_tokens":  2000,
					},
				}))
			},
			expectedOutput: "max_tokens: 2000",
		},
		{
			name:          "show non-existent profile",
			args:          []string{"show", "nonexistent"},
			expectedError: "failed to get profile 'nonexistent'",
		},
		{
			name:           "show without profile name",
			args:           []string{"show"},
			expectedOutput: "Current profile: default",
		},

		// Create command
		{
			name:           "create profile",
			args:           []string{"create", "test"},
			expectedOutput: "Created profile: test",
		},
		{
			name: "create profile with options",
			args: []string{"create", "quick"},
			flags: map[string]interface{}{
				"provider":    "openai",
				"model":       "gpt-3.5-turbo",
				"description": "Fast responses",
			},
			expectedOutput: "Created profile: quick",
		},
		{
			name: "create profile copying from existing",
			args: []string{"create", "work-copy"},
			flags: map[string]interface{}{
				"copy-from": "work",
			},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{
					"provider": "openai",
					"model":    "gpt-4",
				}))
			},
			expectedOutput: "Created profile: work-copy",
		},
		{
			name: "create existing profile",
			args: []string{"create", "existing"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.existing", map[string]interface{}{}))
			},
			expectedError: "profile 'existing' already exists",
		},
		{
			name:          "create without name",
			args:          []string{"create"},
			expectedError: "missing argument - name required",
		},

		// Switch command
		{
			name: "switch profile",
			args: []string{"switch", "work"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{}))
			},
			expectedOutput: "Switched to profile: work",
		},
		{
			name: "switch by direct name",
			args: []string{"work"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{}))
			},
			expectedOutput: "Switched to profile: work",
		},
		{
			name:          "switch to non-existent profile",
			args:          []string{"switch", "nonexistent"},
			expectedError: "profile 'nonexistent' not found",
		},
		{
			name:          "switch without name",
			args:          []string{"switch"},
			expectedError: "missing argument - name required",
		},

		// Delete command
		{
			name:          "delete default profile",
			args:          []string{"delete", "default"},
			expectedError: "cannot delete default profile",
		},
		{
			name: "delete current profile",
			args: []string{"delete", "current"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.current", map[string]interface{}{}))
				require.NoError(t, c.SetValue("profile.current", "current"))
			},
			expectedError: "cannot delete current profile 'current'",
		},
		{
			name: "delete profile",
			args: []string{"delete", "test"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.test", map[string]interface{}{}))
			},
			expectedError: "profile deletion not implemented",
		},
		{
			name:          "delete non-existent profile",
			args:          []string{"delete", "nonexistent"},
			expectedError: "profile 'nonexistent' not found",
		},
		{
			name:          "delete without name",
			args:          []string{"delete"},
			expectedError: "missing argument - name required",
		},

		// Update command
		{
			name: "update profile settings",
			args: []string{"update", "work", "temperature=0.5", "max_tokens=1000"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{}))
			},
			expectedOutput: "Updated profile: work",
		},
		{
			name:          "update non-existent profile",
			args:          []string{"update", "nonexistent", "key=value"},
			expectedError: "profile 'nonexistent' not found",
		},
		{
			name: "update with invalid format",
			args: []string{"update", "work", "invalid_format"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{}))
			},
			expectedError: "invalid update format: invalid_format",
		},
		{
			name:          "update without values",
			args:          []string{"update", "work"},
			expectedError: "missing argument - name and key=value required",
		},

		// Copy command
		{
			name: "copy profile",
			args: []string{"copy", "work", "work-backup"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{
					"provider": "openai",
					"model":    "gpt-4",
				}))
			},
			expectedOutput: "Copied profile 'work' to 'work-backup'",
		},
		{
			name:          "copy non-existent profile",
			args:          []string{"copy", "nonexistent", "backup"},
			expectedError: "source profile 'nonexistent' not found",
		},
		{
			name: "copy to existing profile",
			args: []string{"copy", "work", "existing"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{}))
				require.NoError(t, c.SetValue("profiles.existing", map[string]interface{}{}))
			},
			expectedError: "destination profile 'existing' already exists",
		},
		{
			name:          "copy without destination",
			args:          []string{"copy", "work"},
			expectedError: "missing argument - source and destination required",
		},

		// Export command
		{
			name: "export profile",
			args: []string{"export", "work"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{
					"provider": "openai",
					"model":    "gpt-4",
				}))
			},
			expectedOutput: "provider: openai",
		},
		{
			name:  "export profile JSON",
			args:  []string{"export", "work"},
			flags: map[string]interface{}{"format": "json"},
			setupConfig: func(c *config.Config) {
				require.NoError(t, c.SetValue("profiles.work", map[string]interface{}{
					"provider": "openai",
				}))
			},
			expectedOutput: `"Provider": "openai"`,
		},
		{
			name:          "export non-existent profile",
			args:          []string{"export", "nonexistent"},
			expectedError: "failed to export profile",
		},
		{
			name:          "export without name",
			args:          []string{"export"},
			expectedError: "missing argument - name required",
		},

		// Import command
		{
			name:          "import profile",
			args:          []string{"import", "test", "profile.yaml"},
			expectedError: "profile import not implemented",
		},
		{
			name:          "import without filename",
			args:          []string{"import", "test"},
			expectedError: "missing argument - name and filename required",
		},

		// Invalid commands
		{
			name:          "invalid subcommand",
			args:          []string{"invalid"},
			expectedError: "profile 'invalid' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test config
			cfg := createTestConfig(t)
			if tt.setupConfig != nil {
				tt.setupConfig(cfg)
			}

			cmd := NewProfileCommand(cfg)

			ctx := context.Background()
			var stdout, stderr bytes.Buffer
			exec := &command.ExecutionContext{
				Args:   tt.args,
				Flags:  tt.flags,
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
				assert.Contains(t, output, tt.expectedOutput)
			}
		})
	}
}

func TestProfileCommand_Metadata(t *testing.T) {
	cmd := NewProfileCommand(nil)
	meta := cmd.Metadata()

	assert.Equal(t, "profile", meta.Name)
	assert.Contains(t, meta.Aliases, "prof")
	assert.Equal(t, command.CategoryShared, meta.Category)
	assert.NotEmpty(t, meta.Description)
	assert.NotEmpty(t, meta.LongDescription)
	assert.Contains(t, meta.LongDescription, "list")
	assert.Contains(t, meta.LongDescription, "show")
	assert.Contains(t, meta.LongDescription, "create")
	assert.Contains(t, meta.LongDescription, "switch")
	assert.Contains(t, meta.LongDescription, "delete")
	assert.Contains(t, meta.LongDescription, "update")
	assert.Contains(t, meta.LongDescription, "copy")
	assert.Contains(t, meta.LongDescription, "export")
	assert.Contains(t, meta.LongDescription, "import")

	// Check flags
	assert.Len(t, meta.Flags, 1)
	formatFlag := meta.Flags[0]
	assert.Equal(t, "format", formatFlag.Name)
	assert.Equal(t, "f", formatFlag.Short)
	assert.Equal(t, command.FlagTypeString, formatFlag.Type)
	assert.Equal(t, "text", formatFlag.Default)
}

func TestProfileCommand_Validate(t *testing.T) {
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
			cmd := NewProfileCommand(tt.config)
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

func TestProfileCommand_CompleteLifecycle(t *testing.T) {
	cfg := createTestConfig(t)
	cmd := NewProfileCommand(cfg)
	ctx := context.Background()

	// 1. Create a profile
	exec := &command.ExecutionContext{
		Args: []string{"create", "test"},
		Flags: map[string]interface{}{
			"provider":    "anthropic",
			"model":       "claude-3",
			"description": "Test profile",
		},
		Data: make(map[string]interface{}),
	}
	err := cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "Created profile: test")

	// 2. List profiles
	exec = &command.ExecutionContext{
		Args: []string{"list"},
		Data: make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output := exec.Data["output"].(string)
	assert.Contains(t, output, "test")

	// 3. Switch to the new profile
	exec = &command.ExecutionContext{
		Args: []string{"switch", "test"},
		Data: make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "Switched to profile: test")

	// 4. Update the profile
	exec = &command.ExecutionContext{
		Args: []string{"update", "test", "temperature=0.8", "max_tokens=3000"},
		Data: make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "Updated profile: test")

	// 5. Show the profile details
	exec = &command.ExecutionContext{
		Args: []string{"show", "test"},
		Data: make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output = exec.Data["output"].(string)
	assert.Contains(t, output, "Profile: test")
	assert.Contains(t, output, "Provider: anthropic")
	assert.Contains(t, output, "Model: claude-3")
	assert.Contains(t, output, "Description: Test profile")

	// 6. Copy the profile
	exec = &command.ExecutionContext{
		Args: []string{"copy", "test", "test-backup"},
		Data: make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	assert.Contains(t, exec.Data["output"], "Copied profile 'test' to 'test-backup'")

	// 7. Export the profile
	exec = &command.ExecutionContext{
		Args:  []string{"export", "test"},
		Flags: map[string]interface{}{"format": "json"},
		Data:  make(map[string]interface{}),
	}
	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)
	output = exec.Data["output"].(string)
	assert.Contains(t, output, "claude-3")
	assert.Contains(t, output, "anthropic")
}
