//go:build integration
// +build integration

package main

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/magellai/pkg/config"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

func TestConfigurationInit(t *testing.T) {
	// Initialize configuration
	err := config.Init()
	require.NoError(t, err)
	assert.NotNil(t, config.Manager)

	// Set some values
	err = config.Manager.SetValue("provider.default", "openai")
	assert.NoError(t, err)

	// Get a value
	val := config.Manager.Get("provider.default")
	assert.Equal(t, "openai", val)
}

// Simple mock backend for testing
type testMockBackend struct {
	sessions map[string]*domain.Session
}

func newTestMockBackend() *testMockBackend {
	return &testMockBackend{
		sessions: make(map[string]*domain.Session),
	}
}

func (mb *testMockBackend) NewSession(name string) *domain.Session {
	session := domain.NewSession(storage.GenerateSessionID())
	session.Name = name
	return session
}

func (mb *testMockBackend) SaveSession(session *domain.Session) error {
	mb.sessions[session.ID] = session
	return nil
}

func (mb *testMockBackend) LoadSession(id string) (*domain.Session, error) {
	session, ok := mb.sessions[id]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (mb *testMockBackend) ListSessions() ([]*domain.SessionInfo, error) {
	var sessions []*domain.SessionInfo
	for _, session := range mb.sessions {
		sessions = append(sessions, session.ToSessionInfo())
	}
	return sessions, nil
}

func (mb *testMockBackend) DeleteSession(id string) error {
	delete(mb.sessions, id)
	return nil
}

func (mb *testMockBackend) SearchSessions(query string) ([]*domain.SearchResult, error) {
	return []*domain.SearchResult{}, nil
}

func (mb *testMockBackend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	return nil
}

func (mb *testMockBackend) Close() error {
	return nil
}

func TestMockStorageBackend(t *testing.T) {
	// Create mock backend
	backend := newTestMockBackend()
	require.NotNil(t, backend)

	// Test basic session operations
	session := backend.NewSession("test-session")
	assert.NotNil(t, session)
	assert.Equal(t, "test-session", session.Name)

	// Save the session
	err := backend.SaveSession(session)
	assert.NoError(t, err)

	// Load the session
	loaded, err := backend.LoadSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, loaded.ID)

	// List sessions
	sessions, err := backend.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, session.ID, sessions[0].ID)
}

func TestConfigSchema(t *testing.T) {
	// Initialize configuration
	err := config.Init()
	require.NoError(t, err)

	// Load defaults
	err = config.Manager.Load(nil)
	require.NoError(t, err)

	// Get schema
	schema, err := config.Manager.GetSchema()
	require.NoError(t, err)
	assert.NotNil(t, schema)

	// Verify defaults
	assert.Equal(t, "info", schema.Log.Level)
	assert.Equal(t, "text", schema.Output.Format)
	assert.True(t, schema.Output.Color)
}

func TestConfigProfileOverride(t *testing.T) {
	// Initialize configuration
	err := config.Init()
	require.NoError(t, err)

	// Set base values
	err = config.Manager.SetValue("log.level", "info")
	require.NoError(t, err)
	err = config.Manager.SetValue("output.format", "text")
	require.NoError(t, err)

	// Create profile settings
	err = config.Manager.SetValue("profiles.test.log.level", "debug")
	require.NoError(t, err)
	err = config.Manager.SetValue("profiles.test.output.format", "json")
	require.NoError(t, err)

	// Apply profile
	err = config.Manager.SetProfile("test")
	require.NoError(t, err)

	// Verify profile overrides
	schema, err := config.Manager.GetSchema()
	require.NoError(t, err)
	assert.Equal(t, "debug", schema.Log.Level)
	assert.Equal(t, "json", schema.Output.Format)
}
