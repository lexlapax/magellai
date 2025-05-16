// ABOUTME: Constants for the command package
// ABOUTME: Defines common constants used across commands

package command

// OutputFormat represents the format of command output
type OutputFormat string

const (
	// OutputFormatText is the default text output format
	OutputFormatText OutputFormat = "text"
	
	// OutputFormatJSON is JSON output format
	OutputFormatJSON OutputFormat = "json"
	
	// OutputFormatYAML is YAML output format
	OutputFormatYAML OutputFormat = "yaml"
)