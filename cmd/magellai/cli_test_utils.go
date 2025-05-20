// ABOUTME: Test utilities for CLI integration tests
// ABOUTME: Provides common functionality for testing the CLI application

//go:build integration
// +build integration

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// StorageType represents the storage backend type for tests
type StorageType string

const (
	// StorageTypeFilesystem represents the filesystem storage backend
	StorageTypeFilesystem StorageType = "filesystem"
	// StorageTypeSQLite represents the SQLite storage backend
	StorageTypeSQLite StorageType = "sqlite"
)

// TestEnv represents the test environment for CLI tests
type TestEnv struct {
	// TempDir is the temporary directory for test data
	TempDir string
	// ConfigPath is the path to the test config file
	ConfigPath string
	// StorageDir is the storage directory for the test
	StorageDir string
	// SQLiteDBPath is the path to the SQLite database (if applicable)
	SQLiteDBPath string
	// StorageType is the type of storage backend to use
	StorageType StorageType
	// UseMock determines whether to use mock providers or real ones
	UseMock bool
	// BinaryPath is the path to the test binary
	BinaryPath string
	// Cleanup is a function to clean up the test environment
	Cleanup func()
}

// SetupTestEnv creates a test environment for CLI tests
func SetupTestEnv(t *testing.T, storageType StorageType, useMock bool) *TestEnv {
	t.Helper()

	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "magellai-test-*")
	require.NoError(t, err)

	// Create storage directory
	storageDir := filepath.Join(tempDir, "storage")
	err = os.MkdirAll(storageDir, 0755)
	require.NoError(t, err)

	// Determine config file to use
	var configTemplate string
	if useMock {
		configTemplate = filepath.Join("testdata", "test.mock.config.yaml")
	} else {
		configTemplate = filepath.Join("testdata", "test.config.yaml")
	}

	// Create SQLite path if needed
	sqliteDBPath := filepath.Join(tempDir, "magellai-test.db")

	// Read config template
	configData, err := os.ReadFile(configTemplate)
	require.NoError(t, err)

	// Replace placeholders with actual paths
	configContent := string(configData)
	configContent = strings.ReplaceAll(configContent, "STORAGE_DIR_PLACEHOLDER", storageDir)
	configContent = strings.ReplaceAll(configContent, "SQLITE_DB_PLACEHOLDER", sqliteDBPath)

	// Write config to test directory
	configPath := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Build test binary
	binaryPath := filepath.Join(tempDir, "test-magellai")
	buildArgs := []string{"build", "-o", binaryPath}
	
	// Add build tags if using SQLite
	if storageType == StorageTypeSQLite {
		buildArgs = append(buildArgs, "-tags", "sqlite")
	}
	
	buildArgs = append(buildArgs, ".")
	cmd := exec.Command("go", buildArgs...)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build test binary: %s", string(output))

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return &TestEnv{
		TempDir:      tempDir,
		ConfigPath:   configPath,
		StorageDir:   storageDir,
		SQLiteDBPath: sqliteDBPath,
		StorageType:  storageType,
		UseMock:      useMock,
		BinaryPath:   binaryPath,
		Cleanup:      cleanup,
	}
}

// RunCommand runs a CLI command with the given arguments
func (env *TestEnv) RunCommand(args ...string) (string, error) {
	// Add config file flag to args
	configArgs := append([]string{"--config-file", env.ConfigPath}, args...)
	
	// Determine environment variables based on storage type
	cmd := exec.Command(env.BinaryPath, configArgs...)
	if env.StorageType == StorageTypeSQLite {
		cmd.Env = append(os.Environ(), 
			fmt.Sprintf("MAGELLAI_STORAGE_TYPE=sqlite"),
			fmt.Sprintf("MAGELLAI_STORAGE_CONNECTION_STRING=%s", env.SQLiteDBPath),
		)
	} else {
		cmd.Env = append(os.Environ(), 
			fmt.Sprintf("MAGELLAI_STORAGE_TYPE=filesystem"),
			fmt.Sprintf("MAGELLAI_STORAGE_DIR=%s", env.StorageDir),
		)
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// RunInteractiveCommand runs a CLI command with the given input and returns its output
func (env *TestEnv) RunInteractiveCommand(input string, args ...string) (string, error) {
	// Add config file flag to args
	configArgs := append([]string{"--config-file", env.ConfigPath}, args...)
	
	// Determine environment variables based on storage type
	cmd := exec.Command(env.BinaryPath, configArgs...)
	if env.StorageType == StorageTypeSQLite {
		cmd.Env = append(os.Environ(), 
			fmt.Sprintf("MAGELLAI_STORAGE_TYPE=sqlite"),
			fmt.Sprintf("MAGELLAI_STORAGE_CONNECTION_STRING=%s", env.SQLiteDBPath),
		)
	} else {
		cmd.Env = append(os.Environ(), 
			fmt.Sprintf("MAGELLAI_STORAGE_TYPE=filesystem"),
			fmt.Sprintf("MAGELLAI_STORAGE_DIR=%s", env.StorageDir),
		)
	}

	// Set up pipes for stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Start the command
	err = cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Write input to stdin and close
	_, err = io.WriteString(stdin, input+"\n")
	if err != nil {
		return "", fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	// Create a channel to signal command completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for the command to complete or timeout
	select {
	case err := <-done:
		if err != nil {
			return stderr.String(), fmt.Errorf("%w: %s", err, stderr.String())
		}
	case <-time.After(10 * time.Second):
		// Kill the process if it times out
		cmd.Process.Kill()
		return "", fmt.Errorf("command timed out after 10 seconds")
	}

	return stdout.String(), nil
}

// ForEachStorageType runs a test for each storage type
func ForEachStorageType(t *testing.T, useMock bool, testFunc func(t *testing.T, env *TestEnv)) {
	t.Helper()
	
	storageTypes := []StorageType{StorageTypeFilesystem, StorageTypeSQLite}
	
	for _, storageType := range storageTypes {
		storageType := storageType // Shadow variable for closure
		
		// Skip SQLite tests if running on a system without SQLite support
		if storageType == StorageTypeSQLite {
			if _, err := exec.LookPath("sqlite3"); err != nil {
				t.Logf("Skipping SQLite tests: %v", err)
				continue
			}
		}
		
		t.Run(fmt.Sprintf("Storage=%s", storageType), func(t *testing.T) {
			env := SetupTestEnv(t, storageType, useMock)
			defer env.Cleanup()
			
			testFunc(t, env)
		})
	}
}

// WithMockEnv runs a test with a mock environment
func WithMockEnv(t *testing.T, storageType StorageType, testFunc func(t *testing.T, env *TestEnv)) {
	t.Helper()
	
	env := SetupTestEnv(t, storageType, true)
	defer env.Cleanup()
	
	testFunc(t, env)
}

// WithLiveEnv runs a test with a live environment (real providers)
func WithLiveEnv(t *testing.T, storageType StorageType, testFunc func(t *testing.T, env *TestEnv)) {
	t.Helper()
	
	// Skip tests if API keys are not set
	if os.Getenv("ANTHROPIC_API_KEY") == "" && os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping live provider tests: no API keys set")
	}
	
	env := SetupTestEnv(t, storageType, false)
	defer env.Cleanup()
	
	testFunc(t, env)
}