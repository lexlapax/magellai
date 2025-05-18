# Session Merging Architecture

## Overview

Session merging functionality allows combining two conversation sessions into one. This document describes the technical implementation and architecture.

## Domain Model

### Core Types

```go
// MergeType defines how sessions should be merged
type MergeType int

const (
    MergeTypeContinuation MergeType = iota  // Append messages
    MergeTypeRebase                         // Replay messages
    MergeTypeCherryPick                     // Select specific messages
)

// MergeOptions configures how sessions are merged
type MergeOptions struct {
    Type         MergeType
    SourceID     string
    TargetID     string
    MergePoint   int      // Message index in target
    MessageIDs   []string // For cherry-pick mode
    CreateBranch bool     // Create new branch for result
    BranchName   string   // Name for new branch
}

// MergeResult contains the result of a merge operation
type MergeResult struct {
    SessionID      string
    MergedCount    int
    ConflictCount  int
    Conflicts      []MergeConflict
    NewBranchID    string
}
```

## Architecture Layers

### 1. Domain Layer (`pkg/domain/session.go`)

Core merge logic implemented as methods on Session:

```go
// Check if merge is possible
func (s *Session) CanMerge(other *Session) error

// Prepare session for merge (validation + setup)
func (s *Session) PrepareForMerge(source *Session, options MergeOptions) (*Session, error)

// Execute the actual merge
func (s *Session) ExecuteMerge(source *Session, options MergeOptions) (*Session, *MergeResult, error)
```

### 2. Storage Layer (`pkg/storage/backend.go`)

Backend interface with merge operation:

```go
type Backend interface {
    // ... other methods ...
    MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error)
}
```

Implementations:
- Filesystem backend (`pkg/storage/filesystem/filesystem.go`)
- SQLite backend (`pkg/storage/sqlite/sqlite.go`)
- Mock backend for testing (`pkg/repl/mock_backend_test.go`)

### 3. REPL Layer (`pkg/repl/commands_branch.go`)

User-facing command implementation:

```go
func (r *REPL) cmdMerge(args []string) error
```

Handles:
- Command parsing
- Option processing
- User interaction
- Result display

## Merge Algorithms

### Continuation Merge
1. Load both sessions
2. Identify starting point in source
3. Clone messages from source
4. Append to target (or new branch)
5. Update metadata

```go
case MergeTypeContinuation:
    startIndex := 0
    if source.IsBranch() && source.ParentID == s.ID {
        startIndex = source.BranchPoint
    }
    
    for i := startIndex; i < len(source.Conversation.Messages); i++ {
        msg := source.Conversation.Messages[i]
        newMsg := msg.Clone()
        newMsg.ID = generateMessageID()
        mergeSession.Conversation.AddMessage(newMsg)
        result.MergedCount++
    }
```

### Rebase Merge
1. Load both sessions
2. Optionally truncate target at merge point
3. Clone all source messages
4. Append to target
5. Generate new IDs

```go
case MergeTypeRebase:
    if options.MergePoint > 0 && options.MergePoint < len(mergeSession.Conversation.Messages) {
        mergeSession.Conversation.Messages = mergeSession.Conversation.Messages[:options.MergePoint]
    }
    
    for _, msg := range source.Conversation.Messages {
        newMsg := msg.Clone()
        newMsg.ID = generateMessageID()
        mergeSession.Conversation.AddMessage(newMsg)
        result.MergedCount++
    }
```

## Branch Management

When `CreateBranch` is true:
1. Create new session with parent relationship
2. Copy target session properties
3. Execute merge into new session
4. Update parent's child list
5. Add merge metadata

## Data Flow

1. User issues `/merge` command
2. REPL parses command and options
3. StorageManager.MergeSessions called
4. Backend loads both sessions
5. Session.ExecuteMerge performs merge
6. Result saved to storage
7. User notified of result

## Transaction Support

SQLite backend uses transactions:
```go
tx, err := b.db.Begin()
defer tx.Rollback()
// ... merge operations ...
tx.Commit()
```

## Error Handling

### Validation Errors
- Cannot merge session with itself
- Sessions must have conversations
- Invalid message indices
- Unknown merge types

### Storage Errors
- Session not found
- Save failures
- Transaction failures

## Testing Strategy

### Unit Tests
- Domain merge logic (`session_merge_test.go`)
- Individual merge types
- Edge cases and validation

### Integration Tests
- Storage backend merge operations
- REPL command parsing
- End-to-end merge scenarios

### Mock Testing
- MockStorageBackend implements merge
- Isolated REPL testing
- Command validation

## Performance Considerations

- Messages are cloned, not referenced
- Large conversations may use significant memory
- SQLite uses transactions for consistency
- Consider streaming for very large merges

## Future Enhancements

1. **Conflict Detection**
   - Identify overlapping content
   - Merge conflict resolution UI
   - Three-way merge support

2. **Cherry-pick Mode**
   - Select specific messages
   - Message filtering
   - Interactive selection

3. **Merge Strategies**
   - Custom merge algorithms
   - ML-based conflict resolution
   - Semantic merging

4. **Performance**
   - Streaming merge for large sessions
   - Incremental saves
   - Background processing

5. **Undo Support**
   - Merge history tracking
   - Revert operations
   - Merge snapshots