# Domain Layer Implementation Plan

## Important Update: Package Naming

Based on the analysis, we should use `pkg/domain/` instead of `pkg/session/` as originally planned in TODO.md. This better reflects the scope of the package, which will contain all domain entities, not just session-related types.

## Implementation Phases

### Phase 1: Create Domain Package Structure

#### 1.1 Create Directory Structure
```bash
mkdir -p pkg/domain
```

#### 1.2 Create Domain Type Files
- `pkg/domain/session.go` - Session and SessionInfo types
- `pkg/domain/message.go` - Message and MessageRole types
- `pkg/domain/attachment.go` - Attachment and AttachmentType types
- `pkg/domain/conversation.go` - Conversation type
- `pkg/domain/search.go` - SearchResult and SearchMatch types
- `pkg/domain/provider.go` - Provider, Model, and ModelCapability types
- `pkg/domain/types.go` - Shared enums and constants
- `pkg/domain/doc.go` - Package documentation

### Phase 2: Implement Domain Types

#### 2.1 Session Domain (`pkg/domain/session.go`)
```go
package domain

import "time"

// Session represents a complete chat session with all its data
type Session struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name,omitempty"`
    Conversation *Conversation          `json:"conversation"`
    Config       map[string]interface{} `json:"config,omitempty"`
    Created      time.Time              `json:"created"`
    Updated      time.Time              `json:"updated"`
    Tags         []string               `json:"tags,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SessionInfo provides summary information about a session
type SessionInfo struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Created      time.Time `json:"created"`
    Updated      time.Time `json:"updated"`
    MessageCount int       `json:"message_count"`
    Model        string    `json:"model,omitempty"`
    Provider     string    `json:"provider,omitempty"`
    Tags         []string  `json:"tags,omitempty"`
}
```

#### 2.2 Message Domain (`pkg/domain/message.go`)
```go
package domain

import "time"

// Message represents a single message within a conversation
type Message struct {
    ID          string                 `json:"id"`
    Role        MessageRole            `json:"role"`
    Content     string                 `json:"content"`
    Timestamp   time.Time              `json:"timestamp"`
    Attachments []Attachment           `json:"attachments,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageRole represents the role of a message sender
type MessageRole string

const (
    MessageRoleUser      MessageRole = "user"
    MessageRoleAssistant MessageRole = "assistant"
    MessageRoleSystem    MessageRole = "system"
)
```

### Phase 3: Refactor Storage Package

#### 3.1 Update Storage Types
- Remove duplicate types from `pkg/storage/types.go`
- Import domain types: `import "github.com/lexlapax/magellai/pkg/domain"`
- Keep only storage-specific types (if any)

#### 3.2 Update Storage Interface
```go
package storage

import "github.com/lexlapax/magellai/pkg/domain"

type Backend interface {
    // Session operations using domain types
    CreateSession(session *domain.Session) error
    GetSession(id string) (*domain.Session, error)
    UpdateSession(session *domain.Session) error
    DeleteSession(id string) error
    ListSessions() ([]*domain.SessionInfo, error)
    
    // Search operations using domain types
    SearchSessions(query string) ([]*domain.SearchResult, error)
}
```

### Phase 4: Refactor REPL Package

#### 4.1 Update REPL Types
- Remove duplicate types from `pkg/repl/types.go`
- Import domain types
- Keep only REPL-specific types

#### 4.2 Simplify or Remove Adapter
- Analyze if adapter.go is still needed
- Remove conversion functions for domain types
- Keep only necessary infrastructure adaptations

### Phase 5: Update Tests

#### 5.1 Create Domain Tests
- `pkg/domain/session_test.go`
- `pkg/domain/message_test.go`
- `pkg/domain/conversation_test.go`

#### 5.2 Update Existing Tests
- Update storage package tests
- Update REPL package tests
- Fix integration tests

## Rollback Plan

If issues arise during implementation:

1. **Git Branch Protection**: Work on a feature branch
2. **Incremental Commits**: Commit after each successful phase
3. **Test Coverage**: Ensure tests pass before proceeding
4. **Rollback Points**: Tag stable states for easy rollback

## Success Metrics

1. **Zero Type Duplication**: No duplicate domain types across packages
2. **Test Coverage**: Maintain or improve test coverage
3. **Clean Architecture**: Clear separation of concerns
4. **Performance**: No performance regression
5. **Developer Experience**: Easier to understand and maintain

## Timeline

- **Day 1**: Create domain package and implement types
- **Day 2**: Refactor storage package
- **Day 3**: Refactor REPL package  
- **Day 4**: Update tests and documentation
- **Day 5**: Integration testing and cleanup

## Next Steps

1. Update TODO.md to reflect `pkg/domain/` instead of `pkg/session/`
2. Create the domain package structure
3. Begin implementing domain types
4. Proceed with refactoring in phases