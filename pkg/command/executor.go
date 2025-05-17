// ABOUTME: Command execution framework for running commands
// ABOUTME: Provides validation, execution, and error handling

package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
)

// ValidationErrorType represents the type of validation error
type ValidationErrorType int

const (
	// ValidationErrorMissingRequired indicates a required flag was not provided
	ValidationErrorMissingRequired ValidationErrorType = iota
	// ValidationErrorInvalidType indicates a flag value has the wrong type
	ValidationErrorInvalidType
	// ValidationErrorInvalidArgumentCount indicates wrong number of arguments
	ValidationErrorInvalidArgumentCount
	// ValidationErrorInvalidArgumentType indicates an argument has the wrong type
	ValidationErrorInvalidArgumentType
	// ValidationErrorCustom indicates a custom validation error
	ValidationErrorCustom
)

// ValidationError represents a detailed command validation error
type ValidationError struct {
	Command string
	Flag    string
	Type    ValidationErrorType
	Message string
}

func (e *ValidationError) Error() string {
	if e.Flag != "" {
		return fmt.Sprintf("validation error in command %s (flag %s): %s", e.Command, e.Flag, e.Message)
	}
	return fmt.Sprintf("validation error in command %s: %s", e.Command, e.Message)
}

// CommandExecutor handles command execution with validation and error handling
type CommandExecutor struct {
	registry      *Registry
	defaultStdin  io.Reader
	defaultStdout io.Writer
	defaultStderr io.Writer
	preExecute    []ExecutorHook
	postExecute   []ExecutorHook
}

// ExecutorHook is a function that runs before or after command execution
type ExecutorHook func(ctx context.Context, cmd Interface, exec *ExecutionContext) error

// ExecutorOption configures the executor
type ExecutorOption func(*CommandExecutor)

