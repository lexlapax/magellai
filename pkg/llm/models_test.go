// ABOUTME: Unit tests for model registry and information functions
// ABOUTME: Tests GetAvailableModels, GetModelInfo, and model utilities

package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAvailableModels(t *testing.T) {
	models := GetAvailableModels()

	// Check we have models from all providers
	providers := make(map[string]bool)
	for _, model := range models {
		providers[model.Provider] = true
	}

	assert.True(t, providers[ProviderOpenAI], "Should have OpenAI models")
	assert.True(t, providers[ProviderAnthropic], "Should have Anthropic models")
	assert.True(t, providers[ProviderGemini], "Should have Gemini models")

	// Check some known models exist
	found := make(map[string]bool)
	for _, model := range models {
		key := model.Provider + "/" + model.Model
		found[key] = true
	}

	assert.True(t, found["openai/gpt-4"], "Should have GPT-4")
	assert.True(t, found["anthropic/claude-3-opus"], "Should have Claude 3 Opus")
	assert.True(t, found["gemini/pro"], "Should have Gemini Pro")
}

func TestGetModelInfo(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		model       string
		expectError bool
		checkInfo   func(t *testing.T, info ModelInfo)
	}{
		{
			name:        "valid OpenAI model",
			provider:    "openai",
			model:       "gpt-4",
			expectError: false,
			checkInfo: func(t *testing.T, info ModelInfo) {
				assert.Equal(t, "openai", info.Provider)
				assert.Equal(t, "gpt-4", info.Model)
				assert.Equal(t, "GPT-4", info.DisplayName)
				assert.True(t, info.Capabilities.Text)
				assert.False(t, info.Capabilities.Image)
			},
		},
		{
			name:        "valid Anthropic model",
			provider:    "anthropic",
			model:       "claude-3-opus",
			expectError: false,
			checkInfo: func(t *testing.T, info ModelInfo) {
				assert.Equal(t, "anthropic", info.Provider)
				assert.Equal(t, "claude-3-opus", info.Model)
				assert.Equal(t, "Claude 3 Opus", info.DisplayName)
				assert.True(t, info.Capabilities.Text)
				assert.True(t, info.Capabilities.Image)
			},
		},
		{
			name:        "valid multimodal model",
			provider:    "gemini",
			model:       "ultra",
			expectError: false,
			checkInfo: func(t *testing.T, info ModelInfo) {
				assert.Equal(t, "gemini", info.Provider)
				assert.Equal(t, "ultra", info.Model)
				assert.True(t, info.Capabilities.Text)
				assert.True(t, info.Capabilities.Image)
				assert.True(t, info.Capabilities.Audio)
				assert.True(t, info.Capabilities.Video)
			},
		},
		{
			name:        "case insensitive provider",
			provider:    "OpenAI",
			model:       "gpt-4",
			expectError: false,
			checkInfo: func(t *testing.T, info ModelInfo) {
				assert.Equal(t, ProviderOpenAI, info.Provider)
			},
		},
		{
			name:        "case insensitive model",
			provider:    "openai",
			model:       "GPT-4",
			expectError: false,
			checkInfo: func(t *testing.T, info ModelInfo) {
				assert.Equal(t, "gpt-4", info.Model)
			},
		},
		{
			name:        "invalid provider",
			provider:    "invalid",
			model:       "model",
			expectError: true,
		},
		{
			name:        "invalid model",
			provider:    "openai",
			model:       "invalid-model",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetModelInfo(tt.provider, tt.model)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkInfo != nil {
					tt.checkInfo(t, info)
				}
			}
		})
	}
}
