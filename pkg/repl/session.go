// ABOUTME: Manages REPL sessions including persistence, save/load functionality, and metadata
// ABOUTME: Provides methods for creating, saving, loading, and listing sessions

package repl

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
)

// No need to seed rand as of Go 1.20

// Session represents a complete REPL session with conversation and metadata
type Session struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name,omitempty"`
	Conversation *Conversation          `json:"conversation"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Created      time.Time              `json:"created"`
	Updated      time.Time              `json:"updated"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SessionManager handles session persistence and lifecycle
type SessionManager struct {
	StorageDir string
}

// NewSessionManager creates a new session manager with the given storage directory
func NewSessionManager(storageDir string) (*SessionManager, error) {
	logging.LogDebug("Creating session manager", "storageDir", storageDir)

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logging.LogError(err, "Failed to create storage directory", "storageDir", storageDir)
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	logging.LogDebug("Session manager created successfully", "storageDir", storageDir)
	return &SessionManager{
		StorageDir: storageDir,
	}, nil
}

// NewSession creates a new session with a conversation
func (sm *SessionManager) NewSession(name string) *Session {
	now := time.Now()
	sessionID := generateSessionID()

	logging.LogInfo("Creating new session", "id", sessionID, "name", name)

	session := &Session{
		ID:           sessionID,
		Name:         name,
		Conversation: NewConversation(sessionID),
		Config:       make(map[string]interface{}),
		Created:      now,
		Updated:      now,
		Metadata:     make(map[string]interface{}),
	}

	logging.LogDebug("Session created successfully", "id", sessionID, "name", name)
	return session
}

// SaveSession persists a session to disk
func (sm *SessionManager) SaveSession(session *Session) error {
	start := time.Now()
	session.Updated = time.Now()

	logging.LogInfo("Saving session", "id", session.ID, "name", session.Name)

	// Create session file path
	filename := fmt.Sprintf("%s.json", session.ID)
	filepath := filepath.Join(sm.StorageDir, filename)

	logging.LogDebug("Writing session file", "path", filepath)

	// Marshal session to JSON
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		logging.LogError(err, "Failed to marshal session", "id", session.ID)
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		logging.LogError(err, "Failed to write session file", "path", filepath)
		return fmt.Errorf("failed to write session file: %w", err)
	}

	duration := time.Since(start)
	logging.LogInfo("Session saved successfully", "id", session.ID, "path", filepath)
	logging.LogDebug("Session save duration", "id", session.ID, "duration", duration)
	return nil
}

// LoadSession loads a session from disk by ID
func (sm *SessionManager) LoadSession(id string) (*Session, error) {
	start := time.Now()
	logging.LogInfo("Loading session", "id", id)

	filename := fmt.Sprintf("%s.json", id)
	filepath := filepath.Join(sm.StorageDir, filename)

	logging.LogDebug("Reading session file", "path", filepath)

	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			logging.LogDebug("Session not found", "id", id)
			return nil, fmt.Errorf("session not found: %s", id)
		}
		logging.LogError(err, "Failed to read session file", "path", filepath)
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Unmarshal session
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		logging.LogError(err, "Failed to unmarshal session", "id", id)
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	duration := time.Since(start)
	logging.LogInfo("Session loaded successfully", "id", id, "name", session.Name)
	logging.LogDebug("Session load duration", "id", id, "duration", duration)
	return &session, nil
}

// ListSessions returns a list of all available sessions
func (sm *SessionManager) ListSessions() ([]*SessionInfo, error) {
	start := time.Now()
	logging.LogDebug("Listing sessions", "storageDir", sm.StorageDir)

	entries, err := os.ReadDir(sm.StorageDir)
	if err != nil {
		logging.LogError(err, "Failed to read storage directory", "storageDir", sm.StorageDir)
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var sessions []*SessionInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Read session file to get basic info
		filepath := filepath.Join(sm.StorageDir, entry.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			logging.LogDebug("Failed to read session file", "path", filepath, "error", err)
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			logging.LogDebug("Failed to unmarshal session", "path", filepath, "error", err)
			continue
		}

		info := &SessionInfo{
			ID:           session.ID,
			Name:         session.Name,
			Created:      session.Created,
			Updated:      session.Updated,
			MessageCount: len(session.Conversation.Messages),
			Tags:         session.Tags,
		}

		sessions = append(sessions, info)
	}

	duration := time.Since(start)
	logging.LogDebug("Listed sessions", "count", len(sessions))
	logging.LogDebug("Session list duration", "count", len(sessions), "duration", duration)
	return sessions, nil
}

