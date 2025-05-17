# Domain Layer Refactoring Report

## Executive Summary

The Magellai codebase currently suffers from significant type duplication across multiple packages (`pkg/repl`, `pkg/storage`, `pkg/llm`), with identical or near-identical types defined in each package. This duplication creates maintenance overhead, increases the likelihood of bugs, and makes the codebase harder to understand. This report analyzes the current state and proposes a comprehensive refactoring strategy to introduce a proper domain layer.

## Current State Analysis

### 1. Type Duplication Issues

#### Duplicate Types Identified

| Type | Packages | Differences |
|------|-----------|-------------|
| `Session` | storage, repl | Storage has flattened fields, REPL has nested Conversation |
| `SessionInfo` | storage, repl | Identical structure |
| `SearchResult` | storage, repl | Identical structure |
| `SearchMatch` | storage, repl | Identical structure |
| `Message` | storage, repl, llm | Different attachment types |
| `Attachment` | storage, llm | Different field names and structure |

#### Adapter Pattern Overhead

The current architecture uses `pkg/repl/adapter.go` with numerous conversion functions:
- `ToStorageSession` / `FromStorageSession`
- `ToStorageMessage` / `FromStorageMessage`
- `ToStorageSearchResult` / `FromStorageSearchResult`
- `ToStorageSessionInfo` / `FromStorageSessionInfo`

These conversions add unnecessary complexity and potential points of failure.

### 2. Architectural Issues

1. **Missing Domain Layer**: No central package defining core business entities
2. **Infrastructure Coupling**: Business logic mixed with infrastructure concerns
3. **Type Ownership Confusion**: Unclear which package "owns" each type
4. **Conversion Overhead**: Excessive type conversions between layers

### 3. Impact on Development

- **Maintenance Burden**: Changes must be synchronized across multiple packages
- **Bug Risk**: Type conversions can introduce subtle bugs
- **Onboarding Difficulty**: New developers struggle to understand type relationships
- **Testing Complexity**: Tests must account for type conversions

## Proposed Solution: Domain Layer Introduction

### 1. New Package Structure

Create a new domain layer to house all business entities:

```
pkg/
├── domain/          # NEW: Core business entities
│   ├── session.go   # Session, SessionInfo
│   ├── message.go   # Message, MessageRole
│   ├── attachment.go # Attachment, AttachmentType
│   ├── conversation.go # Conversation
│   ├── search.go    # SearchResult, SearchMatch
│   ├── provider.go  # Provider, Model, ModelCapability
│   └── types.go     # Shared types and enums
├── storage/         # Infrastructure layer
├── repl/           # Application layer
└── llm/            # Infrastructure layer
```

### 2. Domain Types Design

```go
// pkg/domain/session.go
package domain

import "time"

type Session struct {
    ID           string
    Name         string
    Conversation *Conversation
    Created      time.Time
    Updated      time.Time
    Tags         []string
    Config       map[string]interface{}
    Metadata     map[string]interface{}
}

type SessionInfo struct {
    ID           string
    Name         string
    Created      time.Time
    Updated      time.Time
    MessageCount int
    Model        string
    Provider     string
    Tags         []string
}

// pkg/domain/message.go
type Message struct {
    ID          string
    Role        MessageRole
    Content     string
    Attachments []Attachment
    Timestamp   time.Time
    Metadata    map[string]interface{}
}

type MessageRole string

const (
    MessageRoleUser      MessageRole = "user"
    MessageRoleAssistant MessageRole = "assistant"
    MessageRoleSystem    MessageRole = "system"
)

// pkg/domain/conversation.go
type Conversation struct {
    ID           string
    Messages     []Message
    Model        string
    Provider     string
    Temperature  float64
    MaxTokens    int
    SystemPrompt string
    Created      time.Time
    Updated      time.Time
    Metadata     map[string]interface{}
}

// pkg/domain/attachment.go
type Attachment struct {
    ID       string
    Type     AttachmentType
    Content  []byte
    FilePath string
    Name     string
    MimeType string
    Size     int64
    URL      string
    Metadata map[string]interface{}
}

type AttachmentType string

const (
    AttachmentTypeImage AttachmentType = "image"
    AttachmentTypeFile  AttachmentType = "file"
    AttachmentTypeText  AttachmentType = "text"
    AttachmentTypeAudio AttachmentType = "audio"
    AttachmentTypeVideo AttachmentType = "video"
)

// pkg/domain/search.go
type SearchResult struct {
    Session *SessionInfo
    Matches []SearchMatch
}

type SearchMatch struct {
    Type     string
    Role     string
    Content  string
    Context  string
    Position int
}
```

