// ABOUTME: SQLite implementation of the storage backend interface
// ABOUTME: Stores sessions in a SQLite database with FTS5 search support

//go:build sqlite || db

package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

// Backend implements the storage.Backend interface using SQLite
type Backend struct {
	db     *sql.DB
	userID string
}

// New creates a new SQLite storage backend
func New(config storage.Config) (storage.Backend, error) {
	dbPath, ok := config["db_path"].(string)
	if !ok || dbPath == "" {
		// Default to user's data directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".config", "magellai", "sessions.db")
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get current user ID for multi-tenant support
	currentUser, err := user.Current()
	userID := "default"
	if err == nil {
		userID = currentUser.Username
	}

	backend := &Backend{
		db:     db,
		userID: userID,
	}

	// Initialize database schema
	if err := backend.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return backend, nil
}

// init registers the SQLite backend with the storage factory
func init() {
	storage.RegisterBackend(storage.SQLiteBackend, New)
}

// initSchema creates the necessary database tables
func (b *Backend) initSchema() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT,
			config TEXT,
			created TIMESTAMP,
			updated TIMESTAMP,
			metadata TEXT,
			conversation_id TEXT,
			tags TEXT,
			UNIQUE(user_id, id)
		)`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			model TEXT,
			provider TEXT,
			temperature REAL,
			max_tokens INTEGER,
			system_prompt TEXT,
			created TIMESTAMP,
			updated TIMESTAMP,
			metadata TEXT,
			UNIQUE(user_id, id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT,
			timestamp TIMESTAMP,
			attachments TEXT,
			metadata TEXT,
			position INTEGER,
			FOREIGN KEY (conversation_id, user_id) REFERENCES conversations(id, user_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			session_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			tag TEXT NOT NULL,
			PRIMARY KEY (session_id, user_id, tag),
			FOREIGN KEY (session_id, user_id) REFERENCES sessions(id, user_id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_user ON conversations(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tags_session ON tags(session_id, user_id)`,
	}

	for _, schema := range schemas {
		if _, err := b.db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	// Try to create FTS5 virtual table for search
	b.createFTSTable()

	return nil
}

// createFTSTable attempts to create FTS5 virtual table
func (b *Backend) createFTSTable() {
	fts5Schema := `CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
		conversation_id UNINDEXED,
		user_id UNINDEXED,
		content,
		role,
		tokenize='porter'
	)`

	if _, err := b.db.Exec(fts5Schema); err != nil {
		logging.LogDebug("FTS5 table creation failed, falling back to LIKE queries", "error", err)
	}
}

// NewSession creates a new session
func (b *Backend) NewSession(name string) *domain.Session {
	sessionID := storage.GenerateSessionID()
	session := domain.NewSession(sessionID)
	session.Name = name
	return session
}

