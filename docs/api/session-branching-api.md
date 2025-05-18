# Session Branching API Reference

## Domain Types

### Session

Extended session type with branching support:

```go
type Session struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name,omitempty"`
    Conversation *Conversation          `json:"conversation"`
    Config       map[string]interface{} `json:"config,omitempty"`
    Created      time.Time              `json:"created"`
    Updated      time.Time              `json:"updated"`
    Tags         []string               `json:"tags,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
    
    // Branching fields
    ParentID     string                 `json:"parent_id,omitempty"`
    BranchPoint  int                    `json:"branch_point,omitempty"`
    BranchName   string                 `json:"branch_name,omitempty"`
    ChildIDs     []string               `json:"child_ids,omitempty"`
}
```

### SessionInfo

Extended session info with branch information:

```go
type SessionInfo struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Created      time.Time `json:"created"`
    Updated      time.Time `json:"updated"`
    MessageCount int       `json:"message_count"`
    Model        string    `json:"model,omitempty"`
    Provider     string    `json:"provider,omitempty"`
    Tags         []string  `json:"tags,omitempty"`
    
    // Branch information
    ParentID     string    `json:"parent_id,omitempty"`
    BranchName   string    `json:"branch_name,omitempty"`
    ChildCount   int       `json:"child_count,omitempty"`
    IsBranch     bool      `json:"is_branch,omitempty"`
}
```

### BranchTree

Hierarchical branch structure:

```go
type BranchTree struct {
    Session  *SessionInfo  `json:"session"`
    Children []*BranchTree `json:"children,omitempty"`
}
```

## Methods

### Session Methods

#### CreateBranch

Creates a new branch from the current session.

```go
func (s *Session) CreateBranch(branchID string, branchName string, messageIndex int) (*Session, error)
```

**Parameters:**
- `branchID`: Unique identifier for the new branch
- `branchName`: Human-readable name for the branch
- `messageIndex`: Index in the message history where branching occurs (0-based)

**Returns:**
- `*Session`: The newly created branch session
- `error`: Error if parameters are invalid

**Errors:**
- Returns error if `messageIndex` is negative or exceeds message count

**Example:**
```go
session := domain.NewSession("main-session")
// Add messages...

branch, err := session.CreateBranch("branch-1", "Experiment", 5)
if err != nil {
    return fmt.Errorf("failed to create branch: %w", err)
}
```

#### AddChild

Adds a child branch ID to the session.

```go
func (s *Session) AddChild(childID string)
```

**Parameters:**
- `childID`: ID of the child branch to add

**Notes:**
- Idempotent - adding the same child multiple times has no effect
- Automatically updates the session timestamp

#### RemoveChild

Removes a child branch ID from the session.

```go
func (s *Session) RemoveChild(childID string)
```

**Parameters:**
- `childID`: ID of the child branch to remove

**Notes:**
- Safe to call even if child doesn't exist
- Automatically updates the session timestamp

#### IsBranch

Checks if the session is a branch of another session.

```go
func (s *Session) IsBranch() bool
```

**Returns:**
- `true` if the session has a parent ID
- `false` if it's a root session

#### HasBranches

Checks if the session has any child branches.

```go
func (s *Session) HasBranches() bool
```

**Returns:**
- `true` if the session has child IDs
- `false` if no branches exist

## Storage Backend Interface

### GetChildren

Retrieves all direct child branches of a session.

```go
func (b *Backend) GetChildren(sessionID string) ([]*domain.SessionInfo, error)
```

**Parameters:**
- `sessionID`: ID of the parent session

**Returns:**
- `[]*domain.SessionInfo`: List of child session summaries
- `error`: Error if session not found or retrieval fails

**Example:**
```go
children, err := storage.GetChildren("parent-session-id")
if err != nil {
    return fmt.Errorf("failed to get children: %w", err)
}

for _, child := range children {
    fmt.Printf("Branch: %s (%s)\n", child.Name, child.ID)
}
```

### GetBranchTree

Retrieves the complete branch tree starting from a session.

```go
func (b *Backend) GetBranchTree(sessionID string) (*domain.BranchTree, error)
```

