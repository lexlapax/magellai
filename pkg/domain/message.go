// ABOUTME: Domain types for messages including Message and MessageRole
// ABOUTME: Core business entities for conversation messages and their metadata

package domain

import (
	"time"
)

// Message represents a single message within a conversation.
type Message struct {
	ID          string                 `json:"id"`
	Role        MessageRole            `json:"role"`
	Content     string                 `json:"content"`
	Timestamp   time.Time              `json:"timestamp"`
	Attachments []Attachment           `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageRole represents the role of a message sender.
type MessageRole string

// MessageRole constants define the possible roles in a conversation.
const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// NewMessage creates a new message with the given parameters.
func NewMessage(id string, role MessageRole, content string) *Message {
	return &Message{
		ID:          id,
		Role:        role,
		Content:     content,
		Timestamp:   time.Now(),
		Attachments: []Attachment{},
		Metadata:    make(map[string]interface{}),
	}
}

// AddAttachment adds an attachment to the message.
func (m *Message) AddAttachment(attachment Attachment) {
	m.Attachments = append(m.Attachments, attachment)
}

// RemoveAttachment removes an attachment from the message by ID.
func (m *Message) RemoveAttachment(attachmentID string) {
	attachments := make([]Attachment, 0, len(m.Attachments))
	for _, a := range m.Attachments {
		if a.ID != attachmentID {
			attachments = append(attachments, a)
		}
	}
	m.Attachments = attachments
}

// IsValid validates the message fields.
func (m *Message) IsValid() bool {
	return m.ID != "" &&
		m.Role != "" &&
		(m.Role == MessageRoleUser || m.Role == MessageRoleAssistant || m.Role == MessageRoleSystem) &&
		(m.Content != "" || len(m.Attachments) > 0) // Message must have content or attachments
}

// String returns the message role as a string.
func (r MessageRole) String() string {
	return string(r)
}

// IsValid checks if the message role is valid.
func (r MessageRole) IsValid() bool {
	return r == MessageRoleUser || r == MessageRoleAssistant || r == MessageRoleSystem
}

// Clone creates a deep copy of the message.
func (m *Message) Clone() Message {
	clone := Message{
		ID:        m.ID,
		Role:      m.Role,
		Content:   m.Content,
		Timestamp: m.Timestamp,
		Attachments: make([]Attachment, len(m.Attachments)),
		Metadata:    make(map[string]interface{}),
	}
	
	// Deep copy attachments
	copy(clone.Attachments, m.Attachments)
	
	// Deep copy metadata
	for k, v := range m.Metadata {
		clone.Metadata[k] = v
	}
	
	return clone
}
