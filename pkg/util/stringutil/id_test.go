// ABOUTME: Tests for ID generation utility functions
// ABOUTME: Validates ID generation for sessions, messages, and other entities

package stringutil

import (
	"regexp"
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Test with prefix
	id := GenerateID("test", 8)

	// Verify format: prefix-timestamp-randomstring
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("GenerateID with prefix should have 3 parts, got %d: %s", len(parts), id)
	}

	if parts[0] != "test" {
		t.Errorf("GenerateID prefix incorrect: got %s, want %s", parts[0], "test")
	}

	// Timestamp should match format 20060102T150405Z
	timestampRegex := regexp.MustCompile(`^\d{8}T\d{6}Z$`)
	if !timestampRegex.MatchString(parts[1]) {
		t.Errorf("GenerateID timestamp format incorrect: %s", parts[1])
	}

	// Random part should be 8 characters
	if len(parts[2]) != 8 {
		t.Errorf("GenerateID random part should be 8 characters, got %d: %s",
			len(parts[2]), parts[2])
	}

	// Test without prefix
	id = GenerateID("", 8)

	// Verify format: timestamp-randomstring
	parts = strings.Split(id, "-")
	if len(parts) != 2 {
		t.Errorf("GenerateID without prefix should have 2 parts, got %d: %s", len(parts), id)
	}

	// Timestamp should match format 20060102T150405Z
	if !timestampRegex.MatchString(parts[0]) {
		t.Errorf("GenerateID timestamp format incorrect: %s", parts[0])
	}

	// Random part should be 8 characters
	if len(parts[1]) != 8 {
		t.Errorf("GenerateID random part should be 8 characters, got %d: %s",
			len(parts[1]), parts[1])
	}
}

func TestGenerateSessionID(t *testing.T) {
	id := GenerateSessionID()

	// Verify format: ses-timestamp-randomstring
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("GenerateSessionID should have 3 parts, got %d: %s", len(parts), id)
	}

	if parts[0] != "ses" {
		t.Errorf("GenerateSessionID prefix incorrect: got %s, want %s", parts[0], "ses")
	}
}

func TestGenerateMessageID(t *testing.T) {
	id := GenerateMessageID()

	// Verify format: msg-timestamp-randomstring
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("GenerateMessageID should have 3 parts, got %d: %s", len(parts), id)
	}

	if parts[0] != "msg" {
		t.Errorf("GenerateMessageID prefix incorrect: got %s, want %s", parts[0], "msg")
	}
}

func TestGenerateAttachmentID(t *testing.T) {
	id := GenerateAttachmentID()

	// Verify format: att-timestamp-randomstring
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("GenerateAttachmentID should have 3 parts, got %d: %s", len(parts), id)
	}

	if parts[0] != "att" {
		t.Errorf("GenerateAttachmentID prefix incorrect: got %s, want %s", parts[0], "att")
	}
}

func TestGenerateRequestID(t *testing.T) {
	id := GenerateRequestID()

	// Verify format: req-timestamp-randomstring
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("GenerateRequestID should have 3 parts, got %d: %s", len(parts), id)
	}

	if parts[0] != "req" {
		t.Errorf("GenerateRequestID prefix incorrect: got %s, want %s", parts[0], "req")
	}
}

func TestIDUniqueness(t *testing.T) {
	// Generate a large number of IDs and check for duplicates
	idCount := 1000
	ids := make(map[string]bool, idCount)

	for i := 0; i < idCount; i++ {
		id := GenerateID("test", 8)
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}
