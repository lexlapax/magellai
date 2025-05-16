# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magellai is a command-line interface (CLI) tool and REPL that interacts with Large Language Models (LLMs). It operates in two primary modes:
- **`ask` mode**: One-shot queries
- **`chat` mode**: Interactive conversations (REPL)

The project follows a library-first design where the core intelligence (LLM providers, prompt orchestration, tools, agents, workflows) is implemented as a reusable Go module.

## Current Status (Phase 3.2 Complete)

✅ Phase 1: Core Foundation - Complete
✅ Phase 2.1: Configuration Management with Koanf - Complete
✅ Phase 2.2: Configuration Schema - Complete  
✅ Phase 2.3: Configuration Utilities - Complete (mostly)
✅ Phase 2.4: Unified Command System - Complete
✅ Phase 2.5: Core Commands Implementation - Complete
✅ Phase 2.6: Models Static Inventory - Complete
✅ Phase 3.1: CLI Structure Setup - Complete
✅ Phase 3.2: Ask Command Implementation - Complete

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

✅ Command execution framework
  - Command executor with validation and error handling
  - Pre/post execution hooks for extensibility
  - Argument and flag parsing with type validation
  - Context-aware flag type checking
  - Support for boolean flags without explicit values
  - Comprehensive error handling with custom validation errors
  - Full unit test coverage with all edge cases

✅ Models static inventory (Phase 2.6)
  - Created comprehensive models.json file in root directory
  - Defined JSON schema with metadata and model information
  - Included all current models from OpenAI, Anthropic, and Google
  - Detailed capability breakdown (text, image, audio, video, file)
  - Read/write permissions for each capability
  - Additional capabilities: function_calling, streaming, json_mode
  - Model metadata: context_window, max_output_tokens, pricing, training_cutoff
  - Created pkg/models package for loading and querying models.json
  - Full test coverage for models package

### Additional Improvements:
- Cleaned up help implementation by removing old help.go file
- Consolidated all help tests into a single help_test.go file
- Removed all backup and intermediary files
- Fixed all linting errors
- Added command executor with proper validation
- Implemented intelligent flag parsing logic
- Created static models inventory with comprehensive model information

### Phase 3.1 Completed:
✅ CLI Framework research and selection
  - Evaluated multiple frameworks (Cobra, Kong, urfave/cli, Kingpin, go-flags, docopt)
  - Selected Kong + kongplete based on minimal dependencies, flexibility, testing ease
  - Documented decision in docs/technical/cli_framework_analysis.md

✅ Kong CLI implementation
  - Implemented main.go with Kong framework
  - Created root command with global flags
  - Integrated command registry with Kong commands
  - Created stub implementations for ask and chat
  - Connected CLI to unified command system
  - Fixed type assertion issues for CLI interface
  - Registered all core commands from pkg/command/core

✅ Version command implementation
  - Created core version command in pkg/command/core/version.go
  - Implemented both --version flag and version subcommand for Unix compatibility
  - Support for build-time version injection via ldflags
  - Support for both text and JSON output formats
  - Made --version flag visible in help text
  - Comprehensive unit tests in version_test.go
  - Full integration with command registry and executor

✅ Main.go test coverage
  - Created main_test.go with comprehensive CLI tests
  - Tests for version flag behavior and subcommand
  - Tests for global flags (config, profile, verbose, debug)
  - Tests for command parsing and registration
  - Created integration_test.go for end-to-end testing
  - Fixed test issues with exec.Command approach
  - Achieved full test coverage for error handling

### Phase 3.2 Completed:
✅ Ask command implementation
  - Full ask command implementation in pkg/command/core/ask.go
  - Support for multimodal attachments (files, images, etc.)
  - Streaming response support with proper output handling
  - Provider selection based on model format (provider/model)
  - Integration with Kong CLI framework
  - Complete flag support:
    - --model/-m for model selection
    - --attach/-a for file attachments (repeatable)
    - --stream for streaming responses
    - --temperature/-t for temperature control
    - --max-tokens for response length
    - --system/-s for system prompts
    - --format for response format hints
    - --output for output format (global flag)
  - Full test coverage for ask command functionality
  - Proper error handling and validation
  - Deleted old pkg/magellai.go and pkg/magellai_test.go (initial ask implementations)
  - Fixed all test failures related to Flags type changes
  - Ensured all ExecutionContext instances have Flags field initialized

### Next: Phase 3.3 - Chat Command Implementation
Now that the ask command is complete, the next step is to implement the chat command:
- [ ] Implement chat subcommand
- [ ] Launch REPL mode
- [ ] Profile selection
- [ ] Session resume support
- [ ] Initial attachments support

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
  magellai/  → CLI entry point using Kong
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
# Run unit tests only (fast)
make test

# Run integration tests only
make test-integration

# Run all tests (unit and integration)
make test-all

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
- **Kong**: For CLI command structure and argument parsing
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

### pkg/command
- Command execution framework with validation
  - CommandExecutor for orchestrating command execution
  - Pre/post execution hooks for extensibility
  - Comprehensive argument and flag parsing
  - Intelligent flag type validation
  - Support for boolean flags without explicit values
  - Custom validation error types for detailed error reporting
  - Integration with the command registry
  - Context-aware execution with proper I/O handling
  - Full test coverage including edge cases

### pkg/models
- Static model inventory management from models.json
  - Load and parse models.json from root directory
  - Query models by provider, name, or full name (provider/model)
  - Filter models by capabilities (text, image, audio, video, file)
  - Support for read/write capability queries
  - List all providers and model families
  - Get models with specific features (streaming, json_mode, function_calling)
  - Model metadata access (context window, pricing, training cutoff)
  - Comprehensive test coverage

## Workflow Conventions

- Always update TODO.md after completion and before compaction.
- Stop and ask after every task section or major task before continuing on to another task.