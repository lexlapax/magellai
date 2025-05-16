// ABOUTME: Configuration management using koanf with multi-layer support
// ABOUTME: Handles config loading, merging, validation, and profile management

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	// "github.com/lexlapax/magellai/internal/logging"
	"github.com/spf13/pflag"
)

const (
	// ConfigEnvPrefix is the prefix for environment variables
	ConfigEnvPrefix = "MAGELLAI_"

	// Default paths
	SystemConfigPath  = "/etc/magellai/config.yaml"
	UserConfigDir     = "~/.config/magellai"
	UserConfigFile    = "config.yaml"
	ProjectConfigFile = ".magellai.yaml"
)

// Config represents the global configuration
type Config struct {
	// Koanf instance for configuration management
	koanf      *koanf.Koanf
	mu         sync.RWMutex
	watchers   []func()
	currentDir string

	// Configuration layers in order of precedence (lowest to highest)
	defaults map[string]interface{}
	profile  string // current profile name
}

// Manager is the global configuration manager instance
var Manager *Config

// Init initializes the configuration manager
func Init() error {
	Manager = &Config{
		koanf:    koanf.New("."),
		defaults: getDefaultConfig(),
	}

	// Get current working directory for project config search
	cwd, err := os.Getwd()
	if err != nil {
		// logging.Warn("Failed to get current directory", "error", err)
		cwd = "."
	}
	Manager.currentDir = cwd

	return nil
}

// Load loads configuration from all sources in precedence order
func (c *Config) Load(flags *pflag.FlagSet) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Load defaults
	if err := c.koanf.Load(confmap.Provider(c.defaults, "."), nil); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	// 2. Load system config if exists
	_ = c.loadFile(SystemConfigPath) // Ignore error if file doesn't exist

	// 3. Load user config
	userConfig := expandPath(filepath.Join(UserConfigDir, UserConfigFile))
	_ = c.loadFile(userConfig) // Ignore error if file doesn't exist

	// 4. Load project config (search upward from current directory)
	if projectConfig := c.findProjectConfig(); projectConfig != "" {
		_ = c.loadFile(projectConfig) // Ignore error if file doesn't exist
	}

	// 5. Load environment variables
	if err := c.koanf.Load(env.Provider(ConfigEnvPrefix, ".", func(s string) string {
		// Convert MAGELLAI_PROVIDER_API_KEY to provider.api_key
		s = strings.ToLower(strings.TrimPrefix(s, ConfigEnvPrefix))
		s = strings.ReplaceAll(s, "_", ".")
		return s
	}), nil); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// 6. Load command-line flags (if provided)
	if flags != nil {
		if err := c.koanf.Load(posflag.Provider(flags, ".", c.koanf), nil); err != nil {
			return fmt.Errorf("failed to load command-line flags: %w", err)
		}
	}

	// Apply profile overrides if a profile is set
	if c.profile != "" {
		if err := c.applyProfile(c.profile); err != nil {
			return fmt.Errorf("failed to apply profile: %w", err)
		}
	}

	return nil
}

// LoadFile loads a specific configuration file
func (c *Config) LoadFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loadFile(path)
}

// loadFile is the internal file loading method (not thread-safe)
func (c *Config) loadFile(path string) error {
	expandedPath := expandPath(path)

	// Check if file exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return err
	}

	// Load the file
	if err := c.koanf.Load(file.Provider(expandedPath), yaml.Parser()); err != nil {
		return fmt.Errorf("failed to load config file %s: %w", path, err)
	}

	// logging.Debug("Loaded config file", "path", expandedPath)
	return nil
}

// findProjectConfig searches for a project config file starting from current directory
func (c *Config) findProjectConfig() string {
	dir := c.currentDir
	for {
		configPath := filepath.Join(dir, ProjectConfigFile)
		if _, err := os.Stat(configPath); err == nil {
			// logging.Debug("Found project config", "path", configPath)
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}
	return ""
}

// SetProfile sets the active profile
func (c *Config) SetProfile(profile string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.profile = profile
	return c.applyProfile(profile)
}

// applyProfile applies profile-specific overrides (not thread-safe)
func (c *Config) applyProfile(profile string) error {
	profileKey := fmt.Sprintf("profiles.%s", profile)
	if !c.koanf.Exists(profileKey) {
		return fmt.Errorf("profile '%s' not found", profile)
	}

	// Get profile config
	profileConfig := c.koanf.Cut(profileKey)

	// Merge profile config over current config
	if err := c.koanf.Merge(profileConfig); err != nil {
		return fmt.Errorf("failed to apply profile '%s': %w", profile, err)
	}

	// logging.Debug("Applied profile", "profile", profile)
	return nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"log": map[string]interface{}{
			"level":  "info",
			"format": "text",
		},
		"provider": map[string]interface{}{
			"default": "openai",
		},
		"output": map[string]interface{}{
			"format": "text",
			"color":  true,
		},
		"session": map[string]interface{}{
			"directory": expandPath("~/.config/magellai/sessions"),
			"autosave":  true,
		},
		"plugin": map[string]interface{}{
			"directory": expandPath("~/.config/magellai/plugins"),
		},
	}
}

// expandPath expands ~ to user home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// Watch enables configuration file watching
func (c *Config) Watch(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.watchers = append(c.watchers, callback)
	// TODO: Implement file watching using fsnotify
}

// notifyWatchers notifies all registered watchers of config changes
func (c *Config) notifyWatchers() {
	for _, watcher := range c.watchers {
		go watcher()
	}
}
