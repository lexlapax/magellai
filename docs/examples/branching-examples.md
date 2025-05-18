# Session Branching Examples

This document provides practical examples of using the session branching feature in magellai.

## Basic Branching

### Example 1: Simple Branch Creation

```bash
# Start a conversation
$ magellai chat
> Hello, I need help with programming
AI: I'd be happy to help with programming! What language or topic are you interested in?

> I'm working with Python
AI: Python is a great choice! What specifically would you like to know about Python?

# Create a branch to explore web development
> /branch web-dev
Created branch 'web-dev' (ID: session_1234) at message 4
To switch to this branch, use: /switch session_1234

> Tell me about Python web frameworks
AI: Python has several popular web frameworks including Django, Flask, and FastAPI...

# Go back to main conversation and create another branch
> /switch session_main
Switched to branch 'main' (ID: session_main)

> /branch data-science at 2
Created branch 'data-science' (ID: session_5678) at message 2

> Tell me about Python for data science
AI: Python is extremely popular in data science. Key libraries include NumPy, Pandas...
```

### Example 2: Using Branch Tree

```bash
# View your conversation structure
> /tree
Session Branch Tree:
main - Initial Conversation (ID: session_main) - 4 messages
├─ web-dev - Web Development (ID: session_1234) - 6 messages *
│  └─ django-specific - Django Deep Dive (ID: session_9012) - 3 messages
└─ data-science - Data Science Path (ID: session_5678) - 8 messages

# List branches with details
> /branches
Branches of current session:
  main - Initial Conversation (ID: session_main) - 4 messages
* web-dev - Web Development (ID: session_1234) - 6 messages
  data-science - Data Science Path (ID: session_5678) - 8 messages
```

## Advanced Usage

### Example 3: A/B Testing Prompts

```bash
# Original question
> Explain machine learning
AI: Machine learning is a subset of artificial intelligence...

# Create branches to test different prompt styles
> /branch technical-explanation at 1
> Explain machine learning using technical terms and mathematical concepts
AI: Machine learning encompasses algorithms that optimize parameters θ through...

> /switch session_main
> /branch simple-explanation at 1
> Explain machine learning like I'm 5 years old
AI: Imagine you're teaching a robot to recognize cats and dogs...

> /switch session_main
> /branch practical-explanation at 1
> Explain machine learning with real-world examples
AI: Machine learning is like how Netflix recommends movies...
```

### Example 4: Iterative Refinement

```bash
# Start with a code request
> Write a Python function to sort a list
AI: Here's a simple sorting function...

# Branch to explore different approaches
> /branch bubble-sort
> Write a bubble sort implementation
AI: Here's bubble sort in Python...

> /switch session_main
> /branch quick-sort at 1
> Write a quicksort implementation
AI: Here's quicksort in Python...

> /switch session_main
> /branch optimize at 2
> Can you optimize this for large lists?
AI: For large lists, here are optimization strategies...

# Create sub-branch for specific optimization
> /branch parallel-processing
> How would I parallelize this sorting?
AI: You can use multiprocessing for parallel sorting...
```

## Workflow Examples

### Example 5: Research Workflow

```bash
# Main research question
> I need to understand quantum computing applications
AI: Quantum computing has several key application areas...

# Branch for different application domains
> /branch cryptography
> Tell me about quantum computing in cryptography
AI: Quantum computing poses both threats and opportunities for cryptography...

> /switch session_main
> /branch drug-discovery at 1
> How is quantum computing used in drug discovery?
AI: Quantum computing can simulate molecular interactions...

> /switch session_main
> /branch optimization at 1
> What optimization problems can quantum computing solve?
AI: Quantum computing excels at certain optimization problems...

# Deep dive into specific area
> /switch session_cryptography
> /branch shor-algorithm
> Explain Shor's algorithm in detail
AI: Shor's algorithm is a quantum algorithm for integer factorization...
```

### Example 6: Debugging Workflow

```bash
# Initial problem
> My Python script is throwing an AttributeError
AI: Can you share the error message and relevant code?

> [shares code and error]
AI: I see the issue. The error occurs because...

# Branch to explore different solutions
> /branch solution-1
> What if I check if the attribute exists first?
AI: Yes, you can use hasattr() to check...

> /switch session_main
> /branch solution-2 at 2
> Could I use a try-except block instead?
AI: Yes, here's how to handle it with exception handling...

> /switch session_main
> /branch refactor at 2
> Should I refactor this code to avoid the issue?
AI: Refactoring might be the best long-term solution...
```

## Best Practices Examples

### Example 7: Documentation Branches

```bash
# Main documentation discussion
> I need to document my API
AI: API documentation should include...

# Branch for different documentation styles
> /branch openapi
> Show me OpenAPI/Swagger documentation
AI: Here's how to document your API using OpenAPI...

> /switch session_main
> /branch markdown at 1
> Show me Markdown API documentation
AI: Here's a Markdown template for API docs...

> /switch session_main
> /branch docstrings at 1
> How should I write Python docstrings for my API?
AI: Python docstrings for APIs should follow...
```

### Example 8: Learning Branches

```bash
# Learning a new technology
> I want to learn React
AI: React is a JavaScript library for building user interfaces...

# Branch based on experience level
> /branch beginner
> I have no JavaScript experience
AI: Let's start with JavaScript basics first...

> /switch session_main
> /branch intermediate at 1
> I know JavaScript but not React
AI: Great! Let's dive into React concepts...

> /switch session_main
> /branch advanced at 1
> I know React basics, what about advanced patterns?
AI: Let's explore React hooks, context, and performance...

# Sub-branch for specific topics
> /branch hooks-deep-dive
> Explain useEffect in detail
AI: useEffect is React's way of handling side effects...
```

## Command Reference

- `/branch <name>` - Create branch at current point
- `/branch <name> at <n>` - Create branch at message n
- `/branches` - List all branches
- `/tree` - Show branch hierarchy
- `/switch <id>` - Switch to different branch

## Tips

1. Use descriptive branch names
2. Branch before major topic changes
3. Create sub-branches for detailed exploration
4. Use `/tree` to visualize conversation structure
5. Switch between branches to compare approaches
6. Regularly save important branches