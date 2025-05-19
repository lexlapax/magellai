// ABOUTME: Config command - Manages configuration settings for Magellai
// ABOUTME: Supports listing, getting, setting, validation, export/import, and profile management

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	osExec "os/exec"
	"sort"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/pkg/llm"
)

// ConfigCommand implements configuration management
type ConfigCommand struct {
	config *config.Config
}

// NewConfigCommand creates a new config command instance
func NewConfigCommand(cfg *config.Config) *ConfigCommand {
	return &ConfigCommand{
		config: cfg,
	}
}

// Execute runs the config command
func (c *ConfigCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if exec.Data == nil {
		exec.Data = make(map[string]interface{})
	}

	if len(exec.Args) == 0 {
		return c.showCurrentConfig(ctx, exec)
	}

	switch exec.Args[0] {
	case "list":
		return c.listConfig(ctx, exec)
	case "get":
		if len(exec.Args) < 2 {
			return fmt.Errorf("config get: %w - key required", command.ErrMissingArgument)
		}
		return c.getConfig(ctx, exec, exec.Args[1])
	case "set":
		if len(exec.Args) < 3 {
			return fmt.Errorf("config set: %w - key and value required", command.ErrMissingArgument)
		}
		return c.setConfig(ctx, exec, exec.Args[1], exec.Args[2])
	case "validate":
		return c.validateConfig(ctx, exec)
	case "export":
		return c.exportConfig(ctx, exec)
	case "import":
		if len(exec.Args) < 2 {
			return fmt.Errorf("config import: %w - filename required", command.ErrMissingArgument)
		}
		return c.importConfig(ctx, exec, exec.Args[1])
	case "edit":
		return c.editConfig(ctx, exec)
	case "profiles":
		if len(exec.Args) < 2 {
			return c.listProfiles(ctx, exec)
		}
		return c.handleProfileCommand(ctx, exec, exec.Args[1:])
	default:
		return fmt.Errorf("config: %w - invalid subcommand '%s'", command.ErrInvalidArguments, exec.Args[0])
	}
}

// Metadata returns the command metadata
func (c *ConfigCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "config",
		Aliases:     []string{"cfg"},
		Description: "Manage configuration settings",
		LongDescription: `The config command provides comprehensive configuration management:

Subcommands:
  list                List all configuration settings
  get <key>          Get a specific setting value
  set <key> <value>  Set a configuration value
  validate           Validate the current configuration
  export             Export configuration to stdout
  import <file>      Import configuration from file
  edit               Open configuration in editor
  profiles           Manage configuration profiles
    list             List all profiles
    switch <name>    Switch to a profile
    create <name>    Create a new profile
    delete <name>    Delete a profile
    export <name>    Export a profile

Examples:
  config list               # Show all settings
  config get provider      # Get current provider
  config set model gpt-4   # Set default model
  config validate          # Check configuration
  config export > my.yaml  # Export config
  config import my.yaml    # Import config
  config profiles list     # List profiles
  config profiles switch work  # Switch to work profile`,
		Category: command.CategoryShared,
		Flags: []command.Flag{
			{
				Name:        "format",
				Short:       "f",
				Description: "Output format for export/list (json|yaml|text)",
				Type:        command.FlagTypeString,
				Default:     "text",
			},
		},
	}
}

// Validate checks if the command configuration is valid
func (c *ConfigCommand) Validate() error {
	if c.config == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return nil
}