// SaveSession persists a session to the database
func (b *Backend) SaveSession(session *domain.Session) error {
	tx, err := b.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Marshal JSON fields
	configJSON, _ := json.Marshal(session.Config)
	metadataJSON, _ := json.Marshal(session.Metadata)

	// Insert or update session
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO sessions 
		(id, user_id, name, config, created, updated, metadata, conversation_id, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, b.userID, session.Name, string(configJSON),
		session.Created, session.Updated, string(metadataJSON),
		session.Conversation.ID, strings.Join(session.Tags, ","),
	)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Save conversation
	convMetadataJSON, _ := json.Marshal(session.Conversation.Metadata)
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO conversations
		(id, user_id, model, provider, temperature, max_tokens, system_prompt, created, updated, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.Conversation.ID, b.userID, session.Conversation.Model,
		session.Conversation.Provider, session.Conversation.Temperature,
		session.Conversation.MaxTokens, session.Conversation.SystemPrompt,
		session.Conversation.Created, session.Conversation.Updated,
		string(convMetadataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to save conversation: %w", err)
	}

	// Delete existing messages and tags
	if _, err := tx.Exec("DELETE FROM messages WHERE conversation_id = ? AND user_id = ?", session.Conversation.ID, b.userID); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM tags WHERE session_id = ? AND user_id = ?", session.ID, b.userID); err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// Insert messages
	for idx, msg := range session.Conversation.Messages {
		attachmentsJSON, _ := json.Marshal(msg.Attachments)
		metadataJSON, _ := json.Marshal(msg.Metadata)

		_, err = tx.Exec(`
			INSERT INTO messages 
			(id, conversation_id, user_id, role, content, timestamp, attachments, metadata, position)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			msg.ID, session.Conversation.ID, b.userID, string(msg.Role), msg.Content,
			msg.Timestamp, string(attachmentsJSON), string(metadataJSON), idx,
		)
		if err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}

		// Update FTS table if it exists
		tx.Exec(`
			INSERT INTO messages_fts (conversation_id, user_id, content)
			VALUES (?, ?, ?)`,
			session.Conversation.ID, b.userID, msg.Content,
		)
	}

	// Insert tags
	for _, tag := range session.Tags {
		_, err = tx.Exec(`
			INSERT INTO tags (session_id, user_id, tag)
			VALUES (?, ?, ?)`,
			session.ID, b.userID, tag,
		)
		if err != nil {
			return fmt.Errorf("failed to save tag: %w", err)
		}
	}

	return tx.Commit()
}

// LoadSession loads a session from the database
func (b *Backend) LoadSession(id string) (*domain.Session, error) {
	var session domain.Session
	var configJSON, metadataJSON sql.NullString
	var conversationID string
	var tagsStr string

	row := b.db.QueryRow(`
		SELECT id, name, config, created, updated, metadata, conversation_id, tags
		FROM sessions 
		WHERE id = ? AND user_id = ?`,
		id, b.userID,
	)

	err := row.Scan(
		&session.ID, &session.Name, &configJSON, &session.Created,
		&session.Updated, &metadataJSON, &conversationID, &tagsStr,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: %s", storage.ErrSessionNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Unmarshal JSON fields
	if configJSON.Valid {
		json.Unmarshal([]byte(configJSON.String), &session.Config)
	} else {
		session.Config = make(map[string]interface{})
	}
	if metadataJSON.Valid {
		json.Unmarshal([]byte(metadataJSON.String), &session.Metadata)
	} else {
		session.Metadata = make(map[string]interface{})
	}

	// Parse tags
	if tagsStr != "" {
		session.Tags = strings.Split(tagsStr, ",")
	} else {
		session.Tags = []string{}
	}

	// Load conversation
	var conv domain.Conversation
	var convMetadataJSON sql.NullString
	var temperature sql.NullFloat64
	var maxTokens sql.NullInt64
	var systemPrompt sql.NullString

	row = b.db.QueryRow(`
		SELECT id, model, provider, temperature, max_tokens, system_prompt, 
		       created, updated, metadata
		FROM conversations 
		WHERE id = ? AND user_id = ?`,
		conversationID, b.userID,
	)

	err = row.Scan(
		&conv.ID, &conv.Model, &conv.Provider, &temperature, &maxTokens,
		&systemPrompt, &conv.Created, &conv.Updated, &convMetadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}

	if temperature.Valid {
		conv.Temperature = temperature.Float64
	}
	if maxTokens.Valid {
		conv.MaxTokens = int(maxTokens.Int64)
	}
	if systemPrompt.Valid {
		conv.SystemPrompt = systemPrompt.String
	}
	if convMetadataJSON.Valid {
		json.Unmarshal([]byte(convMetadataJSON.String), &conv.Metadata)
	} else {
		conv.Metadata = make(map[string]interface{})
	}

	// Load messages
	rows, err := b.db.Query(`
		SELECT id, role, content, timestamp, attachments, metadata
		FROM messages
		WHERE conversation_id = ? AND user_id = ?
		ORDER BY position`,
		conversationID, b.userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	defer rows.Close()

	conv.Messages = []domain.Message{}
	for rows.Next() {
		var msg domain.Message
		var roleStr string
		var attachmentsJSON, msgMetadataJSON sql.NullString

		err := rows.Scan(
			&msg.ID, &roleStr, &msg.Content, &msg.Timestamp,
			&attachmentsJSON, &msgMetadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Role = domain.MessageRole(roleStr)

		if attachmentsJSON.Valid {
			json.Unmarshal([]byte(attachmentsJSON.String), &msg.Attachments)
		} else {
			msg.Attachments = []domain.Attachment{}
		}
		if msgMetadataJSON.Valid {
			json.Unmarshal([]byte(msgMetadataJSON.String), &msg.Metadata)
		} else {
			msg.Metadata = make(map[string]interface{})
		}

		conv.Messages = append(conv.Messages, msg)
	}

	session.Conversation = &conv

	return &session, nil
}

// ListSessions returns a list of all sessions for the current user
func (b *Backend) ListSessions() ([]*domain.SessionInfo, error) {
	rows, err := b.db.Query(`
		SELECT s.id, s.name, s.created, s.updated, s.tags,
		       c.model, c.provider,
		       COUNT(m.id) as message_count
		FROM sessions s
		JOIN conversations c ON s.conversation_id = c.id AND s.user_id = c.user_id
		LEFT JOIN messages m ON c.id = m.conversation_id AND c.user_id = m.user_id
		WHERE s.user_id = ?
		GROUP BY s.id
		ORDER BY s.updated DESC`,
		b.userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.SessionInfo
	for rows.Next() {
		var info domain.SessionInfo
		var tagsStr string

		err := rows.Scan(
			&info.ID, &info.Name, &info.Created, &info.Updated, &tagsStr,
			&info.Model, &info.Provider, &info.MessageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session info: %w", err)
		}

		// Parse tags
		if tagsStr != "" {
			info.Tags = strings.Split(tagsStr, ",")
		} else {
			info.Tags = []string{}
		}

		sessions = append(sessions, &info)
	}

	return sessions, nil
}

// DeleteSession removes a session from the database
func (b *Backend) DeleteSession(id string) error {
	tx, err := b.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get conversation ID first
	var conversationID string
	err = tx.QueryRow("SELECT conversation_id FROM sessions WHERE id = ? AND user_id = ?", id, b.userID).Scan(&conversationID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("%w: %s", storage.ErrSessionNotFound, id)
	}
	if err != nil {
		return fmt.Errorf("failed to get conversation ID: %w", err)
	}

	// Delete session (cascades to tags)
	result, err := tx.Exec("DELETE FROM sessions WHERE id = ? AND user_id = ?", id, b.userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deletion result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("%w: %s", storage.ErrSessionNotFound, id)
	}

	// Delete conversation (cascades to messages)
	_, err = tx.Exec("DELETE FROM conversations WHERE id = ? AND user_id = ?", conversationID, b.userID)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return tx.Commit()
}

// SearchSessions searches for sessions by text content
func (b *Backend) SearchSessions(query string) ([]*domain.SearchResult, error) {
	// For now, don't use FTS5 - simplify for testing
	ftsAvailable := false

	sessions, err := b.ListSessions()
	if err != nil {
		return nil, err
	}

	var results []*domain.SearchResult
	lowerQuery := strings.ToLower(query)

	for _, info := range sessions {
		session, err := b.LoadSession(info.ID)
		if err != nil {
			continue
		}

		result := domain.NewSearchResult(info)

		// Search in system prompt
		if session.Conversation.SystemPrompt != "" && strings.Contains(strings.ToLower(session.Conversation.SystemPrompt), lowerQuery) {
			result.AddMatch(domain.SearchMatch{
				Type:    domain.SearchMatchTypeSystemPrompt,
				Content: extractSnippet(session.Conversation.SystemPrompt, query, 50),
				Context: "System Prompt",
			})
		}

		// Search in messages using FTS5 or LIKE
		if ftsAvailable {
			rows, err := b.db.Query(`
				SELECT content, role, position 
				FROM messages m
				JOIN messages_fts ON m.conversation_id = messages_fts.conversation_id 
				    AND m.user_id = messages_fts.user_id
				WHERE messages_fts MATCH ? 
				    AND m.conversation_id = ? 
				    AND m.user_id = ?`,
				query, session.Conversation.ID, b.userID,
			)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var content, role string
					var position int
					if err := rows.Scan(&content, &role, &position); err == nil {
						result.AddMatch(domain.SearchMatch{
							Type:     domain.SearchMatchTypeMessage,
							Role:     role,
							Content:  extractSnippet(content, query, 50),
							Context:  fmt.Sprintf("Message %d (%s)", position+1, role),
							Position: position,
						})
					}
				}
			}
		} else {
			// Fallback to searching in loaded messages
			for idx, msg := range session.Conversation.Messages {
				if strings.Contains(strings.ToLower(msg.Content), lowerQuery) {
					result.AddMatch(domain.SearchMatch{
						Type:     domain.SearchMatchTypeMessage,
						Role:     string(msg.Role),
						Content:  extractSnippet(msg.Content, query, 50),
						Context:  fmt.Sprintf("Message %d (%s)", idx+1, msg.Role),
						Position: idx,
					})
				}
			}
		}

		// Search in session name
		if strings.Contains(strings.ToLower(session.Name), lowerQuery) {
			result.AddMatch(domain.SearchMatch{
				Type:    domain.SearchMatchTypeName,
				Content: session.Name,
				Context: "Session Name",
			})
		}

		// Search in tags
		for _, tag := range session.Tags {
			if strings.Contains(strings.ToLower(tag), lowerQuery) {
				result.AddMatch(domain.SearchMatch{
					Type:    domain.SearchMatchTypeTag,
					Content: tag,
					Context: "Tag",
				})
			}
		}

		if result.HasMatches() {
			results = append(results, result)
		}
	}

	return results, nil
}

// ExportSession exports a session in the specified format
func (b *Backend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	session, err := b.LoadSession(id)
	if err != nil {
		return err
	}

	switch format {
	case domain.ExportFormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(session)

	case domain.ExportFormatMarkdown:
		return exportMarkdown(session, w)

	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// Close closes the database connection
func (b *Backend) Close() error {
	return b.db.Close()
}

// Helper functions

func extractSnippet(content, query string, contextRadius int) string {
	// Similar to filesystem backend but without word boundary adjustment
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

	return prefix + strings.TrimSpace(content[start:end]) + suffix
}

func exportMarkdown(session *domain.Session, w io.Writer) error {
	fmt.Fprintf(w, "# Session: %s\n\n", session.Name)
	fmt.Fprintf(w, "ID: %s\n", session.ID)
	fmt.Fprintf(w, "Created: %s\n", session.Created.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated: %s\n\n", session.Updated.Format(time.RFC3339))

	if len(session.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n\n", strings.Join(session.Tags, ", "))
	}

	fmt.Fprintln(w, "## Conversation")

	for _, msg := range session.Conversation.Messages {
		// Capitalize first letter of role
		role := string(msg.Role)
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
					name = string(att.Type) + "_attachment"
				}
				fmt.Fprintf(w, "- %s (%s)\n", name, att.MimeType)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}

// GetChildren returns all direct child branches of a session
func (b *Backend) GetChildren(sessionID string) ([]*domain.SessionInfo, error) {
	parent, err := b.LoadSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent session: %w", err)
	}

	children := make([]*domain.SessionInfo, 0, len(parent.ChildIDs))
	for _, childID := range parent.ChildIDs {
		child, err := b.LoadSession(childID)
		if err != nil {
			// Skip missing children
			continue
		}
		children = append(children, child.ToSessionInfo())
	}

	return children, nil
}

// GetBranchTree returns the full branch tree starting from a session
func (b *Backend) GetBranchTree(sessionID string) (*domain.BranchTree, error) {
	session, err := b.LoadSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	tree := &domain.BranchTree{
		Session:  session.ToSessionInfo(),
		Children: make([]*domain.BranchTree, 0),
	}

	// Recursively build the tree
	for _, childID := range session.ChildIDs {
		childTree, err := b.GetBranchTree(childID)
		if err != nil {
			// Skip missing children
			continue
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

// MergeSessions merges two sessions according to the specified options
func (b *Backend) MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error) {
	logging.LogInfo("Merging sessions", "target", targetID, "source", sourceID, "type", options.Type)

	// Begin transaction
	tx, err := b.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Load both sessions
	targetSession, err := b.LoadSession(targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load target session: %w", err)
	}

	sourceSession, err := b.LoadSession(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source session: %w", err)
	}

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

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logging.LogInfo("Sessions merged successfully", "target", targetID, "source", sourceID, "mergedCount", result.MergedCount)
	return result, nil
}
