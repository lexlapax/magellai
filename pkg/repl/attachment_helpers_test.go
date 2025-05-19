// ABOUTME: Tests for attachment helper functions in the REPL
// ABOUTME: Tests file attachment creation and display name generation

package repl

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
)

func TestCreateFileAttachmentFromPath(t *testing.T) {
	// Create temp directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		filename      string
		content       string
		wantType      domain.AttachmentType
		wantMimeType  string
		wantError     bool
		errorContains string
		setupFunc     func(string) error
	}{
		{
			name:         "image jpeg file",
			filename:     "test.jpg",
			content:      "fake jpeg data",
			wantType:     domain.AttachmentTypeImage,
			wantMimeType: "image/jpeg",
		},
		{
			name:         "image png file",
			filename:     "test.png",
			content:      "fake png data",
			wantType:     domain.AttachmentTypeImage,
			wantMimeType: "image/png",
		},
		{
			name:         "image gif file",
			filename:     "test.gif",
			content:      "fake gif data",
			wantType:     domain.AttachmentTypeImage,
			wantMimeType: "image/gif",
		},
		{
			name:         "image webp file",
			filename:     "test.webp",
			content:      "fake webp data",
			wantType:     domain.AttachmentTypeImage,
			wantMimeType: "image/webp",
		},
		{
			name:         "audio mp3 file",
			filename:     "test.mp3",
			content:      "fake mp3 data",
			wantType:     domain.AttachmentTypeAudio,
			wantMimeType: "audio/mpeg",
		},
		{
			name:         "audio wav file",
			filename:     "test.wav",
			content:      "fake wav data",
			wantType:     domain.AttachmentTypeAudio,
			wantMimeType: "audio/wav",
		},
		{
			name:         "audio m4a file",
			filename:     "test.m4a",
			content:      "fake m4a data",
			wantType:     domain.AttachmentTypeAudio,
			wantMimeType: "audio/m4a",
		},
		{
			name:         "audio ogg file",
			filename:     "test.ogg",
			content:      "fake ogg data",
			wantType:     domain.AttachmentTypeAudio,
			wantMimeType: "audio/ogg",
		},
		{
			name:         "video mp4 file",
			filename:     "test.mp4",
			content:      "fake mp4 data",
			wantType:     domain.AttachmentTypeVideo,
			wantMimeType: "video/mp4",
		},
		{
			name:         "video avi file",
			filename:     "test.avi",
			content:      "fake avi data",
			wantType:     domain.AttachmentTypeVideo,
			wantMimeType: "video/avi",
		},
		{
			name:         "video mov file",
			filename:     "test.mov",
			content:      "fake mov data",
			wantType:     domain.AttachmentTypeVideo,
			wantMimeType: "video/mov",
		},
		{
			name:         "video webm file",
			filename:     "test.webm",
			content:      "fake webm data",
			wantType:     domain.AttachmentTypeVideo,
			wantMimeType: "video/webm",
		},
		{
			name:         "unknown file type",
			filename:     "test.txt",
			content:      "plain text data",
			wantType:     domain.AttachmentTypeFile,
			wantMimeType: "application/octet-stream",
		},
		{
			name:         "file with uppercase extension",
			filename:     "test.JPG",
			content:      "fake jpeg data",
			wantType:     domain.AttachmentTypeImage,
			wantMimeType: "image/jpeg",
		},
		{
			name:         "file with mixed case extension",
			filename:     "test.Mp3",
			content:      "fake mp3 data",
			wantType:     domain.AttachmentTypeAudio,
			wantMimeType: "audio/mpeg",
		},
		{
			name:          "non-existent file",
			filename:      "does-not-exist.jpg",
			wantError:     true,
			errorContains: "failed to read file",
			setupFunc: func(path string) error {
				// Don't create the file
				return nil
			},
		},
		{
			name:         "empty file",
			filename:     "empty.jpg",
			content:      "",
			wantType:     domain.AttachmentTypeImage,
			wantMimeType: "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.filename)

			// Setup test file
			if tt.setupFunc != nil {
				err := tt.setupFunc(filePath)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			} else {
				err := os.WriteFile(filePath, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			// Test the function
			attachment, err := createFileAttachmentFromPath(filePath)

			// Check error expectations
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error should contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify attachment type
			if attachment.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, attachment.Type)
			}

			// Verify MIME type
			if attachment.MimeType != tt.wantMimeType {
				t.Errorf("expected MIME type %s, got %s", tt.wantMimeType, attachment.MimeType)
			}

			// Verify file path is preserved
			if attachment.FilePath != filePath {
				t.Errorf("expected file path %s, got %s", filePath, attachment.FilePath)
			}

			// Verify content is base64 encoded
			expectedContent := base64.StdEncoding.EncodeToString([]byte(tt.content))
			if string(attachment.Content) != expectedContent {
				t.Errorf("expected content %s, got %s", expectedContent, string(attachment.Content))
			}
		})
	}
}

func TestGetAttachmentDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		att      domain.Attachment
		expected string
	}{
		{
			name: "attachment with file path",
			att: domain.Attachment{
				Type:     domain.AttachmentTypeImage,
				FilePath: "/path/to/file/image.jpg",
			},
			expected: "image.jpg",
		},
		{
			name: "attachment with empty file path",
			att: domain.Attachment{
				Type:     domain.AttachmentTypeImage,
				FilePath: "",
			},
			expected: "image_attachment",
		},
		{
			name: "attachment with path only",
			att: domain.Attachment{
				Type:     domain.AttachmentTypeAudio,
				FilePath: "/",
			},
			expected: "/",
		},
		{
			name: "text attachment without path",
			att: domain.Attachment{
				Type: domain.AttachmentTypeText,
			},
			expected: "text_attachment",
		},
		{
			name: "file attachment without path",
			att: domain.Attachment{
				Type: domain.AttachmentTypeFile,
			},
			expected: "file_attachment",
		},
		{
			name: "video attachment without path",
			att: domain.Attachment{
				Type: domain.AttachmentTypeVideo,
			},
			expected: "video_attachment",
		},
		{
			name: "audio attachment without path",
			att: domain.Attachment{
				Type: domain.AttachmentTypeAudio,
			},
			expected: "audio_attachment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAttachmentDisplayName(tt.att)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetDomainAttachmentDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		att      domain.Attachment
		expected string
	}{
		{
			name: "attachment with display name",
			att: domain.Attachment{
				Type: domain.AttachmentTypeImage,
				Name: "my-image.png",
			},
			expected: "my-image.png",
		},
		{
			name: "attachment without display name",
			att: domain.Attachment{
				Type: domain.AttachmentTypeImage,
			},
			expected: "image_attachment",
		},
		{
			name: "text attachment without display name",
			att: domain.Attachment{
				Type: domain.AttachmentTypeText,
			},
			expected: "text_attachment",
		},
		{
			name: "file attachment with name",
			att: domain.Attachment{
				Type: domain.AttachmentTypeFile,
				Name: "document.pdf",
			},
			expected: "document.pdf",
		},
		{
			name: "video attachment without name",
			att: domain.Attachment{
				Type: domain.AttachmentTypeVideo,
			},
			expected: "video_attachment",
		},
		{
			name: "audio attachment with empty name",
			att: domain.Attachment{
				Type: domain.AttachmentTypeAudio,
				Name: "",
			},
			expected: "audio_attachment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDomainAttachmentDisplayName(tt.att)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFilepathOperations(t *testing.T) {
	// Test with various path separators and patterns
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "unix style path",
			path:     "/path/to/file.txt",
			expected: "file.txt",
		},
		{
			name:     "windows style path",
			path:     "C:\\Users\\test\\file.txt",
			expected: filepath.Base("C:\\Users\\test\\file.txt"), // Let filepath handle it correctly
		},
		{
			name:     "relative path",
			path:     "relative/path/file.txt",
			expected: "file.txt",
		},
		{
			name:     "just filename",
			path:     "file.txt",
			expected: "file.txt",
		},
		{
			name:     "path with special characters",
			path:     "/path/to/file-with-dashes_and_underscores.txt",
			expected: "file-with-dashes_and_underscores.txt",
		},
		{
			name:     "path with spaces",
			path:     "/path/to/file with spaces.txt",
			expected: "file with spaces.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			att := domain.Attachment{
				Type:     domain.AttachmentTypeFile,
				FilePath: tt.path,
			}
			result := getAttachmentDisplayName(att)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCreateFileAttachmentLargeFile(t *testing.T) {
	// Test with a larger file to ensure base64 encoding works correctly
	tempDir := t.TempDir()
	largePath := filepath.Join(tempDir, "large.jpg")

	// Create a 1MB test file
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	err := os.WriteFile(largePath, largeData, 0644)
	if err != nil {
		t.Fatalf("failed to create large test file: %v", err)
	}

	attachment, err := createFileAttachmentFromPath(largePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the attachment was created
	if attachment.Type != domain.AttachmentTypeImage {
		t.Errorf("expected image type, got %s", attachment.Type)
	}

	// Decode the base64 content and verify it matches original
	decoded, err := base64.StdEncoding.DecodeString(string(attachment.Content))
	if err != nil {
		t.Fatalf("failed to decode base64 content: %v", err)
	}

	if len(decoded) != len(largeData) {
		t.Errorf("decoded data length mismatch: expected %d, got %d", len(largeData), len(decoded))
	}

	// Verify a sample of the data
	for i := 0; i < min(100, len(decoded)); i++ {
		if decoded[i] != largeData[i] {
			t.Errorf("data mismatch at index %d: expected %d, got %d", i, largeData[i], decoded[i])
		}
	}
}

func TestMimeTypeMapping(t *testing.T) {
	// Test that MIME types are correctly mapped
	mappings := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".m4a":  "audio/m4a",
		".ogg":  "audio/ogg",
		".mp4":  "video/mp4",
		".avi":  "video/avi",
		".mov":  "video/mov",
		".webm": "video/webm",
		".txt":  "application/octet-stream",
		".pdf":  "application/octet-stream",
		".doc":  "application/octet-stream",
	}

	tempDir := t.TempDir()
	testContent := []byte("test content")

	for ext, expectedMime := range mappings {
		t.Run(ext, func(t *testing.T) {
			filename := "test" + ext
			filePath := filepath.Join(tempDir, filename)

			err := os.WriteFile(filePath, testContent, 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			attachment, err := createFileAttachmentFromPath(filePath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if attachment.MimeType != expectedMime {
				t.Errorf("expected MIME type %s for extension %s, got %s", expectedMime, ext, attachment.MimeType)
			}
		})
	}
}

// Helper function for compatibility
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}