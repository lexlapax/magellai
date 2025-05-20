// ABOUTME: Test fixtures for configuration objects
// ABOUTME: Provides reusable test configurations for testing

package fixtures

import (
	"io"
	"os"
	"path/filepath"
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

// CleanupTestConfigFile removes a test configuration file
func CleanupTestConfigFile(configPath string) error {
	if configPath == "" {
		return nil
	}

	return os.Remove(configPath)
}

// GetDefaultTestConfig returns a default test configuration as YAML
func GetDefaultTestConfig() string {
	return `
# Test configuration
provider:
  default: openai
  openai:
    api_key: ${OPENAI_API_KEY}
    default_model: gpt-4o
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    default_model: claude-2

model:
  default: openai/gpt-4o
  settings:
    openai/gpt-4o:
      temperature: 0.7
      max_tokens: 2048
    anthropic/claude-2:
      temperature: 0.7
      max_tokens: 4096

session:
  directory: ~/.magellai/sessions
  autosave: true
  storage:
    type: filesystem
    settings:
      base_dir: ~/.magellai
`
}

// CreateConfigDir creates a test configuration directory
func CreateConfigDir() (string, func(), error) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "magellai-test")
	if err != nil {
		return "", nil, err
	}

	// Create config file
	_, err = CreateTestConfigFile(dir, "")
	if err != nil {
		os.RemoveAll(dir)
		return "", nil, err
	}

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup, nil
}

// CreateConfigFileWithContent creates a config file with the specified content
func CreateConfigFileWithContent(content string) (string, func(), error) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "magellai-test")
	if err != nil {
		return "", nil, err
	}

	// Create config file
	configPath, err := CreateTestConfigFile(dir, content)
	if err != nil {
		os.RemoveAll(dir)
		return "", nil, err
	}

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return configPath, cleanup, nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
}
