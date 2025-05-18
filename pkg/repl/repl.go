// ABOUTME: Provides the interactive REPL interface for the chat functionality
// ABOUTME: Handles user input, command parsing, and message processing

package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem" // Register filesystem backend
	_ "github.com/lexlapax/magellai/pkg/storage/sqlite"     // Register SQLite backend
	"github.com/lexlapax/magellai/pkg/utils"
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
	config         ConfigInterface
	provider       llm.Provider
	session        *Session
	manager        *SessionManager
	reader         *bufio.Reader
	writer         io.Writer
	promptStyle    string
	multiline      bool
	exitOnEOF      bool
	autoSave       bool
	autoSaveTimer  *time.Timer
	lastSaveTime   time.Time
	autoRecovery   *AutoRecoveryManager
	registry       *command.Registry
	cmdHistory     []string              // Command history
	readline       *ReadlineInterface    // Readline interface for tab completion
	isTerminal     bool                  // Whether we're running in a terminal
	colorFormatter *utils.ColorFormatter // Color formatter for output
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

	// Check for crash recovery first if no specific session is requested
	if opts.SessionID == "" {
		// Create auto-recovery manager to check for recoverable sessions
		tempAutoRecovery, err := NewAutoRecoveryManager(DefaultAutoRecoveryConfig(), backend)
		if err == nil {
			recoveryState, err := tempAutoRecovery.CheckRecovery()
			if err == nil && recoveryState != nil {
				fmt.Fprintf(opts.Writer, "Found recoverable session from previous crash.\n")
				fmt.Fprintf(opts.Writer, "Session ID: %s\n", recoveryState.SessionID)
				fmt.Fprintf(opts.Writer, "Session Name: %s\n", recoveryState.SessionName)
				fmt.Fprintf(opts.Writer, "Last saved: %s\n", recoveryState.Timestamp.Format("2006-01-02 15:04:05"))
				fmt.Fprint(opts.Writer, "Recover this session? (y/n): ")

				reader := bufio.NewReader(opts.Reader)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))

				if response == "y" || response == "yes" {
					session, err = tempAutoRecovery.RecoverSession(recoveryState)
					if err != nil {
						logging.LogWarn("Failed to recover session", "error", err)
					} else {
						logging.LogInfo("Recovered session from crash", "id", session.ID)
						fmt.Fprintf(opts.Writer, "Session recovered successfully.\n\n")
						// Clear the recovery state since we've recovered
						if err := tempAutoRecovery.ClearRecoveryState(); err != nil {
							logging.LogWarn("Failed to clear recovery state after recovery", "error", err)
						}
					}
				} else {
					// User declined recovery, clear the state
					if err := tempAutoRecovery.ClearRecoveryState(); err != nil {
						logging.LogWarn("Failed to clear recovery state after decline", "error", err)
					}
				}
			}
		}
	}

	if opts.SessionID != "" {
		// Resume existing session
		logging.LogInfo("Resuming existing session", "sessionID", opts.SessionID)
		session, err = manager.StorageManager.LoadSession(opts.SessionID)
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
		registry:     command.NewRegistry(),
		cmdHistory:   make([]string, 0),
		isTerminal:   utils.IsTerminal(),
	}

	// Initialize color formatter if in terminal
	enableColors := repl.isTerminal && cfg.GetBool("repl.colors.enabled")
	repl.colorFormatter = utils.NewColorFormatter(enableColors, nil)

	// Register all REPL commands
	if err := RegisterREPLCommands(repl, repl.registry); err != nil {
		logging.LogError(err, "Failed to register REPL commands")
		return nil, fmt.Errorf("failed to register REPL commands: %w", err)
	}

	// Initialize readline if in terminal mode
	if repl.isTerminal {
		// Get command names from registry for tab completion
		commands := make([]string, 0)
		for _, cmd := range repl.registry.List(command.CategoryShared) {
			// Add the actual command name (without prefix)
			meta := cmd.Metadata()
			commands = append(commands, meta.Name)
			// Add aliases if any
			commands = append(commands, meta.Aliases...)
		}

		historyFile := ""
		if opts.StorageDir != "" {
			historyFile = filepath.Join(opts.StorageDir, ".repl_history")
		}

		// Use colored prompt if colors are enabled
		prompt := repl.promptStyle
		if repl.colorFormatter.Enabled() {
			prompt = repl.colorFormatter.FormatPrompt(prompt)
		}

		readlineConfig := &ReadlineConfig{
			Prompt:           prompt,
			HistoryFile:      historyFile,
			EnableCompletion: true,
			MultilineMode:    repl.multiline,
		}

		readlineInterface, err := NewReadlineInterface(readlineConfig)
		if err != nil {
			logging.LogWarn("Failed to initialize readline, falling back to standard input", "error", err)
			repl.isTerminal = false
		} else {
			repl.readline = readlineInterface
			// Update completer with actual command names
			if completer, ok := repl.readline.Instance.Config.AutoComplete.(*replCompleter); ok {
				completer.commands = commands
			}
		}
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

	// Initialize auto-recovery
	autoRecoveryConfig := DefaultAutoRecoveryConfig()
	if cfg.Exists("session.auto_recovery") {
		// Allow configuration override
		if cfg.Exists("session.auto_recovery.enabled") {
			autoRecoveryConfig.Enabled = cfg.GetBool("session.auto_recovery.enabled")
		}
		if cfg.Exists("session.auto_recovery.interval") {
			interval := cfg.GetString("session.auto_recovery.interval")
			if duration, err := time.ParseDuration(interval); err == nil {
				autoRecoveryConfig.SaveInterval = duration
			}
		}
		if cfg.Exists("session.auto_recovery.max_age") {
			age := cfg.GetString("session.auto_recovery.max_age")
			if duration, err := time.ParseDuration(age); err == nil {
				autoRecoveryConfig.MaxRecoveryAge = duration
			}
		}
	}

	autoRecovery, err := NewAutoRecoveryManager(autoRecoveryConfig, manager.StorageManager)
	if err != nil {
		logging.LogWarn("Failed to create auto-recovery manager", "error", err)
		// Continue without auto-recovery
	} else {
		repl.autoRecovery = autoRecovery
		logging.LogInfo("Auto-recovery initialized", "enabled", autoRecoveryConfig.Enabled)
	}

	return repl, nil
}

