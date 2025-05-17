// ABOUTME: Configuration validation functions
// ABOUTME: Ensures configuration values meet required constraints

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/storage"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field string
	Value interface{}
	Error string
}

// Validate validates the configuration
func (c *Config) Validate() error {
	logging.LogDebug("Starting configuration validation")

	var errors []ValidationError

	// Validate log configuration
	if err := c.validateLogConfig(); err != nil {
		errors = append(errors, err...)
	}

	// Validate provider configuration
	if err := c.validateProviderConfig(); err != nil {
		errors = append(errors, err...)
	}

	// Validate output configuration
	if err := c.validateOutputConfig(); err != nil {
		errors = append(errors, err...)
	}

	// Validate session configuration
	if err := c.validateSessionConfig(); err != nil {
		errors = append(errors, err...)
	}

	// Validate plugin configuration
	if err := c.validatePluginConfig(); err != nil {
		errors = append(errors, err...)
	}

	// Validate profiles
	if err := c.validateProfiles(); err != nil {
		errors = append(errors, err...)
	}

	if len(errors) > 0 {
		for _, err := range errors {
			logging.LogWarn("Configuration validation error",
				"field", err.Field,
				"value", err.Value,
				"error", err.Error)
		}
		logging.LogError(nil, "Configuration validation failed", "errorCount", len(errors))
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	logging.LogDebug("Configuration validation completed successfully")
	return nil
}

// validateLogConfig validates logging configuration
func (c *Config) validateLogConfig() []ValidationError {
	var errors []ValidationError

	level := c.GetString("log.level")
	validLevels := []string{"debug", "info", "warn", "error"}
	if !containsString(validLevels, level) {
		errors = append(errors, ValidationError{
			Field: "log.level",
			Value: level,
			Error: fmt.Sprintf("invalid log level, must be one of: %v", validLevels),
		})
	}

	format := c.GetString("log.format")
	validFormats := []string{"text", "json"}
	if !containsString(validFormats, format) {
		errors = append(errors, ValidationError{
			Field: "log.format",
			Value: format,
			Error: fmt.Sprintf("invalid log format, must be one of: %v", validFormats),
		})
	}

	return errors
}

// validateProviderConfig validates provider configuration
func (c *Config) validateProviderConfig() []ValidationError {
	var errors []ValidationError

	defaultProvider := c.GetString("provider.default")
	validProviders := []string{"openai", "anthropic", "gemini"}
	if defaultProvider != "" && !containsString(validProviders, defaultProvider) {
		errors = append(errors, ValidationError{
			Field: "provider.default",
			Value: defaultProvider,
			Error: fmt.Sprintf("invalid default provider, must be one of: %v", validProviders),
		})
	}

	// Validate specific provider configurations if they exist
	for _, provider := range validProviders {
		if c.Exists(fmt.Sprintf("provider.%s", provider)) {
			if err := c.validateSpecificProvider(provider); err != nil {
				errors = append(errors, err...)
			}
		}
	}

	return errors
}

// validateSpecificProvider validates a specific provider's configuration
func (c *Config) validateSpecificProvider(provider string) []ValidationError {
	var errors []ValidationError

	// Check if API key is provided (either in config or environment)
	apiKey := c.GetProviderAPIKey(provider)
	if apiKey == "" {
		errors = append(errors, ValidationError{
			Field: fmt.Sprintf("provider.%s.api_key", provider),
			Value: nil,
			Error: fmt.Sprintf("API key for %s not found in config or environment", provider),
		})
	}

	// Validate timeout if specified
	timeoutKey := fmt.Sprintf("provider.%s.timeout", provider)
	if c.Exists(timeoutKey) {
		timeout := c.GetDuration(timeoutKey)
		if timeout <= 0 {
			errors = append(errors, ValidationError{
				Field: timeoutKey,
				Value: timeout,
				Error: "timeout must be greater than 0",
			})
		}
	}

	// Validate max_retries if specified
	retriesKey := fmt.Sprintf("provider.%s.max_retries", provider)
	if c.Exists(retriesKey) {
		retries := c.GetInt(retriesKey)
		if retries < 0 {
			errors = append(errors, ValidationError{
				Field: retriesKey,
				Value: retries,
				Error: "max_retries must be >= 0",
			})
		}
	}

	return errors
}

// validateOutputConfig validates output configuration
func (c *Config) validateOutputConfig() []ValidationError {
	var errors []ValidationError

	format := c.GetString("output.format")
	validFormats := []string{"text", "json", "markdown"}
	if !containsString(validFormats, format) {
		errors = append(errors, ValidationError{
			Field: "output.format",
			Value: format,
			Error: fmt.Sprintf("invalid output format, must be one of: %v", validFormats),
		})
	}

	return errors
}

// validateSessionConfig validates session configuration
func (c *Config) validateSessionConfig() []ValidationError {
	var errors []ValidationError

	sessionDir := c.GetString("session.directory")
	if sessionDir != "" {
		expandedDir := expandPath(sessionDir)
		// Create directory if it doesn't exist
		if err := os.MkdirAll(expandedDir, 0755); err != nil {
			errors = append(errors, ValidationError{
				Field: "session.directory",
				Value: sessionDir,
				Error: fmt.Sprintf("failed to create session directory: %v", err),
			})
		}
	}

	// Validate max_age if specified
	if c.Exists("session.max_age") {
		maxAge := c.GetDuration("session.max_age")
		if maxAge < 0 {
			errors = append(errors, ValidationError{
				Field: "session.max_age",
				Value: maxAge,
				Error: "max_age must be >= 0",
			})
		}
	}

	// Validate storage configuration
	storageType := c.GetString("session.storage.type")
	if storageType != "" {
		// Check if the storage backend is available
		if !storage.IsBackendAvailable(storage.BackendType(storageType)) {
			availableBackends := storage.GetAvailableBackends()
			errors = append(errors, ValidationError{
				Field: "session.storage.type",
				Value: storageType,
				Error: fmt.Sprintf("storage backend '%s' is not available. Available backends: %v",
					storageType, availableBackends),
			})
		}

		// Validate storage-specific settings
		switch storage.BackendType(storageType) {
		case storage.FileSystemBackend:
			// Validate filesystem storage settings
			baseDir := c.GetString("session.storage.settings.base_dir")
			if baseDir != "" {
				expandedDir := expandPath(baseDir)
				if err := os.MkdirAll(expandedDir, 0755); err != nil {
					errors = append(errors, ValidationError{
						Field: "session.storage.settings.base_dir",
						Value: baseDir,
						Error: fmt.Sprintf("failed to create storage base directory: %v", err),
					})
				}
			}
		case storage.SQLiteBackend:
			// Validate SQLite storage settings
			dbPath := c.GetString("session.storage.settings.db_path")
			if dbPath != "" {
				expandedPath := expandPath(dbPath)
				dbDir := filepath.Dir(expandedPath)
				if err := os.MkdirAll(dbDir, 0755); err != nil {
					errors = append(errors, ValidationError{
						Field: "session.storage.settings.db_path",
						Value: dbPath,
						Error: fmt.Sprintf("failed to create database directory: %v", err),
					})
				}
			}
		}
	}

	return errors
}

// validatePluginConfig validates plugin configuration
func (c *Config) validatePluginConfig() []ValidationError {
	var errors []ValidationError

	pluginDir := c.GetString("plugin.directory")
	if pluginDir != "" {
		expandedDir := expandPath(pluginDir)
		// Create directory if it doesn't exist
		if err := os.MkdirAll(expandedDir, 0755); err != nil {
			errors = append(errors, ValidationError{
				Field: "plugin.directory",
				Value: pluginDir,
				Error: fmt.Sprintf("failed to create plugin directory: %v", err),
			})
		}
	}

	// Validate plugin paths
	pluginPaths := c.GetStringSlice("plugin.path")
	for i, path := range pluginPaths {
		expandedPath := expandPath(path)
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			errors = append(errors, ValidationError{
				Field: fmt.Sprintf("plugin.path[%d]", i),
				Value: path,
				Error: "plugin path does not exist",
			})
		}
	}

	return errors
}

