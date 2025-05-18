# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magellai is a command-line interface (CLI) tool and REPL that interacts with Large Language Models (LLMs). It operates in two primary modes:
- **`ask` mode**: One-shot queries
- **`chat` mode**: Interactive conversations (REPL)

The project follows a library-first design where the core intelligence (LLM providers, prompt orchestration, tools, agents, workflows) is implemented as a reusable Go module.

## Current Status (Phase 4.8 - Code abstraction and redundancy checks)

✅ Phase 1: Core Foundation - Complete
✅ Phase 2: Configuration and Command Foundation - Complete  
✅ Phase 3: CLI with Kong - Complete
🚧 Phase 4: Advanced REPL Features - In Progress
  ✅ Phase 4.1: Extended REPL Commands - Complete
  ✅ Phase 4.1.1: Fix logging and file attachment issues - Complete
  ✅ Phase 4.2.1: Session Storage library abstraction - Complete
  ✅ Phase 4.2.2: Session Auto-save functionality - Complete
    ✅ Auto-save functionality - Complete
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
  🔲 Phase 4.4: REPL Integration with Unified Command System - Pending
  🔲 Phase 4.5: REPL UI Enhancements - Pending
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
  🚧 Phase 4.8: Run code, abstraction, redundancy checks and fixes - Next
  🔲 Phase 4.9: Documentation and architecture updates - Pending (moved from 4.6)
  🔲 Phase 4.10: Final validation and rollout - Pending (moved from 4.6)

## Development Conventions

- **Workflow Task Completion**: 
  - Run full `make`, `make test`, `make test-integration`, `make lint` after every task completion

### Rest of the file remains the same... (previous content continues)