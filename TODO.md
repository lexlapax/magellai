# Magellai Implementation TODO List

This document provides a detailed, phased implementation plan for the Magellai project following the library-first design approach.

**Current Status**: Phase 4.12 - Final validation and rollout

## Phase 4: Advanced REPL Features (Week 4)

### 4.8 Configuration - defaults, sample etc. ✅ (Completed)

### 4.9 Code abstraction and redundancy checks ✅ (Completed)

### 4.10 Manual test suite for cmd line (cli and repl both) ✅ (Completed)

### 4.11 Documentation and architecture updates ✅ (Completed)

### 4.12 Cmd line and repl improvements (UI and others)
- [x] API_KEYS - if no config file, use environment variables to read API Keys and use default provider and model after coming up with running configuration
- [ ] change `magellai config generate` to print out to stdout instead of writing to a file, 
  - [ ] change help and documentation accordingly


### 4.13 Final validation and rollout 🚧 (In Progress)
- [ ] Run full test suite
- [ ] Manual testing of all features
- [ ] Update CHANGELOG.md
- [ ] Create release notes
- [ ] Plan deployment strategy

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
  - [ ] Project-local config (`./.magellai.yaml`)
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

### 8.1 REPL/CLI UI Advanced features.
- [ ] Custom prompt themes (command/core functionality - cli and repl both)
- [ ] Syntax highlighting for code blocks (command/core functionality - cli and repl both)
- [ ] Progress indicators for streaming (command/core functionality - cli and repl both)
- [ ] Rich media rendering (images, tables) (command/core functionality - cli and repl both DEFER till tools/agents)

### 8.2 CLI Help System Advanced Improvements
- [ ] Future improvements for CLI help system:
  - [ ] Add custom help formatter for more control
  - [ ] Integrate Kong help with core help system for unified behavior
  - [ ] Add support for hiding commands with --all flag

### 8.3 Additional Scripting Engines
- [ ] Goja (JavaScript) support
  - [ ] Runtime integration
  - [ ] API bindings
  - [ ] Example scripts
- [ ] Tengo scripting support
  - [ ] Runtime integration
  - [ ] API bindings
  - [ ] Example scripts

### 8.4 Go Plugin Support
- [ ] Native Go plugin interface
  - [ ] Plugin loading mechanism
  - [ ] API stability guarantees
  - [ ] Plugin SDK
  - [ ] Migration guide from binary plugins
  - [ ] Security considerations

### 8.5 Web Interface
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

### 8.6 Advanced REPL Features
- [ ] Enhanced REPL capabilities:
  - [ ] Syntax highlighting
  - [ ] Command history search
  - [ ] Vi/Emacs key bindings
  - [ ] Custom prompt themes
  - [ ] Auto-suggestions
  - [ ] Rich media rendering

### 8.7 Enterprise Features
- [ ] Enterprise enhancements:
  - [ ] SAML/OIDC authentication
  - [ ] Audit logging
  - [ ] Usage analytics
  - [ ] Team collaboration
  - [ ] Policy management
  - [ ] Compliance tools

### 8.8 Additional Session Storage Backends
#### 8.8.1 Additional Database Support
    - [ ] Implement PostgreSQLStorage backend for remote database
    - [ ] Add database connection pooling and retry logic
    - [ ] Add configuration options
    - [ ] Create database migration scripts
#### 8.8.2 Cloud Storage Support 
    - [ ] Implement S3Storage backend for object storage
    - [ ] Implement RedisStorage backend for in-memory cache
    - [ ] Add cloud authentication and credentials support
    - [ ] Implement storage middleware (compression, encryption)
    - [ ] Add multi-tier storage with caching layer
    - [ ] Create cloud deployment documentation
    - [ ] Add configuration options
#### 8.8.3 Advanced Features 
    - [ ] Add storage migration tool for backend switching
    - [ ] Implement storage health checks and monitoring
    - [ ] Add storage backup and restore functionality
    - [ ] Create storage performance optimization guide
    - [ ] Implement storage quota management
    - [ ] Add storage backend plugin architecture

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