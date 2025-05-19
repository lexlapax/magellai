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
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

func init() {
	storage.RegisterBackend(storage.FileSystemBackend, New)
}

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
func (b *Backend) NewSession(name string) *domain.Session {
	sessionID := generateSessionID()

	logging.LogInfo("Creating new session", "id", sessionID, "name", name)

	session := domain.NewSession(sessionID)
	session.Name = name

	return session
}

// SaveSession persists a session to disk
func (b *Backend) SaveSession(session *domain.Session) error {
	start := time.Now()
	session.UpdateTimestamp()

	logging.LogInfo("Saving session", "id", session.ID, "name", session.Name)

	filename := fmt.Sprintf("%s.json", session.ID)
	filepath := filepath.Join(b.baseDir, filename)

	// Use domain type directly - no conversion needed
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
func (b *Backend) LoadSession(id string) (*domain.Session, error) {
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

	// Use domain type directly - no conversion needed
	var session domain.Session
	if err := json.Unmarshal(data, &session); err != nil {
		logging.LogError(err, "Failed to unmarshal session", "id", id)
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	logging.LogInfo("Session loaded successfully", "id", id, "duration", time.Since(start))
	return &session, nil
}

// ListSessions returns a list of all available sessions
func (b *Backend) ListSessions() ([]*domain.SessionInfo, error) {
	logging.LogDebug("Listing sessions", "baseDir", b.baseDir)

	entries, err := os.ReadDir(b.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var sessions []*domain.SessionInfo

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

		// Use domain type directly
		var session domain.Session
		if err := json.Unmarshal(data, &session); err != nil {
			logging.LogDebug("Failed to unmarshal session", "path", filepath, "error", err)
			continue
		}
		sessions = append(sessions, session.ToSessionInfo())
	}

	logging.LogInfo("Listed sessions", "count", len(sessions))
	return sessions, nil
}

// DeleteSession removes a session from disk
func (b *Backend) DeleteSession(id string) error {
	logging.LogInfo("Deleting session", "id", id)

	filename := fmt.Sprintf("%s.json", id)
	filepath := filepath.Join(b.baseDir, filename)

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			logging.LogWarn("Session not found for deletion", "id", id)
			return fmt.Errorf("session not found: %s", id)
		}
		logging.LogError(err, "Failed to delete session", "id", id)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	logging.LogInfo("Session deleted", "id", id)
	return nil
}

// SearchSessions searches for sessions containing the given query
func (b *Backend) SearchSessions(query string) ([]*domain.SearchResult, error) {
	logging.LogInfo("Searching sessions", "query", query)
	lowerQuery := strings.ToLower(query)

	entries, err := os.ReadDir(b.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var results []*domain.SearchResult

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filepath := filepath.Join(b.baseDir, entry.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			continue
		}

		var session domain.Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		// Create session info for searching
		sessionInfo := session.ToSessionInfo()
		result := domain.NewSearchResult(sessionInfo)

		// Search in session name
		if strings.Contains(strings.ToLower(session.Name), lowerQuery) {
			result.AddMatch(domain.NewSearchMatch(
				domain.SearchMatchTypeName,
				"",
				session.Name,
				session.Name,
				-1,
			))
		}

		// Search in messages
		if session.Conversation != nil {
			for i, msg := range session.Conversation.Messages {
				if strings.Contains(strings.ToLower(msg.Content), lowerQuery) {
					snippet := extractSnippet(msg.Content, lowerQuery, 50)
					result.AddMatch(domain.NewSearchMatch(
						domain.SearchMatchTypeMessage,
						string(msg.Role),
						msg.Content,
						snippet,
						i,
					))
				}
			}

			// Search in system prompt
			if strings.Contains(strings.ToLower(session.Conversation.SystemPrompt), lowerQuery) {
				snippet := extractSnippet(session.Conversation.SystemPrompt, lowerQuery, 50)
				result.AddMatch(domain.NewSearchMatch(
					domain.SearchMatchTypeSystemPrompt,
					"",
					session.Conversation.SystemPrompt,
					snippet,
					-1,
				))
			}
		}

		// Search in tags
		for _, tag := range session.Tags {
			if strings.Contains(strings.ToLower(tag), lowerQuery) {
				result.AddMatch(domain.NewSearchMatch(
					domain.SearchMatchTypeTag,
					"",
					tag,
					tag,
					-1,
				))
			}
		}

		if result.HasMatches() {
			results = append(results, result)
		}
	}

	logging.LogInfo("Search completed", "query", query, "results", len(results))
	return results, nil
}

// ExportSession exports a session in the specified format
func (b *Backend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	logging.LogInfo("Exporting session", "id", id, "format", format)

	session, err := b.LoadSession(id)
	if err != nil {
		return err
	}

	switch format {
	case domain.ExportFormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(session); err != nil {
			return fmt.Errorf("failed to encode session as JSON: %w", err)
		}

	case domain.ExportFormatMarkdown:
		if err := exportMarkdown(session, w); err != nil {
			return fmt.Errorf("failed to export session as Markdown: %w", err)
		}

	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	logging.LogInfo("Session exported", "id", id, "format", format)
	return nil
}

// Close cleans up resources (no-op for filesystem backend)
func (b *Backend) Close() error {
	logging.LogDebug("Closing filesystem backend")
	return nil
}

// Helper functions

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func extractSnippet(content, searchTerm string, contextLen int) string {
	idx := strings.Index(strings.ToLower(content), strings.ToLower(searchTerm))
	if idx == -1 {
		return ""
	}

	// Calculate start and end positions for context
	start := idx - contextLen
	if start < 0 {
		start = 0
	}

	end := idx + len(searchTerm) + contextLen
	if end > len(content) {
		end = len(content)
	}

	// Adjust to word boundaries if possible
	if start > 0 {
		// Find start of word
		for start > 0 && content[start-1] != ' ' {
			start--
		}
	}

	// Adjust end to include full word
	if end < len(content) {
		// If we're not at a space, extend to include the current word
		for end < len(content) && content[end] != ' ' {
			end++
		}
		// Now we're at a space, include one more word
		if end < len(content) && content[end] == ' ' {
			end++ // Skip the space
			// Include the next word
			for end < len(content) && content[end] != ' ' {
				end++
			}
		}
	}

	// Extract the snippet
	snippet := content[start:end]

	// Add ellipsis if we didn't start at the beginning
	if start > 0 {
		snippet = "..." + snippet
	}

	// Add ellipsis if we didn't reach the end
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}

func exportMarkdown(session *domain.Session, w io.Writer) error {
	fmt.Fprintf(w, "# Session: %s\n\n", session.Name)
	fmt.Fprintf(w, "**ID:** %s\n", session.ID)
	fmt.Fprintf(w, "**Created:** %s\n", session.Created.Format(time.RFC3339))
	fmt.Fprintf(w, "**Updated:** %s\n", session.Updated.Format(time.RFC3339))

	// Add tags if present
	if len(session.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n", strings.Join(session.Tags, ", "))
	}
	fmt.Fprintf(w, "\n")

	if session.Conversation != nil {
		if session.Conversation.SystemPrompt != "" {
			fmt.Fprintf(w, "## System Prompt\n\n%s\n\n", session.Conversation.SystemPrompt)
		}

		fmt.Fprintf(w, "## Conversation\n\n")
		for _, msg := range session.Conversation.Messages {
			role := string(msg.Role)
			if len(role) > 0 {
				role = strings.ToUpper(role[:1]) + role[1:]
			}
			fmt.Fprintf(w, "### %s\n\n", role)
			fmt.Fprintf(w, "%s\n\n", msg.Content)

			if len(msg.Attachments) > 0 {
				fmt.Fprintf(w, "**Attachments:**\n")
				for _, att := range msg.Attachments {
					name := att.GetDisplayName()
					if name == "" {
						name = string(att.Type) + "_attachment"
					}
					if att.MimeType != "" {
						fmt.Fprintf(w, "- %s (%s)\n", name, att.MimeType)
					} else {
						fmt.Fprintf(w, "- %s (%s)\n", name, att.Type)
					}
				}
				fmt.Fprintf(w, "\n")
			}
		}
	}

	return nil
}

// GetChildren returns all direct child branches of a session
func (b *Backend) GetChildren(sessionID string) ([]*domain.SessionInfo, error) {
	logging.LogDebug("Getting children for session", "sessionID", sessionID)

	// Load the parent session to get child IDs
	parent, err := b.LoadSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent session: %w", err)
	}

	children := make([]*domain.SessionInfo, 0, len(parent.ChildIDs))

	// Load each child session info
	for _, childID := range parent.ChildIDs {
		child, err := b.LoadSession(childID)
		if err != nil {
			logging.LogDebug("Failed to load child session", "childID", childID, "error", err)
			continue // Skip missing children
		}
		children = append(children, child.ToSessionInfo())
	}

	return children, nil
}

// GetBranchTree returns the full branch tree starting from a session
func (b *Backend) GetBranchTree(sessionID string) (*domain.BranchTree, error) {
	logging.LogDebug("Getting branch tree for session", "sessionID", sessionID)

	// Load the root session
	session, err := b.LoadSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Create the tree node
	tree := &domain.BranchTree{
		Session:  session.ToSessionInfo(),
		Children: make([]*domain.BranchTree, 0),
	}

	// Recursively build the tree
	if len(session.ChildIDs) > 0 {
		for _, childID := range session.ChildIDs {
			childTree, err := b.GetBranchTree(childID)
			if err != nil {
				logging.LogDebug("Failed to get child tree", "childID", childID, "error", err)
				continue // Skip missing children
			}
			tree.Children = append(tree.Children, childTree)
		}
	}

	return tree, nil
}

// MergeSessions merges two sessions according to the specified options
func (b *Backend) MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error) {
	logging.LogInfo("Merging sessions", "target", targetID, "source", sourceID, "type", options.Type)

	// Load both sessions
	targetSession, err := b.LoadSession(targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load target session: %w", err)
	}

	sourceSession, err := b.LoadSession(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source session: %w", err)
	}

	// Update the merge options with actual IDs
	options.TargetID = targetID
	options.SourceID = sourceID

	// Execute the merge
	mergedSession, result, err := targetSession.ExecuteMerge(sourceSession, options)
	if err != nil {
		return nil, fmt.Errorf("failed to execute merge: %w", err)
	}

	// Save the merged session
	if err := b.SaveSession(mergedSession); err != nil {
		return nil, fmt.Errorf("failed to save merged session: %w", err)
	}

	// If the merge created a new branch, update the parent
	if result.NewBranchID != "" && options.CreateBranch {
		// Save the updated parent that now has the new child
		if err := b.SaveSession(targetSession); err != nil {
			logging.LogWarn("Failed to update parent session with new child", "parentID", targetID, "error", err)
		}
	}

	logging.LogInfo("Sessions merged successfully", "target", targetID, "source", sourceID, "mergedCount", result.MergedCount)
	return result, nil
}
