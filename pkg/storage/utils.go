// ABOUTME: Shared utility functions for the storage package
// ABOUTME: Provides common functions used by multiple storage backends

package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateSessionID generates a unique session ID with timestamp and random suffix
func GenerateSessionID() string {
	// Generate a random suffix using crypto/rand for guaranteed uniqueness
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp only if random generation fails
		return fmt.Sprintf("%s-%08d", time.Now().Format("20060102-150405-000000000"), time.Now().UnixNano()%100000000)
	}
	return fmt.Sprintf("%s-%08s", time.Now().Format("20060102-150405-000000000"), hex.EncodeToString(b))
}
