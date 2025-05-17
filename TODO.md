# Magellai Implementation TODO List

This document provides a detailed, phased implementation plan for the Magellai project following the library-first design approach.

**Current Status**: Phase 4 In Progress - Advanced REPL Features

## Phase 1: Core Foundation (Week 1) ✅

## Phase 2: Configuration and Command Foundation (Week 2) ✅

## Phase 3: CLI with Kong (Week 3) ✅

## Phase 4: Advanced REPL Features (Week 4)

### 4.1 Extended REPL Commands ✅

### 4.1.1 Fix logging and file attachment issues ✅

### 4.2 Advanced Session Features

### 4.2.1 Session Storage library abstraction
- [ ] Session library abstraction for future database and other storage
  - [x] Check for abstraction first and provide recommendation

#### 4.2.1.1 Interface and Filesystem Implementation ✅
    - [x] Create StorageBackend interface with all session operations 
    - [x] Implement FileSystemStorage backend maintaining current behavior
    - [x] Create factory for storage backends
    - [x] Add storage configuration support in config system
    - [x] Migrate SessionManager to use StorageBackend - remove backward compatibility requirement - replace functionality
    - [x] Fix history_test.go to use new storage abstraction pattern
    - [x] Update session commands to support abstract storage
    - [x] Add unit tests for interface and filesystem implementation
    - [x] Add integration tests 

#### 4.2.1.2 Database Support [add as optional compile time / build time feature to reduce dependency] ✅
    - [x] Ensure Schemas have multi-user/tenant support, default is current user
    - [x] Implement SQLiteStorage backend for local database
    - [x] Add build tags for optional database support
    - [x] Implement FTS5 fallback for systems without FTS5 support
    - [x] Add database-specific configuration options
    - [x] Update documentation for database setup
    - [x] Add performance benchmarks for database vs filesystem
    - [x] Update makefile targets, create new target for benchmarks ✅

### 4.2.1.3 default configs for session storage and cleanup ✅
    - [x] Default session storage should be filestore. ✅
    - [x] Make sure to check if sqllite is feature is compiled in/available when switching config to db/sqlite backend ✅
    - [x] refactor code so that storage is under pkg/storage or something like that so it makes sense ✅
    - [x] Removed obsolete session_filesystem.go file replaced by storage abstraction
    - [x] Created comprehensive tests for storage_manager.go and session_manager.go  
    - [x] Created comprehensive tests for adapter.go with 100% coverage

### 4.2.1.4 Fix domain layer and types
    - [ ] Create domain layer package structure
        - [ ] Create new package `pkg/session/` as the domain layer
        - [ ] Create `pkg/session/types.go` for shared domain types
        - [ ] Add package documentation explaining domain layer purpose
    
    - [ ] Identify and consolidate shared types
        - [ ] Move `SessionInfo` type from both packages to `pkg/session/types.go`
        - [ ] Move `SearchResult` type from both packages to `pkg/session/types.go`
        - [ ] Move `SearchMatch` type from both packages to `pkg/session/types.go`
        - [ ] Analyze `Session` type differences and create unified domain model
        - [ ] Create common `Message` type in domain layer
        - [ ] Decide on attachment strategy (use `llm.Attachment` or create domain type)
    
    - [ ] Refactor storage package
        - [ ] Remove duplicate types from `pkg/storage/types.go`
        - [ ] Import shared types from `pkg/session/`
        - [ ] Keep only storage-specific types (e.g., database models if needed)
        - [ ] Update storage interfaces to use domain types
        - [ ] Update all storage implementations (filesystem, sqlite) to use domain types
    
    - [ ] Refactor REPL package
        - [ ] Remove duplicate types from `pkg/repl/types.go`
        - [ ] Import shared types from `pkg/session/`
        - [ ] Keep only REPL-specific types (e.g., `Conversation`)
        - [ ] Update REPL code to use domain types
        - [ ] Analyze if `pkg/repl/adapter.go` is still needed after refactoring
    
    - [ ] Update type conversions
        - [ ] Eliminate unnecessary conversions between duplicate types
        - [ ] Simplify or remove adapter layer if no longer needed
        - [ ] Update `StorageManager` to work directly with domain types
        - [ ] Update `SessionManager` to work directly with domain types
    
    - [ ] Update tests for new structure
        - [ ] Create tests for domain types in `pkg/session/types_test.go`
        - [ ] Update storage package tests to use domain types
        - [ ] Update REPL package tests to use domain types
        - [ ] Remove or update adapter tests as needed
        - [ ] Ensure all existing tests still pass after refactoring
    
    - [ ] Documentation and cleanup
        - [ ] Update package documentation to reflect new architecture
        - [ ] Create architecture diagram showing domain/application/infrastructure layers
        - [ ] Document type ownership and responsibilities
        - [ ] Remove obsolete code and comments
        - [ ] Update README or contributing guide with new structure

