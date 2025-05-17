// ABOUTME: Unit tests for the configuration package
// ABOUTME: Ensures all configuration functionality works correctly

package config

import (
	"testing"
)

func TestInit(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	if Manager == nil {
		t.Error("Manager should not be nil after Init")
	}

	if Manager.koanf == nil {
		t.Error("koanf instance should not be nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Load defaults
	if err := Manager.Load(nil); err != nil {
		t.Fatalf("Failed to load defaults: %v", err)
	}

	// Test default values
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"log.level", "info"},
		{"log.format", "text"},
		{"provider.default", "openai"},
		{"output.format", "text"},
		{"output.color", true},
		{"session.autosave", true},
	}

	for _, test := range tests {
		actual := Manager.Get(test.key)
		if actual != test.expected {
			t.Errorf("Expected %s to be %v, got %v", test.key, test.expected, actual)
		}
	}
}

func TestTypeAccessors(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	if err := Manager.Load(nil); err != nil {
		t.Fatalf("Failed to load defaults: %v", err)
	}

	// Test string accessor
	level := Manager.GetString("log.level")
	if level != "info" {
		t.Errorf("Expected log.level to be 'info', got %s", level)
	}

	// Test bool accessor
	color := Manager.GetBool("output.color")
	if !color {
		t.Error("Expected output.color to be true")
	}

	// Test existence check
	if !Manager.Exists("log.level") {
		t.Error("Expected log.level to exist")
	}

	if Manager.Exists("nonexistent.key") {
		t.Error("Expected nonexistent.key to not exist")
	}
}

func TestSetValue(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	if err := Manager.Load(nil); err != nil {
		t.Fatalf("Failed to load defaults: %v", err)
	}

	// Set a new value
	err = Manager.SetValue("test.key", "test value")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Verify the value was set
	value := Manager.GetString("test.key")
	if value != "test value" {
		t.Errorf("Expected test.key to be 'test value', got %s", value)
	}
}

func TestParseProviderModel(t *testing.T) {
	tests := []struct {
		input    string
		provider string
		model    string
	}{
		{"openai/gpt-4", "openai", "gpt-4"},
		{"anthropic/claude-3", "anthropic", "claude-3"},
		{"gpt-4", "", "gpt-4"},
		{"", "", ""},
	}

	for _, test := range tests {
		pair := ParseProviderModel(test.input)
		if pair.Provider != test.provider {
			t.Errorf("For input %s, expected provider %s, got %s",
				test.input, test.provider, pair.Provider)
		}
		if pair.Model != test.model {
			t.Errorf("For input %s, expected model %s, got %s",
				test.input, test.model, pair.Model)
		}
	}
}

func TestCommandLineOverrides(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Create test command-line overrides as a map
	cmdlineOverrides := map[string]interface{}{
		"profile":   "test",
		"verbosity": 2,
	}

	// Load with command-line overrides
	err = Manager.Load(cmdlineOverrides)
	if err != nil {
		t.Fatalf("Failed to load config with command-line overrides: %v", err)
	}

	// Check if override values are accessible
	profile := Manager.GetString("profile")
	if profile != "test" {
		t.Errorf("Expected profile to be 'test', got %s", profile)
	}

	verbosity := Manager.GetInt("verbosity")
	if verbosity != 2 {
		t.Errorf("Expected verbosity to be 2, got %d", verbosity)
	}
}
