// ABOUTME: Temporary stub implementations for ask and chat commands
// ABOUTME: These will be replaced with actual implementations later

package main

import (
	"context"
	"fmt"

	"github.com/lexlapax/magellai/pkg/command"
)

// RegisterStubCommands registers temporary stub commands for testing
func RegisterStubCommands(registry *command.Registry) error {
	// Ask command stub
	askCmd := command.NewSimpleCommand(
		&command.Metadata{
			Name:        "ask",
			Description: "Send a one-shot query to the LLM",
			Category:    command.CategoryCLI,
			Flags: []command.Flag{
				{Name: "model", Type: command.FlagTypeString, Description: "Model to use"},
				{Name: "attach", Type: command.FlagTypeStringSlice, Description: "Files to attach"},
				{Name: "stream", Type: command.FlagTypeBool, Description: "Enable streaming"},
				{Name: "temperature", Type: command.FlagTypeFloat, Description: "Temperature setting"},
			},
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			fmt.Fprintf(exec.Stdout, "Ask command not yet implemented. Prompt: %v\n", exec.Args)
			return nil
		},
	)

	// Chat command stub
	chatCmd := command.NewSimpleCommand(
		&command.Metadata{
			Name:        "chat",
			Description: "Start an interactive chat session",
			Category:    command.CategoryCLI,
			Flags: []command.Flag{
				{Name: "model", Type: command.FlagTypeString, Description: "Model to use"},
				{Name: "resume", Type: command.FlagTypeString, Description: "Resume session ID"},
				{Name: "attach", Type: command.FlagTypeStringSlice, Description: "Files to attach"},
			},
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			fmt.Fprintf(exec.Stdout, "Chat command not yet implemented.\n")
			return nil
		},
	)

	// Register commands
	if err := registry.Register(askCmd); err != nil {
		return err
	}
	if err := registry.Register(chatCmd); err != nil {
		return err
	}

	return nil
}
