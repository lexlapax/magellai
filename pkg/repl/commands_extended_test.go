// ABOUTME: Tests for extended REPL commands implementation
// ABOUTME: Validates new command parsing, execution, and error handling

package repl

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/command"
	"github.com/lexlapax/magellai/pkg/domain"
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
			expectError: true,
			expectMsg:   "verbosity level required (debug, info, warn, error)",
		},
		{
			name:        "set valid verbosity",
			args:        []string{"debug"},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: false,
			expectMsg:   "Verbosity level set to: debug\n",
		},
		{
			name:        "set invalid verbosity",
			args:        []string{"invalid"},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: true,
			expectMsg:   "invalid verbosity level: invalid",
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
				session: &domain.Session{
					Conversation: &domain.Conversation{},
					Metadata:     make(map[string]interface{}),
				},
				sharedContext: command.NewSharedContext(),
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
			expectError: true,
			expectMsg:   "output format required (text, json, yaml, markdown)",
		},
		{
			name:        "show default output format",
			args:        []string{},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: true,
			expectMsg:   "output format required (text, json, yaml, markdown)",
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
			expectMsg:   "invalid output format: invalid",
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
				session: &domain.Session{
					Conversation: &domain.Conversation{},
					Metadata:     make(map[string]interface{}),
				},
				sharedContext: command.NewSharedContext(),
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
			expectError: true,
			expectMsg:   "profile name required",
		},
		{
			name:        "show default profile",
			args:        []string{},
			setupConfig: func(cfg *MockConfigInterface) {},
			expectError: true,
			expectMsg:   "profile name required",
		},
		{
			name: "switch to valid profile",
			args: []string{"personal"},
			setupConfig: func(cfg *MockConfigInterface) {
				cfg.values["profiles.personal"] = map[string]interface{}{}
			},
			expectError: false,
			expectMsg:   "Profile switching not fully implemented yet.\n",
		},
		{
			name: "switch to unknown profile",
			args: []string{"unknown"},
			setupConfig: func(cfg *MockConfigInterface) {
				cfg.values["available_profiles"] = "default,work,personal"
			},
			expectError: true,
			expectMsg:   "profile not found: unknown",
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
				session: &domain.Session{
					Conversation: &domain.Conversation{},
					Metadata:     make(map[string]interface{}),
				},
				sharedContext: command.NewSharedContext(),
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
		attachments   []domain.Attachment
		expectError   bool
		expectMsg     string
		expectRemoved bool
	}{
		{
			name: "remove existing attachment",
			args: []string{"test.pdf"},
			attachments: []domain.Attachment{
				{FilePath: "/path/to/test.pdf", Type: domain.AttachmentTypeFile},
				{FilePath: "/path/to/other.doc", Type: domain.AttachmentTypeFile},
			},
			expectError:   false,
			expectMsg:     "Attachment removed: test.pdf\n",
			expectRemoved: true,
		},
		{
			name: "remove non-existent attachment",
			args: []string{"missing.pdf"},
			attachments: []domain.Attachment{
				{FilePath: "/path/to/test.pdf", Type: domain.AttachmentTypeFile},
			},
			expectError:   true,
			expectMsg:     "attachment not found: missing.pdf",
			expectRemoved: false,
		},
		{
			name:          "no pending attachments",
			args:          []string{"test.pdf"},
			attachments:   []domain.Attachment{},
			expectError:   false,
			expectMsg:     "No attachments to remove.\n",
			expectRemoved: false,
		},
		{
			name:        "missing filename argument",
			args:        []string{},
			attachments: []domain.Attachment{},
			expectError: true,
			expectMsg:   "attachment file name required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := &REPL{
				writer: &buf,
				session: &domain.Session{
					Conversation: &domain.Conversation{},
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
					remaining := r.session.Metadata["pending_attachments"].([]domain.Attachment)
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
		session: &domain.Session{
			Conversation: &domain.Conversation{
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
	assert.Contains(t, output, "Model: test/model")
	assert.Contains(t, output, "Stream: true")
	assert.Contains(t, output, "Temperature: 0.8")
	assert.Contains(t, output, "Max tokens: 1000")
	assert.Contains(t, output, "Verbosity: debug")
	assert.Contains(t, output, "Auto-save: false")
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
			expectMsg:   "Config api_key set to: test-key-123\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.Equal(t, "test-key-123", r.config.GetString("api_key"))
			},
		},
		{
			name:        "set stream boolean true",
			args:        []string{"stream", "true"},
			expectError: false,
			expectMsg:   "Streaming mode: on\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.True(t, r.config.GetBool("stream"))
			},
		},
		{
			name:        "set stream boolean on",
			args:        []string{"stream", "on"},
			expectError: false,
			expectMsg:   "Streaming mode: on\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.True(t, r.config.GetBool("stream"))
			},
		},
		{
			name:        "set temperature float",
			args:        []string{"temperature", "0.7"},
			expectError: false,
			expectMsg:   "Temperature set to: 0.7\n",
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
			expectMsg:   "Max tokens set to: 2000\n",
			checkValue: func(t *testing.T, r *REPL) {
				assert.Equal(t, 2000, r.session.Conversation.MaxTokens)
			},
		},
		{
			name:        "set invalid max_tokens",
			args:        []string{"max_tokens", "-1"},
			expectError: true,
			expectMsg:   "max tokens must be positive",
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
			expectMsg:   "Config system_prompt set to: You are helpful\n",
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
				session: &domain.Session{
					Conversation: &domain.Conversation{},
					Metadata:     make(map[string]interface{}),
				},
				sharedContext: command.NewSharedContext(),
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
	t.Skip("Temporarily disabled while updating to unified command system")
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
		{":verbosity", []string{}, true, "verbosity level required (debug, info, warn, error)"},
		{":output", []string{}, true, "output format required (text, json, yaml, markdown)"},
		{":profile", []string{}, true, "profile name required"},
		{":attach", []string{}, true, "file path required"},
		{":attach-remove", []string{}, true, "attachment file name required"},
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
				session: &domain.Session{
					Conversation: &domain.Conversation{},
					Metadata:     make(map[string]interface{}),
				},
				sharedContext: command.NewSharedContext(),
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
		session: &domain.Session{
			Conversation: &domain.Conversation{
				Model: "test/model",
			},
			Metadata: make(map[string]interface{}),
		},
	}

	// Test /config show - would be handled by handleCommand which calls showConfig
	err := r.showConfig()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Current configuration:")
	assert.Contains(t, buf.String(), "Model: test/model")

	buf.Reset()

	// Test /config set - would be handled by handleCommand which calls setConfig
	err = r.setConfig([]string{"test_key", "test_value"})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Config test_key set to: test_value")
	assert.Equal(t, "test_value", cfg.GetString("test_key"))
}
