# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magellai is a command-line interface (CLI) tool and REPL that interacts with Large Language Models (LLMs). It operates in two primary modes:
- **`ask` mode**: One-shot queries
- **`chat` mode**: Interactive conversations (REPL)

The project follows a library-first design where the core intelligence (LLM providers, prompt orchestration, tools, agents, workflows) is implemented as a reusable Go module.

## Current Status (Working on Phase 4.2)

✅ Phase 1: Core Foundation - Complete
✅ Phase 2.1: Configuration Management with Koanf - Complete
✅ Phase 2.2: Configuration Schema - Complete  
✅ Phase 2.3: Configuration Utilities - Complete (mostly)
✅ Phase 2.4: Unified Command System - Complete
✅ Phase 2.5: Core Commands Implementation - Complete
✅ Phase 2.6: Models inventory file - Complete
✅ Phase 3.1: CLI Structure Setup - Complete
✅ Phase 3.2: Ask Command - Complete
✅ Phase 3.2.1: CLI Help System Improvements - Complete (future enhancements pending)
✅ Phase 3.3: Chat Command & REPL Foundation - Complete
✅ Phase 3.4: Configuration Commands (using koanf) - Complete
✅ Phase 3.5: Logging and Verbosity Implementation - Complete
  ✅ Phase 3.5.1: Configuration Logging - Complete
  ✅ Phase 3.5.2: LLM Provider Logging - Complete
  ✅ Phase 3.5.3: Session Management Logging - Complete
  ✅ Phase 3.5.4: Command Execution Logging - Complete
  ✅ Phase 3.5.5: REPL Operations Logging - Complete
  ✅ Phase 3.5.6: File Operations Logging - Complete
  ✅ Phase 3.5.7: User-Facing Operations Logging - Complete
  ✅ Phase 3.5.8: Performance and Metrics Logging - Complete  
  ✅ Phase 3.5.9: Security and Audit Logging - Complete
  ✅ Phase 3.5.10: Testing and Integration - Complete
✅ Phase 3.6: History Commands - Complete
✅ Phase 4.1: Extended REPL Commands - Complete
✅ Phase 4.1.1: Fix logging and file attachment issues - Complete
🚧 Phase 4.2: Advanced Session Features - In Progress
  ✅ Phase 4.2.1: Auto-save functionality - Complete
  ✅ Phase 4.2.2: Session export formats (JSON/Markdown) - Complete
  ✅ Phase 4.2.3: Session search by content - Complete

### Recent Improvements
- Completed Session Management Features (Phase 4.2):
  - Implemented auto-save functionality with timers and signal handling
  - Added session export to JSON and Markdown formats
  - Created full-text session search across messages, prompts, and metadata
  - Fixed history command compilation error with SearchResult structure
  - Improved extract snippet function with word boundary adjustments
  - Complete test coverage for all session features
- Fixed logging and file attachment issues (Phase 4.1.1):
  - Fixed double initialization of logger
  - Logger now properly respects MAGELLAI_LOG_LEVEL environment variable
  - Implemented file attachment fallback for unsupported models
  - Models without file support now receive content as text (e.g., GPT-3.5-turbo)
- Dependency reduction (post Phase 4.1):
  - Updated go-llms from v0.2.1 to v0.2.4 (reduced dependencies by 11)
  - Removed pflag dependency from Magellai (replaced with map-based config)
  - Total reduction: 50 to 40 dependencies (20% reduction)
  - Binary size reduced from 15M to 14M (6.7% reduction)
  - All tests passing and linter compliant
- Previously completed REPL commands (Phase 4.1):
  - Added special commands for model switching, streaming, verbosity, output format
  - Added temperature and max_tokens controls
  - Added profile switching and attachment management
  - Added system prompt management
  - Added /config show and /config set commands
  - Created comprehensive tests for all new commands
- Previously completed:
  - History commands (Phase 3.6) with list, show, delete, export, search
  - Comprehensive logging (Phase 3.5) across all components
  - Chat command and REPL foundation (Phase 3.3)
  - Ask command with pipeline support (Phase 3.2)
  - Configuration commands with koanf integration (Phase 3.4)
  - Added appropriate error logging for all failure scenarios
- User-Facing Operations logging (3.5.7):
  - INFO level for model changes with before/after values
  - INFO level for profile switches with before/after values
  - INFO level for command invocations with command name and arguments
  - INFO level for session starts/ends (already implemented)
  - INFO level for configuration changes with keys and before/after values
  - All user-initiated actions now have comprehensive logging
