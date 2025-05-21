// ABOUTME: Demonstration test that shows context preservation between commands
// ABOUTME: Validates that state is maintained across multiple command executions

package repl

import (
	"context"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestSharedContextPreservation(t *testing.T) {
	// Create a shared context
	sharedContext := command.NewSharedContext()

	// Create an executor with the shared context (not used in this test, just demonstrating creation)
	registry := command.NewRegistry()
	_ = command.NewExecutor(registry, command.WithSharedContext(sharedContext))

	// Simulate setting temperature in one command
	t.Run("Set temperature command", func(t *testing.T) {
		sharedContext.Set(command.SharedContextTemperature, 0.8)

		// In a real scenario, the command would do this
		temp, exists := sharedContext.GetFloat64(command.SharedContextTemperature)
		assert.True(t, exists)
		assert.Equal(t, 0.8, temp)
	})

	// Simulate another command accessing the same context
	t.Run("Access temperature in another command", func(t *testing.T) {
		// The temperature should still be there
		temp, exists := sharedContext.GetFloat64(command.SharedContextTemperature)
		assert.True(t, exists)
		assert.Equal(t, 0.8, temp)

		// Demonstrate that helper methods work too
		assert.Equal(t, 0.8, sharedContext.Temperature())
	})

	// Simulate setting model in one command and accessing in another
	t.Run("Model preservation", func(t *testing.T) {
		// Set model using helper method
		sharedContext.SetModel("openai/gpt-4")

		// Access in "another command"
		model := sharedContext.Model()
		assert.Equal(t, "openai/gpt-4", model)
	})

	// Simulate complex state preservation
	t.Run("Multiple state values", func(t *testing.T) {
		// Set multiple values
		sharedContext.SetMaxTokens(2000)
		sharedContext.SetStream(true)
		sharedContext.SetDebug(true)

		// Access all values - they should all be preserved
		assert.Equal(t, 2000, sharedContext.MaxTokens())
		assert.True(t, sharedContext.Stream())
		assert.True(t, sharedContext.Debug())
		assert.Equal(t, "openai/gpt-4", sharedContext.Model()) // From previous test
		assert.Equal(t, 0.8, sharedContext.Temperature())      // From first test
	})
}

func TestREPLSharedContextIntegration(t *testing.T) {
	// Create test REPL with proper config
	config := &MockConfigInterface{
		values: make(map[string]interface{}),
	}
	config.SetValue("model.default", "openai/gpt-4o")
	repl, err := NewREPL(&REPLOptions{
		Config: config,
		Writer: nil,
		Reader: nil,
	})
	if err != nil {
		t.Fatalf("Failed to create REPL: %v", err)
	}

	// Set values through REPL commands
	t.Run("REPL preserves context", func(t *testing.T) {
		// Set temperature through command
		err := repl.setTemperature([]string{"0.7"})
		assert.NoError(t, err)

		// Temperature should be in both session and shared context
		assert.Equal(t, 0.7, repl.session.Conversation.Temperature)
		assert.Equal(t, 0.7, repl.sharedContext.Temperature())

		// Set model through command
		err = repl.switchModel([]string{"anthropic/claude-3"})
		assert.NoError(t, err)

		// Model should be in both session and shared context
		assert.Equal(t, "anthropic/claude-3", repl.session.Conversation.Model)
		assert.Equal(t, "anthropic/claude-3", repl.sharedContext.Model())
	})
}

// Demonstrate context usage in command execution
func TestCommandExecutionWithSharedContext(t *testing.T) {
	// Create shared context and set initial values
	sharedContext := command.NewSharedContext()
	sharedContext.SetModel("test/model")
	sharedContext.SetTemperature(0.5)

	// Create a simple test command that reads from shared context
	testCmd := command.NewSimpleCommand(
		&command.Metadata{
			Name:        "test",
			Description: "Test command",
			Category:    command.CategoryShared,
		},
		func(ctx context.Context, exec *command.ExecutionContext) error {
			// Access shared context values
			model := exec.SharedContext.Model()
			temp := exec.SharedContext.Temperature()

			// Verify they match what we set
			assert.Equal(t, "test/model", model)
			assert.Equal(t, 0.5, temp)

			// Modify a value
			exec.SharedContext.SetMaxTokens(1000)

			return nil
		},
	)

	// Create executor with shared context
	registry := command.NewRegistry()
	if err := registry.Register(testCmd); err != nil {
		t.Fatalf("Failed to register test command: %v", err)
	}
	executor := command.NewExecutor(registry, command.WithSharedContext(sharedContext))

	// Execute the command
	err := executor.Execute(context.Background(), "test", &command.ExecutionContext{})
	assert.NoError(t, err)

	// Verify the command's modification persisted
	assert.Equal(t, 1000, sharedContext.MaxTokens())
}
