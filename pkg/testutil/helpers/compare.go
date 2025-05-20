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

	// Check model and provider from Conversation
	if expected.Conversation != nil && actual.Conversation != nil {
		if expected.Conversation.Model != actual.Conversation.Model {
			t.Errorf("Session Model mismatch: expected %q, got %q", expected.Conversation.Model, actual.Conversation.Model)
		}

		if expected.Conversation.Provider != actual.Conversation.Provider {
			t.Errorf("Session Provider mismatch: expected %q, got %q", expected.Conversation.Provider, actual.Conversation.Provider)
		}
	}

	// Compare tags
	if !StringSlicesEqual(expected.Tags, actual.Tags) {
		t.Errorf("Session Tags mismatch: expected %v, got %v", expected.Tags, actual.Tags)
	}

	// Compare message count (if conversations exist)
	if expected.Conversation != nil && actual.Conversation != nil {
		if len(expected.Conversation.Messages) != len(actual.Conversation.Messages) {
			t.Errorf("Message count mismatch: expected %d, got %d",
				len(expected.Conversation.Messages), len(actual.Conversation.Messages))
		}
	}
}

// StringSlicesEqual compares two string slices for equality
func StringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Sort both slices for comparison
	sortedA := make([]string, len(a))
	sortedB := make([]string, len(b))
	copy(sortedA, a)
	copy(sortedB, b)
	sort.Strings(sortedA)
	sort.Strings(sortedB)

	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}

	return true
}

// MapsEqual compares two maps for equality
func MapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valA := range a {
		valB, exists := b[key]
		if !exists {
			return false
		}

		if !reflect.DeepEqual(valA, valB) {
			return false
		}
	}

	return true
}

// AssertErrorContains checks if an error contains expected substring
func AssertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Errorf("Expected error containing %q, but got nil", expected)
		return
	}

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error containing %q, but got %q", expected, err.Error())
	}
}

// AssertNoError checks that no error occurred
func AssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
}

// CompareDomainMessages compares slices of domain messages
func CompareDomainMessages(t *testing.T, expected, actual []domain.Message) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("Message count mismatch: expected %d, got %d", len(expected), len(actual))
		return
	}

	for i := range expected {
		expectedMsg := &expected[i]
		actualMsg := &actual[i]
		CompareMessages(t, expectedMsg, actualMsg)
	}
}

// CompareSessionInfo compares two SessionInfo objects
func CompareSessionInfo(t *testing.T, expected, actual *domain.SessionInfo) {
	t.Helper()

	if expected.ID != actual.ID {
		t.Errorf("SessionInfo ID mismatch: expected %q, got %q", expected.ID, actual.ID)
	}

	if expected.Name != actual.Name {
		t.Errorf("SessionInfo Name mismatch: expected %q, got %q", expected.Name, actual.Name)
	}

	if expected.MessageCount != actual.MessageCount {
		t.Errorf("SessionInfo MessageCount mismatch: expected %d, got %d",
			expected.MessageCount, actual.MessageCount)
	}

	if expected.Model != actual.Model {
		t.Errorf("SessionInfo Model mismatch: expected %q, got %q", expected.Model, actual.Model)
	}

	if expected.Provider != actual.Provider {
		t.Errorf("SessionInfo Provider mismatch: expected %q, got %q", expected.Provider, actual.Provider)
	}

	if !StringSlicesEqual(expected.Tags, actual.Tags) {
		t.Errorf("SessionInfo Tags mismatch: expected %v, got %v", expected.Tags, actual.Tags)
	}
}
