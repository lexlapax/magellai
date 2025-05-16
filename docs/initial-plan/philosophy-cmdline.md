# Magellai: Unix Philosophy Meets AI Language Models

Magellai aims to be the first-class Unix citizen in the LLM ecosystem, bridging deterministic command-line traditions with the probabilistic nature of language models. This vision document outlines how Magellai will embrace Unix principles while addressing the unique challenges of LLM integration.

## The Unix Foundation for AI Tools

Unix philosophy has stood the test of time for over five decades because its core principles—simplicity, modularity, composability—align with how humans naturally solve problems. **LLM tools must embrace these principles** rather than reinvent patterns that have proven effective across generations of software.

The traditional terminal remains the most efficient interface for many tasks because it enables precise control, scriptability, and composition of tools. By bringing LLMs into this ecosystem as good citizens rather than foreign entities, we can enhance existing workflows rather than replace them.

Modern CLI tools follow established design patterns that make them intuitive, consistent, and powerful. After analyzing how git, kubectl, docker, and aws-cli implement key CLI features, clear best practices emerge that balance usability with flexibility.

### Core Principles Guiding Magellai

Magellai's design embraces Unix philosophy while acknowledging LLMs' unique characteristics:

1. **Do one thing well**: Focus on being an exceptional interface to language models, not trying to become an application framework
2. **Composability first**: Design every feature to work within pipelines and with existing tools
3. **Progressive disclosure**: Simple for basic tasks, powerful for complex ones
4. **Human text interfaces**: Text remains the universal interface, with structured outputs when appropriate
5. **Local-first operation**: Respect privacy and network independence with local execution options

## Command Architecture: From Simple Prompts to Complex Workflows

Magellai adopts a layered command structure that supports the full spectrum from simple one-off prompts to sophisticated agent workflows.

### Command Hierarchy

The command structure follows a git-style subcommand pattern with consistent verb-noun conventions:

```
magellai [global options] <command> [subcommand] [options]
```

**Four main command categories** represent increasing levels of capability:

1. **Core commands**: Direct LLM interactions 
   - `magellai ask` - Simple prompt-response
   - `magellai chat` - Interactive REPL/conversation mode
   - `magellai repl` - Alias for chat, emphasizing the REPL nature

2. **Tool commands**: Enhanced capabilities that don't use an llm
   - `magellai tool run calculate` - Structured content generation
   - `magellai tool run get-weather` - Text transformations

3. **Agent commands**: Autonomous capabilities that use llms and call tools
   - `magellai agent run transform summarize` - Execute an agent with a summarize subcommand
   - `magellai agent run extract keywords` - Execute an agent with a extract keywords subcommand
   - `magellai agent run get-news` - Execute an agent with specific capabilities
   - `magellai agent run get-news` - Execute an agent with specific capabilities
   - `magellai agent run research` - Define custom agents
   - `magellai agent list` - View available agents

4. **Workflow commands**: Multi-step processes and worklfows that may use multiple agents and tools
   - `magellai workflow run deep-research` - Execute a defined workflow
   - `magellai workflow list` - Veiew available workflows
   - `magellai workflow define research` - Create new workflows
   - `magellai workflow visualize deep-research` - Display workflow structure

This hierarchy follows established patterns: **Tools should use no more than 2-3 levels of nesting** in command hierarchies to balance organization with usability. The verb-noun pattern (like Git's `git commit`) works well for focused tools, while noun-verb (like kubectl's `kubectl pods get`) is better for tools with many resource types.

### Command Composability

Magellai supports both external and internal composition:

- **External composition** (Unix pipeline style):
  ```bash
  cat document.txt | magellai agent transform summarize | magellai agent extract keywords > keywords.txt
  ```

- **Internal composition** (workflow definition):
  ```bash
  magellai workflow define research --steps "transform:summarize,extract:keywords,generate:report"
  ```

The tool preserves metadata across transformations to enable rich pipelines while maintaining Unix compatibility.

### Flag and Command Interleaving

Following modern CLI patterns, Magellai supports flexible flag positioning. **Most modern CLI tools support flexible flag positioning**, allowing flags to appear before, after, or mixed with commands:

