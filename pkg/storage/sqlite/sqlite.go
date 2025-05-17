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
			model TEXT,
			provider TEXT,
			temperature REAL,
			max_tokens INTEGER,
			system_prompt TEXT,
			UNIQUE(user_id, id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT,
			timestamp TIMESTAMP,
			attachments TEXT,
			metadata TEXT,
			position INTEGER,
			FOREIGN KEY (session_id, user_id) REFERENCES sessions(id, user_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			session_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			tag TEXT NOT NULL,
			PRIMARY KEY (session_id, user_id, tag),
			FOREIGN KEY (session_id, user_id) REFERENCES sessions(id, user_id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, user_id)`,
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
		session_id UNINDEXED,
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
func (b *Backend) NewSession(name string) *storage.Session {
	sessionID := storage.GenerateSessionID()
	now := time.Now()

	return &storage.Session{
		ID:       sessionID,
		Name:     name,
		Messages: []storage.Message{},
		Config:   make(map[string]interface{}),
		Created:  now,
		Updated:  now,
		Metadata: make(map[string]interface{}),
		Tags:     []string{},
	}
}

// SaveSession persists a session to the database
func (b *Backend) SaveSession(session *storage.Session) error {
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
		(id, user_id, name, config, created, updated, metadata, model, provider, temperature, max_tokens, system_prompt)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, b.userID, session.Name, string(configJSON),
		session.Created, session.Updated, string(metadataJSON),
		session.Model, session.Provider, session.Temperature,
		session.MaxTokens, session.SystemPrompt,
	)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Delete existing messages and tags
	if _, err := tx.Exec("DELETE FROM messages WHERE session_id = ? AND user_id = ?", session.ID, b.userID); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM tags WHERE session_id = ? AND user_id = ?", session.ID, b.userID); err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// Insert messages
	for idx, msg := range session.Messages {
		attachmentsJSON, _ := json.Marshal(msg.Attachments)
		metadataJSON, _ := json.Marshal(msg.Metadata)

		_, err = tx.Exec(`
			INSERT INTO messages 
			(id, session_id, user_id, role, content, timestamp, attachments, metadata, position)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			msg.ID, session.ID, b.userID, msg.Role, msg.Content,
			msg.Timestamp, string(attachmentsJSON), string(metadataJSON), idx,
		)
		if err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}

		// Update FTS table if it exists
		tx.Exec(`
			INSERT INTO messages_fts (session_id, user_id, content, role)
			VALUES (?, ?, ?, ?)`,
			session.ID, b.userID, msg.Content, msg.Role,
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
func (b *Backend) LoadSession(id string) (*storage.Session, error) {
	var session storage.Session
	var configJSON, metadataJSON sql.NullString
	var temperature sql.NullFloat64
	var maxTokens sql.NullInt64
	var systemPrompt sql.NullString

	row := b.db.QueryRow(`
		SELECT id, name, config, created, updated, metadata, model, provider, 
		       temperature, max_tokens, system_prompt
		FROM sessions 
		WHERE id = ? AND user_id = ?`,
		id, b.userID,
	)

	err := row.Scan(
		&session.ID, &session.Name, &configJSON, &session.Created,
		&session.Updated, &metadataJSON, &session.Model, &session.Provider,
		&temperature, &maxTokens, &systemPrompt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Unmarshal JSON fields
	if configJSON.Valid {
		json.Unmarshal([]byte(configJSON.String), &session.Config)
	}
	if metadataJSON.Valid {
		json.Unmarshal([]byte(metadataJSON.String), &session.Metadata)
	}
	if temperature.Valid {
		session.Temperature = temperature.Float64
	}
	if maxTokens.Valid {
		session.MaxTokens = int(maxTokens.Int64)
	}
	if systemPrompt.Valid {
		session.SystemPrompt = systemPrompt.String
	}

	// Load messages
	rows, err := b.db.Query(`
		SELECT id, role, content, timestamp, attachments, metadata
		FROM messages
		WHERE session_id = ? AND user_id = ?
		ORDER BY position`,
		id, b.userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	defer rows.Close()

	session.Messages = []storage.Message{}
	for rows.Next() {
		var msg storage.Message
		var attachmentsJSON, msgMetadataJSON sql.NullString

		err := rows.Scan(
			&msg.ID, &msg.Role, &msg.Content, &msg.Timestamp,
			&attachmentsJSON, &msgMetadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if attachmentsJSON.Valid {
			json.Unmarshal([]byte(attachmentsJSON.String), &msg.Attachments)
		}
		if msgMetadataJSON.Valid {
			json.Unmarshal([]byte(msgMetadataJSON.String), &msg.Metadata)
		}

		session.Messages = append(session.Messages, msg)
	}

	// Load tags
	rows, err = b.db.Query(`
		SELECT tag FROM tags WHERE session_id = ? AND user_id = ?`,
		id, b.userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	defer rows.Close()

	session.Tags = []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		session.Tags = append(session.Tags, tag)
	}

	return &session, nil
}

// ListSessions returns a list of all sessions for the current user
func (b *Backend) ListSessions() ([]*storage.SessionInfo, error) {
	rows, err := b.db.Query(`
		SELECT s.id, s.name, s.created, s.updated, s.model, s.provider,
		       COUNT(m.id) as message_count
		FROM sessions s
		LEFT JOIN messages m ON s.id = m.session_id AND s.user_id = m.user_id
		WHERE s.user_id = ?
		GROUP BY s.id
		ORDER BY s.updated DESC`,
		b.userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*storage.SessionInfo
	for rows.Next() {
		var info storage.SessionInfo
		err := rows.Scan(
			&info.ID, &info.Name, &info.Created, &info.Updated,
			&info.Model, &info.Provider, &info.MessageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session info: %w", err)
		}

		// Load tags
		tagRows, err := b.db.Query(`
			SELECT tag FROM tags WHERE session_id = ? AND user_id = ?`,
			info.ID, b.userID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags: %w", err)
		}

		info.Tags = []string{}
		for tagRows.Next() {
			var tag string
			if err := tagRows.Scan(&tag); err != nil {
				tagRows.Close()
				return nil, fmt.Errorf("failed to scan tag: %w", err)
			}
			info.Tags = append(info.Tags, tag)
		}
		tagRows.Close()

		sessions = append(sessions, &info)
	}

	return sessions, nil
}

// DeleteSession removes a session from the database
func (b *Backend) DeleteSession(id string) error {
	result, err := b.db.Exec("DELETE FROM sessions WHERE id = ? AND user_id = ?", id, b.userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deletion result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	return nil
}

// SearchSessions searches for sessions by text content
func (b *Backend) SearchSessions(query string) ([]*storage.SearchResult, error) {
	// First, check if FTS5 is available
	var ftsAvailable bool
	row := b.db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='messages_fts'")
	if err := row.Scan(&ftsAvailable); err == sql.ErrNoRows {
		ftsAvailable = false
	}

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

		// Search in messages using FTS5 or LIKE
		if ftsAvailable {
			rows, err := b.db.Query(`
				SELECT content, role, position 
				FROM messages m
				JOIN messages_fts ON m.session_id = messages_fts.session_id 
				    AND m.user_id = messages_fts.user_id
				WHERE messages_fts MATCH ? 
				    AND m.session_id = ? 
				    AND m.user_id = ?`,
				query, session.ID, b.userID,
			)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var content, role string
					var position int
					if err := rows.Scan(&content, &role, &position); err == nil {
						matches = append(matches, storage.SearchMatch{
							Type:     "message",
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

	return results, nil
}

// ExportSession exports a session in the specified format
func (b *Backend) ExportSession(id string, format storage.ExportFormat, w io.Writer) error {
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

func exportMarkdown(session *storage.Session, w io.Writer) error {
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
