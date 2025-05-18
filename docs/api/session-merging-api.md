# Session Merging API Reference

## Domain Types

### MergeType

```go
type MergeType int

const (
    MergeTypeContinuation MergeType = iota  // Append messages after merge point
    MergeTypeRebase                         // Replay all messages on target
    MergeTypeCherryPick                     // Select specific messages
)
```

### MergeOptions

```go
type MergeOptions struct {
    Type         MergeType    // Type of merge operation
    SourceID     string       // ID of source session
    TargetID     string       // ID of target session  
    MergePoint   int          // Index in target where merge begins
    MessageIDs   []string     // Message IDs for cherry-pick mode
    CreateBranch bool         // Whether to create new branch
    BranchName   string       // Name for new branch
}
```

### MergeResult

```go
type MergeResult struct {
    SessionID      string           // ID of merged session
    MergedCount    int              // Number of messages merged
    ConflictCount  int              // Number of conflicts
    Conflicts      []MergeConflict  // Conflict details
    NewBranchID    string           // ID of new branch (if created)
}
```

### MergeConflict

```go
type MergeConflict struct {
    Index        int      // Message index where conflict occurred
    SourceMsg    Message  // Message from source session
    TargetMsg    Message  // Message from target session
    Resolution   string   // How conflict was resolved
}
```

## Session Methods

### CanMerge

```go
func (s *Session) CanMerge(other *Session) error
```

Checks if this session can be merged with another.

**Parameters:**
- `other`: The session to merge with

**Returns:**
- `error`: nil if merge is possible, error describing why not

**Example:**
```go
if err := targetSession.CanMerge(sourceSession); err != nil {
    log.Printf("Cannot merge: %v", err)
}
```

### ExecuteMerge

```go
func (s *Session) ExecuteMerge(source *Session, options MergeOptions) (*Session, *MergeResult, error)
```

Performs the actual merge operation.

**Parameters:**
- `source`: Session to merge from
- `options`: Configuration for the merge

**Returns:**
- `*Session`: The merged session (may be new if branching)
- `*MergeResult`: Details about the merge
- `error`: Any error that occurred

**Example:**
```go
options := MergeOptions{
    Type:         MergeTypeContinuation,
    CreateBranch: true,
    BranchName:   "Merged Session",
}

mergedSession, result, err := targetSession.ExecuteMerge(sourceSession, options)
if err != nil {
    return err
}

log.Printf("Merged %d messages into %s", result.MergedCount, result.SessionID)
```

## Storage Backend Interface

### MergeSessions

```go
func MergeSessions(targetID, sourceID string, options MergeOptions) (*MergeResult, error)
```

Backend method to merge two sessions.

**Parameters:**
- `targetID`: ID of target session
- `sourceID`: ID of source session
- `options`: Merge configuration

**Returns:**
- `*MergeResult`: Merge operation results
- `error`: Any error that occurred

## REPL Commands

### /merge

```
/merge <source_session_id> [options]
```

**Options:**
- `--type <type>`: Merge type (continuation, rebase)
- `--create-branch`: Create new branch for merge
- `--branch-name <name>`: Name for new branch

**Examples:**
```bash
# Simple continuation merge
/merge session_1234567890

# Rebase merge with new branch
/merge session_1234567890 --type rebase --create-branch --branch-name "Rebased"

# Merge into current session
/merge session_other_branch --type continuation
```

## Code Examples

### Basic Merge

```go
// Load sessions
target, _ := storage.LoadSession(targetID)
source, _ := storage.LoadSession(sourceID)

// Configure merge
options := domain.MergeOptions{
    Type:     domain.MergeTypeContinuation,
    SourceID: source.ID,
    TargetID: target.ID,
}

// Execute merge
result, err := storage.MergeSessions(target.ID, source.ID, options)
if err != nil {
    return fmt.Errorf("merge failed: %w", err)
}

fmt.Printf("Merged %d messages\n", result.MergedCount)
```

### Branch Merge

```go
// Create branch during merge
options := domain.MergeOptions{
    Type:         domain.MergeTypeRebase,
    CreateBranch: true,
    BranchName:   "Feature Integration",
}

result, err := storage.MergeSessions(mainBranch, featureBranch, options)
if err != nil {
    return err
}

// Switch to new branch
newBranch, _ := storage.LoadSession(result.NewBranchID)
```

### Advanced Merge with Validation

```go
// Pre-merge validation
if err := target.CanMerge(source); err != nil {
    return fmt.Errorf("incompatible sessions: %w", err)
}

// Check for common ancestor
if ancestor := target.GetCommonAncestor(source); ancestor != nil {
    log.Printf("Common ancestor: %s", *ancestor)
}

// Execute merge with full options
options := domain.MergeOptions{
    Type:         domain.MergeTypeContinuation,
    MergePoint:   len(target.Conversation.Messages),
    CreateBranch: true,
    BranchName:   fmt.Sprintf("Merge_%s_%s", target.Name, source.Name),
}

merged, result, err := target.ExecuteMerge(source, options)
if err != nil {
    return err
}

// Handle results
if result.ConflictCount > 0 {
    for _, conflict := range result.Conflicts {
        log.Printf("Conflict at %d: %s", conflict.Index, conflict.Resolution)
    }
}
```

## Error Handling

Common errors and their meanings:

| Error | Description |
|-------|-------------|
| `cannot merge session with itself` | Attempting to merge a session with itself |
| `both sessions must have conversations` | One or both sessions lack conversation data |
| `invalid message index for branching` | MergePoint is out of bounds |
| `unsupported merge type` | Unknown MergeType value |
| `session not found` | Source or target session doesn't exist |

## Best Practices

1. **Always validate before merging**
   ```go
   if err := target.CanMerge(source); err != nil {
       return err
   }
   ```

2. **Use branches for safety**
   ```go
   options.CreateBranch = true
   ```

3. **Check results**
   ```go
   if result.MergedCount == 0 {
       log.Warn("No messages were merged")
   }
   ```

4. **Handle transactions in storage backends**
   ```go
   tx, _ := db.Begin()
   defer tx.Rollback()
   // ... merge operations ...
   tx.Commit()
   ```

5. **Log merge operations**
   ```go
   logging.LogInfo("Merging sessions", 
       "source", sourceID, 
       "target", targetID, 
       "type", options.Type)
   ```