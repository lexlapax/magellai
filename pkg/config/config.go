// ABOUTME: Configuration management using koanf with multi-layer support
// ABOUTME: Handles config loading, merging, validation, and profile management

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/lexlapax/magellai/internal/logging"
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
	defaults    map[string]interface{}
	profile     string   // current profile name
	loadedFiles []string // list of successfully loaded config files
}

// Manager is the global configuration manager instance
var Manager *Config

// GetLoadedFiles returns the list of configuration files that were successfully loaded
func (c *Config) GetLoadedFiles() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	files := make([]string, len(c.loadedFiles))
	copy(files, c.loadedFiles)
	return files
}

// Init initializes the configuration manager
func Init() error {
	logging.LogInfo("Initializing configuration manager")

	Manager = &Config{
		koanf:    koanf.New("."),
		defaults: GetCompleteDefaultConfig(),
	}

	// Get current working directory for project config search
	cwd, err := os.Getwd()
	if err != nil {
		logging.LogWarn("Failed to get current directory", "error", err)
		cwd = "."
	}
	Manager.currentDir = cwd

	logging.LogDebug("Configuration manager initialized", "currentDir", cwd)
	return nil
}

// Load loads configuration from all sources in precedence order
func (c *Config) Load(cmdlineOverrides map[string]interface{}) error {
	start := time.Now()
	logging.LogInfo("Loading configuration from all sources")

	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Load defaults
	logging.LogDebug("Loading default configuration")
	if err := c.koanf.Load(confmap.Provider(c.defaults, "."), nil); err != nil {
		logging.LogError(err, "Failed to load defaults")
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	// 2. Load system config if exists
	logging.LogDebug("Loading system configuration", "path", SystemConfigPath)
	_ = c.loadFile(SystemConfigPath) // Ignore error if file doesn't exist

	// 3. Load user config
	userConfig := expandPath(filepath.Join(UserConfigDir, UserConfigFile))
	logging.LogDebug("Loading user configuration", "path", userConfig)
	_ = c.loadFile(userConfig) // Ignore error if file doesn't exist

	// 4. Load project config (search upward from current directory)
	logging.LogDebug("Searching for project configuration")
	if projectConfig := c.findProjectConfig(); projectConfig != "" {
		logging.LogInfo("Found project configuration", "path", projectConfig)
		_ = c.loadFile(projectConfig) // Ignore error if file doesn't exist
	}

	// 5. Load environment variables
	logging.LogDebug("Loading environment variables", "prefix", ConfigEnvPrefix)
	if err := c.koanf.Load(env.Provider(ConfigEnvPrefix, ".", func(s string) string {
		// Convert MAGELLAI_PROVIDER_API_KEY to provider.api_key
		s = strings.ToLower(strings.TrimPrefix(s, ConfigEnvPrefix))
		s = strings.ReplaceAll(s, "_", ".")
		return s
	}), nil); err != nil {
		logging.LogError(err, "Failed to load environment variables")
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// 6. Load command-line overrides (if provided)
	if len(cmdlineOverrides) > 0 {
		logging.LogDebug("Loading command-line configuration overrides")
		if err := c.koanf.Load(confmap.Provider(cmdlineOverrides, "."), nil); err != nil {
			logging.LogError(err, "Failed to load command-line overrides")
			return fmt.Errorf("failed to load command-line overrides: %w", err)
		}
	}

	// Apply profile overrides if a profile is set
	if c.profile != "" {
		logging.LogInfo("Applying profile overrides", "profile", c.profile)
		if err := c.applyProfile(c.profile); err != nil {
			logging.LogError(err, "Failed to apply profile", "profile", c.profile)
			return fmt.Errorf("failed to apply profile: %w", err)
		}
	}

	duration := time.Since(start)
	logging.LogInfo("Configuration loading completed successfully")
	logging.LogDebug("Configuration load time", "duration", duration)
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

	logging.LogDebug("Attempting to load configuration file", "path", path, "expandedPath", expandedPath)

	// Check if file exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		logging.LogDebug("Configuration file not found", "path", expandedPath)
		return err
	}

	// Load the file
	if err := c.koanf.Load(file.Provider(expandedPath), yaml.Parser()); err != nil {
		logging.LogError(err, "Failed to load config file", "path", expandedPath)
		return fmt.Errorf("failed to load config file %s: %w", path, err)
	}

	// Track successfully loaded files
	c.loadedFiles = append(c.loadedFiles, expandedPath)

	logging.LogDebug("Loaded config file successfully", "path", expandedPath)
	return nil
}

// findProjectConfig searches for a project config file starting from current directory
func (c *Config) findProjectConfig() string {
	logging.LogDebug("Searching for project configuration file", "startDir", c.currentDir, "filename", ProjectConfigFile)

	dir := c.currentDir
	for {
		configPath := filepath.Join(dir, ProjectConfigFile)
		logging.LogDebug("Checking for project config", "path", configPath)

		if _, err := os.Stat(configPath); err == nil {
			logging.LogDebug("Found project config", "path", configPath)
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
	logging.LogInfo("Switching to profile", "profile", profile)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.profile = profile
	err := c.applyProfile(profile)
	if err != nil {
		logging.LogError(err, "Failed to set profile", "profile", profile)
		return err
	}

	logging.LogInfo("Successfully switched to profile", "profile", profile)
	return nil
}

// applyProfile applies profile-specific overrides (not thread-safe)
func (c *Config) applyProfile(profile string) error {
	logging.LogDebug("Applying profile configuration", "profile", profile)

	profileKey := fmt.Sprintf("profiles.%s", profile)
	if !c.koanf.Exists(profileKey) {
		logging.LogWarn("Profile not found", "profile", profile)
		return fmt.Errorf("profile '%s' not found", profile)
	}

	// Get profile config
	profileConfig := c.koanf.Cut(profileKey)

	// Merge profile config over current config
	if err := c.koanf.Merge(profileConfig); err != nil {
		logging.LogError(err, "Failed to merge profile configuration", "profile", profile)
		return fmt.Errorf("failed to apply profile '%s': %w", profile, err)
	}

	logging.LogDebug("Applied profile successfully", "profile", profile)
	return nil
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
	logging.LogDebug("Adding configuration watcher")

	c.mu.Lock()
	defer c.mu.Unlock()

	c.watchers = append(c.watchers, callback)
	logging.LogDebug("Configuration watcher added", "totalWatchers", len(c.watchers))
	// TODO: Implement file watching using fsnotify
}

// notifyWatchers notifies all registered watchers of config changes
func (c *Config) notifyWatchers() {
	for _, watcher := range c.watchers {
		go watcher()
	}
}

// GetPrimaryConfigFile returns the path to the primary configuration file
func (c *Config) GetPrimaryConfigFile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Try user config first
	userConfig := expandPath(filepath.Join(UserConfigDir, UserConfigFile))
	if _, err := os.Stat(userConfig); err == nil {
		return userConfig
	}

	// Try project config
	projectConfig := c.findProjectConfig()
	if projectConfig != "" {
		return projectConfig
	}

	// Fallback to system config
	if _, err := os.Stat(SystemConfigPath); err == nil {
		return SystemConfigPath
	}

	// Default to user config path even if it doesn't exist
	return userConfig
}

// Reload reloads the configuration from all sources
func (c *Config) Reload() error {
	logging.LogInfo("Reloading configuration from all sources")

	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new koanf instance and store the old one
	oldKoanf := c.koanf
	c.koanf = koanf.New(".")

	// Load defaults first
	if err := c.loadDefaults(); err != nil {
		logging.LogError(err, "Failed to reload defaults")
		c.koanf = oldKoanf
		return err
	}

	// Load system config if exists
	_ = c.loadFile(SystemConfigPath)

	// Load user config
	userConfig := expandPath(filepath.Join(UserConfigDir, UserConfigFile))
	_ = c.loadFile(userConfig)

	// Load project config (search upward from current directory)
	if projectConfig := c.findProjectConfig(); projectConfig != "" {
		_ = c.loadFile(projectConfig)
	}

	// Load environment variables
	if err := c.koanf.Load(env.Provider(ConfigEnvPrefix, ".", func(s string) string {
		s = strings.ToLower(strings.TrimPrefix(s, ConfigEnvPrefix))
		s = strings.ReplaceAll(s, "_", ".")
		return s
	}), nil); err != nil {
		c.koanf = oldKoanf
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Apply profile overrides if a profile is set
	if c.profile != "" {
		if err := c.applyProfile(c.profile); err != nil {
			logging.LogError(err, "Failed to apply profile during reload", "profile", c.profile)
			c.koanf = oldKoanf
			return fmt.Errorf("failed to apply profile: %w", err)
		}
	}

	c.notifyWatchers()
	logging.LogInfo("Configuration reload completed successfully")
	return nil
}

// loadDefaults loads the default configuration
func (c *Config) loadDefaults() error {
	return c.koanf.Load(confmap.Provider(c.defaults, "."), nil)
}

// DeleteKey deletes a configuration key
func (c *Config) DeleteKey(key string) error {
	logging.LogInfo("Deleting configuration key", "key", key)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key exists
	if !c.koanf.Exists(key) {
		logging.LogWarn("Key not found for deletion", "key", key)
		return fmt.Errorf("key not found: %s", key)
	}

	// koanf doesn't have direct delete support, so we need to work around it
	// Get all config, delete the key, and reload
	allConfig := c.koanf.All()
	deleteNestedKey(allConfig, key)

	// Create new koanf instance with updated config
	newKoanf := koanf.New(".")
	if err := newKoanf.Load(confmap.Provider(allConfig, "."), nil); err != nil {
		logging.LogError(err, "Failed to reload config after delete", "key", key)
		return fmt.Errorf("failed to reload config after delete: %w", err)
	}

	c.koanf = newKoanf
	c.notifyWatchers()

	logging.LogInfo("Successfully deleted configuration key", "key", key)
	return nil
}

// deleteNestedKey deletes a nested key from a map
func deleteNestedKey(m map[string]interface{}, key string) {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		delete(m, key)
		return
	}

	current := m
	for i := 0; i < len(parts)-1; i++ {
		if next, ok := current[parts[i]].(map[string]interface{}); ok {
			current = next
		} else {
			return // Key path doesn't exist
		}
	}

	delete(current, parts[len(parts)-1])
}