### 4.2.2 Session Auto-save functionality
- [ ] Enhance session management:
  - [x] Auto-save functionality ✅
  - [x] Session export formats (JSON, Markdown) ✅
  - [x] Session search by content ✅
  - [ ] Session tags and metadata
  - [ ] Session branching/forking
  - [ ] Session merging


### 4.3 Error Handling & Recovery
- [ ] Ensure loglevels are implemented at the library level with cmd line just passing the argument through
  - [ ] set default loglevel to warn
- [ ] Implement robust error handling:
  - [ ] Graceful network error recovery
  - [ ] Provider fallback mechanisms
  - [ ] Session auto-recovery after crashes
  - [ ] Partial response handling
  - [ ] Rate limit handling
  - [ ] Context length management

### 4.4 REPL Integration with Unified Command System
- [ ] Connect REPL to unified command system:
  - [ ] Route REPL commands through command registry
  - [ ] Support both `/` and `:` command prefixes
  - [ ] Integrate with existing core commands (config, model, alias, etc.)
  - [ ] Maintain command history across modes
  - [ ] Support command aliases in REPL
  - [ ] Context preservation between commands

### 4.5 REPL UI Enhancements
- [ ] Improve REPL interface:
  - [ ] Tab completion for commands
  - [ ] Syntax highlighting for code blocks
  - [ ] ANSI color output when TTY
  - [ ] Non-interactive mode detection
  - [ ] Custom prompt themes
  - [ ] Progress indicators for streaming
  - [ ] Rich media rendering (images, tables)

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

### 8.1 CLI Help System Advanced Improvements
- [ ] Future improvements for CLI help system:
  - [ ] Add custom help formatter for more control
  - [ ] Integrate Kong help with core help system for unified behavior
  - [ ] Add support for hiding commands with --all flag

### 8.2 Additional Scripting Engines
- [ ] Goja (JavaScript) support
  - [ ] Runtime integration
  - [ ] API bindings
  - [ ] Example scripts
- [ ] Tengo scripting support
  - [ ] Runtime integration
  - [ ] API bindings
  - [ ] Example scripts

### 8.3 Go Plugin Support
- [ ] Native Go plugin interface
  - [ ] Plugin loading mechanism
  - [ ] API stability guarantees
  - [ ] Plugin SDK
  - [ ] Migration guide from binary plugins
  - [ ] Security considerations

### 8.5 Additional Session Storage Backends
#### 8.5.1 Additional Database Support
    - [ ] Implement PostgreSQLStorage backend for remote database
    - [ ] Add database connection pooling and retry logic
    - [ ] Add configuration options
    - [ ] Create database migration scripts
#### 8.5.2 Cloud Storage Support 
    - [ ] Implement S3Storage backend for object storage
    - [ ] Implement RedisStorage backend for in-memory cache
    - [ ] Add cloud authentication and credentials support
    - [ ] Implement storage middleware (compression, encryption)
    - [ ] Add multi-tier storage with caching layer
    - [ ] Create cloud deployment documentation
    - [ ] Add configuration options
#### 8.5.3 Advanced Features 
    - [ ] Add storage migration tool for backend switching
    - [ ] Implement storage health checks and monitoring
    - [ ] Add storage backup and restore functionality
    - [ ] Create storage performance optimization guide
    - [ ] Implement storage quota management
    - [ ] Add storage backend plugin architecture

### 8.4 Web Interface
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

### 8.5 Advanced REPL Features
- [ ] Enhanced REPL capabilities:
  - [ ] Syntax highlighting
  - [ ] Command history search
  - [ ] Vi/Emacs key bindings
  - [ ] Custom prompt themes
  - [ ] Auto-suggestions
  - [ ] Rich media rendering

### 8.6 Enterprise Features
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