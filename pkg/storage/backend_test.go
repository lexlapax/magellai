// ABOUTME: Tests for the storage backend interface
// ABOUTME: Ensures the interface is properly defined and constants are correct

package storage

import (
	"io"
	"testing"

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

// MockBackend implements the Backend interface for testing
type MockBackend struct {
	sessions map[string]*Session
	closed   bool
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		sessions: make(map[string]*Session),
	}
}

func (m *MockBackend) NewSession(name string) *Session {
	return &Session{
		ID:   "mock-" + name,
		Name: name,
	}
}

func (m *MockBackend) SaveSession(session *Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *MockBackend) LoadSession(id string) (*Session, error) {
	session, exists := m.sessions[id]
	if !exists {
		return nil, nil
	}
	return session, nil
}

func (m *MockBackend) ListSessions() ([]*SessionInfo, error) {
	var infos []*SessionInfo
	for _, session := range m.sessions {
		infos = append(infos, &SessionInfo{
			ID:   session.ID,
			Name: session.Name,
		})
	}
	return infos, nil
}

func (m *MockBackend) DeleteSession(id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *MockBackend) SearchSessions(query string) ([]*SearchResult, error) {
	return nil, nil
}

func (m *MockBackend) ExportSession(id string, format ExportFormat, w io.Writer) error {
	return nil
}

func (m *MockBackend) Close() error {
	m.closed = true
	return nil
}

// Compile-time check that MockBackend implements Backend
var _ Backend = (*MockBackend)(nil)

func TestMockBackend_Implementation(t *testing.T) {
	backend := NewMockBackend()

	// Test NewSession
	session := backend.NewSession("test")
	assert.NotNil(t, session)
	assert.Equal(t, "mock-test", session.ID)
	assert.Equal(t, "test", session.Name)

	// Test SaveSession
	err := backend.SaveSession(session)
	assert.NoError(t, err)

	// Test LoadSession
	loaded, err := backend.LoadSession(session.ID)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, session.ID, loaded.ID)

	// Test ListSessions
	sessions, err := backend.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)

	// Test DeleteSession
	err = backend.DeleteSession(session.ID)
	assert.NoError(t, err)

	// Test Close
	err = backend.Close()
	assert.NoError(t, err)
	assert.True(t, backend.closed)
}
