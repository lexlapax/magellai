# Dependency Management in Magellai

This document outlines the approach to dependency management and package organization in the Magellai project.

## Package Structure and Dependencies

Magellai follows a layered architecture with the following key packages:

1. **Domain Package** (`pkg/domain`): Contains core domain models shared across the application
2. **Storage Package** (`pkg/storage`): Provides storage implementations for persisting data
3. **LLM Package** (`pkg/llm`): Adapts LLM providers for use with domain types 
4. **REPL Package** (`pkg/repl`): Implements the interactive chat interface
5. **Command Package** (`pkg/command`): Defines the command system for CLI and REPL
6. **UI Package** (`pkg/ui`): Provides user interface utilities
7. **Config Package** (`pkg/config`): Manages application configuration

## Dependency Flow

The intended dependency flow is:

```
domain <- storage <- repl <- command
                  ^
                  |
domain <- llm ----+
```

## Interface Packages

To avoid circular dependencies, we use interface packages that define contracts between components:

1. **`pkg/replapi`**: Defines interfaces for the REPL system that are used by both the REPL package and any package that needs to interact with the REPL (e.g., command).

## Intentional Coupling Points

The following are intentional coupling points in the codebase:

1. **Domain Types**: The `pkg/domain` package is used by most other packages as it contains the core domain models. This is an intentional dependency as these types represent the shared language of the application.

2. **REPL and Command Integration**: The `pkg/replapi` package defines the contract between the REPL system and the command system, allowing them to interact without creating circular dependencies.

3. **Storage Backend Registration**: Storage backends register themselves with the factory in `pkg/storage/factory.go` using the Go init pattern. This creates a coupling between the storage package and specific backends, but avoids needing direct imports in client code.

## Dependency Injection

Magellai uses dependency injection in several key areas:

1. **REPL Creation**: The `pkg/replapi` package provides a factory that creates REPL instances, decoupling the implementation from clients.

2. **Provider Creation**: LLM providers are created through factory functions in `pkg/llm/provider.go`.

3. **Command Registration**: Commands are registered with the command system, which allows for loose coupling between command implementations and their execution.

## Package Boundaries

When adding new code, consider the following guidelines:

1. Place domain models in `pkg/domain`
2. Use interface packages to break circular dependencies
3. Use dependency injection to reduce direct coupling
4. Avoid importing implementation packages; prefer importing interfaces

## Import Rules

1. Never import from a higher layer to a lower layer (e.g., don't import `pkg/repl` from `pkg/storage`)
2. Use interface packages when crossing layer boundaries
3. Prefer dependency injection over direct imports
4. Document any new coupling points

## Testing Dependencies

For testing, we allow more flexible import rules:

1. Test files can import from any package needed for testing
2. Test helper packages can import from implementation packages
3. Mock implementations can be provided for testing purposes

By following these guidelines, we maintain a clean architecture with minimal coupling between components.