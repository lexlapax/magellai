# Magellai User Guide

## Introduction

Welcome to the Magellai user guide! This guide provides comprehensive documentation for using the Magellai CLI tool to interact with Large Language Models (LLMs).

Magellai is a powerful command-line interface and REPL designed to make working with LLMs easy and efficient. Whether you're a developer, researcher, or enthusiast, this guide will help you get the most out of Magellai.

## Getting Started

### Installation

```bash
# Install from source
git clone https://github.com/lexlapax/magellai.git
cd magellai
make build
make install
```

### Quick Start

```bash
# Simple questions with ask mode
magellai ask "Explain quantum computing"

# Interactive conversations with chat mode
magellai chat

# Use a specific model
magellai ask --model openai/gpt-4 "Write a haiku about coding"

# Stream responses
magellai ask --stream "Tell me a story"
```

## Core Features

### Ask Mode

The `ask` mode provides a simple way to get quick answers without entering a full chat session:

```bash
magellai ask "How do I check disk space in Linux?"
```

Options:
- `--model`: Specify model to use (`openai/gpt-4`, `anthropic/claude-3-opus`, etc.)
- `--provider`: Specify provider to use (`openai`, `anthropic`, `gemini`)
- `--stream`: Stream the response as it's generated
- `--file`: Attach files to the request
- `--system`: Provide system instructions

### Chat Mode

The `chat` mode provides an interactive REPL experience for ongoing conversations:

```bash
magellai chat
```

Once in chat mode, you can use special commands by prefixing with `/` or `:`:
- `/help`: Show available commands
- `/model <model>`: Change the model
- `/system <text>`: Set system instructions
- `/export <format>`: Export the conversation
- `/history`: Show session history
- `/switch <session>`: Switch to another session
- `/exit` or `/quit`: Exit chat mode

## Session Management

### Session History

View your past conversations:

```bash
magellai history list
```

Resume a previous session:

```bash
magellai chat --session 1234567890
```

### Session Branching

[Session branching](session-branching-guide.md) allows you to create alternative paths from any point in your conversation history.

Commands:
- `/branch <name>`: Create a new branch from the current session
- `/branches`: List all branches of the current session
- `/tree`: Display the branch tree for the current session
- `/switch <id>`: Switch to a different branch

See the [detailed session branching guide](session-branching-guide.md) for more information.

### Session Merging

[Session merging](session-merging-guide.md) allows you to combine two sessions into one.

Commands:
- `/merge <id>`: Merge another session into the current one
- `/merge <id> rebase`: Rebase current session on another session
- `/merge <id> create-branch <name>`: Merge and create a new branch

See the [detailed session merging guide](session-merging-guide.md) for more information.

## Configuration

### Configuration Files

Magellai uses a flexible configuration system with the following precedence:

1. Built-in defaults
2. System config: `/etc/magellai/config.yaml`
3. User config: `~/.config/magellai/config.yaml`
4. Project config: `.magellai.yaml` (searched upward from current directory)
5. Environment variables: `MAGELLAI_*`
6. Command-line flags

### Create Default Configuration

Generate a default configuration file:

```bash
magellai config generate
```

### Provider Configuration

Configure API keys and other provider settings:

```yaml
providers:
  - id: "openai-1"
    name: "OpenAI"
    type: "openai"
    models:
      - id: "gpt-4"
        name: "GPT-4"
      - id: "gpt-4o"
        name: "GPT-3.5 Turbo"
    options:
      api_key: "env:OPENAI_API_KEY"

  - id: "anthropic-1"
    name: "Anthropic"
    type: "anthropic"
    models:
      - id: "claude-3-opus"
        name: "Claude 3 Opus"
      - id: "claude-3-sonnet"
        name: "Claude 3 Sonnet"
    options:
      api_key: "env:ANTHROPIC_API_KEY"
```

### Profiles

Create profiles for different use cases:

```yaml
profiles:
  - id: "default"
    name: "Default Profile"
    provider: "openai-1"
    model: "gpt-4o"
    options:
      temperature: 0.7
      max_tokens: 1000
  
  - id: "creative"
    name: "Creative Writing"
    provider: "anthropic-1"
    model: "claude-3-opus"
    options:
      temperature: 0.9
      max_tokens: 2000
```

## Advanced Features

### File Attachments

Attach files to your conversations:

```bash
magellai ask --file /path/to/document.pdf "Summarize this document"
```

In chat mode:
```
/attach /path/to/image.jpg
What's in this image?
```

### Provider Fallback

Configure fallback chains for reliable operations:

```yaml
fallback_chains:
  - id: "reliable-chain"
    name: "Reliable Chain"
    providers:
      - "anthropic-1"
      - "openai-1"
      - "gemini-1"
    options:
      max_retries: 3
      retry_delay: 1000
```

Use a fallback chain:
```bash
magellai ask --fallback reliable-chain "What's the capital of France?"
```

## Troubleshooting

### Common Issues

1. **API Key Issues**
   - Ensure API keys are properly set in environment variables
   - Verify configuration file syntax

2. **Connection Problems**
   - Check internet connection
   - Verify provider status
   - Try using a fallback chain

3. **Rate Limiting**
   - Reduce request frequency
   - Use different providers
   - Implement backoff strategies

### Logging

Enable verbose logging for troubleshooting:

```bash
magellai --log-level debug ask "Test question"
```

Log levels:
- `error`: Only show errors
- `warn`: Show warnings and errors (default)
- `info`: Show informational messages
- `debug`: Show all debug information

## Additional Resources

- [Technical Documentation](../technical/README.md): Architecture and implementation details
- [API Documentation](../api/README.md): Programmatic interfaces
- [Examples](../examples/README.md): Example usage patterns
- [Planning Documents](../planning/README.md): Design decisions and philosophy

## Command Reference

| Command | Description |
|---------|-------------|
| `magellai ask` | One-shot query mode |
| `magellai chat` | Interactive chat mode |
| `magellai config` | Configuration management |
| `magellai config show` | Show current configuration |
| `magellai config generate` | Generate default config |
| `magellai history` | Session history management |
| `magellai history list` | List session history |
| `magellai history delete` | Delete sessions |
| `magellai help` | Display help information |
| `magellai version` | Show version information |