```bash
# All these work in Magellai
magellai --profile=work ask "Question?"
magellai ask --profile=work "Question?"
magellai ask "Question?" --profile=work

# Streaming as a flag
magellai ask "Explain quantum physics" --stream
magellai transform summarize --stream < large_document.txt
```

This flexibility comes from modern parsing libraries like spf13/cobra that implement sophisticated flag handling, making commands easier to construct and modify.

## User Experience Philosophy

Magellai balances power and simplicity through several core UX principles:

### Progressive Disclosure

- **Simple first interaction**: `magellai ask "How does photosynthesis work?"`
- **Graduated complexity**: More features revealed as users explore help content
- **Consistent command patterns**: Commands follow predictable patterns so learning one teaches the structure of others

### Dual-Mode Operation

Magellai detects and adapts to interactive versus non-interactive contexts:

- **Interactive mode**: Rich colored output, helpful prompts, and suggested commands
- **Non-interactive mode**: Machine-parseable output, minimal stderr, and appropriate exit codes
- **Automatic detection**: TTY detection with manual override options

### Global Logging Controls

Following established patterns, **logging is implemented through global flags** rather than commands:

```bash
# Numeric level pattern (kubectl style)
magellai -v=3 ask "Question?"  # 0=silent, higher=more verbose

# Boolean pattern (docker/aws style)
magellai --debug transform summarize  # debug mode on/off
```

Magellai adopts conventional flag names like `-v/--verbose` and `--debug` that users already recognize. Environment variables serve as alternatives to command-line flags, making them suitable for CI/CD pipelines and scripts (e.g., `MAGELLAI_VERBOSE=3`).

### Input Flexibility

- **Multiple input sources**: Command arguments, stdin, files, URLs
- **Fragment integration**: Load context from diverse sources (inspired by simonw/llm)
- **Context persistence**: Session state maintained across commands when appropriate

### Output Control

- **Structured formats**: JSON, YAML, and plain text outputs
- **Format validation**: Schema enforcement for consistent structured data
- **Streaming support**: Real-time output for long-running operations (via `--stream` flag)
- **Pagination control**: Smart handling of large outputs

## Interactive REPL Experience

The `magellai chat` (aliased as `magellai repl`) command provides a rich interactive environment that combines natural conversation with programmatic control.

### REPL Command Modes

The REPL operates in three distinct modes, distinguished by prefix characters:

1. **Conversation Mode (default)**: Direct text input for natural LLM interaction
   ```
   magellai> How does quantum entanglement work?
   ```

2. **Command Mode (`/` prefix)**: Access to CLI commands within the REPL
   ```
   magellai> /config set model gpt-4
   magellai> /workflow run research quantum computing
   magellai> /context add paper.pdf
   ```

3. **Script Mode (`:` prefix)**: Direct script execution in supported languages
   ```
   magellai> :lua result = llm.generate("Explain this", "openai", "gpt-4")
   magellai> :js console.log(await generateAsync("Hello", "openai", "gpt-4"))
   magellai> :tengo fmt.println(llm.generate("Test", "openai", "gpt-4"))
   ```

### Available REPL Commands

Essential REPL commands available via the `/` prefix:

```
# Model and Provider Management
/provider <name>              # Switch to different provider
/model <name>                 # Change the active model
/models                       # List available models
/temperature <0.0-2.0>        # Adjust temperature setting

# Context Management
/context add <file>           # Add file to conversation context
/context list                 # Show current context items
/context clear                # Clear all context
/context remove <file>        # Remove specific context item

# Session Management
/session save <name>          # Save current conversation
/session load <name>          # Load previous conversation
/session export <file>        # Export session to file
/history                      # Show conversation history
/clear                        # Clear current conversation

# Workflow and Agent Control
/workflow run <name> [args]   # Execute a workflow
/agent run <name> [args]      # Run an agent
/script run <file>            # Execute a script file

# Configuration
/config set <key> <value>     # Set configuration value
/config get <key>             # Get configuration value
/config list                  # Show current configuration

# Help and Information
/help [command]               # Show help for commands
/commands                     # List all available commands
/providers                    # List configured providers
/info                         # Show system information

# Control Commands
/stream on|off                # Toggle streaming responses
/format json|yaml|text        # Set output format
/verbose on|off               # Toggle verbose output
/exit                         # Exit the REPL
/quit                         # Alias for exit
```

