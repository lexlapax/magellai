// ABOUTME: Interactive REPL for chat sessions with LLMs
// ABOUTME: Handles user input, command execution, and session management

/*
Package repl implements the interactive chat environment for Magellai.

The REPL (Read-Eval-Print Loop) provides an interactive command-line interface
for conducting conversations with LLMs. It handles user input, command parsing,
message processing, and maintains conversation state across interactions.

Key Components:
  - REPL: The main interactive loop that processes user input
  - Conversation: Manages the state of an ongoing conversation
  - Commands: Built-in commands for session management and utility functions
  - CommandAdapter: Bridges between REPL commands and the unified command system
  - Session Management: Handles session loading, saving, branching, and merging
  - Non-Interactive Mode: Supports batch processing for automation

The REPL implements several advanced features:
  - Session branching for exploring alternative conversation paths
  - Session merging to combine conversations
  - File attachments for multimodal content
  - Auto-recovery from crashes
  - Command history and tab completion
  - Output colorization

Usage:
    // Create a new REPL instance
    repl, err := replapi.NewREPL(&replapi.REPLOptions{
        Config: config,
        Writer: os.Stdout,
        Reader: os.Stdin,
    })
    if err != nil {
        // Handle error
    }

    // Start the interactive loop
    err = repl.Run()

The REPL is designed to be extensible through the command system while
maintaining a clean separation from the core domain logic.
*/
package repl