// ABOUTME: Tests for the storage backend interface
// ABOUTME: Ensures the interface is properly defined and constants are correct

package storage

import (
	"bytes"
	"testing"

	"github.com/lexlapax/magellai/internal/testutil/fixtures"
	"github.com/lexlapax/magellai/internal/testutil/storagemock"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestBackendType_Constants(t *testing.T) {
	// Verify backend type constants
	assert.Equal(t, BackendType("filesystem"), FileSystemBackend)
	assert.Equal(t, BackendType("sqlite"), SQLiteBackend)
	assert.Equal(t, BackendType("postgresql"), PostgreSQLBackend)
	assert.Equal(t, BackendType("memory"), MemoryBackend)
}

func TestConfig_Type(t *testing.T) {
	// Test Config type
	config := Config{
		"base_dir": "/tmp/test",
		"db_path":  "/tmp/test.db",
		"user_id":  "test-user",
	}

	// Verify type assertions work
	baseDir, ok := config["base_dir"].(string)
	assert.True(t, ok)
	assert.Equal(t, "/tmp/test", baseDir)

	dbPath, ok := config["db_path"].(string)
	assert.True(t, ok)
	assert.Equal(t, "/tmp/test.db", dbPath)
}

func TestBackendInterface_Compliance(t *testing.T) {
	// Create mock backend using centralized mock
	mock := storagemock.NewMockBackend()

	// Ensure it implements the storagemock.Backend interface
	var _ storagemock.Backend = mock

	// Test basic operations
	session := mock.NewSession("test-session")
	assert.NotNil(t, session)
	assert.Equal(t, "test-session", session.Name)

	// Save session
	err := mock.SaveSession(session)
	assert.NoError(t, err)

	// Load session
	loaded, err := mock.LoadSession(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)

	// List sessions
	sessions, err := mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)

	// Delete session
	err = mock.DeleteSession(session.ID)
	assert.NoError(t, err)

	// Verify session was deleted
	sessions, err = mock.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 0)

	// Close backend
	err = mock.Close()
	assert.NoError(t, err)
}

func TestBackendInterface_Search(t *testing.T) {
	mock := storagemock.NewMockBackend()

	// Create test sessions using fixtures
	session1 := fixtures.CreateTestSessionWithMessages("session-1", []domain.Message{})
	session2 := fixtures.CreateTestSessionWithMessages("session-2", []domain.Message{})

	// Save sessions
	assert.NoError(t, mock.SaveSession(session1))
	assert.NoError(t, mock.SaveSession(session2))

	// Search (mock implementation returns empty results)
	results, err := mock.SearchSessions("test")
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestBackendInterface_Export(t *testing.T) {
	mock := storagemock.NewMockBackend()

	// Create test session using fixture
	session := fixtures.CreateTestSession("export-test")
	assert.NoError(t, mock.SaveSession(session))

	// Export session
	var buf bytes.Buffer
	err := mock.ExportSession(session.ID, domain.ExportFormatJSON, &buf)
	assert.NoError(t, err)

	// Verify export wrote something
	assert.NotEmpty(t, buf.String())
}

func TestBackendInterface_Branching(t *testing.T) {
	mock := storagemock.NewMockBackend()

	// Create test sessions with branches
	sessionFamily := fixtures.CreateSessionFamily()
	root := sessionFamily["parent-session"]

	// Save all sessions
	assert.NoError(t, mock.SaveSession(root))
	for _, branch := range sessionFamily {
		if branch.ID != root.ID {
			assert.NoError(t, mock.SaveSession(branch))
		}
	}

	// Get children
	children, err := mock.GetChildren(root.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, children)

	// Get branch tree
	tree, err := mock.GetBranchTree(root.ID)
	assert.NoError(t, err)
	assert.NotNil(t, tree)
}

func TestBackendInterface_Merge(t *testing.T) {
	mock := storagemock.NewMockBackend()

	// Create test sessions for merge
	sessionFamily := fixtures.CreateSessionFamily()
	target := sessionFamily["parent-session"]
	source := sessionFamily["branch-a"]

	// Save sessions
	assert.NoError(t, mock.SaveSession(target))
	assert.NoError(t, mock.SaveSession(source))

	// Merge sessions
	options := domain.MergeOptions{
		Type:       domain.MergeTypeContinuation,
		SourceID:   source.ID,
		TargetID:   target.ID,
		MergePoint: 0,
	}
	result, err := mock.MergeSessions(target.ID, source.ID, options)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBackendInterface_ErrorHandling(t *testing.T) {
	mock := storagemock.NewMockBackend()

	// Set error for load operation
	testErr := assert.AnError
	mock.WithLoadError(testErr)

	// Test error propagation
	_, err := mock.LoadSession("test-id")
	assert.Equal(t, testErr, err)
}
