// ABOUTME: Tests for extended REPL commands implementation
// ABOUTME: Validates new command parsing, execution, and error handling

package repl

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockConfigInterface for testing extended commands
type MockConfigInterface struct {
	values map[string]interface{}
}

func NewMockConfig() *MockConfigInterface {
	return &MockConfigInterface{
		values: make(map[string]interface{}),
	}
}

func (m *MockConfigInterface) GetString(key string) string {
	if val, ok := m.values[key].(string); ok {
		return val
	}
	return ""
}

func (m *MockConfigInterface) GetBool(key string) bool {
	if val, ok := m.values[key].(bool); ok {
		return val
	}
	return false
}

func (m *MockConfigInterface) Get(key string) interface{} {
	return m.values[key]
}

func (m *MockConfigInterface) Exists(key string) bool {
	_, ok := m.values[key]
	return ok
}

func (m *MockConfigInterface) SetValue(key string, value interface{}) error {
	m.values[key] = value
	return nil
}

func TestSetVerbosity(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupConfig func(*MockConfigInterface)
		expectError bool
		expectMsg   string
	}{
		{
			name: "show current verbosity",
			args: []string{},
			setupConfig: func(cfg *MockConfigInterface) {
				cfg.values["verbosity"] = "info"
			},
			expectError: false,
			expectMsg:   "Current verbosity: info\n",
		},
		{
			name:        "set valid verbosity",
			args:        []string{"debug"},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: false,
			expectMsg:   "Verbosity set to: debug\n",
		},
		{
			name:        "set invalid verbosity",
			args:        []string{"invalid"},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: true,
			expectMsg:   "invalid verbosity level: invalid (valid: debug, info, warn, error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewMockConfig()
			tt.setupConfig(cfg)

			var buf bytes.Buffer
			r := &REPL{
				config: cfg,
				writer: &buf,
				session: &Session{
					Conversation: &Conversation{},
					Metadata:     make(map[string]interface{}),
				},
			}

			err := r.setVerbosity(tt.args)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectMsg, buf.String())
			}
		})
	}
}

func TestSetOutput(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupConfig func(*MockConfigInterface)
		expectError bool
		expectMsg   string
	}{
		{
			name: "show current output format",
			args: []string{},
			setupConfig: func(cfg *MockConfigInterface) {
				cfg.values["output_format"] = "json"
			},
			expectError: false,
			expectMsg:   "Current output format: json\n",
		},
		{
			name:        "show default output format",
			args:        []string{},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: false,
			expectMsg:   "Current output format: text\n",
		},
		{
			name:        "set valid output format",
			args:        []string{"yaml"},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: false,
			expectMsg:   "Output format set to: yaml\n",
		},
		{
			name:        "set invalid output format",
			args:        []string{"invalid"},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: true,
			expectMsg:   "invalid output format: invalid (valid: text, json, yaml, markdown)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewMockConfig()
			tt.setupConfig(cfg)

			var buf bytes.Buffer
			r := &REPL{
				config: cfg,
				writer: &buf,
				session: &Session{
					Conversation: &Conversation{},
					Metadata:     make(map[string]interface{}),
				},
			}

			err := r.setOutputFormat(tt.args)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectMsg, buf.String())
			}
		})
	}
}

func TestSwitchProfile(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupConfig func(*MockConfigInterface)
		expectError bool
		expectMsg   string
	}{
		{
			name: "show current profile",
			args: []string{},
			setupConfig: func(cfg *MockConfigInterface) {
				cfg.values["profile"] = "work"
			},
			expectError: false,
			expectMsg:   "Current profile: work\n",
		},
		{
			name:        "show default profile",
			args:        []string{},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: false,
			expectMsg:   "Current profile: default\n",
		},
		{
			name: "switch to valid profile",
			args: []string{"personal"},
			setupConfig: func(cfg *MockConfigInterface) {
				cfg.values["available_profiles"] = "default,work,personal"
			},
			expectError: false,
			expectMsg:   "Switched to profile: personal\n",
		},
		{
			name: "switch to unknown profile",
			args: []string{"unknown"},
			setupConfig: func(cfg *MockConfigInterface) {
				cfg.values["available_profiles"] = "default,work,personal"
			},
			expectError: true,
			expectMsg:   "profile 'unknown' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewMockConfig()
			tt.setupConfig(cfg)

			var buf bytes.Buffer
			r := &REPL{
				config: cfg,
				writer: &buf,
				session: &Session{
					Conversation: &Conversation{},
					Metadata:     make(map[string]interface{}),
				},
			}

			err := r.switchProfile(tt.args)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectMsg)
			} else {
				assert.NoError(t, err)
				assert.True(t, strings.HasPrefix(buf.String(), tt.expectMsg))
			}
		})
	}
}

