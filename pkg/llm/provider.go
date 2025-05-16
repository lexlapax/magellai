// ABOUTME: Provider adapter interface that wraps go-llms providers
// ABOUTME: Provides factory methods and configuration helpers for LLM providers
package llm

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Provider is our adapter interface that wraps go-llms domain.Provider
type Provider interface {
	// Generate produces text from a prompt
	Generate(ctx context.Context, prompt string, options ...ProviderOption) (string, error)

	// GenerateMessage produces a response from messages
	GenerateMessage(ctx context.Context, messages []Message, options ...ProviderOption) (*Response, error)

	// GenerateWithSchema produces structured output conforming to a schema
	GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...ProviderOption) (interface{}, error)

	// Stream streams responses token by token
	Stream(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error)

	// StreamMessage streams responses from messages
	StreamMessage(ctx context.Context, messages []Message, options ...ProviderOption) (<-chan StreamChunk, error)

	// GetModelInfo returns information about the current model
	GetModelInfo() ModelInfo
}

// ProviderOption configures provider behavior
type ProviderOption func(*providerConfig)

// providerConfig holds configuration for provider operations
type providerConfig struct {
	temperature      *float64
	maxTokens        *int
	stopSequences    []string
	topP             *float64
	topK             *int
	presencePenalty  *float64
	frequencyPenalty *float64
	seed             *int
	responseFormat   string
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content      string
	Index        int
	FinishReason string
	Error        error
}

// providerAdapter adapts a go-llms Provider to our interface
type providerAdapter struct {
	provider  domain.Provider
	modelInfo ModelInfo
}

// NewProvider creates a provider adapter for the specified provider type
func NewProvider(providerType, model string, apiKey ...string) (Provider, error) {
	var llmProvider domain.Provider
	var err error

	// Use provided API key or fall back to environment variable
	key := ""
	if len(apiKey) > 0 {
		key = apiKey[0]
	} else {
		key = getAPIKeyFromEnv(providerType)
	}

	// Mock provider doesn't need an API key
	if providerType != ProviderMock && key == "" {
		return nil, fmt.Errorf("API key not provided for %s", providerType)
	}

	// Create the underlying go-llms provider
	switch providerType {
	case ProviderOpenAI:
		llmProvider = provider.NewOpenAIProvider(key, model)
		err = nil
	case ProviderAnthropic:
		llmProvider = provider.NewAnthropicProvider(key, model)
		err = nil
	case ProviderGemini:
		llmProvider = provider.NewGeminiProvider(key, model)
		err = nil
	case ProviderMock:
		llmProvider = provider.NewMockProvider()
		err = nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Get model info from registry
	modelInfo, err := GetModelInfo(providerType, model)
	if err != nil {
		// If model not found in registry, create a basic one
		modelInfo = ModelInfo{
			Provider:     providerType,
			Model:        model,
			Capabilities: ModelCapabilities{Text: true}, // Default to text-only
			Description:  fmt.Sprintf("%s model %s", providerType, model),
		}
	}

	return &providerAdapter{
		provider:  llmProvider,
		modelInfo: modelInfo,
	}, nil
}

// getAPIKeyFromEnv retrieves API key from environment variables
func getAPIKeyFromEnv(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return os.Getenv("OPENAI_API_KEY")
	case ProviderAnthropic:
		return os.Getenv("ANTHROPIC_API_KEY")
	case ProviderGemini:
		return os.Getenv("GEMINI_API_KEY")
	default:
		return ""
	}
}

// getModelCapabilities returns capabilities based on known models
// In a real implementation, this would query the provider or use a config
func getModelCapabilities(provider, model string) []ModelCapability {
	// Default capabilities for text models
	capabilities := []ModelCapability{CapabilityText}

	// Add multimodal capabilities for known models
	switch provider {
	case ProviderOpenAI:
		if contains([]string{"gpt-4-vision", "gpt-4o", "gpt-4o-mini"}, model) {
			capabilities = append(capabilities, CapabilityImage)
		}
	case ProviderGemini:
		// Most Gemini models support images
		capabilities = append(capabilities, CapabilityImage)
		// Some support video and audio
		if contains([]string{"gemini-pro-vision", "gemini-1.5-pro"}, model) {
			capabilities = append(capabilities, CapabilityVideo, CapabilityAudio)
		}
	case ProviderAnthropic:
		// Claude 3+ models support images
		if contains([]string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku", "claude-3.5-sonnet"}, model) {
			capabilities = append(capabilities, CapabilityImage)
		}
	}

	return capabilities
}

// Generate produces text from a prompt
func (p *providerAdapter) Generate(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
	llmOptions := p.buildLLMOptions(options...)
	return p.provider.Generate(ctx, prompt, llmOptions...)
}

