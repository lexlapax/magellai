// ABOUTME: Stub for when SQLite support is not compiled in
// ABOUTME: Registers an error-returning factory when SQLite is not available

//go:build !sqlite && !db

package sqlite

import (
	"fmt"

	"github.com/lexlapax/magellai/pkg/storage"
)

func init() {
	storage.RegisterBackend(storage.SQLiteBackend, func(config storage.Config) (storage.Backend, error) {
		return nil, fmt.Errorf("SQLite support not compiled in: build with -tags sqlite or -tags db")
	})
}
