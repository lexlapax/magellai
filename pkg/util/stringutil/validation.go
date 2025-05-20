// ABOUTME: String validation utilities for common validation tasks
// ABOUTME: Provides functions for validating and sanitizing string inputs

package stringutil

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// Common validation regular expressions
var (
	// EmailRegex matches a valid email address
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// URLRegex matches a valid URL (simplified version)
	URLRegex = regexp.MustCompile(`^(http|https)://[a-zA-Z0-9][-a-zA-Z0-9.]*\.[a-zA-Z]{2,}(/[-a-zA-Z0-9%_.~#?&=]*)?$`)

	// AlphanumericRegex matches only alphanumeric characters
	AlphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	// AlphanumericWithSpacesRegex matches alphanumeric characters and spaces
	AlphanumericWithSpacesRegex = regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)

	// FilenameRegex matches safe file names
	FilenameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`)
)

// IsValidEmail checks if a string is a valid email address
func IsValidEmail(email string) bool {
	return email != "" && EmailRegex.MatchString(email)
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(url string) bool {
	return url != "" && URLRegex.MatchString(url)
}

// IsAlphanumeric checks if a string contains only alphanumeric characters
func IsAlphanumeric(str string) bool {
	return str != "" && AlphanumericRegex.MatchString(str)
}

// IsAlphanumericWithSpaces checks if a string contains only alphanumeric characters and spaces
func IsAlphanumericWithSpaces(str string) bool {
	return str != "" && AlphanumericWithSpacesRegex.MatchString(str)
}

// IsValidFilename checks if a string is a valid and safe filename
func IsValidFilename(filename string) bool {
	return filename != "" &&
		!strings.ContainsAny(filename, "/\\:*?\"<>|") &&
		filepath.Base(filename) == filename
}

// IsPrintable checks if a string contains only printable characters
func IsPrintable(str string) bool {
	if str == "" {
		return false
	}

	for _, r := range str {
		if !unicode.IsPrint(r) {
			return false
		}
	}

	return true
}

// HasMinLength checks if a string is at least the specified length
func HasMinLength(str string, minLength int) bool {
	return len(str) >= minLength
}

// HasMaxLength checks if a string is at most the specified length
func HasMaxLength(str string, maxLength int) bool {
	return len(str) <= maxLength
}

// IsInRange checks if a string length is within the specified range
func IsInRange(str string, minLength, maxLength int) bool {
	length := len(str)
	return length >= minLength && length <= maxLength
}

// Contains checks if a string contains a substring
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// ContainsAny checks if a string contains any of the given characters
func ContainsAny(str, chars string) bool {
	return strings.ContainsAny(str, chars)
}

// ContainsAll checks if a string contains all of the given substrings
func ContainsAll(str string, substrs ...string) bool {
	for _, substr := range substrs {
		if !strings.Contains(str, substr) {
			return false
		}
	}
	return true
}

// IsValidID checks if a string is a valid ID in the format used by the application
// Valid IDs have a prefix, timestamp, and random string, separated by hyphens
func IsValidID(id string) bool {
	parts := strings.Split(id, "-")

	// Must have 2 or 3 parts (with or without prefix)
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}

	// Check timestamp format if it has 3 parts
	if len(parts) == 3 {
		// Second part should be timestamp in format 20060102T150405Z
		timestampRegex := regexp.MustCompile(`^\d{8}T\d{6}Z$`)
		if !timestampRegex.MatchString(parts[1]) {
			return false
		}

		// Third part should be random hex string
		randomRegex := regexp.MustCompile(`^[0-9a-f]+$`)
		if !randomRegex.MatchString(parts[2]) {
			return false
		}
	} else {
		// First part should be timestamp in format 20060102T150405Z
		timestampRegex := regexp.MustCompile(`^\d{8}T\d{6}Z$`)
		if !timestampRegex.MatchString(parts[0]) {
			return false
		}

		// Second part should be random hex string
		randomRegex := regexp.MustCompile(`^[0-9a-f]+$`)
		if !randomRegex.MatchString(parts[1]) {
			return false
		}
	}

	return true
}
