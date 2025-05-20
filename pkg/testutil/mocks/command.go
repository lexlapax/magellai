// ABOUTME: Mock implementations for command interface
// ABOUTME: Provides reusable mock commands for testing

package mocks

import (
	"context"
	"sync"

	"github.com/lexlapax/magellai/pkg/command"
)

// MockCommand implements the Command interface for testing
type MockCommand struct {
	mu            sync.Mutex
	executeFunc   func(context.Context, *command.ExecutionContext) error
	validateFunc  func(*command.ExecutionContext) error
	callCounts    map[string]int
	errorToReturn error
	lastContext   *command.ExecutionContext
}

// NewMockCommand creates a new mock command
func NewMockCommand() *MockCommand {
	return &MockCommand{
		callCounts: make(map[string]int),
	}
}

// SetExecuteFunc sets a custom execute function
func (mc *MockCommand) SetExecuteFunc(f func(context.Context, *command.ExecutionContext) error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.executeFunc = f
}

// SetValidateFunc sets a custom validate function
func (mc *MockCommand) SetValidateFunc(f func(*command.ExecutionContext) error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.validateFunc = f
}

// SetError sets the error to return
func (mc *MockCommand) SetError(err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.errorToReturn = err
}

// GetCallCount returns the call count for a method
func (mc *MockCommand) GetCallCount(method string) int {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.callCounts[method]
}

// Execute implements the Command interface
func (mc *MockCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.callCounts["Execute"]++
	mc.lastContext = exec

	if mc.errorToReturn != nil {
		return mc.errorToReturn
	}

	if mc.executeFunc != nil {
		return mc.executeFunc(ctx, exec)
	}

	return nil
}

// Validate implements the Command interface
func (mc *MockCommand) Validate(exec *command.ExecutionContext) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.callCounts["Validate"]++

	if mc.errorToReturn != nil {
		return mc.errorToReturn
	}

	if mc.validateFunc != nil {
		return mc.validateFunc(exec)
	}

	return nil
}

// GetLastContext returns the last execution context
func (mc *MockCommand) GetLastContext() *command.ExecutionContext {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.lastContext
}

// Metadata implements the Command interface
func (mc *MockCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "mock",
		Description: "Mock command for testing",
		Category:    command.CategoryREPL,
		Flags:       []command.Flag{},
		Aliases:     []string{"m", "mk"},
	}
}

// MockConfigInterface implements a minimal config interface for testing
type MockConfigInterface struct {
	mu     sync.Mutex
	values map[string]interface{}
	errors map[string]error
}

// NewMockConfigInterface creates a new mock config
func NewMockConfigInterface() *MockConfigInterface {
	return &MockConfigInterface{
		values: make(map[string]interface{}),
		errors: make(map[string]error),
	}
}

// SetValue sets a configuration value
func (mc *MockConfigInterface) SetValue(key string, value interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.values[key] = value
}

// SetError sets an error for a specific key
func (mc *MockConfigInterface) SetError(key string, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.errors[key] = err
}

// Get implements a getter for testing
func (mc *MockConfigInterface) Get(key string) (interface{}, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if err, exists := mc.errors[key]; exists {
		return nil, err
	}

	value, exists := mc.values[key]
	if !exists {
		return nil, nil
	}

	return value, nil
}
