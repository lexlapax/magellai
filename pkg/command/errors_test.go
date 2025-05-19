// ABOUTME: Tests for command package error definitions
// ABOUTME: Ensures all error constants behave correctly with error wrapping

package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInvalidCommand",
			err:      ErrInvalidCommand,
			expected: "invalid command",
		},
		{
			name:     "ErrCommandNotFound",
			err:      ErrCommandNotFound,
			expected: "command not found",
		},
		{
			name:     "ErrInvalidArguments",
			err:      ErrInvalidArguments,
			expected: "invalid arguments",
		},
		{
			name:     "ErrMissingArgument",
			err:      ErrMissingArgument,
			expected: "missing argument",
		},
		{
			name:     "ErrMissingRequiredFlag",
			err:      ErrMissingRequiredFlag,
			expected: "missing required flag",
		},
		{
			name:     "ErrInvalidFlagValue",
			err:      ErrInvalidFlagValue,
			expected: "invalid flag value",
		},
		{
			name:     "ErrCommandAlreadyRegistered",
			err:      ErrCommandAlreadyRegistered,
			expected: "command already registered",
		},
		{
			name:     "ErrCommandCanceled",
			err:      ErrCommandCanceled,
			expected: "command canceled",
		},
		{
			name:     "ErrInvalidCategory",
			err:      ErrInvalidCategory,
			expected: "invalid command category",
		},
		{
			name:     "ErrNotAvailableInContext",
			err:      ErrNotAvailableInContext,
			expected: "command not available in this context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name    string
		baseErr error
		wrapMsg string
	}{
		{
			name:    "WrappedInvalidCommand",
			baseErr: ErrInvalidCommand,
			wrapMsg: "failed to create command",
		},
		{
			name:    "WrappedCommandNotFound",
			baseErr: ErrCommandNotFound,
			wrapMsg: "unable to locate command",
		},
		{
			name:    "WrappedInvalidArguments",
			baseErr: ErrInvalidArguments,
			wrapMsg: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap the error
			wrapped := errors.Join(errors.New(tt.wrapMsg), tt.baseErr)

			// Should be able to unwrap to original error
			assert.True(t, errors.Is(wrapped, tt.baseErr))
			assert.Contains(t, wrapped.Error(), tt.wrapMsg)
			assert.Contains(t, wrapped.Error(), tt.baseErr.Error())
		})
	}
}

func TestErrorTypeAssertions(t *testing.T) {
	// Test that our errors are actual error types
	var err error

	err = ErrInvalidCommand
	assert.NotNil(t, err)

	err = ErrCommandNotFound
	assert.NotNil(t, err)

	err = ErrInvalidArguments
	assert.NotNil(t, err)

	err = ErrMissingArgument
	assert.NotNil(t, err)

	err = ErrMissingRequiredFlag
	assert.NotNil(t, err)

	err = ErrInvalidFlagValue
	assert.NotNil(t, err)

	err = ErrCommandAlreadyRegistered
	assert.NotNil(t, err)

	err = ErrCommandCanceled
	assert.NotNil(t, err)

	err = ErrInvalidCategory
	assert.NotNil(t, err)

	err = ErrNotAvailableInContext
	assert.NotNil(t, err)
}

func TestErrorComparison(t *testing.T) {
	// Test that errors can be compared with errors.Is
	testCases := []struct {
		name     string
		err1     error
		err2     error
		expected bool
	}{
		{
			name:     "Same error",
			err1:     ErrInvalidCommand,
			err2:     ErrInvalidCommand,
			expected: true,
		},
		{
			name:     "Different errors",
			err1:     ErrInvalidCommand,
			err2:     ErrCommandNotFound,
			expected: false,
		},
		{
			name:     "Wrapped error same base",
			err1:     errors.Join(errors.New("wrapped"), ErrInvalidCommand),
			err2:     ErrInvalidCommand,
			expected: true,
		},
		{
			name:     "Wrapped error different base",
			err1:     errors.Join(errors.New("wrapped"), ErrInvalidCommand),
			err2:     ErrCommandNotFound,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := errors.Is(tc.err1, tc.err2)
			assert.Equal(t, tc.expected, result)
		})
	}
}
