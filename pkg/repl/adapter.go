// ABOUTME: Adapters to convert between REPL types and storage types
// ABOUTME: Provides clean separation between REPL concerns and storage abstraction

package repl

import (
	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/lexlapax/magellai/pkg/storage"
)

// ToStorageSession converts a REPL session to a storage session
func ToStorageSession(replSession *Session) *storage.Session {
	if replSession == nil {
		return nil
	}

	// Convert conversation to storage session fields
	storageSession := &storage.Session{
		ID:       replSession.ID,
		Name:     replSession.Name,
		Config:   replSession.Config,
		Created:  replSession.Created,
		Updated:  replSession.Updated,
		Tags:     replSession.Tags,
		Metadata: replSession.Metadata,
	}

	// Copy conversation fields if present
	if replSession.Conversation != nil {
		storageSession.Model = replSession.Conversation.Model
		storageSession.Provider = replSession.Conversation.Provider
		storageSession.Temperature = replSession.Conversation.Temperature
		storageSession.MaxTokens = replSession.Conversation.MaxTokens
		storageSession.SystemPrompt = replSession.Conversation.SystemPrompt

		// Convert messages
		storageSession.Messages = make([]storage.Message, len(replSession.Conversation.Messages))
		for i, msg := range replSession.Conversation.Messages {
			storageSession.Messages[i] = ToStorageMessage(msg)
		}
	}

	return storageSession
}

// FromStorageSession converts a storage session to a REPL session
func FromStorageSession(storageSession *storage.Session) *Session {
	if storageSession == nil {
		return nil
	}

	// Create conversation from storage session fields
	conversation := &Conversation{
		ID:           storageSession.ID,
		Model:        storageSession.Model,
		Provider:     storageSession.Provider,
		Temperature:  storageSession.Temperature,
		MaxTokens:    storageSession.MaxTokens,
		SystemPrompt: storageSession.SystemPrompt,
		Created:      storageSession.Created,
		Updated:      storageSession.Updated,
		Messages:     make([]Message, len(storageSession.Messages)),
	}

	// Convert messages
	for i, msg := range storageSession.Messages {
		conversation.Messages[i] = FromStorageMessage(msg)
	}

	return &Session{
		ID:           storageSession.ID,
		Name:         storageSession.Name,
		Conversation: conversation,
		Config:       storageSession.Config,
		Created:      storageSession.Created,
		Updated:      storageSession.Updated,
		Tags:         storageSession.Tags,
		Metadata:     storageSession.Metadata,
	}
}

// ToStorageMessage converts a REPL message to a storage message
func ToStorageMessage(replMsg Message) storage.Message {
	storageMsg := storage.Message{
		ID:        replMsg.ID,
		Role:      replMsg.Role,
		Content:   replMsg.Content,
		Timestamp: replMsg.Timestamp,
		Metadata:  replMsg.Metadata,
	}

	// Convert attachments
	if len(replMsg.Attachments) > 0 {
		storageMsg.Attachments = make([]storage.Attachment, len(replMsg.Attachments))
		for i, att := range replMsg.Attachments {
			storageMsg.Attachments[i] = llmAttachmentToStorage(att)
		}
	}

	return storageMsg
}

// FromStorageMessage converts a storage message to a REPL message
func FromStorageMessage(storageMsg storage.Message) Message {
	replMsg := Message{
		ID:        storageMsg.ID,
		Role:      storageMsg.Role,
		Content:   storageMsg.Content,
		Timestamp: storageMsg.Timestamp,
		Metadata:  storageMsg.Metadata,
	}

	// Convert attachments
	if len(storageMsg.Attachments) > 0 {
		replMsg.Attachments = make([]llm.Attachment, len(storageMsg.Attachments))
		for i, att := range storageMsg.Attachments {
			replMsg.Attachments[i] = storageAttachmentToLLM(att)
		}
	}

	return replMsg
}

// ToStorageSearchResult converts REPL search results to storage search results
func ToStorageSearchResult(replResult *SearchResult) *storage.SearchResult {
	if replResult == nil {
		return nil
	}

	return &storage.SearchResult{
		Session: ToStorageSessionInfo(replResult.Session),
		Matches: ToStorageSearchMatches(replResult.Matches),
	}
}

// FromStorageSearchResult converts storage search results to REPL search results
func FromStorageSearchResult(storageResult *storage.SearchResult) *SearchResult {
	if storageResult == nil {
		return nil
	}

	return &SearchResult{
		Session: FromStorageSessionInfo(storageResult.Session),
		Matches: FromStorageSearchMatches(storageResult.Matches),
	}
}

// ToStorageSessionInfo converts REPL session info to storage session info
func ToStorageSessionInfo(replInfo *SessionInfo) *storage.SessionInfo {
	if replInfo == nil {
		return nil
	}

	return &storage.SessionInfo{
		ID:           replInfo.ID,
		Name:         replInfo.Name,
		Created:      replInfo.Created,
		Updated:      replInfo.Updated,
		MessageCount: replInfo.MessageCount,
		Tags:         replInfo.Tags,
	}
}

// FromStorageSessionInfo converts storage session info to REPL session info
func FromStorageSessionInfo(storageInfo *storage.SessionInfo) *SessionInfo {
	if storageInfo == nil {
		return nil
	}

	return &SessionInfo{
		ID:           storageInfo.ID,
		Name:         storageInfo.Name,
		Created:      storageInfo.Created,
		Updated:      storageInfo.Updated,
		MessageCount: storageInfo.MessageCount,
		Tags:         storageInfo.Tags,
	}
}

// ToStorageSearchMatches converts REPL search matches to storage search matches
func ToStorageSearchMatches(replMatches []SearchMatch) []storage.SearchMatch {
	storageMatches := make([]storage.SearchMatch, len(replMatches))
	for i, match := range replMatches {
		storageMatches[i] = storage.SearchMatch{
			Type:     match.Type,
			Role:     match.Role,
			Content:  match.Content,
			Context:  match.Context,
			Position: match.Position,
		}
	}
	return storageMatches
}

// FromStorageSearchMatches converts storage search matches to REPL search matches
func FromStorageSearchMatches(storageMatches []storage.SearchMatch) []SearchMatch {
	replMatches := make([]SearchMatch, len(storageMatches))
	for i, match := range storageMatches {
		replMatches[i] = SearchMatch{
			Type:     match.Type,
			Role:     match.Role,
			Content:  match.Content,
			Context:  match.Context,
			Position: match.Position,
		}
	}
	return replMatches
}

// Helper functions for attachment conversion

func llmAttachmentToStorage(llmAtt llm.Attachment) storage.Attachment {
	return storage.Attachment{
		Type:     string(llmAtt.Type),
		URL:      llmAtt.FilePath, // Use FilePath as URL for compatibility
		MimeType: llmAtt.MimeType,
		Name:     llmAtt.FilePath, // Map FilePath to Name
		Content:  llmAtt.Content,
		Metadata: make(map[string]interface{}), // LLM attachments don't have metadata
	}
}

func storageAttachmentToLLM(storageAtt storage.Attachment) llm.Attachment {
	return llm.Attachment{
		Type:     llm.AttachmentType(storageAtt.Type),
		MimeType: storageAtt.MimeType,
		FilePath: storageAtt.Name, // Map Name back to FilePath
		Content:  storageAtt.Content,
	}
}
