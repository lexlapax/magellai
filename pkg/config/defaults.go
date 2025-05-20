// ABOUTME: Comprehensive default configuration generator
// ABOUTME: Provides complete default values for all configuration options

package config

import (
	"os"
	"path/filepath"
)

// GetCompleteDefaultConfig returns a comprehensive default configuration
// with all available options set to sensible defaults
func GetCompleteDefaultConfig() map[string]interface{} {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "magellai")

	return map[string]interface{}{
		// Logging configuration
		"log": map[string]interface{}{
			"level":  "info", // Default to info level
			"format": "text", // text or json
		},

		// Provider configuration
		"provider": map[string]interface{}{
			"default": "openai",
			"openai": map[string]interface{}{
				"api_key":       "", // User should set via env var OPENAI_API_KEY
				"base_url":      "https://api.openai.com/v1",
				"organization":  "",
				"api_version":   "",
				"default_model": "gpt-4o",
				"timeout":       "30s",
				"max_retries":   3,
			},
			"anthropic": map[string]interface{}{
				"api_key":       "", // User should set via env var ANTHROPIC_API_KEY
				"base_url":      "https://api.anthropic.com",
				"api_version":   "2023-06-01",
				"default_model": "claude-3-5-haiku-latest",
				"timeout":       "30s",
				"max_retries":   3,
			},
			"gemini": map[string]interface{}{
				"api_key":       "", // User should set via env var GEMINI_API_KEY
				"base_url":      "https://generativelanguage.googleapis.com/v1beta",
				"project_id":    "",
				"location":      "us-central1",
				"default_model": "gemini-2.0-flash-lite",
				"timeout":       "30s",
				"max_retries":   3,
			},
		},

		// Model configuration
		"model": map[string]interface{}{
			"default": "openai/gpt-4o",
			"settings": map[string]interface{}{
				// Global model settings (can be overridden per model)
				"*": map[string]interface{}{
					"temperature":       0.7,
					"max_tokens":        2048,
					"top_p":             1.0,
					"frequency_penalty": 0.0,
					"presence_penalty":  0.0,
					"stop_sequences":    []string{},
				},
				// Model-specific settings
				"openai/gpt-4o": map[string]interface{}{
					"max_tokens": 4096,
				},
				"openai/gpt-4o-mini": map[string]interface{}{
					"max_tokens": 4096,
				},
				"anthropic/claude-3-5-sonnet-latest": map[string]interface{}{
					"max_tokens": 8192,
				},
				"anthropic/claude-3-5-haiku-latest": map[string]interface{}{
					"max_tokens": 8192,
				},
			},
		},

		// Output configuration
		"output": map[string]interface{}{
			"format": "text", // text, json, yaml, markdown
			"color":  true,   // Enable colored output
			"pretty": true,   // Pretty print JSON/YAML output
		},

		// Session configuration
		"session": map[string]interface{}{
			"directory":   filepath.Join(configDir, "sessions"),
			"autosave":    true,
			"max_age":     "0s", // 0 means no expiration
			"compression": false,
			"storage": map[string]interface{}{
				"type": "filesystem",
				"settings": map[string]interface{}{
					"base_dir": filepath.Join(configDir, "sessions"),
				},
			},
			"auto_recovery": map[string]interface{}{
				"enabled":  true,
				"interval": "30s",
				"max_age":  "24h",
			},
		},

		// REPL configuration
		"repl": map[string]interface{}{
			"colors": map[string]interface{}{
				"enabled": true,
			},
			"prompt_style": "> ",
			"multiline":    false,
			"history_file": filepath.Join(configDir, ".repl_history"),
			"auto_save": map[string]interface{}{
				"enabled":  true,
				"interval": "5m",
			},
		},

		// Plugin configuration
		"plugin": map[string]interface{}{
			"directory": filepath.Join(configDir, "plugins"),
			"path":      []string{}, // Additional paths to search for plugins
			"enabled":   []string{}, // Explicitly enabled plugins
			"disabled":  []string{}, // Explicitly disabled plugins
		},

		// Profiles configuration
		"profiles": map[string]interface{}{
			"fast": map[string]interface{}{
				"description": "Fast responses with lower quality",
				"provider":    "gemini",
				"model":       "gemini-2.0-flash-lite",
				"settings": map[string]interface{}{
					"temperature": 0.3,
					"max_tokens":  1024,
				},
			},
			"quality": map[string]interface{}{
				"description": "High-quality responses, slower",
				"provider":    "openai",
				"model":       "o3",
				"settings": map[string]interface{}{
					"temperature": 0.7,
					"max_tokens":  4096,
				},
			},
			"creative": map[string]interface{}{
				"description": "Creative and diverse responses",
				"provider":    "anthropic",
				"model":       "claude-3-7-sonnet-latest",
				"settings": map[string]interface{}{
					"temperature": 0.9,
					"max_tokens":  4096,
				},
			},
		},

		// Command aliases
		"aliases": map[string]interface{}{
			"q":    "exit",
			"quit": "exit",
			"cls":  "clear",
			"h":    "help",
			"?":    "help",
		},

		// CLI settings
		"cli": map[string]interface{}{
			"stream":  true,  // Enable streaming by default
			"verbose": false, // Verbose output
			"confirm": true,  // Confirm destructive operations
		},
	}
}

