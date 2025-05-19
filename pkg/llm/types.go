// ABOUTME: Core types for wrapping go-llms library types using domain types
// ABOUTME: Provides adapter types that bridge between Magellai domain and go-llms types
package llm

import (
	"strings"

	"github.com/lexlapax/magellai/pkg/domain"
)

// Provider name constants
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderGemini    = "gemini"
	ProviderOllama    = "ollama"
	ProviderMock      = "mock"
)

// Model capability flags
type ModelCapability string

const (
	CapabilityText  ModelCapability = "text"
	CapabilityImage ModelCapability = "image"
	CapabilityAudio ModelCapability = "audio"
	CapabilityVideo ModelCapability = "video"
	CapabilityFile  ModelCapability = "file"
)

// Request uses domain.Message instead of custom Message type
type Request struct {
	Messages     []domain.Message `json:"messages"`
	Model        string           `json:"model,omitempty"` // provider/model format
	Temperature  *float64         `json:"temperature,omitempty"`
	MaxTokens    *int             `json:"max_tokens,omitempty"`
	Stream       bool             `json:"stream,omitempty"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
	Options      *PromptParams    `json:"options,omitempty"`
}

// Response wraps go-llms response types
type Response struct {
	Content      string                 `json:"content"`
	Model        string                 `json:"model,omitempty"`
	Usage        *Usage                 `json:"usage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	FinishReason string                 `json:"finish_reason,omitempty"`
}

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// PromptParams maps to go-llms domain.Option
type PromptParams struct {
	Temperature      *float64               `json:"temperature,omitempty"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	TopK             *int                   `json:"top_k,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	Seed             *int                   `json:"seed,omitempty"`
	ResponseFormat   string                 `json:"response_format,omitempty"`
	CustomOptions    map[string]interface{} `json:"custom_options,omitempty"`
}

// Conversion functions moved to adapters.go

// ParseModelString splits a provider/model string into components
func ParseModelString(model string) (provider, modelName string) {
	parts := strings.SplitN(model, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// Default to OpenAI if no provider specified
	return ProviderOpenAI, model
}

// FormatModelString combines provider and model into provider/model format
func FormatModelString(provider, model string) string {
	return provider + "/" + model
}