// NewExecutor creates a new command executor
func NewExecutor(registry *Registry, opts ...ExecutorOption) *CommandExecutor {
	logging.LogDebug("Creating new command executor", "registrySize", len(registry.commands))

	e := &CommandExecutor{
		registry:      registry,
		defaultStdin:  os.Stdin,
		defaultStdout: os.Stdout,
		defaultStderr: os.Stderr,
		preExecute:    make([]ExecutorHook, 0),
		postExecute:   make([]ExecutorHook, 0),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// WithDefaultStreams sets default I/O streams
func WithDefaultStreams(stdin io.Reader, stdout, stderr io.Writer) ExecutorOption {
	return func(e *CommandExecutor) {
		e.defaultStdin = stdin
		e.defaultStdout = stdout
		e.defaultStderr = stderr
	}
}

// WithPreExecuteHook adds a pre-execution hook
func WithPreExecuteHook(hook ExecutorHook) ExecutorOption {
	return func(e *CommandExecutor) {
		e.preExecute = append(e.preExecute, hook)
	}
}

// WithPostExecuteHook adds a post-execution hook
func WithPostExecuteHook(hook ExecutorHook) ExecutorOption {
	return func(e *CommandExecutor) {
		e.postExecute = append(e.postExecute, hook)
	}
}

// Execute runs a command by name with the given context
func (e *CommandExecutor) Execute(ctx context.Context, name string, exec *ExecutionContext) error {
	start := time.Now()
	logging.LogDebug("Executing command", "name", name, "args", exec.Args)

	// Get command from registry
	cmd, err := e.registry.Get(name)
	if err != nil {
		logging.LogError(err, "Command not found", "name", name)
		return err
	}

	// Execute the command
	result := e.ExecuteCommand(ctx, cmd, exec)
	duration := time.Since(start)
	logging.LogDebug("Command total execution time", "name", name, "duration", duration)
	return result
}

// ExecuteCommand runs a specific command with validation and hooks
func (e *CommandExecutor) ExecuteCommand(ctx context.Context, cmd Interface, exec *ExecutionContext) error {
	meta := cmd.Metadata()
	flagCount := 0
	if exec.Flags != nil {
		flagCount = len(exec.Flags.values)
	}
	logging.LogDebug("Starting command execution", "name", meta.Name, "args", exec.Args, "flagCount", flagCount)

	// Set defaults if not provided
	if exec.Stdin == nil {
		exec.Stdin = e.defaultStdin
	}
	if exec.Stdout == nil {
		exec.Stdout = e.defaultStdout
	}
	if exec.Stderr == nil {
		exec.Stderr = e.defaultStderr
	}
	if exec.Context == nil {
		exec.Context = ctx
	}
	if exec.Data == nil {
		exec.Data = make(map[string]interface{})
	}

	// Validate the command
	logging.LogDebug("Validating command", "name", meta.Name)
	if err := e.validateCommand(cmd, exec); err != nil {
		logging.LogError(err, "Command validation failed", "command", meta.Name)
		return err
	}
	logging.LogDebug("Command validation successful", "name", meta.Name)

	// Run pre-execution hooks
	for i, hook := range e.preExecute {
		logging.LogDebug("Running pre-execution hook", "hookIndex", i, "command", meta.Name)
		if err := hook(ctx, cmd, exec); err != nil {
			logging.LogError(err, "Pre-execution hook failed", "hookIndex", i, "command", meta.Name)
			return fmt.Errorf("pre-execute hook failed: %w", err)
		}
	}

	// Execute the command
	logging.LogInfo("Executing command", "name", meta.Name, "args", exec.Args)
	err := cmd.Execute(ctx, exec)

	if err != nil {
		logging.LogError(err, "Command execution failed", "command", meta.Name)
	} else {
		logging.LogDebug("Command execution completed successfully", "name", meta.Name)
	}

	// Run post-execution hooks (even if command failed)
	for i, hook := range e.postExecute {
		logging.LogDebug("Running post-execution hook", "hookIndex", i, "command", meta.Name)
		if hookErr := hook(ctx, cmd, exec); hookErr != nil {
			// If command succeeded but post-hook failed, return hook error
			if err == nil {
				logging.LogError(hookErr, "Post-execution hook failed", "hookIndex", i, "command", meta.Name)
				return fmt.Errorf("post-execute hook failed: %w", hookErr)
			}
			// If both failed, log hook error but return original error
			logging.LogWarn("Post-execution hook failed after command error", "hookError", hookErr, "commandError", err, "hookIndex", i, "command", meta.Name)
			fmt.Fprintf(exec.Stderr, "post-execute hook failed: %v\n", hookErr)
		}
	}

	return err
}

// validateCommand validates command and execution context
func (e *CommandExecutor) validateCommand(cmd Interface, exec *ExecutionContext) error {
	// Validate the command itself
	logging.LogDebug("Validating command structure", "command", cmd.Metadata().Name)
	if err := cmd.Validate(); err != nil {
		logging.LogError(err, "Command structure validation failed", "command", cmd.Metadata().Name)
		return fmt.Errorf("command validation failed: %w", err)
	}

	meta := cmd.Metadata()

	// Validate required flags
	for _, flag := range meta.Flags {
		if flag.Required {
			if !exec.Flags.Has(flag.Name) {
				logging.LogError(nil, "Missing required flag", "command", meta.Name, "flag", flag.Name)
				return &ValidationError{
					Command: meta.Name,
					Flag:    flag.Name,
					Type:    ValidationErrorMissingRequired,
					Message: fmt.Sprintf("required flag '--%s' not provided", flag.Name),
				}
			}
			logging.LogDebug("Required flag present", "command", meta.Name, "flag", flag.Name)
		}
	}

	// Validate flag types
	for _, flag := range meta.Flags {
		if exec.Flags.Has(flag.Name) {
			value := exec.Flags.Get(flag.Name)
			logging.LogDebug("Validating flag type", "command", meta.Name, "flag", flag.Name, "expectedType", flag.Type, "value", value)
			if err := validateFlagType(flag, value); err != nil {
				logging.LogError(err, "Invalid flag type", "command", meta.Name, "flag", flag.Name, "expectedType", flag.Type, "actualValue", value)
				return &ValidationError{
					Command: meta.Name,
					Flag:    flag.Name,
					Type:    ValidationErrorInvalidType,
					Message: err.Error(),
				}
			}
		}
	}

	return nil
}

// validateFlagType validates that a flag value matches the expected type
func validateFlagType(flag Flag, value interface{}) error {
	switch flag.Type {
	case FlagTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case FlagTypeInt:
		switch v := value.(type) {
		case int, int32, int64:
			// Valid integer types
		case float64:
			// Check if float can be safely converted to int
			if v != float64(int(v)) {
				return fmt.Errorf("expected integer, got float with decimal: %v", v)
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case FlagTypeBool:
		// Boolean flags can be set without a value (implicitly true)
		switch v := value.(type) {
		case bool:
			// Valid boolean type
		case string:
			// Accept string values for booleans if they're valid
			if v != "true" && v != "false" && v != "True" && v != "False" && v != "TRUE" && v != "FALSE" {
				return fmt.Errorf("expected boolean, got invalid string: %s", v)
			}
		default:
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case FlagTypeFloat:
		switch value.(type) {
		case float32, float64:
			// Valid float types
		case int, int32, int64:
			// Integers can be treated as floats
		default:
			return fmt.Errorf("expected float, got %T", value)
		}
	case FlagTypeStringSlice:
		switch v := value.(type) {
		case []string:
			// Valid string slice
		case []interface{}:
			// Check if all elements are strings
			for i, elem := range v {
				if _, ok := elem.(string); !ok {
					return fmt.Errorf("expected string at index %d, got %T", i, elem)
				}
			}
		default:
			return fmt.Errorf("expected string slice, got %T", value)
		}
	}
	return nil
}

// ExecuteWithArgs is a convenience method for executing with args and flags
func (e *CommandExecutor) ExecuteWithArgs(ctx context.Context, name string, args []string, flags map[string]interface{}) error {
	exec := &ExecutionContext{
		Args:  args,
		Flags: NewFlags(flags),
		Data:  make(map[string]interface{}),
	}
	return e.Execute(ctx, name, exec)
}

// ParseAndExecute parses command line arguments and executes the command
func (e *CommandExecutor) ParseAndExecute(ctx context.Context, args []string) error {
	logging.LogDebug("Parsing and executing command", "argCount", len(args))

	if len(args) == 0 {
		logging.LogError(ErrMissingArgument, "No command provided")
		return ErrMissingArgument
	}

	// First argument is the command name
	cmdName := args[0]
	logging.LogDebug("Command name extracted", "name", cmdName)

	// Get the command to check its flags
	cmd, err := e.registry.Get(cmdName)
	if err != nil {
		logging.LogError(err, "Failed to get command from registry", "name", cmdName)
		return err
	}

	// Parse remaining arguments with command metadata
	logging.LogDebug("Parsing arguments", "name", cmdName, "remainingArgs", args[1:])
	parsedArgs, parsedFlags, err := parseArgsWithMetadata(args[1:], cmd.Metadata())
	if err != nil {
		logging.LogError(err, "Failed to parse arguments", "command", cmdName)
		return fmt.Errorf("failed to parse arguments: %w", err)
	}
	logging.LogDebug("Arguments parsed", "command", cmdName, "args", parsedArgs, "flagCount", len(parsedFlags))

	// Create execution context
	exec := &ExecutionContext{
		Args:  parsedArgs,
		Flags: NewFlags(parsedFlags),
		Data:  make(map[string]interface{}),
	}

	return e.ExecuteCommand(ctx, cmd, exec)
}

// parseArgsWithMetadata separates positional arguments from flags with knowledge of expected flags
func parseArgsWithMetadata(args []string, meta *Metadata) ([]string, map[string]interface{}, error) {
	positional := make([]string, 0)
	flags := make(map[string]interface{})

	// Build a map of flag names to their types
	flagTypes := make(map[string]FlagType)
	for _, flag := range meta.Flags {
		flagTypes[flag.Name] = flag.Type
		if flag.Short != "" {
			flagTypes[flag.Short] = flag.Type
		}
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		// Check if it's a flag
		if len(arg) > 0 && arg[0] == '-' {
			flagName, flagValue, consumed, err := parseFlagWithType(args[i:], flagTypes)
			if err != nil {
				return nil, nil, err
			}
			flags[flagName] = flagValue
			i += consumed
		} else {
			// Positional argument
			positional = append(positional, arg)
			i++
		}
	}

	return positional, flags, nil
}

// parseFlagWithType parses a single flag and its value with knowledge of the expected type
func parseFlagWithType(args []string, flagTypes map[string]FlagType) (string, interface{}, int, error) {
	if len(args) == 0 {
		return "", nil, 0, fmt.Errorf("empty flag")
	}

	arg := args[0]
	consumed := 1

	// Handle different flag formats
	var flagName string
	var hasValue bool
	var value string

	if len(arg) > 2 && arg[:2] == "--" {
		// Long flag: --flag or --flag=value
		flagName = arg[2:]
		if idx := strings.Index(flagName, "="); idx >= 0 {
			value = flagName[idx+1:]
			flagName = flagName[:idx]
			hasValue = true
		}
	} else if len(arg) > 1 && arg[0] == '-' {
		// Short flag: -f or -f value
		flagName = arg[1:]
	} else {
		return "", nil, 0, fmt.Errorf("invalid flag format: %s", arg)
	}

	// Get the expected type for this flag
	flagType, knownFlag := flagTypes[flagName]

	// If we have a value already (from --flag=value), parse it
	if hasValue {
		parsedValue := parseValueWithType(value, flagType)
		return flagName, parsedValue, consumed, nil
	}

	// For flags without values, check if it's a boolean flag
	if knownFlag && flagType == FlagTypeBool {
		return flagName, true, consumed, nil
	}

	// For non-boolean flags, try to get the value from the next argument
	if consumed < len(args) {
		nextArg := args[consumed]
		// Check if the next argument looks like a flag
		if len(nextArg) > 0 && nextArg[0] == '-' {
			// Next arg is a flag, so this flag has no value
			if knownFlag && flagType == FlagTypeBool {
				return flagName, true, consumed, nil
			}
			// Non-boolean flag without value is an error
			return "", nil, 0, fmt.Errorf("flag %s requires a value", flagName)
		}
		// Use next arg as value
		parsedValue := parseValueWithType(nextArg, flagType)
		return flagName, parsedValue, consumed + 1, nil
	}

	// No next argument available
	if knownFlag && flagType == FlagTypeBool {
		return flagName, true, consumed, nil
	}

	// Unknown flag or non-boolean without value
	return flagName, true, consumed, nil
}

// parseValueWithType parses a string value into the appropriate type
func parseValueWithType(s string, expectedType FlagType) interface{} {
	switch expectedType {
	case FlagTypeBool:
		if s == "true" || s == "True" || s == "TRUE" || s == "1" || s == "yes" || s == "Yes" || s == "YES" {
			return true
		}
		if s == "false" || s == "False" || s == "FALSE" || s == "0" || s == "no" || s == "No" || s == "NO" {
			return false
		}
		// Invalid boolean string will be caught in validation
		return s

	case FlagTypeInt:
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return int(n)
		}
		// Try float and check if it's a whole number
		if f, err := strconv.ParseFloat(s, 64); err == nil && f == float64(int64(f)) {
			return int(f)
		}
		// Invalid integer will be caught in validation
		return s

	case FlagTypeFloat:
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		// Invalid float will be caught in validation
		return s

	case FlagTypeString, FlagTypeDuration:
		return s

	case FlagTypeStringSlice:
		// For now, treat as single string; proper slice handling would need multiple values
		return []string{s}

	default:
		// Unknown type, return as string
		return s
	}
}