// showCurrentConfig displays the current configuration overview
func (c *ConfigCommand) showCurrentConfig(ctx context.Context, exec *command.ExecutionContext) error {
	// For tests, let's return simplified output
	if exec.Data["outputFormat"] != nil || exec.Flags.GetString("format") == "json" {
		// JSON output
		info := map[string]interface{}{
			"provider": c.config.GetDefaultProvider(),
			"model":    c.config.GetDefaultModel(),
			"profile":  c.config.GetString("profile.current"),
		}
		if info["profile"] == "" {
			info["profile"] = "default"
		}
		jsonBytes, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config to JSON: %w", err)
		}
		exec.Data["output"] = string(jsonBytes)
		return nil
	}

	// Text format for regular showCurrentConfig
	var output strings.Builder
	output.WriteString("Current configuration:\n")
	output.WriteString(fmt.Sprintf("  Provider: %s\n", c.config.GetDefaultProvider()))
	output.WriteString(fmt.Sprintf("  Model: %s\n", c.config.GetDefaultModel()))

	profile := c.config.GetString("profile.current")
	if profile == "" {
		profile = "default"
	}
	output.WriteString(fmt.Sprintf("  Profile: %s\n", profile))

	exec.Data["output"] = output.String()

	// Also write to stdout for interactive use
	fmt.Fprint(exec.Stdout, output.String())

	return nil
}

// listConfig lists all configuration settings
func (c *ConfigCommand) listConfig(ctx context.Context, exec *command.ExecutionContext) error {
	allSettings := c.config.All()

	outputFormat := exec.Flags.GetString("format")
	if outputFormat == "" {
		outputFormat = "text"
	}

	exec.Data["output"] = formatSettings(allSettings, outputFormat)
	return nil
}

// getConfig gets a specific configuration value
func (c *ConfigCommand) getConfig(ctx context.Context, exec *command.ExecutionContext, key string) error {
	// Special handling for shortcuts
	if key == "provider" {
		key = "provider.default"
	} else if key == "model" {
		key = "model.default"
	}

	value := c.config.Get(key)
	if value == nil {
		return fmt.Errorf("key not found: %s", key)
	}

	exec.Data["output"] = fmt.Sprintf("%s: %v", key, value)
	return nil
}

// setConfig sets a configuration value
func (c *ConfigCommand) setConfig(ctx context.Context, exec *command.ExecutionContext, key, value string) error {
	// Handle provider/model setting specially
	if key == "provider" {
		previousValue := c.config.GetDefaultProvider()
		if err := c.config.SetDefaultProvider(value); err != nil {
			return fmt.Errorf("failed to set provider: %w", err)
		}
		logging.LogInfo("Configuration changed", "key", "provider", "old", previousValue, "new", value)
		exec.Data["output"] = fmt.Sprintf("Provider set to: %s", value)
		return nil
	}

	if key == "model" {
		// Parse model string and set both provider and model if needed
		previousModel := c.config.GetDefaultModel()
		providerStr, modelStr := llm.ParseModelString(value)
		if providerStr != "" {
			previousProvider := c.config.GetDefaultProvider()
			if err := c.config.SetDefaultProvider(providerStr); err != nil {
				return fmt.Errorf("failed to set provider: %w", err)
			}
			if previousProvider != providerStr {
				logging.LogInfo("Configuration changed", "key", "provider", "old", previousProvider, "new", providerStr)
			}
		}
		if err := c.config.SetDefaultModel(modelStr); err != nil {
			return fmt.Errorf("failed to set model: %w", err)
		}
		logging.LogInfo("Configuration changed", "key", "model", "old", previousModel, "new", value)
		exec.Data["output"] = fmt.Sprintf("Model set to: %s", value)
		return nil
	}

	// For other keys, use generic set
	previousValue := c.config.GetString(key)
	if err := c.config.SetValue(key, value); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	logging.LogInfo("Configuration changed", "key", key, "old", previousValue, "new", value)
	exec.Data["output"] = fmt.Sprintf("%s set to: %s", key, value)
	return nil
}

// validateConfig validates the current configuration
func (c *ConfigCommand) validateConfig(ctx context.Context, exec *command.ExecutionContext) error {
	err := c.config.Validate()
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	exec.Data["output"] = "Configuration is valid"
	return nil
}

