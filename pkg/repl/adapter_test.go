// ABOUTME: Tests for adapter functions that convert between REPL and storage types
// ABOUTME: Ensures proper conversion logic and handles nil values correctly

package repl

import (
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/lexlapax/magellai/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToStorageSession(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		result := ToStorageSession(nil)
		assert.Nil(t, result)
	})

	t.Run("session without conversation", func(t *testing.T) {
		replSession := &Session{
			ID:       "test-id",
			Name:     "Test Session",
			Config:   map[string]interface{}{"key": "value"},
			Created:  time.Now(),
			Updated:  time.Now(),
			Tags:     []string{"tag1", "tag2"},
			Metadata: map[string]interface{}{"meta": "data"},
		}

		storageSession := ToStorageSession(replSession)
		require.NotNil(t, storageSession)
		assert.Equal(t, replSession.ID, storageSession.ID)
		assert.Equal(t, replSession.Name, storageSession.Name)
		assert.Equal(t, replSession.Config, storageSession.Config)
		assert.Equal(t, replSession.Created, storageSession.Created)
		assert.Equal(t, replSession.Updated, storageSession.Updated)
		assert.Equal(t, replSession.Tags, storageSession.Tags)
		assert.Equal(t, replSession.Metadata, storageSession.Metadata)
		assert.Empty(t, storageSession.Messages)
	})

	t.Run("session with conversation", func(t *testing.T) {
		now := time.Now()
		replSession := &Session{
			ID:   "test-id",
			Name: "Test Session",
			Conversation: &Conversation{
				ID:           "conv-id",
				Model:        "gpt-4",
				Provider:     "openai",
				Temperature:  0.7,
				MaxTokens:    2000,
				SystemPrompt: "You are helpful",
				Created:      now,
				Updated:      now,
				Messages: []Message{
					{
						ID:        "msg-1",
						Role:      "user",
						Content:   "Hello",
						Timestamp: now,
						Attachments: []llm.Attachment{
							{
								Type:     llm.AttachmentTypeImage,
								MimeType: "image/png",
								FilePath: "image.png",
								Content:  "image data",
							},
						},
					},
					{
						ID:        "msg-2",
						Role:      "assistant",
						Content:   "Hi there!",
						Timestamp: now,
					},
				},
			},
			Created: now,
			Updated: now,
		}

		storageSession := ToStorageSession(replSession)
		require.NotNil(t, storageSession)
		assert.Equal(t, replSession.ID, storageSession.ID)
		assert.Equal(t, replSession.Name, storageSession.Name)
		assert.Equal(t, replSession.Conversation.Model, storageSession.Model)
		assert.Equal(t, replSession.Conversation.Provider, storageSession.Provider)
		assert.Equal(t, replSession.Conversation.Temperature, storageSession.Temperature)
		assert.Equal(t, replSession.Conversation.MaxTokens, storageSession.MaxTokens)
		assert.Equal(t, replSession.Conversation.SystemPrompt, storageSession.SystemPrompt)
		assert.Len(t, storageSession.Messages, 2)

		// Check message conversion
		assert.Equal(t, "msg-1", storageSession.Messages[0].ID)
		assert.Equal(t, "user", storageSession.Messages[0].Role)
		assert.Equal(t, "Hello", storageSession.Messages[0].Content)
		assert.Len(t, storageSession.Messages[0].Attachments, 1)
		assert.Equal(t, "image", storageSession.Messages[0].Attachments[0].Type)
		assert.Equal(t, "image.png", storageSession.Messages[0].Attachments[0].Name)
	})
}

