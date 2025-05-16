// ABOUTME: Context-aware help command and formatter implementation
// ABOUTME: Provides unified help system for CLI and REPL interfaces

package core

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
)

// HelpFormatter formats help text for commands
type HelpFormatter interface {
	// FormatCommand formats help for a single command
	FormatCommand(cmd command.Interface) string

	// FormatCommandList formats a list of commands
	FormatCommandList(commands []command.Interface, category command.Category) string

	// FormatError formats an error message with help
	FormatError(err error, suggestion string) string
}

// HelpCommand provides context-aware help functionality
type HelpCommand struct {
	registry  *command.Registry
	formatter HelpFormatter
	category  command.Category
	config    *config.Config
}

// NewHelpCommand creates a new help command
func NewHelpCommand(registry *command.Registry, config *config.Config) *HelpCommand {
	return &HelpCommand{
		registry:  registry,
		formatter: NewContextAwareHelpFormatter(config),
		category:  command.CategoryShared,
		config:    config,
	}
}

// Execute displays help information
func (h *HelpCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	// Check if we need to show all commands
	showAll := exec.Flags.GetBool("all")

	// Check if we need to show aliases
	showAliases := !exec.Flags.GetBool("no-aliases")

	// Update formatter settings
	if formatter, ok := h.formatter.(*ContextAwareHelpFormatter); ok {
		formatter.ShowHidden = showAll
		formatter.ShowAliases = showAliases
		formatter.Category = h.category
	}

	// Check if specific command help is requested
	if len(exec.Args) > 0 {
		cmdName := exec.Args[0]

		// Try exact match
		cmd, err := h.registry.Get(cmdName)
		if err != nil {
			// Try alias resolution
			aliasKey := fmt.Sprintf("aliases.%s", cmdName)
			if aliasedCmd := h.config.GetString(aliasKey); aliasedCmd != "" {
				// Extract the command from the alias
				parts := strings.Fields(aliasedCmd)
				if len(parts) > 0 {
					cmd, err = h.registry.Get(parts[0])
				}
			}

			if err != nil {
				// Try to find similar commands
				similar := h.registry.Search(cmdName)
				if len(similar) > 0 {
					suggestion := similar[0].Metadata().Name
					fmt.Fprintln(exec.Stderr, h.formatter.FormatError(err, suggestion))
				} else {
					fmt.Fprintln(exec.Stderr, h.formatter.FormatError(err, ""))
				}
				return err
			}
		}

		// Show specific command help
		fmt.Fprintln(exec.Stdout, h.formatter.FormatCommand(cmd))
		return nil
	}

	// Show general help
	commands := h.registry.List(h.category)
	fmt.Fprintln(exec.Stdout, h.formatter.FormatCommandList(commands, h.category))

	// Show aliases if in REPL mode
	if h.category == command.CategoryREPL && showAliases {
		aliasesMap := h.config.Get("aliases")
		if aliasesMap != nil {
			if aliases, ok := aliasesMap.(map[string]interface{}); ok {
				if len(aliases) > 0 {
					fmt.Fprintln(exec.Stdout, h.formatAliases(aliases))
				}
			}
		}
	}

	return nil
}

// formatAliases formats the aliases for display
func (h *HelpCommand) formatAliases(aliases map[string]interface{}) string {
	var b strings.Builder
	b.WriteString("\nAliases:\n")

	// Sort aliases for consistent display
	var names []string
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	for _, name := range names {
		value := ""
		if v, ok := aliases[name]; ok {
			value = fmt.Sprintf("%v", v)
		}
		fmt.Fprintf(w, "  %s\t→ %s\n", name, value)
	}
	w.Flush()

	return b.String()
}

