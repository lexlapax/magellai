// ABOUTME: Core interfaces for the unified command system
// ABOUTME: Defines command interface, metadata, and context for CLI/REPL/API use

package command

import (
	"context"
	"io"
)

// Interface defines the main command interface for all commands
type Interface interface {
	// Execute runs the command with the given context
	Execute(ctx context.Context, exec *ExecutionContext) error

	// Metadata returns the command's metadata
	Metadata() *Metadata

	// Validate checks if the command can be executed with given arguments
	Validate() error
}

// ExecutionContext provides runtime context for command execution
type ExecutionContext struct {
	// Command arguments (positional)
	Args []string

	// Command flags (key-value pairs)
	Flags map[string]interface{}

	// Input/Output streams
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Additional context data
	Data map[string]interface{}

	// Global configuration
	Config interface{} // Will be replaced with *config.Config when integrated

	// Parent context for cancellation
	Context context.Context
}

// Metadata describes a command's properties
type Metadata struct {
	// Name is the primary command name
	Name string

	// Aliases are alternative names for the command
	Aliases []string

	// Description is a short description of the command
	Description string

	// LongDescription provides detailed information about the command
	LongDescription string

	// Category defines the command type
	Category Category

	// Flags defines the command-specific flags
	Flags []Flag

	// Hidden indicates if the command should be hidden from help
	Hidden bool

	// Deprecated marks the command as deprecated
	Deprecated string
}

// Category defines command availability across interfaces
type Category int

const (
	// CategoryShared - available in CLI, REPL, and API
	CategoryShared Category = iota

	// CategoryCLI - only available in CLI
	CategoryCLI

	// CategoryREPL - only available in REPL
	CategoryREPL

	// CategoryAPI - only available in API (future)
	CategoryAPI
)

// Flag defines a command flag
type Flag struct {
	// Name is the flag name
	Name string

	// Short is the short flag name (single character)
	Short string

	// Description describes the flag
	Description string

	// Type is the flag value type
	Type FlagType

	// Default is the default value
	Default interface{}

	// Required indicates if the flag is required
	Required bool

	// Multiple allows multiple values
	Multiple bool
}

// FlagType defines the type of a flag value
type FlagType int

const (
	// FlagTypeString is a string flag
	FlagTypeString FlagType = iota

	// FlagTypeInt is an integer flag
	FlagTypeInt

	// FlagTypeBool is a boolean flag
	FlagTypeBool

	// FlagTypeFloat is a float flag
	FlagTypeFloat

	// FlagTypeDuration is a duration flag
	FlagTypeDuration

	// FlagTypeStringSlice is a string slice flag
	FlagTypeStringSlice
)

// Executor is a function that executes a command
type Executor func(ctx context.Context, exec *ExecutionContext) error

// SimpleCommand provides a basic implementation of Interface
type SimpleCommand struct {
	meta     *Metadata
	executor Executor
}

// NewSimpleCommand creates a new simple command
func NewSimpleCommand(meta *Metadata, executor Executor) *SimpleCommand {
	return &SimpleCommand{
		meta:     meta,
		executor: executor,
	}
}

// Execute implements Interface
func (c *SimpleCommand) Execute(ctx context.Context, exec *ExecutionContext) error {
	return c.executor(ctx, exec)
}

// Metadata implements Interface
func (c *SimpleCommand) Metadata() *Metadata {
	return c.meta
}

// Validate implements Interface
func (c *SimpleCommand) Validate() error {
	// Basic validation - can be extended
	if c.meta.Name == "" {
		return ErrInvalidCommand
	}
	return nil
}
