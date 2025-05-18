// ABOUTME: Tests for adapter functions that convert between REPL and domain types
// ABOUTME: Ensures proper conversion logic and handles nil values correctly

package repl

import (
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToDomainSession(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		result := ToDomainSession(nil)
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

		domainSession := ToDomainSession(replSession)
		require.NotNil(t, domainSession)
		assert.Equal(t, replSession.ID, domainSession.ID)
		assert.Equal(t, replSession.Name, domainSession.Name)
		assert.Equal(t, replSession.Config, domainSession.Config)
		assert.Equal(t, replSession.Created, domainSession.Created)
		assert.Equal(t, replSession.Updated, domainSession.Updated)
		assert.Equal(t, replSession.Tags, domainSession.Tags)
		assert.Equal(t, replSession.Metadata, domainSession.Metadata)
		assert.Nil(t, domainSession.Conversation)
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

		domainSession := ToDomainSession(replSession)
		require.NotNil(t, domainSession)
		assert.Equal(t, replSession.ID, domainSession.ID)
		assert.Equal(t, replSession.Name, domainSession.Name)
		require.NotNil(t, domainSession.Conversation)
		assert.Equal(t, replSession.Conversation.Model, domainSession.Conversation.Model)
		assert.Equal(t, replSession.Conversation.Provider, domainSession.Conversation.Provider)
		assert.Equal(t, replSession.Conversation.Temperature, domainSession.Conversation.Temperature)
		assert.Equal(t, replSession.Conversation.MaxTokens, domainSession.Conversation.MaxTokens)
		assert.Equal(t, replSession.Conversation.SystemPrompt, domainSession.Conversation.SystemPrompt)
		assert.Len(t, domainSession.Conversation.Messages, 2)

		// Check message conversion
		assert.Equal(t, "msg-1", domainSession.Conversation.Messages[0].ID)
		assert.Equal(t, domain.MessageRoleUser, domainSession.Conversation.Messages[0].Role)
		assert.Equal(t, "Hello", domainSession.Conversation.Messages[0].Content)
		assert.Len(t, domainSession.Conversation.Messages[0].Attachments, 1)
		assert.Equal(t, domain.AttachmentTypeImage, domainSession.Conversation.Messages[0].Attachments[0].Type)
		assert.Equal(t, "image.png", domainSession.Conversation.Messages[0].Attachments[0].Name)
	})
}

