// ABOUTME: Tests for the command execution framework
// ABOUTME: Validates execution, hooks, validation, and error handling

package command_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandExecutor verifies basic command execution
func TestCommandExecutor(t *testing.T) {
	registry := command.NewRegistry()
	executor := command.NewExecutor(registry)

	// Create a simple command
	executed := false
	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name:        "test",
			Description: "Test command",
			Category:    command.CategoryShared,
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			executed = true
			fmt.Fprintf(exec.Stdout, "Hello from test command\n")
			return nil
		},
	)

	// Register the command
	require.NoError(t, registry.Register(cmd))

	// Execute by name
	exec := &command.ExecutionContext{
		Stdout: &strings.Builder{},
		Stderr: &strings.Builder{},
		Data:   make(map[string]interface{}),
	}

	err := executor.Execute(context.Background(), "test", exec)
	assert.NoError(t, err)
	assert.True(t, executed)
	assert.Contains(t, exec.Stdout.(*strings.Builder).String(), "Hello from test command")
}

// TestCommandNotFound verifies error when command doesn't exist
func TestCommandNotFound(t *testing.T) {
	registry := command.NewRegistry()
	executor := command.NewExecutor(registry)

	exec := &command.ExecutionContext{
		Data: make(map[string]interface{}),
	}

	err := executor.Execute(context.Background(), "nonexistent", exec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestPreExecuteHooks verifies pre-execution hooks
func TestPreExecuteHooks(t *testing.T) {
	registry := command.NewRegistry()

	hookCalled := false
	executor := command.NewExecutor(registry, command.WithPreExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
		hookCalled = true
		exec.Data["hook"] = "pre"
		return nil
	}))

	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name: "test",
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			assert.Equal(t, "pre", exec.Data["hook"])
			return nil
		},
	)

	require.NoError(t, registry.Register(cmd))

	exec := &command.ExecutionContext{
		Data: make(map[string]interface{}),
	}

	err := executor.Execute(context.Background(), "test", exec)
	assert.NoError(t, err)
	assert.True(t, hookCalled)
}

// TestPostExecuteHooks verifies post-execution hooks
func TestPostExecuteHooks(t *testing.T) {
	registry := command.NewRegistry()

	hookCalled := false
	executor := command.NewExecutor(registry, command.WithPostExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
		hookCalled = true
		exec.Data["hook"] = "post"
		return nil
	}))

	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name: "test",
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			exec.Data["command"] = "executed"
			return nil
		},
	)

	require.NoError(t, registry.Register(cmd))

	exec := &command.ExecutionContext{
		Data: make(map[string]interface{}),
	}

	err := executor.Execute(context.Background(), "test", exec)
	assert.NoError(t, err)
	assert.True(t, hookCalled)
	assert.Equal(t, "executed", exec.Data["command"])
	assert.Equal(t, "post", exec.Data["hook"])
}

// TestHookError verifies hook error handling
func TestHookError(t *testing.T) {
	registry := command.NewRegistry()

	tests := []struct {
		name    string
		preErr  error
		postErr error
		cmdErr  error
		wantErr string
	}{
		{
			name:    "pre-hook error",
			preErr:  errors.New("pre-hook failed"),
			wantErr: "pre-execute hook failed",
		},
		{
			name:    "post-hook error with successful command",
			postErr: errors.New("post-hook failed"),
			wantErr: "post-execute hook failed",
		},
		{
			name:    "both command and post-hook error",
			cmdErr:  errors.New("command failed"),
			postErr: errors.New("post-hook failed"),
			wantErr: "command failed", // Command error takes precedence
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var exec *command.CommandExecutor

			if tc.preErr != nil {
				exec = command.NewExecutor(registry, command.WithPreExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
					return tc.preErr
				}))
			} else if tc.postErr != nil {
				exec = command.NewExecutor(registry, command.WithPostExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
					return tc.postErr
				}))
			} else {
				exec = command.NewExecutor(registry)
			}

			cmd := command.NewSimpleCommand(
				&command.Metadata{
					Name: "test",
				},
				func(ctx context.Context, exec *command.ExecutionContext) error {
					return tc.cmdErr
				},
			)

			ctx := &command.ExecutionContext{
				Stdout: &strings.Builder{},
				Stderr: &strings.Builder{},
				Data:   make(map[string]interface{}),
			}

			err := exec.ExecuteCommand(context.Background(), cmd, ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestFlagValidation verifies flag validation
func TestFlagValidation(t *testing.T) {
	registry := command.NewRegistry()
	executor := command.NewExecutor(registry)

	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name: "test",
			Flags: []command.Flag{
				{
					Name:     "name",
					Type:     command.FlagTypeString,
					Required: true,
				},
				{
					Name:     "count",
					Type:     command.FlagTypeInt,
					Required: false,
				},
			},
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			return nil
		},
	)

	tests := []struct {
		name    string
		flags   map[string]interface{}
		wantErr bool
		errType command.ValidationErrorType
	}{
		{
			name:    "missing required flag",
			flags:   map[string]interface{}{},
			wantErr: true,
			errType: command.ValidationErrorMissingRequired,
		},
		{
			name: "invalid flag type",
			flags: map[string]interface{}{
				"name":  "test",
				"count": "not a number",
			},
			wantErr: true,
			errType: command.ValidationErrorInvalidType,
		},
		{
			name: "valid flags",
			flags: map[string]interface{}{
				"name":  "test",
				"count": 42,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exec := &command.ExecutionContext{
				Flags: command.NewFlags(tc.flags),
				Data:  make(map[string]interface{}),
			}

			err := executor.ExecuteCommand(context.Background(), cmd, exec)

			if tc.wantErr {
				assert.Error(t, err)
				var validationErr *command.ValidationError
				if errors.As(err, &validationErr) {
					assert.Equal(t, tc.errType, validationErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestParseAndExecute verifies argument parsing and execution
func TestParseAndExecute(t *testing.T) {
	registry := command.NewRegistry()

	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name: "echo",
			Flags: []command.Flag{
				{
					Name: "upper",
					Type: command.FlagTypeBool,
				},
			},
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			text := strings.Join(exec.Args, " ")
			if upper := exec.Flags.GetBool("upper"); upper {
				text = strings.ToUpper(text)
			}
			fmt.Fprint(exec.Stdout, text)
			return nil
		},
	)

	require.NoError(t, registry.Register(cmd))

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple echo",
			args:     []string{"echo", "hello", "world"},
			expected: "hello world",
		},
		{
			name:     "echo with flag",
			args:     []string{"echo", "--upper", "hello", "world"},
			expected: "HELLO WORLD",
		},
		{
			name:     "echo with flag value",
			args:     []string{"echo", "--upper=true", "hello"},
			expected: "HELLO",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := &strings.Builder{}
			executor := command.NewExecutor(registry, command.WithDefaultStreams(nil, output, nil))

			err := executor.ParseAndExecute(context.Background(), tc.args)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, output.String())
		})
	}
}

