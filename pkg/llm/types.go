// ABOUTME: Core types for wrapping go-llms library types
// ABOUTME: Provides adapter types that bridge between Magellai and go-llms domain types
package llm

import (
	"strings"
	
	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Provider name constants
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderGemini    = "gemini"
	ProviderOllama    = "ollama"
	ProviderMock      = "mock"
)

// Model capability flags
type ModelCapability string

const (
	CapabilityText  ModelCapability = "text"
	CapabilityImage ModelCapability = "image"
	CapabilityAudio ModelCapability = "audio"
	CapabilityVideo ModelCapability = "video"
	CapabilityFile  ModelCapability = "file"
)

// ModelInfo represents metadata about a model
// NOTE: We intentionally don't maintain a hard-coded list of models
// as they change frequently. Instead, providers will query their
// available models at runtime or use configuration.
type ModelInfo struct {
	Provider     string            `json:"provider"`
	Model        string            `json:"model"`
	Capabilities []ModelCapability `json:"capabilities"`
	MaxTokens    int               `json:"max_tokens,omitempty"`
	Description  string            `json:"description,omitempty"`
}

// Request wraps go-llms domain.Message for Magellai usage
type Request struct {
	Messages    []Message      `json:"messages"`
	Model       string         `json:"model,omitempty"`       // provider/model format
	Temperature *float64       `json:"temperature,omitempty"`
	MaxTokens   *int           `json:"max_tokens,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	SystemPrompt string        `json:"system_prompt,omitempty"`
	Options     *PromptParams  `json:"options,omitempty"`
}

// Message wraps go-llms domain.Message
type Message struct {
	Role        string        `json:"role"`         // system, user, assistant
	Content     string        `json:"content"`
	Attachments []Attachment  `json:"attachments,omitempty"`
}

// Response wraps go-llms response types
type Response struct {
	Content      string                 `json:"content"`
	Model        string                 `json:"model,omitempty"`
	Usage        *Usage                 `json:"usage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	FinishReason string                 `json:"finish_reason,omitempty"`
}

// Usage tracks token usage
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Attachment represents multimodal content
type Attachment struct {
	Type     AttachmentType `json:"type"`
	Content  string         `json:"content,omitempty"`  // For text or base64 data
	FilePath string         `json:"file_path,omitempty"` // For file references
	MimeType string         `json:"mime_type,omitempty"`
}

// AttachmentType defines the type of attachment
type AttachmentType string

const (
	AttachmentTypeImage AttachmentType = "image"
	AttachmentTypeAudio AttachmentType = "audio"
	AttachmentTypeVideo AttachmentType = "video"
	AttachmentTypeFile  AttachmentType = "file"
	AttachmentTypeText  AttachmentType = "text"
)

// PromptParams maps to go-llms domain.Option
type PromptParams struct {
	Temperature      *float64               `json:"temperature,omitempty"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	TopK             *int                   `json:"top_k,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	Seed             *int                   `json:"seed,omitempty"`
	ResponseFormat   string                 `json:"response_format,omitempty"`
	CustomOptions    map[string]interface{} `json:"custom_options,omitempty"`
}

// Conversion methods to go-llms types

// ToLLMMessage converts a Message to go-llms domain.Message
func (m Message) ToLLMMessage() domain.Message {
	llmMsg := domain.Message{
		Role: domain.Role(m.Role),
	}
	
	// Convert simple text content
	if m.Content != "" && len(m.Attachments) == 0 {
		llmMsg.Content = []domain.ContentPart{
			{
				Type: domain.ContentTypeText,
				Text: m.Content,
			},
		}
		return llmMsg
	}
	
	// Convert with attachments
	llmMsg.Content = make([]domain.ContentPart, 0)
	
	// Add text content first if present
	if m.Content != "" {
		llmMsg.Content = append(llmMsg.Content, domain.ContentPart{
			Type: domain.ContentTypeText,
			Text: m.Content,
		})
	}
	
	// Add attachments
	for _, att := range m.Attachments {
		if part := att.ToLLMContentPart(); part != nil {
			llmMsg.Content = append(llmMsg.Content, *part)
		}
	}
	
	return llmMsg
}

