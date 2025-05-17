//go:build !sqlite && !db
// +build !sqlite,!db

// ABOUTME: Placeholder for when SQLite support is not compiled in
// ABOUTME: Provides error message when SQLite is requested but not available

package repl

import "fmt"

// RegisterSQLiteStorage registers a placeholder when SQLite is not compiled in
func init() {
	RegisterStorageBackend(SQLiteStorage, func(config map[string]interface{}) (StorageBackend, error) {
		return nil, fmt.Errorf("SQLite storage not available: compile with -tags sqlite or -tags db")
	})
}
