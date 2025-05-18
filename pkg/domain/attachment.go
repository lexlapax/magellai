// ABOUTME: Domain types for attachments including Attachment and AttachmentType
// ABOUTME: Core business entities for multimodal content attached to messages

package domain

// Attachment represents multimodal content attached to a message.
type Attachment struct {
	ID       string                 `json:"id"`
	Type     AttachmentType         `json:"type"`
	Content  []byte                 `json:"content,omitempty"`
	FilePath string                 `json:"file_path,omitempty"`
	URL      string                 `json:"url,omitempty"`
	Name     string                 `json:"name,omitempty"`
	MimeType string                 `json:"mime_type,omitempty"`
	Size     int64                  `json:"size,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AttachmentType represents the type of attachment.
type AttachmentType string

// AttachmentType constants define the possible attachment types.
const (
	AttachmentTypeImage AttachmentType = "image"
	AttachmentTypeFile  AttachmentType = "file"
	AttachmentTypeText  AttachmentType = "text"
	AttachmentTypeAudio AttachmentType = "audio"
	AttachmentTypeVideo AttachmentType = "video"
)

// NewAttachment creates a new attachment with the given parameters.
func NewAttachment(id string, attachmentType AttachmentType) *Attachment {
	return &Attachment{
		ID:       id,
		Type:     attachmentType,
		Metadata: make(map[string]interface{}),
	}
}

// IsValid validates the attachment fields.
func (a *Attachment) IsValid() bool {
	return a.ID != "" && 
		a.Type.IsValid() && 
		(len(a.Content) > 0 || a.FilePath != "" || a.URL != "")
}

// HasContent returns true if the attachment has content data.
func (a *Attachment) HasContent() bool {
	return len(a.Content) > 0
}

// HasReference returns true if the attachment has a file path or URL reference.
func (a *Attachment) HasReference() bool {
	return a.FilePath != "" || a.URL != ""
}

// GetDisplayName returns a display name for the attachment.
func (a *Attachment) GetDisplayName() string {
	if a.Name != "" {
		return a.Name
	}
	if a.FilePath != "" {
		return a.FilePath
	}
	if a.URL != "" {
		return a.URL
	}
	return a.ID
}

// String returns the attachment type as a string.
func (t AttachmentType) String() string {
	return string(t)
}

// IsValid checks if the attachment type is valid.
func (t AttachmentType) IsValid() bool {
	return t == AttachmentTypeImage || 
		t == AttachmentTypeFile || 
		t == AttachmentTypeText || 
		t == AttachmentTypeAudio || 
		t == AttachmentTypeVideo
}