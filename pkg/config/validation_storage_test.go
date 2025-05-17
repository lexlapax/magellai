// ABOUTME: Tests for configuration validation with storage backend checks
// ABOUTME: Ensures storage backend availability is correctly validated

package config

import (
	"testing"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem" // Register filesystem backend
	_ "github.com/lexlapax/magellai/pkg/storage/sqlite"     // Register SQLite backend
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStorageBackend(t *testing.T) {
	tests := []struct {
		name        string
		config      func(t *testing.T) map[string]interface{}
		expectError bool
		errorField  string
	}{
		{
			name: "filesystem backend - always available",
			config: func(t *testing.T) map[string]interface{} {
				return map[string]interface{}{
					"session": map[string]interface{}{
						"storage": map[string]interface{}{
							"type": "filesystem",
							"settings": map[string]interface{}{
								"base_dir": t.TempDir(),
							},
						},
					},
				}
			},
			expectError: false,
		},
		{
			name: "sqlite backend - should fail if not compiled in",
			config: func(t *testing.T) map[string]interface{} {
				return map[string]interface{}{
					"session": map[string]interface{}{
						"storage": map[string]interface{}{
							"type": "sqlite",
							"settings": map[string]interface{}{
								"db_path": t.TempDir() + "/test.db",
							},
						},
					},
				}
			},
			// This will succeed if SQLite is compiled in, fail otherwise
			expectError: !storage.IsBackendAvailable(storage.SQLiteBackend),
			errorField:  "session.storage.type",
		},
		{
			name: "invalid backend type",
			config: func(t *testing.T) map[string]interface{} {
				return map[string]interface{}{
					"session": map[string]interface{}{
						"storage": map[string]interface{}{
							"type": "nosuchbackend",
						},
					},
				}
			},
			expectError: true,
			errorField:  "session.storage.type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a Config instance directly for testing
			testConfig := tt.config(t)
			config := &Config{
				koanf:    koanf.New("."),
				defaults: testConfig,
			}

			// Load the test configuration
			require.NoError(t, config.koanf.Load(confmap.Provider(testConfig, "."), nil))

			// Validate session configuration
			errors := config.validateSessionConfig()

			if tt.expectError {
				assert.NotEmpty(t, errors)
				// Check if the error contains the expected field
				if tt.errorField != "" {
					found := false
					for _, err := range errors {
						if err.Field == tt.errorField {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error in field %s but not found", tt.errorField)
				}
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestGetAvailableBackends(t *testing.T) {
	backends := storage.GetAvailableBackends()

	// Filesystem should always be available
	found := false
	for _, b := range backends {
		if b == storage.FileSystemBackend {
			found = true
			break
		}
	}
	assert.True(t, found, "FileSystemBackend should always be available")

	// SQLite may or may not be available depending on build tags
	t.Logf("Available backends: %v", backends)
}
