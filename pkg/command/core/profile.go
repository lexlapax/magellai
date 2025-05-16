// ABOUTME: Profile command - Manages configuration profiles in Magellai
// ABOUTME: Provides complete profile lifecycle management including create, switch, delete, list

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

// ProfileCommand implements profile management
type ProfileCommand struct {
	config *config.Config
}

// NewProfileCommand creates a new profile command instance
func NewProfileCommand(cfg *config.Config) *ProfileCommand {
	return &ProfileCommand{
		config: cfg,
	}
}

// Execute runs the profile command
func (p *ProfileCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if exec.Data == nil {
		exec.Data = make(map[string]interface{})
	}

	if len(exec.Args) == 0 {
		return p.showCurrentProfile(ctx, exec)
	}

	switch exec.Args[0] {
	case "list":
		return p.listProfiles(ctx, exec)
	case "show":
		if len(exec.Args) > 1 {
			return p.showProfile(ctx, exec, exec.Args[1])
		}
		return p.showCurrentProfile(ctx, exec)
	case "create":
		if len(exec.Args) < 2 {
			return fmt.Errorf("profile create: %w - name required", command.ErrMissingArgument)
		}
		return p.createProfile(ctx, exec, exec.Args[1])
	case "switch":
		if len(exec.Args) < 2 {
			return fmt.Errorf("profile switch: %w - name required", command.ErrMissingArgument)
		}
		return p.switchProfile(ctx, exec, exec.Args[1])
	case "delete":
		if len(exec.Args) < 2 {
			return fmt.Errorf("profile delete: %w - name required", command.ErrMissingArgument)
		}
		return p.deleteProfile(ctx, exec, exec.Args[1])
	case "update":
		if len(exec.Args) < 3 {
			return fmt.Errorf("profile update: %w - name and key=value required", command.ErrMissingArgument)
		}
		return p.updateProfile(ctx, exec, exec.Args[1], exec.Args[2:])
	case "copy":
		if len(exec.Args) < 3 {
			return fmt.Errorf("profile copy: %w - source and destination required", command.ErrMissingArgument)
		}
		return p.copyProfile(ctx, exec, exec.Args[1], exec.Args[2])
	case "export":
		if len(exec.Args) < 2 {
			return fmt.Errorf("profile export: %w - name required", command.ErrMissingArgument)
		}
		return p.exportProfile(ctx, exec, exec.Args[1])
	case "import":
		if len(exec.Args) < 3 {
			return fmt.Errorf("profile import: %w - name and filename required", command.ErrMissingArgument)
		}
		return p.importProfile(ctx, exec, exec.Args[1], exec.Args[2])
	default:
		// If no valid subcommand, assume it's a profile name to switch to
		return p.switchProfile(ctx, exec, exec.Args[0])
	}
}

// Metadata returns the command metadata
func (p *ProfileCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "profile",
		Aliases:     []string{"prof"},
		Description: "Manage configuration profiles",
		LongDescription: `The profile command provides complete profile lifecycle management:

Subcommands:
  list               List all available profiles
  show [name]        Show profile details (current if none specified)
  create <name>      Create a new profile
  switch <name>      Switch to a different profile
  delete <name>      Delete a profile
  update <name> k=v  Update profile settings
  copy <src> <dst>   Copy a profile
  export <name>      Export profile configuration
  import <name> <f>  Import profile from file

Examples:
  profile                  # Show current profile
  profile list             # List all profiles
  profile show work        # Show work profile details
  profile create fast      # Create new fast profile
  profile switch work      # Switch to work profile
  profile work             # Also switches to work profile
  profile update work temperature=0.5
  profile copy work home   # Copy work to home
  profile export work      # Export work profile
  profile import test p.yaml`,
		Category: command.CategoryShared,
		Flags: []command.Flag{
			{
				Name:        "format",
				Short:       "f",
				Description: "Output format (json|yaml|text)",
				Type:        command.FlagTypeString,
				Default:     "text",
			},
		},
	}
}

// Validate checks if the command configuration is valid
func (p *ProfileCommand) Validate() error {
	if p.config == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return nil
}

// showCurrentProfile displays the current active profile
func (p *ProfileCommand) showCurrentProfile(ctx context.Context, exec *command.ExecutionContext) error {
	current := p.config.GetString("profile.current")
	if current == "" {
		current = "default"
	}

	// Get profile details
	profileConfig, err := p.config.GetProfile(current)
	if err != nil {
		// Profile might not exist, just show basic info
		exec.Data["output"] = fmt.Sprintf("Current profile: %s", current)
		return nil
	}

	// Format output
	outputFormat := p.getOutputFormat(exec)
	switch outputFormat {
	case "json":
		// For JSON, just use text output for current profile showing
		exec.Data["output"] = fmt.Sprintf("Current profile: %s", current)
		return nil
	default:
		var output strings.Builder
		output.WriteString(fmt.Sprintf("Current profile: %s\n", current))
		if profileConfig.Provider != "" {
			output.WriteString(fmt.Sprintf("  Provider: %s\n", profileConfig.Provider))
		}
		if profileConfig.Model != "" {
			output.WriteString(fmt.Sprintf("  Model: %s\n", profileConfig.Model))
		}
		if profileConfig.Description != "" {
			output.WriteString(fmt.Sprintf("  Description: %s\n", profileConfig.Description))
		}
		exec.Data["output"] = output.String()
	}

	return nil
}

