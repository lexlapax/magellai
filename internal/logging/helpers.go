// ABOUTME: Helper functions for common logging patterns
// ABOUTME: Provides structured logging helpers for components like LLM providers, sessions, etc.

package logging

import (
	"fmt"
	"time"
)

// Provider logging helpers

// LogProviderOperation logs a provider operation with standard fields
func LogProviderOperation(logger *Logger, operation string, provider string, model string, level string, message string, args ...any) {
	standardFields := []any{
		"operation", operation,
		"provider", provider,
		"model", model,
	}

	// Combine standard fields with the custom ones
	combinedArgs := append(standardFields, args...)

	switch level {
	case "debug":
		logger.Debug(message, combinedArgs...)
	case "info":
		logger.Info(message, combinedArgs...)
	case "warn":
		logger.Warn(message, combinedArgs...)
	case "error":
		logger.Error(message, combinedArgs...)
	default:
		logger.Info(message, combinedArgs...)
	}
}

// LogProviderError logs a provider error with standard fields
func LogProviderError(logger *Logger, operation string, provider string, model string, err error, args ...any) {
	standardFields := []any{
		"operation", operation,
		"provider", provider,
		"model", model,
		"error", err,
	}

	combinedArgs := append(standardFields, args...)
	logger.Error(fmt.Sprintf("Provider error during %s", operation), combinedArgs...)
}

// LogProviderFallback logs a provider fallback attempt
func LogProviderFallback(logger *Logger, operation string, primaryProvider string, fallbackProvider string, reason string, args ...any) {
	standardFields := []any{
		"operation", operation,
		"primary_provider", primaryProvider,
		"fallback_provider", fallbackProvider,
		"reason", reason,
	}

	combinedArgs := append(standardFields, args...)
	logger.Info("Provider fallback attempt", combinedArgs...)
}

// LogContextManagement logs context management operations
func LogContextManagement(logger *Logger, operation string, provider string, model string, originalTokens int, newTokens int, args ...any) {
	standardFields := []any{
		"operation", operation,
		"provider", provider,
		"model", model,
		"original_tokens", originalTokens,
		"new_tokens", newTokens,
		"tokens_saved", originalTokens - newTokens,
	}

	combinedArgs := append(standardFields, args...)
	logger.Info("Context management", combinedArgs...)
}

// Session and Branching logging helpers

// LogSessionOperation logs session operations with standard fields
func LogSessionOperation(logger *Logger, operation string, sessionID string, level string, message string, args ...any) {
	standardFields := []any{
		"operation", operation,
		"session_id", sessionID,
		"timestamp", time.Now().Format(time.RFC3339),
	}

	combinedArgs := append(standardFields, args...)

	switch level {
	case "debug":
		logger.Debug(message, combinedArgs...)
	case "info":
		logger.Info(message, combinedArgs...)
	case "warn":
		logger.Warn(message, combinedArgs...)
	case "error":
		logger.Error(message, combinedArgs...)
	default:
		logger.Info(message, combinedArgs...)
	}
}

// LogBranchOperation logs branch operations with standard fields
func LogBranchOperation(logger *Logger, operation string, branchID string, parentID string, level string, message string, args ...any) {
	standardFields := []any{
		"operation", operation,
		"branch_id", branchID,
		"parent_id", parentID,
		"timestamp", time.Now().Format(time.RFC3339),
	}

	combinedArgs := append(standardFields, args...)

	switch level {
	case "debug":
		logger.Debug(message, combinedArgs...)
	case "info":
		logger.Info(message, combinedArgs...)
	case "warn":
		logger.Warn(message, combinedArgs...)
	case "error":
		logger.Error(message, combinedArgs...)
	default:
		logger.Info(message, combinedArgs...)
	}
}

// LogMergeOperation logs merge operations with standard fields
func LogMergeOperation(logger *Logger, operation string, sourceID string, targetID string, mergeType string, level string, message string, args ...any) {
	standardFields := []any{
		"operation", operation,
		"source_id", sourceID,
		"target_id", targetID,
		"merge_type", mergeType,
		"timestamp", time.Now().Format(time.RFC3339),
	}

	combinedArgs := append(standardFields, args...)

	switch level {
	case "debug":
		logger.Debug(message, combinedArgs...)
	case "info":
		logger.Info(message, combinedArgs...)
	case "warn":
		logger.Warn(message, combinedArgs...)
	case "error":
		logger.Error(message, combinedArgs...)
	default:
		logger.Info(message, combinedArgs...)
	}
}

// Stream and Partial Response logging helpers

// LogStreamOperation logs streaming operations with standard fields
func LogStreamOperation(logger *Logger, operation string, provider string, model string, level string, message string, args ...any) {
	standardFields := []any{
		"operation", operation,
		"provider", provider,
		"model", model,
		"timestamp", time.Now().Format(time.RFC3339),
	}

	combinedArgs := append(standardFields, args...)

	switch level {
	case "debug":
		logger.Debug(message, combinedArgs...)
	case "info":
		logger.Info(message, combinedArgs...)
	case "warn":
		logger.Warn(message, combinedArgs...)
	case "error":
		logger.Error(message, combinedArgs...)
	default:
		logger.Info(message, combinedArgs...)
	}
}

// LogStreamRecovery logs stream recovery attempts with standard fields
func LogStreamRecovery(logger *Logger, provider string, model string, contentLength int, attemptNumber int, maxAttempts int, level string, message string, args ...any) {
	standardFields := []any{
		"provider", provider,
		"model", model,
		"content_length", contentLength,
		"attempt", attemptNumber,
		"max_attempts", maxAttempts,
		"timestamp", time.Now().Format(time.RFC3339),
	}

	combinedArgs := append(standardFields, args...)

	switch level {
	case "debug":
		logger.Debug(message, combinedArgs...)
	case "info":
		logger.Info(message, combinedArgs...)
	case "warn":
		logger.Warn(message, combinedArgs...)
	case "error":
		logger.Error(message, combinedArgs...)
	default:
		logger.Info(message, combinedArgs...)
	}
}
