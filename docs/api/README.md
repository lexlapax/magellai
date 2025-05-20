# Magellai API Documentation

## Introduction

This API documentation covers the programmatic interfaces provided by the Magellai library packages. These APIs allow developers to integrate Magellai's functionality into their own applications.

## Core APIs

Magellai follows a library-first design, meaning all core functionality is implemented in reusable Go libraries. These libraries can be imported and used independently of the CLI tool.

## Session Management API

### Session Branching

The [Session Branching API](session-branching-api.md) allows creating alternative conversation paths from a given point in a conversation history.

Key operations:
- Creating branches
- Managing branch relationships
- Navigating branch trees
- Listing branch information

For detailed documentation, see [Session Branching API](session-branching-api.md).

### Session Merging

The [Session Merging API](session-merging-api.md) enables combining multiple conversations into a single unified conversation.

Key operations:
- Merging sessions as continuations
- Rebasing sessions
- Cherry-picking messages
- Managing merge conflicts

For detailed documentation, see [Session Merging API](session-merging-api.md).

## Storage API

The storage API provides a unified interface for persisting sessions across different storage backends.

```go
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

## LLM Provider API

The LLM provider API offers a unified interface for interacting with multiple LLM providers.

```go
type Provider interface {
    Send(ctx context.Context, messages []domain.Message, options domain.ProviderOptions) (*domain.Message, error)
    SendStream(ctx context.Context, messages []domain.Message, options domain.ProviderOptions) (<-chan domain.StreamResponse, error)
    GetName() string
    GetID() string
    ListModels() []domain.Model
    GetModel(id string) (*domain.Model, error)
}
```

## REPL API

The REPL API provides programmatic access to the interactive REPL functionality.

```go
type REPL interface {
    Start(ctx context.Context) error
    Stop() error
    AddCommand(name string, handler CommandHandler)
    SetPrompt(prompt string)
    SetCompletionHandler(handler CompletionHandler)
    GetSession() *domain.Session
    SetSession(session *domain.Session)
}
```

## Command API

The command API enables creating custom commands for the CLI and REPL.

```go
type Command interface {
    Name() string
    Description() string
    Run(ctx *Context, args []string) error
    Help() string
}
```

## Usage Examples

### Example: Using Session Storage API

```go
package main

import (
    "fmt"
    "github.com/lexlapax/magellai/pkg/domain"
    "github.com/lexlapax/magellai/pkg/storage"
    "github.com/lexlapax/magellai/pkg/storage/filesystem"
)

func main() {
    // Create a filesystem storage backend
    backend, err := filesystem.NewFilesystemBackend("/path/to/storage")
    if err != nil {
        panic(err)
    }
    
    // Initialize backend
    if err := backend.Init(); err != nil {
        panic(err)
    }
    defer backend.Close()
    
    // Create a new session
    session := &domain.Session{
        ID:           "session-123",
        Name:         "Example Session",
        Conversation: &domain.Conversation{},
        Created:      time.Now(),
        Updated:      time.Now(),
    }
    
    // Add a message
    message := &domain.Message{
        ID:        "msg-1",
        Role:      domain.MessageRoleUser,
        Content:   "Hello, world!",
        Timestamp: time.Now(),
    }
    session.Conversation.AddMessage(message)
    
    // Save the session
    if err := backend.CreateSession(session); err != nil {
        panic(err)
    }
    
    // List all sessions
    sessions, err := backend.ListSessions()
    if err != nil {
        panic(err)
    }
    
    for _, s := range sessions {
        fmt.Printf("Session: %s (%s)\n", s.Name, s.ID)
    }
}
```

### Example: Using LLM Provider API

```go
package main

import (
    "context"
    "fmt"
    "github.com/lexlapax/magellai/pkg/domain"
    "github.com/lexlapax/magellai/pkg/llm"
)

func main() {
    // Create a provider
    provider, err := llm.NewProvider("openai", map[string]interface{}{
        "api_key": "your-api-key",
    })
    if err != nil {
        panic(err)
    }
    
    // Prepare messages
    messages := []domain.Message{
        {
            Role:    domain.MessageRoleUser,
            Content: "Hello, how are you?",
        },
    }
    
    // Provider options
    options := domain.ProviderOptions{
        Model:       "gpt-3.5-turbo",
        Temperature: 0.7,
        MaxTokens:   100,
    }
    
    // Send request
    ctx := context.Background()
    response, err := provider.Send(ctx, messages, options)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Response:", response.Content)
}
```

## Related Documentation

- [User Guide](../user-guide/README.md): Documentation for end-users
- [Technical Documentation](../technical/README.md): Implementation details
- [Architecture Overview](../technical/architecture.md): System architecture
- [Domain Layer](../technical/domain-layer-architecture.md): Domain-driven design