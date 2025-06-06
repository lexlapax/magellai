# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magellai is a command-line interface (CLI) tool and REPL that interacts with Large Language Models (LLMs). It operates in two primary modes:
- **`ask` mode**: One-shot queries
- **`chat` mode**: Interactive conversations (REPL)

The project follows a library-first design where the core intelligence (LLM providers, prompt orchestration, tools, agents, workflows) is implemented as a reusable Go module.

## Current Status (Phase 4.12 - Final validation and rollout)

✅ Phase 1: Core Foundation - Complete
✅ Phase 2: Configuration and Command Foundation - Complete  
✅ Phase 3: CLI with Kong - Complete
🚧 Phase 4: Advanced REPL Features - In Progress
  ✅ Phase 4.1: Extended REPL Commands - Complete
  ✅ Phase 4.1.1: Fix logging and file attachment issues - Complete
  ✅ Phase 4.2: Advanced Session Features - Complete
    ✅ Phase 4.2.1: Session Storage library abstraction - Complete
    ✅ Phase 4.2.2: Session Auto-save functionality - Complete
    ✅ Session export formats (JSON, Markdown) - Complete
    ✅ Session search by content - Complete
    ✅ Session tags and metadata - Complete
    ✅ Session branching/forking - Complete
    ✅ Session merging - Complete
  ✅ Phase 4.3: Error Handling & Recovery - Complete
    ✅ Log levels implemented at library level, default set to warn
    ✅ Graceful network error recovery with retry logic
    ✅ Provider fallback mechanisms with chain configuration
    ✅ Partial response handling for streaming
    ✅ Rate limit handling with intelligent backoff
    ✅ Context length management with message prioritization
    ✅ Session auto-recovery after crashes - Complete
  ✅ Phase 4.4: REPL Integration with Unified Command System - Complete
    ✅ Route REPL commands through command registry
    ✅ Support both `/` and `:` command prefixes
    ✅ Integrate with existing core commands
    ✅ Maintain command history across modes
    ✅ Support command aliases in REPL
    ✅ Context preservation between commands
  ✅ Phase 4.5: REPL UI Enhancements - Complete
    ✅ Tab completion for commands - Complete
    ✅ ANSI color output when TTY - Complete (including library refactoring)
    ✅ Non-interactive mode detection - Complete
    ✅ scan and fix Context preservation between commands - Complete
  ✅ Phase 4.6: Fix domain layer and types - Complete
    ✅ Domain package structure created
    ✅ All core domain types implemented
    ✅ Storage package refactored to use domain types
    ✅ REPL package refactored to use domain types
    ✅ LLM package updated with domain adapters
    ✅ All tests updated and passing
    ✅ Build and lint checks passing
  ✅ Phase 4.7: Fix tests, test-integration issue - Complete
    ✅ Fixed logging tests that were failing in bulk runs
    ✅ Fixed session export tests creating leftover files
    ✅ All unit and integration tests passing
  ✅ Phase 4.8: Configuration - defaults, sample etc. - Complete
    ✅ with no configuration file, use a default configuration
    ✅ add a flag or command to create an example configuration  
    ✅ show config should show all current runtime configurations
  ✅ Phase 4.9: Code abstraction and redundancy checks - Complete
    ✅ Phase 4.9.1: Type Consolidation - Complete
      • Resolved duplicate Message type definitions across packages
      • Updated pkg/llm to use domain.Message throughout
      • Created comprehensive adapter functions in pkg/llm/adapters.go
      • Fixed MessageRole vs Role type inconsistency
      • Unified Attachment type representations
      • Removed pkg/repl/types.go and migrated to domain types
      • Analyzed pkg/llm/types.go - determined to keep as adapter types
      • All tests passing after type consolidation
    ✅ Phase 4.9.2: Duplicate Conversion Functions - Complete
      • Identified and removed unused pkg/llm/domain_integration.go file
      • Eliminated redundant StorageSession type in pkg/storage/types.go
      • Updated filesystem backend to use domain types directly
      • Removed unnecessary conversion layers
      • Improved JSON serialization efficiency
    ✅ Phase 4.9.3: Package Organization and Structure - Complete
      • Created test package for integration tests
      • Added build tags to all integration test files
      • Moved integration tests from cmd/magellai/ to pkg/test/integration/
      • Fixed ask_pipeline_test.go to work with current architecture
      • Added integration build tags to sqlite tests
      • Updated Makefile targets to include necessary tags
      • Better organized test structure with proper separation
    ✅ Phase 4.9.4: Error Handling Consistency - Complete
      • Created package-specific error.go files for storage, repl, config, and llm packages
      • Standardized error handling with sentinel errors
      • Implemented consistent error wrapping with %w
      • Updated code to use new error constants
      • Removed duplicate error strings
      • All tests passing with new error handling
    ✅ Phase 4.9.5: Missing Tests - Complete
      • Created comprehensive tests for all untested files
      • Added integration tests for critical paths
      • Used table-driven test patterns throughout
      • Added benchmark tests where appropriate
      • Complete test coverage achieved
    ✅ Phase 4.9.6: Fix Integration Test Failures - Complete
      • Fixed hanging test in provider_fallback_integration_test.go
      • Updated test expectations to match implementation
      • Fixed configuration precedence tests with proper environment handling
      • Consolidated integration tests to cmd/magellai/
      • Removed empty pkg/test directories
      • All integration tests now passing
    ✅ Phase 4.9.7: Test Organization and Helpers - Complete
      • Consolidated test helpers and mocks in internal/testutil/mocks/
      • Fixed provider mock implementation to match current interface
      • Improved I/O and context management in tests with proper error handling
      • Standardized loop-based copying with copy() function
      • Fixed all linting issues in test helpers
      • Addressed unused variable warnings in storage mock implementation
    ✅ Phase 4.9.8: Logging and Instrumentation - Complete
      • Standardized logging approach throughout the codebase
      • Replaced all fmt.Print statements with proper logging
      • Implemented structured logging with consistent field naming
      • Added log level configuration in settings
      • Improved error context in log messages
      • Centralized logging initialization in internal/logging
      • Added proper debug logging for troubleshooting
    ✅ Phase 4.9.9: Function and Method Cleanup - Complete
      • Audited and removed unused exported functions
      • Documented or removed experimental/WIP functions
      • Created new stringutil package for consolidated utilities
      • Standardized path handling, ID generation, and validation
      • Added //ABOUTME: sections to all code files
    ✅ Phase 4.9.10: Interface and Contract Consistency - Complete
      • Added compile-time interface implementation checks
      • Standardized method signatures across similar interfaces
      • Enhanced interface documentation with comprehensive comments
      • Fixed inconsistencies between storage.Backend and domain.SessionRepository
      • Updated implementations to match standardized interfaces
      • All tests passing with improved interface consistency
    ✅ Phase 4.9.11: Import and Dependency Cleanup - Complete
      • Created pkg/replapi package to break circular dependencies
      • Implemented dependency injection pattern for REPL instantiation
      • Refactored pkg/command/core and pkg/repl to eliminate import cycle
      • Created documentation for dependency management strategies
      • Documented intentional coupling points in the codebase
      • Improved package boundary definition
  ✅ Phase 4.10: Manual test suite for cmd line - Complete
    ✅ Created comprehensive integration tests for CLI functionality
    ✅ Implemented test utilities for consistent test setup and teardown
    ✅ Added tests for both filesystem and SQLite storage backends
    ✅ Created tests for all core commands (ask, chat, config, etc.)
    ✅ Implemented tests for session branching and merging
    ✅ Added provider fallback tests with error conditions
    ✅ Documented potential issues and improvements in test_issues.md
  ✅ Phase 4.11: Documentation and architecture updates - Complete
  🚧 Phase 4.12: Final validation and rollout - In Progress
    ✅ API_KEYS - Read API keys from environment variables when no config file is present
    ✅ Automatically select default provider based on available API keys
    ✅ Improved error handling for missing API keys

## Development Conventions

- **Workflow Task Completion**: 
  - Run full `make`, `make test`, `make test-integration`, `make lint` after every task completion

### Development Memories

- I'll do git actions myself