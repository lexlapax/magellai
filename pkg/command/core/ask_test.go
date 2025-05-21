// ABOUTME: Unit tests for the ask command
// ABOUTME: Tests prompt handling, provider selection, streaming, and attachments
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
)

func TestAskCommand(t *testing.T) {
	// Initialize config
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.Manager
	if err := cfg.SetValue("model.default", "mock/test-model"); err != nil {
		t.Fatalf("Failed to set model.default: %v", err)
	}

	// Set a mock API key for the mock provider
	if err := cfg.SetValue("provider.mock.api_key", "mock-api-key"); err != nil {
		t.Fatalf("Failed to set provider.mock.api_key: %v", err)
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
	if err := cfg.SetValue("model.default", "mock/test-model"); err != nil {
		t.Fatalf("Failed to set model.default: %v", err)
	}

	// Set a mock API key for the mock provider
	if err := cfg.SetValue("provider.mock.api_key", "mock-api-key"); err != nil {
		t.Fatalf("Failed to set provider.mock.api_key: %v", err)
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

func TestAskCommandAPIKeyResolution(t *testing.T) {
	// Initialize config
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.Manager

	// Test with different providers
	providers := []string{"openai", "anthropic", "gemini"}

	for _, providerName := range providers {
		t.Run(providerName+" provider", func(t *testing.T) {
			// Setup config for this provider
			modelName := "test-model"
			if providerName == "openai" {
				modelName = "gpt-4"
			} else if providerName == "anthropic" {
				modelName = "claude-3"
			} else if providerName == "gemini" {
				modelName = "gemini-pro"
			}

			// Configure the model and API key
			fullModelName := providerName + "/" + modelName
			apiKey := "test-api-key-for-" + providerName

			if err := cfg.SetValue("model.default", fullModelName); err != nil {
				t.Fatalf("Failed to set model.default: %v", err)
			}

			apiKeyPath := "provider." + providerName + ".api_key"
			if err := cfg.SetValue(apiKeyPath, apiKey); err != nil {
				t.Fatalf("Failed to set %s: %v", apiKeyPath, err)
			}

			// Create command instance that will use our config
			cmd := NewAskCommand(cfg)

			// Create a context that will trap all calls to NewProvider
			ctx := context.Background()
			var stdout, stderr bytes.Buffer

			exec := &command.ExecutionContext{
				Context: ctx,
				Args:    []string{"Test with " + providerName},
				Flags:   command.NewFlags(nil),
				Stdout:  &stdout,
				Stderr:  &stderr,
				Config:  cfg,
			}

			// Execute should fail with an error about provider not being found
			// (since we're providing a fake API key), but we should be able to
			// verify it's trying to use the right provider and API key
			err := cmd.Execute(ctx, exec)

			// Should fail since we're using fake API keys and models
			require.Error(t, err)

			// Error should mention the right provider
			require.Contains(t, err.Error(), providerName)

			// API key should be visible in the error or test environment
			// but we can't easily assert this without mocking the llm.NewProvider function
		})
	}
}

func TestAskCommandAttachments(t *testing.T) {
	// Initialize config
	if err := config.Init(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.Manager
	if err := cfg.SetValue("model.default", "mock/test-model"); err != nil {
		t.Fatalf("Failed to set model.default: %v", err)
	}

	// Set a mock API key for the mock provider
	if err := cfg.SetValue("provider.mock.api_key", "mock-api-key"); err != nil {
		t.Fatalf("Failed to set provider.mock.api_key: %v", err)
	}

	cmd := NewAskCommand(cfg)

	t.Run("Single attachment", func(t *testing.T) {
		// Create a temporary file for testing
		tmpFile, err := os.CreateTemp("", "test*.png")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Write some test content
		_, err = tmpFile.WriteString("Test image content")
		require.NoError(t, err)
		tmpFile.Close()

		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"Describe this image"},
			Flags: command.NewFlags(map[string]interface{}{
				"attach": []string{tmpFile.Name()},
			}),
			Stdout: &stdout,
			Stderr: &bytes.Buffer{},
			Config: cfg,
		}

		// Should work with mock provider
		err = cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Multiple attachments", func(t *testing.T) {
		// Create temporary files for testing
		tmpFile1, err := os.CreateTemp("", "test*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile1.Name())

		tmpFile2, err := os.CreateTemp("", "test*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile2.Name())

		// Write some test content
		_, err = tmpFile1.WriteString("Test file 1 content")
		require.NoError(t, err)
		tmpFile1.Close()

		_, err = tmpFile2.WriteString("Test file 2 content")
		require.NoError(t, err)
		tmpFile2.Close()

		ctx := context.Background()
		var stdout bytes.Buffer

		exec := &command.ExecutionContext{
			Context: ctx,
			Args:    []string{"Compare these files"},
			Flags: command.NewFlags(map[string]interface{}{
				"attach": []string{tmpFile1.Name(), tmpFile2.Name()},
			}),
			Stdout: &stdout,
			Stderr: &bytes.Buffer{},
			Config: cfg,
		}

		// Should work with mock provider
		err = cmd.Execute(ctx, exec)
		if err != nil && !strings.Contains(err.Error(), "mock provider") {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
