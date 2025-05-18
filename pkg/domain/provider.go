// ABOUTME: Domain types for LLM providers including Provider, Model, and ModelCapability
// ABOUTME: Core business entities for provider and model configuration

package domain

// Provider represents an LLM provider configuration.
type Provider struct {
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"display_name"`
	Endpoint        string                 `json:"endpoint,omitempty"`
	APIKeyEnvVar    string                 `json:"api_key_env_var,omitempty"`
	Models          []Model                `json:"models"`
	Capabilities    []string               `json:"capabilities"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Model represents a specific model offered by a provider.
type Model struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"display_name"`
	Provider        string                 `json:"provider"`
	ContextWindow   int                    `json:"context_window"`
	MaxOutputTokens int                    `json:"max_output_tokens"`
	InputCost       float64                `json:"input_cost"`  // Cost per 1M tokens
	OutputCost      float64                `json:"output_cost"` // Cost per 1M tokens
	Capabilities    []ModelCapability      `json:"capabilities"`
	Modalities      []string               `json:"modalities"`
	ReleaseDate     string                 `json:"release_date,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ModelCapability represents a specific capability of a model.
type ModelCapability string

// ModelCapability constants define the possible model capabilities.
const (
	ModelCapabilityStreaming        ModelCapability = "streaming"
	ModelCapabilityToolCalling      ModelCapability = "tool_calling"
	ModelCapabilityVision           ModelCapability = "vision"
	ModelCapabilityAudio            ModelCapability = "audio"
	ModelCapabilityVideo            ModelCapability = "video"
	ModelCapabilityFileAttachments  ModelCapability = "file_attachments"
	ModelCapabilityStructuredOutput ModelCapability = "structured_output"
	ModelCapabilityFunctionCalling  ModelCapability = "function_calling"
	ModelCapabilityJSONMode         ModelCapability = "json_mode"
)

// ModalityType represents the type of content a model can process.
type ModalityType string

// ModalityType constants define the possible modalities.
const (
	ModalityText  ModalityType = "text"
	ModalityImage ModalityType = "image"
	ModalityAudio ModalityType = "audio"
	ModalityVideo ModalityType = "video"
	ModalityFile  ModalityType = "file"
)

// NewProvider creates a new provider with the given name.
func NewProvider(name, displayName string) *Provider {
	return &Provider{
		Name:         name,
		DisplayName:  displayName,
		Models:       []Model{},
		Capabilities: []string{},
		Metadata:     make(map[string]interface{}),
	}
}

// AddModel adds a model to the provider.
func (p *Provider) AddModel(model Model) {
	p.Models = append(p.Models, model)
}

// GetModel retrieves a model by ID.
func (p *Provider) GetModel(modelID string) *Model {
	for _, model := range p.Models {
		if model.ID == modelID {
			return &model
		}
	}
	return nil
}

// HasCapability checks if the model has a specific capability.
func (m *Model) HasCapability(capability ModelCapability) bool {
	for _, c := range m.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// SupportsModality checks if the model supports a specific modality.
func (m *Model) SupportsModality(modality string) bool {
	for _, mod := range m.Modalities {
		if mod == modality {
			return true
		}
	}
	return false
}

// GetFullID returns the full model ID in provider/model format.
func (m *Model) GetFullID() string {
	return m.Provider + "/" + m.Name
}

// String returns the capability as a string.
func (c ModelCapability) String() string {
	return string(c)
}

// String returns the modality as a string.
func (m ModalityType) String() string {
	return string(m)
}