### Multi-line Input Support

The REPL supports multi-line input for complex prompts:

```
magellai> """
... Analyze the following code and suggest improvements:
... 
... def factorial(n):
...     if n <= 1:
...         return 1
...     return n * factorial(n-1)
... """
```

### Context Persistence

The REPL maintains context across commands within a session:

```
magellai> What is the capital of France?
[Response: The capital of France is Paris...]

magellai> What is its population?
[Response: The population of Paris is approximately 2.1 million...]
```

### Smart Command Completion

Tab completion provides intelligent suggestions:

```
magellai> /con<TAB>
/config   /context

magellai> /context a<TAB>
/context add
```

### REPL Configuration

The REPL can be customized through configuration:

```yaml
# ~/.config/magellai/config.yaml
repl:
  prompt: "magellai> "
  multiline_prompt: "... "
  history_file: ~/.magellai_history
  history_size: 10000
  auto_save_session: true
  default_format: text
  tab_completion: true
  syntax_highlighting: true
  vi_mode: false  # Use vi-style keybindings
```

### Stream Flag vs Command

Based on the REPL design, streaming is better implemented as a flag rather than a separate command:

```bash
# Command-line usage
magellai ask "Explain quantum physics" --stream

# REPL usage
magellai> /stream on
magellai> Explain quantum physics in detail
[streaming response...]

# One-off streaming in REPL
magellai> /ask --stream "Generate a long story"
```

This approach provides flexibility while maintaining consistency across interactive and non-interactive modes.

### REPL Integration with Workflows

The REPL seamlessly integrates with workflows and scripts:

```
magellai> /workflow run research "quantum computing"
[Running research workflow...]
[Step 1/5: Initial research...]
[Step 2/5: Gathering sources...]

magellai> /script run analysis.lua --data=results.json
[Executing Lua script...]
[Script output: Analysis complete, 15 insights found]

magellai> Can you summarize the insights from that analysis?
[Response: Based on the analysis script output, here are the key insights...]
```

### Error Handling in REPL

The REPL provides graceful error handling with helpful suggestions:

```
magellai> /modle gpt-4
Error: Unknown command '/modle'. Did you mean '/model'?

magellai> /context add nonexistent.pdf
Error: File 'nonexistent.pdf' not found. 
Available files in current directory:
  - document.pdf
  - notes.txt
  - research.md
```

### REPL Shortcuts and Aliases

Common shortcuts for frequent operations:

```
# Quick provider switching
magellai> /o4    # Shorthand for /provider openai /model gpt-4
magellai> /c3    # Shorthand for /provider anthropic /model claude-3

# Quick context operations
magellai> /+file.txt   # Shorthand for /context add file.txt
magellai> /-file.txt   # Shorthand for /context remove file.txt

# Session shortcuts
magellai> /save        # Quick save with auto-generated name
magellai> //           # Repeat last command
```

## Configuration Management

Magellai implements a comprehensive configuration system with well-defined precedence.

### Configuration Command Pattern

Following established CLI patterns, Magellai uses a dedicated `config` command for configuration management:

```bash
# Common config operations
magellai config set default.provider openai
magellai config get default.model
magellai config list
magellai config edit
```

This pattern mirrors Git's comprehensive configuration model:

```bash
magellai config --list                    # View all settings
magellai config --global user.name "Val"  # Set with scope
magellai config --edit                    # Open in editor
```

### Configuration Hierarchy

Configuration sources in order of precedence (following standard patterns):

1. **Command-line flags**: Immediate override for specific invocation
2. **Environment variables**: `MAGELLAI_*` prefixed variables
3. **Project configuration**: `.magellai.yaml` in project directory
4. **User configuration**: `~/.config/magellai/config.yaml`
5. **System configuration**: `/etc/magellai/config.yaml`
6. **Default values**: Built into application

### Configuration Format

Magellai uses YAML as its primary configuration format:

```yaml
# Example ~/.config/magellai/config.yaml
default:
  provider: openai
  model: gpt-4
  temperature: 0.7

providers:
  openai:
    api_key: ${OPENAI_API_KEY}
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    
commands:
  ask:
    model: gpt-3.5-turbo
    temperature: 0.3
```

