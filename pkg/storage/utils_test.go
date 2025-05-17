// ABOUTME: Tests for storage package utility functions
// ABOUTME: Ensures shared functions work correctly across all backends

package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSessionID(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateSessionID()
		assert.NotEmpty(t, id)
		assert.False(t, ids[id], "Duplicate ID generated: %s", id)
		ids[id] = true

		// Verify format: YYYYMMDD-HHMMSS-NNNNNNNNN-RRRRRRRR (8 hex chars)
		assert.True(t, strings.Contains(id, "-"))
		assert.Regexp(t, `^\d{8}-\d{6}-\d{9}-[0-9a-f]{8}$`, id)
	}

	// Test that IDs have correct number of segments
	id := GenerateSessionID()
	parts := strings.Split(id, "-")
	assert.Len(t, parts, 4, "ID should have 4 parts separated by hyphens")

	// Test individual segments
	assert.Len(t, parts[0], 8, "Date part should be 8 digits (YYYYMMDD)")
	assert.Len(t, parts[1], 6, "Time part should be 6 digits (HHMMSS)")
	assert.Len(t, parts[2], 9, "Nanosecond part should be 9 digits")
	assert.Len(t, parts[3], 8, "Random hex part should be 8 characters")

	// Verify hex part is valid hexadecimal
	assert.Regexp(t, `^[0-9a-f]+$`, parts[3], "Last part should be hexadecimal")
}
