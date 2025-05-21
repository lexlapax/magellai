// ABOUTME: Integration test for config show command
// ABOUTME: Verifies CLI displays all runtime configurations

//go:build cmdline
// +build cmdline

package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestConfigShowCLI(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "test-magellai", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer func() {
		_ = exec.Command("rm", "test-magellai").Run()
	}()

	// Test config show command
	cmd = exec.Command("./test-magellai", "config", "show")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, out.String())
	}

	output := out.String()

	// Verify it shows runtime configurations
	expectedKeys := []string{
		"provider",
		"model",
		"log",
		"session",
		"repl",
		"profiles",
		"aliases",
		"output",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("Expected configuration key '%s' not found in output", key)
		}
	}

	// Verify it shows default values
	if !strings.Contains(output, "provider.default: openai") {
		t.Error("Default provider value not shown")
	}

	if !strings.Contains(output, "log.level:") {
		t.Error("Log level not shown")
	}

	t.Logf("Config show output (first 500 chars):\n%s", output[:min(500, len(output))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
