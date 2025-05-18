// ABOUTME: Domain types for conversations including Conversation management
// ABOUTME: Core business entity for managing chat conversation state and history

package domain

import (
	"time"
)

// Conversation manages the state of an interactive conversation.
type Conversation struct {
	ID           string                 `json:"id"`
	Messages     []Message              `json:"messages"`
	Model        string                 `json:"model"`
	Provider     string                 `json:"provider"`
	Temperature  float64                `json:"temperature"`
	MaxTokens    int                    `json:"max_tokens"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Created      time.Time              `json:"created"`
	Updated      time.Time              `json:"updated"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NewConversation creates a new conversation with the given ID.
func NewConversation(id string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:          id,
		Messages:    []Message{},
		Created:     now,
		Updated:     now,
		Metadata:    make(map[string]interface{}),
		Temperature: 0.7, // Default temperature
		MaxTokens:   0,   // Default to model's maximum
	}
}

// AddMessage adds a new message to the conversation.
func (c *Conversation) AddMessage(message Message) {
	c.Messages = append(c.Messages, message)
	c.Updated = time.Now()
}

// GetLastMessage returns the last message in the conversation, or nil if empty.
func (c *Conversation) GetLastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

// GetMessageCount returns the total number of messages in the conversation.
func (c *Conversation) GetMessageCount() int {
	return len(c.Messages)
}

// GetUserMessageCount returns the number of user messages in the conversation.
func (c *Conversation) GetUserMessageCount() int {
	count := 0
	for _, msg := range c.Messages {
		if msg.Role == MessageRoleUser {
			count++
		}
	}
	return count
}

// GetAssistantMessageCount returns the number of assistant messages in the conversation.
func (c *Conversation) GetAssistantMessageCount() int {
	count := 0
	for _, msg := range c.Messages {
		if msg.Role == MessageRoleAssistant {
			count++
		}
	}
	return count
}

// SetModel sets the model and provider for the conversation.
func (c *Conversation) SetModel(provider, model string) {
	c.Provider = provider
	c.Model = model
	c.Updated = time.Now()
}

// SetParameters sets the conversation parameters.
func (c *Conversation) SetParameters(temperature float64, maxTokens int) {
	c.Temperature = temperature
	c.MaxTokens = maxTokens
	c.Updated = time.Now()
}

// SetSystemPrompt sets the system prompt for the conversation.
func (c *Conversation) SetSystemPrompt(prompt string) {
	c.SystemPrompt = prompt
	c.Updated = time.Now()
}

// ClearMessages removes all messages from the conversation.
func (c *Conversation) ClearMessages() {
	c.Messages = []Message{}
	c.Updated = time.Now()
}

// IsEmpty returns true if the conversation has no messages.
func (c *Conversation) IsEmpty() bool {
	return len(c.Messages) == 0
}

// Clone creates a deep copy of the conversation.
func (c *Conversation) Clone() *Conversation {
	clone := &Conversation{
		ID:           c.ID,
		Model:        c.Model,
		Provider:     c.Provider,
		Temperature:  c.Temperature,
		MaxTokens:    c.MaxTokens,
		SystemPrompt: c.SystemPrompt,
		Created:      c.Created,
		Updated:      c.Updated,
		Messages:     make([]Message, len(c.Messages)),
		Metadata:     make(map[string]interface{}),
	}

	// Deep copy messages
	for i, msg := range c.Messages {
		clone.Messages[i] = msg.Clone()
	}

	// Deep copy metadata
	for k, v := range c.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}
