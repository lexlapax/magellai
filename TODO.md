# Magellai Implementation TODO List

This document provides a detailed, phased implementation plan for the Magellai project following the library-first design approach.

## Phase 1: Core Foundation (Week 1)

### 1.1 Project Setup
- [x] Initialize Go module structure (`go mod init github.com/lexlapax/magellai`)
- [x] Add go-llms dependency (`go get github.com/lexlapax/go-llms@v0.2.1`)
- [x] Add go-llms as git submodule for source reference
- [x] Create directory structure:
  - [x] `cmd/magellai/` - CLI entry point
  - [x] `pkg/llm/` - LLM provider adapter/wrapper for go-llms
  - [x] `pkg/config/` - Configuration management
  - [x] `pkg/session/` - Session storage
  - [x] `internal/` - Internal utilities
  - [x] put build binaries in `bin/`
- [x] Create Makefile for builds, tests, and other tasks
- [x] Create a license file - use MIT license in root directory
- [x] Create a simple README.md on root directory with links to documentation under [docs]

### 1.2 Core Infrastructure
- [x] Create logging infrastructure using `slog`
- [x] Setup configuration directory structure
  - [x] `~/.config/magellai/` for general config
  - [x] `~/.config/magellai/sessions/` for session storage
  - [x] `~/.config/magellai/plugins/` for plugin installations

### 1.3 Core Data Models and go-llms Integration
- [x] Create wrapper types in `pkg/llm/types.go` that adapt go-llms types
  - [x] `Request` struct wrapping go-llms domain.Message
  - [x] `Response` struct wrapping go-llms domain.Response
  - [x] `PromptParams` mapping to go-llms domain.Option
  - [x] `Attachment` struct for multimodal content
  - [x] Define constants for providers and models

### 1.4 LLM Provider Adapter
- [x] Create `pkg/llm/provider.go` adapter interface for go-llms
  - [x] Wrap go-llms domain.Provider interface
  - [x] Adapt Generate/GenerateMessage methods
  - [x] Adapt Stream/StreamMessage methods
  - [x] Provider factory for OpenAI, Anthropic, Gemini
  - [x] Configuration helpers for API keys
  - [x] Model capability system with ModelInfo struct and capability flags (text, audio, video, image, file) 

### 1.5 Provider Implementations
- [x] Create provider adapters using go-llms
  - [x] OpenAI adapter using go-llms provider.OpenAI
  - [x] Anthropic adapter using go-llms provider.Anthropic
  - [x] Gemini adapter using go-llms provider.Gemini
  - [x] Mock provider for testing
  - [x] Unit tests for each provider

### 1.6 High-Level Ask Function
- [x] Implement `Ask()` function in `pkg/magellai.go`
  - [x] Use go-llms domain.Provider interface
  - [x] Provider/model selection logic (from config)
  - [x] Convert prompts to go-llms messages
  - [x] Response formatting
  - [x] Error handling with go-llms error types
  - [x] Comprehensive unit tests
  - [x] Support for streaming responses
  - [x] Support for multimodal attachments (AskWithAttachments)

## Phase 2: Configuration and Command Foundation (Week 2)

### 2.1 Configuration Management with Koanf
- [ ] Install koanf dependency (`go get github.com/knadh/koanf/v2`)
- [ ] Create `pkg/config/config.go` with koanf integration
  - [ ] Multi-layer configuration support:
    - [ ] Default values (embedded)
    - [ ] System config (`/etc/magellai/config.yaml`)
    - [ ] User config (`~/.config/magellai/config.yaml`)
    - [ ] Project config (`.magellai.yaml` - search upward)
    - [ ] Environment variables (`MAGELLAI_*`)
    - [ ] Command-line flags (highest precedence)
  - [ ] Profile system implementation
  - [ ] Configuration validation
  - [ ] Type-safe configuration access
  - [ ] Configuration watchers for live reload
  - [ ] Configuration merging strategies

### 2.2 Configuration Schema
- [ ] Define configuration structure in `pkg/config/schema.go`
  - [ ] Provider configurations (API keys, endpoints)
  - [ ] Model settings using `provider/model` format
  - [ ] Model-specific settings (temperature, max tokens)
  - [ ] Output preferences (format, colors)
  - [ ] Session storage settings
  - [ ] Plugin directories
  - [ ] Logging configuration
  - [ ] Profile definitions
  - [ ] Aliases for common commands
  - [ ] Model parsing utilities (split provider/model strings)

