// ABOUTME: Command system for CLI and REPL command execution
// ABOUTME: Provides unified interface for all commands with execution context

/*
Package command implements the unified command system for Magellai.

This package provides a consistent framework for defining, discovering, and executing
commands across different contexts (CLI, REPL, API). It uses a registry pattern
to manage commands and maintains execution context across command invocations.

Key Components:
  - Interface: Core command interface that all commands implement
  - Registry: Central registry for command discovery and lookup
  - Executor: Executes commands with proper context management
  - ExecutionContext: Shared context for command execution
  - Discovery: Auto-discovery of commands using reflection
  - Constants: Shared constants for command names and categories
  - Flags: Common flag definitions and handling

The command system supports several advanced features:
  - Command categories and help text generation
  - Consistent error handling and validation
  - Shared context across command invocations
  - Automatic command aliasing
  - Integration with both CLI and REPL environments

Usage:

	// Register a command
	registry := command.NewRegistry()
	registry.Register(myCommand)

	// Execute a command
	ctx := context.Background()
	execCtx := &command.ExecutionContext{
	    Args:   []string{"--option", "value"},
	    Config: config,
	    Stdin:  os.Stdin,
	    Stdout: os.Stdout,
	    Stderr: os.Stderr,
	}
	err := registry.Execute(ctx, "command-name", execCtx)

The package is designed to be extensible while providing a consistent
interface for all command-driven interactions with the application.
*/
package command
