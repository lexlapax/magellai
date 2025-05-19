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
	"github.com/lexlapax/magellai/pkg/ui"
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
	// If formatter is ContextAwareHelpFormatter, use its color formatter
	var colorFormatter *ui.ColorFormatter
	if formatter, ok := h.formatter.(*ContextAwareHelpFormatter); ok {
		colorFormatter = formatter.colorFormatter
	} else {
		// Fallback to default formatter without colors
		colorFormatter = ui.NewColorFormatter(false, nil)
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(colorFormatter.FormatInfo("Aliases:"))
	b.WriteString("\n")

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

		aliasName := colorFormatter.FormatCommand(name)
		aliasValue := colorFormatter.FormatPrompt(value)
		fmt.Fprintf(w, "  %s\t→ %s\n", aliasName, aliasValue)
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
	ShowHidden     bool
	ShowAliases    bool
	MaxWidth       int
	IndentSpaces   int
	Category       command.Category
	config         *config.Config
	colorFormatter *ui.ColorFormatter
}

// NewContextAwareHelpFormatter creates a new context-aware help formatter
func NewContextAwareHelpFormatter(config *config.Config) *ContextAwareHelpFormatter {
	// Check if colors are enabled in config
	colorsEnabled := config.GetBool("repl.colors.enabled")

	// Only enable colors if we're in a terminal
	if colorsEnabled && !ui.IsTerminal() {
		colorsEnabled = false
	}

	return &ContextAwareHelpFormatter{
		ShowHidden:     false,
		ShowAliases:    true,
		MaxWidth:       80,
		IndentSpaces:   2,
		config:         config,
		colorFormatter: ui.NewColorFormatter(colorsEnabled, nil),
	}
}

