// ABOUTME: Implements replapi.REPL interface for the REPL struct
// ABOUTME: Provides accessor methods required by the interface

package repl

import (
	"errors"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/replapi"
)

// Ensure REPL implements replapi.REPL
var _ replapi.REPL = (*REPL)(nil)

// GetSession returns the current session
func (r *REPL) GetSession() *domain.Session {
	return r.session
}

// ExecuteCommand executes a command with the given name and arguments
func (r *REPL) ExecuteCommand(cmdName string, args []string) error {
	logging.LogDebug("Executing command via API", "command", cmdName, "args", args)

	// Convert command name to prefixed version if needed
	if !strings.HasPrefix(cmdName, "/") && !strings.HasPrefix(cmdName, ":") {
		// Default to slash command
		cmdName = "/" + cmdName
	}

	// Call the handleCommand method with the proper prefix
	return r.handleCommand(cmdName)
}

// ProcessMessage processes a user message and returns the assistant's response
func (r *REPL) ProcessMessage(content string) (string, error) {
	logging.LogDebug("Processing message via API", "length", len(content))

	// Add message to conversation
	msg := domain.NewMessage("", domain.MessageRoleUser, content)
	r.session.Conversation.AddMessage(*msg)

	// Get last message after processing
	err := r.processMessage(content)
	if err != nil {
		return "", err
	}

	// Return the assistant's response
	lastMsg := r.session.Conversation.GetLastMessage()
	if lastMsg != nil && lastMsg.Role == domain.MessageRoleAssistant {
		return lastMsg.Content, nil
	}

	return "", errors.New("no assistant response generated")
}
