// ABOUTME: Tests for string validation utility functions
// ABOUTME: Validates regular expressions and validation helper functions

package stringutil

import (
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user@subdomain.example.com",
	}

	invalidEmails := []string{
		"",
		"invalid",
		"invalid@",
		"@example.com",
		"user@.com",
		"user@example",
	}

	for _, email := range validEmails {
		if !IsValidEmail(email) {
			t.Errorf("IsValidEmail(%s) should return true", email)
		}
	}

	for _, email := range invalidEmails {
		if IsValidEmail(email) {
			t.Errorf("IsValidEmail(%s) should return false", email)
		}
	}
}

func TestIsValidURL(t *testing.T) {
	validURLs := []string{
		"http://example.com",
		"https://example.com",
		"http://example.com/path",
		"https://example.com/path?query=value",
		"https://subdomain.example.com",
	}

	invalidURLs := []string{
		"",
		"invalid",
		"example.com",
		"ftp://example.com",
		"http://",
		"http://.com",
	}

	for _, url := range validURLs {
		if !IsValidURL(url) {
			t.Errorf("IsValidURL(%s) should return true", url)
		}
	}

	for _, url := range invalidURLs {
		if IsValidURL(url) {
			t.Errorf("IsValidURL(%s) should return false", url)
		}
	}
}

func TestIsAlphanumeric(t *testing.T) {
	validInputs := []string{
		"abc123",
		"ABC123",
		"123456",
	}

	invalidInputs := []string{
		"",
		"abc 123",
		"abc-123",
		"abc_123",
		"abc.123",
	}

	for _, input := range validInputs {
		if !IsAlphanumeric(input) {
			t.Errorf("IsAlphanumeric(%s) should return true", input)
		}
	}

	for _, input := range invalidInputs {
		if IsAlphanumeric(input) {
			t.Errorf("IsAlphanumeric(%s) should return false", input)
		}
	}
}

func TestIsAlphanumericWithSpaces(t *testing.T) {
	validInputs := []string{
		"abc 123",
		"ABC 123",
		"hello world",
	}

	invalidInputs := []string{
		"",
		"abc-123",
		"abc_123",
		"abc.123",
	}

	for _, input := range validInputs {
		if !IsAlphanumericWithSpaces(input) {
			t.Errorf("IsAlphanumericWithSpaces(%s) should return true", input)
		}
	}

	for _, input := range invalidInputs {
		if IsAlphanumericWithSpaces(input) {
			t.Errorf("IsAlphanumericWithSpaces(%s) should return false", input)
		}
	}
}

func TestIsValidFilename(t *testing.T) {
	validFilenames := []string{
		"file.txt",
		"file-name.txt",
		"file_name.txt",
		"file.name.txt",
	}

	invalidFilenames := []string{
		"",
		"file/name.txt",
		"file\\name.txt",
		"file:name.txt",
		"file*name.txt",
		"file?name.txt",
		"file\"name.txt",
		"file<name.txt",
		"file>name.txt",
		"file|name.txt",
	}

	for _, filename := range validFilenames {
		if !IsValidFilename(filename) {
			t.Errorf("IsValidFilename(%s) should return true", filename)
		}
	}

	for _, filename := range invalidFilenames {
		if IsValidFilename(filename) {
			t.Errorf("IsValidFilename(%s) should return false", filename)
		}
	}
}

func TestIsPrintable(t *testing.T) {
	validInputs := []string{
		"Hello, world!",
		"1234567890",
		"!@#$%^&*()",
	}

	invalidInputs := []string{
		"",
		"Hello\nworld",
		"Hello\tworld",
		"Hello\x00world",
	}

	for _, input := range validInputs {
		if !IsPrintable(input) {
			t.Errorf("IsPrintable(%s) should return true", input)
		}
	}

	for _, input := range invalidInputs {
		if IsPrintable(input) {
			t.Errorf("IsPrintable(%s) should return false", input)
		}
	}
}

func TestLengthValidators(t *testing.T) {
	// Test HasMinLength
	if !HasMinLength("hello", 5) {
		t.Errorf("HasMinLength should return true for string of exact minimum length")
	}
	if !HasMinLength("hello", 4) {
		t.Errorf("HasMinLength should return true for string longer than minimum length")
	}
	if HasMinLength("hello", 6) {
		t.Errorf("HasMinLength should return false for string shorter than minimum length")
	}

	// Test HasMaxLength
	if !HasMaxLength("hello", 5) {
		t.Errorf("HasMaxLength should return true for string of exact maximum length")
	}
	if !HasMaxLength("hello", 6) {
		t.Errorf("HasMaxLength should return true for string shorter than maximum length")
	}
	if HasMaxLength("hello", 4) {
		t.Errorf("HasMaxLength should return false for string longer than maximum length")
	}

	// Test IsInRange
	if !IsInRange("hello", 4, 6) {
		t.Errorf("IsInRange should return true for string within range")
	}
	if !IsInRange("hello", 5, 5) {
		t.Errorf("IsInRange should return true for string of exact range limits")
	}
	if IsInRange("hello", 6, 10) {
		t.Errorf("IsInRange should return false for string shorter than minimum range")
	}
	if IsInRange("hello", 1, 4) {
		t.Errorf("IsInRange should return false for string longer than maximum range")
	}
}

func TestContainsValidators(t *testing.T) {
	// Test Contains
	if !Contains("hello world", "world") {
		t.Errorf("Contains should return true for string containing substring")
	}
	if Contains("hello world", "universe") {
		t.Errorf("Contains should return false for string not containing substring")
	}

	// Test ContainsAny
	if !ContainsAny("hello world", "abcde") {
		t.Errorf("ContainsAny should return true for string containing any of the characters")
	}
	if ContainsAny("hello world", "xyz") {
		t.Errorf("ContainsAny should return false for string not containing any of the characters")
	}

	// Test ContainsAll
	if !ContainsAll("hello world", "hello", "world") {
		t.Errorf("ContainsAll should return true for string containing all substrings")
	}
	if ContainsAll("hello world", "hello", "universe") {
		t.Errorf("ContainsAll should return false for string not containing all substrings")
	}
}

func TestIsValidID(t *testing.T) {
	// Create a valid ID using our generator
	validIDWithPrefix := GenerateID("test", 8)
	validIDWithoutPrefix := GenerateID("", 8)

	invalidIDs := []string{
		"",
		"invalid",
		"invalid-id",
		"prefix-only-",
		"-invalid-format",
		"too-many-parts-here-invalid",
		"test-20060102150405-invalid", // Invalid timestamp format
		"test-20060102T150405Z-xyz",   // Invalid random part
	}

	// Valid IDs from our generator should pass validation
	if !IsValidID(validIDWithPrefix) {
		t.Errorf("IsValidID(%s) should return true for valid ID with prefix", validIDWithPrefix)
	}

	if !IsValidID(validIDWithoutPrefix) {
		t.Errorf("IsValidID(%s) should return true for valid ID without prefix", validIDWithoutPrefix)
	}

	// Invalid IDs should fail validation
	for _, id := range invalidIDs {
		if IsValidID(id) {
			t.Errorf("IsValidID(%s) should return false for invalid ID", id)
		}
	}
}
