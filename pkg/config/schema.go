// ABOUTME: Configuration schema definitions and types
// ABOUTME: Provides type-safe configuration structure and parsing utilities

package config

import (
	"strings"
	"time"
)

// ProviderModelPair represents a provider/model combination
type ProviderModelPair struct {
	Provider string
	Model    string
}

// ParseProviderModel parses a provider/model string into separate components
func ParseProviderModel(s string) ProviderModelPair {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) == 2 {
		return ProviderModelPair{
			Provider: parts[0],
			Model:    parts[1],
		}
	}
	// If no slash, assume it's just a model for the default provider
	return ProviderModelPair{
		Provider: "",
		Model:    s,
	}
}

// Schema represents the complete configuration structure
type Schema struct {
	Log      LogConfig                `koanf:"log"`
	Provider ProviderConfig           `koanf:"provider"`
	Model    ModelConfig              `koanf:"model"`
	Output   OutputConfig             `koanf:"output"`
	Session  SessionConfig            `koanf:"session"`
	Plugin   PluginConfig             `koanf:"plugin"`
	Profiles map[string]ProfileConfig `koanf:"profiles"`
	Aliases  map[string]string        `koanf:"aliases"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
}

// ProviderConfig represents provider configuration
type ProviderConfig struct {
	Default   string           `koanf:"default"`
	OpenAI    *OpenAIConfig    `koanf:"openai"`
	Anthropic *AnthropicConfig `koanf:"anthropic"`
	Gemini    *GeminiConfig    `koanf:"gemini"`
	// Generic map for provider-specific settings
	Settings map[string]map[string]interface{} `koanf:"settings"`
}

// ModelConfig represents model-specific configuration
type ModelConfig struct {
	// Default model in provider/model format
	Default string `koanf:"default"`
	// Model-specific settings keyed by provider/model
	Settings map[string]ModelSettings `koanf:"settings"`
}

// ModelSettings represents model-specific parameters
type ModelSettings struct {
	Temperature      *float64 `koanf:"temperature"`
	MaxTokens        *int     `koanf:"max_tokens"`
	TopP             *float64 `koanf:"top_p"`
	FrequencyPenalty *float64 `koanf:"frequency_penalty"`
	PresencePenalty  *float64 `koanf:"presence_penalty"`
	StopSequences    []string `koanf:"stop_sequences"`
	// Provider-specific settings
	Extra map[string]interface{} `koanf:"extra"`
}

// OutputConfig represents output preferences
type OutputConfig struct {
	Format string `koanf:"format"` // text, json, markdown
	Color  bool   `koanf:"color"`
	Pretty bool   `koanf:"pretty"`
}

// SessionConfig represents session storage settings
type SessionConfig struct {
	Directory   string        `koanf:"directory"`
	AutoSave    bool          `koanf:"autosave"`
	MaxAge      time.Duration `koanf:"max_age"`
	Compression bool          `koanf:"compression"`
	Storage     StorageConfig `koanf:"storage"`
}

// StorageConfig represents storage backend configuration
type StorageConfig struct {
	Type     string                 `koanf:"type"`     // filesystem, sqlite, postgresql, etc.
	Settings map[string]interface{} `koanf:"settings"` // Backend-specific settings
}

// PluginConfig represents plugin configuration
type PluginConfig struct {
	Directory string   `koanf:"directory"`
	Path      []string `koanf:"path"`
	Enabled   []string `koanf:"enabled"`
	Disabled  []string `koanf:"disabled"`
}

// ProfileConfig represents a configuration profile
type ProfileConfig struct {
	Description string                 `koanf:"description"`
	Provider    string                 `koanf:"provider"`
	Model       string                 `koanf:"model"`
	Settings    map[string]interface{} `koanf:"settings"`
}

// Provider-specific configurations

// OpenAIConfig represents OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey       string        `koanf:"api_key"`
	BaseURL      string        `koanf:"base_url"`
	Organization string        `koanf:"organization"`
	APIVersion   string        `koanf:"api_version"`
	DefaultModel string        `koanf:"default_model"`
	Timeout      time.Duration `koanf:"timeout"`
	MaxRetries   int           `koanf:"max_retries"`
}

// AnthropicConfig represents Anthropic-specific configuration
type AnthropicConfig struct {
	APIKey       string        `koanf:"api_key"`
	BaseURL      string        `koanf:"base_url"`
	APIVersion   string        `koanf:"api_version"`
	DefaultModel string        `koanf:"default_model"`
	Timeout      time.Duration `koanf:"timeout"`
	MaxRetries   int           `koanf:"max_retries"`
}

// GeminiConfig represents Google Gemini-specific configuration
type GeminiConfig struct {
	APIKey       string        `koanf:"api_key"`
	BaseURL      string        `koanf:"base_url"`
	ProjectID    string        `koanf:"project_id"`
	Location     string        `koanf:"location"`
	DefaultModel string        `koanf:"default_model"`
	Timeout      time.Duration `koanf:"timeout"`
	MaxRetries   int           `koanf:"max_retries"`
}
