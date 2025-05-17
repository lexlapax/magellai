//go:build sqlite || db
// +build sqlite db

// ABOUTME: SQLite storage backend for session persistence
// ABOUTME: Provides database-based session storage with multi-tenant support

package repl

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os/user"
	"strings"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorageBackend struct {
	db     *sql.DB
	userID string
}

// NewSQLiteStorage creates a new SQLite storage backend
func NewSQLiteStorage(config map[string]interface{}) (*SQLiteStorageBackend, error) {
	// Extract database path from config
	dbPath, ok := config["path"].(string)
	if !ok || dbPath == "" {
		return nil, fmt.Errorf("database path not specified in configuration")
	}

	// Get current user for multi-tenant support
	currentUser, err := user.Current()
	userID := "default"
	if err == nil {
		userID = currentUser.Username
	}

	// Allow overriding user ID from config
	if configUserID, ok := config["user_id"].(string); ok && configUserID != "" {
		userID = configUserID
	}

	logging.LogInfo("Opening SQLite database", "path", dbPath, "user", userID)

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logging.LogError(err, "Failed to open SQLite database", "path", dbPath)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create storage instance
	storage := &SQLiteStorageBackend{
		db:     db,
		userID: userID,
	}

	// Initialize database schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the database tables if they don't exist
func (s *SQLiteStorageBackend) initSchema() error {
	// SQLite doesn't support INDEX in CREATE TABLE, so create them separately
	tables := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT,
			created TIMESTAMP NOT NULL,
			updated TIMESTAMP NOT NULL,
			metadata TEXT,
			conversation TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_created ON sessions(created)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_updated ON sessions(updated)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			attachments TEXT,
			timestamp TIMESTAMP NOT NULL,
			sequence INTEGER NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session_user ON messages(session_id, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_content ON messages(content)`, // Add simple index for content searching
	}

	for _, table := range tables {
		_, err := s.db.Exec(table)
		if err != nil {
			logging.LogError(err, "Failed to create database schema", "sql", table)
			return err
		}
	}

	// Try to create FTS5 table if available
	fts5Available := s.checkFTS5Available()
	if fts5Available {
		fts5Tables := []string{
			`CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
				session_id UNINDEXED,
				user_id UNINDEXED,
				role UNINDEXED,
				content,
				content=messages,
				content_rowid=id
			)`,
			`CREATE TRIGGER IF NOT EXISTS messages_ai AFTER INSERT ON messages BEGIN
				INSERT INTO messages_fts(session_id, user_id, role, content) 
				VALUES (new.session_id, new.user_id, new.role, new.content);
			END`,
			`CREATE TRIGGER IF NOT EXISTS messages_ad AFTER DELETE ON messages BEGIN
				DELETE FROM messages_fts WHERE rowid = old.id;
			END`,
			`CREATE TRIGGER IF NOT EXISTS messages_au AFTER UPDATE ON messages BEGIN
				DELETE FROM messages_fts WHERE rowid = old.id;
				INSERT INTO messages_fts(session_id, user_id, role, content) 
				VALUES (new.session_id, new.user_id, new.role, new.content);
			END`,
		}

		for _, table := range fts5Tables {
			_, err := s.db.Exec(table)
			if err != nil {
				logging.LogWarn("FTS5 table creation failed, falling back to simple search", "error", err)
				break
			}
		}
	} else {
		logging.LogDebug("FTS5 not available, using simple search")
	}

	logging.LogDebug("Database schema initialized successfully")
	return nil
}

// checkFTS5Available checks if FTS5 is available
func (s *SQLiteStorageBackend) checkFTS5Available() bool {
	var dummy string
	err := s.db.QueryRow("SELECT sqlite_compileoption_used('ENABLE_FTS5')").Scan(&dummy)
	return err == nil
}

// NewSession creates a new session
func (s *SQLiteStorageBackend) NewSession(name string) *Session {
	return &Session{
		ID:           generateSessionID(),
		Name:         name,
		Created:      time.Now(),
		Updated:      time.Now(),
		Conversation: NewConversation(generateSessionID()),
		Metadata:     make(map[string]interface{}),
	}
}

// SaveSession saves a session to the database
func (s *SQLiteStorageBackend) SaveSession(session *Session) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Serialize metadata
	metadataJSON, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert or update session
	_, err = tx.Exec(`
		INSERT INTO sessions (id, user_id, name, created, updated, metadata, conversation)
		VALUES (?, ?, ?, ?, ?, ?, '')
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			updated = excluded.updated,
			metadata = excluded.metadata
	`, session.ID, s.userID, session.Name, session.Created, session.Updated, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Delete existing messages for this session
	_, err = tx.Exec("DELETE FROM messages WHERE session_id = ? AND user_id = ?", session.ID, s.userID)
	if err != nil {
		return fmt.Errorf("failed to delete old messages: %w", err)
	}

	// Insert messages
	for i, msg := range session.Conversation.Messages {
		attachmentsJSON, err := json.Marshal(msg.Attachments)
		if err != nil {
			return fmt.Errorf("failed to marshal attachments: %w", err)
		}

		_, err = tx.Exec(`
			INSERT INTO messages (session_id, user_id, role, content, attachments, timestamp, sequence)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, session.ID, s.userID, msg.Role, msg.Content, string(attachmentsJSON), time.Now(), i)
		if err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}
	}

	return tx.Commit()
}

