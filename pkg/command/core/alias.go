// ABOUTME: Alias command - Manages command aliases for CLI and REPL use
// ABOUTME: Provides add, remove, list, show functionality for aliases

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
)

// AliasCommand implements alias management
type AliasCommand struct {
	config *config.Config
}

// NewAliasCommand creates a new alias command instance
func NewAliasCommand(cfg *config.Config) *AliasCommand {
	return &AliasCommand{
		config: cfg,
	}
}

// Execute runs the alias command
func (a *AliasCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if exec.Data == nil {
		exec.Data = make(map[string]interface{})
	}

	if len(exec.Args) == 0 {
		return a.listAliases(ctx, exec)
	}

	switch exec.Args[0] {
	case "list":
		return a.listAliases(ctx, exec)
	case "add", "set":
		if len(exec.Args) < 3 {
			return fmt.Errorf("alias add: %w - name and command required", command.ErrMissingArgument)
		}
		return a.addAlias(ctx, exec, exec.Args[1], strings.Join(exec.Args[2:], " "))
	case "remove", "delete", "rm":
		if len(exec.Args) < 2 {
			return fmt.Errorf("alias remove: %w - name required", command.ErrMissingArgument)
		}
		return a.removeAlias(ctx, exec, exec.Args[1])
	case "show", "get":
		if len(exec.Args) < 2 {
			return fmt.Errorf("alias show: %w - name required", command.ErrMissingArgument)
		}
		return a.showAlias(ctx, exec, exec.Args[1])
	case "clear":
		return a.clearAliases(ctx, exec)
	case "export":
		return a.exportAliases(ctx, exec)
	case "import":
		if len(exec.Args) < 2 {
			return fmt.Errorf("alias import: %w - filename required", command.ErrMissingArgument)
		}
		return a.importAliases(ctx, exec, exec.Args[1])
	default:
		// If no valid subcommand, assume it's showing a specific alias
		return a.showAlias(ctx, exec, exec.Args[0])
	}
}

// Metadata returns the command metadata
func (a *AliasCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "alias",
		Aliases:     []string{"aliases"},
		Description: "Manage command aliases",
		LongDescription: `The alias command manages command aliases for both CLI and REPL use:

Subcommands:
  list              List all aliases
  add <n> <cmd>     Add or update an alias
  remove <n>        Remove an alias
  show <n>          Show a specific alias
  clear             Remove all aliases
  export            Export aliases
  import <f>        Import aliases from file

Examples:
  alias                       # List all aliases
  alias list                  # List all aliases
  alias add gpt4 "model gpt-4"                 # Create alias for model switch
  alias add claude "model anthropic/claude-3"  # Create provider/model alias
  alias add fast "model gpt-3.5-turbo --temperature 0.1"  # With options
  alias show gpt4             # Show specific alias
  alias remove gpt4           # Remove an alias
  alias clear                 # Remove all aliases
  alias export > aliases.json # Export aliases
  alias import aliases.json   # Import aliases

Note: Aliases can be used in both CLI and REPL modes. In CLI, they're expanded
before command execution. In REPL, they can be invoked directly.`,
		Category: command.CategoryShared,
		Flags: []command.Flag{
			{
				Name:        "format",
				Short:       "f",
				Description: "Output format (json|yaml|text)",
				Type:        command.FlagTypeString,
				Default:     "text",
			},
			{
				Name:        "scope",
				Short:       "s",
				Description: "Alias scope (all|cli|repl)",
				Type:        command.FlagTypeString,
				Default:     "all",
			},
		},
	}
}

// Validate checks if the command configuration is valid
func (a *AliasCommand) Validate() error {
	if a.config == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return nil
}

