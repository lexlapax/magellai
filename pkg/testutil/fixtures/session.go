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
func CreateTestSessionWithMessages(id string, messageCount int) *domain.Session {
	session := CreateTestSession(id)

	// Add system message
	session.Conversation.SetSystemPrompt("You are a helpful test assistant")

	// Add user and assistant messages
	for i := 0; i < messageCount; i++ {
		userMsg := domain.Message{
			ID:        fmt.Sprintf("msg-user-%d", i+1),
			Role:      domain.MessageRoleUser,
			Content:   fmt.Sprintf("Test user message %d", i+1),
			Timestamp: time.Now().Add(time.Duration(i*2) * time.Minute),
			Metadata:  make(map[string]interface{}),
		}
		session.Conversation.AddMessage(userMsg)

		assistantMsg := domain.Message{
			ID:        fmt.Sprintf("msg-assistant-%d", i+1),
			Role:      domain.MessageRoleAssistant,
			Content:   fmt.Sprintf("Test assistant response %d", i+1),
			Timestamp: time.Now().Add(time.Duration(i*2+1) * time.Minute),
			Metadata:  make(map[string]interface{}),
		}
		session.Conversation.AddMessage(assistantMsg)
	}

	return session
}

// CreateTestSessionWithAttachments creates a session with attachments
func CreateTestSessionWithAttachments(id string) *domain.Session {
	session := CreateTestSessionWithMessages(id, 1)

	// Find the last user message
	for i := len(session.Conversation.Messages) - 1; i >= 0; i-- {
		if session.Conversation.Messages[i].Role == domain.MessageRoleUser {
			// Add attachments to the last user message
			attachments := []domain.Attachment{
				{
					ID:       "attach-1",
					Type:     domain.AttachmentTypeText,
					Name:     "test.txt",
					Content:  []byte("Test text content"),
					MimeType: "text/plain",
					Size:     int64(len([]byte("Test text content"))),
				},
				{
					ID:       "attach-2",
					Type:     domain.AttachmentTypeImage,
					Name:     "test.jpg",
					Content:  []byte("fake image content"),
					MimeType: "image/jpeg",
					Size:     int64(len([]byte("fake image content"))),
				},
			}
			session.Conversation.Messages[i].Attachments = attachments
			break
		}
	}

	return session
}

// CreateTestBranchSession creates a session with branches
func CreateTestBranchSession(parentID string, branchName string) *domain.Session {
	branch := CreateTestSession("")
	branch.ParentID = parentID
	branch.BranchName = branchName
	branch.BranchPoint = 2 // Branched at message index 2

	return branch
}

// CreateTestSessionTree creates a tree of sessions for testing branching
func CreateTestSessionTree() (*domain.Session, []*domain.Session) {
	// Create root session
	root := CreateTestSessionWithMessages("root-session", 3)

	// Create branches
	branch1 := CreateTestBranchSession(root.ID, "feature-branch")
	branch1.ID = "branch-1"
	root.AddChild(branch1.ID)

	branch2 := CreateTestBranchSession(root.ID, "experiment-branch")
	branch2.ID = "branch-2"
	root.AddChild(branch2.ID)

	// Create a sub-branch
	subBranch := CreateTestBranchSession(branch1.ID, "sub-feature")
	subBranch.ID = "sub-branch-1"
	branch1.AddChild(subBranch.ID)

	return root, []*domain.Session{branch1, branch2, subBranch}
}

// CreateTestSessionForMerge creates sessions ready for merge testing
func CreateTestSessionForMerge() (*domain.Session, *domain.Session) {
	// Create target session
	target := CreateTestSessionWithMessages("target-session", 3)

	// Create source session with different messages
	source := CreateTestSessionWithMessages("source-session", 2)
	source.Conversation.AddMessage(domain.Message{
		ID:        "unique-msg",
		Role:      domain.MessageRoleUser,
		Content:   "Unique message for merge testing",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	})

	return target, source
}

// CreateMinimalSession creates a minimal session for quick testing
func CreateMinimalSession() *domain.Session {
	return domain.NewSession("minimal-session")
}

// CreateSessionInfo creates a test SessionInfo object
func CreateSessionInfo(id string) *domain.SessionInfo {
	now := time.Now()
	return &domain.SessionInfo{
		ID:           id,
		Name:         "Test Session Info",
		Created:      now.Add(-24 * time.Hour),
		Updated:      now,
		MessageCount: 10,
		Model:        "test/model",
		Provider:     "test",
		Tags:         []string{"test", "info"},
		ParentID:     "",
		BranchName:   "",
		ChildCount:   0,
		IsBranch:     false,
	}
}