// FormatCommand formats help for a single command
func (f *ContextAwareHelpFormatter) FormatCommand(cmd command.Interface) string {
	meta := cmd.Metadata()
	var b strings.Builder

	// Command name and aliases
	b.WriteString(f.colorFormatter.FormatInfo("Command: "))
	b.WriteString(f.colorFormatter.FormatCommand(meta.Name))
	if f.ShowAliases && len(meta.Aliases) > 0 {
		b.WriteString(" (aliases: ")
		b.WriteString(f.colorFormatter.FormatCommand(strings.Join(meta.Aliases, ", ")))
		b.WriteString(")")
	}
	b.WriteString("\n\n")

	// Description
	if meta.Description != "" {
		b.WriteString(f.colorFormatter.FormatInfo("Description: "))
		b.WriteString(meta.Description)
		b.WriteString("\n\n")
	}

	// Long description
	if meta.LongDescription != "" {
		b.WriteString(f.colorFormatter.FormatInfo("Details:"))
		b.WriteString("\n")
		b.WriteString(f.wrapText(meta.LongDescription, f.IndentSpaces))
		b.WriteString("\n\n")
	}

	// Flags
	if len(meta.Flags) > 0 {
		b.WriteString(f.colorFormatter.FormatInfo("Flags:"))
		b.WriteString("\n")
		b.WriteString(f.formatFlags(meta.Flags))
		b.WriteString("\n")
	}

	// Category
	b.WriteString(f.colorFormatter.FormatInfo("Available in: "))
	b.WriteString(f.formatCategory(meta.Category))
	b.WriteString("\n")

	// Context-specific information
	if f.Category == command.CategoryCLI && meta.Category == command.CategoryShared {
		b.WriteString("\n")
		b.WriteString(f.colorFormatter.FormatInfo("CLI-specific usage:"))
		b.WriteString("\n")
		b.WriteString(f.colorFormatter.FormatPrompt(fmt.Sprintf("  magellai %s [flags] [args]", meta.Name)))
		b.WriteString("\n")
	} else if f.Category == command.CategoryREPL && meta.Category == command.CategoryShared {
		b.WriteString("\n")
		b.WriteString(f.colorFormatter.FormatInfo("REPL-specific usage:"))
		b.WriteString("\n")
		b.WriteString(f.colorFormatter.FormatPrompt(fmt.Sprintf("  /%s [flags] [args]", meta.Name)))
		b.WriteString("\n")
	}

	// Deprecated warning
	if meta.Deprecated != "" {
		b.WriteString("\n")
		b.WriteString(f.colorFormatter.FormatError(fmt.Sprintf("DEPRECATED: %s", meta.Deprecated)))
		b.WriteString("\n")
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
	b.WriteString(f.colorFormatter.FormatInfo(fmt.Sprintf("%s:", title)))
	b.WriteString("\n\n")

	// Context-specific intro
	if category == command.CategoryCLI {
		b.WriteString("Use ")
		b.WriteString(f.colorFormatter.FormatPrompt("'magellai <command> --help'"))
		b.WriteString(" for more information about a command.\n\n")
	} else if category == command.CategoryREPL {
		b.WriteString("Use ")
		b.WriteString(f.colorFormatter.FormatPrompt("'/help <command>'"))
		b.WriteString(" for more information about a command.\n")
		b.WriteString("Use ")
		b.WriteString(f.colorFormatter.FormatPrompt("'/<command>'"))
		b.WriteString(" to execute a command in the REPL.\n\n")
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

			b.WriteString(f.colorFormatter.FormatInfo(fmt.Sprintf("%s:", title)))
			b.WriteString("\n")
			b.WriteString(f.formatCommandTable(cmds))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// FormatError formats an error message with help
func (f *ContextAwareHelpFormatter) FormatError(err error, suggestion string) string {
	var b strings.Builder

	b.WriteString(f.colorFormatter.FormatError(fmt.Sprintf("Error: %v", err)))
	b.WriteString("\n")

	if suggestion != "" {
		b.WriteString("\n")
		b.WriteString(f.colorFormatter.FormatInfo("Did you mean: "))
		b.WriteString(f.colorFormatter.FormatCommand(suggestion))
		b.WriteString("?\n")
	}

	// Context-specific help message
	b.WriteString("\n")
	if f.Category == command.CategoryCLI {
		b.WriteString(f.colorFormatter.FormatInfo("Use "))
		b.WriteString(f.colorFormatter.FormatPrompt("'magellai help'"))
		b.WriteString(f.colorFormatter.FormatInfo(" to see available commands."))
	} else if f.Category == command.CategoryREPL {
		b.WriteString(f.colorFormatter.FormatInfo("Use "))
		b.WriteString(f.colorFormatter.FormatPrompt("'/help'"))
		b.WriteString(f.colorFormatter.FormatInfo(" to see available commands."))
	} else {
		b.WriteString(f.colorFormatter.FormatInfo("Use "))
		b.WriteString(f.colorFormatter.FormatPrompt("'help'"))
		b.WriteString(f.colorFormatter.FormatInfo(" to see available commands."))
	}
	b.WriteString("\n")

	return b.String()
}

// formatFlags formats command flags
func (f *ContextAwareHelpFormatter) formatFlags(flags []command.Flag) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	for _, flag := range flags {
		shortFlag := ""
		if flag.Short != "" {
			shortFlag = f.colorFormatter.FormatCommand(fmt.Sprintf("-%s", flag.Short)) + ", "
		}

		required := ""
		if flag.Required {
			required = f.colorFormatter.FormatError(" (required)")
		}

		defaultVal := ""
		if flag.Default != nil {
			defaultVal = f.colorFormatter.FormatInfo(fmt.Sprintf(" (default: %v)", flag.Default))
		}

		flagName := f.colorFormatter.FormatCommand(fmt.Sprintf("--%s", flag.Name))
		fmt.Fprintf(w, "  %s%s\t%s%s%s\n",
			shortFlag, flagName, flag.Description, required, defaultVal)
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

		name := f.colorFormatter.FormatCommand(meta.Name)
		if f.ShowAliases && len(meta.Aliases) > 0 {
			aliases := f.colorFormatter.FormatCommand(strings.Join(meta.Aliases, ", "))
			name = fmt.Sprintf("%s (%s)", name, aliases)
		}

		desc := meta.Description
		if meta.Deprecated != "" {
			desc = f.colorFormatter.FormatError(fmt.Sprintf("[DEPRECATED] %s", desc))
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
