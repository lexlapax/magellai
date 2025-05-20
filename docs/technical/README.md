# Magellai Technical Documentation

## Introduction

This technical documentation provides in-depth information about the architecture, design decisions, and implementation details of the Magellai project. It's intended for developers who want to understand the internal workings of the system, contribute to the codebase, or build extensions.

## Core Architecture

- [**Architecture Overview**](architecture.md): High-level architecture and design principles
- [**Domain Layer Architecture**](domain-layer-architecture.md): Domain-driven design implementation
- [**Type Ownership**](type-ownership.md): Definitive ownership of types across packages
- [**Interface Contract Consistency**](interface-contract-consistency.md): Interface design principles

## Domain Implementation

- [**Domain Layer Implementation Plan**](domain-layer-implementation-plan.md): Original implementation roadmap
- [**Domain Layer Refactoring Report**](domain-layer-refactoring-report.md): Summary of domain refactoring

## Storage System

- [**Session Storage Abstraction**](session-storage-abstraction.md): Storage layer design
- [**Database Setup**](database-setup.md): SQLite integration details
- [**Storage Benchmark Results**](storage-benchmark-results.md): Performance comparison of storage backends

## Advanced Features

- [**Session Branching**](session-branching.md): Technical implementation of conversation branching
- [**Session Merging**](session-merging.md): Technical implementation of conversation merging

## Code Organization

- [**Dependency Management**](dependency-management.md): Strategies for managing dependencies
- [**Dependency Reduction**](dependency-reduction.md): Efforts to minimize external dependencies
- [**Error Handling Standardization**](error-handling-standardization.md): Error handling patterns

## CLI Implementation

- [**CLI Framework Analysis**](cli_framework_analysis.md): Evaluation of CLI framework options
- [**Color Refactoring**](color-refactoring-summary.md): Terminal color support implementation

## Interface Design

- [**Interface Documentation Analysis**](interface-documentation-analysis.md): Review of interface documentation
- [**Interface Signature Analysis**](interface-signature-analysis.md): Method signature consistency
- [**Interface Implementation Checks**](interface-implementation-checks.md): Compile-time interface checks
- [**Interface Consistency Summary**](interface-consistency-summary.md): Interface improvements
- [**Interface Consistency Implementation**](interface-consistency-implementation.md): Implementation details

## Package Structure

The Magellai codebase follows a clean package structure:

```
magellai/
├── cmd/               # CLI entry points
│   └── magellai/      # Main CLI application
├── internal/          # Private implementation details
│   ├── configdir/     # Configuration directory helpers
│   ├── logging/       # Logging utilities
│   └── testutil/      # Test utilities and mocks
└── pkg/               # Public library packages
    ├── command/       # Command implementation
    │   └── core/      # Core commands (ask, chat, etc.)
    ├── config/        # Configuration management
    ├── domain/        # Domain types and rules
    ├── llm/           # LLM provider integration
    ├── models/        # Model management
    ├── repl/          # REPL implementation
    │   └── session/   # Session management
    ├── replapi/       # REPL public interfaces
    ├── storage/       # Storage backends
    │   ├── filesystem/# File-based storage
    │   └── sqlite/    # SQLite storage
    ├── testutil/      # Test utilities
    └── ui/            # User interface helpers
```

## Key Interfaces

The system is built around several key interfaces that define the boundaries between components:

### Domain Interfaces

```go
// SessionRepository defines the contract for session storage operations
type SessionRepository interface {
    Create(session *Session) error
    Get(id string) (*Session, error)
    Update(session *Session) error
    Delete(id string) error
    List() ([]*SessionInfo, error)
    Search(query string) ([]*SearchResult, error)
    GetChildren(sessionID string) ([]*SessionInfo, error)
    GetBranchTree(sessionID string) (*BranchTree, error)
}

// ProviderRepository defines the contract for provider operations
type ProviderRepository interface {
    GetProvider(id string) (*Provider, error)
    ListProviders() ([]*Provider, error)
    GetModelByID(providerID, modelID string) (*Model, error)
}
```

### Storage Interface

```go
// Backend defines the interface for storage implementations
type Backend interface {
    // Basic operations
    Init() error
    Close() error
    
    // Session operations
    CreateSession(session *domain.Session) error
    GetSession(id string) (*domain.Session, error)
    UpdateSession(session *domain.Session) error
    DeleteSession(id string) error
    ListSessions() ([]*domain.SessionInfo, error)
    
    // Search operations
    SearchSessions(query string) ([]*domain.SearchResult, error)
    
    // Branch operations
    GetChildren(sessionID string) ([]*domain.SessionInfo, error)
    GetBranchTree(sessionID string) (*domain.BranchTree, error)
    
    // Merge operations
    MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error)
}
```

### LLM Provider Interface

```go
// Provider defines the interface for LLM provider implementations
type Provider interface {
    Send(ctx context.Context, messages []domain.Message, options domain.ProviderOptions) (*domain.Message, error)
    SendStream(ctx context.Context, messages []domain.Message, options domain.ProviderOptions) (<-chan domain.StreamResponse, error)
    GetName() string
    GetID() string
    ListModels() []domain.Model
    GetModel(id string) (*domain.Model, error)
}
```

### Command Interface

```go
// Command defines the interface for command implementations
type Command interface {
    Name() string
    Description() string
    Run(ctx *Context, args []string) error
    Help() string
}
```

## Testing Approach

The Magellai codebase follows a comprehensive testing strategy:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test interactions between components
3. **End-to-End Tests**: Test the entire system from CLI to storage
4. **Table-Driven Tests**: Use table-driven approach for thorough coverage
5. **Mocks and Stubs**: Use test doubles for external dependencies

## Future Development

Upcoming technical work includes:

1. **Plugin System**: Extensible architecture for plugins
2. **Tool Framework**: Integration with external tools
3. **Agent Framework**: Advanced multi-step reasoning
4. **Workflow Engine**: Composable LLM workflows
5. **HTTP API**: RESTful API for programmatic access

## Implementation Guidelines

When contributing to the Magellai codebase, follow these guidelines:

1. **Library-First Approach**: Core logic lives in `pkg/` packages
2. **Clean Architecture**: Maintain separation of concerns
3. **Type Ownership**: Respect domain layer as the source of truth
4. **Interface Contracts**: Follow consistent interface patterns
5. **Error Handling**: Use standard error handling patterns
6. **Testing**: Write comprehensive tests for all changes
7. **Documentation**: Update documentation to reflect changes