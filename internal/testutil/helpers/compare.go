// ABOUTME: Comparison helper functions for testing
// ABOUTME: Provides utilities for comparing complex objects in tests

package helpers

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
)

// CompareMessages compares two messages for equality
func CompareMessages(t *testing.T, expected, actual *domain.Message) {
	t.Helper()

	if expected.ID != actual.ID {
		t.Errorf("Message ID mismatch: expected %q, got %q", expected.ID, actual.ID)
	}

	if expected.Role != actual.Role {
		t.Errorf("Message Role mismatch: expected %q, got %q", expected.Role, actual.Role)
	}

	if expected.Content != actual.Content {
		t.Errorf("Message Content mismatch: expected %q, got %q", expected.Content, actual.Content)
	}

	// Compare attachments count
	if len(expected.Attachments) != len(actual.Attachments) {
		t.Errorf("Attachment count mismatch: expected %d, got %d",
			len(expected.Attachments), len(actual.Attachments))
	}
}

// CompareSessions compares two sessions for equality
func CompareSessions(t *testing.T, expected, actual *domain.Session) {
	t.Helper()

	if expected.ID != actual.ID {
		t.Errorf("Session ID mismatch: expected %q, got %q", expected.ID, actual.ID)
	}

	if expected.Name != actual.Name {
		t.Errorf("Session Name mismatch: expected %q, got %q", expected.Name, actual.Name)
	}

	// Compare model and provider through conversation
	if expected.Conversation != nil && actual.Conversation != nil {
		if expected.Conversation.Model != actual.Conversation.Model {
			t.Errorf("Conversation Model mismatch: expected %q, got %q",
				expected.Conversation.Model, actual.Conversation.Model)
		}

		if expected.Conversation.Provider != actual.Conversation.Provider {
			t.Errorf("Conversation Provider mismatch: expected %q, got %q",
				expected.Conversation.Provider, actual.Conversation.Provider)
		}
	}

	// Compare metadata
	if len(expected.Metadata) != len(actual.Metadata) {
		t.Errorf("Metadata count mismatch: expected %d, got %d",
			len(expected.Metadata), len(actual.Metadata))
	} else {
		for k, v := range expected.Metadata {
			actVal, ok := actual.Metadata[k]
			if !ok {
				t.Errorf("Metadata missing key %q", k)
				continue
			}

			if v != actVal {
				t.Errorf("Metadata value mismatch for key %q: expected %v, got %v", k, v, actVal)
			}
		}
	}

	// Compare conversations
	if expected.Conversation != nil && actual.Conversation != nil {
		if expected.Conversation.ID != actual.Conversation.ID {
			t.Errorf("Conversation ID mismatch: expected %q, got %q",
				expected.Conversation.ID, actual.Conversation.ID)
		}

		if len(expected.Conversation.Messages) != len(actual.Conversation.Messages) {
			t.Errorf("Conversation message count mismatch: expected %d, got %d",
				len(expected.Conversation.Messages), len(actual.Conversation.Messages))
		}
	}

	// Compare branch information
	if expected.ParentID != actual.ParentID {
		t.Errorf("ParentID mismatch: expected %q, got %q",
			expected.ParentID, actual.ParentID)
	}

	if expected.BranchName != actual.BranchName {
		t.Errorf("BranchName mismatch: expected %q, got %q",
			expected.BranchName, actual.BranchName)
	}

	if expected.BranchPoint != actual.BranchPoint {
		t.Errorf("BranchPoint mismatch: expected %d, got %d",
			expected.BranchPoint, actual.BranchPoint)
	}

	// Compare tags (order-independent)
	if !StringSlicesEqual(expected.Tags, actual.Tags) {
		t.Errorf("Tags mismatch: expected %v, got %v", expected.Tags, actual.Tags)
	}
}

// CompareStringMaps compares two string maps for equality
func CompareStringMaps(t *testing.T, expected, actual map[string]string, label string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("%s count mismatch: expected %d, got %d", label, len(expected), len(actual))
		return
	}

	for k, v := range expected {
		actVal, ok := actual[k]
		if !ok {
			t.Errorf("%s missing key %q", label, k)
			continue
		}

		if v != actVal {
			t.Errorf("%s value mismatch for key %q: expected %q, got %q", label, k, v, actVal)
		}
	}
}

// StringSlicesEqual determines if two string slices are equal (order-independent)
func StringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Sort copies of the slices
	sortedA := make([]string, len(a))
	sortedB := make([]string, len(b))

	copy(sortedA, a)
	copy(sortedB, b)

	sort.Strings(sortedA)
	sort.Strings(sortedB)

	return reflect.DeepEqual(sortedA, sortedB)
}

// StringSlicesEqualOrdered determines if two string slices are equal (order-dependent)
func StringSlicesEqualOrdered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// StringContains checks if a string contains a substring (case-insensitive)
func StringContains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}
