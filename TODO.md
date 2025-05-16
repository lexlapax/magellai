# Magellai Implementation TODO List

This document provides a detailed, phased implementation plan for the Magellai project following the library-first design approach.

## Phase 1: Core Foundation (Week 1) ✅

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

## Phase 2: Configuration and Command Foundation (Week 2) ⚙️

### 2.1 Configuration Management with Koanf ✅
- [x] Install koanf dependency (`go get github.com/knadh/koanf/v2`)
- [x] Create `pkg/config/config.go` with koanf integration
  - [x] Multi-layer configuration support:
    - [x] Default values (embedded)
    - [x] System config (`/etc/magellai/config.yaml`)
    - [x] User config (`~/.config/magellai/config.yaml`)
    - [x] Project config (`.magellai.yaml` - search upward)
    - [x] Environment variables (`MAGELLAI_*`)
    - [x] Command-line flags (highest precedence)
  - [x] Profile system implementation
  - [x] Configuration validation
  - [x] Type-safe configuration access
  - [x] Configuration watchers for live reload
  - [x] Configuration merging strategies

### 2.2 Configuration Schema ✅
- [x] Define configuration structure in `pkg/config/schema.go`
  - [x] Provider configurations (API keys, endpoints)
  - [x] Model settings using `provider/model` format
  - [x] Model-specific settings (temperature, max tokens)
  - [x] Output preferences (format, colors)
  - [x] Session storage settings
  - [x] Plugin directories
  - [x] Logging configuration
  - [x] Profile definitions
  - [x] Aliases for common commands
  - [x] Model parsing utilities (split provider/model strings)

### 2.3 Configuration Utilities ✅
- [x] Implement configuration helpers in `pkg/config/utils.go`
  - [x] GetString/GetInt/GetBool methods
  - [x] SetValue with validation
  - [x] Profile switching logic
  - [x] Configuration export/import
  - [ ] Migration from old config formats
  - [x] Environment variable mapping
  - [x] Secret handling (API keys)
  - [ ] Configuration debugging tools

### 2.4 Unified Command System ✅
- [x] Create directory structure for unified command management `pkg/command`
- [x] Design command registry system to be central
  - [x] Command interface for all commands (CLI and REPL - could be rest-api in the future)
  - [x] Command metadata (name, description, flags, availability)
  - [x] Command execution context
  - [x] Command validation and error handling
- [x] Define command categories:
  - [x] CLI-only commands (e.g., `ask`, `chat`)
  - [x] REPL-only commands (e.g., `/reset`, `/exit`)
  - [x] Shared commands (e.g., `model`, `config`)
  - [x] Flag-to-command mapping for REPL (e.g., `--stream` becomes `:stream`)
- [x] Create command discovery and registration mechanism
- [x] Implement help system that works across CLI and REPL

### 2.5 Core Commands Implementation ✅
- [x] Implement shared commands in `pkg/command/core/`:
  - [x] `model` - Switch between LLM models,
    - [x] `model` should take argument of the form `<provider>/<modelname`>
    - [x] this automatically switches `provider` - Switch between providers
  - [x] `config` - Configuration management
    - [x] Comprehensive subcommands (list, get, set, validate, export, import)
    - [x] Profile management (create, switch, delete, export)
    - [x] Full unit test coverage
    - [x] Fixed all linting errors (error checks)
  - [x] `profile` - Profile management
    - [x] Complete lifecycle management (create, switch, update, copy, delete)
    - [x] Profile export/import functionality
    - [x] Show current and specific profile details
    - [x] List all available profiles
    - [x] Full unit test coverage with lifecycle tests
    - [x] Fixed test ordering issues for map comparisons
  - [x] `alias` - Alias management
    - [x] Add, remove, list, show, clear aliases
    - [x] Support for both CLI and REPL aliases
    - [x] Scope management (cli/repl/all)
    - [x] Export/import functionality
    - [x] Full unit test coverage
  - [x] `help` - Context-aware help
    - [x] Context-aware display for CLI vs REPL
    - [x] Command categorization
    - [x] Alias resolution
    - [x] Comprehensive unit tests
    - [x] Consolidated help functionality into core/help.go
    - [x] Removed old help files and tests
- [x] Create command execution framework
  - [x] Command executor with validation and error handling
  - [x] Pre/post execution hooks
  - [x] Argument and flag parsing with type validation
  - [x] Comprehensive unit tests
- [x] Add command validation and error handling
  - [x] Flag type validation
  - [x] Required flag checking
  - [x] Custom validation error types
  - [x] Contextual error messages
- [x] Unit tests for each command (model, config, profile, alias, and help commands complete)

### 2.6 Models static inventory file `models.json` ✅
- [x] A statically created `models.json` in root directory - this will/can be used for help and other things later
  - [x] version no (semantic versioning), and date as file metadata on top
  - [x] list of models by provider
  - [x] each model has name, description, url for model documentation/modelcard and a capability list, and last updated in models.json and other metadata
    - [x] capability list should be something like text and sub capability like read/consume, write/generate - possible capabilities are text, file, image, audio, video
- [x] Created pkg/models for loading and querying models.json
  - [x] Load inventory from root directory
  - [x] Query models by provider, name, capabilities
  - [x] Get models with specific capabilities
  - [x] List providers and model families
  - [x] Comprehensive unit tests
