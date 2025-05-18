package llm

import (
	"context"
	"testing"
	"time"

	schemadomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/magellai/pkg/domain"
)

func TestDomainProviderAdapter(t *testing.T) {
	// Create a mock provider
	mockProvider := NewMockProvider()

	// Set up mock responses
	mockResponse := &Response{
		Content: "Hello from the model!",
		Model:   "mock/test",
		Usage: &Usage{
			InputTokens:  10,
			OutputTokens: 5,
			TotalTokens:  15,
		},
		FinishReason: "stop",
	}
	mockProvider.SetResponse(mockResponse)

	// Wrap with domain provider
	domainProvider := NewDomainProvider(mockProvider)

	// Create domain messages
	messages := []*domain.Message{
		{
			ID:        "msg1",
			Role:      domain.MessageRoleUser,
			Content:   "Hello!",
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
	}

	// Test GenerateDomainMessage
	ctx := context.Background()
	response, err := domainProvider.GenerateDomainMessage(ctx, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate response
	if response.Role != domain.MessageRoleAssistant {
		t.Errorf("expected role %s, got %s", domain.MessageRoleAssistant, response.Role)
	}
	if response.Content != mockResponse.Content {
		t.Errorf("expected content %s, got %s", mockResponse.Content, response.Content)
	}
	if response.Metadata["model"] != mockResponse.Model {
		t.Errorf("expected model %s, got %s", mockResponse.Model, response.Metadata["model"])
	}
	if response.Metadata["finish_reason"] != mockResponse.FinishReason {
		t.Errorf("expected finish reason %s, got %s", mockResponse.FinishReason, response.Metadata["finish_reason"])
	}

	// Check usage metadata
	usage, ok := response.Metadata["usage"].(*Usage)
	if !ok {
		t.Fatalf("expected usage metadata to be *Usage, got %T", response.Metadata["usage"])
	}
	if usage.TotalTokens != mockResponse.Usage.TotalTokens {
		t.Errorf("expected total tokens %d, got %d", mockResponse.Usage.TotalTokens, usage.TotalTokens)
	}
}

func TestDomainProviderStreamAdapter(t *testing.T) {
	// Create a mock provider
	mockProvider := NewMockProvider()

	// Set up streaming response
	chunks := []StreamChunk{
		{Content: "Hello", Index: 0},
		{Content: " from", Index: 1},
		{Content: " the", Index: 2},
		{Content: " model!", Index: 3, FinishReason: "stop"},
	}
	mockProvider.SetStreamChunks(chunks)

	// Wrap with domain provider
	domainProvider := NewDomainProvider(mockProvider)

	// Create domain messages
	messages := []*domain.Message{
		{
			ID:        "msg1",
			Role:      domain.MessageRoleUser,
			Content:   "Hello!",
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
	}

	// Test StreamDomainMessage
	ctx := context.Background()
	stream, err := domainProvider.StreamDomainMessage(ctx, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Collect streamed messages
	var streamedMessages []*domain.Message
	for msg := range stream {
		streamedMessages = append(streamedMessages, msg)
	}

	// Should have received messages for each chunk
	if len(streamedMessages) != len(chunks) {
		t.Errorf("expected %d messages, got %d", len(chunks), len(streamedMessages))
	}

	// Check final message content
	lastMsg := streamedMessages[len(streamedMessages)-1]
	expectedContent := "Hello from the model!"
	if lastMsg.Content != expectedContent {
		t.Errorf("expected final content %s, got %s", expectedContent, lastMsg.Content)
	}
	if lastMsg.Role != domain.MessageRoleAssistant {
		t.Errorf("expected role %s, got %s", domain.MessageRoleAssistant, lastMsg.Role)
	}
	if lastMsg.Metadata["finish_reason"] != "stop" {
		t.Errorf("expected finish reason 'stop', got %s", lastMsg.Metadata["finish_reason"])
	}
}

// MockProvider for testing
type MockProvider struct {
	response     *Response
	streamChunks []StreamChunk
	modelInfo    ModelInfo
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		modelInfo: ModelInfo{
			Provider: "mock",
			Model:    "test",
			Capabilities: ModelCapabilities{
				Text: true,
			},
		},
	}
}

func (m *MockProvider) SetResponse(response *Response) {
	m.response = response
}

func (m *MockProvider) SetStreamChunks(chunks []StreamChunk) {
	m.streamChunks = chunks
}

func (m *MockProvider) Generate(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
	if m.response != nil {
		return m.response.Content, nil
	}
	return "Mock response", nil
}

func (m *MockProvider) GenerateMessage(ctx context.Context, messages []Message, options ...ProviderOption) (*Response, error) {
	if m.response != nil {
		return m.response, nil
	}
	return &Response{Content: "Mock response"}, nil
}

func (m *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemadomain.Schema, options ...ProviderOption) (interface{}, error) {
	return m.response, nil
}

func (m *MockProvider) Stream(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		for _, chunk := range m.streamChunks {
			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (m *MockProvider) StreamMessage(ctx context.Context, messages []Message, options ...ProviderOption) (<-chan StreamChunk, error) {
	return m.Stream(ctx, "", options...)
}

func (m *MockProvider) GetModelInfo() ModelInfo {
	return m.modelInfo
}
