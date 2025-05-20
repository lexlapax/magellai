# Interface and Contract Consistency

This document outlines the findings and plan for improving interface consistency across the Magellai codebase, part of Phase 4.9.10.

## Overview

Interfaces are critical in Magellai's design as they provide clear contracts between components, allowing for loose coupling and easier testing. This phase aims to standardize interfaces across the codebase, ensure proper documentation, and enforce compile-time checks for implementations.

## Identified Interfaces

We identified the following key interfaces across the codebase:

### Command Package Interfaces
- `command.Interface`: Core command interface for all commands
- `command.discoverer`: Internal interface for command discovery

### Domain Package Interfaces
- `domain.SessionRepository`: Contract for session persistence
- `domain.ProviderRepository`: Contract for provider/model configuration

### Storage Package Interfaces
- `storage.Backend`: Interface for session storage implementations

### LLM Package Interfaces
- `llm.Provider`: Adapter interface wrapping go-llms providers
- `llm.DomainProvider`: Extends Provider with domain type support

## Current Issues

After reviewing these interfaces, we identified several areas for improvement:

1. **Inconsistent Documentation**: Some interfaces have thorough documentation, while others have minimal or missing docs.

2. **Missing Compile-Time Checks**: Not all interfaces have compile-time implementation checks (e.g., `var _ Interface = (*Implementation)(nil)`).

3. **Method Signature Inconsistency**: Similar operations across different interfaces have inconsistent parameter and return types.

4. **Naming Convention Variations**: Method names don't follow consistent patterns across similar interfaces.

5. **Error Handling Variations**: Some interfaces return explicit errors, while others use different patterns.

## Action Plan

### 1. Standardize Interface Documentation

- Add consistent godoc comments to all interface methods
- Document parameters, return values, and error conditions
- Include usage examples where appropriate

### 2. Add Compile-Time Implementation Checks

- Add `var _ Interface = (*Implementation)(nil)` statements for all implementations
- Place these checks after interface definitions or in implementation files
- Ensure all concrete types properly implement their interfaces

### 3. Normalize Method Signatures

- Standardize context passing (e.g., first parameter)
- Ensure consistent error handling (e.g., always return errors as the last value)
- Align parameter ordering across similar methods

### 4. Apply Interface Segregation

- Review large interfaces for potential splitting
- Create focused interfaces for specific functionality
- Compose interfaces where appropriate

## Interface-Specific Tasks

### Command Package

- Improve documentation for `command.Interface`
- Add compile-time checks for all command implementations
- Consider splitting `ExecutionContext` into more focused types

### Domain Package

- Standardize repository interface methods
- Ensure all repository implementations have compile-time checks
- Document error conditions for all methods

### Storage Package

- Align `Backend` interface with `domain.SessionRepository`
- Add implementation checks for all storage backends
- Improve documentation for branch/merge operations

### LLM Package

- Normalize method signatures between `Provider` and `DomainProvider`
- Document provider options and their effects
- Add implementation checks for all provider adapters

## Success Criteria

The interface and contract consistency work will be considered successful when:

1. All interfaces have comprehensive documentation
2. All implementations include compile-time interface checks
3. Method signatures follow consistent patterns
4. Large interfaces are appropriately segregated
5. All tests pass after these changes

## Implementation Guidelines

When improving interfaces, we'll follow these guidelines:

1. Make minimal changes to maintain backward compatibility
2. Update documentation first, then add compile-time checks
3. Take an iterative approach to method signature normalization
4. Write tests for any functional changes
5. Maintain consistent naming conventions across the codebase