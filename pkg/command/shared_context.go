// ABOUTME: Shared context for preserving state between command executions
// ABOUTME: Provides thread-safe storage for cross-command state in REPL sessions

package command

import (
	"sync"
)

// SharedContext holds persistent state between command executions
type SharedContext struct {
	mu    sync.RWMutex
	state map[string]interface{}
}

// NewSharedContext creates a new shared context
func NewSharedContext() *SharedContext {
	return &SharedContext{
		state: make(map[string]interface{}),
	}
}

// Get retrieves a value from the shared context
func (sc *SharedContext) Get(key string) (interface{}, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	value, exists := sc.state[key]
	return value, exists
}

// Set stores a value in the shared context
func (sc *SharedContext) Set(key string, value interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.state[key] = value
}

// Delete removes a value from the shared context
func (sc *SharedContext) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	delete(sc.state, key)
}

// GetString retrieves a string value from the shared context
func (sc *SharedContext) GetString(key string) (string, bool) {
	if value, exists := sc.Get(key); exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt retrieves an integer value from the shared context
func (sc *SharedContext) GetInt(key string) (int, bool) {
	if value, exists := sc.Get(key); exists {
		if num, ok := value.(int); ok {
			return num, true
		}
	}
	return 0, false
}

// GetFloat64 retrieves a float64 value from the shared context
func (sc *SharedContext) GetFloat64(key string) (float64, bool) {
	if value, exists := sc.Get(key); exists {
		if num, ok := value.(float64); ok {
			return num, true
		}
	}
	return 0, false
}

// GetBool retrieves a boolean value from the shared context
func (sc *SharedContext) GetBool(key string) (bool, bool) {
	if value, exists := sc.Get(key); exists {
		if b, ok := value.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// Clone creates a copy of the shared context state
func (sc *SharedContext) Clone() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	clone := make(map[string]interface{}, len(sc.state))
	for k, v := range sc.state {
		clone[k] = v
	}
	return clone
}

// Clear removes all values from the shared context
func (sc *SharedContext) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.state = make(map[string]interface{})
}

// Common keys used in shared context
const (
	// Model settings
	SharedContextModel       = "model"
	SharedContextProvider    = "provider"
	SharedContextTemperature = "temperature"
	SharedContextMaxTokens   = "max_tokens"
	SharedContextStream      = "stream"

	// Session state
	SharedContextSessionID   = "session_id"
	SharedContextSessionName = "session_name"

	// UI state
	SharedContextMultiline = "multiline"
	SharedContextPrompt    = "prompt"
	SharedContextVerbosity = "verbosity"
	SharedContextOutput    = "output"

	// Attachments
	SharedContextPendingAttachments = "pending_attachments"
)