// validateProfiles validates profile configurations
func (c *Config) validateProfiles() []ValidationError {
	var errors []ValidationError

	if !c.Exists("profiles") {
		return errors
	}

	profiles := c.Get("profiles")
	if profileMap, ok := profiles.(map[string]interface{}); ok {
		for name := range profileMap {
			if err := c.validateProfile(name); err != nil {
				errors = append(errors, err...)
			}
		}
	}

	return errors
}

// validateProfile validates a specific profile
func (c *Config) validateProfile(name string) []ValidationError {
	var errors []ValidationError

	profile, err := c.GetProfile(name)
	if err != nil {
		errors = append(errors, ValidationError{
			Field: fmt.Sprintf("profiles.%s", name),
			Value: nil,
			Error: err.Error(),
		})
		return errors
	}

	// Validate provider if specified
	if profile.Provider != "" {
		validProviders := []string{"openai", "anthropic", "gemini"}
		if !containsString(validProviders, profile.Provider) {
			errors = append(errors, ValidationError{
				Field: fmt.Sprintf("profiles.%s.provider", name),
				Value: profile.Provider,
				Error: fmt.Sprintf("invalid provider, must be one of: %v", validProviders),
			})
		}
	}

	// Validate model format if specified
	if profile.Model != "" {
		pair := ParseProviderModel(profile.Model)
		if pair.Provider == "" && profile.Provider == "" && c.GetDefaultProvider() == "" {
			errors = append(errors, ValidationError{
				Field: fmt.Sprintf("profiles.%s.model", name),
				Value: profile.Model,
				Error: "model specified without provider and no default provider set",
			})
		}
	}

	return errors
}

// containsString checks if a string is in a slice
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ValidateModelSettings validates model-specific settings
func ValidateModelSettings(settings *ModelSettings) []ValidationError {
	var errors []ValidationError

	if settings.Temperature != nil {
		if *settings.Temperature < 0 || *settings.Temperature > 2 {
			errors = append(errors, ValidationError{
				Field: "temperature",
				Value: *settings.Temperature,
				Error: "temperature must be between 0 and 2",
			})
		}
	}

	if settings.TopP != nil {
		if *settings.TopP < 0 || *settings.TopP > 1 {
			errors = append(errors, ValidationError{
				Field: "top_p",
				Value: *settings.TopP,
				Error: "top_p must be between 0 and 1",
			})
		}
	}

	if settings.FrequencyPenalty != nil {
		if *settings.FrequencyPenalty < -2 || *settings.FrequencyPenalty > 2 {
			errors = append(errors, ValidationError{
				Field: "frequency_penalty",
				Value: *settings.FrequencyPenalty,
				Error: "frequency_penalty must be between -2 and 2",
			})
		}
	}

	if settings.PresencePenalty != nil {
		if *settings.PresencePenalty < -2 || *settings.PresencePenalty > 2 {
			errors = append(errors, ValidationError{
				Field: "presence_penalty",
				Value: *settings.PresencePenalty,
				Error: "presence_penalty must be between -2 and 2",
			})
		}
	}

	if settings.MaxTokens != nil {
		if *settings.MaxTokens <= 0 {
			errors = append(errors, ValidationError{
				Field: "max_tokens",
				Value: *settings.MaxTokens,
				Error: "max_tokens must be greater than 0",
			})
		}
	}

	return errors
}