- Performance and Metrics logging (3.5.8):
  - DEBUG level timing for configuration load operations
  - DEBUG level timing for LLM response generation (all methods)
  - DEBUG level timing for session operations (save, load, list)
  - DEBUG level timing for command execution
  - Duration tracking for streaming operations
  - Performance metrics for all major system operations
- Security and Audit logging (3.5.9):
  - DEBUG level API key usage with sanitization (shows only partial key)
  - INFO level configuration modifications (config changes, profile operations)
  - DEBUG level file access attempts (already implemented in 3.5.6)
  - ERROR level error conditions (comprehensive throughout codebase)
  - Added sanitizeAPIKey function for safe logging of sensitive data
  - Added audit logging for profile creation and copying
- Testing and Integration (3.5.10):
  - Created comprehensive logging tests in logger_test.go
  - Added tests for sensitive data sanitization
  - Added tests for different verbosity levels
  - Created performance benchmarks for logging
  - Created pkg/llm/sanitization_test.go for API key sanitization testing
  - All tests passing with minimal performance impact verified
- Fixed compilation errors related to GetProfile and len(exec.Flags)
- Fixed integration test to build test binary in bin directory for proper cleanup
- Fixed logging infrastructure to handle nil errors gracefully
- All operations now have appropriate logging with context
- Completed all sections of Phase 3.5
- Marked partial sections (3.2 and 3.2.1) for revisit in TODO.md
- All tests passing, code formatted, and linted successfully
- Phase 3.5 is now fully complete!
- History Commands (3.6):
  - Implemented all history subcommands (list, show, delete, export, search)
  - Created comprehensive history command with proper metadata and validation
  - Added tabular output for list and search commands
  - Implemented JSON and Markdown export formats
  - Added proper session detail display with message truncation
  - Complete test coverage with mocked configuration directory
  - Integrated with CLI command structure
  - All subcommands tested and working correctly

## Architecture

### Core Library Structure
```
pkg/
  llm/       → Provider drivers (OpenAI, Anthropic, Gemini)
  tool/      → Tool registry and execution
  agent/     → Multi-step autonomous capabilities
  workflow/  → YAML parser and DAG engine
  config/    → Configuration management with profiles
  repl/      → Interactive loop, conversation management, and session persistence
  command/   → Unified command system for CLI and REPL
  models/    → Model inventory and capability querying
```

### Front-End Structure
```
cmd/
  magellai/  → CLI entry point using Kong
```

### Plugin Architecture
- External binaries via naming convention: `magellai-tool-*`, `magellai-agent-*`
- Go plugins for tighter integration (optional)
- JSON-RPC or stdio communication protocol

## Common Development Commands

### Build Commands
```bash
# Build the main binary
go build -o magellai cmd/magellai/main.go

# Build with race detection
go build -race -o magellai cmd/magellai/main.go

# Install the binary
go install ./cmd/magellai
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/llm/...

# Run tests with race detection
go test -race ./...
```

### Linting and Formatting
```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run go vet
go vet ./...
```

## Key Design Patterns

### Command Structure
- Git-style subcommands: `magellai <verb> [options]`
- Consistent flag patterns across commands
- Progressive disclosure (simple first, complex as needed)

### Configuration Precedence (highest to lowest)
1. REPL commands (runtime only)
2. Command-line flags
3. Environment variables (`MAGELLAI_*`)
4. Profile overrides
5. Project configuration (`./.magellai.yaml`)
6. User configuration (`~/.config/magellai/config.yaml`)
7. System configuration (`/etc/magellai/config.yaml`)
8. Built-in defaults

### I/O Patterns
- Unix pipeline compatibility
- Streaming support via `--stream` flag
- Multiple input sources (args, stdin, files)
- Structured output formats (JSON, YAML, text)

### REPL Commands
- Conversation mode: Direct text input
- Command mode: `/` prefix (e.g., `/model gpt-4`)
- Special commands: `:` prefix (e.g., `:stream on`)

## Development Guidelines

### Library-First Approach
- Keep all LLM logic in the library (`pkg/`)
- Front-ends should only handle I/O and flag parsing
- Ensure library remains flag-free and testable

### Plugin Development
- Follow naming convention: `magellai-<type>-<name>`
- Implement JSON-RPC protocol for communication
- Register tools/agents/flows in appropriate registries

