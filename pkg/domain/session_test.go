package domain

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	id := "test-session-123"
	session := NewSession(id)
	
	if session.ID != id {
		t.Errorf("Expected session ID %s, got %s", id, session.ID)
	}
	
	if session.Conversation == nil {
		t.Error("Expected conversation to be initialized")
	}
	
	if session.Conversation.ID != id {
		t.Errorf("Expected conversation ID %s, got %s", id, session.Conversation.ID)
	}
	
	if session.Config == nil {
		t.Error("Expected config map to be initialized")
	}
	
	if session.Metadata == nil {
		t.Error("Expected metadata map to be initialized")
	}
	
	if len(session.Tags) != 0 {
		t.Error("Expected tags to be empty")
	}
}

func TestSessionAddTag(t *testing.T) {
	session := NewSession("test")
	
	// Add first tag
	session.AddTag("important")
	if len(session.Tags) != 1 || session.Tags[0] != "important" {
		t.Error("Failed to add first tag")
	}
	
	// Add duplicate tag (should not be added)
	session.AddTag("important")
	if len(session.Tags) != 1 {
		t.Error("Duplicate tag should not be added")
	}
	
	// Add second tag
	session.AddTag("review")
	if len(session.Tags) != 2 {
		t.Error("Failed to add second tag")
	}
}

func TestSessionRemoveTag(t *testing.T) {
	session := NewSession("test")
	session.Tags = []string{"important", "review", "archive"}
	
	session.RemoveTag("review")
	
	if len(session.Tags) != 2 {
		t.Error("Failed to remove tag")
	}
	
	for _, tag := range session.Tags {
		if tag == "review" {
			t.Error("Tag was not removed")
		}
	}
}

func TestSessionToSessionInfo(t *testing.T) {
	session := NewSession("test-session")
	session.Name = "Test Session"
	session.Tags = []string{"test", "demo"}
	
	// Add some messages
	session.Conversation.AddMessage(*NewMessage("msg1", MessageRoleUser, "Hello"))
	session.Conversation.AddMessage(*NewMessage("msg2", MessageRoleAssistant, "Hi there"))
	session.Conversation.SetModel("openai", "gpt-4")
	
	info := session.ToSessionInfo()
	
	if info.ID != session.ID {
		t.Error("SessionInfo ID mismatch")
	}
	
	if info.Name != session.Name {
		t.Error("SessionInfo Name mismatch")
	}
	
	if info.MessageCount != 2 {
		t.Errorf("Expected 2 messages, got %d", info.MessageCount)
	}
	
	if info.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", info.Model)
	}
	
	if info.Provider != "openai" {
		t.Errorf("Expected provider openai, got %s", info.Provider)
	}
}

func TestSessionUpdateTimestamp(t *testing.T) {
	session := NewSession("test")
	originalTime := session.Updated
	
	// Sleep briefly to ensure time difference
	time.Sleep(10 * time.Millisecond)
	
	session.UpdateTimestamp()
	
	if !session.Updated.After(originalTime) {
		t.Error("Updated timestamp should be after original")
	}
}