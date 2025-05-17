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
	"time"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem" // Register filesystem backend
	_ "github.com/lexlapax/magellai/pkg/storage/sqlite"     // Register SQLite backend
)

// ConfigInterface defines the minimal interface needed for configuration
type ConfigInterface interface {
	GetString(key string) string
	GetBool(key string) bool
	Get(key string) interface{}
	Exists(key string) bool
	SetValue(key string, value interface{}) error
}

// REPL represents the Read-Eval-Print Loop for interactive chat
type REPL struct {
	config        ConfigInterface
	provider      llm.Provider
	session       *Session
	manager       *SessionManager
	reader        *bufio.Reader
	writer        io.Writer
	promptStyle   string
	multiline     bool
	exitOnEOF     bool
	autoSave      bool
	autoSaveTimer *time.Timer
	lastSaveTime  time.Time
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
	logging.LogDebug("Creating new REPL instance", "sessionID", opts.SessionID, "model", opts.Model)

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
			logging.LogError(err, "Failed to get home directory")
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		opts.StorageDir = filepath.Join(homeDir, ".config", "magellai", "sessions")
		logging.LogDebug("Using default session storage directory", "dir", opts.StorageDir)
	}
	// Load current configuration
	cfg := opts.Config

	// Create storage backend
	logging.LogDebug("Creating storage backend", "storageDir", opts.StorageDir)

	// Get storage configuration from config
	storageType := "filesystem" // Default to filesystem
	if cfg.Exists("session.storage.type") {
		storageType = cfg.GetString("session.storage.type")
	}

	storageConfig := make(map[string]interface{})
	storageConfig["base_dir"] = opts.StorageDir
	if cfg.Exists("session.storage.settings") {
		settings := cfg.Get("session.storage.settings")
		if m, ok := settings.(map[string]interface{}); ok {
			for k, v := range m {
				storageConfig[k] = v
			}
		}
	}

	// Create storage using the new storage package
	backend, err := CreateStorageManager(storage.BackendType(storageType), storage.Config(storageConfig))
	if err != nil {
		logging.LogError(err, "Failed to create storage backend", "type", storageType)
		return nil, fmt.Errorf("failed to create storage backend: %w", err)
	}

	// Create session manager (backend is a StorageManager, not a Backend)
	logging.LogDebug("Creating session manager")
	manager := &SessionManager{StorageManager: backend}

	var session *Session
	if opts.SessionID != "" {
		// Resume existing session
		logging.LogInfo("Resuming existing session", "sessionID", opts.SessionID)
		session, err = manager.LoadSession(opts.SessionID)
		if err != nil {
			logging.LogError(err, "Failed to load session", "sessionID", opts.SessionID)
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
	} else {
		// Create new session
		logging.LogInfo("Creating new session")
		session, err = manager.NewSession("Interactive Chat")
		if err != nil {
			logging.LogError(err, "Failed to create new session")
			return nil, fmt.Errorf("failed to create new session: %w", err)
		}
	}

	// Override model if specified
	modelStr := cfg.GetString("model")
	if opts.Model != "" {
		logging.LogDebug("Overriding model from options", "model", opts.Model)
		modelStr = opts.Model
	}

	// Parse model string to get provider and model
	logging.LogDebug("Parsing model string", "modelStr", modelStr)
	parts := strings.Split(modelStr, "/")
	if len(parts) != 2 {
		logging.LogError(nil, "Invalid model format", "modelStr", modelStr)
		return nil, fmt.Errorf("invalid model format, expected provider/model")
	}
	providerType := parts[0]
	modelName := parts[1]
	logging.LogDebug("Parsed model configuration", "provider", providerType, "model", modelName)

	// Create provider
	logging.LogInfo("Creating LLM provider", "provider", providerType, "model", modelName)
	provider, err := llm.NewProvider(providerType, modelName)
	if err != nil {
		logging.LogError(err, "Failed to create provider", "provider", providerType, "model", modelName)
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Update session with model
	session.Conversation.Model = modelStr
	session.Conversation.Provider = providerType

	autoSave := cfg.GetBool("repl.auto_save.enabled")

	repl := &REPL{
		config:       cfg,
		provider:     provider,
		session:      session,
		manager:      manager,
		reader:       bufio.NewReader(opts.Reader),
		writer:       opts.Writer,
		promptStyle:  opts.PromptStyle,
		exitOnEOF:    true,
		autoSave:     autoSave,
		lastSaveTime: time.Now(),
	}

	// Setup auto-save timer if enabled
	if autoSave {
		interval := cfg.GetString("repl.auto_save.interval")
		if interval == "" {
			interval = "5m" // Default to 5 minutes
		}
		duration, err := time.ParseDuration(interval)
		if err != nil {
			logging.LogWarn("Invalid auto-save interval, using default", "interval", interval, "error", err)
			duration = 5 * time.Minute
		}
		repl.scheduleAutoSave(duration)
		logging.LogInfo("Auto-save enabled", "interval", duration)
	}

	return repl, nil
}

// Run starts the REPL loop
func (r *REPL) Run() error {
	logging.LogInfo("Starting REPL session", "sessionID", r.session.ID, "model", r.session.Conversation.Model)

	// Print welcome message
	fmt.Fprintf(r.writer, "magellai chat - Interactive LLM chat (type /help for commands)\n")
	fmt.Fprintf(r.writer, "Model: %s\n", r.session.Conversation.Model)
	fmt.Fprintf(r.writer, "Session: %s\n\n", r.session.ID)

	// Cleanup function to stop auto-save
	defer func() {
		if r.autoSave {
			r.stopAutoSave()
			logging.LogInfo("Stopped auto-save timer")
		}
	}()

	// Main REPL loop
	for {
		// Print prompt
		fmt.Fprint(r.writer, r.promptStyle)

		// Read input
		logging.LogDebug("Reading user input")
		input, err := r.readInput()
		if err != nil {
			if err == io.EOF && r.exitOnEOF {
				logging.LogInfo("EOF received, exiting REPL")
				fmt.Fprintln(r.writer, "\nGoodbye!")
				return nil
			}
			logging.LogError(err, "Read error")
			return fmt.Errorf("read error: %w", err)
		}

		// Skip empty input
		input = strings.TrimSpace(input)
		if input == "" {
			logging.LogDebug("Empty input, skipping")
			continue
		}
		logging.LogDebug("Processing user input", "inputLength", len(input))

		// Check for commands
		if strings.HasPrefix(input, "/") {
			logging.LogDebug("Processing command", "command", input)
			if err := r.handleCommand(input); err != nil {
				logging.LogError(err, "Command error", "command", input)
				fmt.Fprintf(r.writer, "Error: %v\n", err)
			}
			continue
		}

		// Check for special commands (: prefix)
		if strings.HasPrefix(input, ":") {
			logging.LogDebug("Processing special command", "command", input)
			if err := r.handleSpecialCommand(input); err != nil {
				logging.LogError(err, "Special command error", "command", input)
				fmt.Fprintf(r.writer, "Error: %v\n", err)
			}
			continue
		}

		// Process as conversation
		logging.LogDebug("Processing message", "messageLength", len(input))
		if err := r.processMessage(input); err != nil {
			logging.LogError(err, "Message processing error")
			fmt.Fprintf(r.writer, "Error: %v\n", err)
		}

		// Trigger auto-save after processing message
		if r.autoSave {
			if err := r.performAutoSave(); err != nil {
				logging.LogWarn("Failed to auto-save session", "sessionID", r.session.ID, "error", err)
			}
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
	logging.LogDebug("Processing message", "message", message)
	// Get pending attachments
	var attachments []llm.Attachment
	if r.session.Metadata != nil {
		if pending, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment); ok {
			attachments = pending
			logging.LogDebug("Found pending attachments", "count", len(attachments))
			// Clear pending attachments
			delete(r.session.Metadata, "pending_attachments")
		}
	}

	// Add user message to conversation
	logging.LogDebug("Adding user message to conversation", "attachmentCount", len(attachments))
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
		logging.LogDebug("Using streaming mode")
		// Start response
		fmt.Fprint(r.writer, "\n")

		// Stream response chunks
		var fullResponse strings.Builder
		stream, err := r.provider.StreamMessage(ctx, messages, opts...)
		if err != nil {
			logging.LogError(err, "Failed to start stream")
			return fmt.Errorf("failed to start stream: %w", err)
		}

		for chunk := range stream {
			if chunk.Error != nil {
				logging.LogError(chunk.Error, "Stream error")
				return fmt.Errorf("stream error: %w", chunk.Error)
			}
			fmt.Fprint(r.writer, chunk.Content)
			fullResponse.WriteString(chunk.Content)
		}
		logging.LogDebug("Stream completed", "responseLength", fullResponse.Len())

		fmt.Fprintln(r.writer, "")

		// Add assistant message to conversation
		r.session.Conversation.AddMessage("assistant", fullResponse.String(), nil)
	} else {
		logging.LogDebug("Using non-streaming mode")
		// Non-streaming response
		resp, err := r.provider.GenerateMessage(ctx, messages, opts...)
		if err != nil {
			logging.LogError(err, "Failed to generate message")
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
	logging.LogDebug("Handling command", "cmd", cmd)

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]
	logging.LogDebug("Parsed command", "command", command, "argCount", len(args))

	switch command {
	case "/help":
		logging.LogDebug("Showing help")
		return r.showHelp()
	case "/exit", "/quit":
		logging.LogInfo("User requested exit")
		// Save session before exiting
		if err := r.manager.SaveSession(r.session); err != nil {
			logging.LogWarn("Failed to save session on exit", "error", err)
			fmt.Fprintf(r.writer, "Warning: Failed to save session: %v\n", err)
		}
		fmt.Fprintln(r.writer, "Goodbye!")
		os.Exit(0)
		return nil // This line is never reached but satisfies the compiler
	case "/save":
		logging.LogDebug("Saving session", "args", args)
		return r.saveSession(args)
	case "/load":
		logging.LogDebug("Loading session", "args", args)
		return r.loadSession(args)
	case "/reset":
		logging.LogDebug("Resetting conversation")
		return r.resetConversation()
	case "/model":
		logging.LogDebug("Showing model")
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
	case "/config":
		if len(args) == 0 {
			return fmt.Errorf("usage: /config show|set <key> <value>")
		}
		subcommand := args[0]
		switch subcommand {
		case "show":
			return r.showConfig()
		case "set":
			if len(args) < 2 {
				return fmt.Errorf("usage: /config set <key> <value>")
			}
			return r.setConfig(args[1:])
		default:
			return fmt.Errorf("unknown config subcommand: %s", subcommand)
		}
	case "/export":
		return r.exportSession(args)
	case "/search":
		return r.searchSessions(args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// handleSpecialCommand handles special commands (starting with :)
func (r *REPL) handleSpecialCommand(cmd string) error {
	logging.LogDebug("Handling special command", "cmd", cmd)

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]
	logging.LogDebug("Parsed special command", "command", command, "argCount", len(args))

	switch command {
	case ":model":
		logging.LogDebug("Switching model", "args", args)
		return r.switchModel(args)
	case ":stream":
		logging.LogDebug("Toggling streaming", "args", args)
		return r.toggleStreaming(args)
	case ":temperature":
		logging.LogDebug("Setting temperature", "args", args)
		return r.setTemperature(args)
	case ":max_tokens":
		logging.LogDebug("Setting max tokens", "args", args)
		return r.setMaxTokens(args)
	case ":multiline":
		logging.LogDebug("Toggling multiline")
		return r.toggleMultiline()
	case ":verbosity":
		logging.LogDebug("Setting verbosity", "args", args)
		return r.setVerbosity(args)
	case ":output":
		logging.LogDebug("Setting output format", "args", args)
		return r.setOutput(args)
	case ":profile":
		logging.LogDebug("Switching profile", "args", args)
		return r.switchProfile(args)
	case ":attach":
		logging.LogDebug("Attaching file", "args", args)
		return r.attachFile(args)
	case ":attach-remove":
		logging.LogDebug("Removing attachment", "args", args)
		return r.removeAttachment(args)
	case ":attach-list":
		logging.LogDebug("Listing attachments")
		return r.listAttachments()
	case ":system":
		logging.LogDebug("Setting system prompt", "args", args)
		return r.setSystemPrompt(args)
	default:
		logging.LogDebug("Unknown special command", "command", command)
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
  /search <query>    Search sessions by content
  /attach <file>     Attach a file to the next message
  /attachments       List current attachments
  /config show       Display current configuration
  /config set <k> <v> Set configuration value
  /export <fmt> [f]  Export session (json, markdown)

SPECIAL COMMANDS:
  :model <name>         Switch to a different model
  :stream on/off     Toggle streaming mode
  :verbosity <level> Set verbosity (debug, info, warn, error)
  :output <format>   Set output format (text, json, yaml, markdown)
  :temperature <n>   Set generation temperature (0.0-2.0)
  :max_tokens <n>    Set maximum response tokens
  :profile <n>       Switch to a different profile
  :attach <file>     Add attachment for next message
  :attach-remove <f> Remove a pending attachment
  :attach-list       List all pending attachments
  :system [prompt]   Set or show system prompt
  :multiline         Toggle multi-line input mode

Type your message and press Enter to send.
`)
	return nil
}

// scheduleAutoSave sets up the auto-save timer
func (r *REPL) scheduleAutoSave(interval time.Duration) {
	if r.autoSaveTimer != nil {
		r.autoSaveTimer.Stop()
	}

	logging.LogDebug("Scheduling auto-save", "interval", interval)
	r.autoSaveTimer = time.AfterFunc(interval, func() {
		logging.LogDebug("Auto-save timer triggered")
		if err := r.performAutoSave(); err != nil {
			logging.LogError(err, "Auto-save failed")
		}
		// Reschedule the timer
		r.scheduleAutoSave(interval)
	})
}

// performAutoSave saves the current session
func (r *REPL) performAutoSave() error {
	// Don't save if no changes since last save
	if r.session.Updated.Before(r.lastSaveTime) || r.session.Updated.Equal(r.lastSaveTime) {
		logging.LogDebug("No changes since last save, skipping auto-save")
		return nil
	}

	logging.LogInfo("Performing auto-save", "sessionID", r.session.ID)
	if err := r.manager.SaveSession(r.session); err != nil {
		return fmt.Errorf("auto-save failed: %w", err)
	}

	r.lastSaveTime = time.Now()
	logging.LogInfo("Auto-save completed", "sessionID", r.session.ID)
	return nil
}

// stopAutoSave stops the auto-save timer
func (r *REPL) stopAutoSave() {
	if r.autoSaveTimer != nil {
		r.autoSaveTimer.Stop()
		r.autoSaveTimer = nil
		logging.LogDebug("Auto-save timer stopped")
	}
}
