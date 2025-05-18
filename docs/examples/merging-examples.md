# Session Merging Examples

## Basic Merge Operations

### Example 1: Simple Continuation Merge

```bash
# Start first session exploring Python
$ magellai chat
/save python-basics
> What are Python decorators?
[AI explains decorators]
> Show me an example
[AI provides example]
/save

# Start second session on advanced Python
$ magellai chat  
/save python-advanced
> Explain metaclasses in Python
[AI explains metaclasses]
> How do decorators relate to metaclasses?
[AI explains relationship]
/save

# Merge sessions together
/load python-basics
/merge python-advanced --create-branch --branch-name "Complete Python Guide"
# Now you have both basic and advanced topics in one session
```

### Example 2: Feature Branch Integration

```bash
# Main development session
$ magellai chat
/save main-api
> Let's design a REST API for a todo app
[AI provides initial design]
> Add user authentication endpoints
[AI adds auth endpoints]
/save

# Branch for database design
/branch db-design
> Focus on the database schema for this API
[AI designs schema]
> Add indexes for performance
[AI adds indexes]
/save

# Another branch for frontend
/switch main-api
/branch frontend-integration  
> Design React components for this API
[AI designs components]
/save

# Merge everything back
/switch main-api
/merge db-design --create-branch --branch-name "Full Stack Todo App"
/switch "Full Stack Todo App"
/merge frontend-integration
# Complete application design in one session
```

### Example 3: Research Consolidation

```bash
# Research on different topics
$ magellai chat
/save research-ml
> Explain transformer architecture
[AI explains transformers]
> Compare with LSTM networks
[AI provides comparison]
/save

# Separate research on applications
$ magellai chat
/save research-applications
> What are practical uses of transformers?
[AI lists applications]
> Focus on NLP applications
[AI details NLP uses]
/save

# Combine research
/load research-ml
/merge research-applications --branch-name "ML Research Complete"
/export markdown ml-research.md
```

## Advanced Merge Patterns

### Pattern 1: Sequential Development

```bash
# Initial specification
/new project-spec
> Define requirements for a chat application
[AI provides requirements]
/save

# Technical design phase
/new technical-design
> Design the architecture for these requirements
[AI creates architecture]
/save

# Implementation planning
/new implementation-plan
> Create implementation tasks for this architecture
[AI lists tasks]
/save

# Merge into complete project plan
/load project-spec
/merge technical-design --create-branch --branch-name "Project Phase 1"
/switch "Project Phase 1"
/merge implementation-plan
/export markdown project-plan.md
```

### Pattern 2: Parallel Exploration

```bash
# Explore option A
/new option-a
> Implement solution using Python Flask
[AI provides Flask implementation]
/save

# Explore option B  
/new option-b
> Implement solution using Node.js Express
[AI provides Express implementation]
/save

# Compare and merge
/new comparison
> Compare Flask vs Express for this use case
[AI provides comparison]
/merge option-a --create-branch --branch-name "All Options"
/switch "All Options"
/merge option-b
# Now have all options in one place for decision making
```

### Pattern 3: Iterative Refinement

```bash
# Initial draft
/new story-draft-1
> Write a short story about time travel
[AI writes story]
/save

# Feedback session
/new story-feedback
> What could improve this story?
[AI provides feedback]
> Suggest character development ideas
[AI suggests improvements]
/save

# Second draft incorporating feedback
/new story-draft-2
> Rewrite the story with these improvements
[AI creates improved version]
/save

# Merge all iterations
/load story-draft-1
/merge story-feedback --create-branch --branch-name "Story Development"
/switch "Story Development"
/merge story-draft-2
# Complete story development history
```

## Workflow Examples

### Academic Research Workflow

```bash
# Literature review
/new lit-review
> Summarize recent papers on quantum computing
[AI provides summaries]
/save

# Methodology development
/new methodology
> Design experimental methodology for quantum research
[AI designs methodology]
/save

# Results analysis
/new results
> Analyze these quantum computing results
[AI analyzes results]
/save

# Merge into paper
/load lit-review
/merge methodology --create-branch --branch-name "Research Paper"
/switch "Research Paper"
/merge results
/export markdown quantum-research.md
```

### Software Debugging Workflow

```bash
# Bug report
/new bug-report
> User reports app crashes on login
[AI analyzes report]
/save

# Investigation
/new investigation
> Check authentication service logs
[AI investigates]
> Found null pointer exception
/save

# Solution development
/new solution
> Fix null pointer in auth service
[AI provides fix]
> Add unit tests
[AI adds tests]
/save

# Merge into resolution
/load bug-report
/merge investigation --create-branch --branch-name "Bug Fix #123"
/switch "Bug Fix #123"
/merge solution
# Complete bug resolution with history
```

### Content Creation Workflow

```bash
# Blog post outline
/new blog-outline
> Create outline for AI trends in 2024
[AI creates outline]
/save

# Research section
/new blog-research
> Research current AI developments
[AI provides research]
/save

# Writing section
/new blog-writing
> Write introduction based on outline
[AI writes intro]
> Continue with main points
[AI writes content]
/save

# Merge into final post
/load blog-outline
/merge blog-research --create-branch --branch-name "AI Trends Blog Post"
/switch "AI Trends Blog Post"  
/merge blog-writing
/export markdown blog-post.md
```

## Best Practices

1. **Name your sessions clearly** - Use descriptive names for easy identification
2. **Save before merging** - Always save important work before merge operations
3. **Use branches for experiments** - Create branches when trying different approaches
4. **Document merge purpose** - Use clear branch names that explain the merge
5. **Export important merges** - Export merged sessions for backup or sharing

## Common Pitfalls

1. **Forgetting to save** - Unsaved changes may be lost during merge
2. **Wrong merge direction** - Consider which session should be the base
3. **Overwriting important sessions** - Use `--create-branch` to preserve originals
4. **Circular dependencies** - Avoid merging branches back into themselves
5. **Large session merges** - Very large sessions may take time to merge

## Tips for Effective Merging

- Use `/tree` to visualize your session structure before merging
- Create a "master" session for important projects
- Merge related work regularly to avoid drift
- Export merged sessions for team sharing
- Use descriptive commit messages in branch names