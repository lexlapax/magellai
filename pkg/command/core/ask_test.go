// ABOUTME: Unit tests for the ask command
// ABOUTME: Tests prompt handling, provider selection, streaming, and attachments
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
)

func TestAskCommand(t *testing.T) {
	// Initialize config
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.Manager
	if err := cfg.SetValue("model", "mock/test-model"); err != nil {
		t.Fatalf("Failed to set model: %v", err)
	}
	if err := cfg.SetValue("defaults.system_prompt", "You are a helpful assistant"); err != nil {
		t.Fatalf("Failed to set system prompt: %v", err)
	}
	if err := cfg.SetValue("output", "text"); err != nil {
		t.Fatalf("Failed to set output format: %v", err)
	}

	cmd := NewAskCommand(cfg)

	t.Run("Metadata", func(t *testing.T) {
		meta := cmd.Metadata()
		if meta.Name != "ask" {
			t.Errorf("expected command name 'ask', got %s", meta.Name)
		}
		if meta.Category != command.CategoryCLI {
			t.Errorf("expected CLI category, got %v", meta.Category)
		}
		if len(meta.Flags) == 0 {
			t.Error("expected command to have flags")
		}
	})

	t.Run("Validate", func(t *testing.T) {
		err := cmd.Validate()
		if err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("Execute", func(t *testing.T) {
		tests := []struct {
			name    string
			args    []string
			flags   map[string]interface{}
			wantErr bool
		}{
			{
				name:    "Simple prompt",
				args:    []string{"Hello, world!"},
				flags:   map[string]interface{}{},
				wantErr: false,
			},
			{
				name:    "Prompt with model flag",
				args:    []string{"Test prompt"},
				flags:   map[string]interface{}{"model": "mock/test"},
				wantErr: false,
			},
			{
				name:    "Prompt with system flag",
				args:    []string{"Calculate 2+2"},
				flags:   map[string]interface{}{"system": "You are a math tutor"},
				wantErr: false,
			},
			{
				name: "Prompt with temperature",
				args: []string{"Tell me a story"},
				flags: map[string]interface{}{
					"temperature": 0.8,
				},
				wantErr: false,
			},
			{
				name: "Prompt with max tokens",
				args: []string{"Explain quantum computing"},
				flags: map[string]interface{}{
					"max-tokens": 500,
				},
				wantErr: false,
			},
			{
				name:    "No prompt",
				args:    []string{},
				flags:   map[string]interface{}{},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := context.Background()
				var stdout, stderr bytes.Buffer

				exec := &command.ExecutionContext{
					Context: ctx,
					Args:    tt.args,
					Flags:   command.NewFlags(tt.flags),
					Stdout:  &stdout,
					Stderr:  &stderr,
					Config:  cfg,
				}

				err := cmd.Execute(ctx, exec)
				if (err != nil) != tt.wantErr {
					t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr && stdout.Len() == 0 && (err == nil || !strings.Contains(err.Error(), "mock provider")) {
					t.Error("expected output but got none")
				}
			})
		}
	})

	t.Run("JSON output format", func(t *testing.T) {
		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"Test JSON output"},
			Flags: command.NewFlags(map[string]interface{}{
				"output": "json",
			}),
			Stdout: &stdout,
			Stderr: &bytes.Buffer{},
			Config: cfg,
		}

		err := cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Fatalf("Execute() error = %v", err)
		}

		// If we got output, check if it's valid JSON
		if stdout.Len() > 0 {
			var result map[string]interface{}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v", err)
			}

			// Verify expected fields
			if _, ok := result["content"]; !ok {
				t.Error("JSON output missing 'content' field")
			}
			if _, ok := result["model"]; !ok {
				t.Error("JSON output missing 'model' field")
			}
			if _, ok := result["provider"]; !ok {
				t.Error("JSON output missing 'provider' field")
			}
		}
	})

	t.Run("Multi-word prompt", func(t *testing.T) {
		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"This", "is", "a", "multi", "word", "prompt"},
			Flags:   command.NewFlags(map[string]interface{}{}),
			Stdout:  &stdout,
			Stderr:  &bytes.Buffer{},
			Config:  cfg,
		}

		err := cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Fatalf("Execute() error = %v", err)
		}

		// The prompt should be joined with spaces
		if stdout.Len() == 0 && (err == nil || !strings.Contains(err.Error(), "mock provider")) {
			t.Error("expected output or mock provider error but got none")
		}
	})
}

func TestAskCommandStreaming(t *testing.T) {
	// Initialize config
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.Manager
	if err := cfg.SetValue("model", "mock/test-model"); err != nil {
		t.Fatalf("Failed to set model: %v", err)
	}

	cmd := NewAskCommand(cfg)

	t.Run("Streaming text output", func(t *testing.T) {
		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"Test streaming"},
			Flags: command.NewFlags(map[string]interface{}{
				"stream": true,
			}),
			Stdout: &stdout,
			Stderr: &bytes.Buffer{},
			Config: cfg,
		}

		err := cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("Streaming JSON output", func(t *testing.T) {
		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"Test streaming JSON"},
			Flags: command.NewFlags(map[string]interface{}{
				"stream": true,
				"output": "json",
			}),
			Stdout: &stdout,
			Stderr: &bytes.Buffer{},
			Config: cfg,
		}

		err := cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Fatalf("Execute() error = %v", err)
		}

		// If we got output, check if it's valid JSON
		if stdout.Len() > 0 {
			var result map[string]interface{}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("Failed to parse JSON output: %v", err)
			}
		}
	})
}

func TestAskCommandAttachments(t *testing.T) {
	// Initialize config
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.Manager
	if err := cfg.SetValue("model", "mock/test-model"); err != nil {
		t.Fatalf("Failed to set model: %v", err)
	}

	cmd := NewAskCommand(cfg)

	t.Run("Single attachment", func(t *testing.T) {
		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"Describe this image"},
			Flags: command.NewFlags(map[string]interface{}{
				"attach": []string{"test.png"},
			}),
			Stdout: &stdout,
			Stderr: &bytes.Buffer{},
			Config: cfg,
		}

		// Should work with mock provider
		err := cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Multiple attachments", func(t *testing.T) {
		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"Compare these files"},
			Flags: command.NewFlags(map[string]interface{}{
				"attach": []string{"file1.txt", "file2.txt"},
			}),
			Stdout: &stdout,
			Stderr: &bytes.Buffer{},
			Config: cfg,
		}

		// Should work with mock provider
		err := cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