### 2.3 Configuration Utilities
- [ ] Implement configuration helpers in `pkg/config/utils.go`
  - [ ] GetString/GetInt/GetBool methods
  - [ ] SetValue with validation
  - [ ] Profile switching logic
  - [ ] Configuration export/import
  - [ ] Migration from old config formats
  - [ ] Environment variable mapping
  - [ ] Secret handling (API keys)
  - [ ] Configuration debugging tools

### 2.4 Unified Command System
- [ ] Create directory structure for unified command management `pkg/command`
- [ ] Design command registry system to be central
  - [ ] Command interface for all commands (CLI and REPL)
  - [ ] Command metadata (name, description, flags, availability)
  - [ ] Command execution context
  - [ ] Command validation and error handling
- [ ] Define command categories:
  - [ ] CLI-only commands (e.g., `ask`, `chat`)
  - [ ] REPL-only commands (e.g., `/reset`, `/exit`)
  - [ ] Shared commands (e.g., `model`, `config`)
  - [ ] Flag-to-command mapping for REPL (e.g., `--stream` becomes `/stream`)
- [ ] Create command discovery and registration mechanism
- [ ] Implement help system that works across CLI and REPL

### 2.5 Core Commands Implementation
- [ ] Implement shared commands in `pkg/command/core/`:
  - [ ] `model` - Switch between LLM models,
    - [ ] `model` should take argument of the form `<provider>/<modelname`>
    - [ ] this automatically switches `provider` - Switch between providers
  - [ ] `config` - Configuration management
  - [ ] `profile` - Profile management
  - [ ] `alias` - Alias management
  - [ ] `help` - Context-aware help
- [ ] Create command execution framework
- [ ] Add command validation and error handling
- [ ] Unit tests for each command

## Phase 3: CLI with Cobra (Week 3)

### 3.1 CLI Structure Setup
- [ ] Install Cobra dependency
- [ ] Create main.go in `cmd/magellai/`
- [ ] Define root command with global flags
- [ ] Implement version command
- [ ] Setup global flag parsing:
  - [ ] `--verbosity/-v` - Log verbosity level
  - [ ] `--output/-o` - Output format [text|json|markdown]
  - [ ] `--config/-c` - Config file to use
  - [ ] `--profile` - Configuration profile
  - [ ] `--no-color` - Disable color output
  - [ ] `--version` - Show version info

### 3.2 Ask Command
- [ ] Implement `ask` subcommand
  - [ ] Prompt as positional argument
  - [ ] Command-specific flags:
    - [ ] `--attach/-a` - File attachments (repeatable)
    - [ ] `--model/-m` - Provider/model selection
    - [ ] `--temperature/-t` - Model temperature
    - [ ] `--stream` - Enable streaming
    - [ ] `--format` - Response format hints
  - [ ] Pipeline support (stdin/stdout)
  - [ ] Integrate with unified command system

### 3.3 Chat Command
- [ ] Implement `chat` subcommand
  - [ ] Launch REPL mode
  - [ ] Profile selection
  - [ ] Session resume support (`--resume <id>`)
  - [ ] Initial attachments support
  - [ ] Pass control to REPL implementation

### 3.4 Configuration Commands (using koanf)
- [ ] Implement config subcommands:
  - [ ] `config set <key> <value>` - Set configuration value using koanf
  - [ ] `config get <key>` - Get configuration value via koanf
  - [ ] `config list` - List all settings from koanf
  - [ ] `config edit` - Open config in editor
  - [ ] `config validate` - Validate configuration
  - [ ] `config export` - Export current config
  - [ ] `config import <file>` - Import configuration
  - [ ] `config profiles list` - List profiles
  - [ ] `config profiles create <name>` - Create profile
  - [ ] `config profiles delete <name>` - Delete profile
  - [ ] `config profiles export <name>` - Export profile
  - [ ] `config profiles switch <name>` - Switch active profile

