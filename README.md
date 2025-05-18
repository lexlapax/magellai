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
- Extensible plugin architecture
- Tool and agent framework
- Workflow automation
- Session management and history

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

## Documentation

- [Getting Started](docs/user-guide/getting-started.md)
- [Configuration Guide](docs/user-guide/README.md)
- [Plugin Development](docs/technical/README.md)
- [API Reference](docs/api/README.md)
- [Architecture Overview](docs/technical/architecture.md)

### Planning & Design

- [CLI UX Research with ChatGPT O3](docs/planning/cli-ux-research-chatgpt-o3.md)
- [Configuration Precedence](docs/planning/configuration-precedence.md)
- [Philosophy of CLI Design](docs/planning/philosophy-cmdline.md)
- [UX-Driven Implementation](docs/planning/ux-driven-implementation.md)

## Configuration

Configuration files are loaded in the following order (later sources override earlier ones):

1. Built-in defaults
2. System config: `/etc/magellai/config.yaml`
3. User config: `~/.config/magellai/config.yaml`
4. Project config: `.magellai.yaml` (searched upward from current directory)
5. Environment variables: `MAGELLAI_*`
6. Command-line flags

See [Configuration Guide](docs/user-guide/README.md) for detailed information.

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