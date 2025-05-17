// ABOUTME: Mock storage backend for testing
// ABOUTME: Provides a mock implementation of storage.Backend for unit tests

package repl

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lexlapax/magellai/pkg/storage"
)

// MockStorageBackend implements storage.Backend for testing
type MockStorageBackend struct {
	sessions map[string]*storage.Session
	calls    map[string]int
	err      error
}

// NewMockStorageBackend creates a new mock storage backend
func NewMockStorageBackend() *MockStorageBackend {
	return &MockStorageBackend{
		sessions: make(map[string]*storage.Session),
		calls:    make(map[string]int),
	}
}

func (m *MockStorageBackend) NewSession(name string) *storage.Session {
	m.calls["NewSession"]++
	return &storage.Session{
		ID:       "test-session-" + fmt.Sprint(time.Now().Unix()),
		Name:     name,
		Messages: []storage.Message{},
		Config:   make(map[string]interface{}),
		Created:  time.Now(),
		Updated:  time.Now(),
		Metadata: make(map[string]interface{}),
		Tags:     []string{},
	}
}

func (m *MockStorageBackend) SaveSession(session *storage.Session) error {
	m.calls["SaveSession"]++
	if m.err != nil {
		return m.err
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *MockStorageBackend) LoadSession(id string) (*storage.Session, error) {
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

func (m *MockStorageBackend) ListSessions() ([]*storage.SessionInfo, error) {
	m.calls["ListSessions"]++
	if m.err != nil {
		return nil, m.err
	}
	var infos []*storage.SessionInfo
	for _, session := range m.sessions {
		infos = append(infos, &storage.SessionInfo{
			ID:           session.ID,
			Name:         session.Name,
			Created:      session.Created,
			Updated:      session.Updated,
			MessageCount: len(session.Messages),
			Tags:         session.Tags,
		})
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

func (m *MockStorageBackend) SearchSessions(query string) ([]*storage.SearchResult, error) {
	m.calls["SearchSessions"]++
	if m.err != nil {
		return nil, m.err
	}
	var results []*storage.SearchResult
	// Simple mock implementation
	for _, session := range m.sessions {
		if strings.Contains(strings.ToLower(session.Name), strings.ToLower(query)) {
			results = append(results, &storage.SearchResult{
				Session: &storage.SessionInfo{
					ID:   session.ID,
					Name: session.Name,
				},
				Matches: []storage.SearchMatch{
					{
						Type:    "name",
						Content: session.Name,
						Context: "Session Name",
					},
				},
			})
		}
	}
	return results, nil
}

func (m *MockStorageBackend) ExportSession(id string, format storage.ExportFormat, w io.Writer) error {
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
