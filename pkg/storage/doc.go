// ABOUTME: Storage package providing persistence for sessions and conversations
// ABOUTME: Implements multiple backend options including filesystem and SQLite

/*
Package storage provides session persistence capabilities for Magellai.

The storage package offers a unified Backend interface with multiple implementations
that handle saving, loading, searching, and managing chat sessions. It's designed
to abstract storage concerns from the rest of the application while providing
robust persistence capabilities.

Key Components:
  - Backend Interface: The core interface implemented by all storage providers
  - Factory: Creates storage backends based on configuration
  - File System Backend: Stores sessions as JSON files in a directory structure
  - SQLite Backend: Provides efficient querying and search for sessions
  - Error Handling: Standardized error types for all storage operations

Usage:

	// Create a storage backend from configuration
	backend, err := storage.NewBackend(config)
	if err != nil {
	    // Handle error
	}
	defer backend.Close()

	// Store a session
	err = backend.SaveSession(session)

	// Retrieve a session
	session, err := backend.GetSession(sessionID)

	// Search across sessions
	results, err := backend.SearchSessions(query)

The Backend interface is aligned with domain.SessionRepository to ensure
proper domain separation while providing storage-specific functionality.
*/
package storage
