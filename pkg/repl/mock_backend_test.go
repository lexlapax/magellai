// ABOUTME: Mock storage backend for testing
// ABOUTME: Provides a mock implementation of storage.Backend for unit tests

package repl

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

// MockStorageBackend implements storage.Backend for testing
type MockStorageBackend struct {
	sessions map[string]*domain.Session
	calls    map[string]int
	err      error
}

// NewMockStorageBackend creates a new mock storage backend
func NewMockStorageBackend() *MockStorageBackend {
	return &MockStorageBackend{
		sessions: make(map[string]*domain.Session),
		calls:    make(map[string]int),
	}
}

// AsBackend returns the mock as a storage.Backend interface
func (m *MockStorageBackend) AsBackend() storage.Backend {
	return m
}

func (m *MockStorageBackend) NewSession(name string) *domain.Session {
	m.calls["NewSession"]++
	sessionID := "test-session-" + fmt.Sprint(time.Now().Unix())
	session := domain.NewSession(sessionID)
	session.Name = name
	return session
}

func (m *MockStorageBackend) SaveSession(session *domain.Session) error {
	m.calls["SaveSession"]++
	if m.err != nil {
		return m.err
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *MockStorageBackend) LoadSession(id string) (*domain.Session, error) {
	m.calls["LoadSession"]++
	if m.err != nil {
		return nil, m.err
	}
	session, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

func (m *MockStorageBackend) ListSessions() ([]*domain.SessionInfo, error) {
	m.calls["ListSessions"]++
	if m.err != nil {
		return nil, m.err
	}
	var infos []*domain.SessionInfo
	for _, session := range m.sessions {
		infos = append(infos, session.ToSessionInfo())
	}
	return infos, nil
}

func (m *MockStorageBackend) DeleteSession(id string) error {
	m.calls["DeleteSession"]++
	if m.err != nil {
		return m.err
	}
	_, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	delete(m.sessions, id)
	return nil
}

func (m *MockStorageBackend) SearchSessions(query string) ([]*domain.SearchResult, error) {
	m.calls["SearchSessions"]++
	if m.err != nil {
		return nil, m.err
	}
	var results []*domain.SearchResult
	// Simple mock implementation
	for _, session := range m.sessions {
		if strings.Contains(strings.ToLower(session.Name), strings.ToLower(query)) {
			sessionInfo := session.ToSessionInfo()
			result := domain.NewSearchResult(sessionInfo)
			result.AddMatch(domain.NewSearchMatch(
				domain.SearchMatchTypeName,
				"",
				session.Name,
				"Session Name",
				-1,
			))
			results = append(results, result)
		}
	}
	return results, nil
}

func (m *MockStorageBackend) ExportSession(id string, format domain.ExportFormat, w io.Writer) error {
	m.calls["ExportSession"]++
	if m.err != nil {
		return m.err
	}
	session, ok := m.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	_, err := fmt.Fprintf(w, "Exported session %s (%s)", session.ID, format)
	return err
}

func (m *MockStorageBackend) Close() error {
	m.calls["Close"]++
	return m.err
}
