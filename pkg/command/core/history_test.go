package core

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/repl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryCommand_Metadata(t *testing.T) {
	cmd := NewHistoryCommand()
	metadata := cmd.Metadata()

	assert.Equal(t, "history", metadata.Name)
	assert.NotEqual(t, command.Category(0), metadata.Category)
	assert.NotEmpty(t, metadata.Description)
	assert.NotEmpty(t, metadata.LongDescription)
	assert.Len(t, metadata.Flags, 1)
	assert.Equal(t, "format", metadata.Flags[0].Name)
}

func TestHistoryCommand_Validate(t *testing.T) {
	cmd := NewHistoryCommand()
	err := cmd.Validate()
	assert.NoError(t, err)
}

func TestHistoryCommand_ExecuteList(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "magellai-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test config directory
	cleanup := setupTestConfig(tmpDir)
	defer cleanup()

	// Create a session manager with storage backend
	sessionsDir := filepath.Join(tmpDir, ".config", "magellai", "sessions")
	storage, err := repl.CreateStorageBackend(repl.FileSystemStorage, map[string]interface{}{
		"base_dir": sessionsDir,
	})
	require.NoError(t, err)

	manager, err := repl.NewSessionManager(storage)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Create and save a test session
	session, err := manager.NewSession("test-session")
	require.NoError(t, err)
	session.Conversation.AddMessage("user", "test message", nil)
	err = manager.SaveSession(session)
	require.NoError(t, err)

	// Create command and execute list
	cmd := NewHistoryCommand()
	ctx := context.Background()

	// Create execution context
	var buf bytes.Buffer
	exec := &command.ExecutionContext{
		Args:    []string{"list"},
		Flags:   command.NewFlags(nil),
		Stdout:  &buf,
		Stderr:  os.Stderr,
		Context: ctx,
		Data:    make(map[string]interface{}),
	}

	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "test-session")
}

func TestHistoryCommand_ExecuteShow(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "magellai-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test config directory
	cleanup := setupTestConfig(tmpDir)
	defer cleanup()

	// Create a session manager with storage backend
	sessionsDir := filepath.Join(tmpDir, ".config", "magellai", "sessions")
	storage, err := repl.CreateStorageBackend(repl.FileSystemStorage, map[string]interface{}{
		"base_dir": sessionsDir,
	})
	require.NoError(t, err)

	manager, err := repl.NewSessionManager(storage)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Create and save a test session
	session, err := manager.NewSession("test-session")
	require.NoError(t, err)
	session.Conversation.AddMessage("user", "test message", nil)
	err = manager.SaveSession(session)
	require.NoError(t, err)

	// Create command and execute show
	cmd := NewHistoryCommand()
	ctx := context.Background()

	// Create execution context
	var buf bytes.Buffer
	exec := &command.ExecutionContext{
		Args:    []string{"show", session.ID},
		Flags:   command.NewFlags(nil),
		Stdout:  &buf,
		Stderr:  os.Stderr,
		Context: ctx,
		Data:    make(map[string]interface{}),
	}

	err = cmd.Execute(ctx, exec)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Session ID:")
	assert.Contains(t, output, "test-session")
	assert.Contains(t, output, "test message")
}

func TestHistoryCommand_ExecuteInvalid(t *testing.T) {
	cmd := NewHistoryCommand()
	ctx := context.Background()

	// Test no arguments
	exec := &command.ExecutionContext{
		Args:    []string{},
		Flags:   command.NewFlags(nil),
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Context: ctx,
		Data:    make(map[string]interface{}),
	}

	err := cmd.Execute(ctx, exec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no subcommand specified")

	// Test unknown subcommand
	exec.Args = []string{"unknown"}
	err = cmd.Execute(ctx, exec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown subcommand")

	// Test show without ID
	exec.Args = []string{"show"}
	err = cmd.Execute(ctx, exec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session ID required")
}
