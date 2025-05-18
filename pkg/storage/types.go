// ABOUTME: Storage-specific types and conversions for the storage package
// ABOUTME: Uses domain types from pkg/domain and provides any storage-specific functionality

package storage

import (
	"github.com/lexlapax/magellai/pkg/domain"
)

// StorageSession is a flattened representation of domain.Session for storage purposes.
// This type is only needed if we need to transform the domain structure for storage,
// but we'll try to use domain types directly first.
type StorageSession struct {
	*domain.Session
	
	// Flattened conversation fields for easier storage/querying
	Model        string  `json:"model,omitempty"`
	Provider     string  `json:"provider,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
	SystemPrompt string  `json:"system_prompt,omitempty"`
	Messages     []domain.Message `json:"messages"`
}

// ToStorageSession converts a domain.Session to a storage-specific format if needed.
// This flattens the conversation data for easier storage and querying.
func ToStorageSession(session *domain.Session) *StorageSession {
	if session == nil {
		return nil
	}
	
	storageSession := &StorageSession{
		Session: session,
	}
	
	// Flatten conversation fields if present
	if session.Conversation != nil {
		storageSession.Model = session.Conversation.Model
		storageSession.Provider = session.Conversation.Provider
		storageSession.Temperature = session.Conversation.Temperature
		storageSession.MaxTokens = session.Conversation.MaxTokens
		storageSession.SystemPrompt = session.Conversation.SystemPrompt
		storageSession.Messages = session.Conversation.Messages
	}
	
	return storageSession
}

// ToDomainSession converts a storage session back to domain.Session.
func ToDomainSession(storageSession *StorageSession) *domain.Session {
	if storageSession == nil {
		return nil
	}
	
	// Reconstruct the domain session
	session := &domain.Session{
		ID:       storageSession.ID,
		Name:     storageSession.Name,
		Created:  storageSession.Created,
		Updated:  storageSession.Updated,
		Tags:     storageSession.Tags,
		Config:   storageSession.Config,
		Metadata: storageSession.Metadata,
	}
	
	// Reconstruct conversation from flattened fields
	if len(storageSession.Messages) > 0 || storageSession.Model != "" {
		conversation := &domain.Conversation{
			ID:           session.ID,
			Messages:     storageSession.Messages,
			Model:        storageSession.Model,
			Provider:     storageSession.Provider,
			Temperature:  storageSession.Temperature,
			MaxTokens:    storageSession.MaxTokens,
			SystemPrompt: storageSession.SystemPrompt,
			Created:      session.Created,
			Updated:      session.Updated,
		}
		session.Conversation = conversation
	}
	
	return session
}