**Parameters:**
- `sessionID`: ID of the root session

**Returns:**
- `*domain.BranchTree`: Hierarchical tree structure
- `error`: Error if session not found or tree building fails

**Example:**
```go
tree, err := storage.GetBranchTree("root-session-id")
if err != nil {
    return fmt.Errorf("failed to get tree: %w", err)
}

// Recursively process tree
var printTree func(*domain.BranchTree, string)
printTree = func(node *domain.BranchTree, prefix string) {
    fmt.Printf("%s%s\n", prefix, node.Session.Name)
    for _, child := range node.Children {
        printTree(child, prefix + "  ")
    }
}
printTree(tree, "")
```

## Integration Examples

### Creating a Branch Programmatically

```go
func createExperimentBranch(storage storage.Backend, sessionID string) error {
    // Load the parent session
    parent, err := storage.LoadSession(sessionID)
    if err != nil {
        return fmt.Errorf("failed to load parent: %w", err)
    }
    
    // Create branch at current message count
    branchID := fmt.Sprintf("branch_%d", time.Now().Unix())
    messageIndex := len(parent.Conversation.Messages)
    
    branch, err := parent.CreateBranch(branchID, "Experiment", messageIndex)
    if err != nil {
        return fmt.Errorf("failed to create branch: %w", err)
    }
    
    // Save both sessions
    if err := storage.SaveSession(parent); err != nil {
        return fmt.Errorf("failed to save parent: %w", err)
    }
    
    if err := storage.SaveSession(branch); err != nil {
        return fmt.Errorf("failed to save branch: %w", err)
    }
    
    return nil
}
```

### Finding All Branches of a Session

```go
func findAllBranches(storage storage.Backend, sessionID string) ([]*domain.SessionInfo, error) {
    tree, err := storage.GetBranchTree(sessionID)
    if err != nil {
        return nil, fmt.Errorf("failed to get tree: %w", err)
    }
    
    var branches []*domain.SessionInfo
    
    var collectBranches func(*domain.BranchTree)
    collectBranches = func(node *domain.BranchTree) {
        branches = append(branches, node.Session)
        for _, child := range node.Children {
            collectBranches(child)
        }
    }
    
    collectBranches(tree)
    return branches, nil
}
```

### Switching Active Branch

```go
func switchToBranch(storage storage.Backend, currentSession, targetBranchID string) error {
    // Save current session if needed
    current, err := storage.LoadSession(currentSession)
    if err != nil {
        return fmt.Errorf("failed to load current: %w", err)
    }
    
    if current.Updated.After(lastSaveTime) {
        if err := storage.SaveSession(current); err != nil {
            return fmt.Errorf("failed to save current: %w", err)
        }
    }
    
    // Load target branch
    branch, err := storage.LoadSession(targetBranchID)
    if err != nil {
        return fmt.Errorf("failed to load branch: %w", err)
    }
    
    // Update active session reference
    activeSession = branch
    
    return nil
}
```

## Error Handling

Common error scenarios and handling:

```go
// Invalid branch point
branch, err := session.CreateBranch(id, name, -1)
if err != nil {
    // Error: "invalid message index for branching"
}

// Branch from beyond message count
branch, err := session.CreateBranch(id, name, 999)
if err != nil {
    // Error: "invalid message index for branching"
}

// Non-existent session
children, err := storage.GetChildren("non-existent")
if err != nil {
    // Error: "session not found: non-existent"
}
```

## Performance Considerations

1. **Tree Depth**: Deep branch trees may impact `GetBranchTree` performance
2. **Child Count**: Sessions with many children may slow down tree operations
3. **Message Copying**: Branch creation copies messages, consider memory usage
4. **Recursive Operations**: Tree traversal is recursive, monitor stack usage

## Best Practices

1. **ID Generation**: Use timestamp-based or UUID for branch IDs
2. **Name Conventions**: Use descriptive, searchable branch names
3. **Cleanup**: Regularly prune unused branches
4. **Validation**: Always validate message indices before branching
5. **Persistence**: Save both parent and branch after creation
6. **Error Handling**: Handle all error cases gracefully
7. **Concurrency**: Synchronize branch operations in multi-threaded environments