// Metadata returns the command metadata
func (h *HelpCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "help",
		Aliases:     []string{"h", "?"},
		Description: "Display context-aware help information",
		Category:    command.CategoryShared,
		Flags: []command.Flag{
			{
				Name:        "all",
				Short:       "a",
				Description: "Show all commands including hidden ones",
				Type:        command.FlagTypeBool,
				Default:     false,
			},
			{
				Name:        "no-aliases",
				Short:       "n",
				Description: "Hide command aliases",
				Type:        command.FlagTypeBool,
				Default:     false,
			},
		},
	}
}

// Validate validates the command
func (h *HelpCommand) Validate() error {
	if h.registry == nil {
		return fmt.Errorf("help command requires a registry")
	}
	if h.formatter == nil {
		return fmt.Errorf("help command requires a formatter")
	}
	if h.config == nil {
		return fmt.Errorf("help command requires a config")
	}
	return nil
}

// ContextAwareHelpFormatter provides context-aware help formatting
type ContextAwareHelpFormatter struct {
	ShowHidden   bool
	ShowAliases  bool
	MaxWidth     int
	IndentSpaces int
	Category     command.Category
	config       *config.Config
}

// NewContextAwareHelpFormatter creates a new context-aware help formatter
func NewContextAwareHelpFormatter(config *config.Config) *ContextAwareHelpFormatter {
	return &ContextAwareHelpFormatter{
		ShowHidden:   false,
		ShowAliases:  true,
		MaxWidth:     80,
		IndentSpaces: 2,
		config:       config,
	}
}

// FormatCommand formats help for a single command
func (f *ContextAwareHelpFormatter) FormatCommand(cmd command.Interface) string {
	meta := cmd.Metadata()
	var b strings.Builder

	// Command name and aliases
	b.WriteString(fmt.Sprintf("Command: %s", meta.Name))
	if f.ShowAliases && len(meta.Aliases) > 0 {
		b.WriteString(fmt.Sprintf(" (aliases: %s)", strings.Join(meta.Aliases, ", ")))
	}
	b.WriteString("\n\n")

	// Description
	if meta.Description != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n\n", meta.Description))
	}

	// Long description
	if meta.LongDescription != "" {
		b.WriteString("Details:\n")
		b.WriteString(f.wrapText(meta.LongDescription, f.IndentSpaces))
		b.WriteString("\n\n")
	}

	// Flags
	if len(meta.Flags) > 0 {
		b.WriteString("Flags:\n")
		b.WriteString(f.formatFlags(meta.Flags))
		b.WriteString("\n")
	}

	// Category
	b.WriteString(fmt.Sprintf("Available in: %s\n", f.formatCategory(meta.Category)))

	// Context-specific information
	if f.Category == command.CategoryCLI && meta.Category == command.CategoryShared {
		b.WriteString("\nCLI-specific usage:\n")
		b.WriteString(fmt.Sprintf("  magellai %s [flags] [args]\n", meta.Name))
	} else if f.Category == command.CategoryREPL && meta.Category == command.CategoryShared {
		b.WriteString("\nREPL-specific usage:\n")
		b.WriteString(fmt.Sprintf("  /%s [flags] [args]\n", meta.Name))
	}

	// Deprecated warning
	if meta.Deprecated != "" {
		b.WriteString(fmt.Sprintf("\nDEPRECATED: %s\n", meta.Deprecated))
	}

	return b.String()
}

