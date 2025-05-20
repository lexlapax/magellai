// ABOUTME: Mock implementations for command interfaces
// ABOUTME: Provides reusable mock commands for testing

package mocks

import (
	"context"
	"sync"

	"github.com/lexlapax/magellai/pkg/command"
)

// MockCommand implements the command.Interface for testing
type MockCommand struct {
	mu           sync.RWMutex
	metadata     *command.Metadata
	executeFunc  func(ctx context.Context, exec *command.ExecutionContext) error
	validateFunc func() error
	executeCalls int
	lastExec     *command.ExecutionContext
}

// NewMockCommand creates a new mock command
func NewMockCommand(name, description string) *MockCommand {
	return &MockCommand{
		metadata: &command.Metadata{
			Name:        name,
			Description: description,
			Category:    command.CategoryShared,
			Flags:       []command.Flag{},
		},
	}
}

// WithCategory sets the command category
func (m *MockCommand) WithCategory(category command.Category) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata.Category = category
	return m
}

// WithAliases sets the command aliases
func (m *MockCommand) WithAliases(aliases ...string) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata.Aliases = aliases
	return m
}

// WithLongDescription sets the command's long description
func (m *MockCommand) WithLongDescription(desc string) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata.LongDescription = desc
	return m
}

// WithFlag adds a flag to the command
func (m *MockCommand) WithFlag(flag command.Flag) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata.Flags = append(m.metadata.Flags, flag)
	return m
}

// WithFlags sets all flags for the command
func (m *MockCommand) WithFlags(flags []command.Flag) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata.Flags = flags
	return m
}

// WithHidden sets the hidden state of the command
func (m *MockCommand) WithHidden(hidden bool) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata.Hidden = hidden
	return m
}

// WithDeprecated marks the command as deprecated
func (m *MockCommand) WithDeprecated(message string) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata.Deprecated = message
	return m
}

// WithExecuteFunc sets a custom execute function
func (m *MockCommand) WithExecuteFunc(fn func(ctx context.Context, exec *command.ExecutionContext) error) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executeFunc = fn
	return m
}

// WithValidateFunc sets a custom validation function
func (m *MockCommand) WithValidateFunc(fn func() error) *MockCommand {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validateFunc = fn
	return m
}

// Execute implements the command.Interface Execute method
func (m *MockCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	m.mu.Lock()
	m.executeCalls++
	m.lastExec = exec
	executeFunc := m.executeFunc
	m.mu.Unlock()

	if executeFunc != nil {
		return executeFunc(ctx, exec)
	}
	return nil
}

// Metadata implements the command.Interface Metadata method
func (m *MockCommand) Metadata() *command.Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metadata
}

// Validate implements the command.Interface Validate method
func (m *MockCommand) Validate() error {
	m.mu.RLock()
	validateFunc := m.validateFunc
	m.mu.RUnlock()

	if validateFunc != nil {
		return validateFunc()
	}
	return nil
}

// GetExecuteCalls returns the number of times Execute was called
func (m *MockCommand) GetExecuteCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.executeCalls
}

// GetLastExec returns the last execution context
func (m *MockCommand) GetLastExec() *command.ExecutionContext {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastExec
}

// MockSharedContext implements a mock command.SharedContext for testing
type MockSharedContext struct {
	mu    sync.RWMutex
	state map[string]interface{}
}

// NewMockSharedContext creates a new mock shared context
func NewMockSharedContext() *MockSharedContext {
	return &MockSharedContext{
		state: make(map[string]interface{}),
	}
}

// Get implements the SharedContext Get method
func (sc *MockSharedContext) Get(key string) (interface{}, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	value, exists := sc.state[key]
	return value, exists
}

// Set implements the SharedContext Set method
func (sc *MockSharedContext) Set(key string, value interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.state[key] = value
}

// Delete implements the SharedContext Delete method
func (sc *MockSharedContext) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.state, key)
}

// GetString retrieves a string value from the mock shared context
func (sc *MockSharedContext) GetString(key string) (string, bool) {
	value, exists := sc.Get(key)
	if !exists {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

// GetBool retrieves a boolean value from the mock shared context
func (sc *MockSharedContext) GetBool(key string) (bool, bool) {
	value, exists := sc.Get(key)
	if !exists {
		return false, false
	}

	b, ok := value.(bool)
	return b, ok
}

// GetInt retrieves an integer value from the mock shared context
func (sc *MockSharedContext) GetInt(key string) (int, bool) {
	value, exists := sc.Get(key)
	if !exists {
		return 0, false
	}

	i, ok := value.(int)
	return i, ok
}

// MockExecutionContext creates a mock execution context for testing
func MockExecutionContext(args []string) *command.ExecutionContext {
	return &command.ExecutionContext{
		Args:          args,
		Flags:         command.NewFlags(nil),
		Data:          make(map[string]interface{}),
		SharedContext: command.NewSharedContext(),
	}
}