// Run starts the REPL loop
func (r *REPL) Run() error {
	logging.LogInfo("Starting REPL session", "sessionID", r.session.ID, "model", r.session.Conversation.Model)

	// Print welcome message
	if r.colorFormatter.Enabled() {
		fmt.Fprintf(r.writer, "%s (type %s for commands)\n",
			r.colorFormatter.FormatInfo("magellai chat - Interactive LLM chat"),
			r.colorFormatter.FormatCommand("/help"))
		fmt.Fprintf(r.writer, "%s: %s\n",
			r.colorFormatter.FormatInfo("Model"),
			r.colorFormatter.FormatHighlight(r.session.Conversation.Model))
		fmt.Fprintf(r.writer, "%s: %s\n\n",
			r.colorFormatter.FormatInfo("Session"),
			r.colorFormatter.FormatHighlight(r.session.ID))
	} else {
		fmt.Fprintf(r.writer, "magellai chat - Interactive LLM chat (type /help for commands)\n")
		fmt.Fprintf(r.writer, "Model: %s\n", r.session.Conversation.Model)
		fmt.Fprintf(r.writer, "Session: %s\n\n", r.session.ID)
	}

	// Start auto-recovery if available
	if r.autoRecovery != nil {
		if err := r.autoRecovery.Start(); err != nil {
			logging.LogWarn("Failed to start auto-recovery", "error", err)
		}
	}

	// Setup signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logging.LogInfo("Received signal, saving recovery state", "signal", sig)

		// Force save recovery state
		if r.autoRecovery != nil {
			if err := r.autoRecovery.ForceRecoverySave(); err != nil {
				logging.LogError(err, "Failed to save recovery state on signal")
			} else {
				logging.LogInfo("Recovery state saved successfully")
			}
		}

		// Exit
		os.Exit(0)
	}()

	// Cleanup function to stop auto-save and auto-recovery
	defer func() {
		if r.autoSave {
			r.stopAutoSave()
			logging.LogInfo("Stopped auto-save timer")
		}
		if r.autoRecovery != nil {
			r.autoRecovery.Stop()
			// Save recovery state one final time
			if err := r.autoRecovery.SaveRecoveryState(); err != nil {
				logging.LogWarn("Failed to save final recovery state", "error", err)
			}
			logging.LogInfo("Stopped auto-recovery")
		}
		if r.readline != nil {
			r.readline.Close()
			logging.LogInfo("Closed readline interface")
		}
	}()

	// Main REPL loop
	for {
		// Read input
		logging.LogDebug("Reading user input")
		var input string
		var err error

		if r.readline != nil {
			// Use readline for input
			input, err = r.readline.ReadLine()
		} else {
			// Fallback to standard input
			prompt := r.promptStyle
			if r.colorFormatter.Enabled() {
				prompt = r.colorFormatter.FormatPrompt(prompt)
			}
			fmt.Fprint(r.writer, prompt)
			input, err = r.readInput()
		}

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
				if r.colorFormatter.Enabled() {
					fmt.Fprintf(r.writer, "%s: %v\n", r.colorFormatter.FormatError("Error"), err)
				} else {
					fmt.Fprintf(r.writer, "Error: %v\n", err)
				}
			}
			continue
		}

		// Check for special commands (: prefix)
		if strings.HasPrefix(input, ":") {
			logging.LogDebug("Processing special command", "command", input)
			if err := r.handleSpecialCommand(input); err != nil {
				logging.LogError(err, "Special command error", "command", input)
				if r.colorFormatter.Enabled() {
					fmt.Fprintf(r.writer, "%s: %v\n", r.colorFormatter.FormatError("Error"), err)
				} else {
					fmt.Fprintf(r.writer, "Error: %v\n", err)
				}
			}
			continue
		}

		// Process as conversation
		logging.LogDebug("Processing message", "messageLength", len(input))
		if err := r.processMessage(input); err != nil {
			logging.LogError(err, "Message processing error")
			if r.colorFormatter.Enabled() {
				fmt.Fprintf(r.writer, "%s: %v\n", r.colorFormatter.FormatError("Error"), err)
			} else {
				fmt.Fprintf(r.writer, "Error: %v\n", err)
			}
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
	AddMessageToConversation(r.session.Conversation, "user", message, attachments)

	// Save recovery state after user message
	if r.autoRecovery != nil {
		go func() {
			if err := r.autoRecovery.SaveRecoveryState(); err != nil {
				logging.LogWarn("Failed to save recovery state after user message", "error", err)
			}
		}()
	}

	// Get conversation history
	messages := GetHistory(r.session.Conversation)

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
			content := chunk.Content
			if r.colorFormatter.Enabled() {
				content = r.colorFormatter.FormatAssistantMessage(content)
			}
			fmt.Fprint(r.writer, content)
			fullResponse.WriteString(chunk.Content)
		}
		logging.LogDebug("Stream completed", "responseLength", fullResponse.Len())

		fmt.Fprintln(r.writer, "")

		// Add assistant message to conversation
		AddMessageToConversation(r.session.Conversation, "assistant", fullResponse.String(), nil)

		// Trigger recovery save after message
		if r.autoRecovery != nil {
			go func() {
				if err := r.autoRecovery.SaveRecoveryState(); err != nil {
					logging.LogWarn("Failed to save recovery state after message", "error", err)
				}
			}()
		}
	} else {
		logging.LogDebug("Using non-streaming mode")
		// Non-streaming response
		resp, err := r.provider.GenerateMessage(ctx, messages, opts...)
		if err != nil {
			logging.LogError(err, "Failed to generate message")
			return fmt.Errorf("failed to generate response: %w", err)
		}

		// Print response
		content := resp.Content
		if r.colorFormatter.Enabled() {
			content = r.colorFormatter.FormatAssistantMessage(content)
		}
		fmt.Fprintf(r.writer, "\n%s\n\n", content)

		// Add assistant message to conversation
		AddMessageToConversation(r.session.Conversation, "assistant", resp.Content, nil)

		// Trigger recovery save after message
		if r.autoRecovery != nil {
			go func() {
				if err := r.autoRecovery.SaveRecoveryState(); err != nil {
					logging.LogWarn("Failed to save recovery state after message", "error", err)
				}
			}()
		}
	}

	return nil
}

