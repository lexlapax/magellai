// ABOUTME: Package for working with the static models inventory file
// ABOUTME: Provides types and utilities for loading and querying models.json

package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Metadata represents the file metadata
type Metadata struct {
	Version       string `json:"version"`
	LastUpdated   string `json:"last_updated"`
	Description   string `json:"description"`
	SchemaVersion string `json:"schema_version"`
}

// Capabilities represents what a model can do
type Capabilities struct {
	Text            MediaCapability `json:"text"`
	Image           MediaCapability `json:"image"`
	Audio           MediaCapability `json:"audio"`
	Video           MediaCapability `json:"video"`
	File            MediaCapability `json:"file"`
	FunctionCalling bool            `json:"function_calling"`
	Streaming       bool            `json:"streaming"`
	JSONMode        bool            `json:"json_mode"`
}

// MediaCapability represents read/write capabilities for a media type
type MediaCapability struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

// Pricing represents model pricing information
type Pricing struct {
	InputPer1kTokens  float64 `json:"input_per_1k_tokens"`
	OutputPer1kTokens float64 `json:"output_per_1k_tokens"`
}

// Model represents a single model in the inventory
type Model struct {
	Provider         string       `json:"provider"`
	Name             string       `json:"name"`
	DisplayName      string       `json:"display_name"`
	Description      string       `json:"description"`
	DocumentationURL string       `json:"documentation_url"`
	Capabilities     Capabilities `json:"capabilities"`
	ContextWindow    int          `json:"context_window"`
	MaxOutputTokens  int          `json:"max_output_tokens"`
	TrainingCutoff   string       `json:"training_cutoff"`
	ModelFamily      string       `json:"model_family"`
	Pricing          Pricing      `json:"pricing"`
	LastUpdated      string       `json:"last_updated"`
}

// Inventory represents the complete models inventory
type Inventory struct {
	Metadata Metadata `json:"_metadata"`
	Models   []Model  `json:"models"`
}

// LoadInventory loads the models.json file from the root directory
func LoadInventory(rootPath string) (*Inventory, error) {
	filePath := filepath.Join(rootPath, "models.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read models.json: %w", err)
	}

	var inventory Inventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, fmt.Errorf("failed to parse models.json: %w", err)
	}

	return &inventory, nil
}

// GetModel returns a specific model by provider and name
func (inv *Inventory) GetModel(provider, name string) *Model {
	for _, model := range inv.Models {
		if model.Provider == provider && model.Name == name {
			return &model
		}
	}
	return nil
}

// GetModelByFullName returns a model by its full name (provider/model)
func (inv *Inventory) GetModelByFullName(fullName string) *Model {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return nil
	}
	return inv.GetModel(parts[0], parts[1])
}

// ListProviders returns a list of unique providers
func (inv *Inventory) ListProviders() []string {
	providers := make(map[string]bool)
	for _, model := range inv.Models {
		providers[model.Provider] = true
	}

	result := make([]string, 0, len(providers))
	for provider := range providers {
		result = append(result, provider)
	}
	return result
}

// GetModelsByProvider returns all models for a specific provider
func (inv *Inventory) GetModelsByProvider(provider string) []Model {
	var result []Model
	for _, model := range inv.Models {
		if model.Provider == provider {
			result = append(result, model)
		}
	}
	return result
}

// GetModelsWithCapability returns models that have a specific capability
func (inv *Inventory) GetModelsWithCapability(mediaType string, operation string) []Model {
	var result []Model
	for _, model := range inv.Models {
		var capability MediaCapability
		var hasCapability bool

		switch mediaType {
		case "text":
			capability = model.Capabilities.Text
			hasCapability = true
		case "image":
			capability = model.Capabilities.Image
			hasCapability = true
		case "audio":
			capability = model.Capabilities.Audio
			hasCapability = true
		case "video":
			capability = model.Capabilities.Video
			hasCapability = true
		case "file":
			capability = model.Capabilities.File
			hasCapability = true
		case "function_calling":
			hasCapability = model.Capabilities.FunctionCalling
		case "streaming":
			hasCapability = model.Capabilities.Streaming
		case "json_mode":
			hasCapability = model.Capabilities.JSONMode
		default:
			continue
		}

		// For non-media capabilities
		if mediaType == "function_calling" || mediaType == "streaming" || mediaType == "json_mode" {
			if hasCapability {
				result = append(result, model)
			}
			continue
		}

		// For media capabilities, check read/write operation
		if operation == "read" && capability.Read {
			result = append(result, model)
		} else if operation == "write" && capability.Write {
			result = append(result, model)
		} else if operation == "" && (capability.Read || capability.Write) {
			// If no operation specified, include if it has either capability
			result = append(result, model)
		}
	}
	return result
}

// GetLatestVersion returns the metadata version
func (inv *Inventory) GetLatestVersion() string {
	return inv.Metadata.Version
}

// GetLastUpdated returns when the inventory was last updated
func (inv *Inventory) GetLastUpdated() (time.Time, error) {
	return time.Parse("2006-01-02", inv.Metadata.LastUpdated)
}
