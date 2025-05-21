// ABOUTME: Tests for API key resolution from environment variables
// ABOUTME: Verifies config system properly handles API keys from different sources

package config

import (
	"os"
	"testing"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyResolution(t *testing.T) {
	// We don't need to initialize logging for tests

	// Setup test cases
	testCases := []struct {
		name            string
		envVars         map[string]string
		existingConfig  map[string]interface{}
		expectedKeys    map[string]string
		expectedDefault string
	}{
		{
			name: "OpenAI key from environment",
			envVars: map[string]string{
				EnvOpenAIKey: "test-openai-key",
			},
			existingConfig: map[string]interface{}{},
			expectedKeys: map[string]string{
				"provider.openai.api_key": "test-openai-key",
			},
			expectedDefault: "openai",
		},
		{
			name: "Anthropic key from environment",
			envVars: map[string]string{
				EnvAnthropicKey: "test-anthropic-key",
			},
			existingConfig: map[string]interface{}{},
			expectedKeys: map[string]string{
				"provider.anthropic.api_key": "test-anthropic-key",
			},
			expectedDefault: "anthropic",
		},
		{
			name: "Gemini key from environment",
			envVars: map[string]string{
				EnvGeminiKey: "test-gemini-key",
			},
			existingConfig: map[string]interface{}{},
			expectedKeys: map[string]string{
				"provider.gemini.api_key": "test-gemini-key",
			},
			expectedDefault: "gemini",
		},
		{
			name: "Multiple keys - OpenAI preferred",
			envVars: map[string]string{
				EnvOpenAIKey:    "test-openai-key",
				EnvAnthropicKey: "test-anthropic-key",
			},
			existingConfig: map[string]interface{}{},
			expectedKeys: map[string]string{
				"provider.openai.api_key":    "test-openai-key",
				"provider.anthropic.api_key": "test-anthropic-key",
			},
			expectedDefault: "openai",
		},
		{
			name: "Config file has precedence",
			envVars: map[string]string{
				EnvOpenAIKey: "test-openai-key",
			},
			existingConfig: map[string]interface{}{
				"provider": map[string]interface{}{
					"default": "anthropic",
					"openai": map[string]interface{}{
						"api_key": "config-openai-key",
					},
				},
			},
			expectedKeys: map[string]string{
				"provider.openai.api_key": "config-openai-key",
			},
			expectedDefault: "anthropic",
		},
		{
			name: "Environment fills in missing keys",
			envVars: map[string]string{
				EnvOpenAIKey: "test-openai-key",
			},
			existingConfig: map[string]interface{}{
				"provider": map[string]interface{}{
					"default": "anthropic",
					"anthropic": map[string]interface{}{
						"api_key": "config-anthropic-key",
					},
				},
			},
			expectedKeys: map[string]string{
				"provider.openai.api_key":    "test-openai-key",
				"provider.anthropic.api_key": "config-anthropic-key",
			},
			expectedDefault: "anthropic",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear environment variables first
			os.Unsetenv(EnvOpenAIKey)
			os.Unsetenv(EnvAnthropicKey)
			os.Unsetenv(EnvGeminiKey)

			// Set environment variables for this test case
			for key, value := range tc.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Create a new Config instance
			cfg := &Config{
				koanf:    koanf.New("."),
				defaults: GetCompleteDefaultConfig(),
			}

			// Load existing config if any
			if len(tc.existingConfig) > 0 {
				err := cfg.koanf.Load(confmap.Provider(tc.existingConfig, "."), nil)
				assert.NoError(t, err)
			}

			// Call loadProviderAPIKeys
			err := cfg.loadProviderAPIKeys()
			assert.NoError(t, err)

			// Check API keys
			for path, expectedValue := range tc.expectedKeys {
				value := cfg.koanf.String(path)
				assert.Equal(t, expectedValue, value, "Unexpected value for %s", path)
			}

			// Check default provider
			defaultProvider := cfg.koanf.String("provider.default")
			assert.Equal(t, tc.expectedDefault, defaultProvider)
		})
	}
}
