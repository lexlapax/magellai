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
	if err := registry.Register(chatCmd); err != nil {
		return err
	}

	return nil
}
