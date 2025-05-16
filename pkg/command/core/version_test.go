// ABOUTME: Unit tests for the version command
// ABOUTME: Tests text and JSON output formats

package core

import (
	"context"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		commit         string
		date           string
		flags          map[string]interface{}
		data           map[string]interface{}
		expectedOutput string
		expectedJSON   bool
	}{
		{
			name:           "simple version",
			version:        "1.0.0",
			commit:         "none",
			date:           "unknown",
			expectedOutput: "magellai version 1.0.0",
			expectedJSON:   false,
		},
		{
			name:           "version with commit",
			version:        "1.0.0",
			commit:         "abc123",
			date:           "unknown",
			expectedOutput: "magellai version 1.0.0 (commit: abc123)",
			expectedJSON:   false,
		},
		{
			name:           "version with commit and date",
			version:        "1.0.0",
			commit:         "abc123",
			date:           "2023-12-25",
			expectedOutput: "magellai version 1.0.0 (commit: abc123, built: 2023-12-25)",
			expectedJSON:   false,
		},
		{
			name:           "JSON format via flag",
			version:        "1.0.0",
			commit:         "abc123",
			date:           "2023-12-25",
			flags:          map[string]interface{}{"format": "json"},
			expectedOutput: `"version": "1.0.0"`,
			expectedJSON:   true,
		},
		{
			name:           "JSON format via data",
			version:        "1.0.0",
			commit:         "abc123",
			date:           "2023-12-25",
			data:           map[string]interface{}{"outputFormat": "json"},
			expectedOutput: `"commit": "abc123"`,
			expectedJSON:   true,
		},
		{
			name:           "dev version",
			version:        "dev",
			commit:         "",
			date:           "",
			expectedOutput: "magellai version dev",
			expectedJSON:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCommand(tt.version, tt.commit, tt.date)
			ctx := context.Background()

			exec := &command.ExecutionContext{
				Flags: tt.flags,
				Data:  tt.data,
			}

			if exec.Flags == nil {
				exec.Flags = make(map[string]interface{})
			}
			if exec.Data == nil {
				exec.Data = make(map[string]interface{})
			}

			err := cmd.Execute(ctx, exec)
			assert.NoError(t, err)

			output, ok := exec.Data["output"].(string)
			assert.True(t, ok, "output should be a string")

			if tt.expectedJSON {
				assert.Contains(t, output, tt.expectedOutput)
				assert.Contains(t, output, "\"version\"")
			} else {
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestVersionCommand_Metadata(t *testing.T) {
	cmd := NewVersionCommand("1.0.0", "abc123", "2023-12-25")
	meta := cmd.Metadata()

	assert.Equal(t, "version", meta.Name)
	assert.Contains(t, meta.Aliases, "ver")
	assert.Contains(t, meta.Aliases, "v")
	assert.NotEmpty(t, meta.Description)
	assert.NotEmpty(t, meta.LongDescription)
	assert.Equal(t, command.CategoryShared, meta.Category)
	assert.Len(t, meta.Flags, 1)
	assert.Equal(t, "format", meta.Flags[0].Name)
}

func TestVersionCommand_Validate(t *testing.T) {
	cmd := NewVersionCommand("1.0.0", "abc123", "2023-12-25")
	err := cmd.Validate()
	assert.NoError(t, err)
}
