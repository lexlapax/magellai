// ABOUTME: Command validation and error handling utilities
// ABOUTME: Provides validation for command arguments, flags, and execution context

package command

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// Validator provides validation for commands
type Validator struct {
	cmd Interface
}

// NewValidator creates a new command validator
func NewValidator(cmd Interface) *Validator {
	return &Validator{cmd: cmd}
}

// ValidateArgs validates command arguments
func (v *Validator) ValidateArgs(args []string) error {
	meta := v.cmd.Metadata()

	// Check minimum/maximum args if needed
	// This is a basic implementation - can be extended with min/max arg counts
	if meta.Name == "" {
		return ErrInvalidCommand
	}

	return nil
}

// ValidateFlags validates command flags
func (v *Validator) ValidateFlags(flags *Flags) error {
	meta := v.cmd.Metadata()

	// Check required flags
	for _, flag := range meta.Flags {
		if flag.Required {
			if !flags.Has(flag.Name) {
				return fmt.Errorf("%w: %s", ErrMissingRequiredFlag, flag.Name)
			}
		}
	}

	// Validate flag types
	for _, flag := range meta.Flags {
		if flags.Has(flag.Name) {
			value := flags.Get(flag.Name)
			if err := validateFlagValue(flag, value); err != nil {
				return fmt.Errorf("%w for flag '%s': %v", ErrInvalidFlagValue, flag.Name, err)
			}
		}
	}

	return nil
}

// ValidateContext validates the execution context
func (v *Validator) ValidateContext(ctx *ExecutionContext) error {
	if ctx == nil {
		return fmt.Errorf("execution context is nil")
	}

	// Validate args
	if err := v.ValidateArgs(ctx.Args); err != nil {
		return err
	}

	// Validate flags
	if err := v.ValidateFlags(ctx.Flags); err != nil {
		return err
	}

	// Check for required I/O streams
	if ctx.Stdout == nil {
		return fmt.Errorf("stdout is required")
	}

	if ctx.Stderr == nil {
		return fmt.Errorf("stderr is required")
	}

	return nil
}

// validateFlagValue validates a single flag value against its type
func validateFlagValue(flag Flag, value interface{}) error {
	switch flag.Type {
	case FlagTypeString:
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

	case FlagTypeInt:
		switch v := value.(type) {
		case int:
			// Already correct type
		case string:
			if _, err := strconv.Atoi(v); err != nil {
				return fmt.Errorf("invalid integer value: %s", v)
			}
		default:
			return fmt.Errorf("expected int, got %T", value)
		}

	case FlagTypeBool:
		switch v := value.(type) {
		case bool:
			// Already correct type
		case string:
			if _, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("invalid boolean value: %s", v)
			}
		default:
			return fmt.Errorf("expected bool, got %T", value)
		}

	case FlagTypeFloat:
		switch v := value.(type) {
		case float64:
			// Already correct type
		case float32:
			// Acceptable
		case string:
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return fmt.Errorf("invalid float value: %s", v)
			}
		default:
			return fmt.Errorf("expected float, got %T", value)
		}

	case FlagTypeDuration:
		switch v := value.(type) {
		case time.Duration:
			// Already correct type
		case string:
			if _, err := time.ParseDuration(v); err != nil {
				return fmt.Errorf("invalid duration value: %s", v)
			}
		default:
			return fmt.Errorf("expected duration, got %T", value)
		}

	case FlagTypeStringSlice:
		switch value.(type) {
		case []string:
			// Already correct type
		case string:
			// Single string is acceptable for a slice
		default:
			return fmt.Errorf("expected string slice, got %T", value)
		}
	}

	return nil
}

// ParseFlagValue converts a string value to the appropriate type
func ParseFlagValue(flag Flag, value string) (interface{}, error) {
	switch flag.Type {
	case FlagTypeString:
		return value, nil

	case FlagTypeInt:
		return strconv.Atoi(value)

	case FlagTypeBool:
		return strconv.ParseBool(value)

	case FlagTypeFloat:
		return strconv.ParseFloat(value, 64)

	case FlagTypeDuration:
		return time.ParseDuration(value)

	case FlagTypeStringSlice:
		// For now, return as single element slice
		// Can be enhanced to parse comma-separated values
		return []string{value}, nil

	default:
		return nil, fmt.Errorf("unknown flag type: %v", flag.Type)
	}
}

// DefaultExecutor provides a default command executor that validates before execution
func DefaultExecutor(cmd Interface, executor Executor) Executor {
	return func(ctx context.Context, exec *ExecutionContext) error {
		// Create validator
		validator := NewValidator(cmd)

		// Validate the execution context
		if err := validator.ValidateContext(exec); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Execute the command
		return executor(ctx, exec)
	}
}
