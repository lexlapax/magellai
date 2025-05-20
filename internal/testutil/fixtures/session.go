// ABOUTME: Test fixtures for session objects
// ABOUTME: Provides reusable test data for session-related tests

package fixtures

import (
	"fmt"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
)

// CreateTestSession creates a test session with default values
func CreateTestSession(id string) *domain.Session {
	if id == "" {
		id = "test-session-1"
	}

	now := time.Now()
	session := domain.NewSession(id)
	session.Name = "Test Session"
	session.Created = now
	session.Updated = now
	session.Tags = []string{"test", "fixture"}
	session.Metadata = map[string]interface{}{
		"test":    true,
		"purpose": "testing",
	}

	// Set conversation properties
	session.Conversation.Model = "test/model"
	session.Conversation.Provider = "test"
	session.Conversation.Temperature = 0.7
	session.Conversation.MaxTokens = 100

	return session
}

// CreateTestSessionWithMessages creates a session with test messages
func CreateTestSessionWithMessages(id string, messages []domain.Message) *domain.Session {
	session := CreateTestSession(id)

	// Set system prompt
	session.Conversation.SetSystemPrompt("You are a helpful test assistant")

	// Add provided messages
	for _, msg := range messages {
		session.Conversation.AddMessage(msg)
	}

	// If no messages were provided, add some default ones
	if len(messages) == 0 {
		// Add a few test messages
		userMsg := domain.Message{
			ID:        "msg-user-1",
			Role:      domain.MessageRoleUser,
			Content:   "Hello from test",
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
		session.Conversation.AddMessage(userMsg)

		assistantMsg := domain.Message{
			ID:        "msg-assistant-1",
			Role:      domain.MessageRoleAssistant,
			Content:   "Hello! This is a test response.",
			Timestamp: time.Now().Add(time.Minute),
			Metadata:  make(map[string]interface{}),
		}
		session.Conversation.AddMessage(assistantMsg)
	}

	return session
}

// CreateSessionFamily creates a set of related sessions for branch/merge testing
func CreateSessionFamily() map[string]*domain.Session {
	now := time.Now()

	// Create parent session
	parent := CreateTestSession("parent-session")

	// Add some messages to parent
	parent.Conversation.Messages = []domain.Message{
		{
			ID:        "msg-1",
			Role:      domain.MessageRoleUser,
			Content:   "Hello from the parent session",
			Timestamp: now.Add(-2 * time.Hour),
			Metadata:  make(map[string]interface{}),
		},
		{
			ID:        "msg-2",
			Role:      domain.MessageRoleAssistant,
			Content:   "Hello! How can I help you today?",
			Timestamp: now.Add(-1*time.Hour - 55*time.Minute),
			Metadata:  make(map[string]interface{}),
		},
		{
			ID:        "msg-3",
			Role:      domain.MessageRoleUser,
			Content:   "Tell me about branch testing",
			Timestamp: now.Add(-1*time.Hour - 50*time.Minute),
			Metadata:  make(map[string]interface{}),
		},
		{
			ID:        "msg-4",
			Role:      domain.MessageRoleAssistant,
			Content:   "Branch testing is a way to create parallel sessions from an existing one.",
			Timestamp: now.Add(-1*time.Hour - 45*time.Minute),
			Metadata:  make(map[string]interface{}),
		},
	}

	// Create branch A from parent (branches after message 2)
	branchA, err := parent.CreateBranch("branch-a", "Branch A", 2)
	if err != nil {
		panic(fmt.Sprintf("Failed to create branch A: %v", err))
	}

	// Add branch A specific messages
	branchA.Conversation.AddMessage(domain.Message{
		ID:        "msg-a-1",
		Role:      domain.MessageRoleUser,
		Content:   "Let's explore branch A",
		Timestamp: now.Add(-1 * time.Hour),
		Metadata:  make(map[string]interface{}),
	})

	branchA.Conversation.AddMessage(domain.Message{
		ID:        "msg-a-2",
		Role:      domain.MessageRoleAssistant,
		Content:   "This is branch A content",
		Timestamp: now.Add(-55 * time.Minute),
		Metadata:  make(map[string]interface{}),
	})

	// Create branch B from parent (branches after message 3)
	branchB, err := parent.CreateBranch("branch-b", "Branch B", 3)
	if err != nil {
		panic(fmt.Sprintf("Failed to create branch B: %v", err))
	}

	// Add branch B specific messages
	branchB.Conversation.AddMessage(domain.Message{
		ID:        "msg-b-1",
		Role:      domain.MessageRoleUser,
		Content:   "Let's explore branch B",
		Timestamp: now.Add(-40 * time.Minute),
		Metadata:  make(map[string]interface{}),
	})

	branchB.Conversation.AddMessage(domain.Message{
		ID:        "msg-b-2",
		Role:      domain.MessageRoleAssistant,
		Content:   "This is branch B content",
		Timestamp: now.Add(-35 * time.Minute),
		Metadata:  make(map[string]interface{}),
	})

	// Return all sessions
	return map[string]*domain.Session{
		"parent-session": parent,
		"branch-a":       branchA,
		"branch-b":       branchB,
	}
}
