// ABOUTME: Adapters to convert between REPL types and domain types
// ABOUTME: Provides clean separation between REPL concerns and domain abstraction

package repl

import (
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// ToDomainSession converts a REPL session to a domain session
func ToDomainSession(replSession *Session) *domain.Session {
	if replSession == nil {
		return nil
	}

	// Convert to domain session
	domainSession := &domain.Session{
		ID:       replSession.ID,
		Name:     replSession.Name,
		Config:   replSession.Config,
		Created:  replSession.Created,
		Updated:  replSession.Updated,
		Tags:     replSession.Tags,
		Metadata: replSession.Metadata,
	}

	// Convert conversation if present
	if replSession.Conversation != nil {
		domainSession.Conversation = &domain.Conversation{
			ID:           replSession.Conversation.ID,
			Model:        replSession.Conversation.Model,
			Provider:     replSession.Conversation.Provider,
			Temperature:  replSession.Conversation.Temperature,
			MaxTokens:    replSession.Conversation.MaxTokens,
			SystemPrompt: replSession.Conversation.SystemPrompt,
			Created:      replSession.Conversation.Created,
			Updated:      replSession.Conversation.Updated,
			Messages:     make([]domain.Message, len(replSession.Conversation.Messages)),
		}

		// Convert messages
		for i, msg := range replSession.Conversation.Messages {
			domainSession.Conversation.Messages[i] = ToDomainMessage(msg)
		}
	}

	return domainSession
}

// FromDomainSession converts a domain session to a REPL session
func FromDomainSession(domainSession *domain.Session) *Session {
	if domainSession == nil {
		return nil
	}

	replSession := &Session{
		ID:       domainSession.ID,
		Name:     domainSession.Name,
		Config:   domainSession.Config,
		Created:  domainSession.Created,
		Updated:  domainSession.Updated,
		Tags:     domainSession.Tags,
		Metadata: domainSession.Metadata,
	}

	// Convert conversation if present
	if domainSession.Conversation != nil {
		replSession.Conversation = &Conversation{
			ID:           domainSession.Conversation.ID,
			Model:        domainSession.Conversation.Model,
			Provider:     domainSession.Conversation.Provider,
			Temperature:  domainSession.Conversation.Temperature,
			MaxTokens:    domainSession.Conversation.MaxTokens,
			SystemPrompt: domainSession.Conversation.SystemPrompt,
			Created:      domainSession.Conversation.Created,
			Updated:      domainSession.Conversation.Updated,
			Messages:     make([]Message, len(domainSession.Conversation.Messages)),
		}

		// Convert messages
		for i, msg := range domainSession.Conversation.Messages {
			replSession.Conversation.Messages[i] = FromDomainMessage(msg)
		}
	}

	return replSession
}

// ToDomainMessage converts a REPL message to a domain message
func ToDomainMessage(replMsg Message) domain.Message {
	domainMsg := domain.Message{
		ID:        replMsg.ID,
		Role:      domain.MessageRole(replMsg.Role),
		Content:   replMsg.Content,
		Timestamp: replMsg.Timestamp,
		Metadata:  replMsg.Metadata,
	}

	// Convert attachments
	if len(replMsg.Attachments) > 0 {
		domainMsg.Attachments = make([]domain.Attachment, len(replMsg.Attachments))
		for i, att := range replMsg.Attachments {
			domainMsg.Attachments[i] = llmAttachmentToDomain(att)
		}
	}

	return domainMsg
}

// FromDomainMessage converts a domain message to a REPL message
func FromDomainMessage(domainMsg domain.Message) Message {
	replMsg := Message{
		ID:        domainMsg.ID,
		Role:      string(domainMsg.Role),
		Content:   domainMsg.Content,
		Timestamp: domainMsg.Timestamp,
		Metadata:  domainMsg.Metadata,
	}

	// Convert attachments
	if len(domainMsg.Attachments) > 0 {
		replMsg.Attachments = make([]llm.Attachment, len(domainMsg.Attachments))
		for i, att := range domainMsg.Attachments {
			replMsg.Attachments[i] = domainAttachmentToLLM(att)
		}
	}

	return replMsg
}

// ToDomainSearchResult converts REPL search results to domain search results
func ToDomainSearchResult(replResult *SearchResult) *domain.SearchResult {
	if replResult == nil {
		return nil
	}

	result := domain.NewSearchResult(ToDomainSessionInfo(replResult.Session))
	for _, match := range replResult.Matches {
		result.AddMatch(ToDomainSearchMatch(match))
	}
	return result
}

// FromDomainSearchResult converts domain search results to REPL search results
func FromDomainSearchResult(domainResult *domain.SearchResult) *SearchResult {
	if domainResult == nil {
		return nil
	}

	return &SearchResult{
		Session: FromDomainSessionInfo(domainResult.Session),
		Matches: FromDomainSearchMatches(domainResult.Matches),
	}
}

// ToDomainSessionInfo converts REPL session info to domain session info
func ToDomainSessionInfo(replInfo *SessionInfo) *domain.SessionInfo {
	if replInfo == nil {
		return nil
	}

	return &domain.SessionInfo{
		ID:           replInfo.ID,
		Name:         replInfo.Name,
		Created:      replInfo.Created,
		Updated:      replInfo.Updated,
		MessageCount: replInfo.MessageCount,
		Tags:         replInfo.Tags,
	}
}

// FromDomainSessionInfo converts domain session info to REPL session info
func FromDomainSessionInfo(domainInfo *domain.SessionInfo) *SessionInfo {
	if domainInfo == nil {
		return nil
	}

	return &SessionInfo{
		ID:           domainInfo.ID,
		Name:         domainInfo.Name,
		Created:      domainInfo.Created,
		Updated:      domainInfo.Updated,
		MessageCount: domainInfo.MessageCount,
		Tags:         domainInfo.Tags,
	}
}

// ToDomainSearchMatch converts a REPL search match to a domain search match
func ToDomainSearchMatch(replMatch SearchMatch) domain.SearchMatch {
	return domain.NewSearchMatch(
		replMatch.Type,
		replMatch.Role,
		replMatch.Content,
		replMatch.Context,
		replMatch.Position,
	)
}

// FromDomainSearchMatches converts domain search matches to REPL search matches
func FromDomainSearchMatches(domainMatches []domain.SearchMatch) []SearchMatch {
	replMatches := make([]SearchMatch, len(domainMatches))
	for i, match := range domainMatches {
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

func llmAttachmentToDomain(llmAtt llm.Attachment) domain.Attachment {
	return domain.Attachment{
		Type:     domain.AttachmentType(llmAtt.Type),
		URL:      llmAtt.FilePath, // Use FilePath as URL for compatibility
		MimeType: llmAtt.MimeType,
		Name:     llmAtt.FilePath, // Map FilePath to Name
		Content:  []byte(llmAtt.Content),
		Metadata: make(map[string]interface{}), // LLM attachments don't have metadata
	}
}

func domainAttachmentToLLM(domainAtt domain.Attachment) llm.Attachment {
	return llm.Attachment{
		Type:     llm.AttachmentType(domainAtt.Type),
		MimeType: domainAtt.MimeType,
		FilePath: domainAtt.Name, // Map Name back to FilePath
		Content:  string(domainAtt.Content),
	}
}