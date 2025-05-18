// ABOUTME: Provides readline functionality for REPL including tab completion
// ABOUTME: Handles interactive input with history and completion support

package repl

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
)

// ReadlineConfig contains configuration for readline
type ReadlineConfig struct {
	Prompt           string
	HistoryFile      string
	EnableCompletion bool
	MultilineMode    bool
}

// ReadlineInterface wraps readline functionality
type ReadlineInterface struct {
	Instance *readline.Instance
	config   *ReadlineConfig
}

// NewReadlineInterface creates a new readline interface
func NewReadlineInterface(config *ReadlineConfig) (*ReadlineInterface, error) {
	logging.LogDebug("Creating readline interface", "prompt", config.Prompt)

	// Create readline config
	readlineConfig := &readline.Config{
		Prompt:      config.Prompt,
		HistoryFile: config.HistoryFile,
		EOFPrompt:   "exit",
	}

	// Setup auto completion if enabled
	if config.EnableCompletion {
		readlineConfig.AutoComplete = &replCompleter{
			commands: getCommandNames(),
		}
	}

	// Create readline instance
	instance, err := readline.NewEx(readlineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	return &ReadlineInterface{
		Instance: instance,
		config:   config,
	}, nil
}

// ReadLine reads a line with completion and history support
func (r *ReadlineInterface) ReadLine() (string, error) {
	return r.Instance.Readline()
}

// SetPrompt changes the prompt
func (r *ReadlineInterface) SetPrompt(prompt string) {
	r.Instance.SetPrompt(prompt)
}

// Close closes the readline interface
func (r *ReadlineInterface) Close() error {
	return r.Instance.Close()
}

// replCompleter implements readline.AutoCompleter
type replCompleter struct {
	commands []string
	registry *command.Registry
}

// Do implements the completion logic
func (c *replCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	logging.LogDebug("Tab completion requested", "line", string(line), "pos", pos)

	lineStr := string(line[:pos])

	// Check if this is a command (starts with / or :)
	if !strings.HasPrefix(lineStr, "/") && !strings.HasPrefix(lineStr, ":") {
		return nil, 0
	}

	// Extract the command prefix
	prefix := lineStr[1:] // Remove the / or :

	// Find matching commands
	var candidates [][]rune
	for _, cmd := range c.commands {
		if strings.HasPrefix(cmd, prefix) {
			// Add the full command with the original prefix character
			fullCmd := lineStr[0:1] + cmd
			candidates = append(candidates, []rune(fullCmd))
		}
	}

	logging.LogDebug("Found completions", "count", len(candidates), "prefix", prefix)

	if len(candidates) == 0 {
		return nil, 0
	}

	// Return completions starting from the beginning of the line
	return candidates, 0
}

// getCommandNames returns all available REPL command names
func getCommandNames() []string {
	// This will be populated from the command registry
	return []string{
		"help",
		"exit",
		"quit",
		"clear",
		"history",
		"save",
		"export",
		"multiline",
		"model",
		"config",
		"attach",
		"session",
		"context",
		"reset",
		"undo",
		"redo",
		"theme",
		"alias",
		"info",
		"stats",
		"logs",
		"system",
	}
}

// ReadlineWriter implements io.Writer for readline
type ReadlineWriter struct {
	Instance *readline.Instance
}

// Write implements io.Writer
func (w *ReadlineWriter) Write(p []byte) (n int, err error) {
	// Use readline's output to maintain prompt position
	return w.Instance.Stdout().Write(p)
}
