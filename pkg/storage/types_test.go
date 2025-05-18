// ABOUTME: Tests for storage types and conversions between domain and storage types
// ABOUTME: Ensures proper JSON marshaling/unmarshaling and type conversion

package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageSession_JSONMarshal(t *testing.T) {
	// Create a domain session
	domainSession := domain.NewSession("test-session-id")
	domainSession.Name = "Test Session"
	domainSession.Tags = []string{"test", "example"}
	domainSession.Config["temperature"] = 0.7
	domainSession.Config["max_tokens"] = 100
	domainSession.Metadata["source"] = "unit-test"
	
	// Set up conversation
	domainSession.Conversation.SetModel("openai", "gpt-4")
	domainSession.Conversation.SetParameters(0.7, 100)
	domainSession.Conversation.SetSystemPrompt("You are a helpful assistant")
	domainSession.Conversation.AddMessage(*domain.NewMessage("msg-1", domain.MessageRoleUser, "Hello"))
	domainSession.Conversation.AddMessage(*domain.NewMessage("msg-2", domain.MessageRoleAssistant, "Hi there!"))
	
	// Convert to storage session
	storageSession := ToStorageSession(domainSession)
	
	// Test JSON marshaling
	data, err := json.Marshal(storageSession)
	require.NoError(t, err)
	
	// Test JSON unmarshaling
	var unmarshaled StorageSession
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	
	// Verify data
	assert.Equal(t, storageSession.ID, unmarshaled.ID)
	assert.Equal(t, storageSession.Name, unmarshaled.Name)
	assert.Equal(t, storageSession.Model, unmarshaled.Model)
	assert.Equal(t, storageSession.Provider, unmarshaled.Provider)
	assert.Equal(t, storageSession.Temperature, unmarshaled.Temperature)
	assert.Equal(t, storageSession.MaxTokens, unmarshaled.MaxTokens)
	assert.Equal(t, storageSession.SystemPrompt, unmarshaled.SystemPrompt)
	assert.Len(t, unmarshaled.Messages, 2)
	assert.Equal(t, storageSession.Tags, unmarshaled.Tags)
	
	// Test conversion back to domain
	convertedBack := ToDomainSession(&unmarshaled)
	assert.NotNil(t, convertedBack)
	assert.Equal(t, domainSession.ID, convertedBack.ID)
	assert.Equal(t, domainSession.Name, convertedBack.Name)
	assert.NotNil(t, convertedBack.Conversation)
	assert.Len(t, convertedBack.Conversation.Messages, 2)
}

func TestToStorageSession_NilInput(t *testing.T) {
	result := ToStorageSession(nil)
	assert.Nil(t, result)
}

func TestToDomainSession_NilInput(t *testing.T) {
	result := ToDomainSession(nil)
	assert.Nil(t, result)
}

func TestDomainMessage_Conversion(t *testing.T) {
	// Create domain message with attachment
	msg := domain.NewMessage("msg-1", domain.MessageRoleUser, "Test message")
	attachment := domain.NewAttachment("att-1", domain.AttachmentTypeImage)
	attachment.Name = "test.jpg"
	attachment.MimeType = "image/jpeg"
	attachment.Content = []byte("image data")
	msg.AddAttachment(*attachment)
	
	// Test that the message works correctly
	assert.Equal(t, "msg-1", msg.ID)
	assert.Equal(t, domain.MessageRoleUser, msg.Role)
	assert.Equal(t, "Test message", msg.Content)
	assert.Len(t, msg.Attachments, 1)
}

func TestStorageSession_WithoutConversation(t *testing.T) {
	// Create a minimal domain session without conversation data
	domainSession := &domain.Session{
		ID:       "test-id",
		Name:     "Test",
		Created:  time.Now(),
		Updated:  time.Now(),
		Tags:     []string{},
		Config:   make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
		// Conversation is nil
	}
	
	// Convert to storage
	storageSession := ToStorageSession(domainSession)
	assert.NotNil(t, storageSession)
	assert.Equal(t, domainSession.ID, storageSession.ID)
	assert.Empty(t, storageSession.Messages)
	assert.Empty(t, storageSession.Model)
	
	// Convert back to domain
	convertedBack := ToDomainSession(storageSession)
	assert.NotNil(t, convertedBack)
	assert.Equal(t, domainSession.ID, convertedBack.ID)
	assert.Nil(t, convertedBack.Conversation)
}

func TestStorageSession_CompleteRoundTrip(t *testing.T) {
	// Create a complete domain session
	session := domain.NewSession("complete-test")
	session.Name = "Complete Test Session"
	session.Tags = []string{"test", "complete", "roundtrip"}
	session.Config["feature"] = "enabled"
	session.Metadata["version"] = "1.0"
	
	// Add conversation with all fields
	session.Conversation.SetModel("anthropic", "claude-3")
	session.Conversation.SetParameters(0.8, 2000)
	session.Conversation.SetSystemPrompt("You are Claude, an AI assistant.")
	
	// Add messages with attachments
	userMsg := domain.NewMessage("msg-1", domain.MessageRoleUser, "Please analyze this image")
	imageAtt := domain.NewAttachment("att-1", domain.AttachmentTypeImage)
	imageAtt.Name = "screenshot.png"
	imageAtt.MimeType = "image/png"
	imageAtt.URL = "https://example.com/image.png"
	userMsg.AddAttachment(*imageAtt)
	session.Conversation.AddMessage(*userMsg)
	
	assistantMsg := domain.NewMessage("msg-2", domain.MessageRoleAssistant, "I can see the image shows...")
	session.Conversation.AddMessage(*assistantMsg)
	
	// Convert to storage format
	storageSession := ToStorageSession(session)
	
	// Simulate JSON persistence
	jsonData, err := json.MarshalIndent(storageSession, "", "  ")
	require.NoError(t, err)
	
	// Parse back from JSON
	var parsedStorage StorageSession
	err = json.Unmarshal(jsonData, &parsedStorage)
	require.NoError(t, err)
	
	// Convert back to domain
	finalSession := ToDomainSession(&parsedStorage)
	
	// Verify everything is preserved
	assert.Equal(t, session.ID, finalSession.ID)
	assert.Equal(t, session.Name, finalSession.Name)
	assert.Equal(t, session.Tags, finalSession.Tags)
	assert.Equal(t, session.Config["feature"], finalSession.Config["feature"])
	assert.Equal(t, session.Metadata["version"], finalSession.Metadata["version"])
	
	// Verify conversation
	assert.NotNil(t, finalSession.Conversation)
	assert.Equal(t, session.Conversation.Model, finalSession.Conversation.Model)
	assert.Equal(t, session.Conversation.Provider, finalSession.Conversation.Provider)
	assert.Equal(t, session.Conversation.Temperature, finalSession.Conversation.Temperature)
	assert.Equal(t, session.Conversation.MaxTokens, finalSession.Conversation.MaxTokens)
	assert.Equal(t, session.Conversation.SystemPrompt, finalSession.Conversation.SystemPrompt)
	
	// Verify messages
	assert.Len(t, finalSession.Conversation.Messages, 2)
	assert.Equal(t, "Please analyze this image", finalSession.Conversation.Messages[0].Content)
	assert.Len(t, finalSession.Conversation.Messages[0].Attachments, 1)
	assert.Equal(t, "screenshot.png", finalSession.Conversation.Messages[0].Attachments[0].Name)
}