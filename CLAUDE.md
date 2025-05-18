# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Magellai is a command-line interface (CLI) tool and REPL that interacts with Large Language Models (LLMs). It operates in two primary modes:
- **`ask` mode**: One-shot queries
- **`chat` mode**: Interactive conversations (REPL)

The project follows a library-first design where the core intelligence (LLM providers, prompt orchestration, tools, agents, workflows) is implemented as a reusable Go module.

## Current Status (Phase 4.6 - Fix domain layer and types)

✅ Phase 1: Core Foundation - Complete
✅ Phase 2.1: Configuration Management with Koanf - Complete
✅ Phase 2.2: Configuration Schema - Complete  
✅ Phase 2.3: Configuration Utilities - Complete (mostly)
✅ Phase 2.4: Unified Command System - Complete
✅ Phase 2.5: Core Commands Implementation - Complete
✅ Phase 2.6: Models inventory file - Complete
✅ Phase 3.1: CLI Structure Setup - Complete
✅ Phase 3.2: Ask Command - Complete
✅ Phase 3.2.1: CLI Help System Improvements - Complete
✅ Phase 3.3: Chat Command & REPL Foundation - Complete
✅ Phase 3.4: Configuration Commands (using koanf) - Complete
✅ Phase 3.5: Logging and Verbosity Implementation - Complete
  ✅ Phase 3.5.1: Configuration Logging - Complete
  ✅ Phase 3.5.2: LLM Provider Logging - Complete
  ✅ Phase 3.5.3: Session Management Logging - Complete
🚧 Phase 4: Advanced REPL Features - In Progress
  ✅ Phase 4.1: Extended REPL Commands - Complete
  ✅ Phase 4.1.1: Fix logging and file attachment issues - Complete
  ✅ Phase 4.2.1: Session Storage library abstraction - Complete
    ✅ Phase 4.2.1.1: Interface and Filesystem Implementation - Complete
    ✅ Phase 4.2.1.2: Database Support (SQLite) - Complete
    ✅ Phase 4.2.1.3: Default session storage configuration & refactoring - Complete
  ✅ Phase 4.2.2: Session Auto-save functionality - Complete (mostly)
    ✅ Auto-save functionality - Complete
    ✅ Session export formats (JSON, Markdown) - Complete
    ✅ Session search by content - Complete
    🔲 Session tags and metadata - Pending
    🔲 Session branching/forking - Pending
    🔲 Session merging - Pending
  🚧 Phase 4.6: Fix domain layer and types - In Progress
    ✅ Domain package structure created
    ✅ All core domain types implemented (Session, Message, Attachment, Conversation, etc.)
    ✅ Comprehensive tests for domain layer
    🚧 Refactoring storage package to use domain types - Next

## Development Conventions

- **Workflow Task Completion**: 
  - Run full `make`, `make test`, `make test-integration`, `make lint` after every task completion

### Rest of the file remains the same... (previous content continues)