// ToLLMContentPart converts an Attachment to go-llms domain.ContentPart
func (a Attachment) ToLLMContentPart() *domain.ContentPart {
	switch a.Type {
	case AttachmentTypeImage:
		return &domain.ContentPart{
			Type: domain.ContentTypeImage,
			Image: &domain.ImageContent{
				Source: domain.SourceInfo{
					Type: domain.SourceTypeURL,
					URL:  a.Content,
					MediaType: a.MimeType,
				},
			},
		}
	case AttachmentTypeText:
		return &domain.ContentPart{
			Type: domain.ContentTypeText,
			Text: a.Content,
		}
	case AttachmentTypeFile:
		return &domain.ContentPart{
			Type: domain.ContentTypeFile,
			File: &domain.FileContent{
				FileName: a.FilePath,
				FileData: a.Content, // Expects base64
				MimeType: a.MimeType,
			},
		}
	case AttachmentTypeVideo:
		return &domain.ContentPart{
			Type: domain.ContentTypeVideo,
			Video: &domain.VideoContent{
				Source: domain.SourceInfo{
					Type: domain.SourceTypeURL,
					URL:  a.FilePath,
					MediaType: a.MimeType,
				},
			},
		}
	case AttachmentTypeAudio:
		return &domain.ContentPart{
			Type: domain.ContentTypeAudio,
			Audio: &domain.AudioContent{
				Source: domain.SourceInfo{
					Type: domain.SourceTypeURL,
					URL:  a.FilePath,
					MediaType: a.MimeType,
				},
			},
		}
	}
	return nil
}

// FromLLMMessage converts a go-llms domain.Message to Message
func FromLLMMessage(msg domain.Message) Message {
	m := Message{
		Role: string(msg.Role),
	}
	
	// Convert content parts
	for _, part := range msg.Content {
		switch part.Type {
		case domain.ContentTypeText:
			if m.Content == "" {
				m.Content = part.Text
			} else {
				// Multiple text parts become attachments
				m.Attachments = append(m.Attachments, Attachment{
					Type:    AttachmentTypeText,
					Content: part.Text,
				})
			}
		default:
			if att := fromLLMContentPart(part); att != nil {
				m.Attachments = append(m.Attachments, *att)
			}
		}
	}
	
	return m
}

// fromLLMContentPart converts a go-llms domain.ContentPart to Attachment
func fromLLMContentPart(part domain.ContentPart) *Attachment {
	switch part.Type {
	case domain.ContentTypeImage:
		if part.Image != nil {
			return &Attachment{
				Type:     AttachmentTypeImage,
				Content:  part.Image.Source.URL,
				MimeType: part.Image.Source.MediaType,
			}
		}
	case domain.ContentTypeFile:
		if part.File != nil {
			return &Attachment{
				Type:     AttachmentTypeFile,
				FilePath: part.File.FileName,
				Content:  part.File.FileData,
				MimeType: part.File.MimeType,
			}
		}
	case domain.ContentTypeVideo:
		if part.Video != nil {
			return &Attachment{
				Type:     AttachmentTypeVideo,
				FilePath: part.Video.Source.URL,
				MimeType: part.Video.Source.MediaType,
			}
		}
	case domain.ContentTypeAudio:
		if part.Audio != nil {
			return &Attachment{
				Type:     AttachmentTypeAudio,
				FilePath: part.Audio.Source.URL,
				MimeType: part.Audio.Source.MediaType,
			}
		}
	}
	return nil
}

// ParseModelString splits a provider/model string into components
func ParseModelString(model string) (provider, modelName string) {
	parts := strings.SplitN(model, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// Default to OpenAI if no provider specified
	return ProviderOpenAI, model
}

// FormatModelString combines provider and model into provider/model format
func FormatModelString(provider, model string) string {
	return provider + "/" + model
}