// GenerateExampleConfig generates a well-commented example configuration file
func GenerateExampleConfig() string {
	return `# Magellai Configuration File
# This is an example configuration with all available options

# Logging configuration
log:
  level: info      # Options: debug, info, warn, error
  format: text     # Options: text, json

# Provider configuration
provider:
  default: openai  # Default provider to use
  
  # OpenAI configuration
  openai:
    api_key: ""    # Set via environment variable: OPENAI_API_KEY
    base_url: "https://api.openai.com/v1"
    organization: ""
    api_version: ""
    default_model: "gpt-4o"
    timeout: "30s"
    max_retries: 3
  
  # Anthropic (Claude) configuration
  anthropic:
    api_key: ""    # Set via environment variable: ANTHROPIC_API_KEY
    base_url: "https://api.anthropic.com"
    api_version: "2023-06-01"
    default_model: "claude-3-5-haiku-latest"
    timeout: "30s"
    max_retries: 3
  
  # Google Gemini configuration
  gemini:
    api_key: ""    # Set via environment variable: GEMINI_API_KEY
    base_url: "https://generativelanguage.googleapis.com/v1beta"
    project_id: ""
    location: "us-central1"
    default_model: "gemini-2.0-flash-lite"
    timeout: "30s"
    max_retries: 3

# Model configuration
model:
  default: "openai/gpt-4o"  # Default model in provider/model format
  settings:
    # Global settings (applied to all models unless overridden)
    "*":
      temperature: 0.7
      max_tokens: 2048
      top_p: 1.0
      frequency_penalty: 0.0
      presence_penalty: 0.0
      stop_sequences: []
    
    # Model-specific settings
    "openai/gpt-4o":
      max_tokens: 4096
    "openai/gpt-4o-mini":
      max_tokens: 4096
    "anthropic/claude-3-5-sonnet-latest":
      max_tokens: 8192
    "anthropic/claude-3-5-haiku-latest":
      max_tokens: 8192

# Output configuration
output:
  format: text     # Options: text, json, yaml, markdown
  color: true      # Enable colored output
  pretty: true     # Pretty print JSON/YAML output

# Session configuration
session:
  directory: "~/.config/magellai/sessions"
  autosave: true
  max_age: "0s"    # 0 means no expiration
  compression: false
  storage:
    type: filesystem
    settings:
      base_dir: "~/.config/magellai/sessions"
  auto_recovery:
    enabled: true
    interval: "30s"
    max_age: "24h"

# REPL configuration
repl:
  colors:
    enabled: true
  prompt_style: "> "
  multiline: false
  history_file: "~/.config/magellai/.repl_history"
  auto_save:
    enabled: true
    interval: "5m"

# Plugin configuration
plugin:
  directory: "~/.config/magellai/plugins"
  path: []         # Additional paths to search for plugins
  enabled: []      # Explicitly enabled plugins
  disabled: []     # Explicitly disabled plugins

# Profiles - Named configurations for different use cases
profiles:
  fast:
    description: "Fast responses with lower quality"
    provider: gemini
    model: gemini-2.0-flash-lite
    settings:
      temperature: 0.3
      max_tokens: 1024
  
  quality:
    description: "High-quality responses, slower"
    provider: openai
    model: o3
    settings:
      temperature: 0.7
      max_tokens: 4096
  
  creative:
    description: "Creative and diverse responses"
    provider: anthropic
    model: claude-3-7-sonnet-latest
    settings:
      temperature: 0.9
      max_tokens: 4096

# Command aliases
aliases:
  q: exit
  quit: exit
  cls: clear
  h: help
  "?": help

# CLI settings
cli:
  stream: true     # Enable streaming by default
  verbose: false   # Verbose output
  confirm: true    # Confirm destructive operations
`
}

// GetConfigTemplate returns a configuration template with explanations
func GetConfigTemplate() string {
	return GenerateExampleConfig()
}
