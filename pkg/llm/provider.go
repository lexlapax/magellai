// ABOUTME: Provider adapter interface that wraps go-llms providers using domain types
// ABOUTME: Provides factory methods, streaming capabilities, and provider configuration options
package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/domain"
)

// SanitizeAPIKey creates a sanitized version of an API key for logging
func SanitizeAPIKey(key string) string {
	if len(key) <= 3 {
		return "***"
	}
	if len(key) <= 10 {
		return key[:2] + "..." + key[len(key)-2:]
	}
	// For API keys like "sk-ant-...", show more of the prefix
	if strings.HasPrefix(key, "sk-ant-") && len(key) > 10 {
		return key[:6] + "..." + key[len(key)-4:]
	}
	// For longer keys, show first 3-6 and last 4
	if len(key) > 20 {
		return key[:6] + "..." + key[len(key)-4:]
	}
	return key[:3] + "..." + key[len(key)-4:]
}

// Provider is our adapter interface that wraps go-llms domain.Provider
type Provider interface {
	// Generate produces text from a prompt
	Generate(ctx context.Context, prompt string, options ...ProviderOption) (string, error)

	// GenerateMessage produces a response from messages
	GenerateMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (*Response, error)

	// GenerateWithSchema produces structured output conforming to a schema
	GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...ProviderOption) (interface{}, error)

	// Stream streams responses token by token
	Stream(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error)

	// StreamMessage streams responses from messages
	StreamMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (<-chan StreamChunk, error)

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

// providerAdapter wraps a go-llms provider
type providerAdapter struct {
	provider llmdomain.Provider
	name     string
	model    string
	config   *ProviderConfig
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	APIKey         string
	BaseURL        string
	OrgID          string
	DefaultModel   string
	DefaultOptions *PromptParams
}

// Note: ModelInfo is defined in models.go

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content      string
	Error        error
	Done         bool
	FinishReason string
	Index        int
}

// Ensure providerAdapter implements Provider
var _ Provider = (*providerAdapter)(nil)

