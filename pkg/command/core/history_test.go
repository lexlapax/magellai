package core

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/repl"
	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem" // Register filesystem backend
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test message
func createTestMessage(role, content string) domain.Message {
	return domain.Message{
		ID:        uuid.New().String(),
		Role:      domain.MessageRole(role),
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

func TestHistoryCommand_Metadata(t *testing.T) {
	cmd := NewHistoryCommand()
	metadata := cmd.Metadata()

	assert.Equal(t, "history", metadata.Name)
	assert.NotEmpty(t, metadata.Description)
	assert.NotEmpty(t, metadata.LongDescription)
}

func TestHistoryCommand_Execute_Invalid(t *testing.T) {
	cmd := NewHistoryCommand()
	var output bytes.Buffer
	exec := &command.ExecutionContext{
		Args:   []string{}, // No subcommand
		Flags:  command.NewFlags(nil),
		Stdout: &output,
		Data:   make(map[string]interface{}),
	}

	err := cmd.Execute(context.Background(), exec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no subcommand specified")
}

func TestHistoryCommand_Execute_List(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "history-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := repl.NewStorageManager(backend)
	require.NoError(t, err)

	manager, err := repl.NewSessionManager(storageManager)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Create and save a test session
	session, err := manager.NewSession("test-session")
	require.NoError(t, err)
	msg := createTestMessage("user", "test message")
	session.Conversation.AddMessage(msg)
	err = manager.SaveSession(session)
	require.NoError(t, err)

	// Create command and execute list
	cmd := NewHistoryCommand()
	var output bytes.Buffer
	exec := &command.ExecutionContext{
		Args:   []string{"list"},
		Flags:  command.NewFlags(nil),
		Stdout: &output,
		Data: map[string]interface{}{
			"session_manager": manager,
		},
	}

	err = cmd.Execute(context.Background(), exec)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "test-session")
}

func TestHistoryCommand_Execute_Show(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "history-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := repl.NewStorageManager(backend)
	require.NoError(t, err)

	manager, err := repl.NewSessionManager(storageManager)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Create and save a test session
	session, err := manager.NewSession("test-session")
	require.NoError(t, err)
	msg := createTestMessage("user", "test message")
	session.Conversation.AddMessage(msg)
	err = manager.SaveSession(session)
	require.NoError(t, err)

	// Create command and execute show
	cmd := NewHistoryCommand()
	var output bytes.Buffer
	exec := &command.ExecutionContext{
		Args:   []string{"show", session.ID},
		Flags:  command.NewFlags(nil),
		Stdout: &output,
		Data: map[string]interface{}{
			"session_manager": manager,
		},
	}

	err = cmd.Execute(context.Background(), exec)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "test message")
}

func TestHistoryCommand_Execute_Export(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "history-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := repl.NewStorageManager(backend)
	require.NoError(t, err)

	manager, err := repl.NewSessionManager(storageManager)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Create and save a test session
	session, err := manager.NewSession("test-session")
	require.NoError(t, err)
	msg := createTestMessage("user", "test message")
	session.Conversation.AddMessage(msg)
	err = manager.SaveSession(session)
	require.NoError(t, err)

	// Test JSON export
	t.Run("json export", func(t *testing.T) {
		cmd := NewHistoryCommand()
		var output bytes.Buffer

		// Create flags
		flags := command.NewFlags(nil)
		flags.Set("format", "json")

		exec := &command.ExecutionContext{
			Args:   []string{"export", session.ID},
			Flags:  flags,
			Stdout: &output,
			Data: map[string]interface{}{
				"session_manager": manager,
			},
		}

		err = cmd.Execute(context.Background(), exec)
		assert.NoError(t, err)
		// Output should contain the JSON data
		assert.Contains(t, output.String(), "test message")
	})

	// Test Markdown export
	t.Run("markdown export", func(t *testing.T) {
		cmd := NewHistoryCommand()
		var output bytes.Buffer

		// Create flags
		flags := command.NewFlags(nil)
		flags.Set("format", "markdown")

		exec := &command.ExecutionContext{
			Args:   []string{"export", session.ID},
			Flags:  flags,
			Stdout: &output,
			Data: map[string]interface{}{
				"session_manager": manager,
			},
		}

		err = cmd.Execute(context.Background(), exec)
		assert.NoError(t, err)
		// Output should contain the Markdown data
		assert.Contains(t, output.String(), "# Session")
		assert.Contains(t, output.String(), "test message")
	})
}

func TestHistoryCommand_Execute_Delete(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "history-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := repl.NewStorageManager(backend)
	require.NoError(t, err)

	manager, err := repl.NewSessionManager(storageManager)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Create and save a test session
	session, err := manager.NewSession("test-session")
	require.NoError(t, err)
	err = manager.SaveSession(session)
	require.NoError(t, err)

	// Delete the session
	cmd := NewHistoryCommand()
	var output bytes.Buffer
	exec := &command.ExecutionContext{
		Args:   []string{"delete", session.ID},
		Flags:  command.NewFlags(nil),
		Stdout: &output,
		Data: map[string]interface{}{
			"session_manager": manager,
		},
	}

	err = cmd.Execute(context.Background(), exec)
	assert.NoError(t, err)

	// Verify session was deleted
	_, err = manager.LoadSession(session.ID)
	assert.Error(t, err)
}

func TestHistoryCommand_Execute_Search(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "history-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := repl.NewStorageManager(backend)
	require.NoError(t, err)

	manager, err := repl.NewSessionManager(storageManager)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Create and save test sessions
	session1, err := manager.NewSession("Python session")
	require.NoError(t, err)
	msg1 := createTestMessage("user", "How to use Python decorators?")
	session1.Conversation.AddMessage(msg1)
	err = manager.SaveSession(session1)
	require.NoError(t, err)

	session2, err := manager.NewSession("JavaScript session")
	require.NoError(t, err)
	msg2 := createTestMessage("user", "What is JavaScript async/await?")
	session2.Conversation.AddMessage(msg2)
	err = manager.SaveSession(session2)
	require.NoError(t, err)

	// Search for Python
	cmd := NewHistoryCommand()
	var output bytes.Buffer
	exec := &command.ExecutionContext{
		Args:   []string{"search", "Python"},
		Flags:  command.NewFlags(nil),
		Stdout: &output,
		Data: map[string]interface{}{
			"session_manager": manager,
		},
	}

	err = cmd.Execute(context.Background(), exec)
	assert.NoError(t, err)
	outputStr := output.String()
	assert.Contains(t, outputStr, "Python session")
	assert.Contains(t, outputStr, "decorators")
	assert.NotContains(t, outputStr, "JavaScript")
}
