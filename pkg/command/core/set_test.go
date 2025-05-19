// ABOUTME: Tests for the Set and Get commands
// ABOUTME: Verifies shared context manipulation through commands

package core

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
)

func TestSetCommand_ShowSettings(t *testing.T) {
	sharedContext := command.NewSharedContext()
	sharedContext.SetModel("gpt-4")
	sharedContext.SetTemperature(0.7)

	cmd := NewSetCommand(sharedContext)
	var stdout bytes.Buffer

	exec := &command.ExecutionContext{
		Args:   []string{},
		Stdout: &stdout,
	}

	err := cmd.Execute(context.Background(), exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "model: gpt-4") {
		t.Errorf("expected model in output, got: %s", output)
	}
	if !strings.Contains(output, "temperature: 0.70") {
		t.Errorf("expected temperature in output, got: %s", output)
	}
}

func TestSetCommand_SetModel(t *testing.T) {
	sharedContext := command.NewSharedContext()
	cmd := NewSetCommand(sharedContext)
	var stdout bytes.Buffer

	exec := &command.ExecutionContext{
		Args:   []string{"model", "claude-3"},
		Stdout: &stdout,
	}

	err := cmd.Execute(context.Background(), exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sharedContext.Model() != "claude-3" {
		t.Errorf("expected model claude-3, got %s", sharedContext.Model())
	}

	output := stdout.String()
	if !strings.Contains(output, "Model set to: claude-3") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestSetCommand_SetTemperature(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantVal float64
	}{
		{
			name:    "valid temperature",
			args:    []string{"temperature", "0.8"},
			wantErr: false,
			wantVal: 0.8,
		},
		{
			name:    "temperature alias",
			args:    []string{"temp", "1.5"},
			wantErr: false,
			wantVal: 1.5,
		},
		{
			name:    "invalid temperature format",
			args:    []string{"temperature", "invalid"},
			wantErr: true,
		},
		{
			name:    "temperature too low",
			args:    []string{"temperature", "-0.5"},
			wantErr: true,
		},
		{
			name:    "temperature too high",
			args:    []string{"temperature", "2.5"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sharedContext := command.NewSharedContext()
			cmd := NewSetCommand(sharedContext)
			var stdout bytes.Buffer

			exec := &command.ExecutionContext{
				Args:   tt.args,
				Stdout: &stdout,
			}

			err := cmd.Execute(context.Background(), exec)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && sharedContext.Temperature() != tt.wantVal {
				t.Errorf("expected temperature %f, got %f", tt.wantVal, sharedContext.Temperature())
			}
		})
	}
}

func TestSetCommand_SetMaxTokens(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantVal int
	}{
		{
			name:    "valid max_tokens",
			args:    []string{"max_tokens", "1000"},
			wantErr: false,
			wantVal: 1000,
		},
		{
			name:    "maxtokens alias",
			args:    []string{"maxtokens", "500"},
			wantErr: false,
			wantVal: 500,
		},
		{
			name:    "tokens alias",
			args:    []string{"tokens", "2000"},
			wantErr: false,
			wantVal: 2000,
		},
		{
			name:    "invalid format",
			args:    []string{"max_tokens", "invalid"},
			wantErr: true,
		},
		{
			name:    "negative value",
			args:    []string{"max_tokens", "-100"},
			wantErr: true,
		},
		{
			name:    "zero value",
			args:    []string{"max_tokens", "0"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sharedContext := command.NewSharedContext()
			cmd := NewSetCommand(sharedContext)
			var stdout bytes.Buffer

			exec := &command.ExecutionContext{
				Args:   tt.args,
				Stdout: &stdout,
			}

			err := cmd.Execute(context.Background(), exec)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && sharedContext.MaxTokens() != tt.wantVal {
				t.Errorf("expected maxTokens %d, got %d", tt.wantVal, sharedContext.MaxTokens())
			}
		})
	}
}

