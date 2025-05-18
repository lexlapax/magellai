package domain

import (
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	id := "msg-123"
	role := MessageRoleUser
	content := "Hello, world!"

	msg := NewMessage(id, role, content)

	if msg.ID != id {
		t.Errorf("Expected message ID %s, got %s", id, msg.ID)
	}

	if msg.Role != role {
		t.Errorf("Expected role %s, got %s", role, msg.Role)
	}

	if msg.Content != content {
		t.Errorf("Expected content %s, got %s", content, msg.Content)
	}

	if msg.Metadata == nil {
		t.Error("Expected metadata map to be initialized")
	}

	if len(msg.Attachments) != 0 {
		t.Error("Expected attachments to be empty")
	}

	// Check timestamp is recent
	if time.Since(msg.Timestamp) > time.Second {
		t.Error("Timestamp should be recent")
	}
}

func TestMessageAddAttachment(t *testing.T) {
	msg := NewMessage("msg-123", MessageRoleUser, "Check this file")
	attachment := NewAttachment("att-1", AttachmentTypeFile)
	attachment.Name = "document.pdf"

	msg.AddAttachment(*attachment)

	if len(msg.Attachments) != 1 {
		t.Error("Failed to add attachment")
	}

	if msg.Attachments[0].ID != "att-1" {
		t.Error("Attachment ID mismatch")
	}
}

func TestMessageRemoveAttachment(t *testing.T) {
	msg := NewMessage("msg-123", MessageRoleUser, "Multiple attachments")

	att1 := NewAttachment("att-1", AttachmentTypeFile)
	att2 := NewAttachment("att-2", AttachmentTypeImage)
	att3 := NewAttachment("att-3", AttachmentTypeText)

	msg.Attachments = []Attachment{*att1, *att2, *att3}

	msg.RemoveAttachment("att-2")

	if len(msg.Attachments) != 2 {
		t.Error("Failed to remove attachment")
	}

	for _, att := range msg.Attachments {
		if att.ID == "att-2" {
			t.Error("Attachment was not removed")
		}
	}
}

func TestMessageIsValid(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected bool
	}{
		{
			name: "valid message with content",
			message: Message{
				ID:      "msg-1",
				Role:    MessageRoleUser,
				Content: "Hello",
			},
			expected: true,
		},
		{
			name: "valid message with attachment only",
			message: Message{
				ID:   "msg-2",
				Role: MessageRoleAssistant,
				Attachments: []Attachment{
					*NewAttachment("att-1", AttachmentTypeImage),
				},
			},
			expected: true,
		},
		{
			name: "invalid - no ID",
			message: Message{
				Role:    MessageRoleUser,
				Content: "Hello",
			},
			expected: false,
		},
		{
			name: "invalid - no role",
			message: Message{
				ID:      "msg-3",
				Content: "Hello",
			},
			expected: false,
		},
		{
			name: "invalid - invalid role",
			message: Message{
				ID:      "msg-4",
				Role:    "invalid",
				Content: "Hello",
			},
			expected: false,
		},
		{
			name: "invalid - no content or attachments",
			message: Message{
				ID:   "msg-5",
				Role: MessageRoleUser,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.IsValid()
			if result != tt.expected {
				t.Errorf("Expected IsValid() to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMessageRole(t *testing.T) {
	tests := []struct {
		role     MessageRole
		expected bool
	}{
		{MessageRoleUser, true},
		{MessageRoleAssistant, true},
		{MessageRoleSystem, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if tt.role.IsValid() != tt.expected {
				t.Errorf("Expected IsValid() to return %v for role %s", tt.expected, tt.role)
			}
		})
	}
}
