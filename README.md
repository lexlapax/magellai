# Magellai

A powerful command-line interface (CLI) and REPL for interacting with Large Language Models (LLMs).

## Overview

Magellai is a library-first CLI tool that provides a unified interface for multiple LLM providers including OpenAI, Anthropic, and Google Gemini. It operates in two primary modes:

- **`ask` mode**: One-shot queries for quick interactions
- **`chat` mode**: Interactive conversations through a REPL interface

The architecture follows a clean domain-driven design with a central domain layer containing all business entities, ensuring clear separation of concerns and maintainable code structure.

## Features

- Multiple LLM provider support (OpenAI, Anthropic, Gemini)
- Interactive REPL with rich features
- Streaming responses
- Configurable profiles for different use cases
- Session management with branching and merging
- Provider fallback mechanisms
- Comprehensive error handling
- File attachments
- Storage backends (filesystem, SQLite)

## Installation

```bash
# Build from source
make build

# Install to $GOPATH/bin
make install
```

## Quick Start

```bash
# Simple ask mode
magellai ask "Explain quantum computing"

# Interactive chat mode
magellai chat

# Use a specific model
magellai ask --model openai/gpt-4 "Write a haiku about coding"

# Stream responses
magellai ask --stream "Tell me a story"
```

## Architecture

Magellai follows a clean, layered architecture:

```
┌─────────────────┐
│    CLI Layer    │
└────────┬────────┘
         │
┌────────▼────────┐
│ Application Layer│
└────────┬────────┘
         │
┌────────▼────────┐
│   Domain Layer  │
└────────┬────────┘
         │
┌────────▼────────┐
│Infrastructure Layer│
└─────────────────┘
```

The domain layer is at the core, containing all business entities and rules. Application logic is built around the domain, while infrastructure adapters implement technical capabilities.

For detailed architecture information, see [Architecture Documentation](docs/technical/architecture.md).

## Documentation

### User Documentation

- [User Guide](docs/user-guide/README.md): Complete guide for Magellai users
- [Session Branching Guide](docs/user-guide/session-branching-guide.md): Using branch features
- [Session Merging Guide](docs/user-guide/session-merging-guide.md): Combining conversations

### Technical Documentation

- [Technical Guide](docs/technical/README.md): Technical implementation details
- [Architecture](docs/technical/architecture.md): System architecture and design
- [Domain Layer](docs/technical/domain-layer-architecture.md): Domain-driven design
- [Type Ownership](docs/technical/type-ownership.md): Type definitions and ownership

### API Documentation

- [API Reference](docs/api/README.md): Programmatic interfaces
- [Session Branching API](docs/api/session-branching-api.md): Branching API details
- [Session Merging API](docs/api/session-merging-api.md): Merging API details

### Examples

- [Usage Examples](docs/examples/README.md): Practical examples and tutorials
- [Branching Examples](docs/examples/branching-examples.md): Branch usage patterns
- [Merging Examples](docs/examples/merging-examples.md): Merge usage patterns

### Planning & Design

- [CLI UX Research](docs/planning/cli-ux-research-chatgpt-o3.md): UX research findings
- [Configuration Precedence](docs/planning/configuration-precedence.md): Config design
- [Philosophy of CLI Design](docs/planning/philosophy-cmdline.md): Design principles
- [UX-Driven Implementation](docs/planning/ux-driven-implementation.md): UX approach

## Configuration

Configuration files are loaded in the following order (later sources override earlier ones):

1. Built-in defaults
2. System config: `/etc/magellai/config.yaml`
3. User config: `~/.config/magellai/config.yaml`
4. Project config: `.magellai.yaml` (searched upward from current directory)
5. Environment variables: `MAGELLAI_*`
6. Command-line flags

Generate a default configuration:

```bash
magellai config generate
```

## Development

```bash
# Run unit tests (fast)
make test

# Run integration tests
make test-integration

# Run all tests
make test-all

# Run linter
make lint

# Format code
make fmt

# Run all pre-commit checks (includes unit tests)
make pre-commit
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## Acknowledgments

Built with [go-llms](https://github.com/lexlapax/go-llms) - a unified Go library for LLM providers.