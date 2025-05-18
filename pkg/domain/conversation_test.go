package domain

import (
	"testing"
	"time"
)

func TestNewConversation(t *testing.T) {
	id := "conv-123"
	conv := NewConversation(id)

	if conv.ID != id {
		t.Errorf("Expected conversation ID %s, got %s", id, conv.ID)
	}

	if len(conv.Messages) != 0 {
		t.Error("Expected empty messages")
	}

	if conv.Temperature != DefaultTemperature {
		t.Errorf("Expected default temperature %f, got %f", DefaultTemperature, conv.Temperature)
	}

	if conv.MaxTokens != DefaultMaxTokens {
		t.Errorf("Expected default max tokens %d, got %d", DefaultMaxTokens, conv.MaxTokens)
	}

	if conv.Metadata == nil {
		t.Error("Expected metadata map to be initialized")
	}
}

func TestConversationAddMessage(t *testing.T) {
	conv := NewConversation("test")
	originalTime := conv.Updated

	// Sleep briefly to ensure time difference
	time.Sleep(10 * time.Millisecond)

	msg := *NewMessage("msg-1", MessageRoleUser, "Hello")
	conv.AddMessage(msg)

	if len(conv.Messages) != 1 {
		t.Error("Failed to add message")
	}

	if conv.Messages[0].ID != "msg-1" {
		t.Error("Message ID mismatch")
	}

	if !conv.Updated.After(originalTime) {
		t.Error("Updated timestamp should be after original")
	}
}

func TestConversationGetLastMessage(t *testing.T) {
	conv := NewConversation("test")

	// Empty conversation
	if conv.GetLastMessage() != nil {
		t.Error("Expected nil for empty conversation")
	}

	// Add messages
	msg1 := *NewMessage("msg-1", MessageRoleUser, "First")
	msg2 := *NewMessage("msg-2", MessageRoleAssistant, "Second")
	conv.AddMessage(msg1)
	conv.AddMessage(msg2)

	last := conv.GetLastMessage()
	if last == nil {
		t.Fatal("Expected to get last message")
	}

	if last.ID != "msg-2" {
		t.Error("Wrong last message returned")
	}
}

func TestConversationMessageCounts(t *testing.T) {
	conv := NewConversation("test")

	// Add different types of messages
	conv.AddMessage(*NewMessage("1", MessageRoleUser, "User 1"))
	conv.AddMessage(*NewMessage("2", MessageRoleAssistant, "Assistant 1"))
	conv.AddMessage(*NewMessage("3", MessageRoleUser, "User 2"))
	conv.AddMessage(*NewMessage("4", MessageRoleSystem, "System"))
	conv.AddMessage(*NewMessage("5", MessageRoleAssistant, "Assistant 2"))

	if conv.GetMessageCount() != 5 {
		t.Errorf("Expected 5 total messages, got %d", conv.GetMessageCount())
	}

	if conv.GetUserMessageCount() != 2 {
		t.Errorf("Expected 2 user messages, got %d", conv.GetUserMessageCount())
	}

	if conv.GetAssistantMessageCount() != 2 {
		t.Errorf("Expected 2 assistant messages, got %d", conv.GetAssistantMessageCount())
	}
}

func TestConversationSetModel(t *testing.T) {
	conv := NewConversation("test")
	originalTime := conv.Updated

	time.Sleep(10 * time.Millisecond)

	conv.SetModel("anthropic", "claude-3")

	if conv.Provider != "anthropic" {
		t.Errorf("Expected provider anthropic, got %s", conv.Provider)
	}

	if conv.Model != "claude-3" {
		t.Errorf("Expected model claude-3, got %s", conv.Model)
	}

	if !conv.Updated.After(originalTime) {
		t.Error("Updated timestamp should be after original")
	}
}

func TestConversationSetParameters(t *testing.T) {
	conv := NewConversation("test")

	conv.SetParameters(0.9, 2000)

	if conv.Temperature != 0.9 {
		t.Errorf("Expected temperature 0.9, got %f", conv.Temperature)
	}

	if conv.MaxTokens != 2000 {
		t.Errorf("Expected max tokens 2000, got %d", conv.MaxTokens)
	}
}

func TestConversationClearMessages(t *testing.T) {
	conv := NewConversation("test")

	// Add some messages
	conv.AddMessage(*NewMessage("1", MessageRoleUser, "Test"))
	conv.AddMessage(*NewMessage("2", MessageRoleAssistant, "Response"))

	if len(conv.Messages) != 2 {
		t.Error("Failed to add messages")
	}

	conv.ClearMessages()

	if len(conv.Messages) != 0 {
		t.Error("Failed to clear messages")
	}

	if !conv.IsEmpty() {
		t.Error("Conversation should be empty after clearing")
	}
}
