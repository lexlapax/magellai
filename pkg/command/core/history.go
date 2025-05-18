// ABOUTME: Implements the history command for managing and viewing REPL session history
// ABOUTME: Provides subcommands for listing, showing, deleting, exporting, and searching sessions

package core

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/lexlapax/magellai/internal/configdir"
	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/repl"
	"github.com/lexlapax/magellai/pkg/storage"
)

// HistoryCommand implements the history command
type HistoryCommand struct {
	subcommand string
	sessionID  string
	searchTerm string
	format     string
}

// NewHistoryCommand creates a new history command
func NewHistoryCommand() *HistoryCommand {
	return &HistoryCommand{}
}

func (c *HistoryCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	// Initialize data map if needed
	if exec.Data == nil {
		exec.Data = make(map[string]interface{})
	}

	if len(exec.Args) == 0 {
		return fmt.Errorf("no subcommand specified")
	}

	c.subcommand = exec.Args[0]

	// Process flags
	if format, ok := exec.Flags.Get("format").(string); ok {
		c.format = format
	} else {
		c.format = "json" // default format
	}

	// Get session storage directory
	paths, err := configdir.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get config paths: %v", err)
	}

	// Create storage manager using filesystem backend
	manager, err := repl.CreateStorageManager(storage.FileSystemBackend, storage.Config{
		"base_dir": paths.Sessions,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage manager: %v", err)
	}

	// Create session manager wrapping storage manager
	sessionManager := &repl.SessionManager{StorageManager: manager}

	switch c.subcommand {
	case "list":
		return c.executeList(ctx, exec, sessionManager)
	case "show":
		if len(exec.Args) < 2 {
			return fmt.Errorf("session ID required for show command")
		}
		c.sessionID = exec.Args[1]
		return c.executeShow(ctx, exec, sessionManager)
	case "delete":
		if len(exec.Args) < 2 {
			return fmt.Errorf("session ID required for delete command")
		}
		c.sessionID = exec.Args[1]
		return c.executeDelete(ctx, exec, sessionManager)
	case "export":
		if len(exec.Args) < 2 {
			return fmt.Errorf("session ID required for export command")
		}
		c.sessionID = exec.Args[1]
		return c.executeExport(ctx, exec, sessionManager)
	case "search":
		if len(exec.Args) < 2 {
			return fmt.Errorf("search term required for search command")
		}
		c.searchTerm = strings.Join(exec.Args[1:], " ")
		return c.executeSearch(ctx, exec, sessionManager)
	default:
		return fmt.Errorf("unknown subcommand: %s", c.subcommand)
	}
}

func (c *HistoryCommand) executeList(ctx context.Context, exec *command.ExecutionContext, manager *repl.SessionManager) error {
	logging.LogInfo("Listing sessions")

	sessions, err := manager.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Fprintln(exec.Stdout, "No sessions found")
		return nil
	}

	// Format output as table
	w := tabwriter.NewWriter(exec.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCREATED\tUPDATED\tMESSAGES")

	for _, session := range sessions {
		created := session.Created.Format("2006-01-02 15:04")
		updated := session.Updated.Format("2006-01-02 15:04")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
			session.ID,
			session.Name,
			created,
			updated,
			session.MessageCount)
	}

	w.Flush()
	exec.Data["sessions"] = sessions
	return nil
}

func (c *HistoryCommand) executeShow(ctx context.Context, exec *command.ExecutionContext, manager *repl.SessionManager) error {
	logging.LogInfo("Showing session details", "id", c.sessionID)

	session, err := manager.StorageManager.LoadSession(c.sessionID)
	if err != nil {
		return fmt.Errorf("failed to load session: %v", err)
	}

	// Format session details
	fmt.Fprintf(exec.Stdout, "Session ID: %s\n", session.ID)
	if session.Name != "" {
		fmt.Fprintf(exec.Stdout, "Name: %s\n", session.Name)
	}
	fmt.Fprintf(exec.Stdout, "Created: %s\n", session.Created.Format(time.RFC3339))
	fmt.Fprintf(exec.Stdout, "Updated: %s\n", session.Updated.Format(time.RFC3339))
	if len(session.Tags) > 0 {
		fmt.Fprintf(exec.Stdout, "Tags: %s\n", strings.Join(session.Tags, ", "))
	}
	fmt.Fprintf(exec.Stdout, "\nMessages (%d):\n", len(session.Conversation.Messages))

	for i, msg := range session.Conversation.Messages {
		timestamp := msg.Timestamp.Format("15:04:05")
		content := msg.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		fmt.Fprintf(exec.Stdout, "  [%d] %s (%s): %s\n", i+1, timestamp, msg.Role, content)
		if len(msg.Attachments) > 0 {
			fmt.Fprintf(exec.Stdout, "      Attachments: %d\n", len(msg.Attachments))
		}
	}

	exec.Data["session"] = session
	return nil
}

func (c *HistoryCommand) executeDelete(ctx context.Context, exec *command.ExecutionContext, manager *repl.SessionManager) error {
	logging.LogInfo("Deleting session", "id", c.sessionID)

	err := manager.DeleteSession(c.sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}

	fmt.Fprintf(exec.Stdout, "Session %s deleted\n", c.sessionID)
	exec.Data["deleted_id"] = c.sessionID
	return nil
}

func (c *HistoryCommand) executeExport(ctx context.Context, exec *command.ExecutionContext, manager *repl.SessionManager) error {
	logging.LogInfo("Exporting session", "id", c.sessionID, "format", c.format)

	err := manager.ExportSession(c.sessionID, c.format, exec.Stdout)
	if err != nil {
		return fmt.Errorf("failed to export session: %v", err)
	}

	exec.Data["exported_id"] = c.sessionID
	exec.Data["format"] = c.format
	return nil
}

func (c *HistoryCommand) executeSearch(ctx context.Context, exec *command.ExecutionContext, manager *repl.SessionManager) error {
	logging.LogInfo("Searching sessions", "query", c.searchTerm)

	sessions, err := manager.SearchSessions(c.searchTerm)
	if err != nil {
		return fmt.Errorf("failed to search sessions: %v", err)
	}

	if len(sessions) == 0 {
		fmt.Fprintf(exec.Stdout, "No sessions found matching '%s'\n", c.searchTerm)
		return nil
	}

	// Format output as table
	w := tabwriter.NewWriter(exec.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCREATED\tTAGS")

	for _, result := range sessions {
		created := result.Session.Created.Format("2006-01-02 15:04")
		tags := strings.Join(result.Session.Tags, ", ")
		if tags == "" {
			tags = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			result.Session.ID,
			result.Session.Name,
			created,
			tags)
	}

	w.Flush()
	exec.Data["sessions"] = sessions
	exec.Data["query"] = c.searchTerm
	return nil
}

func (c *HistoryCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "history",
		Category:    command.CategoryCLI | command.CategoryREPL,
		Description: "Manage REPL session history",
		LongDescription: `The history command allows you to manage and view REPL session history.

Subcommands:
  list    - List all sessions
  show    - Show detailed information about a specific session
  delete  - Delete a specific session
  export  - Export a session in JSON or markdown format
  search  - Search sessions by content

Examples:
  magellai history list
  magellai history show <session-id>
  magellai history delete <session-id>
  magellai history export <session-id> --format=markdown
  magellai history search "python code"`,
		Flags: []command.Flag{
			{
				Name:        "format",
				Description: "Export format (json|markdown)",
				Default:     "json",
			},
		},
	}
}

func (c *HistoryCommand) Validate() error {
	return nil
}
