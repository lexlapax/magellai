// ABOUTME: Helper functions for working with attachments in the REPL
// ABOUTME: Provides utilities for creating and displaying attachments

package repl

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/llm"
)

// createFileAttachmentFromPath creates an attachment from a file path
func createFileAttachmentFromPath(filePath string) (llm.Attachment, error) {
	// Read file contents
	data, err := os.ReadFile(filePath)
	if err != nil {
		return llm.Attachment{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine attachment type based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	var attachType llm.AttachmentType
	var mimeType string

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		attachType = llm.AttachmentTypeImage
		mimeType = "image/" + strings.TrimPrefix(ext, ".")
		if ext == ".jpg" {
			mimeType = "image/jpeg"
		}
	case ".mp3", ".wav", ".m4a", ".ogg":
		attachType = llm.AttachmentTypeAudio
		mimeType = "audio/" + strings.TrimPrefix(ext, ".")
		if ext == ".mp3" {
			mimeType = "audio/mpeg"
		}
	case ".mp4", ".avi", ".mov", ".webm":
		attachType = llm.AttachmentTypeVideo
		mimeType = "video/" + strings.TrimPrefix(ext, ".")
	default:
		attachType = llm.AttachmentTypeFile
		mimeType = "application/octet-stream"
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)

	attachment := llm.Attachment{
		Type:     attachType,
		Content:  encoded,
		FilePath: filePath,
		MimeType: mimeType,
	}

	return attachment, nil
}

// getAttachmentDisplayName returns a display name for an attachment
func getAttachmentDisplayName(att llm.Attachment) string {
	if att.FilePath != "" {
		return filepath.Base(att.FilePath)
	}
	return string(att.Type) + "_attachment"
}

// getDomainAttachmentDisplayName returns a display name for a domain attachment
func getDomainAttachmentDisplayName(att domain.Attachment) string {
	if att.GetDisplayName() != "" {
		return att.GetDisplayName()
	}
	return string(att.Type) + "_attachment"
}