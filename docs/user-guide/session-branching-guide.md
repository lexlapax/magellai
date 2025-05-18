# Session Branching User Guide

## What is Session Branching?

Session branching allows you to create alternative conversation paths from any point in your chat history. Think of it like creating a "save point" in a video game - you can explore different conversation directions without losing your original discussion.

## Why Use Branching?

- **Experimentation**: Try different approaches to the same problem
- **Comparison**: Compare responses to different phrasings
- **Safety**: Explore risky questions without affecting your main conversation
- **Organization**: Keep different topics in separate branches

## How to Use Branching

### Creating a Branch

To create a branch from your current conversation:

```
/branch experiment
```

This creates a branch at the end of your current conversation. To branch from a specific message:

```
/branch experiment at 5
```

This creates a branch starting after message #5.

### Viewing Branches

To see all branches of your current session:

```
/branches
```

Output example:
```
Branches of current session:
  main - Original Conversation (ID: session_001) - 15 messages
* experiment - Testing New Ideas (ID: session_002) - 8 messages
  api-questions - API Discussion (ID: session_003) - 12 messages
```

The `*` indicates your current branch.

### Viewing the Branch Tree

To see the hierarchical structure of all branches:

```
/tree
```

Output example:
```
Session Branch Tree:
Original Conversation (ID: session_001) - 15 messages
├─ Testing New Ideas (ID: session_002) - 8 messages *
│  └─ Deep Dive (ID: session_004) - 3 messages
└─ API Discussion (ID: session_003) - 12 messages
```

### Switching Between Branches

To switch to a different branch:

```
/switch session_003
```

This switches your active session to the "API Discussion" branch.

## Practical Examples

### Example 1: Exploring Different Solutions

```
User: How can I optimize my database queries?
AI: Here are several approaches...

# Create a branch to explore indexing
/branch indexing-approach

User: Tell me more about database indexing
AI: Database indexing works by...

# Create another branch for query optimization
/switch session_001  # Go back to original
/branch query-optimization

User: What about query optimization techniques?
AI: Query optimization involves...
```

### Example 2: Trying Different Prompts

```
User: Explain quantum computing
AI: Quantum computing is...

# Branch to try a simpler explanation
/branch simple-explanation at 1

User: Explain quantum computing to a 5-year-old
AI: Imagine you have magical coins...

# Branch to try a technical explanation
/switch session_001
/branch technical-explanation at 1

User: Explain quantum computing with mathematical formulas
AI: Quantum states are represented by |ψ⟩...
```

## Tips and Best Practices

1. **Name Branches Descriptively**: Use names that clearly indicate the purpose of each branch

2. **Branch Early**: Create branches before making significant direction changes in your conversation

3. **Regular Cleanup**: Review and delete branches you no longer need

4. **Document Branch Purpose**: Use the first message in a branch to document why you created it

5. **Use Branch Trees**: Regularly check `/tree` to understand your conversation structure

## Common Workflows

### Iterative Refinement

1. Start with a basic question
2. Branch to explore different aspects
3. Branch from the best responses to go deeper
4. Compare branches to find the best approach

### A/B Testing Prompts

1. Ask your initial question
2. Create multiple branches with different phrasings
3. Compare the quality of responses
4. Continue with the most effective approach

### Topic Organization

1. Use your main session for general discussion
2. Create branches for specific topics
3. Switch between branches as you change topics
4. Keep related discussions in the same branch

## Limitations

- Branches cannot be merged (yet)
- Very deep branch trees may impact performance
- Branch names cannot be changed after creation
- Deleting a parent session doesn't delete its branches

## Troubleshooting

**Q: I can't find my branch**
A: Use `/branches` to list all branches, or `/tree` to see the full structure

**Q: How do I know which branch I'm in?**
A: The current branch is marked with `*` in `/branches` output

**Q: Can I undo a branch creation?**
A: You can delete unwanted branches using `/delete <session_id>`

**Q: What happens to branches if I delete the parent?**
A: Branches remain accessible but become "orphaned" - they still work but aren't shown in the tree