// ABOUTME: Tests for storage types including Session, Message, and Attachment
// ABOUTME: Ensures proper JSON marshaling/unmarshaling and type conversion

package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_JSONMarshal(t *testing.T) {
	session := &Session{
		ID:   "test-session-id",
		Name: "Test Session",
		Messages: []Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Hello",
				Timestamp: time.Now(),
			},
			{
				ID:        "msg-2",
				Role:      "assistant",
				Content:   "Hi there!",
				Timestamp: time.Now(),
			},
		},
		Config: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		},
		Created: time.Now(),
		Updated: time.Now(),
		Tags:    []string{"test", "example"},
		Metadata: map[string]interface{}{
			"source": "unit-test",
		},
		Model:        "gpt-4",
		Provider:     "openai",
		Temperature:  0.7,
		MaxTokens:    100,
		SystemPrompt: "You are a helpful assistant",
	}

	// Test JSON marshaling
	data, err := json.Marshal(session)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON unmarshaling
	var decoded Session
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, session.ID, decoded.ID)
	assert.Equal(t, session.Name, decoded.Name)
	assert.Len(t, decoded.Messages, 2)
	assert.Equal(t, session.Model, decoded.Model)
	assert.Equal(t, session.Provider, decoded.Provider)
	assert.Equal(t, session.Temperature, decoded.Temperature)
	assert.Equal(t, session.MaxTokens, decoded.MaxTokens)
	assert.Equal(t, session.SystemPrompt, decoded.SystemPrompt)
	assert.Equal(t, session.Tags, decoded.Tags)
}

func TestMessage_JSONMarshal(t *testing.T) {
	msg := &Message{
		ID:        "msg-123",
		Role:      "user",
		Content:   "Test message",
		Timestamp: time.Now(),
		Attachments: []Attachment{
			{
				Type:     "image",
				URL:      "https://example.com/image.jpg",
				MimeType: "image/jpeg",
				Name:     "test.jpg",
				Size:     1024,
			},
		},
		Metadata: map[string]interface{}{
			"custom": "value",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON unmarshaling
	var decoded Message
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, msg.ID, decoded.ID)
	assert.Equal(t, msg.Role, decoded.Role)
	assert.Equal(t, msg.Content, decoded.Content)
	assert.Len(t, decoded.Attachments, 1)
	assert.Equal(t, msg.Attachments[0].Type, decoded.Attachments[0].Type)
	assert.Equal(t, msg.Attachments[0].URL, decoded.Attachments[0].URL)
}

func TestAttachment_Types(t *testing.T) {
	tests := []struct {
		name string
		att  Attachment
	}{
		{
			name: "image attachment",
			att: Attachment{
				Type:     "image",
				URL:      "https://example.com/image.jpg",
				MimeType: "image/jpeg",
				Name:     "test.jpg",
				Size:     1024,
			},
		},
		{
			name: "text attachment",
			att: Attachment{
				Type:    "text",
				Content: "This is text content",
				Name:    "text.txt",
			},
		},
		{
			name: "file attachment",
			att: Attachment{
				Type:     "file",
				URL:      "/path/to/file.pdf",
				MimeType: "application/pdf",
				Name:     "document.pdf",
				Size:     2048,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON round trip
			data, err := json.Marshal(tt.att)
			require.NoError(t, err)

			var decoded Attachment
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.att.Type, decoded.Type)
			assert.Equal(t, tt.att.Name, decoded.Name)
			if tt.att.URL != "" {
				assert.Equal(t, tt.att.URL, decoded.URL)
			}
			if tt.att.Content != "" {
				assert.Equal(t, tt.att.Content, decoded.Content)
			}
		})
	}
}

func TestSessionInfo_Conversion(t *testing.T) {
	info := &SessionInfo{
		ID:           "session-123",
		Name:         "Test Session",
		Created:      time.Now(),
		Updated:      time.Now().Add(time.Hour),
		MessageCount: 10,
		Model:        "gpt-4",
		Provider:     "openai",
		Tags:         []string{"test", "example"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(info)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded SessionInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info.ID, decoded.ID)
	assert.Equal(t, info.Name, decoded.Name)
	assert.Equal(t, info.MessageCount, decoded.MessageCount)
	assert.Equal(t, info.Model, decoded.Model)
	assert.Equal(t, info.Provider, decoded.Provider)
	assert.Equal(t, info.Tags, decoded.Tags)
}

func TestSearchResult_Structure(t *testing.T) {
	result := &SearchResult{
		Session: &SessionInfo{
			ID:   "session-123",
			Name: "Test Session",
		},
		Matches: []SearchMatch{
			{
				Type:     "message",
				Role:     "user",
				Content:  "...found text...",
				Context:  "Message 1 (user)",
				Position: 0,
			},
			{
				Type:    "system_prompt",
				Content: "...prompt match...",
				Context: "System Prompt",
			},
		},
	}

	// Verify structure
	assert.NotNil(t, result.Session)
	assert.Len(t, result.Matches, 2)
	assert.Equal(t, "message", result.Matches[0].Type)
	assert.Equal(t, "system_prompt", result.Matches[1].Type)
}

func TestExportFormat_Constants(t *testing.T) {
	// Verify export format constants
	assert.Equal(t, ExportFormat("json"), ExportFormatJSON)
	assert.Equal(t, ExportFormat("markdown"), ExportFormatMarkdown)
	assert.Equal(t, ExportFormat("text"), ExportFormatText)
}
