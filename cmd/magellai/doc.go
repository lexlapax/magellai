// ABOUTME: Main entry point for the Magellai CLI application
// ABOUTME: Handles command-line parsing, execution, and global flags

/*
Package main implements the Magellai command-line application.

This package serves as the main entry point for the Magellai CLI tool,
providing command-line parsing, execution flow, and integration of all
components. It uses Kong for command-line parsing and configuration handling.

Key Components:
  - CLI Structure: Defines the command hierarchy and global flags
  - Command Registration: Discovers and registers all available commands
  - Configuration Management: Handles loading and validation of configuration
  - Command Execution: Routes commands to appropriate handlers
  - Error Handling: Global error handling and reporting

The CLI supports two primary modes of operation:
  - Ask Mode: One-shot queries for quick interactions
  - Chat Mode: Interactive REPL for ongoing conversations

Environment Variables:
  - MAGELLAI_CONFIG_DIR: Directory containing configuration files
  - MAGELLAI_CONFIG_FILE: Path to the main configuration file
  - MAGELLAI_LOG_LEVEL: Logging level (debug, info, warn, error)
  - Various provider API keys (ANTHROPIC_API_KEY, OPENAI_API_KEY, etc.)

The application follows a library-first design, where all core functionality
is implemented in package libraries, with the main package providing the
command-line interface and orchestration.
*/
package main
