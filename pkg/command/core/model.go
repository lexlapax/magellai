// ABOUTME: Model command implementation for switching between LLM models
// ABOUTME: Supports list, select, info, and validation operations

package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/pkg/llm"
)

// OutputFormat constants for different output formats
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

// ModelCommand implements the model command for managing LLM models
type ModelCommand struct {
	config *config.Config
}

// NewModelCommand creates a new model command instance
func NewModelCommand(cfg *config.Config) *ModelCommand {
	return &ModelCommand{
		config: cfg,
	}
}

// Execute executes the model command
func (c *ModelCommand) Execute(ctx context.Context, exec *command.ExecutionContext) error {
	// Initialize data map if needed
	if exec.Data == nil {
		exec.Data = make(map[string]interface{})
	}
	// Handle subcommands based on first argument
	if len(exec.Args) > 0 {
		switch exec.Args[0] {
		case "list":
			return c.listModels(ctx, exec)
		case "info":
			return c.showModelInfo(ctx, exec)
		default:
			// If not a subcommand, handle model selection
			return c.selectModel(ctx, exec)
		}
	}
	// Show current model if no arguments
	return c.showCurrentModel(ctx, exec)
}

// Metadata returns the command metadata
func (c *ModelCommand) Metadata() *command.Metadata {
	return &command.Metadata{
		Name:        "model",
		Description: "Manage and switch between LLM models",
		Category:    command.CategoryShared,
		Flags: []command.Flag{
			{
				Name:        "provider",
				Short:       "p",
				Description: "Filter models by provider",
				Type:        command.FlagTypeString,
				Required:    false,
			},
			{
				Name:        "capabilities",
				Short:       "c",
				Description: "Filter models by capabilities (text,audio,video,image,file)",
				Type:        command.FlagTypeString,
				Required:    false,
			},
		},
		LongDescription: `The model command manages LLM models. Examples:
			model                          # Show current model
			model openai/gpt-4            # Switch to OpenAI GPT-4
			model anthropic/claude-3-opus # Switch to Anthropic Claude 3 Opus  
			model list                    # List all available models
			model list --provider openai  # List OpenAI models
			model info gemini/pro         # Show info about Gemini Pro`,
	}
}

// Validate validates the command configuration
func (c *ModelCommand) Validate() error {
	if c.config == nil {
		return command.ErrInvalidCommand
	}
	return nil
}

// capitalizeProviderName capitalizes the first letter of a provider name
func capitalizeProviderName(provider string) string {
	if len(provider) == 0 {
		return provider
	}
	return strings.ToUpper(provider[:1]) + provider[1:]
}

// listModels lists available models
func (c *ModelCommand) listModels(ctx context.Context, exec *command.ExecutionContext) error {
	// Get filter options
	providerFilter := exec.Flags.GetString("provider")
	capabilitiesFilter := exec.Flags.GetString("capabilities")

	// Get available models from providers
	models := llm.GetAvailableModels()

	// Apply filters
	filteredModels := []llm.ModelInfo{}
	for _, model := range models {
		// Provider filter
		if providerFilter != "" && !strings.EqualFold(model.Provider, providerFilter) {
			continue
		}

		// Capabilities filter
		if capabilitiesFilter != "" {
			requiredCaps := strings.Split(capabilitiesFilter, ",")
			hasAllCaps := true
			for _, cap := range requiredCaps {
				cap = strings.TrimSpace(cap)
				switch strings.ToLower(cap) {
				case "text":
					if !model.Capabilities.Text {
						hasAllCaps = false
					}
				case "audio":
					if !model.Capabilities.Audio {
						hasAllCaps = false
					}
				case "video":
					if !model.Capabilities.Video {
						hasAllCaps = false
					}
				case "image":
					if !model.Capabilities.Image {
						hasAllCaps = false
					}
				case "file":
					if !model.Capabilities.File {
						hasAllCaps = false
					}
				}
			}
			if !hasAllCaps {
				continue
			}
		}

		filteredModels = append(filteredModels, model)
	}

	// Format output
	if outputFormat, ok := exec.Data["outputFormat"]; ok && outputFormat == OutputFormatJSON {
		exec.Data["output"] = filteredModels
		return nil
	}

	// Text output
	output := strings.Builder{}
	output.WriteString("Available Models:\n\n")

	currentProvider := ""
	currentModel := c.config.GetDefaultModel()

	for _, model := range filteredModels {
		if model.Provider != currentProvider {
			if currentProvider != "" {
				output.WriteString("\n")
			}
			output.WriteString(fmt.Sprintf("%s:\n", capitalizeProviderName(model.Provider)))
			currentProvider = model.Provider
		}

		modelName := fmt.Sprintf("%s/%s", model.Provider, model.Model)
		indicator := "  "
		if modelName == currentModel {
			indicator = "* "
		}

		caps := []string{}
		if model.Capabilities.Text {
			caps = append(caps, "text")
		}
		if model.Capabilities.Audio {
			caps = append(caps, "audio")
		}
		if model.Capabilities.Video {
			caps = append(caps, "video")
		}
		if model.Capabilities.Image {
			caps = append(caps, "image")
		}
		if model.Capabilities.File {
			caps = append(caps, "file")
		}

		output.WriteString(fmt.Sprintf("%s%s [%s]\n", indicator, model.Model, strings.Join(caps, ", ")))
	}

	exec.Data["output"] = output.String()
	return nil
}