// DeleteSession removes a session from disk
func (sm *SessionManager) DeleteSession(id string) error {
	logging.LogInfo("Deleting session", "id", id)

	filename := fmt.Sprintf("%s.json", id)
	filepath := filepath.Join(sm.StorageDir, filename)

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			logging.LogDebug("Session not found for deletion", "id", id)
			return fmt.Errorf("session not found: %s", id)
		}
		logging.LogError(err, "Failed to delete session", "id", id, "path", filepath)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	logging.LogInfo("Session deleted successfully", "id", id)
	return nil
}

// SearchResult represents a search result with context
type SearchResult struct {
	Session  *SessionInfo
	Matches  []SearchMatch
}

// SearchMatch represents a single match with context
type SearchMatch struct {
	Type     string // "message", "system_prompt", "name", "tag"
	Role     string // for messages: "user", "assistant", "system"
	Content  string // the actual matched content snippet
	Context  string // surrounding context
	Position int    // message index if applicable
}

// SearchSessions searches for sessions by text in messages, system prompts, names, and tags
func (sm *SessionManager) SearchSessions(query string) ([]*SearchResult, error) {
	start := time.Now()
	logging.LogInfo("Searching sessions", "query", query)

	sessions, err := sm.ListSessions()
	if err != nil {
		logging.LogError(err, "Failed to list sessions for search")
		return nil, err
	}

	var results []*SearchResult
	lowerQuery := strings.ToLower(query)

	for _, info := range sessions {
		// Load full session to search content
		session, err := sm.LoadSession(info.ID)
		if err != nil {
			logging.LogDebug("Failed to load session for search", "id", info.ID, "error", err)
			continue
		}

		var matches []SearchMatch

		// Search in system prompt
		if session.Conversation.SystemPrompt != "" {
			lowerContent := strings.ToLower(session.Conversation.SystemPrompt)
			if strings.Contains(lowerContent, lowerQuery) {
				match := SearchMatch{
					Type:    "system_prompt",
					Content: extractSnippet(session.Conversation.SystemPrompt, query, 50),
					Context: "System Prompt",
				}
				matches = append(matches, match)
			}
		}

		// Search in messages
		for idx, msg := range session.Conversation.Messages {
			lowerContent := strings.ToLower(msg.Content)
			if strings.Contains(lowerContent, lowerQuery) {
				match := SearchMatch{
					Type:     "message",
					Role:     msg.Role,
					Content:  extractSnippet(msg.Content, query, 50),
					Context:  fmt.Sprintf("Message %d (%s)", idx+1, msg.Role),
					Position: idx,
				}
				matches = append(matches, match)
			}
		}

		// Search in session name
		if strings.Contains(strings.ToLower(session.Name), lowerQuery) {
			match := SearchMatch{
				Type:    "name",
				Content: session.Name,
				Context: "Session Name",
			}
			matches = append(matches, match)
		}

		// Search in tags
		for _, tag := range session.Tags {
			if strings.Contains(strings.ToLower(tag), lowerQuery) {
				match := SearchMatch{
					Type:    "tag",
					Content: tag,
					Context: "Tag",
				}
				matches = append(matches, match)
			}
		}

		// Add to results if matches found
		if len(matches) > 0 {
			result := &SearchResult{
				Session: info,
				Matches: matches,
			}
			results = append(results, result)
		}
	}

	duration := time.Since(start)
	logging.LogInfo("Search completed", "query", query, "sessions_searched", len(sessions), "results", len(results))
	logging.LogDebug("Search duration", "duration", duration)
	return results, nil
}

