// ABOUTME: Configuration utility functions for type-safe access
// ABOUTME: Provides helper methods for getting/setting config values

package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

// GetString returns a string value from the configuration
func (c *Config) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.String(key)
}

// GetInt returns an integer value from the configuration
func (c *Config) GetInt(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.Int(key)
}

// GetBool returns a boolean value from the configuration
func (c *Config) GetBool(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.Bool(key)
}

// GetFloat64 returns a float64 value from the configuration
func (c *Config) GetFloat64(key string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.Float64(key)
}

// GetStringSlice returns a string slice from the configuration
func (c *Config) GetStringSlice(key string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.Strings(key)
}

// GetDuration returns a time.Duration from the configuration
func (c *Config) GetDuration(key string) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.Duration(key)
}

// Get returns an interface{} value from the configuration
func (c *Config) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.Get(key)
}

// Exists checks if a key exists in the configuration
func (c *Config) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.Exists(key)
}

// All returns all configuration as a map
func (c *Config) All() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.koanf.All()
}

// SetValue sets a value with validation
func (c *Config) SetValue(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set the value
	if err := c.koanf.Set(key, value); err != nil {
		return fmt.Errorf("failed to set value for key '%s': %w", key, err)
	}

	// Notify watchers
	c.notifyWatchers()

	return nil
}

// GetProviderConfig returns the configuration for a specific provider
func (c *Config) GetProviderConfig(provider string) (map[string]interface{}, error) {
	key := fmt.Sprintf("provider.%s", provider)
	if !c.Exists(key) {
		return nil, fmt.Errorf("provider '%s' not configured", provider)
	}

	config := c.Get(key)
	if configMap, ok := config.(map[string]interface{}); ok {
		return configMap, nil
	}

	return nil, fmt.Errorf("invalid configuration for provider '%s'", provider)
}

// GetModelSettings returns the settings for a specific model
func (c *Config) GetModelSettings(providerModel string) (*ModelSettings, error) {
	key := fmt.Sprintf("model.settings.%s", providerModel)
	if !c.Exists(key) {
		// Return default settings if model-specific settings don't exist
		return &ModelSettings{}, nil
	}

	var settings ModelSettings
	if err := c.koanf.Unmarshal(key, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model settings: %w", err)
	}

	return &settings, nil
}

// GetDefaultProvider returns the default provider
func (c *Config) GetDefaultProvider() string {
	return c.GetString("provider.default")
}

// GetDefaultModel returns the default model
func (c *Config) GetDefaultModel() string {
	return c.GetString("model.default")
}

// SetDefaultProvider sets the default provider
func (c *Config) SetDefaultProvider(provider string) error {
	return c.SetValue("provider.default", provider)
}

// SetDefaultModel sets the default model
func (c *Config) SetDefaultModel(model string) error {
	return c.SetValue("model.default", model)
}

// GetProfile returns a specific profile configuration
func (c *Config) GetProfile(name string) (*ProfileConfig, error) {
	key := fmt.Sprintf("profiles.%s", name)
	if !c.Exists(key) {
		return nil, fmt.Errorf("%w: %s", ErrProfileNotFound, name)
	}

	var profile ProfileConfig
	if err := c.koanf.Unmarshal(key, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// GetSchema returns the entire configuration as a typed schema
func (c *Config) GetSchema() (*Schema, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var schema Schema
	if err := c.koanf.Unmarshal("", &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &schema, nil
}

// GetProviderAPIKey returns the API key for a specific provider
func (c *Config) GetProviderAPIKey(provider string) string {
	// First check environment variable
	envKey := fmt.Sprintf("%s%s_API_KEY", ConfigEnvPrefix, strings.ToUpper(provider))
	if apiKey := os.Getenv(envKey); apiKey != "" {
		return apiKey
	}

	// Then check config
	configKey := fmt.Sprintf("provider.%s.api_key", strings.ToLower(provider))
	return c.GetString(configKey)
}

// MergeProfile merges a profile's settings into the current configuration
func (c *Config) MergeProfile(profileName string) error {
	profile, err := c.GetProfile(profileName)
	if err != nil {
		return err
	}

	// Apply profile settings
	for key, value := range profile.Settings {
		if err := c.SetValue(key, value); err != nil {
			return fmt.Errorf("failed to apply profile setting '%s': %w", key, err)
		}
	}

	// Set provider and model if specified
	if profile.Provider != "" {
		if err := c.SetDefaultProvider(profile.Provider); err != nil {
			return fmt.Errorf("failed to set default provider: %w", err)
		}
	}
	if profile.Model != "" {
		if err := c.SetDefaultModel(profile.Model); err != nil {
			return fmt.Errorf("failed to set default model: %w", err)
		}
	}

	return nil
}

// Export exports the current configuration
func (c *Config) Export() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.koanf.Marshal(yaml.Parser())
}

// Import imports configuration from data
func (c *Config) Import(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a temporary koanf instance to validate the data
	temp := koanf.New(".")
	if err := temp.Load(rawbytes.Provider(data), yaml.Parser()); err != nil {
		return fmt.Errorf("failed to parse configuration data: %w", err)
	}

	// If valid, merge into the main configuration
	if err := c.koanf.Merge(temp); err != nil {
		return fmt.Errorf("failed to merge configuration: %w", err)
	}

	// Notify watchers
	c.notifyWatchers()

	return nil
}
