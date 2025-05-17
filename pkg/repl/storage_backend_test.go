package repl

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFileSystemBackend tests the file system storage backend
func TestFileSystemBackend(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "magellai-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a filesystem backend
	storage, err := NewFileSystemBackend(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create filesystem backend: %v", err)
	}

	t.Run("NewSession", func(t *testing.T) {
		session := storage.NewSession("test session")
		if session == nil {
			t.Fatal("NewSession returned nil")
		}
		if session.Name != "test session" {
			t.Errorf("Expected session name 'test session', got '%s'", session.Name)
		}
		if session.ID == "" {
			t.Error("Session ID is empty")
		}
	})

	t.Run("SaveAndLoadSession", func(t *testing.T) {
		// Create a new session
		session := storage.NewSession("save-test")
		session.Config = map[string]interface{}{
			"model": "test-model",
		}

		// Save it
		err := storage.SaveSession(session)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Load it back
		loaded, err := storage.LoadSession(session.ID)
		if err != nil {
			t.Fatalf("Failed to load session: %v", err)
		}

		if loaded.ID != session.ID {
			t.Errorf("Expected ID %s, got %s", session.ID, loaded.ID)
		}
		if loaded.Name != session.Name {
			t.Errorf("Expected name %s, got %s", session.Name, loaded.Name)
		}
		if loaded.Config["model"] != "test-model" {
			t.Errorf("Expected model config to be preserved")
		}
	})

	t.Run("ListSessions", func(t *testing.T) {
		// Create multiple sessions
		session1 := storage.NewSession("list-test-1")
		err := storage.SaveSession(session1)
		if err != nil {
			t.Fatalf("Failed to save session1: %v", err)
		}

		session2 := storage.NewSession("list-test-2")
		err = storage.SaveSession(session2)
		if err != nil {
			t.Fatalf("Failed to save session2: %v", err)
		}

		// List sessions
		sessions, err := storage.ListSessions()
		if err != nil {
			t.Fatalf("Failed to list sessions: %v", err)
		}

		// Should have at least 2 sessions
		if len(sessions) < 2 {
			t.Errorf("Expected at least 2 sessions, got %d", len(sessions))
		}

		// Verify both sessions are present
		found1, found2 := false, false
		for _, info := range sessions {
			if info.ID == session1.ID {
				found1 = true
			}
			if info.ID == session2.ID {
				found2 = true
			}
		}

		if !found1 || !found2 {
			t.Error("Not all sessions were listed")
		}
	})

	t.Run("DeleteSession", func(t *testing.T) {
		// Create a session
		session := storage.NewSession("delete-test")
		err := storage.SaveSession(session)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Delete it
		err = storage.DeleteSession(session.ID)
		if err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}

		// Try to load it - should fail
		_, err = storage.LoadSession(session.ID)
		if err == nil {
			t.Error("Expected error loading deleted session")
		}
	})

	t.Run("SearchSessions", func(t *testing.T) {
		// Create sessions with specific content
		session := storage.NewSession("search-test")
		session.Conversation = &Conversation{
			Messages: []Message{
				{Role: "user", Content: "What is the capital of France?"},
				{Role: "assistant", Content: "The capital of France is Paris."},
			},
		}
		err := storage.SaveSession(session)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Search for content
		results, err := storage.SearchSessions("capital")
		if err != nil {
			t.Fatalf("Failed to search sessions: %v", err)
		}

		found := false
		for _, result := range results {
			if result.Session.ID == session.ID {
				found = true
				// Check that we have matches
				if len(result.Matches) == 0 {
					t.Error("Expected at least one match")
					continue
				}
				// Check the first match
				match := result.Matches[0]
				if match.Type != "message" {
					t.Errorf("Expected match type 'message', got '%s'", match.Type)
				}
				if !strings.Contains(match.Content, "capital") {
					t.Errorf("Expected match content to contain 'capital'")
				}
			}
		}

		if !found {
			t.Error("Session not found in search results")
		}
	})

	t.Run("ExportSession", func(t *testing.T) {
		// Create a session
		session := storage.NewSession("export-test")
		session.Conversation = &Conversation{
			Messages: []Message{
				{Role: "user", Content: "Test message"},
			},
		}
		err := storage.SaveSession(session)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Export as JSON
		var jsonBuf bytes.Buffer
		err = storage.ExportSession(session.ID, "json", &jsonBuf)
		if err != nil {
			t.Fatalf("Failed to export session as JSON: %v", err)
		}

		jsonOutput := jsonBuf.String()
		if !strings.Contains(jsonOutput, "export-test") {
			t.Error("JSON export doesn't contain session name")
		}

		// Export as Markdown
		var mdBuf bytes.Buffer
		err = storage.ExportSession(session.ID, "markdown", &mdBuf)
		if err != nil {
			t.Fatalf("Failed to export session as Markdown: %v", err)
		}

		mdOutput := mdBuf.String()
		if !strings.Contains(mdOutput, "export-test") {
			t.Error("Markdown export doesn't contain session name")
		}
		if !strings.Contains(mdOutput, "## User") {
			t.Error("Markdown export doesn't have proper formatting")
		}
	})
}

// TestStorageFactory tests the storage factory
func TestStorageFactory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "magellai-factory-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("CreateFileSystemBackend", func(t *testing.T) {
		config := map[string]interface{}{
			"base_dir": tmpDir,
		}

		storage, err := CreateStorageBackend(FileSystemStorage, config)
		if err != nil {
			t.Fatalf("Failed to create storage backend: %v", err)
		}

		// Test that it works
		session := storage.NewSession("factory-test")
		err = storage.SaveSession(session)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		// Verify file exists
		sessionFile := filepath.Join(tmpDir, session.ID+".json")
		if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
			t.Error("Session file was not created")
		}
	})

	t.Run("UnknownBackendType", func(t *testing.T) {
		_, err := CreateStorageBackend(StorageType("unknown"), nil)
		if err == nil {
			t.Error("Expected error for unknown storage type")
		}
		if !strings.Contains(err.Error(), "unknown storage type") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

// TestSessionManager tests the SessionManager with storage backend
func TestSessionManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "magellai-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a storage backend
	storage, err := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
		"base_dir": tmpDir,
	})
	if err != nil {
		t.Fatalf("Failed to create storage backend: %v", err)
	}

	// Create session manager
	manager, err := NewSessionManager(storage)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	t.Run("CreateAndSaveSession", func(t *testing.T) {
		session, err := manager.NewSession("manager-test")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		if session == nil {
			t.Fatal("NewSession returned nil")
		}

		// Load it back
		loaded, err := manager.LoadSession(session.ID)
		if err != nil {
			t.Fatalf("Failed to load session: %v", err)
		}

		if loaded.ID != session.ID {
			t.Errorf("Expected ID %s, got %s", session.ID, loaded.ID)
		}
	})
}
