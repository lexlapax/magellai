// ABOUTME: Ask command implementation for one-shot LLM queries
// ABOUTME: Supports multimodal attachments, streaming, and provider selection
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// AskCommand implements the ask command for one-shot queries
type AskCommand struct {
	config *config.Config
}

// NewAskCommand creates a new ask command instance
func NewAskCommand(cfg *config.Config) *AskCommand {
	return &AskCommand{
		config: cfg,
	}
}

// Metadata returns the command metadata
func (c *AskCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:            "ask",
		Description:     "Send a one-shot query to the LLM",
		Category:        command.CategoryCLI,
		LongDescription: "Send a one-shot query to the LLM with optional attachments and streaming",
		Flags: []command.Flag{
			{
				Name:        "model",
				Short:       "m",
				Type:        command.FlagTypeString,
				Description: "Model to use (provider/model format)",
			},
			{
				Name:        "attach",
				Short:       "a",
				Type:        command.FlagTypeStringSlice,
				Description: "Files to attach (can be used multiple times)",
			},
			{
				Name:        "stream",
				Type:        command.FlagTypeBool,
				Description: "Enable streaming responses",
			},
			{
				Name:        "temperature",
				Short:       "t",
				Type:        command.FlagTypeFloat,
				Description: "Temperature setting (0-1)",
			},
			{
				Name:        "max-tokens",
				Type:        command.FlagTypeInt,
				Description: "Maximum tokens in response",
			},
			{
				Name:        "system",
				Short:       "s",
				Type:        command.FlagTypeString,
				Description: "System prompt",
			},
			{
				Name:        "format",
				Type:        command.FlagTypeString,
				Description: "Response format (text, json, markdown)",
			},
			{
				Name:        "output",
				Short:       "o",
				Type:        command.FlagTypeString,
				Description: "Output format (text, json)",
			},
		},
	}
}

// Execute runs the ask command
func (c *AskCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	// Validate we have a prompt
	if len(exec.Args) == 0 {
		return fmt.Errorf("prompt is required")
	}

	// Combine args into the prompt
	prompt := strings.Join(exec.Args, " ")

	// Get model from flags, profile, or config
	model := exec.Flags.GetString("model")
	if model == "" {
		// Check current profile for model setting
		profileName := c.config.GetString("profile.current")
		logging.LogDebug("Current profile from config", "profile", profileName)
		
		if profileName != "" {
			// Check if profile specifies a model
			profileConfig, err := c.config.GetProfile(profileName)
			if err != nil {
				logging.LogWarn("Failed to get profile config", "profile", profileName, "error", err)
			} else {
				logging.LogDebug("Profile config", "profile", profileName, "provider", profileConfig.Provider, "model", profileConfig.Model)
				
				if profileConfig.Provider != "" {
					// Construct model string from profile settings
					if profileConfig.Model != "" {
						model = fmt.Sprintf("%s/%s", profileConfig.Provider, profileConfig.Model)
						logging.LogDebug("Using model from profile", "profile", profileName, "model", model)
					} else {
						// Only provider specified in profile, use default model for that provider
						providerDefaultModel := c.config.GetString(fmt.Sprintf("provider.%s.default_model", profileConfig.Provider))
						if providerDefaultModel != "" {
							model = fmt.Sprintf("%s/%s", profileConfig.Provider, providerDefaultModel)
							logging.LogDebug("Using provider from profile with default model", "profile", profileName, "model", model)
						}
					}
				}
			}
		}
		
		// If still no model, fall back to global default
		if model == "" {
			model = c.config.GetString("model.default")
			logging.LogDebug("Using global default model", "model", model)
			if model == "" {
				model = "openai/gpt-4o" // fallback default
				logging.LogDebug("Using hardcoded fallback model", "model", model)
			}
		}
	} else {
		logging.LogDebug("Using model from command line flag", "model", model)
	}

	// Parse provider and model
	providerName, modelName := llm.ParseModelString(model)

	// Get API key from config for the provider
	apiKey := c.config.GetString(fmt.Sprintf("provider.%s.api_key", providerName))

	// Create the provider, passing the API key from config
	provider, err := llm.NewProvider(providerName, modelName, apiKey)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Build provider options
	var opts []llm.ProviderOption

	if temp := exec.Flags.GetFloat("temperature"); temp != 0 {
		opts = append(opts, llm.WithTemperature(temp))
	}

	if maxTokens := exec.Flags.GetInt("max-tokens"); maxTokens > 0 {
		opts = append(opts, llm.WithMaxTokens(maxTokens))
	}

	if format := exec.Flags.GetString("format"); format != "" {
		opts = append(opts, llm.WithResponseFormat(format))
	}

	// Build messages
	messages := []domain.Message{}

	// Add system prompt if provided
	if system := exec.Flags.GetString("system"); system != "" {
		messages = append(messages, domain.Message{
			Role:    "system",
			Content: system,
		})
	} else if defaultSystem := c.config.GetString("defaults.system_prompt"); defaultSystem != "" {
		messages = append(messages, domain.Message{
			Role:    "system",
			Content: defaultSystem,
		})
	}

	// Process attachments
	attachments := []domain.Attachment{}
	attachFiles := exec.Flags.GetStringSlice("attach")

	// Check if the model supports file attachments
	modelInfo := provider.GetModelInfo()
	supportsFiles := modelInfo.Capabilities.File

	for _, file := range attachFiles {
		if supportsFiles {
			// Create file attachment for models that support it
			attachment := domain.Attachment{
				Type:     domain.AttachmentTypeFile,
				FilePath: file,
			}
			attachments = append(attachments, attachment)
		} else {
			// For models that don't support files, read the content as text
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", file, err)
			}
			// Convert file content to text attachment
			attachment := domain.Attachment{
				Type:    domain.AttachmentTypeText,
				Content: []byte(fmt.Sprintf("Content of %s:\n\n%s", file, string(content))),
			}
			attachments = append(attachments, attachment)
		}
	}

	// Add user message with prompt and attachments
	userMessage := domain.Message{
		Role:    "user",
		Content: prompt,
	}
	if len(attachments) > 0 {
		userMessage.Attachments = attachments
	}
	messages = append(messages, userMessage)

	// Handle streaming vs non-streaming
	if exec.Flags.GetBool("stream") {
		return c.executeStreaming(ctx, exec, provider, messages, opts)
	}

	return c.executeNonStreaming(ctx, exec, provider, messages, opts)
}

