# Session Branching Documentation

## Overview

Session branching allows users to create alternative conversation paths from any point in their chat history. This feature enables experimentation with different conversation directions without losing the original conversation thread.

## Architecture

### Domain Layer

The branching functionality is implemented in the domain layer with the following key components:

#### Session Type Extensions

```go
type Session struct {
    // ... existing fields ...
    
    // Branching support
    ParentID     string   `json:"parent_id,omitempty"`     // ID of parent session if this is a branch
    BranchPoint  int      `json:"branch_point,omitempty"`  // Message index where branch occurred
    BranchName   string   `json:"branch_name,omitempty"`   // Optional name for this branch
    ChildIDs     []string `json:"child_ids,omitempty"`     // IDs of child branches
}
```

#### Key Methods

##### CreateBranch

Creates a new branch from the current session at a specified message index.

```go
func (s *Session) CreateBranch(branchID string, branchName string, messageIndex int) (*Session, error)
```

**Parameters:**
- `branchID`: Unique identifier for the new branch
- `branchName`: Human-readable name for the branch
- `messageIndex`: Index in the message history where the branch should diverge

**Returns:**
- `*Session`: The newly created branch session
- `error`: Error if the message index is invalid

**Usage Example:**
```go
parent := domain.NewSession("parent-1")
// Add some messages to parent...

branch, err := parent.CreateBranch("branch-1", "Alternative Path", 3)
if err != nil {
    log.Fatal(err)
}
```

##### Branch Management Methods

```go
// AddChild adds a child branch ID to this session
func (s *Session) AddChild(childID string)

// RemoveChild removes a child branch ID from this session
func (s *Session) RemoveChild(childID string)

// IsBranch returns true if this session is a branch of another session
func (s *Session) IsBranch() bool

// HasBranches returns true if this session has child branches
func (s *Session) HasBranches() bool
```

### Storage Layer

The storage backend interface includes branch-specific operations:

```go
type Backend interface {
    // ... existing methods ...
    
    // GetChildren returns all direct child branches of a session
    GetChildren(sessionID string) ([]*domain.SessionInfo, error)
    
    // GetBranchTree returns the full branch tree starting from a session
    GetBranchTree(sessionID string) (*domain.BranchTree, error)
}
```

#### BranchTree Structure

```go
type BranchTree struct {
    Session  *SessionInfo  `json:"session"`
    Children []*BranchTree `json:"children,omitempty"`
}
```

### REPL Commands

The following commands are available in the REPL interface:

#### /branch

Creates a new branch from the current session.

**Syntax:**
```
/branch <name> [at <message_index>]
```

**Parameters:**
- `name`: Name for the new branch
- `message_index`: Optional. Specific message to branch from (defaults to end)

**Example:**
```
> /branch experiment at 5
Created branch 'experiment' (ID: session_1234567890) at message 5
To switch to this branch, use: /switch session_1234567890
```

#### /branches

Lists all branches of the current session or its parent.

**Syntax:**
```
/branches
```

**Example:**
```
> /branches
Branches of current session:
  main-path - Main Session (ID: session_111) - 10 messages, created 2024-01-15 10:30
* experiment - Experimental Branch (ID: session_123) - 7 messages, created 2024-01-15 11:45
  test-branch - Testing Ideas (ID: session_456) - 3 messages, created 2024-01-15 12:00
```

#### /tree

Displays the branch tree for the current session.

**Syntax:**
```
/tree
```

**Example:**
```
> /tree
Session Branch Tree:
Main Session (ID: session_111) - 10 messages *
├─ Experimental Branch (ID: session_123) - 7 messages
│  └─ Sub-experiment (ID: session_789) - 2 messages
└─ Testing Ideas (ID: session_456) - 3 messages
```

#### /switch

Switches to a different branch.

**Syntax:**
```
/switch <branch_id>
```

**Example:**
```
> /switch session_456
Switched to branch 'Testing Ideas' (ID: session_456)
Branch of: parent session (ID: session_111)
Branched at message: 5
Messages: 3
```

## Implementation Details

### Branch Creation Process

1. Validates the message index is within bounds
2. Creates a new session with:
   - Copy of parent's configuration
   - Copy of parent's tags
   - Messages up to the branch point
   - Reference to parent session
3. Updates parent's child list
4. Saves both parent and branch sessions

### Tree Traversal

The `GetBranchTree` method recursively builds the tree structure:
1. Loads the root session
2. For each child ID, recursively calls `GetBranchTree`
3. Builds a hierarchical structure for visualization

### Data Persistence

Branch relationships are persisted in the session data:
- Parent sessions maintain a list of child IDs
- Branch sessions store their parent ID and branch point
- This allows reconstruction of the tree structure

## Edge Cases and Limitations

1. **Circular References**: The system prevents circular references by design
2. **Deep Nesting**: No hard limit on branch depth, but very deep trees may impact performance
3. **Orphaned Branches**: If a parent is deleted, branches become orphaned but remain accessible
4. **Concurrent Modifications**: Branch operations should be synchronized in multi-user scenarios

## Best Practices

1. **Naming Conventions**: Use descriptive branch names to track different experiment paths
2. **Regular Cleanup**: Periodically review and remove unused branches
3. **Branch Points**: Choose meaningful points in conversation to branch from
4. **Documentation**: Document the purpose of each branch for future reference

## Future Enhancements

1. **Branch Merging**: Combine branches back into parent sessions
2. **Branch Comparison**: Visual diff between branches
3. **Branch Templates**: Create template branches for common scenarios
4. **Branch Permissions**: Control who can create/modify branches in shared environments

## Related Functions

- `SessionManager.SaveSession()`: Persists branch changes
- `StorageManager.LoadSession()`: Retrieves branch sessions
- `Session.ToSessionInfo()`: Includes branch metadata in session summaries

## Testing

The implementation includes comprehensive tests:
- Unit tests for domain layer branch operations
- Integration tests for storage backend
- REPL command tests for user interactions

Example test case:
```go
func TestSession_CreateBranch(t *testing.T) {
    parent := NewSession("parent-1")
    // Add test messages...
    
    branch, err := parent.CreateBranch("branch-1", "Test Branch", 2)
    assert.NoError(t, err)
    assert.Equal(t, parent.ID, branch.ParentID)
    assert.Equal(t, 2, branch.BranchPoint)
    assert.Contains(t, parent.ChildIDs, branch.ID)
}
```