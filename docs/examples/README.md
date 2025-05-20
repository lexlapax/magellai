# Magellai Usage Examples

This directory contains practical examples and tutorials for using Magellai in various scenarios.

## Session Management Examples

- [**Branching Examples**](branching-examples.md): Examples of using session branching for different conversation paths
- [**Merging Examples**](merging-examples.md): Examples of merging sessions using different strategies

## CLI Usage Examples

### Basic Commands

```bash
# Simple question in ask mode
magellai ask "What is the capital of France?"

# Interactive chat
magellai chat

# Using a specific model
magellai ask --model openai/gpt-4 "Explain quantum computing"

# Streaming response
magellai ask --stream "Write a short story about a robot"

# Setting system instructions
magellai ask --system "You are a helpful assistant" "How can you help me?"

# Attaching files
magellai ask --file /path/to/document.pdf "Summarize this document"
```

### Configuration Examples

```bash
# Show current configuration
magellai config show

# Generate default configuration
magellai config generate

# Set default model
magellai config set model openai/gpt-4

# Set default provider
magellai config set provider anthropic-1
```

### Session History

```bash
# List all sessions
magellai history list

# Show specific session
magellai history show 1234567890

# Delete session
magellai history delete 1234567890

# Continue previous session
magellai chat --session 1234567890
```

## REPL Command Examples

Once in chat mode (using `magellai chat`), you can use these commands:

```
# Get help
/help

# Change model
/model openai/gpt-4

# Set system instructions
/system You are a helpful assistant that speaks like a pirate.

# Attach file
/attach /path/to/image.jpg

# Show message history
/history

# Export conversation
/export markdown ~/conversation.md

# Create branch
/branch alternative-approach

# List branches
/branches

# Switch to another branch
/switch session_1234567890

# Merge sessions
/merge session_1234567890

# Exit chat
/exit
```

## Advanced Usage Examples

### Provider Fallback

```bash
# Define fallback chain in config.yaml
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

# Use fallback chain
magellai ask --fallback reliable-chain "What's the capital of France?"
```

### Session Branching Workflow

See detailed examples in [branching-examples.md](branching-examples.md).

```bash
# Start chat
magellai chat

# Have a conversation...

# Create a branch to explore alternative approach
/branch alternative-approach

# Continue conversation in branch...

# Create another branch from original
/switch main-session
/branch another-approach

# View branch tree
/tree
```

### Session Merging Workflow

See detailed examples in [merging-examples.md](merging-examples.md).

```bash
# Start with two sessions
magellai chat --session session1
# Have a conversation...

# Switch to another session
/switch session2
# Have another conversation...

# Merge session1 into session2
/merge session1

# Create a merged branch
/merge session1 create-branch merged-approach
```

## API Usage Examples

For programmatic API usage examples, see the [API Documentation](../api/README.md).

## Related Documentation

- [User Guide](../user-guide/README.md): Complete user documentation
- [Technical Guide](../technical/README.md): Technical implementation details
- [API Documentation](../api/README.md): Programmatic interfaces