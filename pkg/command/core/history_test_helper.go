package core

import (
	"os"
	"path/filepath"
)

// setupTestConfig sets the HOME environment variable to use a test directory
func setupTestConfig(tmpDir string) (cleanup func()) {
	// Save old HOME
	oldHome := os.Getenv("HOME")

	// Set new HOME to our test directory
	os.Setenv("HOME", tmpDir)

	// Create the .config/magellai structure
	configDir := filepath.Join(tmpDir, ".config", "magellai")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(configDir, "sessions"), 0755); err != nil {
		panic(err)
	}

	// Return cleanup function
	return func() {
		os.Setenv("HOME", oldHome)
	}
}
