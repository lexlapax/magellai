// ABOUTME: Unit tests for the logging infrastructure
// ABOUTME: Tests configuration, initialization, and logging functionality
package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != "info" {
		t.Errorf("expected default level 'info', got %s", config.Level)
	}
	if config.Format != "text" {
		t.Errorf("expected default format 'text', got %s", config.Format)
	}
	if config.OutputPath != "stderr" {
		t.Errorf("expected default output 'stderr', got %s", config.OutputPath)
	}
	if config.AddSource != false {
		t.Error("expected AddSource to be false by default")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
		wantErr  bool
	}{
		{"debug", slog.LevelDebug, false},
		{"info", slog.LevelInfo, false},
		{"warn", slog.LevelWarn, false},
		{"warning", slog.LevelWarn, false},
		{"error", slog.LevelError, false},
		{"invalid", slog.LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := parseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLevel(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && level != tt.expected {
				t.Errorf("parseLevel(%s) = %v, want %v", tt.input, level, tt.expected)
			}
		})
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Reset the default logger for testing
	defaultLogger = nil
	once = sync.Once{}

	config := LogConfig{
		Level:      "debug",
		Format:     "text",
		OutputPath: "stdout",
		AddSource:  false,
	}

	err := Initialize(config)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}

	if logger.GetLevel() != slog.LevelDebug {
		t.Errorf("expected level debug, got %v", logger.GetLevel())
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer

	// Create a logger with JSON format
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelInfo,
		output: &buf,
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("JSON output missing message: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("JSON output missing key-value pair: %s", output)
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer

	// Create a logger with text format
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelInfo,
		output: &buf,
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("text output missing message: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("text output missing key-value pair: %s", output)
	}
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelInfo,
		output: &buf,
	}

	// Create a new logger with additional attributes
	childLogger := logger.With("component", "test")
	childLogger.Info("child message")

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Errorf("output missing component attribute: %s", output)
	}
	if !strings.Contains(output, "child message") {
		t.Errorf("output missing message: %s", output)
	}
}

func TestLogHelpers(t *testing.T) {
	var buf bytes.Buffer

	// Initialize logger with buffer output
	defaultLogger = nil
	once = sync.Once{}

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	defaultLogger = &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelDebug,
		output: &buf,
	}

	// Test helper functions
	LogDebug("debug message", "key", "value")
	LogInfo("info message", "key", "value")
	LogWarn("warn message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("output missing debug message")
	}
	if !strings.Contains(output, "info message") {
		t.Error("output missing info message")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("output missing warn message")
	}
}
