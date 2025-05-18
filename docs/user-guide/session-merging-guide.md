# Session Merging Guide

This guide explains how to merge conversation sessions in magellai, allowing you to combine different conversational paths.

## Overview

Session merging allows you to combine two separate conversation sessions into one, which is particularly useful when:
- You have parallel conversations that need to be combined
- You want to merge branches back together
- You need to combine different exploration paths

## Merge Types

### 1. Continuation Merge
Appends messages from the source session after the target session's messages.

```bash
/merge <source_session_id>
```

### 2. Rebase Merge
Replays source messages on top of the target session, optionally from a specific point.

```bash
/merge <source_session_id> --type rebase
```

## Basic Usage

### Simple Merge
To merge another session into your current session:

```bash
# In a chat session
/merge session_1747543012345678
```

This appends all messages from the source session to your current conversation.

### Creating a Branch During Merge
To create a new branch instead of modifying the current session:

```bash
/merge session_1747543012345678 --create-branch --branch-name "Combined Ideas"
```

This creates a new session that contains the merged result, leaving both original sessions unchanged.

### Rebase Merge
To replay messages from a source session:

```bash
/merge session_1747543012345678 --type rebase
```

## Advanced Options

### Merge Types
- `continuation` (default): Appends messages after the current conversation
- `rebase`: Replays all source messages on top of the target

### Command Options
- `--type <type>`: Specify merge type (continuation or rebase)
- `--create-branch`: Create a new branch for the merge result
- `--branch-name <name>`: Name for the new branch (if creating one)

## Examples

### Example 1: Combining Research Sessions
```bash
# Start with a session about AI models
/new AI Research
> Tell me about transformer models
...

# In another terminal, research about applications
/new AI Applications  
> What are practical uses of transformers?
...

# Back in first session, merge the applications research
/merge session_applications --branch-name "AI Complete Research"
# Now you have both conversations combined
```

### Example 2: Merging Parallel Branches
```bash
# Working on main session
/save main-work
> Let's design a web API

# Create a branch for database design
/branch database-design
/switch session_branch_id
> Focus on database schema design
...

# Create another branch for frontend
/switch main-work  
/branch frontend-design
/switch session_branch_id
> Let's design the React components
...

# Merge both branches back
/switch main-work
/merge session_database_branch --create-branch --branch-name "Full Stack Design"
/switch session_fullstack_branch  
/merge session_frontend_branch
```

## Best Practices

1. **Use Branches**: When merging, consider creating a branch to preserve original sessions
2. **Name Your Merges**: Use descriptive branch names when creating merge branches
3. **Review Before Merging**: Check session contents with `/history` before merging
4. **Save Before Merging**: Always save important sessions before merge operations

## Merge Conflicts

Currently, magellai performs simple merges without conflict detection. Messages are combined based on the merge type:
- **Continuation**: Simply appends messages
- **Rebase**: Replays messages in order

## Related Commands

- `/branch`: Create a new branch from current session
- `/branches`: List all branches
- `/tree`: View branch hierarchy
- `/switch`: Switch between branches
- `/save`: Save current session
- `/history`: View conversation history

## Tips

1. Use `/tree` to visualize your session structure before merging
2. Create descriptive branch names to track merge purposes
3. Use `/export` to backup sessions before complex merges
4. Consider merge direction - which session should be the base?

## Troubleshooting

### "Cannot merge session with itself"
This error occurs when trying to merge a session into itself. Ensure you're using different session IDs.

### "Session not found"
Verify the session ID using `/sessions` command to list all available sessions.

### Unexpected Results
- Check merge type - continuation vs rebase behave differently
- Verify you're in the correct target session before merging
- Use `/history` to review the merged conversation