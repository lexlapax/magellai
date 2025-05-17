// ABOUTME: Filesystem implementation of the storage backend interface
// ABOUTME: Stores sessions as JSON files on disk with configurable base directory

package filesystem

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/storage"
)

// Backend implements the storage.Backend interface using filesystem storage
type Backend struct {
	baseDir string
}

// New creates a new filesystem storage backend
func New(config storage.Config) (storage.Backend, error) {
	baseDir, ok := config["base_dir"].(string)
	if !ok || baseDir == "" {
		return nil, fmt.Errorf("filesystem backend requires 'base_dir' configuration")
	}

	logging.LogDebug("Creating filesystem backend", "baseDir", baseDir)

	// Ensure directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		logging.LogError(err, "Failed to create storage directory", "baseDir", baseDir)
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Backend{baseDir: baseDir}, nil
}

// init registers the filesystem backend with the storage factory
func init() {
	storage.RegisterBackend(storage.FileSystemBackend, New)
}

// NewSession creates a new session
func (b *Backend) NewSession(name string) *storage.Session {
	now := time.Now()
	sessionID := storage.GenerateSessionID()

	logging.LogInfo("Creating new session", "id", sessionID, "name", name)

	return &storage.Session{
		ID:       sessionID,
		Name:     name,
		Messages: []storage.Message{},
		Config:   make(map[string]interface{}),
		Created:  now,
		Updated:  now,
		Metadata: make(map[string]interface{}),
	}
}

// SaveSession persists a session to disk
func (b *Backend) SaveSession(session *storage.Session) error {
	start := time.Now()
	session.Updated = time.Now()

	logging.LogInfo("Saving session", "id", session.ID, "name", session.Name)

	filename := fmt.Sprintf("%s.json", session.ID)
	filepath := filepath.Join(b.baseDir, filename)

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		logging.LogError(err, "Failed to marshal session", "id", session.ID)
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		logging.LogError(err, "Failed to write session file", "path", filepath)
		return fmt.Errorf("failed to write session file: %w", err)
	}

	logging.LogInfo("Session saved successfully", "id", session.ID, "duration", time.Since(start))
	return nil
}

