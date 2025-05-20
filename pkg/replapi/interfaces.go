// ABOUTME: Defines interfaces and types shared between REPL and command packages
// ABOUTME: Breaks circular dependency between pkg/command/core and pkg/repl

package replapi

import (
	"io"

	"github.com/lexlapax/magellai/pkg/domain"
)

// ConfigInterface defines the minimal interface needed for configuration
type ConfigInterface interface {
	GetString(key string) string
	GetBool(key string) bool
	Get(key string) interface{}
	Exists(key string) bool
	SetValue(key string, value interface{}) error
}

// REPLOptions contains initialization options for REPL
type REPLOptions struct {
	Config      ConfigInterface
	StorageDir  string
	PromptStyle string
	SessionID   string // Optional: resume existing session
	Model       string // Optional: override default model
	Writer      io.Writer
	Reader      io.Reader
}

// REPL defines the minimal interface for chat REPL functionality
type REPL interface {
	// Run starts the REPL loop and returns when complete
	Run() error

	// ExecuteCommand executes a command with the given name and arguments
	ExecuteCommand(cmdName string, args []string) error

	// GetSession gets the current session
	GetSession() *domain.Session

	// ProcessMessage processes a user message and returns the response
	ProcessMessage(content string) (string, error)
}

// Factory defines a function type that creates a new REPL instance
type Factory func(opts *REPLOptions) (REPL, error)