// listAliases lists all configured aliases
func (a *AliasCommand) listAliases(ctx context.Context, exec *command.ExecutionContext) error {
	// Get CLI command aliases
	cliAliases := a.getAllAliases("aliases")

	// Get REPL-specific aliases
	replAliases := a.getAllAliases("repl.aliases")

	scope := a.getScope(exec)
	outputFormat := a.getOutputFormat(exec)

	// Combine aliases based on scope
	aliases := make(map[string]string)
	if scope == "all" || scope == "cli" {
		for k, v := range cliAliases {
			aliases[k] = v
		}
	}
	if scope == "all" || scope == "repl" {
		for k, v := range replAliases {
			aliases[fmt.Sprintf("%s (repl)", k)] = v
		}
	}

	if len(aliases) == 0 {
		exec.Data["output"] = "No aliases defined"
		return nil
	}

	switch outputFormat {
	case "json":
		data := map[string]interface{}{
			"aliases": aliases,
			"count":   len(aliases),
		}
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		exec.Data["output"] = string(jsonData)
	default:
		// Text format
		var output strings.Builder
		output.WriteString("Defined aliases:\n")

		// Sort aliases for consistent output
		names := make([]string, 0, len(aliases))
		for name := range aliases {
			names = append(names, name)
		}
		sort.Strings(names)

		// Find longest name for alignment
		maxLen := 0
		for _, name := range names {
			if len(name) > maxLen {
				maxLen = len(name)
			}
		}

		for _, name := range names {
			cmd := aliases[name]
			output.WriteString(fmt.Sprintf("  %-*s → %s\n", maxLen+2, name, cmd))
		}

		exec.Data["output"] = output.String()
	}

	return nil
}

// addAlias adds or updates an alias
func (a *AliasCommand) addAlias(ctx context.Context, exec *command.ExecutionContext, name, command string) error {
	// Validate alias name
	if strings.Contains(name, " ") {
		return fmt.Errorf("alias name cannot contain spaces")
	}

	// Check for reserved names
	reserved := []string{"help", "version", "config", "model", "profile", "alias"}
	for _, r := range reserved {
		if name == r {
			return fmt.Errorf("cannot override reserved command: %s", name)
		}
	}

	scope := a.getScope(exec)

	// Set alias based on scope
	if scope == "repl" {
		key := fmt.Sprintf("repl.aliases.%s", name)
		if err := a.config.SetValue(key, command); err != nil {
			return fmt.Errorf("failed to set alias: %w", err)
		}
	} else {
		// Default to CLI aliases
		key := fmt.Sprintf("aliases.%s", name)
		if err := a.config.SetValue(key, command); err != nil {
			return fmt.Errorf("failed to set alias: %w", err)
		}
	}

	exec.Data["output"] = fmt.Sprintf("Alias '%s' created: %s", name, command)
	return nil
}

// removeAlias removes an alias
func (a *AliasCommand) removeAlias(ctx context.Context, exec *command.ExecutionContext, name string) error {
	scope := a.getScope(exec)
	found := false

	// Try to remove from CLI aliases
	if scope == "all" || scope == "cli" {
		key := fmt.Sprintf("aliases.%s", name)
		if a.config.Exists(key) {
			// TODO: Config package needs a way to delete keys
			// For now, we'll set it to empty string
			if err := a.config.SetValue(key, ""); err != nil {
				return fmt.Errorf("failed to remove alias: %w", err)
			}
			found = true
		}
	}

	// Try to remove from REPL aliases
	if scope == "all" || scope == "repl" {
		key := fmt.Sprintf("repl.aliases.%s", name)
		if a.config.Exists(key) {
			// TODO: Config package needs a way to delete keys
			// For now, we'll set it to empty string
			if err := a.config.SetValue(key, ""); err != nil {
				return fmt.Errorf("failed to remove alias: %w", err)
			}
			found = true
		}
	}

	if !found {
		return fmt.Errorf("alias '%s' not found", name)
	}

	exec.Data["output"] = fmt.Sprintf("Alias '%s' removed", name)
	return nil
}