// LoadSession loads a session from the database
func (s *SQLiteStorageBackend) LoadSession(id string) (*Session, error) {
	// Load session metadata
	var session Session
	var metadataJSON string
	err := s.db.QueryRow(`
		SELECT id, name, created, updated, metadata
		FROM sessions
		WHERE id = ? AND user_id = ?
	`, id, s.userID).Scan(&session.ID, &session.Name, &session.Created, &session.Updated, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal([]byte(metadataJSON), &session.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	session.Conversation = NewConversation(session.ID)

	// Load messages
	rows, err := s.db.Query(`
		SELECT role, content, attachments
		FROM messages
		WHERE session_id = ? AND user_id = ?
		ORDER BY sequence
	`, id, s.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		var attachmentsJSON string
		if err := rows.Scan(&msg.Role, &msg.Content, &attachmentsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Unmarshal attachments if present
		if attachmentsJSON != "" && attachmentsJSON != "null" {
			if err := json.Unmarshal([]byte(attachmentsJSON), &msg.Attachments); err != nil {
				return nil, fmt.Errorf("failed to unmarshal attachments: %w", err)
			}
		}

		session.Conversation.Messages = append(session.Conversation.Messages, msg)
	}

	return &session, nil
}

// ListSessions returns a list of all sessions for the current user
func (s *SQLiteStorageBackend) ListSessions() ([]*SessionInfo, error) {
	rows, err := s.db.Query(`
		SELECT id, name, created, updated,
			(SELECT COUNT(*) FROM messages WHERE session_id = sessions.id AND user_id = sessions.user_id) as message_count
		FROM sessions
		WHERE user_id = ?
		ORDER BY updated DESC
	`, s.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*SessionInfo
	for rows.Next() {
		var info SessionInfo
		if err := rows.Scan(&info.ID, &info.Name, &info.Created, &info.Updated, &info.MessageCount); err != nil {
			return nil, fmt.Errorf("failed to scan session info: %w", err)
		}
		sessions = append(sessions, &info)
	}

	return sessions, nil
}

// DeleteSession deletes a session from the database
func (s *SQLiteStorageBackend) DeleteSession(id string) error {
	result, err := s.db.Exec("DELETE FROM sessions WHERE id = ? AND user_id = ?", id, s.userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// SearchSessions searches sessions by content using FTS or LIKE
func (s *SQLiteStorageBackend) SearchSessions(query string) ([]*SearchResult, error) {
	if s.hasFTS5() {
		return s.searchWithFTS5(query)
	}
	return s.searchWithLike(query)
}

// hasFTS5 checks if FTS5 table exists
func (s *SQLiteStorageBackend) hasFTS5() bool {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='messages_fts'
	`).Scan(&count)
	return err == nil && count > 0
}

// searchWithFTS5 searches using FTS5
func (s *SQLiteStorageBackend) searchWithFTS5(query string) ([]*SearchResult, error) {
	// Escape special FTS characters
	escapedQuery := strings.ReplaceAll(query, `"`, `""`)

	rows, err := s.db.Query(`
		SELECT DISTINCT
			m.session_id,
			s.name,
			s.updated,
			m.content,
			highlight(messages_fts, 3, '<match>', '</match>') as highlighted_content
		FROM messages_fts m
		JOIN sessions s ON s.id = m.session_id
		WHERE m.user_id = ? AND messages_fts MATCH ?
		ORDER BY rank
		LIMIT 100
	`, s.userID, escapedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search sessions: %w", err)
	}
	defer rows.Close()

	resultMap := make(map[string]*SearchResult)
	for rows.Next() {
		var sessionID, sessionName, content, highlightedContent string
		var updated time.Time
		if err := rows.Scan(&sessionID, &sessionName, &updated, &content, &highlightedContent); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		if result, exists := resultMap[sessionID]; exists {
			// Add to existing matches
			match := SearchMatch{
				Type:    "message",
				Content: highlightedContent,
				Context: content,
			}
			result.Matches = append(result.Matches, match)
		} else {
			// Create new result
			sessionInfo := &SessionInfo{
				ID:      sessionID,
				Name:    sessionName,
				Updated: updated,
			}
			match := SearchMatch{
				Type:    "message",
				Content: highlightedContent,
				Context: content,
			}
			resultMap[sessionID] = &SearchResult{
				Session: sessionInfo,
				Matches: []SearchMatch{match},
			}
		}
	}

	// Convert map to slice
	var results []*SearchResult
	for _, result := range resultMap {
		results = append(results, result)
	}

	return results, nil
}

// searchWithLike searches using LIKE operator
func (s *SQLiteStorageBackend) searchWithLike(query string) ([]*SearchResult, error) {
	likePattern := "%" + query + "%"

	rows, err := s.db.Query(`
		SELECT DISTINCT
			m.session_id,
			s.name,
			s.updated,
			m.content
		FROM messages m
		JOIN sessions s ON s.id = m.session_id
		WHERE m.user_id = ? AND m.content LIKE ?
		ORDER BY m.timestamp DESC
		LIMIT 100
	`, s.userID, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search sessions: %w", err)
	}
	defer rows.Close()

	resultMap := make(map[string]*SearchResult)
	for rows.Next() {
		var sessionID, sessionName, content string
		var updated time.Time
		if err := rows.Scan(&sessionID, &sessionName, &updated, &content); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// Extract snippet around the match
		snippet := extractSnippet(content, query, 100)

		if result, exists := resultMap[sessionID]; exists {
			// Add to existing matches
			match := SearchMatch{
				Type:    "message",
				Content: snippet,
				Context: content,
			}
			result.Matches = append(result.Matches, match)
		} else {
			// Create new result
			sessionInfo := &SessionInfo{
				ID:      sessionID,
				Name:    sessionName,
				Updated: updated,
			}
			match := SearchMatch{
				Type:    "message",
				Content: snippet,
				Context: content,
			}
			resultMap[sessionID] = &SearchResult{
				Session: sessionInfo,
				Matches: []SearchMatch{match},
			}
		}
	}

	// Convert map to slice
	var results []*SearchResult
	for _, result := range resultMap {
		results = append(results, result)
	}

	return results, nil
}

// ExportSession exports a session in the specified format
func (s *SQLiteStorageBackend) ExportSession(id string, format string, w io.Writer) error {
	session, err := s.LoadSession(id)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(session)
	case "markdown":
		// Simple markdown export
		fmt.Fprintf(w, "# Session: %s\n\n", session.Name)
		fmt.Fprintf(w, "**ID:** %s  \n", session.ID)
		fmt.Fprintf(w, "**Created:** %s  \n", session.Created.Format(time.RFC3339))
		fmt.Fprintf(w, "**Updated:** %s  \n\n", session.Updated.Format(time.RFC3339))

		fmt.Fprintf(w, "## Messages\n\n")
		for i, msg := range session.Conversation.Messages {
			fmt.Fprintf(w, "### Message %d - %s\n\n", i+1, msg.Role)
			fmt.Fprintf(w, "%s\n\n", msg.Content)
			if len(msg.Attachments) > 0 {
				fmt.Fprintf(w, "**Attachments:**\n")
				for _, att := range msg.Attachments {
					fmt.Fprintf(w, "- Type: %s", att.Type)
					if att.FilePath != "" {
						fmt.Fprintf(w, ", Path: %s", att.FilePath)
					}
					fmt.Fprintf(w, "\n")
				}
				fmt.Fprintf(w, "\n")
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// Close closes the database connection
func (s *SQLiteStorageBackend) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Register the SQLite storage backend
func init() {
	RegisterStorageBackend(SQLiteStorage, func(config map[string]interface{}) (StorageBackend, error) {
		return NewSQLiteStorage(config)
	})
}