func TestFromStorageSession(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		result := FromStorageSession(nil)
		assert.Nil(t, result)
	})

	t.Run("complete session", func(t *testing.T) {
		now := time.Now()
		storageSession := &storage.Session{
			ID:           "test-id",
			Name:         "Test Session",
			Model:        "gpt-4",
			Provider:     "openai",
			Temperature:  0.7,
			MaxTokens:    2000,
			SystemPrompt: "You are helpful",
			Messages: []storage.Message{
				{
					ID:        "msg-1",
					Role:      "user",
					Content:   "Hello",
					Timestamp: now,
					Attachments: []storage.Attachment{
						{
							Type:     "image",
							MimeType: "image/png",
							Name:     "image.png",
							Content:  "image data",
						},
					},
				},
			},
			Config:   map[string]interface{}{"key": "value"},
			Created:  now,
			Updated:  now,
			Tags:     []string{"tag1", "tag2"},
			Metadata: map[string]interface{}{"meta": "data"},
		}

		replSession := FromStorageSession(storageSession)
		require.NotNil(t, replSession)
		assert.Equal(t, storageSession.ID, replSession.ID)
		assert.Equal(t, storageSession.Name, replSession.Name)
		assert.NotNil(t, replSession.Conversation)
		assert.Equal(t, storageSession.Model, replSession.Conversation.Model)
		assert.Equal(t, storageSession.Provider, replSession.Conversation.Provider)
		assert.Equal(t, storageSession.Temperature, replSession.Conversation.Temperature)
		assert.Equal(t, storageSession.MaxTokens, replSession.Conversation.MaxTokens)
		assert.Equal(t, storageSession.SystemPrompt, replSession.Conversation.SystemPrompt)
		assert.Equal(t, storageSession.Config, replSession.Config)
		assert.Equal(t, storageSession.Tags, replSession.Tags)
		assert.Equal(t, storageSession.Metadata, replSession.Metadata)

		// Check message conversion
		assert.Len(t, replSession.Conversation.Messages, 1)
		assert.Equal(t, "msg-1", replSession.Conversation.Messages[0].ID)
		assert.Equal(t, "user", replSession.Conversation.Messages[0].Role)
		assert.Len(t, replSession.Conversation.Messages[0].Attachments, 1)
		assert.Equal(t, llm.AttachmentTypeImage, replSession.Conversation.Messages[0].Attachments[0].Type)
		assert.Equal(t, "image.png", replSession.Conversation.Messages[0].Attachments[0].FilePath)
	})
}

