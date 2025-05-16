// ABOUTME: Configuration directory management for Magellai
// ABOUTME: Provides functions to create and access user configuration directories
package configdir

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DirPermission is the default permission for directories
	DirPermission = 0755
	// FilePermission is the default permission for files
	FilePermission = 0644
)

// Paths contains all configuration directory paths
type Paths struct {
	Base     string // Base config directory (~/.config/magellai)
	Sessions string // Session storage directory
	Plugins  string // Plugin installation directory
	Logs     string // Log files directory
}

// GetPaths returns the configuration directory paths for the current user
func GetPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("getting home directory: %w", err)
	}

	base := filepath.Join(home, ".config", "magellai")
	
	return Paths{
		Base:     base,
		Sessions: filepath.Join(base, "sessions"),
		Plugins:  filepath.Join(base, "plugins"),
		Logs:     filepath.Join(base, "logs"),
	}, nil
}

// EnsureDirectories creates all necessary configuration directories
func EnsureDirectories() error {
	paths, err := GetPaths()
	if err != nil {
		return err
	}

	// Create all directories
	dirs := []string{
		paths.Base,
		paths.Sessions,
		paths.Plugins,
		paths.Logs,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, DirPermission); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return nil
}

// ConfigFile returns the path to the main configuration file
func ConfigFile() (string, error) {
	paths, err := GetPaths()
	if err != nil {
		return "", err
	}
	return filepath.Join(paths.Base, "config.yaml"), nil
}

// CreateDefaultConfig creates a default configuration file if it doesn't exist
func CreateDefaultConfig() error {
	configPath, err := ConfigFile()
	if err != nil {
		return err
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config already exists
	}

	// Ensure base directory exists
	if err := EnsureDirectories(); err != nil {
		return err
	}

	// Create default configuration
	defaultConfig := `# Magellai Configuration File
# This is the user-level configuration

# Default model settings
default:
  model: openai/gpt-3.5-turbo  # Format: provider/model
  temperature: 0.7
  max_tokens: 2048
  stream: false

# Provider configurations
providers:
  openai:
    api_key: ${OPENAI_API_KEY}
    # base_url: https://api.openai.com/v1
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    # base_url: https://api.anthropic.com
  gemini:
    api_key: ${GEMINI_API_KEY}

# Output preferences
output:
  format: text  # text, json, or markdown
  color: auto   # auto, always, or never

# Logging configuration
logging:
  level: info   # debug, info, warn, or error
  file: ~/.config/magellai/logs/magellai.log

# Session storage
storage:
  sessions: ~/.config/magellai/sessions
  plugins: ~/.config/magellai/plugins

# Named profiles for different use cases
profiles:
  work:
    model: anthropic/claude-3-opus
    temperature: 0.3
  creative:
    model: openai/gpt-4
    temperature: 0.9

# Command aliases
aliases:
  gpt4: "ask --model openai/gpt-4"
  claude: "ask --model anthropic/claude-3-opus"
`

	return os.WriteFile(configPath, []byte(defaultConfig), FilePermission)
}

// ProjectConfigFile looks for a project-specific configuration file
// It searches upward from the current directory for .magellai.yaml
func ProjectConfigFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}

	// Walk up the directory tree looking for .magellai.yaml
	dir := cwd
	for {
		configPath := filepath.Join(dir, ".magellai.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root
			break
		}
		dir = parent
	}

	return "", nil // No project config found
}

// SystemConfigFile returns the path to the system-wide configuration file
func SystemConfigFile() string {
	return "/etc/magellai/config.yaml"
}