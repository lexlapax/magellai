// ABOUTME: Unit tests for configuration directory management
// ABOUTME: Tests directory creation, path resolution, and default config generation
package configdir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetPaths(t *testing.T) {
	paths, err := GetPaths()
	if err != nil {
		t.Fatalf("GetPaths() error = %v", err)
	}

	if paths.Base == "" {
		t.Error("Base path is empty")
	}
	if paths.Sessions == "" {
		t.Error("Sessions path is empty")
	}
	if paths.Plugins == "" {
		t.Error("Plugins path is empty")
	}
	if paths.Logs == "" {
		t.Error("Logs path is empty")
	}

	// Verify paths are properly constructed
	if !filepath.IsAbs(paths.Base) {
		t.Error("Base path is not absolute")
	}
	if !strings.HasPrefix(paths.Sessions, paths.Base) {
		t.Error("Sessions path is not under base path")
	}
	if !strings.HasPrefix(paths.Plugins, paths.Base) {
		t.Error("Plugins path is not under base path")
	}
	if !strings.HasPrefix(paths.Logs, paths.Base) {
		t.Error("Logs path is not under base path")
	}
}

func TestEnsureDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "magellai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Ensure directories are created
	err = EnsureDirectories()
	if err != nil {
		t.Fatalf("EnsureDirectories() error = %v", err)
	}

	// Verify directories exist
	paths, _ := GetPaths()
	dirs := []string{
		paths.Base,
		paths.Sessions,
		paths.Plugins,
		paths.Logs,
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory %s was not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("Path %s is not a directory", dir)
		}
	}
}

func TestConfigFile(t *testing.T) {
	configPath, err := ConfigFile()
	if err != nil {
		t.Fatalf("ConfigFile() error = %v", err)
	}

	if configPath == "" {
		t.Error("Config path is empty")
	}
	if !filepath.IsAbs(configPath) {
		t.Error("Config path is not absolute")
	}
	if filepath.Base(configPath) != "config.yaml" {
		t.Errorf("Expected config file name 'config.yaml', got %s", filepath.Base(configPath))
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "magellai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create default config
	err = CreateDefaultConfig()
	if err != nil {
		t.Fatalf("CreateDefaultConfig() error = %v", err)
	}

	// Verify config file exists
	configPath, _ := ConfigFile()
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Config file was not created: %v", err)
	}
	if info.IsDir() {
		t.Error("Config path is a directory, not a file")
	}

	// Read and verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if len(content) == 0 {
		t.Error("Config file is empty")
	}

	// Test that CreateDefaultConfig doesn't overwrite existing file
	err = CreateDefaultConfig()
	if err != nil {
		t.Error("CreateDefaultConfig() should not error when config already exists")
	}
}

func TestProjectConfigFile(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "magellai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directory structure
	projectDir := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(projectDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Create a project config file
	projectConfig := filepath.Join(projectDir, ".magellai.yaml")
	err = os.WriteFile(projectConfig, []byte("test: config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create project config: %v", err)
	}

	// Test finding config from project root
	originalDir, _ := os.Getwd()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	configPath, err := ProjectConfigFile()
	if err != nil {
		t.Fatalf("ProjectConfigFile() error = %v", err)
	}
	// Resolve symlinks for comparison
	resolvedProjectConfig, _ := filepath.EvalSymlinks(projectConfig)
	resolvedConfigPath, _ := filepath.EvalSymlinks(configPath)
	if resolvedConfigPath != resolvedProjectConfig {
		t.Errorf("Expected %s, got %s", resolvedProjectConfig, resolvedConfigPath)
	}

	// Test finding config from subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change to subdirectory: %v", err)
	}
	configPath, err = ProjectConfigFile()
	if err != nil {
		t.Fatalf("ProjectConfigFile() error = %v", err)
	}
	resolvedConfigPath, _ = filepath.EvalSymlinks(configPath)
	if resolvedConfigPath != resolvedProjectConfig {
		t.Errorf("Expected %s, got %s", resolvedProjectConfig, resolvedConfigPath)
	}

	// Test when no project config exists
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	configPath, err = ProjectConfigFile()
	if err != nil {
		t.Fatalf("ProjectConfigFile() error = %v", err)
	}
	if configPath != "" {
		t.Errorf("Expected empty path when no config exists, got %s", configPath)
	}
}

func TestSystemConfigFile(t *testing.T) {
	systemConfig := SystemConfigFile()
	if systemConfig != "/etc/magellai/config.yaml" {
		t.Errorf("Expected /etc/magellai/config.yaml, got %s", systemConfig)
	}
}