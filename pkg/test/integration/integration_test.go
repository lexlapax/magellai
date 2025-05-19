// ABOUTME: Integration tests for the main CLI application
// ABOUTME: Tests actual command execution and behavior

//go:build integration
// +build integration

package integration

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_BasicE2E tests basic end-to-end functionality
func TestIntegration_BasicE2E(t *testing.T) {
	// Test version command directly
	cmd := exec.Command("go", "run", "./cmd/magellai", "version")
	cmd.Dir = "../../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed with error: %v\nOutput: %s", err, string(output))
	}
	assert.Contains(t, string(output), "magellai version")

	// Test help
	cmd = exec.Command("go", "run", "./cmd/magellai", "--help")
	cmd.Dir = "../../.."
	output, err = cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Usage:")
}

// TestMain_E2E tests the actual main function
// This is closer to a true integration test
func TestMain_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Test version command through built binary
	cmd := exec.Command("go", "build", "-o", "test-magellai", "../../../cmd/magellai")
	require.NoError(t, cmd.Run())
	defer func() {
		_ = exec.Command("rm", "test-magellai").Run()
	}()

	// Run version command
	cmd = exec.Command("./test-magellai", "version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "magellai version")

	// Run help command
	cmd = exec.Command("./test-magellai", "--help")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Usage:")
}