// NewProvider creates a new provider instance
func NewProvider(providerType, model string, apiKey ...string) (Provider, error) {
	logging.LogInfo("Creating new provider", "type", providerType, "model", model)

	// Check for API key
	key := ""
	if len(apiKey) > 0 {
		key = apiKey[0]
	}

	// If key is empty, try to get it from environment variables
	if key == "" {
		key = getAPIKeyFromEnv(providerType)
		if key != "" {
			logging.LogInfo("Using API key from environment variable", "provider", providerType)
		}
	}

	// For non-mock providers, verify that we have an API key
	if key == "" && providerType != ProviderMock {
		envVarName := getEnvVarNameForProvider(providerType)
		return nil, fmt.Errorf("API key required for provider %s. Set %s environment variable or provide it in configuration",
			providerType, envVarName)
	}

	// Create underlying go-llms provider
	var llmProvider llmdomain.Provider
	var err error

	switch providerType {
	case ProviderOpenAI:
		llmProvider = provider.NewOpenAIProvider(key, model)
	case ProviderAnthropic:
		llmProvider = provider.NewAnthropicProvider(key, model)
	case ProviderGemini:
		llmProvider = provider.NewGeminiProvider(key, model)
	case ProviderMock:
		llmProvider = provider.NewMockProvider()
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	return &providerAdapter{
		provider: llmProvider,
		name:     providerType,
		model:    model,
		config:   &ProviderConfig{APIKey: key},
	}, err
}

// Generate produces text from a prompt
func (p *providerAdapter) Generate(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
	// Convert prompt to message
	msg := domain.NewMessage("", domain.MessageRoleUser, prompt)
	messages := []domain.Message{*msg}

	resp, err := p.GenerateMessage(ctx, messages, options...)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateMessage produces a response from messages
func (p *providerAdapter) GenerateMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (*Response, error) {
	logging.LogDebug("Generating message", "provider", p.name, "model", p.model, "messageCount", len(messages))

	// Apply options
	config := &providerConfig{}
	for _, opt := range options {
		opt(config)
	}

	// Convert domain messages to LLM messages
	llmMessages := ToLLMMessages(messages)

	// Create LLM options
	llmOptions := buildLLMOptions(config)

	// Generate response
	llmResp, err := p.provider.GenerateMessage(ctx, llmMessages, llmOptions...)
	if err != nil {
		logging.LogError(err, "Failed to generate message", "provider", p.name)
		return nil, err
	}

	// Convert response
	return convertLLMResponse(&llmResp), nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *providerAdapter) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...ProviderOption) (interface{}, error) {
	// Apply options
	config := &providerConfig{}
	for _, opt := range options {
		opt(config)
	}

	// Build options
	llmOptions := buildLLMOptions(config)

	// Generate with schema
	result, err := p.provider.GenerateWithSchema(ctx, prompt, schema, llmOptions...)
	if err != nil {
		logging.LogError(err, "Failed to generate with schema")
		return nil, err
	}

	return result, nil
}

// Stream streams responses token by token
func (p *providerAdapter) Stream(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
	// Convert prompt to message
	msg := domain.NewMessage("", domain.MessageRoleUser, prompt)
	messages := []domain.Message{*msg}

	return p.StreamMessage(ctx, messages, options...)
}

// StreamMessage streams responses from messages
func (p *providerAdapter) StreamMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (<-chan StreamChunk, error) {
	// Apply options
	config := &providerConfig{}
	for _, opt := range options {
		opt(config)
	}

	// Convert to LLM messages
	llmMessages := ToLLMMessages(messages)

	// Build options
	llmOptions := buildLLMOptions(config)

	// Create stream
	llmStream, err := p.provider.StreamMessage(ctx, llmMessages, llmOptions...)
	if err != nil {
		return nil, err
	}

	// Convert stream
	outStream := make(chan StreamChunk)
	go func() {
		defer close(outStream)
		for chunk := range llmStream {
			outStream <- StreamChunk{
				Content: chunk.Text,
				Done:    chunk.Finished,
			}
		}
	}()

	return outStream, nil
}

// GetModelInfo returns information about the current model
func (p *providerAdapter) GetModelInfo() ModelInfo {
	// TODO: Get actual capabilities from model registry
	return ModelInfo{
		Provider: p.name,
		Model:    p.model,
		Capabilities: ModelCapabilities{
			Text:  true,
			Image: p.name == ProviderOpenAI || p.name == ProviderAnthropic || p.name == ProviderGemini,
			Audio: p.name == ProviderOpenAI || p.name == ProviderGemini,
			Video: p.name == ProviderGemini,
			File:  p.name == ProviderOpenAI || p.name == ProviderAnthropic,
		},
		MaxTokens:     4096,   // Default, should come from model registry
		ContextWindow: 128000, // Default context window
	}
}

// Helper functions

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

// getEnvVarNameForProvider returns the environment variable name for a provider
func getEnvVarNameForProvider(provider string) string {
	switch provider {
	case ProviderOpenAI:
		return "OPENAI_API_KEY"
	case ProviderAnthropic:
		return "ANTHROPIC_API_KEY"
	case ProviderGemini:
		return "GEMINI_API_KEY"
	default:
		return "API_KEY"
	}
}

func buildLLMOptions(config *providerConfig) []llmdomain.Option {
	var options []llmdomain.Option

	if config.temperature != nil {
		options = append(options, llmdomain.WithTemperature(*config.temperature))
	}
	if config.maxTokens != nil {
		options = append(options, llmdomain.WithMaxTokens(*config.maxTokens))
	}
	if len(config.stopSequences) > 0 {
		options = append(options, llmdomain.WithStopSequences(config.stopSequences))
	}
	if config.topP != nil {
		options = append(options, llmdomain.WithTopP(*config.topP))
	}
	// Note: topK, seed, and responseFormat are not supported in go-llms yet
	// We'll need to handle these at the provider level later
	if config.presencePenalty != nil {
		options = append(options, llmdomain.WithPresencePenalty(*config.presencePenalty))
	}
	if config.frequencyPenalty != nil {
		options = append(options, llmdomain.WithFrequencyPenalty(*config.frequencyPenalty))
	}

	return options
}

func convertLLMResponse(resp *llmdomain.Response) *Response {
	return &Response{
		Content:      resp.Content,
		FinishReason: "", // Not available in go-llms Response
		Usage: &Usage{
			InputTokens:  0, // Not available in go-llms Response
			OutputTokens: 0, // Not available in go-llms Response
			TotalTokens:  0, // Not available in go-llms Response
		},
		Metadata: make(map[string]interface{}),
	}
}

// Provider Options

// WithTemperature sets the temperature for generation
func WithTemperature(temp float64) ProviderOption {
	return func(c *providerConfig) {
		c.temperature = &temp
	}
}

// WithMaxTokens sets the maximum tokens for generation
func WithMaxTokens(tokens int) ProviderOption {
	return func(c *providerConfig) {
		c.maxTokens = &tokens
	}
}

// WithStopSequences sets the stop sequences
func WithStopSequences(sequences ...string) ProviderOption {
	return func(c *providerConfig) {
		c.stopSequences = sequences
	}
}

// WithTopP sets the top-p value
func WithTopP(topP float64) ProviderOption {
	return func(c *providerConfig) {
		c.topP = &topP
	}
}

// WithTopK sets the top-k value
func WithTopK(topK int) ProviderOption {
	return func(c *providerConfig) {
		c.topK = &topK
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

// WithSeed sets the random seed
func WithSeed(seed int) ProviderOption {
	return func(c *providerConfig) {
		c.seed = &seed
	}
}

// WithResponseFormat sets the response format
func WithResponseFormat(format string) ProviderOption {
	return func(c *providerConfig) {
		c.responseFormat = format
	}
}
