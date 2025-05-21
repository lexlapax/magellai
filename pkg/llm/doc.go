// ABOUTME: LLM provider management and adaptation layer
// ABOUTME: Connects to LLM services and handles message processing

/*
Package llm provides provider management and adaptation for large language models.

This package bridges between the Magellai domain types and the go-llms library,
offering a unified interface for interacting with various AI providers. It handles
provider configuration, message streaming, error recovery, and context management.

Key Components:
  - Provider: Adapter for go-llms providers using domain types
  - DomainProvider: Domain-specific interface for LLM interactions
  - ResilientProvider: Error handling, retries, and fallback mechanisms
  - ContextManager: Manages message context size and prioritization
  - ErrorHandler: Standardized LLM error handling and recovery
  - Models: Model listing and availability

The package implements several reliability features:
  - Error classification and appropriate recovery strategies
  - Rate limit handling with intelligent backoff
  - Provider fallback chains for high availability
  - Connection reestablishment for streaming responses
  - Context window management to prevent token limit errors

Usage:

	// Create a provider from configuration
	provider, err := llm.NewProvider(config)
	if err != nil {
	    // Handle error
	}

	// Process a complete conversation
	response, err := provider.ProcessMessages(ctx, messages)

	// Stream a response
	channel, err := provider.StreamResponse(ctx, messages)

The package is designed to be resilient to common LLM API issues while
providing a clean interface aligned with the domain layer architecture.
*/
package llm