// SessionInfo represents basic session information for listing
type SessionInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
	MessageCount int       `json:"message_count"`
	Tags         []string  `json:"tags,omitempty"`
}

// ExportSession exports a session to a writer in the specified format
func (sm *SessionManager) ExportSession(id string, format string, w io.Writer) error {
	logging.LogInfo("Exporting session", "id", id, "format", format)

	session, err := sm.LoadSession(id)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(session); err != nil {
			logging.LogError(err, "Failed to export session as JSON", "id", id)
			return err
		}
		logging.LogInfo("Session exported successfully as JSON", "id", id)
		return nil

	case "markdown":
		if err := sm.exportMarkdown(session, w); err != nil {
			logging.LogError(err, "Failed to export session as Markdown", "id", id)
			return err
		}
		logging.LogInfo("Session exported successfully as Markdown", "id", id)
		return nil

	default:
		logging.LogWarn("Unsupported export format", "format", format)
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// Helper function to export session as markdown
func (sm *SessionManager) exportMarkdown(session *Session, w io.Writer) error {
	logging.LogDebug("Exporting session as markdown", "id", session.ID)

	fmt.Fprintf(w, "# Session: %s\n\n", session.Name)
	fmt.Fprintf(w, "ID: %s\n", session.ID)
	fmt.Fprintf(w, "Created: %s\n", session.Created.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated: %s\n\n", session.Updated.Format(time.RFC3339))

	if len(session.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n\n", strings.Join(session.Tags, ", "))
	}

	fmt.Fprintln(w, "## Conversation")

	for _, msg := range session.Conversation.Messages {
		fmt.Fprintf(w, "### %s\n", title(msg.Role))
		fmt.Fprintf(w, "*%s*\n\n", msg.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(w, "%s\n\n", msg.Content)

		if len(msg.Attachments) > 0 {
			fmt.Fprintln(w, "Attachments:")
			for _, att := range msg.Attachments {
				name := att.FilePath
				if name == "" {
					name = string(att.Type) + "_attachment"
				}
				fmt.Fprintf(w, "- %s (%s)\n", name, att.MimeType)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}

// Helper function to generate session ID
func generateSessionID() string {
	// Use nanoseconds plus random component for uniqueness
	id := fmt.Sprintf("%s-%04d", time.Now().Format("20060102-150405-000000000"), rand.Intn(10000))
	logging.LogDebug("Generated session ID", "id", id)
	return id
}

// extractSnippet extracts a snippet with context around the matched query
func extractSnippet(content, query string, contextRadius int) string {
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)
	
	// Find the first occurrence of the query
	idx := strings.Index(lowerContent, lowerQuery)
	if idx == -1 {
		// This shouldn't happen, but if it does, return beginning of content
		if len(content) <= contextRadius*2 {
			return content
		}
		return content[:contextRadius*2] + "..."
	}
	
	// Calculate start and end positions for the snippet
	start := idx - contextRadius
	end := idx + len(query) + contextRadius
	
	// Ensure we don't go out of bounds
	prefix := ""
	suffix := ""
	
	if start < 0 {
		start = 0
	} else {
		prefix = "..."
	}
	
	if end > len(content) {
		end = len(content)
	} else {
		suffix = "..."
	}
	
	// Extract the snippet
	snippet := content[start:end]
	
	// If the snippet is at word boundaries, try to adjust
	if start > 0 && !isWordBoundary(content[start-1]) {
		// Find the previous word boundary
		for i := start; i > 0; i-- {
			if isWordBoundary(content[i-1]) {
				start = i
				snippet = content[start:end]
				break
			}
		}
	}
	
	if end < len(content) && !isWordBoundary(content[end-1]) {
		// Find the next word boundary
		for i := end; i < len(content); i++ {
			if isWordBoundary(content[i]) {
				end = i
				snippet = content[start:end]
				break
			}
		}
	}
	
	return prefix + strings.TrimSpace(snippet) + suffix
}

// isWordBoundary checks if a character is a word boundary
func isWordBoundary(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t' || c == '\r' ||
		c == '.' || c == '!' || c == '?' || c == ',' || c == ';' || c == ':'
}
