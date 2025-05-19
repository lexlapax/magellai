// ABOUTME: Tests for the config generate subcommand
// ABOUTME: Verifies example configuration generation functionality

package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
)

func TestConfigGenerate(t *testing.T) {
	// Initialize config manager for test
	t.Setenv("MAGELLAI_LOG_LEVEL", "error")
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	tests := []struct {
		name      string
		flags     map[string]interface{}
		wantError bool
		checkFile bool
	}{
		{
			name: "generate default",
			flags: map[string]interface{}{
				"output": filepath.Join(t.TempDir(), "default.yaml"),
			},
			wantError: false,
			checkFile: true,
		},
		{
			name: "generate with custom output",
			flags: map[string]interface{}{
				"output": filepath.Join(t.TempDir(), "custom.yaml"),
			},
			wantError: false,
			checkFile: true,
		},
		{
			name: "generate with existing file no force",
			flags: map[string]interface{}{
				"output": createTempFile(t, "existing.yaml"),
			},
			wantError: true,
			checkFile: false,
		},
		{
			name: "generate with existing file with force",
			flags: map[string]interface{}{
				"output": createTempFile(t, "existing-force.yaml"),
				"force":  true,
			},
			wantError: false,
			checkFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewConfigCommand(config.Manager)
			exec := &command.ExecutionContext{
				Args:   []string{"generate"},
				Flags:  command.NewFlags(tt.flags),
				Stdout: &strings.Builder{},
				Stderr: &strings.Builder{},
				Data:   make(map[string]interface{}),
			}

			err := cmd.Execute(context.TODO(), exec)
			if (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.checkFile && err == nil {
				// Check if file was created
				generatedPath, ok := exec.Data["generated_path"].(string)
				if !ok {
					t.Error("Generated path not found in execution data")
					return
				}

				// Check if file exists
				if _, err := os.Stat(generatedPath); os.IsNotExist(err) {
					t.Errorf("Generated file does not exist: %s", generatedPath)
					return
				}

				// Read and verify content
				content, err := os.ReadFile(generatedPath)
				if err != nil {
					t.Errorf("Failed to read generated file: %v", err)
					return
				}

				// Check for essential content
				contentStr := string(content)
				essentialStrings := []string{
					"Magellai Configuration File",
					"provider:",
					"model:",
					"output:",
					"session:",
					"repl:",
					"profiles:",
				}

				for _, str := range essentialStrings {
					if !strings.Contains(contentStr, str) {
						t.Errorf("Generated config missing essential content: %s", str)
					}
				}

				// Clean up generated file if it's not in temp dir
				if !strings.HasPrefix(generatedPath, os.TempDir()) {
					os.Remove(generatedPath)
				}
			}
		})
	}
}

func createTempFile(t *testing.T, name string) string {
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return path
}
