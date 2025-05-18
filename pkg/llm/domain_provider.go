// ABOUTME: Domain-aware provider interface for working with domain types
// ABOUTME: Extends the base Provider interface with domain type support

package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
)

// DomainProvider extends Provider with domain type support
type DomainProvider interface {
	Provider

	// GenerateDomainMessage produces a response from domain messages
	GenerateDomainMessage(ctx context.Context, messages []*domain.Message, options ...ProviderOption) (*domain.Message, error)

	// StreamDomainMessage streams responses from domain messages
	StreamDomainMessage(ctx context.Context, messages []*domain.Message, options ...ProviderOption) (<-chan *domain.Message, error)
}

// domainProviderAdapter adds domain support to any Provider
type domainProviderAdapter struct {
	Provider
}

// NewDomainProvider wraps a Provider with domain type support
func NewDomainProvider(provider Provider) DomainProvider {
	return &domainProviderAdapter{Provider: provider}
}

// GenerateDomainMessage produces a response from domain messages
func (p *domainProviderAdapter) GenerateDomainMessage(ctx context.Context, messages []*domain.Message, options ...ProviderOption) (*domain.Message, error) {
	// Convert domain messages to LLM messages
	llmMessages := FromDomainMessages(messages)

	// Generate response using the underlying provider
	response, err := p.GenerateMessage(ctx, llmMessages, options...)
	if err != nil {
		return nil, err
	}

	// Create domain message from response
	domainMsg := &domain.Message{
		ID:      generateMessageID(),
		Role:    domain.MessageRoleAssistant,
		Content: response.Content,
		Timestamp: time.Now(),
		Metadata: make(map[string]interface{}),
	}

	// Add response metadata
	if response.Model != "" {
		domainMsg.Metadata["model"] = response.Model
	}
	if response.Usage != nil {
		domainMsg.Metadata["usage"] = response.Usage
	}
	if response.FinishReason != "" {
		domainMsg.Metadata["finish_reason"] = response.FinishReason
	}

	return domainMsg, nil
}

// StreamDomainMessage streams responses from domain messages
func (p *domainProviderAdapter) StreamDomainMessage(ctx context.Context, messages []*domain.Message, options ...ProviderOption) (<-chan *domain.Message, error) {
	// Convert domain messages to LLM messages
	llmMessages := FromDomainMessages(messages)

	// Start streaming using the underlying provider
	chunkStream, err := p.StreamMessage(ctx, llmMessages, options...)
	if err != nil {
		return nil, err
	}

	// Create domain message stream
	domainStream := make(chan *domain.Message)
	
	go func() {
		defer close(domainStream)
		
		var content strings.Builder
		var currentMsg *domain.Message
		
		for chunk := range chunkStream {
			if chunk.Error != nil {
				// Send error as metadata
				if currentMsg == nil {
					currentMsg = &domain.Message{
						ID:        generateMessageID(),
						Role:      domain.MessageRoleAssistant,
						Timestamp: time.Now(),
						Metadata:  make(map[string]interface{}),
					}
				}
				currentMsg.Metadata["error"] = chunk.Error.Error()
				select {
				case domainStream <- currentMsg:
				case <-ctx.Done():
					return
				}
				continue
			}
			
			// Accumulate content
			content.WriteString(chunk.Content)
			
			// Create or update message
			if currentMsg == nil {
				currentMsg = &domain.Message{
					ID:        generateMessageID(),
					Role:      domain.MessageRoleAssistant,
					Content:   chunk.Content,
					Timestamp: time.Now(),
					Metadata:  make(map[string]interface{}),
				}
			} else {
				currentMsg.Content = content.String()
			}
			
			// Check if this is the final chunk
			if chunk.FinishReason != "" {
				currentMsg.Metadata["finish_reason"] = chunk.FinishReason
			}
			
			// Send the current state
			select {
			case domainStream <- currentMsg:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return domainStream, nil
}

// generateMessageID creates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}