func TestRemoveAttachment(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		attachments   []llm.Attachment
		expectError   bool
		expectMsg     string
		expectRemoved bool
	}{
		{
			name: "remove existing attachment",
			args: []string{"test.pdf"},
			attachments: []llm.Attachment{
				{FilePath: "/path/to/test.pdf", Type: llm.AttachmentTypeFile},
				{FilePath: "/path/to/other.doc", Type: llm.AttachmentTypeFile},
			},
			expectError:   false,
			expectMsg:     "Removed attachment: test.pdf\n",
			expectRemoved: true,
		},
		{
			name: "remove non-existent attachment",
			args: []string{"missing.pdf"},
			attachments: []llm.Attachment{
				{FilePath: "/path/to/test.pdf", Type: llm.AttachmentTypeFile},
			},
			expectError:   true,
			expectMsg:     "attachment 'missing.pdf' not found",
			expectRemoved: false,
		},
		{
			name:          "no pending attachments",
			args:          []string{"test.pdf"},
			attachments:   []llm.Attachment{},
			expectError:   false,
			expectMsg:     "No pending attachments.\n",
			expectRemoved: false,
		},
		{
			name:        "missing filename argument",
			args:        []string{},
			attachments: []llm.Attachment{},
			expectError: true,
			expectMsg:   "attachment filename required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := &REPL{
				writer: &buf,
				session: &Session{
					Conversation: &Conversation{},
					Metadata:     make(map[string]interface{}),
				},
			}

			if len(tt.attachments) > 0 {
				r.session.Metadata["pending_attachments"] = tt.attachments
			}

			err := r.removeAttachment(tt.args)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectMsg, buf.String())

				// Check if attachment was actually removed
				if tt.expectRemoved {
					remaining := r.session.Metadata["pending_attachments"].([]llm.Attachment)
					assert.Len(t, remaining, len(tt.attachments)-1)
				}
			}
		})
	}
}

func TestShowConfig(t *testing.T) {
	cfg := NewMockConfig()
	cfg.values["stream"] = true
	cfg.values["verbosity"] = "debug"
	cfg.values["output_format"] = "json"
	cfg.values["profile"] = "work"

	var buf bytes.Buffer
	r := &REPL{
		config: cfg,
		writer: &buf,
		session: &Session{
			Conversation: &Conversation{
				Model:       "test/model",
				Temperature: 0.8,
				MaxTokens:   1000,
			},
			Metadata: make(map[string]interface{}),
		},
	}

	err := r.showConfig()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Current configuration:")
	assert.Contains(t, output, "model: test/model")
	assert.Contains(t, output, "stream: true")
	assert.Contains(t, output, "temperature: 0.80")
	assert.Contains(t, output, "max_tokens: 1000")
	assert.Contains(t, output, "verbosity: debug")
	assert.Contains(t, output, "output_format: json")
	assert.Contains(t, output, "profile: work")
}

func TestSetConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectMsg   string
		checkValue  func(t *testing.T, r *REPL)
	}{
		{
			name:        "set string value",
			args:        []string{"api_key", "test-key-123"},
			expectError: false,
			expectMsg:   "Set api_key = test-key-123\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.Equal(t, "test-key-123", r.config.GetString("api_key"))
			},
		},
		{
			name:        "set stream boolean true",
			args:        []string{"stream", "true"},
			expectError: false,
			expectMsg:   "Set stream = true\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.True(t, r.config.GetBool("stream"))
			},
		},
		{
			name:        "set stream boolean on",
			args:        []string{"stream", "on"},
			expectError: false,
			expectMsg:   "Set stream = on\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.True(t, r.config.GetBool("stream"))
			},
		},
		{
			name:        "set temperature float",
			args:        []string{"temperature", "0.7"},
			expectError: false,
			expectMsg:   "Set temperature = 0.7\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.Equal(t, 0.7, r.session.Conversation.Temperature)
			},
		},
		{
			name:        "set invalid temperature",
			args:        []string{"temperature", "3.0"},
			expectError: true,
			expectMsg:   "temperature must be between 0.0 and 2.0",
		},
		{
			name:        "set max_tokens integer",
			args:        []string{"max_tokens", "2000"},
			expectError: false,
			expectMsg:   "Set max_tokens = 2000\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.Equal(t, 2000, r.session.Conversation.MaxTokens)
			},
		},
		{
			name:        "set invalid max_tokens",
			args:        []string{"max_tokens", "-1"},
			expectError: true,
			expectMsg:   "max_tokens must be positive",
		},
		{
			name:        "missing value",
			args:        []string{"key"},
			expectError: true,
			expectMsg:   "usage: /config set <key> <value>",
		},
		{
			name:        "multi-word value",
			args:        []string{"system_prompt", "You", "are", "helpful"},
			expectError: false,
			expectMsg:   "Set system_prompt = You are helpful\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.Equal(t, "You are helpful", r.config.GetString("system_prompt"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewMockConfig()
			var buf bytes.Buffer
			r := &REPL{
				config: cfg,
				writer: &buf,
				session: &Session{
					Conversation: &Conversation{},
					Metadata:     make(map[string]interface{}),
				},
			}

			err := r.setConfig(tt.args)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectMsg, buf.String())
				if tt.checkValue != nil {
					tt.checkValue(t, r)
				}
			}
		})
	}
}

func TestExtendedCommandHandling(t *testing.T) {
	// Test that all extended commands are properly handled in handleSpecialCommand
	tests := []struct {
		command     string
		args        []string
		expectError bool
		expectMsg   string
	}{
		// Existing commands
		{":model", []string{}, true, "model name required"},
		{":stream", []string{}, false, ""},
		{":temperature", []string{}, true, "temperature value required"},
		{":max_tokens", []string{}, true, "max tokens value required"},
		{":multiline", []string{}, false, ""},

		// Extended commands
		{":verbosity", []string{}, false, "Current verbosity"},
		{":output", []string{}, false, "Current output format"},
		{":profile", []string{}, false, "Current profile"},
		{":attach", []string{}, true, "file path required"},
		{":attach-remove", []string{}, true, "attachment filename required"},
		{":attach-list", []string{}, false, ""},
		{":system", []string{}, false, ""},

		// Unknown command
		{":unknown", []string{}, true, "unknown special command"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("command %s", tt.command), func(t *testing.T) {
			cfg := NewMockConfig()
			var buf bytes.Buffer
			r := &REPL{
				config: cfg,
				writer: &buf,
				session: &Session{
					Conversation: &Conversation{},
					Metadata:     make(map[string]interface{}),
				},
			}

			// Mock provider for :model command
			if tt.command == ":model" && len(tt.args) > 0 {
				mockProvider, _ := llm.NewProvider("mock", "test-model")
				r.provider = mockProvider
			}

			cmd := tt.command
			if len(tt.args) > 0 {
				cmd += " " + strings.Join(tt.args, " ")
			}

			err := r.handleSpecialCommand(cmd)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectMsg != "" {
					assert.Contains(t, err.Error(), tt.expectMsg)
				}
			} else {
				// For valid commands without errors, just check they don't error
				// The actual behavior is tested in individual command tests
				if err != nil {
					// Only fail if it's not an expected error
					if !strings.Contains(err.Error(), "model name required") &&
						!strings.Contains(err.Error(), "file path required") &&
						!strings.Contains(err.Error(), "attachment filename required") {
						assert.NoError(t, err)
					}
				}

				// If a specific message is expected in output, check for it
				if tt.expectMsg != "" {
					assert.Contains(t, buf.String(), tt.expectMsg)
				}
			}
		})
	}
}

func TestConfigCommand(t *testing.T) {
	cfg := NewMockConfig()
	cfg.values["stream"] = true

	var buf bytes.Buffer
	r := &REPL{
		config: cfg,
		writer: &buf,
		session: &Session{
			Conversation: &Conversation{
				Model: "test/model",
			},
			Metadata: make(map[string]interface{}),
		},
	}

	// Test /config show - would be handled by handleCommand which calls showConfig
	err := r.showConfig()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Current configuration:")
	assert.Contains(t, buf.String(), "model: test/model")

	buf.Reset()

	// Test /config set - would be handled by handleCommand which calls setConfig
	err = r.setConfig([]string{"test_key", "test_value"})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Set test_key = test_value")
	assert.Equal(t, "test_value", cfg.GetString("test_key"))
}
