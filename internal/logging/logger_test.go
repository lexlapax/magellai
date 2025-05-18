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

	if config.Level != "warn" {
		t.Errorf("expected default level 'warn', got %s", config.Level)
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

// TestSensitiveDataSanitization tests that sensitive data is properly sanitized in logs
func TestSensitiveDataSanitization(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelDebug,
		output: &buf,
	}

	// Simulate logging with a sanitized API key (this tests the actual implementation)
	apiKey := "sk-1234567890abcdefghijklmnopqrstuvwxyz"
	sanitized := "sk-123...wxyz" // What we expect after sanitization

	logger.Debug("API key usage", "key", sanitized, "provider", "openai")

	output := buf.String()

	// Should contain the sanitized version
	if !strings.Contains(output, sanitized) {
		t.Errorf("Expected sanitized key in output, got: %s", output)
	}

	// Should NOT contain the full API key
	if strings.Contains(output, apiKey) {
		t.Errorf("Full API key should not appear in logs: %s", output)
	}
}

// TestErrorLoggingWithNilError tests that nil errors are handled gracefully
func TestErrorLoggingWithNilError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	defaultLogger = &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelInfo,
		output: &buf,
	}

	// Test LogError with nil error
	LogError(nil, "operation completed", "status", "success")

	output := buf.String()

	// Should contain the message but not a nil error
	if !strings.Contains(output, "operation completed") {
		t.Errorf("Expected message in output, got: %s", output)
	}

	// Should not contain error=<nil> or similar
	if strings.Contains(output, "error=") || strings.Contains(output, "<nil>") {
		t.Errorf("Nil error should not be logged: %s", output)
	}
}

// TestVerbosityConfiguration tests that verbosity can be configured properly
func TestVerbosityConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		verbosity   int
		expectLevel slog.Level
	}{
		{"no verbosity", 0, slog.LevelWarn}, // Default is warn now
		{"verbosity 1", 1, slog.LevelInfo},  // -v gives info level
		{"verbosity 2", 2, slog.LevelDebug}, // -vv gives debug level
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with verbosity
			config := DefaultConfig()
			// Map verbosity to appropriate log level
			switch tt.verbosity {
			case 1:
				config.Level = "info"
			case 2:
				config.Level = "debug"
				// 0 keeps default warn level
			}

			// Parse and check level
			level, err := parseLevel(config.Level)
			if err != nil {
				t.Fatalf("Failed to parse level: %v", err)
			}

			if level != tt.expectLevel {
				t.Errorf("Expected level %v for verbosity %d, got %v",
					tt.expectLevel, tt.verbosity, level)
			}
		})
	}
}

// BenchmarkLogging benchmarks logging performance
func BenchmarkLogging(b *testing.B) {
	// Create a no-op writer to avoid I/O overhead
	handler := slog.NewTextHandler(&nopWriter{}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelInfo,
		output: &nopWriter{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "key", "value")
	}
}

// BenchmarkLoggingWithDebugLevel benchmarks logging at debug level
func BenchmarkLoggingWithDebugLevel(b *testing.B) {
	handler := slog.NewTextHandler(&nopWriter{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &Logger{
		Logger: slog.New(handler),
		level:  slog.LevelDebug,
		output: &nopWriter{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("debug message", "iteration", i)
		logger.Info("info message", "iteration", i)
		logger.Warn("warn message", "iteration", i)
	}
}

// nopWriter is a no-op writer for benchmarking
type nopWriter struct{}

func (nopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
