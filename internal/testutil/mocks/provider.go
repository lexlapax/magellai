// ABOUTME: Mock implementations for LLM provider interface
// ABOUTME: Provides reusable mock LLM providers for testing

package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// MockProvider implements the llm.Provider interface for testing
type MockProvider struct {
	mu            sync.RWMutex
	modelInfo     llm.ModelInfo
	responseText  string
	responseMsg   domain.Message
	streamChunks  []string
	options       []llm.ProviderOption
	errorToReturn error
	callCounts    map[string]int
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		modelInfo: llm.ModelInfo{
			Provider:     "mock",
			Model:        "mock-model",
			DisplayName:  "Mock Model",
			Description:  "Mock model for testing",
			MaxTokens:    4096,
			Capabilities: llm.ModelCapabilities{Text: true},
		},
		callCounts: make(map[string]int),
	}
}

// WithModelInfo sets the model info for the mock provider
func (m *MockProvider) WithModelInfo(info llm.ModelInfo) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelInfo = info
	return m
}

// WithResponseText sets the response text
func (m *MockProvider) WithResponseText(text string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseText = text
	return m
}

// WithResponseMessage sets the response message
func (m *MockProvider) WithResponseMessage(msg domain.Message) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseMsg = msg
	return m
}

// WithStreamChunks sets the stream chunks
func (m *MockProvider) WithStreamChunks(chunks []string) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamChunks = chunks
	return m
}

// WithError sets the error to return
func (m *MockProvider) WithError(err error) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorToReturn = err
	return m
}

// WithOptions sets the provider options
func (m *MockProvider) WithOptions(options ...llm.ProviderOption) *MockProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.options = options
	return m
}

// Generate implements the Provider interface
func (m *MockProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCounts["Generate"]++

	if m.errorToReturn != nil {
		return "", m.errorToReturn
	}

	// Apply options - mock just collects them
	m.options = append(m.options, options...)

	// Return predefined response or echo the prompt
	if m.responseText != "" {
		return m.responseText, nil
	}

	return fmt.Sprintf("Mock response to: %s", prompt), nil
}

// GenerateMessage implements the Provider interface
func (m *MockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCounts["GenerateMessage"]++

	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}

	// Apply options - mock just collects them
	m.options = append(m.options, options...)

	// Build response
	var content string
	if m.responseText != "" {
		content = m.responseText
	} else if m.responseMsg.Content != "" {
		content = m.responseMsg.Content
	} else {
		content = "Mock assistant response"
	}

	return &llm.Response{
		Content: content,
		Model:   m.modelInfo.Model,
		Usage: &llm.Usage{
			InputTokens:  100,
			OutputTokens: 50,
			TotalTokens:  150,
		},
		Metadata: map[string]interface{}{
			"mock":    true,
			"elapsed": 0.5,
		},
		FinishReason: "stop",
	}, nil
}

// GenerateWithSchema implements the Provider interface
func (m *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCounts["GenerateWithSchema"]++

	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}

	// Apply options - mock just collects them
	m.options = append(m.options, options...)

	// Return a mock object matching the schema
	// This is a simplified implementation that just returns a mock object
	// A real implementation would try to generate data matching the schema
	return map[string]interface{}{
		"mock_response": true,
		"prompt":        prompt,
		"schema_type":   schema.Type,
	}, nil
}

// Stream implements the Provider interface
func (m *MockProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCounts["Stream"]++

	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}

	// Create channel for stream chunks
	chunkChan := make(chan llm.StreamChunk)

	// Get chunks to send
	chunks := m.streamChunks
	if len(chunks) == 0 {
		chunks = []string{"Mock ", "streaming ", "response ", "for: ", prompt}
	}

	// Send chunks in goroutine
	go func() {
		defer close(chunkChan)

		for i, chunk := range chunks {
			// Check if context is done
			select {
			case <-ctx.Done():
				return
			default:
				// Send chunk
				chunkChan <- llm.StreamChunk{
					Content: chunk,
					Index:   i,
					Done:    i == len(chunks)-1,
				}

				// Add small delay for realism
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	return chunkChan, nil
}