### 3.5 History Commands
- [ ] Implement history subcommands:
  - [ ] `history list` - List all sessions
  - [ ] `history show <id>` - Show session details
  - [ ] `history delete <id>` - Delete session
  - [ ] `history export <id> [--format=json]` - Export session
  - [ ] `history search <term>` - Search sessions

## Phase 4: Conversation, Session Storage & REPL (Week 4)

### 4.1 Conversation Management
- [ ] Create `pkg/conversation/conversation.go`
  - [ ] `Conversation` struct wrapping go-llms messages
  - [ ] Use go-llms domain.Message for history
  - [ ] `NewConversation()` constructor
  - [ ] `Send()` method using go-llms GenerateMessage
  - [ ] Context management with go-llms message roles
  - [ ] Streaming support via go-llms StreamMessage
  - [ ] Conversation state persistence
  - [ ] Token counting and management

### 4.2 Session Storage
- [ ] Implement session persistence in `pkg/session/`
  - [ ] Session storage abstraction for multiple formats
  - [ ] JSON-based storage format
  - [ ] Save/Load conversation methods
  - [ ] Session listing and searching
  - [ ] File-based storage in `~/.config/magellai/sessions/`
  - [ ] Auto-save functionality
  - [ ] Session metadata (timestamps, model used, token counts)
  - [ ] Unit tests for storage operations

### 4.3 REPL Implementation
- [ ] Create `pkg/repl/repl.go`
  - [ ] Interactive loop with prompt handling
  - [ ] Integrate with unified command system
  - [ ] Command mode (/) vs conversation mode
  - [ ] Multi-line input support
  - [ ] ANSI color output when TTY
  - [ ] Non-interactive mode detection
  - [ ] History support (arrow keys)
  - [ ] Tab completion for commands

### 4.4 Core REPL Commands
- [ ] Implement essential REPL commands:
  - [ ] `/help` - Show available commands
  - [ ] `:model <provider/name>` - Switch model
  - [ ] `:stream on|off` - Toggle streaming
  - [ ] `:verbosity int` - set verbosity
  - [ ] `:output <format>`
  - [ ] `:temperature`
  - [ ] `:max_tokens`
  - [ ] `:profile`
  - [ ] `:version`
  - [ ] `:attach <file>`
    - [ ] `:attach-remove <file>` - Remove attachment
    - [ ] `:attach-list` - List attachments
  - [ ] `:stream`
  - [ ] `:format` - response formats hints
  - [ ] `:config load|show` - load config or show path, different from /config
  - [ ] `/save [name]` - Save session
  - [ ] `/load <id>` - Load session
  - [ ] `/reset` - Clear conversation
  - [ ] `/exit` - Exit REPL /quit is equivalent to /exit, so is ^D
  - [ ] `/attach <file>` - Add attachment
  - [ ] `/config show` - Display current config
  - [ ] `/config set <key> <value>` - Set config value
  - [ ] `/profile <name>` - Switch profile
  - [ ] `/alias list` - list aliases
  - [ ] `/alias add <name> <command>` - create a new alias
  - [ ] `/alias remove <name>` - remove an alias



### 4.5 REPL Integration
- [ ] Connect REPL to conversation management
- [ ] Integrate with session storage for persistence
- [ ] Hook up command execution to unified command system
- [ ] Add context management across commands
- [ ] Implement streaming display in REPL
- [ ] Add error recovery and graceful degradation

## Phase 5: Plugin System (Week 5)

### 5.1 Plugin Architecture
- [ ] Design plugin interface in `pkg/plugin/`
  - [ ] Plugin metadata structure
  - [ ] Plugin lifecycle management
  - [ ] Plugin registry
  - [ ] Discovery mechanisms

### 5.2 Binary Plugin Support
- [ ] Implement binary plugin scanner
  - [ ] PATH scanning for `magellai-*` binaries
  - [ ] `~/.config/magellai/plugins/` directory support
  - [ ] Plugin metadata parsing
  - [ ] Name conflict resolution

### 5.3 Plugin Communication
- [ ] Define JSON-RPC protocol
  - [ ] Request/Response message format
  - [ ] Streaming event protocol
  - [ ] Error handling specification
  - [ ] Environment variable passing
  - [ ] Plugin capabilities negotiation

