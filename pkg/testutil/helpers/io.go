// ABOUTME: I/O helper functions for testing
// ABOUTME: Provides utilities for handling I/O in tests

package helpers

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// CaptureOutput captures stdout/stderr during test execution
func CaptureOutput(t *testing.T, f func()) (string, string) {
	t.Helper()

	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	// Redirect stdout and stderr
	os.Stdout = wOut
	os.Stderr = wErr

	// Restore on cleanup
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Run the function
	f()

	// Close writers
	wOut.Close()
	wErr.Close()

	// Read output
	var bufOut, bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T, prefix string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// CreateTempFile creates a temporary file with content
func CreateTempFile(t *testing.T, dir, pattern string, content string) string {
	t.Helper()

	if dir == "" {
		dir = CreateTempDir(t, "test")
	}

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	file.Close()
	return file.Name()
}

// ReadFile reads a file and returns its content
func ReadFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	return string(content)
}

// WriteFile writes content to a file
func WriteFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// FileExists checks if a file exists
func FileExists(t *testing.T, path string) bool {
	t.Helper()

	_, err := os.Stat(path)
	return err == nil
}

// AssertFileContains checks if a file contains expected content
func AssertFileContains(t *testing.T, path, expected string) {
	t.Helper()

	content := ReadFile(t, path)
	if !bytes.Contains([]byte(content), []byte(expected)) {
		t.Errorf("File %s does not contain expected content %q", path, expected)
	}
}

// MockStdin mocks stdin with provided input
func MockStdin(t *testing.T, input string) func() {
	t.Helper()

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	os.Stdin = r

	// Write input to pipe
	go func() {
		defer w.Close()
		_, _ = w.WriteString(input)
	}()

	// Return cleanup function
	return func() {
		os.Stdin = oldStdin
		r.Close()
	}
}

// CreateMockReadWriter creates a mock io.ReadWriter for testing
type MockReadWriter struct {
	*bytes.Buffer
}

// NewMockReadWriter creates a new mock read/writer
func NewMockReadWriter() *MockReadWriter {
	return &MockReadWriter{
		Buffer: &bytes.Buffer{},
	}
}

// Read implements io.Reader
func (m *MockReadWriter) Read(p []byte) (n int, err error) {
	return m.Buffer.Read(p)
}

// Write implements io.Writer
func (m *MockReadWriter) Write(p []byte) (n int, err error) {
	return m.Buffer.Write(p)
}

// Close implements io.Closer
func (m *MockReadWriter) Close() error {
	return nil
}
