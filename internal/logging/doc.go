// ABOUTME: Internal logging package providing structured logging
// ABOUTME: Offers consistent logging interface and level-based filtering

/*
Package logging provides a standardized logging framework for internal use.

This package implements structured logging with consistent field naming and
log level filtering for all components of the application. It centralizes
logging configuration and initialization for consistent behavior.

Key Components:
  - Logger: Main logging interface with level-based methods
  - LogLevel: Enumeration of available log levels
  - Helpers: Utility functions for common logging patterns
  - Configuration: Dynamic log level configuration

The package supports several log levels:
  - Debug: Detailed information for troubleshooting
  - Info: General information about application progress
  - Warn: Potentially problematic situations
  - Error: Error conditions that should be addressed

Usage:

	// Initialize logger with default settings
	logger := logging.NewLogger()

	// Set log level
	logger.SetLevel(logging.LevelDebug)

	// Log with structured fields
	logger.Info("User logged in", "user_id", userId, "source", "web")

	// Log errors with context
	logger.Error("Failed to connect to database", "error", err, "retry", true)

This package is internal to the application and should not be imported by
external code. Public packages should use their own logging abstraction.
*/
package logging
