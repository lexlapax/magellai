// ABOUTME: Set command for modifying shared context values
// ABOUTME: Allows users to set model, temperature, and other state

package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/lexlapax/magellai/pkg/command"
)

// SetCommand manages shared context state
type SetCommand struct {
	SharedContext *command.SharedContext
}

// NewSetCommand creates a new set command
func NewSetCommand(sharedContext *command.SharedContext) command.Interface {
	return &SetCommand{
		SharedContext: sharedContext,
	}
}

// Execute implements command.Interface
func (c *SetCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if len(exec.Args) == 0 {
		// Show current settings
		fmt.Fprintln(exec.Stdout, "Current settings:")

		if model := c.SharedContext.Model(); model != "" {
			fmt.Fprintf(exec.Stdout, "  model: %s\n", model)
		}

		if temp := c.SharedContext.Temperature(); temp != -1 {
			fmt.Fprintf(exec.Stdout, "  temperature: %.2f\n", temp)
		}

		if maxTokens := c.SharedContext.MaxTokens(); maxTokens != -1 {
			fmt.Fprintf(exec.Stdout, "  max_tokens: %d\n", maxTokens)
		}

		fmt.Fprintf(exec.Stdout, "  stream: %v\n", c.SharedContext.Stream())
		fmt.Fprintf(exec.Stdout, "  verbose: %v\n", c.SharedContext.Verbose())
		fmt.Fprintf(exec.Stdout, "  debug: %v\n", c.SharedContext.Debug())

		return nil
	}

	if len(exec.Args) < 2 {
		return fmt.Errorf("usage: set <key> <value>")
	}

	key := strings.ToLower(exec.Args[0])
	value := strings.Join(exec.Args[1:], " ")

	switch key {
	case "model":
		c.SharedContext.SetModel(value)
		fmt.Fprintf(exec.Stdout, "Model set to: %s\n", value)

	case "temperature", "temp":
		temp, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid temperature: %s", value)
		}
		if temp < 0 || temp > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
		c.SharedContext.SetTemperature(temp)
		fmt.Fprintf(exec.Stdout, "Temperature set to: %.2f\n", temp)

	case "max_tokens", "maxtokens", "tokens":
		tokens, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid max_tokens: %s", value)
		}
		if tokens < 1 {
			return fmt.Errorf("max_tokens must be positive")
		}
		c.SharedContext.SetMaxTokens(tokens)
		fmt.Fprintf(exec.Stdout, "Max tokens set to: %d\n", tokens)

	case "stream":
		stream, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid stream value: %s", value)
		}
		c.SharedContext.SetStream(stream)
		fmt.Fprintf(exec.Stdout, "Stream set to: %v\n", stream)

	case "verbose":
		verbose, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid verbose value: %s", value)
		}
		c.SharedContext.SetVerbose(verbose)
		fmt.Fprintf(exec.Stdout, "Verbose set to: %v\n", verbose)

	case "debug":
		debug, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid debug value: %s", value)
		}
		c.SharedContext.SetDebug(debug)
		fmt.Fprintf(exec.Stdout, "Debug set to: %v\n", debug)

	default:
		// Store in generic state
		c.SharedContext.Set(key, value)
		fmt.Fprintf(exec.Stdout, "Set %s = %s\n", key, value)
	}

	return nil
}

// Metadata implements command.Interface
func (c *SetCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "set",
		Aliases:     []string{},
		Description: "Set a configuration value",
		LongDescription: `Set a configuration value for the current session.

Available keys:
  model       - Set the LLM model to use
  temperature - Set the temperature (0.0 to 2.0)
  max_tokens  - Set the maximum tokens
  stream      - Enable/disable streaming (true/false)
  verbose     - Enable/disable verbose mode (true/false)
  debug       - Enable/disable debug mode (true/false)
  
You can also set arbitrary key-value pairs that will be stored in the session context.

Examples:
  set model gpt-4
  set temperature 0.7
  set max_tokens 1000
  set stream false
  set custom-key custom-value`,
		Category: command.CategoryREPL,
		Flags:    []command.Flag{},
	}
}

// Validate implements command.Interface
func (c *SetCommand) Validate() error {
	return nil
}

// GetCommand retrieves values from shared context
type GetCommand struct {
	SharedContext *command.SharedContext
}

// NewGetCommand creates a new get command
func NewGetCommand(sharedContext *command.SharedContext) command.Interface {
	return &GetCommand{
		SharedContext: sharedContext,
	}
}

// Execute implements command.Interface
func (c *GetCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if len(exec.Args) == 0 {
		return fmt.Errorf("usage: get <key>")
	}

	key := strings.ToLower(exec.Args[0])

	switch key {
	case "model":
		fmt.Fprintf(exec.Stdout, "%s\n", c.SharedContext.Model())
	case "temperature", "temp":
		fmt.Fprintf(exec.Stdout, "%.2f\n", c.SharedContext.Temperature())
	case "max_tokens", "maxtokens", "tokens":
		fmt.Fprintf(exec.Stdout, "%d\n", c.SharedContext.MaxTokens())
	case "stream":
		fmt.Fprintf(exec.Stdout, "%v\n", c.SharedContext.Stream())
	case "verbose":
		fmt.Fprintf(exec.Stdout, "%v\n", c.SharedContext.Verbose())
	case "debug":
		fmt.Fprintf(exec.Stdout, "%v\n", c.SharedContext.Debug())
	default:
		// Get from generic state
		val, exists := c.SharedContext.Get(key)
		if !exists {
			return fmt.Errorf("key not found: %s", key)
		}
		fmt.Fprintf(exec.Stdout, "%v\n", val)
	}

	return nil
}

// Metadata implements command.Interface
func (c *GetCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "get",
		Aliases:     []string{},
		Description: "Get a configuration value",
		LongDescription: `Get a configuration value from the current session.

Available keys:
  model       - Get the LLM model
  temperature - Get the temperature
  max_tokens  - Get the maximum tokens
  stream      - Get the streaming setting
  verbose     - Get the verbose setting
  debug       - Get the debug setting
  
You can also get arbitrary key-value pairs stored in the session context.

Examples:
  get model
  get temperature
  get custom-key`,
		Category: command.CategoryREPL,
		Flags:    []command.Flag{},
	}
}

// Validate implements command.Interface
func (c *GetCommand) Validate() error {
	return nil
}
