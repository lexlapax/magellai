// ABOUTME: Model registry and model information helpers
// ABOUTME: Provides functions to get available models and their capabilities

package llm

import (
	"fmt"
	"strings"
)

// ModelCapabilities represents what a model can process
type ModelCapabilities struct {
	Text             bool `json:"text"`
	Audio            bool `json:"audio"`
	Video            bool `json:"video"`
	Image            bool `json:"image"`
	File             bool `json:"file"`
	StructuredOutput bool `json:"structured_output"` // Can generate structured JSON output
}

// ModelInfo represents detailed information about a model
type ModelInfo struct {
	Provider           string            `json:"provider"`
	Model              string            `json:"model"`
	DisplayName        string            `json:"display_name"`
	Description        string            `json:"description"`
	Capabilities       ModelCapabilities `json:"capabilities"`
	MaxTokens          int               `json:"max_tokens,omitempty"`
	ContextWindow      int               `json:"context_window,omitempty"`
	DefaultTemperature float64           `json:"default_temperature,omitempty"`
}

// GetAvailableModels returns a list of available models across all providers
func GetAvailableModels() []ModelInfo {
	models := []ModelInfo{}

	// OpenAI models
	models = append(models,
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "gpt-4",
			DisplayName:   "GPT-4",
			Description:   "Most capable GPT-4 model for complex tasks",
			Capabilities:  ModelCapabilities{Text: true},
			MaxTokens:     4096,
			ContextWindow: 8192,
		},
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "gpt-4-turbo",
			DisplayName:   "GPT-4 Turbo",
			Description:   "Faster GPT-4 with vision capabilities",
			Capabilities:  ModelCapabilities{Text: true, Image: true},
			MaxTokens:     4096,
			ContextWindow: 128000,
		},
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "gpt-4o",
			DisplayName:   "GPT-4 Omni",
			Description:   "Multimodal GPT-4 with vision and audio",
			Capabilities:  ModelCapabilities{Text: true, Image: true, Audio: true},
			MaxTokens:     4096,
			ContextWindow: 128000,
		},
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "gpt-4-vision-preview",
			DisplayName:   "GPT-4 Vision",
			Description:   "GPT-4 with vision capabilities",
			Capabilities:  ModelCapabilities{Text: true, Image: true},
			MaxTokens:     4096,
			ContextWindow: 128000,
		},
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "gpt-3.5-turbo",
			DisplayName:   "GPT-3.5 Turbo",
			Description:   "Fast and cost-effective model",
			Capabilities:  ModelCapabilities{Text: true},
			MaxTokens:     4096,
			ContextWindow: 16384,
		},
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "whisper-1",
			DisplayName:   "Whisper",
			Description:   "Audio transcription model",
			Capabilities:  ModelCapabilities{Audio: true},
			MaxTokens:     0,
			ContextWindow: 0,
		},
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "tts-1",
			DisplayName:   "Text-to-Speech",
			Description:   "Standard text to speech model",
			Capabilities:  ModelCapabilities{Text: true},
			MaxTokens:     0,
			ContextWindow: 0,
		},
		ModelInfo{
			Provider:      ProviderOpenAI,
			Model:         "dall-e-3",
			DisplayName:   "DALL-E 3",
			Description:   "Advanced image generation model",
			Capabilities:  ModelCapabilities{Text: true},
			MaxTokens:     0,
			ContextWindow: 0,
		},
	)

	// Anthropic models
	models = append(models,
		ModelInfo{
			Provider:      ProviderAnthropic,
			Model:         "claude-3-opus",
			DisplayName:   "Claude 3 Opus",
			Description:   "Most powerful Claude model",
			Capabilities:  ModelCapabilities{Text: true, Image: true},
			MaxTokens:     4096,
			ContextWindow: 200000,
		},
		ModelInfo{
			Provider:      ProviderAnthropic,
			Model:         "claude-3-sonnet",
			DisplayName:   "Claude 3 Sonnet",
			Description:   "Balanced performance and cost",
			Capabilities:  ModelCapabilities{Text: true, Image: true},
			MaxTokens:     4096,
			ContextWindow: 200000,
		},
		ModelInfo{
			Provider:      ProviderAnthropic,
			Model:         "claude-3-haiku",
			DisplayName:   "Claude 3 Haiku",
			Description:   "Fast and cost-effective Claude model",
			Capabilities:  ModelCapabilities{Text: true, Image: true},
			MaxTokens:     4096,
			ContextWindow: 200000,
		},
		ModelInfo{
			Provider:      ProviderAnthropic,
			Model:         "claude-3-5-sonnet",
			DisplayName:   "Claude 3.5 Sonnet",
			Description:   "Latest and most capable Claude model",
			Capabilities:  ModelCapabilities{Text: true, Image: true},
			MaxTokens:     8192,
			ContextWindow: 200000,
		},
	)

	// Google Gemini models
	models = append(models,
		ModelInfo{
			Provider:      ProviderGemini,
			Model:         "pro",
			DisplayName:   "Gemini Pro",
			Description:   "Versatile model for various tasks",
			Capabilities:  ModelCapabilities{Text: true},
			MaxTokens:     2048,
			ContextWindow: 32768,
		},
		ModelInfo{
			Provider:      ProviderGemini,
			Model:         "pro-vision",
			DisplayName:   "Gemini Pro Vision",
			Description:   "Multimodal understanding of images and text",
			Capabilities:  ModelCapabilities{Text: true, Image: true},
			MaxTokens:     2048,
			ContextWindow: 32768,
		},
		ModelInfo{
			Provider:      ProviderGemini,
			Model:         "flash",
			DisplayName:   "Gemini Flash",
			Description:   "Fast and efficient model",
			Capabilities:  ModelCapabilities{Text: true},
			MaxTokens:     1024,
			ContextWindow: 32768,
		},
		ModelInfo{
			Provider:      ProviderGemini,
			Model:         "ultra",
			DisplayName:   "Gemini Ultra",
			Description:   "Most capable Gemini model",
			Capabilities:  ModelCapabilities{Text: true, Image: true, Audio: true, Video: true},
			MaxTokens:     2048,
			ContextWindow: 32768,
		},
		ModelInfo{
			Provider:      ProviderGemini,
			Model:         "flash-1.5",
			DisplayName:   "Gemini Flash 1.5",
			Description:   "High-volume, cost-effective model",
			Capabilities:  ModelCapabilities{Text: true, Image: true, Audio: true, Video: true},
			MaxTokens:     8192,
			ContextWindow: 1000000,
		},
		ModelInfo{
			Provider:      ProviderGemini,
			Model:         "pro-1.5",
			DisplayName:   "Gemini Pro 1.5",
			Description:   "Mid-size multimodal model",
			Capabilities:  ModelCapabilities{Text: true, Image: true, Audio: true, Video: true, File: true},
			MaxTokens:     8192,
			ContextWindow: 2000000,
		},
	)

	return models
}

// GetModelInfo returns information about a specific model
func GetModelInfo(provider, model string) (ModelInfo, error) {
	models := GetAvailableModels()

	for _, m := range models {
		if strings.EqualFold(m.Provider, provider) && strings.EqualFold(m.Model, model) {
			return m, nil
		}
	}

	return ModelInfo{}, fmt.Errorf("%w: %s/%s", ErrModelNotFound, provider, model)
}