// listProfiles lists all available profiles
func (p *ProfileCommand) listProfiles(ctx context.Context, exec *command.ExecutionContext) error {
	// Get all profiles from the configuration
	allConfig := p.config.All()
	profiles := make([]string, 0)

	// Look for keys that start with "profiles."
	for key := range allConfig {
		if strings.HasPrefix(key, "profiles.") {
			parts := strings.Split(key, ".")
			if len(parts) >= 2 {
				profileName := parts[1]
				// Skip sub-keys
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

	// Always include default
	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}

	current := p.config.GetString("profile.current")
	if current == "" {
		current = "default"
	}

	// Sort profiles for consistent output
	sort.Strings(profiles)

	outputFormat := p.getOutputFormat(exec)
	if outputFormat == "json" {
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

// showProfile shows details of a specific profile
func (p *ProfileCommand) showProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	profileConfig, err := p.config.GetProfile(name)
	if err != nil {
		return fmt.Errorf("failed to get profile '%s': %w", name, err)
	}

	current := p.config.GetString("profile.current")
	isCurrent := (name == current) || (current == "" && name == "default")

	outputFormat := p.getOutputFormat(exec)
	switch outputFormat {
	case "json":
		data := map[string]interface{}{
			"name":     name,
			"current":  isCurrent,
			"provider": profileConfig.Provider,
			"model":    profileConfig.Model,
			"settings": profileConfig.Settings,
		}
		if profileConfig.Description != "" {
			data["description"] = profileConfig.Description
		}
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		exec.Data["output"] = string(jsonData)
	default:
		var output strings.Builder
		output.WriteString(fmt.Sprintf("Profile: %s", name))
		if isCurrent {
			output.WriteString(" (current)")
		}
		output.WriteString("\n")

		if profileConfig.Description != "" {
			output.WriteString(fmt.Sprintf("  Description: %s\n", profileConfig.Description))
		}
		if profileConfig.Provider != "" {
			output.WriteString(fmt.Sprintf("  Provider: %s\n", profileConfig.Provider))
		}
		if profileConfig.Model != "" {
			output.WriteString(fmt.Sprintf("  Model: %s\n", profileConfig.Model))
		}

		if len(profileConfig.Settings) > 0 {
			output.WriteString("  Settings:\n")
			for key, value := range profileConfig.Settings {
				output.WriteString(fmt.Sprintf("    %s: %v\n", key, value))
			}
		}
		exec.Data["output"] = output.String()
	}

	return nil
}

// createProfile creates a new profile
func (p *ProfileCommand) createProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	key := fmt.Sprintf("profiles.%s", name)
	if p.config.Exists(key) {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	// Get base settings from flags
	description := exec.Flags.GetString("description")
	provider := exec.Flags.GetString("provider")
	model := exec.Flags.GetString("model")
	copyFrom := exec.Flags.GetString("copy-from")

	var profileData map[string]interface{}

	if copyFrom != "" {
		// Copy from existing profile
		sourceProfile, err := p.config.GetProfile(copyFrom)
		if err != nil {
			return fmt.Errorf("failed to get source profile '%s': %w", copyFrom, err)
		}

		profileData = map[string]interface{}{
			"description": sourceProfile.Description,
			"provider":    sourceProfile.Provider,
			"model":       sourceProfile.Model,
			"settings":    sourceProfile.Settings,
		}

		// Override with specified values
		if description != "" {
			profileData["description"] = description
		}
		if provider != "" {
			profileData["provider"] = provider
		}
		if model != "" {
			profileData["model"] = model
		}
	} else {
		// Create new profile from scratch
		profileData = make(map[string]interface{})
		if description != "" {
			profileData["description"] = description
		}
		if provider != "" {
			profileData["provider"] = provider
		}
		if model != "" {
			profileData["model"] = model
		}
	}

	// Create the profile
	if err := p.config.SetValue(key, profileData); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Created profile: %s", name)
	return nil
}

// switchProfile switches to a different profile
func (p *ProfileCommand) switchProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	// Check if profile exists
	key := fmt.Sprintf("profiles.%s", name)
	if !p.config.Exists(key) && name != "default" {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Switch to the profile
	if err := p.config.SetProfile(name); err != nil {
		return fmt.Errorf("failed to switch profile: %w", err)
	}

	// Update current profile in config
	if err := p.config.SetValue("profile.current", name); err != nil {
		return fmt.Errorf("failed to update current profile: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Switched to profile: %s", name)
	return nil
}

// deleteProfile deletes a profile
func (p *ProfileCommand) deleteProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	key := fmt.Sprintf("profiles.%s", name)
	if !p.config.Exists(key) {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Check if this is the current profile
	current := p.config.GetString("profile.current")
	if current == name {
		return fmt.Errorf("cannot delete current profile '%s', switch to another profile first", name)
	}

	// TODO: Config package needs a way to delete keys
	// For now, we'll return an error
	return fmt.Errorf("profile deletion not implemented")
}

// updateProfile updates profile settings
func (p *ProfileCommand) updateProfile(ctx context.Context, exec *command.ExecutionContext, name string, updates []string) error {
	key := fmt.Sprintf("profiles.%s", name)
	if !p.config.Exists(key) && name != "default" {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Parse key=value pairs
	for _, update := range updates {
		parts := strings.SplitN(update, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid update format: %s (expected key=value)", update)
		}

		updateKey := strings.TrimSpace(parts[0])
		updateValue := strings.TrimSpace(parts[1])

		// Handle special keys
		profileKey := fmt.Sprintf("%s.%s", key, updateKey)

		// Convert string values to appropriate types
		var value interface{} = updateValue
		if updateValue == "true" {
			value = true
		} else if updateValue == "false" {
			value = false
		} else {
			// Try to parse as number
			var floatVal float64
			if _, err := fmt.Sscanf(updateValue, "%f", &floatVal); err == nil {
				value = floatVal
			}
		}

		if err := p.config.SetValue(profileKey, value); err != nil {
			return fmt.Errorf("failed to update profile setting '%s': %w", updateKey, err)
		}
	}

	exec.Data["output"] = fmt.Sprintf("Updated profile: %s", name)
	return nil
}

// copyProfile copies a profile to a new name
func (p *ProfileCommand) copyProfile(ctx context.Context, exec *command.ExecutionContext, source, destination string) error {
	// Check if source exists
	sourceKey := fmt.Sprintf("profiles.%s", source)
	if !p.config.Exists(sourceKey) && source != "default" {
		return fmt.Errorf("source profile '%s' not found", source)
	}

	// Check if destination already exists
	destKey := fmt.Sprintf("profiles.%s", destination)
	if p.config.Exists(destKey) {
		return fmt.Errorf("destination profile '%s' already exists", destination)
	}

	// Get source profile
	sourceProfile, err := p.config.GetProfile(source)
	if err != nil {
		return fmt.Errorf("failed to get source profile: %w", err)
	}

	// Create destination profile
	profileData := map[string]interface{}{
		"description": fmt.Sprintf("Copied from %s", source),
		"provider":    sourceProfile.Provider,
		"model":       sourceProfile.Model,
		"settings":    sourceProfile.Settings,
	}

	if err := p.config.SetValue(destKey, profileData); err != nil {
		return fmt.Errorf("failed to create destination profile: %w", err)
	}

	exec.Data["output"] = fmt.Sprintf("Copied profile '%s' to '%s'", source, destination)
	return nil
}

// exportProfile exports a profile configuration
func (p *ProfileCommand) exportProfile(ctx context.Context, exec *command.ExecutionContext, name string) error {
	profile, err := p.config.GetProfile(name)
	if err != nil {
		return fmt.Errorf("failed to export profile: %w", err)
	}

	outputFormat := p.getOutputFormat(exec)

	var output []byte
	switch outputFormat {
	case "json":
		output, err = json.MarshalIndent(profile, "", "  ")
	default:
		// YAML-like format
		var buf strings.Builder
		buf.WriteString(fmt.Sprintf("name: %s\n", name))
		if profile.Description != "" {
			buf.WriteString(fmt.Sprintf("description: %s\n", profile.Description))
		}
		if profile.Provider != "" {
			buf.WriteString(fmt.Sprintf("provider: %s\n", profile.Provider))
		}
		if profile.Model != "" {
			buf.WriteString(fmt.Sprintf("model: %s\n", profile.Model))
		}
		if len(profile.Settings) > 0 {
			buf.WriteString("settings:\n")
			for k, v := range profile.Settings {
				buf.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
			}
		}
		output = []byte(buf.String())
	}

	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	exec.Data["output"] = string(output)
	return nil
}

// importProfile imports a profile from a file
func (p *ProfileCommand) importProfile(ctx context.Context, exec *command.ExecutionContext, name, filename string) error {
	// This would need actual file reading in a real implementation
	return fmt.Errorf("profile import not implemented")
}

// getOutputFormat gets the output format from flags or data
func (p *ProfileCommand) getOutputFormat(exec *command.ExecutionContext) string {
	if format := exec.Flags.GetString("format"); format != "" {
		return format
	}
	if format, ok := exec.Data["outputFormat"].(string); ok && format != "" {
		return format
	}
	return "text"
}
