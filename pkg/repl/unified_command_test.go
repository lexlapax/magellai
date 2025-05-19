// ABOUTME: Tests for unified command system integration
// ABOUTME: Validates command registration, routing, and execution

package repl

import (
	"bytes"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestUnifiedCommandSystem(t *testing.T) {
	// Create a REPL with registry
	cfg := NewMockConfig()
	var buf bytes.Buffer
	r := &REPL{
		config: cfg,
		writer: &buf,
		session: &Session{
			Conversation: &Conversation{},
			Metadata:     make(map[string]interface{}),
		},
		registry:      command.NewRegistry(),
		cmdHistory:    make([]string, 0),
		sharedContext: command.NewSharedContext(),
	}

	// Register commands
	err := RegisterREPLCommands(r, r.registry)
	assert.NoError(t, err)

	// Test slash command routing
	t.Run("slash commands", func(t *testing.T) {
		buf.Reset()
		err := r.handleCommand("/help")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "magellai chat")
	})

	// Test colon command routing
	t.Run("colon commands", func(t *testing.T) {
		buf.Reset()
		err := r.handleSpecialCommand(":multiline")
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "Multi-line mode")
	})

	// Test command history
	t.Run("command history", func(t *testing.T) {
		assert.Len(t, r.cmdHistory, 2)
		assert.Equal(t, "/help", r.cmdHistory[0])
		assert.Equal(t, ":multiline", r.cmdHistory[1])
	})

	// Test unknown command
	t.Run("unknown command", func(t *testing.T) {
		err := r.handleCommand("/unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})

	// Test command registry
	t.Run("command registry", func(t *testing.T) {
		// Check that help command is registered
		cmd, err := r.registry.Get("help")
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Equal(t, "help", cmd.Metadata().Name)

		// Check colon command
		cmd, err = r.registry.Get(":multiline")
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Equal(t, ":multiline", cmd.Metadata().Name)
	})

	// Test command aliases
	t.Run("command aliases", func(t *testing.T) {
		// Check that alias is registered
		cmd, err := r.registry.Get("h")
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Equal(t, "help", cmd.Metadata().Name)

		// Check colon command alias
		cmd, err = r.registry.Get(":ml")
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Equal(t, ":multiline", cmd.Metadata().Name)
	})
}