// showModelInfo shows detailed information about a model
func (c *ModelCommand) showModelInfo(ctx context.Context, exec *command.ExecutionContext) error {
	if len(exec.Args) < 2 {
		return command.ErrMissingArgument
	}

	modelName := exec.Args[1]

	// Parse provider/model format
	provider, model := llm.ParseModelString(modelName)

	// Check if it's in the expected format
	if !strings.Contains(modelName, "/") {
		return fmt.Errorf("invalid model format: %s (expected provider/model)", modelName)
	}

	// Get model info
	modelInfo, err := llm.GetModelInfo(provider, model)
	if err != nil {
		return fmt.Errorf("model not found: %s", modelName)
	}

	// Format output
	if outputFormat, ok := exec.Data["outputFormat"]; ok && outputFormat == OutputFormatJSON {
		exec.Data["output"] = modelInfo
		return nil
	}

	// Text output
	output := strings.Builder{}
	output.WriteString(fmt.Sprintf("Model: %s/%s\n", modelInfo.Provider, modelInfo.Model))
	output.WriteString(fmt.Sprintf("Provider: %s\n", capitalizeProviderName(modelInfo.Provider)))
	output.WriteString(fmt.Sprintf("Display Name: %s\n", modelInfo.DisplayName))
	output.WriteString(fmt.Sprintf("Description: %s\n", modelInfo.Description))
	output.WriteString("\nCapabilities:\n")

	if modelInfo.Capabilities.Text {
		output.WriteString("  - Text processing\n")
	}
	if modelInfo.Capabilities.Audio {
		output.WriteString("  - Audio processing\n")
	}
	if modelInfo.Capabilities.Video {
		output.WriteString("  - Video processing\n")
	}
	if modelInfo.Capabilities.Image {
		output.WriteString("  - Image processing\n")
	}
	if modelInfo.Capabilities.File {
		output.WriteString("  - File processing\n")
	}

	output.WriteString(fmt.Sprintf("\nMax Tokens: %d\n", modelInfo.MaxTokens))
	output.WriteString(fmt.Sprintf("Context Window: %d\n", modelInfo.ContextWindow))

	if modelInfo.DefaultTemperature > 0 {
		output.WriteString(fmt.Sprintf("Default Temperature: %.2f\n", modelInfo.DefaultTemperature))
	}

	exec.Data["output"] = output.String()
	return nil
}

// selectModel switches to a specified model
func (c *ModelCommand) selectModel(ctx context.Context, exec *command.ExecutionContext) error {
	modelName := exec.Args[0]

	// Parse provider/model format
	provider, model := llm.ParseModelString(modelName)

	// Check if it's in the expected format
	if !strings.Contains(modelName, "/") {
		return fmt.Errorf("invalid model format: %s (expected provider/model)", modelName)
	}

	// Validate model exists
	modelInfo, err := llm.GetModelInfo(provider, model)
	if err != nil {
		return fmt.Errorf("model not found: %s", modelName)
	}

	// Get current model before making changes
	currentModel := c.config.GetDefaultModel()

	// Update configuration
	if err := c.config.SetDefaultProvider(provider); err != nil {
		return fmt.Errorf("failed to set provider: %w", err)
	}

	if err := c.config.SetDefaultModel(modelName); err != nil {
		return fmt.Errorf("failed to set model: %w", err)
	}

	// Log the model change
	logging.LogInfo("Model changed", "from", currentModel, "to", modelName)

	// Format output
	if outputFormat, ok := exec.Data["outputFormat"]; ok && outputFormat == OutputFormatJSON {
		exec.Data["output"] = map[string]string{
			"provider": provider,
			"model":    modelName,
			"message":  fmt.Sprintf("Switched to %s", modelName),
		}
		return nil
	}

	exec.Data["output"] = fmt.Sprintf("Switched to %s (%s)", modelInfo.DisplayName, modelName)
	return nil
}

// showCurrentModel displays the currently selected model
func (c *ModelCommand) showCurrentModel(ctx context.Context, exec *command.ExecutionContext) error {
	currentModel := c.config.GetDefaultModel()

	if currentModel == "" {
		exec.Data["output"] = "No model selected"
		return nil
	}

	// Parse provider/model format
	provider, model := llm.ParseModelString(currentModel)

	// Get model info
	modelInfo, err := llm.GetModelInfo(provider, model)
	if err != nil {
		exec.Data["output"] = fmt.Sprintf("Current model: %s (not found in registry)", currentModel)
		return nil
	}

	// Format output
	if outputFormat, ok := exec.Data["outputFormat"]; ok && outputFormat == OutputFormatJSON {
		exec.Data["output"] = map[string]string{
			"provider":     provider,
			"model":        currentModel,
			"display_name": modelInfo.DisplayName,
		}
		return nil
	}

	exec.Data["output"] = fmt.Sprintf("Current model: %s (%s)", modelInfo.DisplayName, currentModel)
	return nil
}