func TestSetCommand_SetBooleans(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		getter  func(*command.SharedContext) bool
		wantErr bool
		wantVal bool
	}{
		{
			name:    "stream true",
			args:    []string{"stream", "true"},
			getter:  (*command.SharedContext).Stream,
			wantErr: false,
			wantVal: true,
		},
		{
			name:    "stream false",
			args:    []string{"stream", "false"},
			getter:  (*command.SharedContext).Stream,
			wantErr: false,
			wantVal: false,
		},
		{
			name:    "verbose true",
			args:    []string{"verbose", "true"},
			getter:  (*command.SharedContext).Verbose,
			wantErr: false,
			wantVal: true,
		},
		{
			name:    "debug false",
			args:    []string{"debug", "false"},
			getter:  (*command.SharedContext).Debug,
			wantErr: false,
			wantVal: false,
		},
		{
			name:    "invalid bool",
			args:    []string{"stream", "invalid"},
			getter:  (*command.SharedContext).Stream,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sharedContext := command.NewSharedContext()
			cmd := NewSetCommand(sharedContext)
			var stdout bytes.Buffer

			exec := &command.ExecutionContext{
				Args:   tt.args,
				Stdout: &stdout,
			}

			err := cmd.Execute(context.Background(), exec)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.getter(sharedContext) != tt.wantVal {
				t.Errorf("expected %v, got %v", tt.wantVal, tt.getter(sharedContext))
			}
		})
	}
}

func TestSetCommand_CustomKeys(t *testing.T) {
	sharedContext := command.NewSharedContext()
	cmd := NewSetCommand(sharedContext)
	var stdout bytes.Buffer

	exec := &command.ExecutionContext{
		Args:   []string{"custom-key", "custom value with spaces"},
		Stdout: &stdout,
	}

	err := cmd.Execute(context.Background(), exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, exists := sharedContext.Get("custom-key")
	if !exists {
		t.Error("expected custom-key to exist")
	}
	if val != "custom value with spaces" {
		t.Errorf("expected 'custom value with spaces', got %v", val)
	}
}

func TestGetCommand_Basic(t *testing.T) {
	sharedContext := command.NewSharedContext()
	sharedContext.SetModel("gpt-4")
	sharedContext.SetTemperature(0.9)
	sharedContext.SetMaxTokens(1500)
	sharedContext.SetStream(false)
	sharedContext.Set("custom", "value")

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "get model",
			args:    []string{"model"},
			want:    "gpt-4\n",
			wantErr: false,
		},
		{
			name:    "get temperature",
			args:    []string{"temperature"},
			want:    "0.90\n",
			wantErr: false,
		},
		{
			name:    "get max_tokens",
			args:    []string{"max_tokens"},
			want:    "1500\n",
			wantErr: false,
		},
		{
			name:    "get stream",
			args:    []string{"stream"},
			want:    "false\n",
			wantErr: false,
		},
		{
			name:    "get custom key",
			args:    []string{"custom"},
			want:    "value\n",
			wantErr: false,
		},
		{
			name:    "get nonexistent key",
			args:    []string{"nonexistent"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGetCommand(sharedContext)
			var stdout bytes.Buffer

			exec := &command.ExecutionContext{
				Args:   tt.args,
				Stdout: &stdout,
			}

			err := cmd.Execute(context.Background(), exec)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && stdout.String() != tt.want {
				t.Errorf("expected output %q, got %q", tt.want, stdout.String())
			}
		})
	}
}

func TestGetCommand_Metadata(t *testing.T) {
	sharedContext := command.NewSharedContext()
	cmd := NewGetCommand(sharedContext)

	meta := cmd.Metadata()
	if meta.Name != "get" {
		t.Errorf("expected name 'get', got %s", meta.Name)
	}
	if meta.Category != command.CategoryREPL {
		t.Error("expected REPL category")
	}
}

func TestSetCommand_Metadata(t *testing.T) {
	sharedContext := command.NewSharedContext()
	cmd := NewSetCommand(sharedContext)

	meta := cmd.Metadata()
	if meta.Name != "set" {
		t.Errorf("expected name 'set', got %s", meta.Name)
	}
	if meta.Category != command.CategoryREPL {
		t.Error("expected REPL category")
	}
}
