// ABOUTME: Tests for path string manipulation utilities
// ABOUTME: Validates functions for path extraction, joining, and normalization

package stringutil

import (
	"testing"
)

func TestExpandPath(t *testing.T) {
	// Test with non-home path
	path := "/usr/local/bin"
	result := ExpandPath(path)
	if result != path {
		t.Errorf("ExpandPath failed to return unchanged path: got %s, want %s", result, path)
	}

	// We can't reliably test home expansion in a unit test environment
	// since it depends on the user's home directory
}

func TestExtractFileName(t *testing.T) {
	testCases := []struct {
		path          string
		withExtension bool
		expected      string
	}{
		{"/user/home/file.txt", true, "file.txt"},
		{"/user/home/file.txt", false, "file"},
		{"file.jpg", true, "file.jpg"},
		{"file.jpg", false, "file"},
		{"file", true, "file"},
		{"file", false, "file"},
		{"", true, ""},
		{"", false, ""},
	}

	for _, tc := range testCases {
		result := ExtractFileName(tc.path, tc.withExtension)
		if result != tc.expected {
			t.Errorf("ExtractFileName(%s, %v) failed: got %s, want %s",
				tc.path, tc.withExtension, result, tc.expected)
		}
	}
}

func TestExtractExtension(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"/user/home/file.txt", ".txt"},
		{"file.jpg", ".jpg"},
		{"file", ""},
		{"/user/home/.gitconfig", ".gitconfig"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := ExtractExtension(tc.path)
		if result != tc.expected {
			t.Errorf("ExtractExtension(%s) failed: got %s, want %s",
				tc.path, result, tc.expected)
		}
	}
}

func TestJoinConfigPath(t *testing.T) {
	testCases := []struct {
		parts    []string
		expected string
	}{
		{[]string{"config", "provider"}, "config.provider"},
		{[]string{"provider", "openai", "api_key"}, "provider.openai.api_key"},
		{[]string{}, ""},
		{[]string{"single"}, "single"},
	}

	for _, tc := range testCases {
		result := JoinConfigPath(tc.parts...)
		if result != tc.expected {
			t.Errorf("JoinConfigPath(%v) failed: got %s, want %s",
				tc.parts, result, tc.expected)
		}
	}
}

func TestSplitConfigPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected []string
	}{
		{"config.provider", []string{"config", "provider"}},
		{"provider.openai.api_key", []string{"provider", "openai", "api_key"}},
		{"", []string{}},
		{"single", []string{"single"}},
	}

	for _, tc := range testCases {
		result := SplitConfigPath(tc.path)
		if len(result) != len(tc.expected) {
			t.Errorf("SplitConfigPath(%s) failed: got %v, want %v",
				tc.path, result, tc.expected)
			continue
		}

		for i := range result {
			if result[i] != tc.expected[i] {
				t.Errorf("SplitConfigPath(%s) failed at index %d: got %s, want %s",
					tc.path, i, result[i], tc.expected[i])
			}
		}
	}
}

func TestSafeFilename(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"file.txt", "file_txt"},
		{"file?.txt", "file__txt"},
		{"file/name", "file_name"},
		{"file\\name", "file_name"},
		{"file:name", "file_name"},
		{"file*name", "file_name"},
		{"file?name", "file_name"},
		{"file\"name", "file_name"},
		{"file<name", "file_name"},
		{"file>name", "file_name"},
		{"file|name", "file_name"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := SafeFilename(tc.name)
		if result != tc.expected {
			t.Errorf("SafeFilename(%s) failed: got %s, want %s",
				tc.name, result, tc.expected)
		}
	}
}
