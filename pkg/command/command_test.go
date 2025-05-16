// ABOUTME: Unit tests for the command package
// ABOUTME: Tests command interface, registry, validation, and help system

package command

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

// TestCommandInterface tests basic command interface implementation
func TestCommandInterface(t *testing.T) {
	// Create a simple command
	cmd := NewSimpleCommand(
		&Metadata{
			Name:        "test",
			Description: "Test command",
			Category:    CategoryShared,
		},
		func(ctx context.Context, exec *ExecutionContext) error {
			fmt.Fprintln(exec.Stdout, "test executed")
			return nil
		},
	)

	// Test metadata
	meta := cmd.Metadata()
	if meta.Name != "test" {
		t.Errorf("Expected name 'test', got %s", meta.Name)
	}

	// Test validation
	if err := cmd.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test execution
	var stdout bytes.Buffer
	exec := &ExecutionContext{
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	}

	if err := cmd.Execute(context.Background(), exec); err != nil {
		t.Errorf("Execution failed: %v", err)
	}

	if stdout.String() != "test executed\n" {
		t.Errorf("Expected 'test executed', got %s", stdout.String())
	}
}

// TestRegistry tests command registry functionality
func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Create test commands
	cmd1 := NewSimpleCommand(
		&Metadata{
			Name:        "cmd1",
			Aliases:     []string{"c1"},
			Description: "Command 1",
			Category:    CategoryShared,
		},
		nil,
	)

	cmd2 := NewSimpleCommand(
		&Metadata{
			Name:        "cmd2",
			Description: "Command 2",
			Category:    CategoryCLI,
		},
		nil,
	)

	// Test registration
	if err := registry.Register(cmd1); err != nil {
		t.Errorf("Failed to register cmd1: %v", err)
	}

	if err := registry.Register(cmd2); err != nil {
		t.Errorf("Failed to register cmd2: %v", err)
	}

	// Test duplicate registration
	if err := registry.Register(cmd1); err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test retrieval
	retrieved, err := registry.Get("cmd1")
	if err != nil {
		t.Errorf("Failed to get cmd1: %v", err)
	}
	if retrieved.Metadata().Name != "cmd1" {
		t.Error("Retrieved wrong command")
	}

	// Test alias retrieval
	retrieved, err = registry.Get("c1")
	if err != nil {
		t.Errorf("Failed to get cmd1 by alias: %v", err)
	}
	if retrieved.Metadata().Name != "cmd1" {
		t.Error("Retrieved wrong command by alias")
	}

	// Test listing
	shared := registry.List(CategoryShared)
	if len(shared) != 2 { // cmd1 is shared, cmd2 is CLI but shared commands are included
		t.Errorf("Expected 2 shared commands, got %d", len(shared))
	}

	cli := registry.List(CategoryCLI)
	if len(cli) != 2 { // Both should be included (cmd1 is shared, cmd2 is CLI)
		t.Errorf("Expected 2 CLI commands, got %d", len(cli))
	}

	// Test search
	results := registry.Search("cmd")
	if len(results) != 2 {
		t.Errorf("Expected 2 search results, got %d", len(results))
	}
}

// TestValidation tests command validation
func TestValidation(t *testing.T) {
	cmd := NewSimpleCommand(
		&Metadata{
			Name: "validate-test",
			Flags: []Flag{
				{
					Name:     "required",
					Type:     FlagTypeString,
					Required: true,
				},
				{
					Name: "optional",
					Type: FlagTypeInt,
				},
			},
		},
		nil,
	)

	validator := NewValidator(cmd)

	// Test missing required flag
	flags := map[string]interface{}{
		"optional": 42,
	}

	if err := validator.ValidateFlags(flags); err == nil {
		t.Error("Expected error for missing required flag")
	}

	// Test with required flag
	flags["required"] = "value"
	if err := validator.ValidateFlags(flags); err != nil {
		t.Errorf("Validation failed with required flag: %v", err)
	}

	// Test invalid flag type
	flags["optional"] = "not-an-int"
	if err := validator.ValidateFlags(flags); err == nil {
		t.Error("Expected error for invalid flag type")
	}
}