### Error Handling
- Use descriptive error codes in defined ranges:
  - 64-73: Input errors
  - 74-83: Model errors
  - 84-93: Network errors
  - 94-103: Processing errors

### Testing Strategy
- Unit tests for library components
- Integration tests for CLI/REPL
- Mock LLM providers for deterministic testing

## Important Files and Conventions

### Configuration Files
- User config: `~/.config/magellai/config.yaml`
- Project config: `.magellai.yaml`
- Plugin directory: `~/.config/magellai/plugins/`
- Session storage: `~/.magellai/sessions/`

### Command Examples
```bash
# Simple ask
magellai ask "Explain quantum computing"

# With attachments
magellai ask -a document.pdf "Summarize this"

# Streaming output
magellai ask --stream "Write a story"

# Using profiles
magellai --profile work ask "Draft an email"

# Start interactive chat
magellai chat

# Resume previous chat session
magellai chat --resume <session-id>

# Chat with specific model
magellai chat --model anthropic/claude-3-sonnet

# Pipeline usage
cat data.json | magellai extract summary
```

## Dependencies

The project uses:
- **go-llms v0.2.1**: LLM provider integration (OpenAI, Anthropic, Gemini)
- **Kong**: For CLI command structure
- **Koanf**: For configuration management
- **Standard Go libraries**: For core functionality

### go-llms Integration

Magellai uses `github.com/lexlapax/go-llms` as the backend LLM wrapper. This library provides:
- Unified interface for multiple LLM providers (OpenAI, Anthropic, Google Gemini)
- Structured response validation with JSON schema
- Tool integration and function calling
- Agent and workflow support
- Multimodal content support (text, images, files, videos, audio)

Key interfaces from go-llms:
- `domain.Provider`: Core LLM provider interface
- `domain.Message`: Message structure for conversations
- `workflow.Agent`: Individual agent with LLM and prompt
- `llmutil.Agent`: Convenience wrapper for common patterns

The library is available as:
- Go module dependency: `github.com/lexlapax/go-llms v0.2.1`
- Source reference: `./go-llms/` (git submodule)

## Implemented Modules

### internal/logging
- Flexible logging wrapper around slog
- Supports JSON and text output formats
- Configurable log levels and output destinations
- Helper functions for common logging patterns

### internal/configdir
- Configuration directory management
- Creates and manages `~/.config/magellai/` structure
- Handles project-specific config files (`.magellai.yaml`)
- Default configuration file generation

### pkg/llm/types
- Core types that wrap go-llms domain types
- Request/Response structs for LLM interactions
- Message type with multimodal attachment support
- Attachment types for image, audio, video, file, text
- PromptParams for configuring LLM behavior
- ModelInfo struct with capability tracking
- Conversion methods between Magellai and go-llms types
- Provider/model string parsing utilities

### pkg/llm/provider
- Provider adapter interface wrapping go-llms providers
- Factory methods for creating providers (OpenAI, Anthropic, Gemini, Mock)
- Configuration option methods (temperature, max tokens, etc.)
- Streaming support with StreamChunk type
- API key management from environment variables
- Model capability detection based on provider/model
- Error handling and validation
- Comprehensive option builder pattern

### pkg/config
- Complete configuration management with Koanf
- Multi-layer configuration support (default, system, user, project, env, flags)
- Profile system with inheritance
- Configuration validation and watchers
- Alias management for commands
- Provider and model configurations

### pkg/command
- Unified command system for CLI and REPL
- Command registry with metadata
- Core commands (model, config, profile, alias, help)
- Command validation and error handling
- Shared command infrastructure

### pkg/models
- Models inventory management
- JSON-based model database
- Query models by capabilities
- Provider and model family support

### pkg/repl
- REPL foundation for interactive chat
- Conversation management with message history
- Session persistence and management
- Interactive command loop
- Command mode (/) vs conversation mode
- Special commands (:) for settings
- Multi-line input support
- Attachment support for multimodal content
- Auto-save functionality with timers
- Session export (JSON, Markdown)
- Full-text session search
- Session list and management

### pkg/command/core/chat
- Chat command implementation
- Integrates CLI with REPL system
- Supports session resumption
- Model selection and attachment handling
- Launches interactive chat session

## Workflow Conventions

- Always update TODO.md after completion and before compaction.
- Stop and ask after every task section or major task before continuing on to another task.