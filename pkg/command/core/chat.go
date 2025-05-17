// ABOUTME: Implements the chat command for interactive LLM conversations
// ABOUTME: Provides REPL functionality for chatting with language models

package core

import (
	"context"
	"fmt"
	"os"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/pkg/repl"
)

// ChatCommand represents the chat command
type ChatCommand struct {
	config *config.Config
}

// NewChatCommand creates a new chat command instance
func NewChatCommand(cfg *config.Config) *ChatCommand {
	return &ChatCommand{
		config: cfg,
	}
}

// Metadata returns the command metadata
func (c *ChatCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:            "chat",
		Description:     "Start an interactive chat session with the LLM",
		Category:        command.CategoryCLI,
		LongDescription: "Start an interactive chat session with the LLM supporting conversation history, attachments, and model switching",
		Flags: []command.Flag{
			{
				Name:        "resume",
				Short:       "r",
				Description: "Resume a previous session by ID",
				Type:        command.FlagTypeString,
				Required:    false,
				Default:     "",
			},
			{
				Name:        "model",
				Short:       "m",
				Description: "Override the default model (provider/model format)",
				Type:        command.FlagTypeString,
				Required:    false,
				Default:     "",
			},
			{
				Name:        "attach",
				Short:       "a",
				Description: "Attach files to the initial message",
				Type:        command.FlagTypeStringSlice,
				Required:    false,
				Default:     []string{},
			},
		},
	}
}

// Execute runs the chat command
func (c *ChatCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	// Get configuration
	cfg := c.config

	// Get flags
	sessionID := exec.Flags.GetString("resume")
	model := exec.Flags.GetString("model")
	attachments := exec.Flags.GetStringSlice("attach")

	// Create REPL options
	opts := &repl.REPLOptions{
		Config:    &replConfigAdapter{cfg},
		SessionID: sessionID,
		Model:     model,
		Writer:    exec.Stdout,
		Reader:    os.Stdin,
	}

	// Create and run REPL
	replInstance, err := repl.NewREPL(opts)
	if err != nil {
		return fmt.Errorf("failed to create REPL: %w", err)
	}

	// TODO: Handle initial attachments
	// For now, we'll skip this as the REPL needs to expose attachment functionality
	_ = attachments

	// Run the REPL
	return replInstance.Run()
}

// Validate checks if the command execution context is valid
func (c *ChatCommand) Validate() error {
	return nil
}

// replConfigAdapter adapts our config to the REPL's ConfigInterface
type replConfigAdapter struct {
	cfg *config.Config
}

func (a *replConfigAdapter) GetString(key string) string {
	return a.cfg.GetString(key)
}

func (a *replConfigAdapter) GetBool(key string) bool {
	return a.cfg.GetBool(key)
}

func (a *replConfigAdapter) SetValue(key string, value interface{}) error {
	return a.cfg.SetValue(key, value)
}