// TestFlagParsing tests flag value parsing
func TestFlagParsing(t *testing.T) {
	tests := []struct {
		flag  Flag
		value string
		want  interface{}
		err   bool
	}{
		{
			flag:  Flag{Type: FlagTypeString},
			value: "test",
			want:  "test",
		},
		{
			flag:  Flag{Type: FlagTypeInt},
			value: "42",
			want:  42,
		},
		{
			flag:  Flag{Type: FlagTypeBool},
			value: "true",
			want:  true,
		},
		{
			flag:  Flag{Type: FlagTypeFloat},
			value: "3.14",
			want:  3.14,
		},
		{
			flag:  Flag{Type: FlagTypeInt},
			value: "not-a-number",
			err:   true,
		},
	}

	for _, test := range tests {
		got, err := ParseFlagValue(test.flag, test.value)
		if test.err {
			if err == nil {
				t.Errorf("Expected error for %s", test.value)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}

		if got != test.want {
			t.Errorf("ParseFlagValue(%v, %s) = %v, want %v",
				test.flag.Type, test.value, got, test.want)
		}
	}
}

// TestREPLMapping tests REPL command mapping
func TestREPLMapping(t *testing.T) {
	mapper := NewREPLMapper()

	mapper.AddMapping(REPLMapping{
		Flag:        "model",
		REPLCommand: ":model",
	})

	mapper.AddMapping(REPLMapping{
		Flag:        "stream",
		REPLCommand: ":stream",
		ValueTransform: func(v interface{}) (string, error) {
			if boolVal, ok := v.(bool); ok {
				if boolVal {
					return "on", nil
				}
				return "off", nil
			}
			return "", fmt.Errorf("expected boolean")
		},
	})

	// Test basic mapping
	cmd, err := mapper.ConvertFlagToCommand("model", "gpt-4")
	if err != nil {
		t.Errorf("Failed to convert flag: %v", err)
	}
	if cmd != ":model gpt-4" {
		t.Errorf("Expected ':model gpt-4', got %s", cmd)
	}

	// Test with transform
	cmd, err = mapper.ConvertFlagToCommand("stream", true)
	if err != nil {
		t.Errorf("Failed to convert flag: %v", err)
	}
	if cmd != ":stream on" {
		t.Errorf("Expected ':stream on', got %s", cmd)
	}

	// Test unknown flag
	_, err = mapper.ConvertFlagToCommand("unknown", "value")
	if err == nil {
		t.Error("Expected error for unknown flag")
	}
}

// TestHelp tests the help system
func TestHelp(t *testing.T) {
	registry := NewRegistry()

	// Register test command
	cmd := NewSimpleCommand(
		&Metadata{
			Name:            "test",
			Aliases:         []string{"t"},
			Description:     "Test command",
			LongDescription: "This is a test command that does testing things.",
			Category:        CategoryShared,
			Flags: []Flag{
				{
					Name:        "verbose",
					Short:       "v",
					Description: "Enable verbose output",
					Type:        FlagTypeBool,
					Default:     false,
				},
			},
		},
		nil,
	)

	if err := registry.Register(cmd); err != nil {
		t.Fatalf("Failed to register test command: %v", err)
	}

	// Create help command
	formatter := NewDefaultHelpFormatter()
	helpCmd := NewHelpCommand(registry, formatter, CategoryShared)

	// Test general help
	var stdout bytes.Buffer
	exec := &ExecutionContext{
		Args:   []string{},
		Stdout: &stdout,
		Stderr: &bytes.Buffer{},
	}

	if err := helpCmd.Execute(context.Background(), exec); err != nil {
		t.Errorf("Help execution failed: %v", err)
	}

	output := stdout.String()
	if !contains(output, "Available Commands") {
		t.Error("Help output missing 'Available Commands'")
	}
	if !contains(output, "test") {
		t.Error("Help output missing test command")
	}

	// Test specific command help
	stdout.Reset()
	exec.Args = []string{"test"}

	if err := helpCmd.Execute(context.Background(), exec); err != nil {
		t.Errorf("Command help execution failed: %v", err)
	}

	output = stdout.String()
	if !contains(output, "Command: test") {
		t.Error("Command help missing command name")
	}
	if !contains(output, "Description: Test command") {
		t.Error("Command help missing description")
	}
	if !contains(output, "verbose") {
		t.Error("Command help missing flags")
	}
}

// Helper function for string contains
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