### 5.4 Plugin Execution
- [ ] Create plugin runner
  - [ ] Process spawning with stdin/stdout
  - [ ] JSON marshaling/unmarshaling
  - [ ] Timeout handling
  - [ ] Resource cleanup
  - [ ] Error recovery

### 5.5 Scripting Engine Interface
- [ ] Design scripting engine interface
  - [ ] Common interface for multiple engines
  - [ ] Tool creation support
  - [ ] Agent creation support
  - [ ] Workflow creation support
  - [ ] Error handling

### 5.6 Gopher-lua Integration
- [ ] Implement Gopher-lua scripting support
  - [ ] Lua runtime initialization
  - [ ] Go function bindings
  - [ ] Tool creation API
  - [ ] Agent creation API
  - [ ] Workflow creation API
  - [ ] Example Lua scripts

### 5.7 Sample Calculator Plugin
- [ ] Create `plugins/calculator/`
  - [ ] Math expression parser
  - [ ] JSON-RPC implementation
  - [ ] Build as `magellai-tool-calculator`
  - [ ] Documentation
  - [ ] Integration tests

### 5.8 Plugin Management Commands
- [ ] Implement plugin commands:
  - [ ] `plugin list` - List installed plugins
  - [ ] `plugin install <source>` - Install plugin
  - [ ] `plugin remove <name>` - Remove plugin
  - [ ] `plugin update [name]` - Update plugin(s)
  - [ ] `plugin info <name>` - Show plugin details

## Phase 6: Tools, Agents & Workflows (Week 6)

### 6.1 Tool Framework
- [ ] Create `pkg/tool/registry.go`
  - [ ] Wrap go-llms toolcall interface
  - [ ] Tool registration and discovery
  - [ ] Built-in tool support
  - [ ] Plugin tool integration
  - [ ] Tool validation

### 6.2 Tool Commands
- [ ] Implement tool commands:
  - [ ] `tool list` - List available tools
  - [ ] `tool run <name> [args]` - Execute tool
  - [ ] `tool info <name>` - Show tool details
  - [ ] `tool test <name>` - Test tool

### 6.3 Agent Framework
- [ ] Implement `pkg/agent/agent.go`
  - [ ] Use go-llms workflow.Agent as base
  - [ ] Extend go-llms llmutil.Agent wrapper
  - [ ] Tool invocation via go-llms toolcall
  - [ ] Multi-step execution
  - [ ] Context persistence
  - [ ] Agent templates

### 6.4 Workflow Engine
- [ ] Create `pkg/workflow/engine.go`
  - [ ] Wrap go-llms workflow.Workflow
  - [ ] YAML workflow parser
  - [ ] Workflow validation
  - [ ] Step dependencies
  - [ ] Error handling and retries
  - [ ] Progress tracking
  - [ ] Workflow templates

### 6.5 Agent & Workflow Commands
- [ ] Implement agent/workflow commands:
  - [ ] `agent list` - List available agents
  - [ ] `agent run <name> [args]` - Execute agent
  - [ ] `agent info <name>` - Show agent details
  - [ ] `workflow list` - List workflows
  - [ ] `workflow run <name> [args]` - Execute workflow
  - [ ] `workflow define <name>` - Define workflow
  - [ ] `workflow visualize <name>` - Show workflow graph
  - [ ] `workflow export <name>` - Export workflow

### 6.6 Built-in Agents
- [ ] Implement example agents:
  - [ ] Researcher agent - Web search and synthesis
  - [ ] Summarizer agent - Content summarization
  - [ ] Code analyzer agent - Code analysis and review
  - [ ] Q&A agent - Question answering
  - [ ] Integration tests for each

## Phase 7: Polish & Documentation (Week 7)

### 7.1 Shell Completion
- [ ] Generate completion scripts:
  - [ ] Bash completion
  - [ ] Zsh completion
  - [ ] Fish completion
  - [ ] PowerShell completion
- [ ] Add `completion` command
- [ ] Installation instructions
- [ ] Test on different shells

### 7.2 Advanced Configuration
- [ ] Implement configuration features:
  - [ ] Profile inheritance
  - [ ] Environment variable overrides
  - [ ] Project-local config (`.magellai.yaml`)
  - [ ] Config schema validation
  - [ ] Migration between versions
  - [ ] Config templates

