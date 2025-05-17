# Domain Layer Architecture

## Current Architecture (Before Refactoring)

```mermaid
graph TB
    subgraph "Current Architecture - Type Duplication"
        CLI[CLI Commands]
        REPL[REPL Package]
        Storage[Storage Package]
        LLM[LLM Package]
        
        REPL_Types["REPL Types<br/>- Session<br/>- Message<br/>- SessionInfo<br/>- SearchResult"]
        Storage_Types["Storage Types<br/>- Session<br/>- Message<br/>- SessionInfo<br/>- SearchResult"]
        LLM_Types["LLM Types<br/>- Message<br/>- Attachment"]
        
        Adapter["Adapter<br/>Complex Type Conversions"]
        
        CLI --> REPL
        REPL --> REPL_Types
        REPL --> Adapter
        Adapter --> Storage_Types
        Storage --> Storage_Types
        LLM --> LLM_Types
        REPL --> LLM_Types
        
        style REPL_Types fill:#ffcccc
        style Storage_Types fill:#ffcccc
        style LLM_Types fill:#ffcccc
        style Adapter fill:#ffffcc
    end
```

## Proposed Architecture (After Refactoring)

```mermaid
graph TB
    subgraph "Proposed Architecture - Domain Layer"
        CLI[CLI Commands]
        
        subgraph "Application Layer"
            REPL[REPL Package]
            Commands[Command Package]
        end
        
        subgraph "Domain Layer"
            Domain["Domain Types<br/>- Session<br/>- Message<br/>- Conversation<br/>- Attachment<br/>- SearchResult<br/>- Provider/Model"]
        end
        
        subgraph "Infrastructure Layer"
            Storage[Storage Package]
            LLM[LLM Package]
            Config[Config Package]
        end
        
        CLI --> Commands
        Commands --> REPL
        REPL --> Domain
        Storage --> Domain
        LLM --> Domain
        Config --> Domain
        
        style Domain fill:#ccffcc
        style REPL fill:#ccccff
        style Storage fill:#ccccff
        style LLM fill:#ccccff
    end
```

## Type Ownership Matrix

| Type | Current Owner(s) | Proposed Owner | Layer |
|------|-----------------|----------------|-------|
| Session | repl, storage | domain | Domain |
| SessionInfo | repl, storage | domain | Domain |
| Message | repl, storage, llm | domain | Domain |
| Attachment | storage, llm | domain | Domain |
| SearchResult | repl, storage | domain | Domain |
| SearchMatch | repl, storage | domain | Domain |
| Conversation | repl | domain | Domain |
| Provider/Model | llm, models | domain | Domain |

## Dependency Flow

### Before Refactoring
```
CLI → REPL ←→ Adapter ←→ Storage
     ↓
    LLM
```

### After Refactoring
```
CLI → REPL → Domain ← Storage
              ↑
             LLM
```

## Package Relationships

```mermaid
graph LR
    subgraph "Presentation"
        CMD[cmd/magellai]
    end
    
    subgraph "Application"
        COMMAND[pkg/command]
        REPL[pkg/repl]
    end
    
    subgraph "Domain"
        DOMAIN[pkg/domain]
    end
    
    subgraph "Infrastructure"
        LLM[pkg/llm]
        STORAGE[pkg/storage]
        CONFIG[pkg/config]
        MODELS[pkg/models]
    end
    
    CMD --> COMMAND
    COMMAND --> REPL
    REPL --> DOMAIN
    COMMAND --> DOMAIN
    LLM --> DOMAIN
    STORAGE --> DOMAIN
    CONFIG --> DOMAIN
    MODELS --> DOMAIN
    
    style DOMAIN fill:#90EE90
```

## Benefits of Domain Layer

1. **Single Source of Truth**: All business entities defined once
2. **Clear Boundaries**: Infrastructure depends on domain, not vice versa
3. **No Type Conversions**: Direct use of domain types across layers
4. **Better Testing**: Domain logic can be tested in isolation
5. **Easier Maintenance**: Changes to business entities happen in one place

## Migration Path

1. **Phase 1**: Create domain package with all types
2. **Phase 2**: Update storage to use domain types
3. **Phase 3**: Update REPL to use domain types
4. **Phase 4**: Update LLM to use domain types
5. **Phase 5**: Remove adapter layer and test everything

## Code Example

### Before (Duplication)
```go
// pkg/repl/types.go
type Session struct {
    ID           string
    Conversation *Conversation
    // ... fields
}

// pkg/storage/types.go
type Session struct {
    ID       string
    Messages []Message
    // ... similar fields
}

// pkg/repl/adapter.go
func ToStorageSession(replSession *Session) *storage.Session {
    // Complex conversion logic
}
```

### After (Domain Layer)
```go
// pkg/domain/session.go
type Session struct {
    ID           string
    Conversation *Conversation
    // ... single definition
}

// pkg/repl/session_manager.go
func (sm *SessionManager) Save(session *domain.Session) error {
    // Direct use of domain type
}

// pkg/storage/backend.go
type Backend interface {
    SaveSession(session *domain.Session) error
    // Direct use of domain type
}
```