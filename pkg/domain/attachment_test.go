package domain

import (
	"testing"
)

func TestNewAttachment(t *testing.T) {
	id := "att-123"
	attType := AttachmentTypeImage

	att := NewAttachment(id, attType)

	if att.ID != id {
		t.Errorf("Expected attachment ID %s, got %s", id, att.ID)
	}

	if att.Type != attType {
		t.Errorf("Expected type %s, got %s", attType, att.Type)
	}

	if att.Metadata == nil {
		t.Error("Expected metadata map to be initialized")
	}
}

func TestAttachmentIsValid(t *testing.T) {
	tests := []struct {
		name       string
		attachment Attachment
		expected   bool
	}{
		{
			name: "valid with content",
			attachment: Attachment{
				ID:      "att-1",
				Type:    AttachmentTypeImage,
				Content: []byte("image data"),
			},
			expected: true,
		},
		{
			name: "valid with file path",
			attachment: Attachment{
				ID:       "att-2",
				Type:     AttachmentTypeFile,
				FilePath: "/path/to/file.pdf",
			},
			expected: true,
		},
		{
			name: "valid with URL",
			attachment: Attachment{
				ID:   "att-3",
				Type: AttachmentTypeAudio,
				URL:  "https://example.com/audio.mp3",
			},
			expected: true,
		},
		{
			name: "invalid - no ID",
			attachment: Attachment{
				Type:    AttachmentTypeText,
				Content: []byte("text"),
			},
			expected: false,
		},
		{
			name: "invalid - no type",
			attachment: Attachment{
				ID:      "att-4",
				Content: []byte("data"),
			},
			expected: false,
		},
		{
			name: "invalid - invalid type",
			attachment: Attachment{
				ID:      "att-5",
				Type:    "invalid",
				Content: []byte("data"),
			},
			expected: false,
		},
		{
			name: "invalid - no content or reference",
			attachment: Attachment{
				ID:   "att-6",
				Type: AttachmentTypeVideo,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attachment.IsValid()
			if result != tt.expected {
				t.Errorf("Expected IsValid() to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAttachmentHelpers(t *testing.T) {
	// Test HasContent
	att1 := Attachment{
		ID:      "att-1",
		Type:    AttachmentTypeImage,
		Content: []byte("image data"),
	}

	if !att1.HasContent() {
		t.Error("Expected HasContent to return true")
	}

	// Test HasReference
	att2 := Attachment{
		ID:       "att-2",
		Type:     AttachmentTypeFile,
		FilePath: "/path/to/file",
	}

	if !att2.HasReference() {
		t.Error("Expected HasReference to return true for FilePath")
	}

	att3 := Attachment{
		ID:   "att-3",
		Type: AttachmentTypeAudio,
		URL:  "https://example.com/audio.mp3",
	}

	if !att3.HasReference() {
		t.Error("Expected HasReference to return true for URL")
	}

	// Test GetDisplayName
	tests := []struct {
		attachment Attachment
		expected   string
	}{
		{
			attachment: Attachment{ID: "att-1", Name: "document.pdf"},
			expected:   "document.pdf",
		},
		{
			attachment: Attachment{ID: "att-2", FilePath: "/path/to/file.txt"},
			expected:   "/path/to/file.txt",
		},
		{
			attachment: Attachment{ID: "att-3", URL: "https://example.com/image.jpg"},
			expected:   "https://example.com/image.jpg",
		},
		{
			attachment: Attachment{ID: "att-4"},
			expected:   "att-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.attachment.GetDisplayName()
			if result != tt.expected {
				t.Errorf("Expected display name %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestAttachmentType(t *testing.T) {
	tests := []struct {
		attType  AttachmentType
		expected bool
	}{
		{AttachmentTypeImage, true},
		{AttachmentTypeFile, true},
		{AttachmentTypeText, true},
		{AttachmentTypeAudio, true},
		{AttachmentTypeVideo, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.attType), func(t *testing.T) {
			if tt.attType.IsValid() != tt.expected {
				t.Errorf("Expected IsValid() to return %v for type %s", tt.expected, tt.attType)
			}
		})
	}
}