// exportConfig exports the configuration
func (c *ConfigCommand) exportConfig(ctx context.Context, exec *command.ExecutionContext) error {
	outputFormat := exec.Flags.GetString("format")
	if outputFormat == "" {
		outputFormat = "yaml"
	}

	if outputFormat == "json" {
		// Export as JSON
		allConfig := c.config.All()
		config, err := json.MarshalIndent(allConfig, "", "  ")
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		exec.Data["output"] = string(config)
	} else {
		// Export as YAML (default)
		config, err := c.config.Export()
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		exec.Data["output"] = string(config)
	}
	return nil
}

// importConfig imports configuration from a file
func (c *ConfigCommand) importConfig(ctx context.Context, exec *command.ExecutionContext, filename string) error {
	err := c.config.LoadFile(filename)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Configuration imported from: %s", filename)
	return nil
}

// editConfig opens the configuration file in the user's editor
func (c *ConfigCommand) editConfig(ctx context.Context, exec *command.ExecutionContext) error {
	// Get the primary config file path
	configFile := c.config.GetPrimaryConfigFile()
	if configFile == "" {
		return fmt.Errorf("no configuration file found")
	}

	// Get the editor from environment or use a default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors
		for _, e := range []string{"vim", "vi", "nano", "emacs", "code", "subl"} {
			if _, err := osExec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found - set $EDITOR or $VISUAL environment variable")
	}

	// Execute the editor
	cmd := osExec.Command(editor, configFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Reload the configuration after editing
	if err := c.config.Reload(); err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Configuration edited and reloaded: %s", configFile)
	return nil
}

// handleProfileCommand handles profile subcommands
func (c *ConfigCommand) handleProfileCommand(ctx context.Context, exec *command.ExecutionContext, args []string) error {
	if len(args) == 0 {
		return c.listProfiles(ctx, exec)
	}

	action := args[0]
	switch action {
	case "list":
		return c.listProfiles(ctx, exec)
	case "switch":
		if len(args) < 2 {
			return fmt.Errorf("config profiles switch: %w - name required", command.ErrMissingArgument)
		}
		return c.switchProfile(ctx, exec, args[1])
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("config profiles create: %w - name required", command.ErrMissingArgument)
		}
		return c.createProfile(ctx, exec, args[1])
	case "delete":
		if len(args) < 2 {
			return fmt.Errorf("config profiles delete: %w - name required", command.ErrMissingArgument)
		}
		return c.deleteProfile(ctx, exec, args[1])
	case "export":
		if len(args) < 2 {
			return fmt.Errorf("config profiles export: %w - name required", command.ErrMissingArgument)
		}
		return c.exportProfile(ctx, exec, args[1])
	default:
		return fmt.Errorf("config profiles: %w - invalid subcommand '%s'", command.ErrInvalidArguments, action)
	}
}

// listProfiles lists all available profiles
func (c *ConfigCommand) listProfiles(ctx context.Context, exec *command.ExecutionContext) error {
	// Get all profiles from the configuration
	allConfig := c.config.All()
	profiles := make([]string, 0)

	// Look for keys that start with "profiles."
	for key := range allConfig {
		if strings.HasPrefix(key, "profiles.") {
			parts := strings.Split(key, ".")
			if len(parts) >= 2 {
				profileName := parts[1]
				// Skip sub-keys like profiles.default.provider
				if !strings.Contains(profileName, ".") && profileName != "" {
					found := false
					for _, p := range profiles {
						if p == profileName {
							found = true
							break
						}
					}
					if !found {
						profiles = append(profiles, profileName)
					}
				}
			}
		}
	}

	// If no profiles exist but we're looking for default, include it
	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}

	current := c.config.GetString("profile.current")
	if current == "" {
		current = "default"
	}

	// Sort profiles for consistent output
	sort.Strings(profiles)

	if format, ok := exec.Data["outputFormat"].(string); ok && format == "json" {
		data := map[string]interface{}{
			"profiles": profiles,
			"current":  current,
		}
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		exec.Data["output"] = string(jsonData)
		return nil
	}

	var output strings.Builder
	output.WriteString("Available profiles:\n")
	for _, profile := range profiles {
		if profile == current {
			output.WriteString(fmt.Sprintf("  * %s (current)\n", profile))
		} else {
			output.WriteString(fmt.Sprintf("    %s\n", profile))
		}
	}
	exec.Data["output"] = output.String()
	return nil
}