// showAlias shows a specific alias
func (a *AliasCommand) showAlias(ctx context.Context, exec *command.ExecutionContext, name string) error {
	// Check CLI aliases
	cliKey := fmt.Sprintf("aliases.%s", name)
	if a.config.Exists(cliKey) {
		command := a.config.GetString(cliKey)
		exec.Data["output"] = fmt.Sprintf("%s → %s", name, command)
		return nil
	}

	// Check REPL aliases
	replKey := fmt.Sprintf("repl.aliases.%s", name)
	if a.config.Exists(replKey) {
		command := a.config.GetString(replKey)
		exec.Data["output"] = fmt.Sprintf("%s (repl) → %s", name, command)
		return nil
	}

	return fmt.Errorf("alias '%s' not found", name)
}

// clearAliases removes all aliases
func (a *AliasCommand) clearAliases(ctx context.Context, exec *command.ExecutionContext) error {
	scope := a.getScope(exec)
	cleared := 0

	// Clear CLI aliases
	if scope == "all" || scope == "cli" {
		aliases := a.getAllAliases("aliases")
		for name := range aliases {
			key := fmt.Sprintf("aliases.%s", name)
			if err := a.config.SetValue(key, ""); err != nil {
				return fmt.Errorf("failed to clear alias '%s': %w", name, err)
			}
			cleared++
		}
	}

	// Clear REPL aliases
	if scope == "all" || scope == "repl" {
		aliases := a.getAllAliases("repl.aliases")
		for name := range aliases {
			key := fmt.Sprintf("repl.aliases.%s", name)
			if err := a.config.SetValue(key, ""); err != nil {
				return fmt.Errorf("failed to clear alias '%s': %w", name, err)
			}
			cleared++
		}
	}

	exec.Data["output"] = fmt.Sprintf("Cleared %d aliases", cleared)
	return nil
}

// exportAliases exports aliases to JSON
func (a *AliasCommand) exportAliases(ctx context.Context, exec *command.ExecutionContext) error {
	scope := a.getScope(exec)

	// Collect aliases
	result := make(map[string]interface{})

	if scope == "all" || scope == "cli" {
		result["cli"] = a.getAllAliases("aliases")
	}

	if scope == "all" || scope == "repl" {
		result["repl"] = a.getAllAliases("repl.aliases")
	}

	// Export as JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to export aliases: %w", err)
	}

	exec.Data["output"] = string(jsonData)
	return nil
}

// importAliases imports aliases from a file
func (a *AliasCommand) importAliases(ctx context.Context, exec *command.ExecutionContext, filename string) error {
	// TODO: Implement file reading and alias import
	return fmt.Errorf("alias import not implemented")
}

// getAllAliases gets all aliases from a specific config path
func (a *AliasCommand) getAllAliases(prefix string) map[string]string {
	aliases := make(map[string]string)

	// Get all config keys
	allConfig := a.config.All()

	// Filter for alias keys
	for key, value := range allConfig {
		if strings.HasPrefix(key, prefix+".") {
			// Extract alias name
			parts := strings.Split(key, ".")
			if len(parts) >= 2 {
				aliasName := parts[len(parts)-1]
				if strValue, ok := value.(string); ok && strValue != "" {
					aliases[aliasName] = strValue
				}
			}
		}
	}

	// Also check if there's a direct alias map
	if aliasMap := a.config.Get(prefix); aliasMap != nil {
		if m, ok := aliasMap.(map[string]interface{}); ok {
			for k, v := range m {
				if strValue, ok := v.(string); ok && strValue != "" {
					aliases[k] = strValue
				}
			}
		}
	}

	return aliases
}

// getScope gets the scope from flags
func (a *AliasCommand) getScope(exec *command.ExecutionContext) string {
	if scope := exec.Flags.GetString("scope"); scope != "" {
		return scope
	}
	return "all"
}

// getOutputFormat gets the output format from flags or data
func (a *AliasCommand) getOutputFormat(exec *command.ExecutionContext) string {
	if format := exec.Flags.GetString("format"); format != "" {
		return format
	}
	if format, ok := exec.Data["outputFormat"].(string); ok && format != "" {
		return format
	}
	return "text"
}
