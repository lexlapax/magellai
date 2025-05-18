// ABOUTME: Adapters to convert between LLM types and domain types
// ABOUTME: Provides conversion for LLM-specific attachment handling

package repl

import (
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// Helper functions for attachment conversion between LLM and domain types

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

// convertDomainMessageToLLM converts a domain message to an LLM message
func convertDomainMessageToLLM(domainMsg domain.Message) llm.Message {
	llmMsg := llm.Message{
		Role:    string(domainMsg.Role),
		Content: domainMsg.Content,
	}

	// Convert attachments
	if len(domainMsg.Attachments) > 0 {
		llmMsg.Attachments = make([]llm.Attachment, len(domainMsg.Attachments))
		for i, att := range domainMsg.Attachments {
			llmMsg.Attachments[i] = domainAttachmentToLLM(att)
		}
	}

	return llmMsg
}
