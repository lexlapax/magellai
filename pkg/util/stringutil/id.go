// ABOUTME: ID generation utilities for creating unique identifiers
// ABOUTME: Provides standardized ID generation for sessions, messages, and other entities

package stringutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID creates a unique ID with timestamp prefix and random suffix
// This is a standardized ID generator for the application
func GenerateID(prefix string, randomLength int) string {
	if randomLength <= 0 {
		randomLength = 8
	}

	// Get current timestamp in a sortable format (RFC3339 without special chars)
	timestamp := time.Now().UTC().Format("20060102T150405Z")

	// Generate random suffix
	randBytes := make([]byte, randomLength/2+1)
	_, err := rand.Read(randBytes)
	if err != nil {
		// Fallback to less random but still unique value if crypto/rand fails
		randStr := fmt.Sprintf("%x", time.Now().UnixNano())
		if len(randStr) > randomLength {
			randStr = randStr[:randomLength]
		}
		return fmt.Sprintf("%s-%s-%s", prefix, timestamp, randStr)
	}

	// Convert to hex string and trim to requested length
	randStr := hex.EncodeToString(randBytes)
	if len(randStr) > randomLength {
		randStr = randStr[:randomLength]
	}

	// Format: prefix-timestamp-randomstring
	if prefix != "" {
		return fmt.Sprintf("%s-%s-%s", prefix, timestamp, randStr)
	}

	return fmt.Sprintf("%s-%s", timestamp, randStr)
}

// GenerateSessionID creates a unique session ID
func GenerateSessionID() string {
	return GenerateID("ses", 8)
}

// GenerateMessageID creates a unique message ID
func GenerateMessageID() string {
	return GenerateID("msg", 8)
}

// GenerateAttachmentID creates a unique attachment ID
func GenerateAttachmentID() string {
	return GenerateID("att", 8)
}

// GenerateRequestID creates a unique request ID for tracking
func GenerateRequestID() string {
	return GenerateID("req", 8)
}