// LoadSession loads a session from disk by ID
func (b *Backend) LoadSession(id string) (*storage.Session, error) {
	start := time.Now()
	logging.LogInfo("Loading session", "id", id)

	filename := fmt.Sprintf("%s.json", id)
	filepath := filepath.Join(b.baseDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session storage.Session
	if err := json.Unmarshal(data, &session); err != nil {
		logging.LogError(err, "Failed to unmarshal session", "id", id)
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	logging.LogInfo("Session loaded successfully", "id", id, "duration", time.Since(start))
	return &session, nil
}

// ListSessions returns a list of all available sessions
func (b *Backend) ListSessions() ([]*storage.SessionInfo, error) {
	logging.LogDebug("Listing sessions", "baseDir", b.baseDir)

	entries, err := os.ReadDir(b.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var sessions []*storage.SessionInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filepath := filepath.Join(b.baseDir, entry.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			logging.LogDebug("Failed to read session file", "path", filepath, "error", err)
			continue
		}

		var session storage.Session
		if err := json.Unmarshal(data, &session); err != nil {
			logging.LogDebug("Failed to unmarshal session", "path", filepath, "error", err)
			continue
		}

		info := &storage.SessionInfo{
			ID:           session.ID,
			Name:         session.Name,
			Created:      session.Created,
			Updated:      session.Updated,
			MessageCount: len(session.Messages),
			Model:        session.Model,
			Provider:     session.Provider,
			Tags:         session.Tags,
		}

		sessions = append(sessions, info)
	}

	return sessions, nil
}

// DeleteSession removes a session from disk
func (b *Backend) DeleteSession(id string) error {
	logging.LogInfo("Deleting session", "id", id)

	filename := fmt.Sprintf("%s.json", id)
	filepath := filepath.Join(b.baseDir, filename)

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", id)
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	logging.LogInfo("Session deleted successfully", "id", id)
	return nil
}

// SearchSessions searches for sessions by text content
func (b *Backend) SearchSessions(query string) ([]*storage.SearchResult, error) {
	logging.LogInfo("Searching sessions", "query", query)

	sessions, err := b.ListSessions()
	if err != nil {
		return nil, err
	}

	var results []*storage.SearchResult
	lowerQuery := strings.ToLower(query)

	for _, info := range sessions {
		session, err := b.LoadSession(info.ID)
		if err != nil {
			continue
		}

		var matches []storage.SearchMatch

		// Search in system prompt
		if session.SystemPrompt != "" && strings.Contains(strings.ToLower(session.SystemPrompt), lowerQuery) {
			matches = append(matches, storage.SearchMatch{
				Type:    "system_prompt",
				Content: extractSnippet(session.SystemPrompt, query, 50),
				Context: "System Prompt",
			})
		}

		// Search in messages
		for idx, msg := range session.Messages {
			if strings.Contains(strings.ToLower(msg.Content), lowerQuery) {
				matches = append(matches, storage.SearchMatch{
					Type:     "message",
					Role:     msg.Role,
					Content:  extractSnippet(msg.Content, query, 50),
					Context:  fmt.Sprintf("Message %d (%s)", idx+1, msg.Role),
					Position: idx,
				})
			}
		}

		// Search in session name
		if strings.Contains(strings.ToLower(session.Name), lowerQuery) {
			matches = append(matches, storage.SearchMatch{
				Type:    "name",
				Content: session.Name,
				Context: "Session Name",
			})
		}

		// Search in tags
		for _, tag := range session.Tags {
			if strings.Contains(strings.ToLower(tag), lowerQuery) {
				matches = append(matches, storage.SearchMatch{
					Type:    "tag",
					Content: tag,
					Context: "Tag",
				})
			}
		}

		if len(matches) > 0 {
			results = append(results, &storage.SearchResult{
				Session: info,
				Matches: matches,
			})
		}
	}

	logging.LogInfo("Search completed", "query", query, "results", len(results))
	return results, nil
}

// ExportSession exports a session to a writer in the specified format
func (b *Backend) ExportSession(id string, format storage.ExportFormat, w io.Writer) error {
	logging.LogInfo("Exporting session", "id", id, "format", format)

	session, err := b.LoadSession(id)
	if err != nil {
		return err
	}

	switch format {
	case storage.ExportFormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(session)

	case storage.ExportFormatMarkdown:
		return b.exportMarkdown(session, w)

	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// Close cleans up resources (no-op for filesystem)
func (b *Backend) Close() error {
	return nil
}

// exportMarkdown exports a session as markdown
func (b *Backend) exportMarkdown(session *storage.Session, w io.Writer) error {
	fmt.Fprintf(w, "# Session: %s\n\n", session.Name)
	fmt.Fprintf(w, "ID: %s\n", session.ID)
	fmt.Fprintf(w, "Created: %s\n", session.Created.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated: %s\n\n", session.Updated.Format(time.RFC3339))

	if len(session.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n\n", strings.Join(session.Tags, ", "))
	}

	fmt.Fprintln(w, "## Conversation")

	for _, msg := range session.Messages {
		// Capitalize first letter of role
		role := msg.Role
		if len(role) > 0 {
			role = strings.ToUpper(role[:1]) + role[1:]
		}
		fmt.Fprintf(w, "### %s\n", role)
		fmt.Fprintf(w, "*%s*\n\n", msg.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(w, "%s\n\n", msg.Content)

		if len(msg.Attachments) > 0 {
			fmt.Fprintln(w, "Attachments:")
			for _, att := range msg.Attachments {
				name := att.Name
				if name == "" {
					name = att.Type + "_attachment"
				}
				fmt.Fprintf(w, "- %s (%s)\n", name, att.MimeType)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}

// extractSnippet extracts a snippet with context around the matched query
func extractSnippet(content, query string, contextRadius int) string {
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerContent, lowerQuery)
	if idx == -1 {
		if len(content) <= contextRadius*2 {
			return content
		}
		return content[:contextRadius*2] + "..."
	}

	start := idx - contextRadius
	end := idx + len(query) + contextRadius

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

	snippet := content[start:end]

	// Adjust to word boundaries
	if start > 0 && !isWordBoundary(content[start-1]) {
		for i := start; i > 0; i-- {
			if isWordBoundary(content[i-1]) {
				start = i
				snippet = content[start:end]
				break
			}
		}
	}

	if end < len(content) && !isWordBoundary(content[end-1]) {
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
