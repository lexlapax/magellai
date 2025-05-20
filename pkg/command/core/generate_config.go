// ABOUTME: Command to generate example configuration file
// ABOUTME: Creates a well-commented configuration template for users

package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
)

// GenerateConfigCommand generates an example configuration file
type GenerateConfigCommand struct{}

// NewGenerateConfigCommand creates a new generate-config command
func NewGenerateConfigCommand() command.Interface {
	return &GenerateConfigCommand{}
}

// Execute implements command.Interface
func (c *GenerateConfigCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	// Get the output path from flags or use default
	outputPath := exec.Flags.GetString("output")
	if outputPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logging.LogError(err, "Failed to get home directory")
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputPath = filepath.Join(homeDir, ".config", "magellai", "config.yaml")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logging.LogError(err, "Failed to create directory", "dir", dir)
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil && !exec.Flags.GetBool("force") {
		return fmt.Errorf("config file already exists at %s. Use --force to overwrite", outputPath)
	}

	// Generate example config
	configContent := config.GenerateExampleConfig()

	// Write to file
	if err := os.WriteFile(outputPath, []byte(configContent), 0644); err != nil {
		logging.LogError(err, "Failed to write config file", "path", outputPath)
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logging.LogInfo("Generated example configuration", "path", outputPath)
	fmt.Fprintf(exec.Stdout, "Successfully generated example configuration at: %s\n", outputPath)

	// Show additional tips
	fmt.Fprintln(exec.Stdout, "\nTips:")
	fmt.Fprintln(exec.Stdout, "- Set API keys via environment variables (e.g., OPENAI_API_KEY)")
	fmt.Fprintln(exec.Stdout, "- Customize settings based on your preferences")
	fmt.Fprintln(exec.Stdout, "- Use profiles for different use cases (fast, quality, creative)")
	fmt.Fprintln(exec.Stdout, "- Check 'magellai config validate' to verify your configuration")

	return nil
}

// Metadata implements command.Interface
func (c *GenerateConfigCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "generate-config",
		Aliases:     []string{"gen-config", "init-config"},
		Description: "Generate an example configuration file",
		LongDescription: `Generate a well-commented example configuration file with all available options.
This creates a template configuration that you can customize for your needs.`,
		Category: command.CategoryShared,
		Flags: []command.Flag{
			{
				Name:        "output",
				Short:       "o",
				Description: "Output path for configuration file",
				Type:        command.FlagTypeString,
				Default:     "",
			},
			{
				Name:        "force",
				Short:       "f",
				Description: "Overwrite existing configuration file",
				Type:        command.FlagTypeBool,
				Default:     false,
			},
		},
	}
}

// Validate implements command.Interface
func (c *GenerateConfigCommand) Validate() error {
	return nil
}