// TestStreamDefaults verifies default stream handling
func TestStreamDefaults(t *testing.T) {
	registry := command.NewRegistry()
	executor := command.NewExecutor(registry)

	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name: "test",
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			assert.NotNil(t, exec.Stdin)
			assert.NotNil(t, exec.Stdout)
			assert.NotNil(t, exec.Stderr)
			assert.NotNil(t, exec.Context)
			assert.NotNil(t, exec.Data)
			return nil
		},
	)

	ctx := &command.ExecutionContext{}
	err := executor.ExecuteCommand(context.Background(), cmd, ctx)
	assert.NoError(t, err)
}

// MockCommand for testing
type MockCommand struct {
	meta        *command.Metadata
	validateErr error
	executeErr  error
	executed    bool
}

func (m *MockCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	m.executed = true
	return m.executeErr
}

func (m *MockCommand) Metadata() *command.Metadata {
	return m.meta
}

func (m *MockCommand) Validate() error {
	return m.validateErr
}

// TestCommandValidation verifies command validation
func TestCommandValidation(t *testing.T) {
	registry := command.NewRegistry()
	executor := command.NewExecutor(registry)

	cmd := &MockCommand{
		meta: &command.Metadata{
			Name: "test",
		},
		validateErr: errors.New("validation failed"),
	}

	exec := &command.ExecutionContext{
		Data: make(map[string]interface{}),
	}

	err := executor.ExecuteCommand(context.Background(), cmd, exec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.False(t, cmd.executed)
}

// TestMultipleHooks verifies execution order of multiple hooks
func TestMultipleHooks(t *testing.T) {
	registry := command.NewRegistry()

	order := []string{}
	executor := command.NewExecutor(registry,
		command.WithPreExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
			order = append(order, "pre1")
			return nil
		}),
		command.WithPreExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
			order = append(order, "pre2")
			return nil
		}),
		command.WithPostExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
			order = append(order, "post1")
			return nil
		}),
		command.WithPostExecuteHook(func(ctx context.Context, cmd command.Interface, exec *command.ExecutionContext) error {
			order = append(order, "post2")
			return nil
		}),
	)

	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name: "test",
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			order = append(order, "command")
			return nil
		},
	)

	exec := &command.ExecutionContext{
		Data: make(map[string]interface{}),
	}

	err := executor.ExecuteCommand(context.Background(), cmd, exec)
	assert.NoError(t, err)
	assert.Equal(t, []string{"pre1", "pre2", "command", "post1", "post2"}, order)
}

// TestExecuteWithArgs verifies the convenience method
func TestExecuteWithArgs(t *testing.T) {
	registry := command.NewRegistry()
	executor := command.NewExecutor(registry, command.WithDefaultStreams(nil, io.Discard, io.Discard))

	cmd := command.NewSimpleCommand(
		&command.Metadata{
			Name: "test",
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			assert.Equal(t, []string{"arg1", "arg2"}, exec.Args)
			assert.Equal(t, "value", exec.Flags.GetString("key"))
			return nil
		},
	)

	require.NoError(t, registry.Register(cmd))

	err := executor.ExecuteWithArgs(
		context.Background(),
		"test",
		[]string{"arg1", "arg2"},
		map[string]interface{}{"key": "value"},
	)
	assert.NoError(t, err)
}
