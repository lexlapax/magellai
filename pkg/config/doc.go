// ABOUTME: Configuration management using koanf with multi-layer support
// ABOUTME: Handles config loading, merging, validation, and profile management

/*
Package config provides configuration management for Magellai.

This package handles configuration from multiple sources (files, environment variables,
defaults) with proper precedence rules. It supports validation, profiles, and
runtime configuration updates for both CLI and REPL environments.

Key Components:
  - Config: Main configuration object with layered sources
  - Schema: JSON Schema for configuration validation
  - Defaults: Default configuration values
  - Validation: Configuration validation against schema
  - Profiles: Support for multiple named configuration profiles
  - Error Handling: Standardized configuration error types

Configuration Layers (in order of precedence):
  1. Command-line flags
  2. Environment variables
  3. Profile-specific configuration
  4. Global configuration file
  5. Default values

Usage:
    // Load configuration
    cfg, err := config.Load(configPath)
    if err != nil {
        // Handle error
    }

    // Access configuration values
    modelName := cfg.GetString("model")
    apiKey := cfg.GetString("anthropic.api_key")

    // Set a configuration value
    err = cfg.SetValue("model", "claude-3-haiku-20240307")

    // Generate default configuration
    defaultConfig, err := config.GenerateDefaultConfig()

The package supports both static configuration for application startup
and dynamic configuration updates during runtime.
*/
package config