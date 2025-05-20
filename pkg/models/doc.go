// ABOUTME: Package for working with the static models inventory file
// ABOUTME: Provides types and utilities for loading and querying models.json

/*
Package models provides functionality for working with the model inventory.

This package handles loading, parsing, and querying the models.json file that
contains metadata about supported LLM models across different providers. It
provides a structured representation of model capabilities and characteristics.

Key Components:
  - Model: Core data structure representing an LLM model
  - Provider: Model provider information
  - Capability: Model capability flags
  - Registry: Collection of models with lookup functionality
  - Loading: Functions for loading models from the inventory file

The models inventory contains information such as:
  - Model identifiers and readable names
  - Provider-specific details
  - Model capabilities (chat, streaming, multimodal, etc.)
  - Context window sizes
  - Pricing information
  - Deprecation status

Usage:
    // Load models from the default location
    modelRegistry, err := models.LoadModels("")
    if err != nil {
        // Handle error
    }

    // Find models by provider
    anthropicModels := modelRegistry.GetModelsByProvider("anthropic")

    // Find models by capability
    streamingModels := modelRegistry.GetModelsByCapability("streaming")

    // Get a specific model
    model, found := modelRegistry.GetModel("claude-3-opus-20240229")

The package ensures that the application has current information about available
models and their capabilities for making informed provider and model selections.
*/
package models