### 3. Layer Responsibilities

#### Domain Layer (`pkg/domain/`)
- Defines core business entities
- Contains business logic and validation
- No dependencies on infrastructure
- Pure Go types and interfaces

#### Application Layer (`pkg/repl/`)
- Orchestrates use cases
- Uses domain entities
- Handles user interactions
- No type duplication

#### Infrastructure Layer (`pkg/storage/`, `pkg/llm/`)
- Implements technical capabilities
- Adapts external systems to domain
- Uses domain types directly
- Handles persistence and external APIs

## Implementation Strategy

### Phase 1: Create Domain Package (Week 1, Days 1-2)

1. Create `pkg/domain/` directory structure
2. Define all domain types in appropriate files
3. Add comprehensive documentation
4. Create unit tests for domain types

### Phase 2: Update Storage Package (Week 1, Days 3-4)

1. Import domain types in storage package
2. Remove duplicate type definitions
3. Update storage interfaces to use domain types
4. Update filesystem and SQLite implementations
5. Fix all storage tests

### Phase 3: Update REPL Package (Week 1, Days 4-5)

1. Import domain types in REPL package
2. Remove duplicate type definitions
3. Refactor conversation to use domain types
4. Remove or simplify adapter.go
5. Update all REPL tests

### Phase 4: Update LLM Package (Week 2, Days 1-2)

1. Analyze LLM message type usage
2. Create adapter for LLM-specific needs
3. Update provider interfaces
4. Fix all LLM tests

### Phase 5: Integration and Testing (Week 2, Days 3-5)

1. Run all tests and fix failures
2. Update integration tests
3. Verify end-to-end functionality
4. Update documentation
5. Create migration guide

## Risk Assessment and Mitigation

### Risks

1. **Breaking Changes**: Type changes may break existing code
   - *Mitigation*: Careful refactoring, comprehensive testing
   
2. **Test Failures**: Many tests will need updates
   - *Mitigation*: Update tests incrementally, maintain coverage
   
3. **Integration Issues**: External packages may have dependencies
   - *Mitigation*: Gradual migration, adapter patterns where needed

### Benefits

1. **Reduced Complexity**: Eliminate duplicate code and conversions
2. **Better Maintainability**: Single source of truth for types
3. **Clearer Architecture**: Well-defined layers and responsibilities
4. **Easier Testing**: Domain logic isolated from infrastructure
5. **Improved Developer Experience**: Clear type ownership

## Recommendations

1. **Immediate Action**: Begin Phase 1 to establish domain layer
2. **Incremental Migration**: Update packages one at a time
3. **Maintain Tests**: Keep test coverage high throughout
4. **Document Changes**: Update architecture documentation
5. **Review Checkpoints**: Review after each phase

## Conclusion

Introducing a proper domain layer will significantly improve the Magellai codebase by eliminating type duplication, clarifying architectural boundaries, and reducing maintenance overhead. The proposed refactoring strategy provides a systematic approach to migrate the existing code while maintaining functionality and test coverage.

The investment in this refactoring will pay dividends in:
- Reduced bugs from type conversions
- Faster feature development
- Easier onboarding for new developers
- Better testability and maintainability

This refactoring aligns with Domain-Driven Design principles and will position Magellai for sustainable growth and evolution.