// handleCommand handles REPL commands (starting with /)
func (r *REPL) handleCommand(cmd string) error {
	logging.LogDebug("Handling command", "cmd", cmd)

	// Add to command history
	r.cmdHistory = append(r.cmdHistory, cmd)

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	commandName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]
	logging.LogDebug("Parsed command", "command", commandName, "argCount", len(args))

	// Look up command in registry
	cmdInterface, err := r.registry.Get(commandName)
	if err != nil {
		// Command not found in registry, check legacy commands
		commandName = "/" + commandName
		return r.handleLegacyCommand(commandName, args)
	}

	// Create execution context
	execCtx := CreateCommandContext(args, r.reader, r.writer, r.writer)
	execCtx.Config = r.config

	// Execute the command
	ctx := context.Background()
	if err := cmdInterface.Execute(ctx, execCtx); err != nil {
		// Handle special exit case
		if errors.Is(err, io.EOF) {
			os.Exit(0)
		}
		return err
	}

	return nil
}

// handleLegacyCommand handles commands not yet migrated to the registry
func (r *REPL) handleLegacyCommand(command string, args []string) error {
	logging.LogDebug("Handling legacy command", "command", command, "argCount", len(args))

	// Most commands should be handled by the registry
	// This should return unknown command error
	return fmt.Errorf("unknown command: %s", command)
}

// handleSpecialCommand handles special commands (starting with :)
func (r *REPL) handleSpecialCommand(cmd string) error {
	logging.LogDebug("Handling special command", "cmd", cmd)

	// Add to command history
	r.cmdHistory = append(r.cmdHistory, cmd)

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	commandName := parts[0] // Keep the : prefix for special commands
	args := parts[1:]
	logging.LogDebug("Parsed special command", "command", commandName, "argCount", len(args))

	// Look up command in registry
	cmdInterface, err := r.registry.Get(commandName)
	if err != nil {
		// Command not found in registry
		return fmt.Errorf("unknown special command: %s", commandName)
	}

	// Create execution context
	execCtx := CreateCommandContext(args, r.reader, r.writer, r.writer)
	execCtx.Config = r.config

	// Execute the command
	ctx := context.Background()
	return cmdInterface.Execute(ctx, execCtx)
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
  /tags              List tags for current session
  /tag <tag>         Add a tag to current session
  /untag <tag>       Remove a tag from current session
  /metadata          Show session metadata
  /meta set <k> <v>  Set metadata value
  /meta del <key>    Delete metadata key
  /branch <name> [at <n>]  Create a new branch at message n
  /branches          List all branches of current session
  /tree              Show session branch tree
  /switch <id>       Switch to a different branch
  /merge <source_id> Merge another session into current

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
