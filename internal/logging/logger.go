// ABOUTME: Logging infrastructure for Magellai using slog
// ABOUTME: Provides configurable logging with support for different output formats and levels
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	mu     sync.RWMutex
	level  slog.Level
	output io.Writer
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// LogConfig represents logging configuration
type LogConfig struct {
	Level      string // debug, info, warn, error
	Format     string // text, json
	OutputPath string // stdout, stderr, or file path
	AddSource  bool   // whether to add source code location
}

// DefaultConfig returns default logging configuration
func DefaultConfig() LogConfig {
	return LogConfig{
		Level:      "info",
		Format:     "text",
		OutputPath: "stderr",
		AddSource:  false,
	}
}

// Initialize sets up the global logger with the provided configuration
func Initialize(config LogConfig) error {
	level, err := parseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	output, err := getOutput(config.OutputPath)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{Key: "time", Value: slog.StringValue(a.Value.Time().Format(time.RFC3339))}
			}
			return a
		},
	}

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	logger := &Logger{
		Logger: slog.New(handler),
		level:  level,
		output: output,
	}

	once.Do(func() {
		defaultLogger = logger
		slog.SetDefault(logger.Logger)
	})

	return nil
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Initialize with default config if not already initialized
		_ = Initialize(DefaultConfig())
	}
	return defaultLogger
}

// With returns a new Logger with the given attributes
func (l *Logger) With(attrs ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(attrs...),
		level:  l.level,
		output: l.output,
	}
}

// WithContext returns a new Logger with the given context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// slog.Logger doesn't have WithContext, so we'll just return the same logger
	// This could be extended to add context-specific attributes in the future
	return l
}

// SetLevel dynamically changes the logging level
func (l *Logger) SetLevel(level slog.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() slog.Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown level: %s", level)
	}
}

// getOutput returns the appropriate io.Writer for the output path
func getOutput(path string) (io.Writer, error) {
	switch path {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
}

// LogError logs an error with additional context
func LogError(err error, msg string, args ...any) {
	logger := GetLogger()
	if _, file, line, ok := runtime.Caller(1); ok {
		args = append(args, "file", file, "line", line)
	}
	if err != nil {
		args = append(args, "error", err.Error())
	}
	logger.Error(msg, args...)
}

// LogDebug logs a debug message
func LogDebug(msg string, args ...any) {
	GetLogger().Debug(msg, args...)
}

// LogInfo logs an info message
func LogInfo(msg string, args ...any) {
	GetLogger().Info(msg, args...)
}

// LogWarn logs a warning message
func LogWarn(msg string, args ...any) {
	GetLogger().Warn(msg, args...)
}

// SetLogLevel sets the global log level by re-initializing the logger
func SetLogLevel(level string) error {
	// Get current configuration
	config := LogConfig{
		Level:      level,
		Format:     "text",   // Preserve current format
		OutputPath: "stderr", // Preserve current output
		AddSource:  false,
	}

	// Re-initialize with new level
	return Initialize(config)
}
