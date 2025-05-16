# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magellai is a command-line interface (CLI) tool and REPL that interacts with Large Language Models (LLMs). It operates in two primary modes:
- **`ask` mode**: One-shot queries
- **`chat` mode**: Interactive conversations (REPL)

The project follows a library-first design where the core intelligence (LLM providers, prompt orchestration, tools, agents, workflows) is implemented as a reusable Go module.

## Current Status (Phase 1.6 Complete)

✅ Project structure set up  
✅ Makefile with build/test/lint targets  
✅ MIT license  
✅ Basic README.md  
✅ Logging infrastructure using slog  
✅ Configuration directory management  
✅ Core data models and go-llms type wrappers  
✅ Model capability system (no hard-coded models)  
✅ Full multimodal support (text, image, audio, video, file)  
✅ Provider adapter interface wrapping go-llms  
✅ Provider factory with configuration helpers  
✅ Streaming support for all providers  
✅ Comprehensive provider options  
✅ All tests passing  
✅ Placeholder main.go for build verification  
✅ Provider implementations (OpenAI, Anthropic, Gemini, Mock)  
✅ Comprehensive unit tests for all providers  
✅ High-level Ask function with multimodal support  
✅ Streaming response support  
✅ Complete error handling  
✅ Full test coverage for Ask functionality  

Next: Phase 2 - Configuration and Command Foundation

Starting with Phase 2.1: Configuration Management with Koanf

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

## Workflow Conventions

- Always update TODO.md after completion and before compaction.
- Stop and ask after every task section or major task before continuing on to another task.