func TestToStorageMessage(t *testing.T) {
	t.Run("message without attachments", func(t *testing.T) {
		now := time.Now()
		replMsg := Message{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Timestamp: now,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		storageMsg := ToStorageMessage(replMsg)
		assert.Equal(t, replMsg.ID, storageMsg.ID)
		assert.Equal(t, replMsg.Role, storageMsg.Role)
		assert.Equal(t, replMsg.Content, storageMsg.Content)
		assert.Equal(t, replMsg.Timestamp, storageMsg.Timestamp)
		assert.Equal(t, replMsg.Metadata, storageMsg.Metadata)
		assert.Empty(t, storageMsg.Attachments)
	})

	t.Run("message with attachments", func(t *testing.T) {
		now := time.Now()
		replMsg := Message{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Check this image",
			Timestamp: now,
			Attachments: []llm.Attachment{
				{
					Type:     llm.AttachmentTypeImage,
					MimeType: "image/jpeg",
					FilePath: "photo.jpg",
					Content:  "image data",
				},
				{
					Type:     llm.AttachmentTypeFile,
					MimeType: "application/pdf",
					FilePath: "document.pdf",
					Content:  "pdf data",
				},
			},
		}

		storageMsg := ToStorageMessage(replMsg)
		assert.Len(t, storageMsg.Attachments, 2)
		assert.Equal(t, "image", storageMsg.Attachments[0].Type)
		assert.Equal(t, "image/jpeg", storageMsg.Attachments[0].MimeType)
		assert.Equal(t, "photo.jpg", storageMsg.Attachments[0].Name)
		assert.Equal(t, "photo.jpg", storageMsg.Attachments[0].URL)
		assert.Equal(t, "file", storageMsg.Attachments[1].Type)
		assert.Equal(t, "application/pdf", storageMsg.Attachments[1].MimeType)
		assert.Equal(t, "document.pdf", storageMsg.Attachments[1].Name)
	})
}

func TestFromStorageMessage(t *testing.T) {
	t.Run("message without attachments", func(t *testing.T) {
		now := time.Now()
		storageMsg := storage.Message{
			ID:        "msg-1",
			Role:      "assistant",
			Content:   "Hello there!",
			Timestamp: now,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		replMsg := FromStorageMessage(storageMsg)
		assert.Equal(t, storageMsg.ID, replMsg.ID)
		assert.Equal(t, storageMsg.Role, replMsg.Role)
		assert.Equal(t, storageMsg.Content, replMsg.Content)
		assert.Equal(t, storageMsg.Timestamp, replMsg.Timestamp)
		assert.Equal(t, storageMsg.Metadata, replMsg.Metadata)
		assert.Empty(t, replMsg.Attachments)
	})

	t.Run("message with attachments", func(t *testing.T) {
		now := time.Now()
		storageMsg := storage.Message{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Here's a video",
			Timestamp: now,
			Attachments: []storage.Attachment{
				{
					Type:     "video",
					MimeType: "video/mp4",
					Name:     "video.mp4",
					URL:      "video.mp4",
					Content:  "video data",
				},
			},
		}

		replMsg := FromStorageMessage(storageMsg)
		assert.Len(t, replMsg.Attachments, 1)
		assert.Equal(t, llm.AttachmentTypeVideo, replMsg.Attachments[0].Type)
		assert.Equal(t, "video/mp4", replMsg.Attachments[0].MimeType)
		assert.Equal(t, "video.mp4", replMsg.Attachments[0].FilePath)
		assert.Equal(t, "video data", replMsg.Attachments[0].Content)
	})
}

func TestSearchResultConversion(t *testing.T) {
	t.Run("ToStorageSearchResult nil", func(t *testing.T) {
		result := ToStorageSearchResult(nil)
		assert.Nil(t, result)
	})

	t.Run("FromStorageSearchResult nil", func(t *testing.T) {
		result := FromStorageSearchResult(nil)
		assert.Nil(t, result)
	})

	t.Run("search result conversion", func(t *testing.T) {
		now := time.Now()
		replResult := &SearchResult{
			Session: &SessionInfo{
				ID:           "session-1",
				Name:         "Test Session",
				Created:      now,
				Updated:      now,
				MessageCount: 5,
				Tags:         []string{"tag1"},
			},
			Matches: []SearchMatch{
				{
					Type:     "message",
					Role:     "user",
					Content:  "test content",
					Context:  "...test content...",
					Position: 10,
				},
			},
		}

		// Convert to storage and back
		storageResult := ToStorageSearchResult(replResult)
		require.NotNil(t, storageResult)
		assert.Equal(t, replResult.Session.ID, storageResult.Session.ID)
		assert.Len(t, storageResult.Matches, 1)
		assert.Equal(t, replResult.Matches[0].Type, storageResult.Matches[0].Type)

		// Convert back
		convertedResult := FromStorageSearchResult(storageResult)
		require.NotNil(t, convertedResult)
		assert.Equal(t, replResult.Session.ID, convertedResult.Session.ID)
		assert.Equal(t, replResult.Session.Name, convertedResult.Session.Name)
		assert.Len(t, convertedResult.Matches, 1)
		assert.Equal(t, replResult.Matches[0].Content, convertedResult.Matches[0].Content)
	})
}

func TestSessionInfoConversion(t *testing.T) {
	t.Run("ToStorageSessionInfo nil", func(t *testing.T) {
		result := ToStorageSessionInfo(nil)
		assert.Nil(t, result)
	})

	t.Run("FromStorageSessionInfo nil", func(t *testing.T) {
		result := FromStorageSessionInfo(nil)
		assert.Nil(t, result)
	})

	t.Run("session info conversion", func(t *testing.T) {
		now := time.Now()
		replInfo := &SessionInfo{
			ID:           "session-1",
			Name:         "Test Session",
			Created:      now,
			Updated:      now,
			MessageCount: 10,
			Tags:         []string{"work", "project"},
		}

		// Convert to storage
		storageInfo := ToStorageSessionInfo(replInfo)
		require.NotNil(t, storageInfo)
		assert.Equal(t, replInfo.ID, storageInfo.ID)
		assert.Equal(t, replInfo.Name, storageInfo.Name)
		assert.Equal(t, replInfo.Created, storageInfo.Created)
		assert.Equal(t, replInfo.Updated, storageInfo.Updated)
		assert.Equal(t, replInfo.MessageCount, storageInfo.MessageCount)
		assert.Equal(t, replInfo.Tags, storageInfo.Tags)

		// Convert back
		convertedInfo := FromStorageSessionInfo(storageInfo)
		require.NotNil(t, convertedInfo)
		assert.Equal(t, replInfo.ID, convertedInfo.ID)
		assert.Equal(t, replInfo.Name, convertedInfo.Name)
		assert.Equal(t, replInfo.MessageCount, convertedInfo.MessageCount)
		assert.Equal(t, replInfo.Tags, convertedInfo.Tags)
	})
}

func TestSearchMatchConversion(t *testing.T) {
	t.Run("empty matches", func(t *testing.T) {
		storageMatches := ToStorageSearchMatches([]SearchMatch{})
		assert.Empty(t, storageMatches)

		replMatches := FromStorageSearchMatches([]storage.SearchMatch{})
		assert.Empty(t, replMatches)
	})

	t.Run("multiple matches", func(t *testing.T) {
		replMatches := []SearchMatch{
			{
				Type:     "message",
				Role:     "user",
				Content:  "first match",
				Context:  "...first match...",
				Position: 0,
			},
			{
				Type:     "name",
				Role:     "",
				Content:  "session name",
				Context:  "Test session name",
				Position: 5,
			},
		}

		// Convert to storage
		storageMatches := ToStorageSearchMatches(replMatches)
		assert.Len(t, storageMatches, 2)
		assert.Equal(t, replMatches[0].Type, storageMatches[0].Type)
		assert.Equal(t, replMatches[1].Content, storageMatches[1].Content)

		// Convert back
		convertedMatches := FromStorageSearchMatches(storageMatches)
		assert.Len(t, convertedMatches, 2)
		assert.Equal(t, replMatches[0].Type, convertedMatches[0].Type)
		assert.Equal(t, replMatches[0].Role, convertedMatches[0].Role)
		assert.Equal(t, replMatches[1].Context, convertedMatches[1].Context)
	})
}

func TestAttachmentConversion(t *testing.T) {
	t.Run("llmAttachmentToStorage", func(t *testing.T) {
		llmAtt := llm.Attachment{
			Type:     llm.AttachmentTypeAudio,
			MimeType: "audio/mp3",
			FilePath: "sound.mp3",
			Content:  "audio data",
		}

		storageAtt := llmAttachmentToStorage(llmAtt)
		assert.Equal(t, "audio", storageAtt.Type)
		assert.Equal(t, "audio/mp3", storageAtt.MimeType)
		assert.Equal(t, "sound.mp3", storageAtt.Name)
		assert.Equal(t, "sound.mp3", storageAtt.URL)
		assert.Equal(t, "audio data", storageAtt.Content)
		assert.NotNil(t, storageAtt.Metadata)
		assert.Empty(t, storageAtt.Metadata)
	})

	t.Run("storageAttachmentToLLM", func(t *testing.T) {
		storageAtt := storage.Attachment{
			Type:     "file",
			MimeType: "text/plain",
			Name:     "document.txt",
			URL:      "http://example.com/document.txt",
			Content:  "text content",
			Metadata: map[string]interface{}{"key": "value"},
		}

		llmAtt := storageAttachmentToLLM(storageAtt)
		assert.Equal(t, llm.AttachmentTypeFile, llmAtt.Type)
		assert.Equal(t, "text/plain", llmAtt.MimeType)
		assert.Equal(t, "document.txt", llmAtt.FilePath)
		assert.Equal(t, "text content", llmAtt.Content)
	})
}
