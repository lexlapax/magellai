// ABOUTME: Mock implementations for LLM provider interface
// ABOUTME: Provides reusable mock LLM providers for testing

package mocks

import (
	"context"
	"fmt"
	"sync"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// MockProvider implements the LLM Provider interface for testing
type MockProvider struct {
	mu            sync.RWMutex
	response      *llm.Response
	streamChunks  []llm.StreamChunk
	modelInfo     llm.ModelInfo
	errorToReturn error
	callCounts    map[string]int
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		modelInfo: llm.ModelInfo{
			Provider: "mock",
			Model:    "test",
			Capabilities: llm.ModelCapabilities{
				Text: true,
			},
		},
		callCounts: make(map[string]int),
	}
}

// SetResponse sets the response to return for Generate calls
func (mp *MockProvider) SetResponse(response *llm.Response) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.response = response
}

// SetStreamChunks sets the chunks to return for Stream calls
func (mp *MockProvider) SetStreamChunks(chunks []llm.StreamChunk) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.streamChunks = chunks
}

// SetError sets the error to return for all operations
func (mp *MockProvider) SetError(err error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.errorToReturn = err
}

// GetCallCount returns the call count for a specific method
func (mp *MockProvider) GetCallCount(method string) int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.callCounts[method]
}

// Generate generates a response
func (mp *MockProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.callCounts["Generate"]++
	if mp.errorToReturn != nil {
		return "", mp.errorToReturn
	}

	if mp.response != nil {
		return mp.response.Content, nil
	}
	return "Mock response", nil
}

// GenerateMessage generates a message response
func (mp *MockProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.callCounts["GenerateMessage"]++
	if mp.errorToReturn != nil {
		return nil, mp.errorToReturn
	}

	if mp.response != nil {
		return mp.response, nil
	}

	return &llm.Response{
		Content: "Mock response",
		Model:   mp.modelInfo.Model,
	}, nil
}

// GenerateWithSchema generates a structured response
func (mp *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.callCounts["GenerateWithSchema"]++
	if mp.errorToReturn != nil {
		return nil, mp.errorToReturn
	}

	return map[string]interface{}{
		"result": "mock structured response",
	}, nil
}

// Stream provides streaming responses
func (mp *MockProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	mp.mu.RLock()
	chunks := mp.streamChunks
	err := mp.errorToReturn
	mp.callCounts["Stream"]++
	mp.mu.RUnlock()

	if err != nil {
		return nil, err
	}

	ch := make(chan llm.StreamChunk, len(chunks))
	go func() {
		defer close(ch)
		for _, chunk := range chunks {
			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// StreamMessage provides streaming message responses
func (mp *MockProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	mp.mu.Lock()
	mp.callCounts["StreamMessage"]++
	mp.mu.Unlock()

	return mp.Stream(ctx, "", options...)
}

// GetModelInfo returns model information
func (mp *MockProvider) GetModelInfo() llm.ModelInfo {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.modelInfo
}

// SetModelInfo sets the model info to return
func (mp *MockProvider) SetModelInfo(info llm.ModelInfo) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.modelInfo = info
}

// MockFailingProvider is a provider that fails after a certain number of calls
type MockFailingProvider struct {
	failAfter int
	calls     int
	mu        sync.Mutex
}

// NewMockFailingProvider creates a provider that fails after n calls
func NewMockFailingProvider(failAfter int) *MockFailingProvider {
	return &MockFailingProvider{
		failAfter: failAfter,
	}
}

// Generate generates a response or fails
func (mfp *MockFailingProvider) Generate(ctx context.Context, prompt string, options ...llm.ProviderOption) (string, error) {
	mfp.mu.Lock()
	defer mfp.mu.Unlock()

	mfp.calls++
	if mfp.calls > mfp.failAfter {
		return "", fmt.Errorf("provider failed after %d calls", mfp.failAfter)
	}
	return "Mock response", nil
}

// GenerateMessage generates a message response or fails
func (mfp *MockFailingProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (*llm.Response, error) {
	mfp.mu.Lock()
	defer mfp.mu.Unlock()

	mfp.calls++
	if mfp.calls > mfp.failAfter {
		return nil, fmt.Errorf("provider failed after %d calls", mfp.failAfter)
	}

	return &llm.Response{
		Content: "Mock response",
		Model:   "mock-model",
	}, nil
}

// GenerateWithSchema generates structured response or fails
func (mfp *MockFailingProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...llm.ProviderOption) (interface{}, error) {
	mfp.mu.Lock()
	defer mfp.mu.Unlock()

	mfp.calls++
	if mfp.calls > mfp.failAfter {
		return nil, fmt.Errorf("provider failed after %d calls", mfp.failAfter)
	}

	return "Mock schema response", nil
}

// Stream provides streaming responses or fails
func (mfp *MockFailingProvider) Stream(ctx context.Context, prompt string, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	mfp.mu.Lock()
	defer mfp.mu.Unlock()

	mfp.calls++
	if mfp.calls > mfp.failAfter {
		return nil, fmt.Errorf("provider failed after %d calls", mfp.failAfter)
	}

	ch := make(chan llm.StreamChunk, 2)
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{Content: "Mock "}
		ch <- llm.StreamChunk{Content: "streaming response"}
	}()
	return ch, nil
}

// StreamMessage provides streaming message responses or fails
func (mfp *MockFailingProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...llm.ProviderOption) (<-chan llm.StreamChunk, error) {
	return mfp.Stream(ctx, "", options...)
}

// GetModelInfo returns model information
func (mfp *MockFailingProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{
		Provider:    "mock",
		Model:       "mock-model",
		DisplayName: "Mock Model",
	}
}
