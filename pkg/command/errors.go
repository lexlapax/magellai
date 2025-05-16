// ABOUTME: Error definitions for the command package
// ABOUTME: Provides standard errors for command validation and execution

package command

import "errors"

var (
	// ErrInvalidCommand indicates an invalid command configuration
	ErrInvalidCommand = errors.New("invalid command")

	// ErrCommandNotFound indicates the command was not found
	ErrCommandNotFound = errors.New("command not found")

	// ErrInvalidArguments indicates invalid command arguments
	ErrInvalidArguments = errors.New("invalid arguments")

	// ErrMissingRequiredFlag indicates a required flag is missing
	ErrMissingRequiredFlag = errors.New("missing required flag")

	// ErrInvalidFlagValue indicates an invalid flag value
	ErrInvalidFlagValue = errors.New("invalid flag value")

	// ErrCommandAlreadyRegistered indicates a command is already registered
	ErrCommandAlreadyRegistered = errors.New("command already registered")

	// ErrCommandCanceled indicates the command was canceled
	ErrCommandCanceled = errors.New("command canceled")

	// ErrInvalidCategory indicates an invalid command category
	ErrInvalidCategory = errors.New("invalid command category")

	// ErrNotAvailableInContext indicates the command is not available in the current context
	ErrNotAvailableInContext = errors.New("command not available in this context")
)