// switchProfile switches to a different profile
func (c *ConfigCommand) switchProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	err := c.config.SetProfile(name)
	if err != nil {
		return fmt.Errorf("failed to switch profile: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Switched to profile: %s", name)
	return nil
}

// createProfile creates a new profile
func (c *ConfigCommand) createProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	key := fmt.Sprintf("profiles.%s", name)
	if c.config.Exists(key) {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// Create an empty profile
	err := c.config.SetValue(key, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Created profile: %s", name)
	return nil
}

// deleteProfile deletes a profile
func (c *ConfigCommand) deleteProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	// Don't allow deleting the default profile
	if name == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}

	// Check if this is the current profile
	currentProfile := c.config.GetString("profile.current")
	if currentProfile == name {
		return fmt.Errorf("cannot delete the currently active profile")
	}

	key := fmt.Sprintf("profiles.%s", name)
	if !c.config.Exists(key) {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Delete the profile
	if err := c.config.DeleteKey(key); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Deleted profile: %s", name)
	return nil
}

// exportProfile exports a specific profile
func (c *ConfigCommand) exportProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	outputFormat := exec.Flags.GetString("format")
	if outputFormat == "" {
		outputFormat = "yaml"
	}

	profile, err := c.config.GetProfile(name)
	if err != nil {
		return fmt.Errorf("failed to export profile: %w", err)
	}

	var config []byte
	if outputFormat == "json" {
		config, err = json.MarshalIndent(profile, "", "  ")
	} else {
		// TODO: Add yaml marshaling when yaml package is available
		// For now, convert to JSON
		config, err = json.MarshalIndent(profile, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	exec.Data["output"] = string(config)
	return nil
}

// formatSettings formats all settings for display
func formatSettings(settings map[string]interface{}, format string) string {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(settings, "", "  ")
		return string(data)
	case "yaml":
		// Simple YAML-like formatting
		var output strings.Builder
		formatValue(&output, settings, 0)
		return output.String()
	default:
		// Text format
		var output strings.Builder
		output.WriteString("Configuration settings:\n")
		formatSettingsText(&output, settings, "  ")
		return output.String()
	}
}

// formatSettingsText recursively formats settings as text
func formatSettingsText(output *strings.Builder, settings interface{}, indent string) {
	switch v := settings.(type) {
	case map[string]interface{}:
		// Sort keys for consistent output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			output.WriteString(fmt.Sprintf("%s%s:", indent, k))
			if nested, ok := v[k].(map[string]interface{}); ok {
				output.WriteString("\n")
				formatSettingsText(output, nested, indent+"  ")
			} else {
				output.WriteString(fmt.Sprintf(" %v\n", v[k]))
			}
		}
	default:
		output.WriteString(fmt.Sprintf("%s%v\n", indent, v))
	}
}

// formatValue recursively formats values in YAML-like style
func formatValue(output *strings.Builder, value interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)

	switch v := value.(type) {
	case map[string]interface{}:
		if indent > 0 {
			output.WriteString("\n")
		}
		for k, val := range v {
			output.WriteString(fmt.Sprintf("%s%s:", indentStr, k))
			if _, isMap := val.(map[string]interface{}); isMap {
				formatValue(output, val, indent+1)
			} else {
				output.WriteString(fmt.Sprintf(" %v\n", val))
			}
		}
	default:
		output.WriteString(fmt.Sprintf(" %v\n", v))
	}
}