// GenerateMessage produces a response from messages
func (p *providerAdapter) GenerateMessage(ctx context.Context, messages []Message, options ...ProviderOption) (*Response, error) {
	// Convert our messages to go-llms messages
	llmMessages := make([]domain.Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = msg.ToLLMMessage()
	}

	llmOptions := p.buildLLMOptions(options...)
	llmResponse, err := p.provider.GenerateMessage(ctx, llmMessages, llmOptions...)
	if err != nil {
		return nil, err
	}

	// Convert go-llms response to our response type
	return &Response{
		Content: llmResponse.Content,
		Model:   p.modelInfo.Model,
		Usage:   nil, // go-llms doesn't provide usage info in the basic Response
	}, nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *providerAdapter) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...ProviderOption) (interface{}, error) {
	llmOptions := p.buildLLMOptions(options...)
	return p.provider.GenerateWithSchema(ctx, prompt, schema, llmOptions...)
}

// Stream streams responses token by token
func (p *providerAdapter) Stream(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
	llmOptions := p.buildLLMOptions(options...)
	llmStream, err := p.provider.Stream(ctx, prompt, llmOptions...)
	if err != nil {
		return nil, err
	}

	// Convert go-llms stream to our stream format
	chunkChan := make(chan StreamChunk)
	go func() {
		defer close(chunkChan)
		for token := range llmStream {
			chunk := StreamChunk{
				Content: token.Text,
			}
			if token.Finished {
				chunk.FinishReason = "stop"
			}
			select {
			case chunkChan <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunkChan, nil
}

// StreamMessage streams responses from messages
func (p *providerAdapter) StreamMessage(ctx context.Context, messages []Message, options ...ProviderOption) (<-chan StreamChunk, error) {
	// Convert our messages to go-llms messages
	llmMessages := make([]domain.Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = msg.ToLLMMessage()
	}

	llmOptions := p.buildLLMOptions(options...)
	llmStream, err := p.provider.StreamMessage(ctx, llmMessages, llmOptions...)
	if err != nil {
		return nil, err
	}

	// Convert go-llms stream to our stream format
	chunkChan := make(chan StreamChunk)
	go func() {
		defer close(chunkChan)
		for token := range llmStream {
			chunk := StreamChunk{
				Content: token.Text,
			}
			if token.Finished {
				chunk.FinishReason = "stop"
			}
			select {
			case chunkChan <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunkChan, nil
}

// GetModelInfo returns information about the current model
func (p *providerAdapter) GetModelInfo() ModelInfo {
	return p.modelInfo
}

// buildLLMOptions converts our options to go-llms options
func (p *providerAdapter) buildLLMOptions(options ...ProviderOption) []domain.Option {
	config := &providerConfig{}
	for _, opt := range options {
		opt(config)
	}

	var llmOptions []domain.Option

	if config.temperature != nil {
		llmOptions = append(llmOptions, domain.WithTemperature(*config.temperature))
	}
	if config.maxTokens != nil {
		llmOptions = append(llmOptions, domain.WithMaxTokens(*config.maxTokens))
	}
	if len(config.stopSequences) > 0 {
		llmOptions = append(llmOptions, domain.WithStopSequences(config.stopSequences))
	}
	if config.topP != nil {
		llmOptions = append(llmOptions, domain.WithTopP(*config.topP))
	}
	if config.presencePenalty != nil {
		llmOptions = append(llmOptions, domain.WithPresencePenalty(*config.presencePenalty))
	}
	if config.frequencyPenalty != nil {
		llmOptions = append(llmOptions, domain.WithFrequencyPenalty(*config.frequencyPenalty))
	}

	return llmOptions
}

// Provider option functions

// WithTemperature sets the temperature for generation
func WithTemperature(temp float64) ProviderOption {
	return func(c *providerConfig) {
		c.temperature = &temp
	}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(tokens int) ProviderOption {
	return func(c *providerConfig) {
		c.maxTokens = &tokens
	}
}

// WithStopSequences sets sequences that stop generation
func WithStopSequences(sequences []string) ProviderOption {
	return func(c *providerConfig) {
		c.stopSequences = sequences
	}
}

// WithTopP sets the nucleus sampling probability
func WithTopP(p float64) ProviderOption {
	return func(c *providerConfig) {
		c.topP = &p
	}
}

// WithTopK sets the top-k sampling parameter
func WithTopK(k int) ProviderOption {
	return func(c *providerConfig) {
		c.topK = &k
	}
}

// WithPresencePenalty sets the presence penalty
func WithPresencePenalty(penalty float64) ProviderOption {
	return func(c *providerConfig) {
		c.presencePenalty = &penalty
	}
}

// WithFrequencyPenalty sets the frequency penalty
func WithFrequencyPenalty(penalty float64) ProviderOption {
	return func(c *providerConfig) {
		c.frequencyPenalty = &penalty
	}
}

// WithSeed sets the random seed for generation
func WithSeed(seed int) ProviderOption {
	return func(c *providerConfig) {
		c.seed = &seed
	}
}

// WithResponseFormat sets the response format (e.g., "json_object")
func WithResponseFormat(format string) ProviderOption {
	return func(c *providerConfig) {
		c.responseFormat = format
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Common errors
var (
	ErrNoAPIKey        = errors.New("no API key provided")
	ErrInvalidProvider = errors.New("invalid provider")
	ErrInvalidModel    = errors.New("invalid model")
)