// executeNonStreaming handles non-streaming requests
func (c *AskCommand) executeNonStreaming(ctx context.Context, exec *command.ExecutionContext, provider llm.Provider, messages []domain.Message, opts []llm.ProviderOption) error {
	// Generate response
	response, err := provider.GenerateMessage(ctx, messages, opts...)
	if err != nil {
		return fmt.Errorf("failed to generate response: %w", err)
	}

	// Output based on format
	outputFormat := exec.Flags.GetString("output")
	if outputFormat == "" {
		// Check global output preference from command line flag
		outputFormat = c.config.GetString("output")
	}

	switch outputFormat {
	case "json":
		jsonOutput := map[string]interface{}{
			"content":       response.Content,
			"model":         provider.GetModelInfo().Model,
			"provider":      provider.GetModelInfo().Provider,
			"finish_reason": response.FinishReason,
		}
		if response.Usage != nil {
			jsonOutput["usage"] = response.Usage
		}

		encoder := json.NewEncoder(exec.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(jsonOutput)

	default: // text
		_, err := fmt.Fprint(exec.Stdout, response.Content)
		return err
	}
}

// executeStreaming handles streaming requests
func (c *AskCommand) executeStreaming(ctx context.Context, exec *command.ExecutionContext, provider llm.Provider, messages []domain.Message, opts []llm.ProviderOption) error {
	// Start streaming
	stream, err := provider.StreamMessage(ctx, messages, opts...)
	if err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}

	// Collect content for final output if needed
	var content strings.Builder
	isJSON := exec.Flags.GetString("output") == "json"

	// Stream chunks to output
	for chunk := range stream {
		if chunk.Error != nil {
			return fmt.Errorf("streaming error: %w", chunk.Error)
		}

		if isJSON {
			// Collect for JSON output
			content.WriteString(chunk.Content)
		} else {
			// Stream directly to output
			fmt.Fprint(exec.Stdout, chunk.Content)
		}
	}

	// Output JSON format if requested
	if isJSON {
		jsonOutput := map[string]interface{}{
			"content":  content.String(),
			"model":    provider.GetModelInfo().Model,
			"provider": provider.GetModelInfo().Provider,
		}

		encoder := json.NewEncoder(exec.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(jsonOutput)
	}

	return nil
}

// Validate implements the Command interface
func (c *AskCommand) Validate() error {
	// Validation is done in Execute for now
	return nil
}