func TestFromDomainSession(t *testing.T) {
	t.Run("nil session", func(t *testing.T) {
		result := FromDomainSession(nil)
		assert.Nil(t, result)
	})

	t.Run("complete session", func(t *testing.T) {
		now := time.Now()
		domainSession := &domain.Session{
			ID:   "test-id",
			Name: "Test Session",
			Conversation: &domain.Conversation{
				ID:           "conv-id",
				Model:        "gpt-4",
				Provider:     "openai",
				Temperature:  0.7,
				MaxTokens:    2000,
				SystemPrompt: "You are helpful",
				Created:      now,
				Updated:      now,
				Messages: []domain.Message{
					{
						ID:        "msg-1",
						Role:      domain.MessageRoleUser,
						Content:   "Hello",
						Timestamp: now,
						Attachments: []domain.Attachment{
							{
								Type:     domain.AttachmentTypeImage,
								MimeType: "image/png",
								Name:     "image.png",
								Content:  []byte("image data"),
							},
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

		replSession := FromDomainSession(domainSession)
		require.NotNil(t, replSession)
		assert.Equal(t, domainSession.ID, replSession.ID)
		assert.Equal(t, domainSession.Name, replSession.Name)
		assert.NotNil(t, replSession.Conversation)
		assert.Equal(t, domainSession.Conversation.Model, replSession.Conversation.Model)
		assert.Equal(t, domainSession.Conversation.Provider, replSession.Conversation.Provider)
		assert.Equal(t, domainSession.Conversation.Temperature, replSession.Conversation.Temperature)
		assert.Equal(t, domainSession.Conversation.MaxTokens, replSession.Conversation.MaxTokens)
		assert.Equal(t, domainSession.Conversation.SystemPrompt, replSession.Conversation.SystemPrompt)
		assert.Equal(t, domainSession.Config, replSession.Config)
		assert.Equal(t, domainSession.Tags, replSession.Tags)
		assert.Equal(t, domainSession.Metadata, replSession.Metadata)

		// Check message conversion
		assert.Len(t, replSession.Conversation.Messages, 1)
		assert.Equal(t, "msg-1", replSession.Conversation.Messages[0].ID)
		assert.Equal(t, "user", replSession.Conversation.Messages[0].Role)
		assert.Len(t, replSession.Conversation.Messages[0].Attachments, 1)
		assert.Equal(t, llm.AttachmentTypeImage, replSession.Conversation.Messages[0].Attachments[0].Type)
		assert.Equal(t, "image.png", replSession.Conversation.Messages[0].Attachments[0].FilePath)
	})
}

func TestToDomainMessage(t *testing.T) {
	t.Run("message without attachments", func(t *testing.T) {
		now := time.Now()
		replMsg := Message{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Timestamp: now,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		domainMsg := ToDomainMessage(replMsg)
		assert.Equal(t, replMsg.ID, domainMsg.ID)
		assert.Equal(t, domain.MessageRoleUser, domainMsg.Role)
		assert.Equal(t, replMsg.Content, domainMsg.Content)
		assert.Equal(t, replMsg.Timestamp, domainMsg.Timestamp)
		assert.Equal(t, replMsg.Metadata, domainMsg.Metadata)
		assert.Empty(t, domainMsg.Attachments)
	})

	t.Run("message with attachments", func(t *testing.T) {
		now := time.Now()
		replMsg := Message{
			ID:        "msg-1",
			Role:      "assistant",
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

		domainMsg := ToDomainMessage(replMsg)
		assert.Len(t, domainMsg.Attachments, 2)
		assert.Equal(t, domain.AttachmentTypeImage, domainMsg.Attachments[0].Type)
		assert.Equal(t, "image/jpeg", domainMsg.Attachments[0].MimeType)
		assert.Equal(t, "photo.jpg", domainMsg.Attachments[0].Name)
		assert.Equal(t, domain.AttachmentTypeFile, domainMsg.Attachments[1].Type)
		assert.Equal(t, "application/pdf", domainMsg.Attachments[1].MimeType)
		assert.Equal(t, "document.pdf", domainMsg.Attachments[1].Name)
	})
}

func TestFromDomainMessage(t *testing.T) {
	t.Run("message without attachments", func(t *testing.T) {
		now := time.Now()
		domainMsg := domain.Message{
			ID:        "msg-1",
			Role:      domain.MessageRoleAssistant,
			Content:   "Hello there!",
			Timestamp: now,
			Metadata:  map[string]interface{}{"key": "value"},
		}

		replMsg := FromDomainMessage(domainMsg)
		assert.Equal(t, domainMsg.ID, replMsg.ID)
		assert.Equal(t, "assistant", replMsg.Role)
		assert.Equal(t, domainMsg.Content, replMsg.Content)
		assert.Equal(t, domainMsg.Timestamp, replMsg.Timestamp)
		assert.Equal(t, domainMsg.Metadata, replMsg.Metadata)
		assert.Empty(t, replMsg.Attachments)
	})

	t.Run("message with attachments", func(t *testing.T) {
		now := time.Now()
		domainMsg := domain.Message{
			ID:        "msg-1",
			Role:      domain.MessageRoleUser,
			Content:   "Here's a video",
			Timestamp: now,
			Attachments: []domain.Attachment{
				{
					Type:     domain.AttachmentTypeVideo,
					MimeType: "video/mp4",
					Name:     "video.mp4",
					URL:      "video.mp4",
					Content:  []byte("video data"),
				},
			},
		}

		replMsg := FromDomainMessage(domainMsg)
		assert.Len(t, replMsg.Attachments, 1)
		assert.Equal(t, llm.AttachmentTypeVideo, replMsg.Attachments[0].Type)
		assert.Equal(t, "video/mp4", replMsg.Attachments[0].MimeType)
		assert.Equal(t, "video.mp4", replMsg.Attachments[0].FilePath)
		assert.Equal(t, "video data", replMsg.Attachments[0].Content)
	})
}

func TestSearchResultConversion(t *testing.T) {
	t.Run("ToDomainSearchResult nil", func(t *testing.T) {
		result := ToDomainSearchResult(nil)
		assert.Nil(t, result)
	})

	t.Run("FromDomainSearchResult nil", func(t *testing.T) {
		result := FromDomainSearchResult(nil)
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

		// Convert to domain and back
		domainResult := ToDomainSearchResult(replResult)
		require.NotNil(t, domainResult)
		assert.Equal(t, replResult.Session.ID, domainResult.Session.ID)
		assert.Equal(t, 1, domainResult.GetMatchCount())
		matches := domainResult.Matches
		assert.Equal(t, replResult.Matches[0].Type, matches[0].Type)

		// Convert back
		convertedResult := FromDomainSearchResult(domainResult)
		require.NotNil(t, convertedResult)
		assert.Equal(t, replResult.Session.ID, convertedResult.Session.ID)
		assert.Equal(t, replResult.Session.Name, convertedResult.Session.Name)
		assert.Len(t, convertedResult.Matches, 1)
		assert.Equal(t, replResult.Matches[0].Content, convertedResult.Matches[0].Content)
	})
}

func TestSessionInfoConversion(t *testing.T) {
	t.Run("ToDomainSessionInfo nil", func(t *testing.T) {
		result := ToDomainSessionInfo(nil)
		assert.Nil(t, result)
	})

	t.Run("FromDomainSessionInfo nil", func(t *testing.T) {
		result := FromDomainSessionInfo(nil)
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

		// Convert to domain
		domainInfo := ToDomainSessionInfo(replInfo)
		require.NotNil(t, domainInfo)
		assert.Equal(t, replInfo.ID, domainInfo.ID)
		assert.Equal(t, replInfo.Name, domainInfo.Name)
		assert.Equal(t, replInfo.Created, domainInfo.Created)
		assert.Equal(t, replInfo.Updated, domainInfo.Updated)
		assert.Equal(t, replInfo.MessageCount, domainInfo.MessageCount)
		assert.Equal(t, replInfo.Tags, domainInfo.Tags)

		// Convert back
		convertedInfo := FromDomainSessionInfo(domainInfo)
		require.NotNil(t, convertedInfo)
		assert.Equal(t, replInfo.ID, convertedInfo.ID)
		assert.Equal(t, replInfo.Name, convertedInfo.Name)
		assert.Equal(t, replInfo.MessageCount, convertedInfo.MessageCount)
		assert.Equal(t, replInfo.Tags, convertedInfo.Tags)
	})
}

func TestSearchMatchConversion(t *testing.T) {
	t.Run("empty matches", func(t *testing.T) {
		domainMatches := make([]domain.SearchMatch, 0)
		for _, match := range []SearchMatch{} {
			domainMatchValue := ToDomainSearchMatch(match)
			domainMatches = append(domainMatches, domainMatchValue)
		}
		assert.Empty(t, domainMatches)

		replMatches := FromDomainSearchMatches([]domain.SearchMatch{})
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

		// Convert to domain
		domainMatches := make([]domain.SearchMatch, len(replMatches))
		for i, match := range replMatches {
			domainMatches[i] = ToDomainSearchMatch(match)
		}
		assert.Len(t, domainMatches, 2)
		assert.Equal(t, replMatches[0].Type, domainMatches[0].Type)
		assert.Equal(t, replMatches[1].Content, domainMatches[1].Content)

		// Convert back
		convertedMatches := FromDomainSearchMatches(domainMatches)
		assert.Len(t, convertedMatches, 2)
		assert.Equal(t, replMatches[0].Type, convertedMatches[0].Type)
		assert.Equal(t, replMatches[0].Role, convertedMatches[0].Role)
		assert.Equal(t, replMatches[1].Context, convertedMatches[1].Context)
	})
}

func TestAttachmentConversion(t *testing.T) {
	t.Run("llmAttachmentToDomain", func(t *testing.T) {
		llmAtt := llm.Attachment{
			Type:     llm.AttachmentTypeAudio,
			MimeType: "audio/mp3",
			FilePath: "sound.mp3",
			Content:  "audio data",
		}

		domainAtt := llmAttachmentToDomain(llmAtt)
		assert.Equal(t, domain.AttachmentTypeAudio, domainAtt.Type)
		assert.Equal(t, "audio/mp3", domainAtt.MimeType)
		assert.Equal(t, "sound.mp3", domainAtt.Name)
		assert.Equal(t, "sound.mp3", domainAtt.URL)
		assert.Equal(t, []byte("audio data"), domainAtt.Content)
		assert.NotNil(t, domainAtt.Metadata)
		assert.Empty(t, domainAtt.Metadata)
	})

	t.Run("domainAttachmentToLLM", func(t *testing.T) {
		domainAtt := domain.Attachment{
			Type:     domain.AttachmentTypeFile,
			MimeType: "text/plain",
			Name:     "document.txt",
			URL:      "http://example.com/document.txt",
			Content:  []byte("text content"),
			Metadata: map[string]interface{}{"key": "value"},
		}

		llmAtt := domainAttachmentToLLM(domainAtt)
		assert.Equal(t, llm.AttachmentTypeFile, llmAtt.Type)
		assert.Equal(t, "text/plain", llmAtt.MimeType)
		assert.Equal(t, "document.txt", llmAtt.FilePath)
		assert.Equal(t, "text content", llmAtt.Content)
	})
}