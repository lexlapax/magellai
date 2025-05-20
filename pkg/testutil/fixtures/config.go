// ABOUTME: Test fixtures for configuration objects
// ABOUTME: Provides reusable test configurations for testing

package fixtures

import (
	"io"
	"os"
	"path/filepath"

	"github.com/lexlapax/magellai/pkg/config"
)

// CreateTestConfigFile creates a temporary test configuration file
func CreateTestConfigFile(dir string, content string) (string, error) {
	if dir == "" {
		dir = os.TempDir()
	}

	if content == "" {
		content = GetDefaultTestConfig()
	}

	configPath := filepath.Join(dir, ".magellai.yaml")
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return configPath, nil
}

// GetDefaultTestConfig returns a default test configuration
func GetDefaultTestConfig() string {
	return `
log:
  level: warn
  format: text

provider:
  default: openai
  
model: openai/gpt-3.5-turbo

output:
  format: text
  color: false
  pretty: true

session:
  directory: ~/.config/magellai/sessions
  autosave: true
  storage:
    type: filesystem

repl:
  prompt: ">>> "
  streaming: true
  auto_save:
    enabled: true
    interval: 5m
  colors:
    enabled: false

profiles:
  test:
    description: Test profile
    provider: mock
    model: mock/test-model
    settings:
      log:
        level: debug
  development:
    description: Development profile
    provider: openai
    model: openai/gpt-3.5-turbo
    settings:
      output:
        format: json
`
}

// GetMinimalTestConfig returns a minimal test configuration
func GetMinimalTestConfig() string {
	return `
provider:
  default: mock
model: mock/test
`
}

// CreateTestProviderConfig creates a test provider configuration
func CreateTestProviderConfig() map[string]interface{} {
	return map[string]interface{}{
		"api_key":     "test-key",
		"base_url":    "https://api.test.com",
		"timeout":     30,
		"max_retries": 3,
	}
}

// CreateTestProfileConfig creates a test profile configuration
func CreateTestProfileConfig(name string) *config.ProfileConfig {
	return &config.ProfileConfig{
		Description: "Test profile " + name,
		Provider:    "test",
		Model:       "test/model",
		Settings: map[string]interface{}{
			"temperature": 0.5,
			"max_tokens":  150,
		},
	}
}

// CreateTestStorageConfig creates a test storage configuration
func CreateTestStorageConfig(storageType string) *config.StorageConfig {
	cfg := &config.StorageConfig{
		Type:     storageType,
		Settings: make(map[string]interface{}),
	}

	switch storageType {
	case "filesystem":
		cfg.Settings["base_dir"] = "/tmp/test-sessions"
		cfg.Settings["user_id"] = "test-user"
	case "sqlite":
		cfg.Settings["db_path"] = "/tmp/test.db"
		cfg.Settings["user_id"] = "test-user"
	case "memory":
		cfg.Settings["max_size"] = 1000
	}

	return cfg
}

// CreateTestEnvironment sets up test environment variables
func CreateTestEnvironment(cleanup func()) {
	// Set test environment variables
	os.Setenv("MAGELLAI_LOG_LEVEL", "warn")
	os.Setenv("MAGELLAI_PROVIDER_DEFAULT", "mock")
	os.Setenv("MAGELLAI_MODEL", "mock/test")

	// Register cleanup if provided
	if cleanup != nil {
		cleanup()
	}
}

// CleanupTestEnvironment removes test environment variables
func CleanupTestEnvironment() {
	os.Unsetenv("MAGELLAI_LOG_LEVEL")
	os.Unsetenv("MAGELLAI_PROVIDER_DEFAULT")
	os.Unsetenv("MAGELLAI_MODEL")
}

// MockConfigManager implements a mock configuration manager for testing
type MockConfigManager struct {
	config map[string]interface{}
}

// NewMockConfigManager creates a new mock config manager
func NewMockConfigManager() *MockConfigManager {
	return &MockConfigManager{
		config: make(map[string]interface{}),
	}
}

// Get returns a configuration value
func (m *MockConfigManager) Get(key string) interface{} {
	return m.config[key]
}

// Set sets a configuration value
func (m *MockConfigManager) Set(key string, value interface{}) error {
	m.config[key] = value
	return nil
}

// LoadFile mocks loading a configuration file
func (m *MockConfigManager) LoadFile(path string) error {
	// Mock implementation - just return success
	return nil
}

// Save mocks saving configuration
func (m *MockConfigManager) Save(w io.Writer) error {
	// Mock implementation - just return success
	return nil
}
