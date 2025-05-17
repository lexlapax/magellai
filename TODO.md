# Magellai Implementation TODO List

This document provides a detailed, phased implementation plan for the Magellai project following the library-first design approach.

## Phase 1: Core Foundation (Week 1) ✅

## Phase 2: Configuration and Command Foundation (Week 2) ✅

## Phase 3: CLI with Kong (Week 3) ✅

### 3.2 Ask Command - Partially Complete (REVISIT)
- [ ] Pipeline support (stdin/stdout)

### 3.2.1 CLI Help System Improvements - Partially Complete (REVISIT)
- [ ] Future improvements (if needed):
  - [ ] Add custom help formatter for more control
  - [ ] Integrate Kong help with core help system for unified behavior
  - [ ] Add support for hiding commands with --all flag

### 3.5 Logging and Verbosity Implementation - Partially Complete (REVISIT)
- [ ] Remaining subsections: 3.5.8, 3.5.9, 3.5.10

#### 3.5.1 Configuration Logging (pkg/config/) ✅

#### 3.5.2 LLM Provider Logging (pkg/llm/) ✅

#### 3.5.3 Session Management Logging (pkg/repl/) ✅

#### 3.5.4 Command Execution Logging (pkg/command/) ✅

#### 3.5.5 REPL Operations Logging (pkg/repl/) ✅

#### 3.5.6 File Operations Logging (internal/configdir/) ✅

#### 3.5.7 User-Facing Operations Logging ✅

#### 3.5.8 Performance and Metrics Logging ✅

#### 3.5.9 Security and Audit Logging
- [ ] API key usage (DEBUG - sanitized)
- [ ] Configuration modifications (INFO)
- [ ] File access attempts (DEBUG)
- [ ] Error conditions (ERROR)

#### 3.5.10 Testing and Integration
- [ ] Add logging tests to verify output
- [ ] Test different verbosity levels
- [ ] Ensure sensitive data is not logged
- [ ] Verify performance impact is minimal

### 3.6 History Commands
- [ ] Implement history subcommands:
  - [ ] `history list` - List all sessions
  - [ ] `history show <id>` - Show session details
  - [ ] `history delete <id>` - Delete session
  - [ ] `history export <id> [--format=json]` - Export session
  - [ ] `history search <term>` - Search sessions


## Phase 4: Advanced REPL Features (Week 4)

### 4.1 Extended REPL Commands
- [ ] Implement additional REPL commands in `pkg/repl/commands.go`:
  - [ ] `:model <provider/name>` - Switch model
  - [ ] `:stream on|off` - Toggle streaming
  - [ ] `:verbosity <level>` - Set verbosity
  - [ ] `:output <format>` - Set output format
  - [ ] `:temperature <value>` - Set temperature
  - [ ] `:max_tokens <value>` - Set max tokens
  - [ ] `:profile <n>` - Switch profile
  - [ ] `:attach <file>` - Add attachment
  - [ ] `:attach-remove <file>` - Remove attachment
  - [ ] `:attach-list` - List attachments
  - [ ] `:system` - System prompt (by itself is system show, with argument is system set)
  - [ ] `/config show` - Display current config
  - [ ] `/config set <key> <value>` - Set config value

### 4.2 Advanced Session Features
- [ ] Enhance session management:
  - [ ] Auto-save functionality
  - [ ] Session export formats (JSON, Markdown)
  - [ ] Session search by content
  - [ ] Session tags and metadata
  - [ ] Session branching/forking
  - [ ] Session merging

### 4.3 REPL UI Enhancements
- [ ] Improve REPL interface:
  - [ ] Tab completion for commands
  - [ ] Syntax highlighting for code blocks
  - [ ] ANSI color output when TTY
  - [ ] Non-interactive mode detection
  - [ ] Custom prompt themes
  - [ ] Progress indicators for streaming
  - [ ] Rich media rendering (images, tables)

### 4.4 REPL Integration with Unified Command System
- [ ] Connect REPL to unified command system:
  - [ ] Route REPL commands through command registry
  - [ ] Support both `/` and `:` command prefixes
  - [ ] Integrate with existing core commands (config, model, alias, etc.)
  - [ ] Maintain command history across modes
  - [ ] Support command aliases in REPL
  - [ ] Context preservation between commands

### 4.5 Error Handling & Recovery
- [ ] Implement robust error handling:
  - [ ] Graceful network error recovery
  - [ ] Provider fallback mechanisms
  - [ ] Session auto-recovery after crashes
  - [ ] Partial response handling
  - [ ] Rate limit handling
  - [ ] Context length management

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

- [ ] a utility to go to the provider websites, parse and create the models.json file
  - [ ] this could potentially be deferred and done as an inbuilt agent or workflow after we complete the workflow tasks below (deferred to Phase 6)

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