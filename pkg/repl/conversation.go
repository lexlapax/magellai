// ABOUTME: Manages conversation state including message history, context, and attachments
// ABOUTME: Provides methods for adding messages, managing context window, and resetting conversations

package repl

import (
	"time"

	"github.com/lexlapax/magellai/pkg/llm"
)

// Message represents a single message in a conversation
type Message struct {
	ID          string           `json:"id"`
	Role        string           `json:"role"` // user, assistant, system
	Content     string           `json:"content"`
	Timestamp   time.Time        `json:"timestamp"`
	Attachments []llm.Attachment `json:"attachments,omitempty"`
	Metadata    map[string]any   `json:"metadata,omitempty"`
}

// Conversation manages the state of an interactive conversation
type Conversation struct {
	ID           string         `json:"id"`
	Messages     []Message      `json:"messages"`
	Model        string         `json:"model"`
	Provider     string         `json:"provider"`
	Temperature  float64        `json:"temperature"`
	MaxTokens    int            `json:"max_tokens"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
	Created      time.Time      `json:"created"`
	Updated      time.Time      `json:"updated"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// NewConversation creates a new conversation with default settings
func NewConversation(id string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:       id,
		Messages: []Message{},
		Created:  now,
		Updated:  now,
		Metadata: make(map[string]any),
	}
}

// AddMessage adds a new message to the conversation
func (c *Conversation) AddMessage(role, content string, attachments []llm.Attachment) {
	msg := Message{
		ID:          generateMessageID(),
		Role:        role,
		Content:     content,
		Timestamp:   time.Now(),
		Attachments: attachments,
		Metadata:    make(map[string]any),
	}
	c.Messages = append(c.Messages, msg)
	c.Updated = time.Now()
}

// SetSystemPrompt sets the system prompt for the conversation
func (c *Conversation) SetSystemPrompt(prompt string) {
	c.SystemPrompt = prompt
	c.Updated = time.Now()
}

// GetHistory returns the conversation history as LLM messages
func (c *Conversation) GetHistory() []llm.Message {
	var messages []llm.Message

	// Add system prompt first if it exists
	if c.SystemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: c.SystemPrompt,
		})
	}

	// Add conversation messages
	for _, msg := range c.Messages {
		llmMsg := llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Convert attachments if present
		if len(msg.Attachments) > 0 {
			llmMsg.Attachments = msg.Attachments
		}

		messages = append(messages, llmMsg)
	}

	return messages
}

// Reset clears the conversation history while preserving settings
func (c *Conversation) Reset() {
	c.Messages = []Message{}
	c.Updated = time.Now()
}

// CountTokens estimates the token count for the conversation
// This is a simplified implementation - real token counting would use tiktoken or similar
func (c *Conversation) CountTokens() int {
	totalTokens := 0

	// Count system prompt tokens
	if c.SystemPrompt != "" {
		totalTokens += estimateTokens(c.SystemPrompt)
	}

	// Count message tokens
	for _, msg := range c.Messages {
		totalTokens += estimateTokens(msg.Content)
		// Add estimated tokens for attachments
		totalTokens += len(msg.Attachments) * 100 // Rough estimate
	}

	return totalTokens
}

// TrimToMaxTokens removes oldest messages to fit within token limit
func (c *Conversation) TrimToMaxTokens(maxTokens int) {
	currentTokens := c.CountTokens()

	// Remove messages from the beginning until we're under the limit
	for currentTokens > maxTokens && len(c.Messages) > 0 {
		// Always keep the last message (usually the user's latest input)
		if len(c.Messages) <= 1 {
			break
		}

		// Remove the first message
		removed := c.Messages[0]
		c.Messages = c.Messages[1:]

		// Recalculate tokens
		currentTokens -= estimateTokens(removed.Content)
		currentTokens -= len(removed.Attachments) * 100
	}

	if currentTokens < c.CountTokens() {
		c.Updated = time.Now()
	}
}

// Helper functions

func generateMessageID() string {
	// Simple ID generation - could be enhanced with UUIDs
	return time.Now().Format("20060102150405.999999")
}

func estimateTokens(text string) int {
	// Very rough estimation: ~4 characters per token
	// In production, use tiktoken or similar library
	return len(text) / 4
}
