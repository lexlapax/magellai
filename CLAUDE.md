# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magellai is a command-line interface (CLI) tool and REPL that interacts with Large Language Models (LLMs). It operates in two primary modes:
- **`ask` mode**: One-shot queries
- **`chat` mode**: Interactive conversations (REPL)

The project follows a library-first design where the core intelligence (LLM providers, prompt orchestration, tools, agents, workflows) is implemented as a reusable Go module.

## Current Status (Phase 2.5 Complete)

✅ Phase 1: Core Foundation - Complete
✅ Phase 2.1: Configuration Management with Koanf - Complete
✅ Phase 2.2: Configuration Schema - Complete  
✅ Phase 2.3: Configuration Utilities - Complete (mostly)
✅ Phase 2.4: Unified Command System - Complete
✅ Phase 2.5: Core Commands Implementation - Complete

### Completed Features:
- Project structure and build system
- Logging infrastructure using slog
- Core data models and go-llms type wrappers
- Multi-modal support (text, image, audio, video, file)
- Provider implementations (OpenAI, Anthropic, Gemini, Mock)
- High-level Ask function with streaming support
- Configuration management with Koanf v2
- Multi-layer configuration support (defaults, system, user, project, env, CLI)
- Profile system for configuration management
- Configuration validation and type-safe access
- Configuration export/import functionality
- Provider/model string parsing utilities
- Unified command system with interface, registry, and validation
- Command categories for CLI, REPL, and API support
- Flag-to-command mapping for REPL
- Command discovery and registration mechanisms
- Unified help system for all interfaces
- Full test coverage for config package (21.6%) and command package (49.0%)
- Model command implementation with list, info, and select functionality
- Config command implementation with comprehensive subcommands
- Profile command implementation with complete lifecycle management

### Phase 2.5 Completed:
✅ Model command implementation
  - List all available models
  - Show model information (capabilities, parameters)
  - Select model using provider/modelname format
  - Automatic provider switching when model changes
  - Comprehensive unit tests

✅ Config command implementation
  - Comprehensive subcommands (list, get, set, validate, export, import)
  - Profile management (create, switch, delete, export)
  - Full unit test coverage
  - Fixed all linting errors (error checks)

✅ Profile command implementation
  - Complete lifecycle management (create, switch, update, copy, delete)
  - Profile export/import functionality
  - Show current and specific profile details
  - List all available profiles
  - Full unit test coverage with lifecycle tests
  - Fixed test ordering issues for map comparisons

✅ Alias command implementation
  - Add, remove, list, show, clear aliases
  - Support for CLI and REPL scopes
  - Export/import functionality
  - Comprehensive unit tests

✅ Help command enhancements
  - Context-aware help for CLI vs REPL
  - Command categorization by interface
  - Alias resolution in help display
  - Formatted command lists with categories
  - Error suggestions for command not found
  - Full unit test coverage
  - Consolidated all help functionality into core package

### Additional Improvements:
- Cleaned up help implementation by removing old help.go file
- Consolidated all help tests into a single help_test.go file
- Removed all backup and intermediary files
- Fixed all linting errors

### Next: Phase 3 - CLI Implementation
With Phase 2.5 now complete, the next steps are:
- [ ] Create command execution framework (final item from Phase 2.5)
- [ ] Add command validation and error handling (final item from Phase 2.5)
- [ ] Begin Phase 3: CLI implementation with a chosen framework (Cobra, urfave/cli, etc.)

## Architecture

### Core Library Structure
```
pkg/
  llm/       → Provider drivers (OpenAI, Anthropic, Gemini)
  tool/      → Tool registry and execution
  agent/     → Multi-step autonomous capabilities
  workflow/  → YAML parser and DAG engine
  session/   → Conversation storage and recall
  config/    → Configuration management with profiles
  repl/      → Interactive loop
```

### Front-End Structure
```
cmd/
  magellai/  → CLI entry point using Cobra
```

### Plugin Architecture
- External binaries via naming convention: `magellai-tool-*`, `magellai-agent-*`
- Go plugins for tighter integration (optional)
- JSON-RPC or stdio communication protocol

## Common Development Commands

(To be determined once the project structure is implemented)

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
- Script mode: `:` prefix (e.g., `:lua result = ...`)

## Development Guidelines

### Library-First Approach
- Keep all LLM logic in the library (`pkg/`)
- Front-ends should only handle I/O and flag parsing
- Ensure library remains flag-free and testable

### Plugin Development
- Follow naming convention: `magellai-<type>-<n>`
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

# Pipeline usage
cat data.json | magellai extract summary
```

## Dependencies

The project uses:
- **go-llms v0.2.1**: LLM provider integration (OpenAI, Anthropic, Gemini)
- **Cobra**: For CLI command structure
- **Koanf v2**: For configuration management (replacing Viper)
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
- Multi-layer configuration management using Koanf v2
- Support for configuration precedence (defaults, system, user, project, env, CLI flags)
- Profile system for different use cases
- Type-safe configuration access methods
- Configuration validation with detailed error reporting
- Configuration schema with provider-specific settings
- Model-specific settings and parameter validation
- Configuration export/import functionality
- Utility functions for provider/model string parsing

### pkg/command/core
- Model command for LLM model management
  - List available models with provider filtering
  - Show model information including capabilities
  - Select and switch models using provider/modelname format
  - Automatic provider switching when model changes
- Config command for configuration management
  - Comprehensive subcommands (list, get, set, validate, export, import)
  - Profile management integrated into config command
  - Support for multiple output formats (text, JSON, YAML)
  - Full configuration lifecycle management
- Profile command for profile lifecycle management
  - Create, switch, update, copy, and delete profiles
  - Show current and specific profile details
  - List all available profiles
  - Export/import profile configurations
  - Profile-specific settings management
- Alias command for command alias management
  - Add, remove, list, show, and clear aliases
  - Support for both CLI and REPL scopes
  - Scope-based filtering (all/cli/repl)
  - Export aliases to JSON format
  - Works with configuration system for persistence
- Enhanced Help command with context awareness
  - Context-aware display for CLI vs REPL interfaces
  - Command categorization by interface availability
  - Alias resolution for finding aliased commands
  - Formatted command lists with proper categorization
  - Error suggestions for mistyped commands
  - Integration with the command registry and config system
  - Consolidated implementation (removed old help.go)
  - Single unified test file for all help functionality

## Workflow Conventions

- Always update TODO.md after completion and before compaction.
- Stop and ask after every task section or major task before continuing on to another task.