// FormatCommandList formats a list of commands
func (f *ContextAwareHelpFormatter) FormatCommandList(commands []command.Interface, category command.Category) string {
	var b strings.Builder

	// Group commands by category
	grouped := f.groupByCategory(commands)

	// Header
	title := "Available Commands"
	if category == command.CategoryCLI {
		title = "CLI Commands"
	} else if category == command.CategoryREPL {
		title = "REPL Commands"
	}
	b.WriteString(fmt.Sprintf("%s:\n\n", title))

	// Context-specific intro
	if category == command.CategoryCLI {
		b.WriteString("Use 'magellai <command> --help' for more information about a command.\n\n")
	} else if category == command.CategoryREPL {
		b.WriteString("Use '/help <command>' for more information about a command.\n")
		b.WriteString("Use '/<command>' to execute a command in the REPL.\n\n")
	}

	// Format each category
	categories := []command.Category{command.CategoryShared, command.CategoryCLI, command.CategoryREPL, command.CategoryAPI}
	for _, cat := range categories {
		if cmds, exists := grouped[cat]; exists && len(cmds) > 0 {
			// Filter based on context
			if category != command.CategoryShared && cat != category && cat != command.CategoryShared {
				continue
			}

			title := f.formatCategory(cat)
			if cat == command.CategoryShared && category != command.CategoryShared {
				title = "Available everywhere"
			}

			b.WriteString(fmt.Sprintf("%s:\n", title))
			b.WriteString(f.formatCommandTable(cmds))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// FormatError formats an error message with help
func (f *ContextAwareHelpFormatter) FormatError(err error, suggestion string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Error: %v\n", err))

	if suggestion != "" {
		b.WriteString(fmt.Sprintf("\nDid you mean: %s?\n", suggestion))
	}

	// Context-specific help message
	if f.Category == command.CategoryCLI {
		b.WriteString("\nUse 'magellai help' to see available commands.\n")
	} else if f.Category == command.CategoryREPL {
		b.WriteString("\nUse '/help' to see available commands.\n")
	} else {
		b.WriteString("\nUse 'help' to see available commands.\n")
	}

	return b.String()
}

// formatFlags formats command flags
func (f *ContextAwareHelpFormatter) formatFlags(flags []command.Flag) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	for _, flag := range flags {
		shortFlag := ""
		if flag.Short != "" {
			shortFlag = fmt.Sprintf("-%s, ", flag.Short)
		}

		required := ""
		if flag.Required {
			required = " (required)"
		}

		defaultVal := ""
		if flag.Default != nil {
			defaultVal = fmt.Sprintf(" (default: %v)", flag.Default)
		}

		fmt.Fprintf(w, "  %s--%s\t%s%s%s\n",
			shortFlag, flag.Name, flag.Description, required, defaultVal)
	}

	w.Flush()
	return b.String()
}

// formatCommandTable formats commands as a table
func (f *ContextAwareHelpFormatter) formatCommandTable(commands []command.Interface) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	// Sort commands by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Metadata().Name < commands[j].Metadata().Name
	})

	for _, cmd := range commands {
		meta := cmd.Metadata()
		if meta.Hidden && !f.ShowHidden {
			continue
		}

		name := meta.Name
		if f.ShowAliases && len(meta.Aliases) > 0 {
			name = fmt.Sprintf("%s (%s)", name, strings.Join(meta.Aliases, ", "))
		}

		desc := meta.Description
		if meta.Deprecated != "" {
			desc = fmt.Sprintf("[DEPRECATED] %s", desc)
		}

		fmt.Fprintf(w, "  %s\t%s\n", name, desc)
	}

	w.Flush()
	return b.String()
}

// formatCategory formats a category name
func (f *ContextAwareHelpFormatter) formatCategory(cat command.Category) string {
	switch cat {
	case command.CategoryShared:
		return "All Interfaces"
	case command.CategoryCLI:
		return "CLI Only"
	case command.CategoryREPL:
		return "REPL Only"
	case command.CategoryAPI:
		return "API Only"
	default:
		return "Unknown"
	}
}

// groupByCategory groups commands by category
func (f *ContextAwareHelpFormatter) groupByCategory(commands []command.Interface) map[command.Category][]command.Interface {
	grouped := make(map[command.Category][]command.Interface)

	for _, cmd := range commands {
		cat := cmd.Metadata().Category
		grouped[cat] = append(grouped[cat], cmd)
	}

	return grouped
}

// wrapText wraps text to fit within max width
func (f *ContextAwareHelpFormatter) wrapText(text string, indent int) string {
	lines := strings.Split(text, "\n")
	var result []string

	indentStr := strings.Repeat(" ", indent)
	for _, line := range lines {
		result = append(result, indentStr+line)
	}

	return strings.Join(result, "\n")
}