- [ ] a utility to go to the provider websites, parse and create the models.json file
  - [ ] this could potentially be deferred and done as an inbuilt agent or workflow after we complete the workflow tasks below (deferred to Phase 6)


## Phase 3: CLI with Kong (Week 3)

### 3.1 CLI Structure Setup ✅
- [x] Research best framework, since we have our own command structure - 
    - [x] Cobra, kong + kongplete, urfave/cli, Kingpin, go-flags, docopt,  
    - [x] criteria - less dependencies, flexible, does not impose hard to get around conventions, easy to test and read, completions support
    - [x] Decision: Kong + kongplete chosen (see docs/technical/cli_framework_analysis.md)
- [x] Install library dependency (Kong + kongplete)
- [x] Create main.go in `cmd/magellai/`
- [x] Define root command with global flags
- [x] Implement version command 
    - [x] to call the command core `version` command with context
- [x] Help command handled by Kong framework (core help command still available for REPL/API)
- [x] Setup global flag parsing: 
  - [x] `--verbosity/-v` - Log verbosity level
  - [x] `--output/-o` - Output format [text|json|markdown]
  - [x] `--configfile/-c` - Config file to use (different from `config` command)
  - [x] `--profile` - Configuration profile
  - [x] `--no-color` - Disable color output
  - [x] `--version` - Show version info (Unix standard flag)
  - [x] Also support version subcommand for advanced usage

### 3.2 Ask Command ✅
- [x] Implement `ask` subcommand
  - [x] Prompt as positional argument
  - [x] Command-specific flags:
    - [x] `--attach/-a` - File attachments (repeatable)
    - [x] `--model/-m` - Provider/model selection
    - [x] `--temperature/-t` - Model temperature
    - [x] `--stream` - Enable streaming
    - [x] `--format` - Response format hints
    - [x] `--max-tokens` - Maximum response tokens
    - [x] `--system/-s` - System prompt
  - [ ] Pipeline support (stdin/stdout)
  - [x] Integrate with unified command system
  - [x] Support global output flag (--output)
  - [x] Full multimodal attachment support
  - [x] Streaming response support
  - [x] Provider selection based on model

### 3.2.1 CLI Help System Improvements
- [ ] Customize Kong help display for progressive disclosure
  - [ ] Override Kong's default help formatter to show only top-level commands
  - [ ] Integrate with centralized help command from pkg/command/core/help.go
  - [ ] Implement custom help handling for `--help` flag
  - [ ] Ensure `magellai --help` shows only main commands
  - [ ] Make `magellai config --help` show config subcommands
  - [ ] Support nested help (e.g., `magellai config profiles --help`)
- [ ] Leverage centralized help command for consistency
  - [ ] Create KongHelpAdapter to bridge Kong help with our help system
  - [ ] Ensure help behavior is consistent between CLI and REPL
  - [ ] Support both `magellai help <command>` and `magellai <command> --help`
- [ ] Implement progressive disclosure pattern
  - [ ] Top-level shows only primary commands (ask, chat, config, etc.)
  - [ ] Subcommand help shows next level of options
  - [ ] Use command metadata to determine what to display at each level
- [ ] Update command registration to support help customization
  - [ ] Add HelpFormatter field to command metadata
  - [ ] Allow commands to specify custom help behavior
  - [ ] Ensure backward compatibility with existing commands
- [ ] Implementation approach:
  - [ ] Study Kong's help system and find extension points
  - [ ] Check if Kong supports custom help formatters or templates
  - [ ] Consider using Kong's BeforeApply hook to intercept help requests
  - [ ] Potentially use Kong's Help struct customization
  - [ ] Ensure solution works with Kong's built-in flag handling

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
  - [ ] `config profiles create <n>` - Create profile
  - [ ] `config profiles delete <n>` - Delete profile
  - [ ] `config profiles export <n>` - Export profile
  - [ ] `config profiles switch <n>` - Switch active profile

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
  - [ ] Session storage abstraction for multiple stores
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
  - [ ] `:configfile load|show` - load config or show path, different from /config
  - [ ] `/save [name]` - Save session
  - [ ] `/load <id>` - Load session
  - [ ] `/reset` - Clear conversation
  - [ ] `/exit` - Exit REPL /quit is equivalent to /exit, so is ^D
  - [ ] `/attach <file>` - Add attachment
  - [ ] `/config show` - Display current config
  - [ ] `/config set <key> <value>` - Set config value
  - [ ] `/profile <n>` - Switch profile
  - [ ] `/alias list` - list aliases
  - [ ] `/alias add <n> <command>` - create a new alias
  - [ ] `/alias remove <n>` - remove an alias



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
  - [ ] `plugin remove <n>` - Remove plugin
  - [ ] `plugin update [name]` - Update plugin(s)
  - [ ] `plugin info <n>` - Show plugin details

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
  - [ ] `tool run <n> [args]` - Execute tool
  - [ ] `tool info <n>` - Show tool details
  - [ ] `tool test <n>` - Test tool

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
  - [ ] `agent run <n> [args]` - Execute agent
  - [ ] `agent info <n>` - Show agent details
  - [ ] `workflow list` - List workflows
  - [ ] `workflow run <n> [args]` - Execute workflow
  - [ ] `workflow define <n>` - Define workflow
  - [ ] `workflow visualize <n>` - Show workflow graph
  - [ ] `workflow export <n>` - Export workflow

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