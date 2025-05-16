// ABOUTME: Tests for the models inventory package
// ABOUTME: Validates loading and querying models.json

package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadInventory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test models.json
	testInventory := &Inventory{
		Metadata: Metadata{
			Version:       "1.0.0",
			LastUpdated:   "2025-01-16",
			Description:   "Test inventory",
			SchemaVersion: "1",
		},
		Models: []Model{
			{
				Provider:    "openai",
				Name:        "gpt-4",
				DisplayName: "GPT-4",
				Description: "Test model",
				Capabilities: Capabilities{
					Text:            MediaCapability{Read: true, Write: true},
					FunctionCalling: true,
					Streaming:       true,
					JSONMode:        true,
				},
				ContextWindow:   8192,
				MaxOutputTokens: 4096,
			},
		},
	}

	// Write test data
	data, err := json.Marshal(testInventory)
	require.NoError(t, err)

	filePath := filepath.Join(tmpDir, "models.json")
	err = os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)

	// Test loading
	loaded, err := LoadInventory(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", loaded.Metadata.Version)
	assert.Len(t, loaded.Models, 1)
	assert.Equal(t, "gpt-4", loaded.Models[0].Name)
}

func TestGetModel(t *testing.T) {
	inventory := &Inventory{
		Models: []Model{
			{Provider: "openai", Name: "gpt-4"},
			{Provider: "anthropic", Name: "claude-3"},
			{Provider: "google", Name: "gemini-pro"},
		},
	}

	// Test existing model
	model := inventory.GetModel("openai", "gpt-4")
	assert.NotNil(t, model)
	assert.Equal(t, "gpt-4", model.Name)

	// Test non-existing model
	model = inventory.GetModel("openai", "gpt-5")
	assert.Nil(t, model)
}

func TestGetModelByFullName(t *testing.T) {
	inventory := &Inventory{
		Models: []Model{
			{Provider: "openai", Name: "gpt-4"},
			{Provider: "anthropic", Name: "claude-3"},
		},
	}

	// Test valid full name
	model := inventory.GetModelByFullName("openai/gpt-4")
	assert.NotNil(t, model)
	assert.Equal(t, "gpt-4", model.Name)

	// Test invalid format
	model = inventory.GetModelByFullName("invalid-format")
	assert.Nil(t, model)

	// Test non-existing model
	model = inventory.GetModelByFullName("openai/gpt-5")
	assert.Nil(t, model)
}

func TestListProviders(t *testing.T) {
	inventory := &Inventory{
		Models: []Model{
			{Provider: "openai", Name: "gpt-4"},
			{Provider: "anthropic", Name: "claude-3"},
			{Provider: "openai", Name: "gpt-3.5"},
			{Provider: "google", Name: "gemini-pro"},
		},
	}

	providers := inventory.ListProviders()
	assert.Len(t, providers, 3)
	assert.Contains(t, providers, "openai")
	assert.Contains(t, providers, "anthropic")
	assert.Contains(t, providers, "google")
}

func TestGetModelsByProvider(t *testing.T) {
	inventory := &Inventory{
		Models: []Model{
			{Provider: "openai", Name: "gpt-4"},
			{Provider: "anthropic", Name: "claude-3"},
			{Provider: "openai", Name: "gpt-3.5"},
			{Provider: "google", Name: "gemini-pro"},
		},
	}

	openaiModels := inventory.GetModelsByProvider("openai")
	assert.Len(t, openaiModels, 2)

	for _, model := range openaiModels {
		assert.Equal(t, "openai", model.Provider)
	}
}

func TestGetModelsWithCapability(t *testing.T) {
	inventory := &Inventory{
		Models: []Model{
			{
				Provider: "openai",
				Name:     "gpt-4",
				Capabilities: Capabilities{
					Text:            MediaCapability{Read: true, Write: true},
					Image:           MediaCapability{Read: true, Write: false},
					FunctionCalling: true,
					Streaming:       true,
					JSONMode:        true,
				},
			},
			{
				Provider: "anthropic",
				Name:     "claude-3",
				Capabilities: Capabilities{
					Text:            MediaCapability{Read: true, Write: true},
					Image:           MediaCapability{Read: true, Write: false},
					FunctionCalling: true,
					Streaming:       true,
					JSONMode:        false,
				},
			},
			{
				Provider: "google",
				Name:     "gemini-pro",
				Capabilities: Capabilities{
					Text:            MediaCapability{Read: true, Write: true},
					Video:           MediaCapability{Read: true, Write: false},
					FunctionCalling: true,
					Streaming:       true,
					JSONMode:        true,
				},
			},
		},
	}

	// Test text writing capability
	textWriters := inventory.GetModelsWithCapability("text", "write")
	assert.Len(t, textWriters, 3)

	// Test image reading capability
	imageReaders := inventory.GetModelsWithCapability("image", "read")
	assert.Len(t, imageReaders, 2)

	// Test video reading capability
	videoReaders := inventory.GetModelsWithCapability("video", "read")
	assert.Len(t, videoReaders, 1)
	assert.Equal(t, "gemini-pro", videoReaders[0].Name)

	// Test JSON mode capability
	jsonModels := inventory.GetModelsWithCapability("json_mode", "")
	assert.Len(t, jsonModels, 2)

	// Test function calling capability
	functionModels := inventory.GetModelsWithCapability("function_calling", "")
	assert.Len(t, functionModels, 3)
}

func TestGetLastUpdated(t *testing.T) {
	inventory := &Inventory{
		Metadata: Metadata{
			LastUpdated: "2025-01-16",
		},
	}

	updated, err := inventory.GetLastUpdated()
	require.NoError(t, err)

	assert.Equal(t, 2025, updated.Year())
	assert.Equal(t, 1, int(updated.Month()))
	assert.Equal(t, 16, updated.Day())
}
