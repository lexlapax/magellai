# CLI Framework Analysis for Magellai

This document analyzes various Go CLI frameworks to determine the best fit for Magellai's command-line interface implementation.

## Requirements

- Less dependencies
- Flexible
- Does not impose hard-to-get-around conventions
- Easy to test and read
- Completions support

## Framework Analysis

### 1. Cobra ⭐⭐⭐
- **Dependencies**: Heavy (includes Viper integration)
- **Flexibility**: Very high, but comes with complexity
- **Conventions**: Opinionated structure with init() functions
- **Testing**: Can be complex due to framework structure
- **Completions**: Excellent built-in support for bash, zsh, fish, PowerShell
- **Fit for Magellai**: Overkill given our existing command system

### 2. urfave/cli (v2) ⭐⭐⭐⭐
- **Dependencies**: Minimal
- **Flexibility**: Good balance of features and simplicity
- **Conventions**: Clean API, allows direct command definitions
- **Testing**: Straightforward with mock io.Writer
- **Completions**: Good support for shell completions
- **Fit for Magellai**: Good match for our needs

### 3. Kong ⭐⭐⭐⭐⭐
- **Dependencies**: Minimal
- **Flexibility**: Excellent, struct-based configuration
- **Conventions**: Uses Go structs, very Go-idiomatic
- **Testing**: Very easy to test due to struct-based approach
- **Completions**: Supports completions via kongplete
- **Fit for Magellai**: Excellent match, integrates well with existing structure

### 4. Standard library flag package ⭐⭐⭐⭐
- **Dependencies**: None (stdlib)
- **Flexibility**: Basic but sufficient
- **Conventions**: None, full control
- **Testing**: Simple and predictable
- **Completions**: Need to implement manually
- **Fit for Magellai**: Could work with our existing command system

### 5. Others (Kingpin, go-flags, docopt)
- **Kingpin**: Deprecated in favor of Kong
- **go-flags**: Less active development
- **docopt**: Too rigid for our needs

## Recommendation: Kong + kongplete

Kong is recommended for the following reasons:

1. **Minimal Dependencies**: Kong has very few dependencies
2. **Perfect Integration**: Kong's struct-based approach aligns perfectly with our existing command system
3. **Flexibility**: Easy to map existing command interfaces to Kong structs
4. **Easy Testing**: Struct-based commands are trivial to test
5. **Completions**: kongplete provides excellent completion support
6. **No Framework Lock-in**: Kong is more of a parser than a framework

### Why Kong fits Magellai

Kong's approach maps seamlessly to our existing command structure:

```go
// Our existing command interface maps perfectly to Kong
type CLI struct {
    Ask    AskCmd    `cmd:"" help:"One-shot query"`
    Chat   ChatCmd   `cmd:"" help:"Interactive chat mode"`
    Config ConfigCmd `cmd:"" help:"Configuration management"`
    // ... other commands
}

// Kong integrates seamlessly with our command executor
type AskCmd struct {
    Prompt      string   `arg:"" help:"The prompt to send"`
    Model       string   `flag:"" short:"m" help:"Model to use"`
    Attachments []string `flag:"" short:"a" help:"Files to attach"`
    Stream      bool     `flag:"" help:"Enable streaming"`
}

func (c *AskCmd) Run(ctx *Context) error {
    // Map to our ExecutionContext
    exec := &command.ExecutionContext{
        Args:  []string{c.Prompt},
        Flags: map[string]interface{}{
            "model":  c.Model,
            "attach": c.Attachments,
            "stream": c.Stream,
        },
    }
    return ctx.Executor.Execute(context.Background(), "ask", exec)
}
```

## Alternative: urfave/cli

If a more traditional approach is preferred, urfave/cli would be the second choice:
- Simple API
- Good completion support
- Active community
- Easy to integrate with existing code

## Why not Cobra?

While Cobra is powerful, it would require significant restructuring of our existing command system. Its init() pattern and framework approach would conflict with our clean command interface design.

## Implementation Plan with Kong

1. Add Kong dependency: `go get github.com/alecthomas/kong`
2. Add kongplete for completions: `go get github.com/willabides/kongplete`
3. Create CLI struct mapping to our commands
4. Implement Run() methods that call our command executor
5. Add completion generation commands
6. Maintain our existing command system as-is

## Decision

Based on this analysis, Kong was chosen as the CLI framework for Magellai due to its minimal dependencies, excellent integration with our existing command system, and clean struct-based approach that aligns with Go idioms.