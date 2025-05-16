// ABOUTME: Version command implementation for showing application version info
// ABOUTME: Displays version, commit hash, and build time in configurable formats

package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lexlapax/magellai/pkg/command"
)

// VersionInfo contains version information
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// VersionCommand implements the version command
type VersionCommand struct {
	info VersionInfo
}

// NewVersionCommand creates a new version command
func NewVersionCommand(version, commit, date string) *VersionCommand {
	return &VersionCommand{
		info: VersionInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
}

// Execute runs the version command
func (c *VersionCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	if exec.Data == nil {
		exec.Data = make(map[string]interface{})
	}

	// Check for JSON output format
	format := ""
	if f, ok := exec.Flags["format"].(string); ok {
		format = f
	} else if f, ok := exec.Data["outputFormat"].(string); ok {
		format = f
	}

	if format == "json" {
		jsonData, err := json.MarshalIndent(c.info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal version info: %w", err)
		}
		exec.Data["output"] = string(jsonData)
	} else {
		// Default text format
		output := fmt.Sprintf("magellai version %s", c.info.Version)
		if c.info.Commit != "" && c.info.Commit != "none" {
			output += fmt.Sprintf(" (commit: %s", c.info.Commit)
			if c.info.Date != "" && c.info.Date != "unknown" {
				output += fmt.Sprintf(", built: %s", c.info.Date)
			}
			output += ")"
		}
		exec.Data["output"] = output
	}

	return nil
}

// Metadata returns the command metadata
func (c *VersionCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "version",
		Aliases:     []string{"ver", "v"},
		Description: "Show version information",
		LongDescription: `The version command displays information about the magellai build:
  - Version number
  - Git commit hash
  - Build date

Examples:
  version           # Show version in text format
  version --format json  # Show version in JSON format`,
		Category: command.CategoryShared,
		Flags: []command.Flag{
			{
				Name:        "format",
				Type:        command.FlagTypeString,
				Description: "Output format (text or json)",
			},
		},
	}
}

// Validate checks if the command can be executed
func (c *VersionCommand) Validate() error {
	return nil
}
