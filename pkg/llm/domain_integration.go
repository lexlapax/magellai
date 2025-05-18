// ABOUTME: Integration point for LLM package to work with domain types
// ABOUTME: Provides factory functions and convenience methods for domain-based interactions

package llm

import (
	"context"
	"fmt"

	"github.com/lexlapax/magellai/pkg/domain"
)

// CreateDomainProvider creates a new domain-aware provider
func CreateDomainProvider(providerType, model string, apiKey ...string) (DomainProvider, error) {
	// Create the base provider
	baseProvider, err := NewProvider(providerType, model, apiKey...)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	
	// Wrap with domain support
	return NewDomainProvider(baseProvider), nil
}

// ConvertModelInfoToDomain converts LLM ModelInfo to domain Model
func ConvertModelInfoToDomain(info ModelInfo) *domain.Model {
	// Convert capabilities
	var capabilities []domain.ModelCapability
	if info.Capabilities.Text {
		capabilities = append(capabilities, domain.ModelCapabilityStructuredOutput)
	}
	if info.Capabilities.Image {
		capabilities = append(capabilities, domain.ModelCapabilityVision)
	}
	if info.Capabilities.Audio {
		capabilities = append(capabilities, domain.ModelCapabilityAudio)
	}
	if info.Capabilities.Video {
		capabilities = append(capabilities, domain.ModelCapabilityVideo)
	}
	if info.Capabilities.File {
		capabilities = append(capabilities, domain.ModelCapabilityFileAttachments)
	}

	// Convert modalities
	var modalities []string
	if info.Capabilities.Text {
		modalities = append(modalities, string(domain.ModalityText))
	}
	if info.Capabilities.Image {
		modalities = append(modalities, string(domain.ModalityImage))
	}
	if info.Capabilities.Audio {
		modalities = append(modalities, string(domain.ModalityAudio))
	}
	if info.Capabilities.Video {
		modalities = append(modalities, string(domain.ModalityVideo))
	}
	if info.Capabilities.File {
		modalities = append(modalities, string(domain.ModalityFile))
	}

	return &domain.Model{
		ID:              fmt.Sprintf("%s/%s", info.Provider, info.Model),
		Name:            info.Model,
		DisplayName:     info.DisplayName,
		Provider:        info.Provider,
		ContextWindow:   info.ContextWindow,
		MaxOutputTokens: info.MaxTokens,
		Capabilities:    capabilities,
		Modalities:      modalities,
		Metadata: map[string]interface{}{
			"description":         info.Description,
			"default_temperature": info.DefaultTemperature,
		},
	}
}

// ConvertProviderToDomain converts LLM provider string to domain Provider
func ConvertProviderToDomain(provider string) *domain.Provider {
	displayName := provider
	switch provider {
	case ProviderOpenAI:
		displayName = "OpenAI"
	case ProviderAnthropic:
		displayName = "Anthropic"
	case ProviderGemini:
		displayName = "Google Gemini"
	case ProviderOllama:
		displayName = "Ollama"
	case ProviderMock:
		displayName = "Mock Provider"
	}

	return &domain.Provider{
		Name:        provider,
		DisplayName: displayName,
		Capabilities: []string{"text", "streaming"},
		Metadata: map[string]interface{}{
			"source": "llm-package",
		},
	}
}

// GenerateWithDomainMessages is a convenience function for one-off generation with domain messages
func GenerateWithDomainMessages(ctx context.Context, providerType, model string, messages []*domain.Message, options ...ProviderOption) (*domain.Message, error) {
	// Get API key from environment if not provided
	apiKey := getAPIKeyFromEnv(providerType)
	if apiKey == "" {
		return nil, fmt.Errorf("no API key found for provider %s", providerType)
	}

	// Create domain provider
	provider, err := CreateDomainProvider(providerType, model, apiKey)
	if err != nil {
		return nil, err
	}

	// Generate response
	return provider.GenerateDomainMessage(ctx, messages, options...)
}

// StreamWithDomainMessages is a convenience function for streaming with domain messages  
func StreamWithDomainMessages(ctx context.Context, providerType, model string, messages []*domain.Message, options ...ProviderOption) (<-chan *domain.Message, error) {
	// Get API key from environment if not provided
	apiKey := getAPIKeyFromEnv(providerType)
	if apiKey == "" {
		return nil, fmt.Errorf("no API key found for provider %s", providerType)
	}

	// Create domain provider
	provider, err := CreateDomainProvider(providerType, model, apiKey)
	if err != nil {
		return nil, err
	}

	// Stream response
	return provider.StreamDomainMessage(ctx, messages, options...)
}