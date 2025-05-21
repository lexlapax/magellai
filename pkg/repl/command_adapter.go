// ABOUTME: Adapter to integrate REPL commands with the unified command system
// ABOUTME: Bridges REPL functionality with command registry for consistent command handling

package repl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
)

// REPLCommandAdapter adapts REPL functions to the unified command interface
type REPLCommandAdapter struct {
	repl     *REPL
	metadata *command.Metadata
	handler  func(*REPL, []string) error
}

// NewREPLCommandAdapter creates a new adapter for a REPL command
func NewREPLCommandAdapter(repl *REPL, meta *command.Metadata, handler func(*REPL, []string) error) *REPLCommandAdapter {
	return &REPLCommandAdapter{
		repl:     repl,
		metadata: meta,
		handler:  handler,
	}
}

// Execute implements command.Interface
func (a *REPLCommandAdapter) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	return a.handler(a.repl, exec.Args)
}

// Metadata implements command.Interface
func (a *REPLCommandAdapter) Metadata() *command.Metadata {
	return a.metadata
}

// Validate implements command.Interface
func (a *REPLCommandAdapter) Validate() error {
	if a.metadata == nil || a.metadata.Name == "" {
		return fmt.Errorf("invalid command metadata")
	}
	return nil
}

// RegisterREPLCommands registers all REPL commands in the unified registry
func RegisterREPLCommands(repl *REPL, registry *command.Registry) error {
	// Define all REPL commands with their metadata
	commands := []struct {
		meta    *command.Metadata
		handler func(*REPL, []string) error
	}{
		// Slash commands
		{
			meta: &command.Metadata{
				Name:        "help",
				Aliases:     []string{"h", "?"},
				Description: "Show help message",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.showHelp()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "exit",
				Aliases:     []string{"quit", "q"},
				Description: "Exit the chat session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				// Save session before exiting
				if err := r.manager.SaveSession(r.session); err != nil {
					fmt.Fprintf(r.writer, "Warning: Failed to save session: %v\n", err)
				}
				fmt.Fprintln(r.writer, "Goodbye!")
				return io.EOF // Signal to exit
			},
		},
		{
			meta: &command.Metadata{
				Name:        "save",
				Description: "Save the current session",
				Category:    command.CategoryREPL,
				Flags: []command.Flag{
					{
						Name:        "name",
						Description: "Session name",
						Type:        command.FlagTypeString,
					},
				},
			},
			handler: func(r *REPL, args []string) error {
				return r.saveSession(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "load",
				Description: "Load a previous session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.loadSession(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "reset",
				Description: "Clear the conversation history",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.resetConversation()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "model",
				Description: "Show current model",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.showModel()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "system",
				Description: "Set or show system prompt",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.setSystemPrompt(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "history",
				Description: "Show conversation history",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.showHistory()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "sessions",
				Description: "List all sessions",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.listSessions()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "attach",
				Description: "Attach a file to the next message",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.attachFile(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "attachments",
				Description: "List current attachments",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.listAttachments()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "config",
				Description: "Show or set configuration",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
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
			},
		},
		{
			meta: &command.Metadata{
				Name:        "export",
				Description: "Export session to file",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.exportSession(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "search",
				Description: "Search sessions by content",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.searchSessions(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "tags",
				Description: "List tags for current session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.listTags()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "tag",
				Description: "Add tag to current session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.addTag(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "untag",
				Description: "Remove tag from current session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.removeTag(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "metadata",
				Description: "Show session metadata",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.showMetadata()
			},
		},
		{
			meta: &command.Metadata{
				Name:        "meta",
				Description: "Set or delete metadata",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				if len(args) == 0 {
					return fmt.Errorf("usage: /meta set <key> <value> or /meta del <key>")
				}
				subcommand := args[0]
				switch subcommand {
				case "set":
					if len(args) < 3 {
						return fmt.Errorf("usage: /meta set <key> <value>")
					}
					return r.setMetadata(args[1], strings.Join(args[2:], " "))
				case "del":
					if len(args) < 2 {
						return fmt.Errorf("usage: /meta del <key>")
					}
					return r.deleteMetadata(args[1])
				default:
					return fmt.Errorf("unknown meta subcommand: %s", subcommand)
				}
			},
		},
		// Colon commands (registered with : prefix)
		{
			meta: &command.Metadata{
				Name:        ":model",
				Description: "Switch to a different model",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.switchModel(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":stream",
				Aliases:     []string{":streaming"},
				Description: "Toggle streaming mode",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.toggleStreaming(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":temperature",
				Aliases:     []string{":temp"},
				Description: "Set temperature parameter",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.setTemperature(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":max_tokens",
				Aliases:     []string{":tokens"},
				Description: "Set maximum tokens",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.setMaxTokens(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":multiline",
				Aliases:     []string{":ml"},
				Description: "Toggle multiline input mode",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.toggleMultiline()
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":verbosity",
				Aliases:     []string{":v"},
				Description: "Set logging verbosity",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.setVerbosity(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":output",
				Description: "Set output format",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.setOutputFormat(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":profile",
				Description: "Switch configuration profile",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.switchProfile(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":attach",
				Description: "Attach file using colon command",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.attachFile(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":attach-remove",
				Description: "Remove attachment",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.removeAttachment(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":attach-list",
				Description: "List attachments",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.listAttachments()
			},
		},
		{
			meta: &command.Metadata{
				Name:        ":system",
				Description: "Set system prompt using colon command",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.setSystemPrompt(args)
			},
		},
		// Branch commands
		{
			meta: &command.Metadata{
				Name:        "branch",
				Description: "Create a new branch from current session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.cmdBranch(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "branches",
				Description: "List all branches of current session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.cmdBranches(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "tree",
				Description: "Show session branch tree",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.cmdTree(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "switch",
				Description: "Switch to a different branch",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.cmdSwitch(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "merge",
				Description: "Merge another session into current session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.cmdMerge(args)
			},
		},
		{
			meta: &command.Metadata{
				Name:        "recover",
				Description: "Recover from crashed session",
				Category:    command.CategoryREPL,
			},
			handler: func(r *REPL, args []string) error {
				return r.cmdRecover(args)
			},
		},
	}

	// Register all commands
	for _, cmd := range commands {
		adapter := NewREPLCommandAdapter(repl, cmd.meta, cmd.handler)
		if err := registry.Register(adapter); err != nil {
			// Log the error but continue with other commands
			logging.LogError(err, "Command already registered", "name", cmd.meta.Name)
		}
	}

	return nil
}

// CreateCommandContext creates an ExecutionContext for a command
func CreateCommandContext(args []string, stdin io.Reader, stdout, stderr io.Writer) *command.ExecutionContext {
	return &command.ExecutionContext{
		Args:   args,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Data:   make(map[string]interface{}),
	}
}

// CreateCommandContextWithShared creates an ExecutionContext with shared context
func CreateCommandContextWithShared(args []string, stdin io.Reader, stdout, stderr io.Writer, shared *command.SharedContext) *command.ExecutionContext {
	return &command.ExecutionContext{
		Args:          args,
		Stdin:         stdin,
		Stdout:        stdout,
		Stderr:        stderr,
		Data:          make(map[string]interface{}),
		SharedContext: shared,
	}
}
