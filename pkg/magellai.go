// ABOUTME: High-level Ask function for one-shot LLM queries
// ABOUTME: Provides simple interface for prompting LLMs with automatic provider selection
package magellai

import (
	"context"
	"fmt"
	"strings"

	"github.com/lexlapax/magellai/pkg/llm"
)

// AskOptions configures the Ask function behavior
type AskOptions struct {
	// Model in provider/model format (e.g., "openai/gpt-4")
	Model string
	// Temperature controls randomness (0-1)
	Temperature *float64
	// MaxTokens limits response length
	MaxTokens *int
	// Stream enables streaming responses
	Stream bool
	// SystemPrompt sets the system message
	SystemPrompt string
	// ResponseFormat specifies the desired output format
	ResponseFormat string
	// Provider-specific options
	ProviderOptions []llm.ProviderOption
}

// AskResult contains the response from the LLM
type AskResult struct {
	Content      string
	Model        string
	Provider     string
	TokensUsed   *llm.Usage
	FinishReason string
}

// Ask sends a prompt to an LLM and returns the response
func Ask(ctx context.Context, prompt string, options *AskOptions) (*AskResult, error) {
	if prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	// Use default options if none provided
	if options == nil {
		options = &AskOptions{}
	}

	// Parse provider and model from the model string
	provider, model := parseModel(options.Model)

	// Create the LLM provider
	llmProvider, err := llm.NewProvider(provider, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Build provider options
	providerOpts := buildProviderOptions(options)

	// Create messages for the LLM
	messages := []llm.Message{}

	// Add system prompt if provided
	if options.SystemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: options.SystemPrompt,
		})
	}

	// Add user prompt
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: prompt,
	})

	// Handle streaming vs non-streaming
	if options.Stream {
		return askStream(ctx, llmProvider, messages, providerOpts)
	}

	return askNonStream(ctx, llmProvider, messages, providerOpts)
}

// askNonStream handles non-streaming requests
func askNonStream(ctx context.Context, provider llm.Provider, messages []llm.Message, opts []llm.ProviderOption) (*AskResult, error) {
	response, err := provider.GenerateMessage(ctx, messages, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	modelInfo := provider.GetModelInfo()

	return &AskResult{
		Content:      response.Content,
		Model:        modelInfo.Model,
		Provider:     modelInfo.Provider,
		TokensUsed:   response.Usage,
		FinishReason: response.FinishReason,
	}, nil
}

// askStream handles streaming requests
func askStream(ctx context.Context, provider llm.Provider, messages []llm.Message, opts []llm.ProviderOption) (*AskResult, error) {
	stream, err := provider.StreamMessage(ctx, messages, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	var content strings.Builder
	var finishReason string

	for chunk := range stream {
		if chunk.Error != nil {
			return nil, fmt.Errorf("streaming error: %w", chunk.Error)
		}
		content.WriteString(chunk.Content)
		if chunk.FinishReason != "" {
			finishReason = chunk.FinishReason
		}
	}

	modelInfo := provider.GetModelInfo()

	return &AskResult{
		Content:      content.String(),
		Model:        modelInfo.Model,
		Provider:     modelInfo.Provider,
		FinishReason: finishReason,
	}, nil
}

// parseModel splits a model string into provider and model parts
func parseModel(modelStr string) (provider, model string) {
	if modelStr == "" {
		// Default to OpenAI GPT-3.5 if no model specified
		return llm.ProviderOpenAI, "gpt-3.5-turbo"
	}

	return llm.ParseModelString(modelStr)
}

// buildProviderOptions converts AskOptions to provider options
func buildProviderOptions(opts *AskOptions) []llm.ProviderOption {
	var providerOpts []llm.ProviderOption

	if opts.Temperature != nil {
		providerOpts = append(providerOpts, llm.WithTemperature(*opts.Temperature))
	}

	if opts.MaxTokens != nil {
		providerOpts = append(providerOpts, llm.WithMaxTokens(*opts.MaxTokens))
	}

	if opts.ResponseFormat != "" {
		providerOpts = append(providerOpts, llm.WithResponseFormat(opts.ResponseFormat))
	}

	// Add any additional provider-specific options
	providerOpts = append(providerOpts, opts.ProviderOptions...)

	return providerOpts
}

// AskWithAttachments sends a prompt with multimodal attachments to an LLM
func AskWithAttachments(ctx context.Context, prompt string, attachments []llm.Attachment, options *AskOptions) (*AskResult, error) {
	if prompt == "" && len(attachments) == 0 {
		return nil, fmt.Errorf("prompt and attachments cannot both be empty")
	}

	// Use default options if none provided
	if options == nil {
		options = &AskOptions{}
	}

	// Parse provider and model from the model string
	provider, model := parseModel(options.Model)

	// Create the LLM provider
	llmProvider, err := llm.NewProvider(provider, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Build provider options
	providerOpts := buildProviderOptions(options)

	// Create messages for the LLM
	messages := []llm.Message{}

	// Add system prompt if provided
	if options.SystemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: options.SystemPrompt,
		})
	}

	// Add user prompt with attachments
	messages = append(messages, llm.Message{
		Role:        "user",
		Content:     prompt,
		Attachments: attachments,
	})

	// Handle streaming vs non-streaming
	if options.Stream {
		return askStream(ctx, llmProvider, messages, providerOpts)
	}

	return askNonStream(ctx, llmProvider, messages, providerOpts)
}