### 7.3 Provider Enhancements
- [ ] Enhance provider support:
  - [ ] Ollama integration for local models
  - [ ] OpenRouter support
  - [ ] Provider fallback chains
  - [ ] Load balancing
  - [ ] Rate limiting
  - [ ] Cost tracking

### 7.4 Documentation
- [ ] Write comprehensive docs:
  - [ ] Getting Started guide
  - [ ] CLI reference (auto-generated)
  - [ ] Configuration guide
  - [ ] Plugin development guide
  - [ ] Scripting guide
  - [ ] API reference
  - [ ] Examples and tutorials
  - [ ] Troubleshooting guide

### 7.5 Testing & CI/CD
- [ ] Complete test coverage:
  - [ ] Unit tests (>80% coverage)
  - [ ] Integration tests for CLI
  - [ ] End-to-end tests
  - [ ] Benchmark tests
  - [ ] Fuzz testing
- [ ] Setup CI/CD:
  - [ ] GitHub Actions workflow
  - [ ] Automated testing
  - [ ] Code quality checks
  - [ ] Security scanning
  - [ ] Release automation
  - [ ] Cross-platform builds

### 7.6 Performance & Optimization
- [ ] Performance improvements:
  - [ ] Response caching
  - [ ] Connection pooling
  - [ ] Parallel provider queries
  - [ ] Memory optimization
  - [ ] Startup time optimization
  - [ ] Profile and benchmark

## Phase 8: Advanced Features (Post-MVP)

### 8.1 Additional Scripting Engines
- [ ] Goja (JavaScript) support
  - [ ] Runtime integration
  - [ ] API bindings
  - [ ] Example scripts
- [ ] Tengo scripting support
  - [ ] Runtime integration
  - [ ] API bindings
  - [ ] Example scripts

### 8.2 Go Plugin Support
- [ ] Native Go plugin interface
  - [ ] Plugin loading mechanism
  - [ ] API stability guarantees
  - [ ] Plugin SDK
  - [ ] Migration guide from binary plugins
  - [ ] Security considerations

### 8.3 Web Interface
- [ ] HTTP API server
  - [ ] RESTful endpoints
  - [ ] WebSocket support for streaming
  - [ ] Authentication/authorization
  - [ ] API documentation (OpenAPI)
- [ ] Web UI
  - [ ] Chat interface
  - [ ] Configuration management
  - [ ] Plugin management
  - [ ] Session history

### 8.4 Advanced REPL Features
- [ ] Enhanced REPL capabilities:
  - [ ] Syntax highlighting
  - [ ] Command history search
  - [ ] Vi/Emacs key bindings
  - [ ] Custom prompt themes
  - [ ] Auto-suggestions
  - [ ] Rich media rendering

### 8.5 Enterprise Features
- [ ] Enterprise enhancements:
  - [ ] SAML/OIDC authentication
  - [ ] Audit logging
  - [ ] Usage analytics
  - [ ] Team collaboration
  - [ ] Policy management
  - [ ] Compliance tools

## Development Guidelines

### Testing Strategy
- Write tests alongside implementation
- Use table-driven tests for comprehensive coverage
- Mock external dependencies (LLM APIs)
- Integration tests for all CLI commands
- E2E tests for critical user journeys
- Performance benchmarks for key operations
- Update Makefile with all test targets

### Code Organization
- Keep packages focused and single-purpose
- Use interfaces for extensibility
- Minimize circular dependencies
- Document all public APIs
- Follow Go best practices and idioms
- Use consistent error handling patterns

### Library-First Approach
- Keep all core logic in the library (`pkg/`)
- Front-ends (CLI/REPL) should only handle I/O and flag parsing
- Ensure library remains flag-free and testable
- Design APIs that can be consumed by multiple front-ends
- Maintain clean separation between library and UI code

### Commit Guidelines
- One feature per commit
- No automatic commits
- No attribution to external entities
- Clear, descriptive commit messages
- Reference issue numbers
- Keep commits atomic and focused

### Review Process
- Self-review before proposing changes
- Run all tests locally
- Update relevant documentation
- Check for breaking changes
- Update CHANGELOG.md
- Consider backward compatibility

### Release Process
- Semantic versioning
- Comprehensive release notes
- Migration guides for breaking changes
- Pre-release testing
- Cross-platform verification
- Documentation updates