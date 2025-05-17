// ABOUTME: Provides the interactive REPL interface for the chat functionality
// ABOUTME: Handles user input, command parsing, and message processing

package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lexlapax/magellai/pkg/llm"
)

// ConfigInterface defines the minimal interface needed for configuration
type ConfigInterface interface {
	GetString(key string) string
	GetBool(key string) bool
	SetValue(key string, value interface{}) error
}

// REPL represents the Read-Eval-Print Loop for interactive chat
type REPL struct {
	config      ConfigInterface
	provider    llm.Provider
	session     *Session
	manager     *SessionManager
	reader      *bufio.Reader
	writer      io.Writer
	promptStyle string
	multiline   bool
	exitOnEOF   bool
}

// REPLOptions contains options for creating a new REPL
type REPLOptions struct {
	Config      ConfigInterface
	StorageDir  string
	PromptStyle string
	SessionID   string // Optional: resume existing session
	Model       string // Optional: override default model
	Writer      io.Writer
	Reader      io.Reader
}

// NewREPL creates a new REPL instance
func NewREPL(opts *REPLOptions) (*REPL, error) {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	if opts.Reader == nil {
		opts.Reader = os.Stdin
	}
	if opts.PromptStyle == "" {
		opts.PromptStyle = "> "
	}
	if opts.StorageDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		opts.StorageDir = filepath.Join(homeDir, ".config", "magellai", "sessions")
	}

	// Create session manager
	manager, err := NewSessionManager(opts.StorageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	// Load current configuration
	cfg := opts.Config

	var session *Session
	if opts.SessionID != "" {
		// Resume existing session
		session, err = manager.LoadSession(opts.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
	} else {
		// Create new session
		session = manager.NewSession("Interactive Chat")
	}

	// Override model if specified
	modelStr := cfg.GetString("model")
	if opts.Model != "" {
		modelStr = opts.Model
	}

	// Parse model string to get provider and model
	parts := strings.Split(modelStr, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid model format, expected provider/model")
	}
	providerType := parts[0]
	modelName := parts[1]

	// Create provider
	provider, err := llm.NewProvider(providerType, modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Update session with model
	session.Conversation.Model = modelStr
	session.Conversation.Provider = providerType

	return &REPL{
		config:      cfg,
		provider:    provider,
		session:     session,
		manager:     manager,
		reader:      bufio.NewReader(opts.Reader),
		writer:      opts.Writer,
		promptStyle: opts.PromptStyle,
		exitOnEOF:   true,
	}, nil
}

// Run starts the REPL loop
func (r *REPL) Run() error {
	// Print welcome message
	fmt.Fprintf(r.writer, "magellai chat - Interactive LLM chat (type /help for commands)\n")
	fmt.Fprintf(r.writer, "Model: %s\n", r.session.Conversation.Model)
	fmt.Fprintf(r.writer, "Session: %s\n\n", r.session.ID)

	// Main REPL loop
	for {
		// Print prompt
		fmt.Fprint(r.writer, r.promptStyle)

		// Read input
		input, err := r.readInput()
		if err != nil {
			if err == io.EOF && r.exitOnEOF {
				fmt.Fprintln(r.writer, "\nGoodbye!")
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		// Skip empty input
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Check for commands
		if strings.HasPrefix(input, "/") {
			if err := r.handleCommand(input); err != nil {
				fmt.Fprintf(r.writer, "Error: %v\n", err)
			}
			continue
		}

		// Check for special commands (: prefix)
		if strings.HasPrefix(input, ":") {
			if err := r.handleSpecialCommand(input); err != nil {
				fmt.Fprintf(r.writer, "Error: %v\n", err)
			}
			continue
		}

		// Process as conversation
		if err := r.processMessage(input); err != nil {
			fmt.Fprintf(r.writer, "Error: %v\n", err)
		}

		// Auto-save session
		if err := r.manager.SaveSession(r.session); err != nil {
			fmt.Fprintf(r.writer, "Warning: Failed to auto-save session: %v\n", err)
		}
	}
}

// readInput reads user input, handling multi-line mode if enabled
func (r *REPL) readInput() (string, error) {
	if !r.multiline {
		return r.reader.ReadString('\n')
	}

	// Multi-line mode
	var lines []string
	for {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line ends multi-line input
			break
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

// processMessage processes a user message and generates a response
func (r *REPL) processMessage(message string) error {
	// Get pending attachments
	var attachments []llm.Attachment
	if r.session.Metadata != nil {
		if pending, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment); ok {
			attachments = pending
			// Clear pending attachments
			delete(r.session.Metadata, "pending_attachments")
		}
	}

	// Add user message to conversation
	r.session.Conversation.AddMessage("user", message, attachments)

	// Get conversation history
	messages := r.session.Conversation.GetHistory()

	// Prepare options
	var opts []llm.ProviderOption

	// Apply model settings
	if temp := r.session.Conversation.Temperature; temp > 0 {
		opts = append(opts, llm.WithTemperature(temp))
	}
	if maxTokens := r.session.Conversation.MaxTokens; maxTokens > 0 {
		opts = append(opts, llm.WithMaxTokens(maxTokens))
	}

	// Create context
	ctx := context.Background()

	// Use streaming if enabled
	if r.config.GetBool("stream") {
		// Start response
		fmt.Fprint(r.writer, "\n")

		// Stream response chunks
		var fullResponse strings.Builder
		stream, err := r.provider.StreamMessage(ctx, messages, opts...)
		if err != nil {
			return fmt.Errorf("failed to start stream: %w", err)
		}

		for chunk := range stream {
			if chunk.Error != nil {
				return fmt.Errorf("stream error: %w", chunk.Error)
			}
			fmt.Fprint(r.writer, chunk.Content)
			fullResponse.WriteString(chunk.Content)
		}

		fmt.Fprintln(r.writer, "")

		// Add assistant message to conversation
		r.session.Conversation.AddMessage("assistant", fullResponse.String(), nil)
	} else {
		// Non-streaming response
		resp, err := r.provider.GenerateMessage(ctx, messages, opts...)
		if err != nil {
			return fmt.Errorf("failed to generate response: %w", err)
		}

		// Print response
		fmt.Fprintf(r.writer, "\n%s\n\n", resp.Content)

		// Add assistant message to conversation
		r.session.Conversation.AddMessage("assistant", resp.Content, nil)
	}

	return nil
}

// handleCommand handles REPL commands (starting with /)
func (r *REPL) handleCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "/help":
		return r.showHelp()
	case "/exit", "/quit":
		// Save session before exiting
		if err := r.manager.SaveSession(r.session); err != nil {
			fmt.Fprintf(r.writer, "Warning: Failed to save session: %v\n", err)
		}
		fmt.Fprintln(r.writer, "Goodbye!")
		os.Exit(0)
		return nil // This line is never reached but satisfies the compiler
	case "/save":
		return r.saveSession(args)
	case "/load":
		return r.loadSession(args)
	case "/reset":
		return r.resetConversation()
	case "/model":
		return r.showModel()
	case "/system":
		return r.setSystemPrompt(args)
	case "/history":
		return r.showHistory()
	case "/sessions":
		return r.listSessions()
	case "/attach":
		return r.attachFile(args)
	case "/attachments":
		return r.listAttachments()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// handleSpecialCommand handles special commands (starting with :)
func (r *REPL) handleSpecialCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case ":model":
		return r.switchModel(args)
	case ":stream":
		return r.toggleStreaming(args)
	case ":temperature":
		return r.setTemperature(args)
	case ":max_tokens":
		return r.setMaxTokens(args)
	case ":multiline":
		return r.toggleMultiline()
	default:
		return fmt.Errorf("unknown special command: %s", command)
	}
}

// Command implementations follow...
// These will be in the next file (commands.go)

// showHelp displays help information
func (r *REPL) showHelp() error {
	fmt.Fprint(r.writer, `
magellai chat - Interactive LLM chat

COMMANDS:
  /help              Show this help message
  /exit, /quit       Exit the chat session
  /save [name]       Save the current session
  /load <id>         Load a previous session
  /reset             Clear the conversation history
  /model             Show current model
  /system [prompt]   Set or show system prompt
  /history           Show conversation history
  /sessions          List all sessions
  /attach <file>     Attach a file to the next message
  /attachments       List current attachments

SPECIAL COMMANDS:
  :model <name>      Switch to a different model
  :stream on/off     Toggle streaming mode
  :temperature <n>   Set generation temperature (0.0-2.0)
  :max_tokens <n>    Set maximum response tokens
  :multiline         Toggle multi-line input mode

Type your message and press Enter to send.
`)
	return nil
}
