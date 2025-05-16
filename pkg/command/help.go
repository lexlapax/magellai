// ABOUTME: Unified help system for CLI and REPL interfaces
// ABOUTME: Provides help formatting and display for all command types

package command

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

// HelpFormatter formats help text for commands
type HelpFormatter interface {
	// FormatCommand formats help for a single command
	FormatCommand(cmd Interface) string

	// FormatCommandList formats a list of commands
	FormatCommandList(commands []Interface, category Category) string

	// FormatError formats an error message with help
	FormatError(err error, suggestion string) string
}

// DefaultHelpFormatter provides default help formatting
type DefaultHelpFormatter struct {
	ShowHidden   bool
	ShowAliases  bool
	MaxWidth     int
	IndentSpaces int
}

// NewDefaultHelpFormatter creates a new default help formatter
func NewDefaultHelpFormatter() *DefaultHelpFormatter {
	return &DefaultHelpFormatter{
		ShowHidden:   false,
		ShowAliases:  true,
		MaxWidth:     80,
		IndentSpaces: 2,
	}
}

// FormatCommand formats help for a single command
func (f *DefaultHelpFormatter) FormatCommand(cmd Interface) string {
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

	// Deprecated warning
	if meta.Deprecated != "" {
		b.WriteString(fmt.Sprintf("\nDEPRECATED: %s\n", meta.Deprecated))
	}

	return b.String()
}

// FormatCommandList formats a list of commands
func (f *DefaultHelpFormatter) FormatCommandList(commands []Interface, category Category) string {
	var b strings.Builder

	// Group commands by category
	grouped := f.groupByCategory(commands)

	// Header
	b.WriteString(fmt.Sprintf("Available Commands (%s):\n\n", f.formatCategory(category)))

	// Format each category
	categories := []Category{CategoryShared, CategoryCLI, CategoryREPL, CategoryAPI}
	for _, cat := range categories {
		if cmds, exists := grouped[cat]; exists && len(cmds) > 0 {
			if category != CategoryShared && cat != category && cat != CategoryShared {
				continue
			}

			b.WriteString(fmt.Sprintf("%s Commands:\n", f.formatCategory(cat)))
			b.WriteString(f.formatCommandTable(cmds))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// FormatError formats an error message with help
func (f *DefaultHelpFormatter) FormatError(err error, suggestion string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Error: %v\n", err))

	if suggestion != "" {
		b.WriteString(fmt.Sprintf("\nDid you mean: %s?\n", suggestion))
	}

	b.WriteString("\nUse 'help' to see available commands.\n")

	return b.String()
}

// formatFlags formats command flags
func (f *DefaultHelpFormatter) formatFlags(flags []Flag) string {
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
func (f *DefaultHelpFormatter) formatCommandTable(commands []Interface) string {
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
func (f *DefaultHelpFormatter) formatCategory(cat Category) string {
	switch cat {
	case CategoryShared:
		return "All Interfaces"
	case CategoryCLI:
		return "CLI Only"
	case CategoryREPL:
		return "REPL Only"
	case CategoryAPI:
		return "API Only"
	default:
		return "Unknown"
	}
}

// groupByCategory groups commands by category
func (f *DefaultHelpFormatter) groupByCategory(commands []Interface) map[Category][]Interface {
	grouped := make(map[Category][]Interface)

	for _, cmd := range commands {
		cat := cmd.Metadata().Category
		grouped[cat] = append(grouped[cat], cmd)
	}

	return grouped
}

// wrapText wraps text to fit within max width
func (f *DefaultHelpFormatter) wrapText(text string, indent int) string {
	// Simple implementation - can be enhanced with proper word wrapping
	lines := strings.Split(text, "\n")
	var result []string

	indentStr := strings.Repeat(" ", indent)
	for _, line := range lines {
		result = append(result, indentStr+line)
	}

	return strings.Join(result, "\n")
}

// HelpCommand is a command that displays help
type HelpCommand struct {
	registry  *Registry
	formatter HelpFormatter
	category  Category
}

// NewHelpCommand creates a new help command
func NewHelpCommand(registry *Registry, formatter HelpFormatter, category Category) *HelpCommand {
	return &HelpCommand{
		registry:  registry,
		formatter: formatter,
		category:  category,
	}
}

// Execute implements Interface
func (h *HelpCommand) Execute(ctx context.Context, exec *ExecutionContext) error {
	// Check if specific command help is requested
	if len(exec.Args) > 0 {
		cmdName := exec.Args[0]
		cmd, err := h.registry.Get(cmdName)
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

		// Show specific command help
		fmt.Fprintln(exec.Stdout, h.formatter.FormatCommand(cmd))
		return nil
	}

	// Show general help
	commands := h.registry.List(h.category)
	fmt.Fprintln(exec.Stdout, h.formatter.FormatCommandList(commands, h.category))
	return nil
}

// Metadata implements Interface
func (h *HelpCommand) Metadata() *Metadata {
	return &Metadata{
		Name:        "help",
		Aliases:     []string{"h", "?"},
		Description: "Display help information",
		Category:    CategoryShared,
		Flags: []Flag{
			{
				Name:        "all",
				Short:       "a",
				Description: "Show all commands including hidden ones",
				Type:        FlagTypeBool,
				Default:     false,
			},
		},
	}
}

// Validate implements Interface
func (h *HelpCommand) Validate() error {
	if h.registry == nil {
		return fmt.Errorf("help command requires a registry")
	}
	if h.formatter == nil {
		return fmt.Errorf("help command requires a formatter")
	}
	return nil
}