### Profile System

Magellai supports named profiles for different use cases:

```bash
# Use specific profile
magellai --profile work ask "Draft a professional email"

# List profiles
magellai config profiles
```

## Extensibility Model

Magellai's plugin architecture enables extending functionality without modifying core code.

### Plugin System Architecture

The plugin system uses a hybrid hook and middleware approach:

1. **Hook system**: Predefined extension points where plugins can inject functionality
   ```
   Core execution flow → Pre-hooks → Command execution → Post-hooks
   ```

2. **Middleware chain**: Request/response transformation pipeline
   ```
   Input → Plugin A → Plugin B → Core processing → Plugin B → Plugin A → Output
   ```

3. **Plugin registry**: Auto-discovery with explicit loading paths
   ```
   ~/.config/magellai/plugins (user plugins)
   /etc/magellai/plugins (system plugins)
   ```

### Plugin Types

Magellai supports several plugin categories:

- **Provider plugins**: Add support for different LLM providers
- **Command plugins**: Add new commands and subcommands
- **Prompt plugins**: Add specialized prompt templates
- **Transform plugins**: Add input/output transformations
- **Integration plugins**: Add connections to other tools and systems

### Extension API Stability

Magellai promises a stable plugin API by:

- **Semantic versioning**: Clear indication of breaking changes
- **Compatibility shims**: Support for older plugin APIs in newer versions
- **Deprecation periods**: Advance notice before removing features

## I/O Handling and Pipeline Integration

As a good Unix citizen, Magellai follows established patterns for I/O handling.

### Standard Streams Usage

- **stdin**: Primary input channel for text content
- **stdout**: Normal output channel for command results
- **stderr**: Used for errors, warnings, and progress information (including logging output)

Magellai is designed to be used anywhere in a pipeline:

- **Beginning**: `magellai generate report --topic "climate change" | less`
- **Middle**: `cat data.json | magellai extract summary | grep important`
- **End**: `find . -name "*.md" | xargs cat | magellai transform summarize`

### Data Formats and Transformation

Magellai supports multiple data formats:

- **Plain text**: Default for simple interactions
- **JSON**: For structured data exchange
- **Markdown**: For rich text output
- **Binary data**: Base64-encoded when necessary

### Stream Processing Model

Magellai processes data in streams rather than batches when possible:

- **Incremental processing**: Process input as it arrives
- **Progressive output**: Generate output as soon as it's available
- **Real-time streaming**: Support for token-by-token output

## Error Handling and Graceful Degradation

LLM interactions introduce unique error scenarios that Magellai handles gracefully.

### Error Categorization

Magellai uses descriptive error codes in defined ranges:

- **64-73**: Input errors (invalid prompts, format issues)
- **74-83**: Model errors (unavailable models, quota limits)
- **84-93**: Network errors (API failures, timeouts)
- **94-103**: Processing errors (parsing failures, unexpected responses)

### Error Reporting Approach

- **Machine-parseable format**: JSON error format in non-interactive mode
- **Human-friendly format**: Colored, detailed errors in interactive mode
- **Actionability**: Every error includes suggested resolution steps

### Graceful Degradation Strategies

Magellai implements multiple degradation strategies:

- **Model fallbacks**: Automatically downgrade to available models
- **Retry with backoff**: Intelligent retry for transient failures
- **Local fallbacks**: Switch to local models when remote APIs are unavailable
- **Partial results**: Return partial results with warnings when appropriate

## Conclusion

Magellai's vision is to bring the power of language models into the Unix ecosystem as a first-class citizen—respecting established patterns while acknowledging LLMs' unique characteristics. By combining Unix philosophy with proven CLI design patterns, Magellai provides a tool that feels familiar to CLI users while unlocking the potential of AI language models.

The design follows evolutionary patterns that have proven effective across decades of tool development: global flags for logging, configuration through dedicated commands, clear command hierarchies with limited nesting, flexible flag positioning, and graceful error handling. These practices create an interface that is both powerful for experts and approachable for newcomers.

Through thoughtful command design, consistent patterns, and a powerful plugin system, Magellai aims to become an essential tool for both casual users and power users of language models on the command line.