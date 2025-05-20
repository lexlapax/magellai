*** this is a continuation of TODO-DONE-ARCHIVE.md**
### 4.9 Code abstraction and redundancy checks

#### 4.9.1 Type Consolidation and Abstraction Issues ✅ (Completed)
  - [x] Resolve duplicate Message type definitions across packages:
    - [x] Consolidate pkg/domain/message.go, pkg/llm/types.go, and go-llms message types
    - [x] Decide on single source of truth (domain types recommended)
    - [x] Remove redundant Message definitions and update all references
    - [x] Updated pkg/llm to use domain.Message throughout
    - [x] Created comprehensive adapter functions in pkg/llm/adapters.go
    - [x] Updated pkg/llm/provider.go to use domain types
    - [x] Updated pkg/repl/conversation.go to use domain types directly (no more conversions)
    - [x] Fixed all compilation errors and tests
  - [x] Resolve MessageRole vs Role type inconsistency:
    - [x] Use domain.MessageRole consistently throughout codebase
    - [x] Add conversion for go-llms Role type (includes "tool" role) in adapters.go
  - [x] Unify Attachment type representations:
    - [x] Consolidated to use domain.Attachment throughout codebase
    - [x] Created adapter functions to convert between domain and go-llms types
    - [x] Fixed attachment type inconsistencies in tests
  - [x] Remove pkg/repl/types.go:
    - [x] Removed type aliases that were pointing to domain types
    - [x] Updated all files in repl package to use domain types directly
    - [x] Fixed all test files to use domain types
    - [x] All tests passing after migration
  - [x] Analyze pkg/llm/types.go:
    - [x] Determined that LLM types should remain as technical adapter types
    - [x] These serve different purpose than domain types (API integration vs business model)
    - [x] Keeping separation maintains proper architectural boundaries

#### 4.9.2 Duplicate Conversion Functions ✅ (Completed)
  - [x] Identify and remove unused pkg/llm/domain_integration.go file (superfluous)
  - [x] Eliminate redundant StorageSession type in pkg/storage/types.go:
    - [x] Refactored to use domain.Session directly
    - [x] Created methods on domain.Session for serialization instead
    - [x] Updated all storage backends to match new approach
  - [x] Update filesystem backend to use domain types directly
  - [x] Remove unnecessary conversion layers:
    - [x] Removed duplicate conversion code from filesystem.go
    - [x] Improved JSON serialization efficiency
    - [x] Eliminated type marshaling overhead
  - [x] Fix all tests to match new structure

#### 4.9.3 Package Organization and Structure ✅ (Completed)
  - [x] Create test package for integration tests
  - [x] Add build tags to integration test files
  - [x] Move integration tests to proper location
  - [x] Fix ask_pipeline_test.go to work with current architecture
  - [x] Add integration build tags to sqlite tests
  - [x] Update Makefile targets to include necessary tags

#### 4.9.4 Error Handling Consistency ✅ (Completed)
  - [x] Create package-specific error.go files for all packages
  - [x] Standardize error handling with sentinel errors
  - [x] Implement consistent error wrapping with %w
  - [x] Update code to use new error constants
  - [x] Remove duplicate error strings

#### 4.9.5 Missing Tests ✅ (Completed)
  - [x] Create comprehensive tests for all untested files
  - [x] Add integration tests for critical paths
  - [x] Use table-driven test patterns throughout
  - [x] Add benchmark tests where appropriate

#### 4.9.6 Fix Integration Test Failures ✅ (Completed)
  - [x] Fix hanging test in provider_fallback_integration_test.go
  - [x] Update test expectations to match implementation
  - [x] Fix configuration precedence tests with proper environment handling
  - [x] Consolidate integration tests
  - [x] Remove empty pkg/test directories

#### 4.9.7 Test Organization and Helpers ✅ (Completed)
  - [x] Consolidate test helpers and mocks
  - [x] Fix provider mock implementation to match current interface
  - [x] Improve I/O and context management in tests
  - [x] Standardize loop-based copying with copy()
  - [x] Fix all linting issues in test helpers

#### 4.9.8 Logging and Instrumentation ✅ (Completed)
  - [x] Standardize logging approach throughout the codebase
  - [x] Replace all fmt.Print statements with proper logging
  - [x] Implement structured logging with consistent field naming
  - [x] Add log level configuration in settings
  - [x] Improve error context in log messages
  - [x] Centralize logging initialization
  - [x] Add proper debug logging for troubleshooting

#### 4.9.9 Function and Method Cleanup ✅ (Completed)
  - [x] Audit and remove unused exported functions
  - [x] Document or remove experimental/WIP functions
  - [x] Create new stringutil package for consolidated utilities
  - [x] Standardize path handling, ID generation, and validation
  - [x] Add //ABOUTME: sections to all code files

#### 4.9.10 Interface and Contract Consistency ✅ (Completed)
  - [x] Add compile-time interface implementation checks using `var _ Interface = (*Implementation)(nil)` pattern
  - [x] Standardize method signatures across similar interfaces
    - [x] Update method names to follow CRUD patterns (Create/Read/Update/Delete)
    - [x] Align storage.Backend with domain.SessionRepository naming conventions
    - [x] Ensure context.Context is consistently used in methods that may block
  - [x] Enhance interface documentation with comprehensive godoc comments
    - [x] Document expected behavior and error cases
    - [x] Add examples where appropriate
    - [x] Clarify ownership of returned objects
  - [x] Fix inconsistencies between storage.Backend and domain.SessionRepository
    - [x] Standardize on SaveSession -> Create/Update pattern
    - [x] Standardize on LoadSession -> Get pattern
    - [x] Standardize on ListSessions -> List pattern
    - [x] Standardize on DeleteSession -> Delete pattern
    - [x] Standardize on SearchSessions -> Search pattern
  - [x] Update implementations to match standardized interfaces
    - [x] Update filesystem.go implementation
    - [x] Update sqlite.go implementation
    - [x] Update session_manager.go to use new method names
    - [x] Update testing_mock.go to match interface changes
  - [x] Ensure all tests passing with improved interface consistency

#### 4.9.11 Import and Dependency Cleanup ✅ (Completed)
  - [x] Remove circular dependencies:
    - [x] Audit import graphs for circular references
    - [x] Fixed circular dependency between pkg/command/core and pkg/repl
    - [x] Created pkg/replapi package for shared interfaces
    - [x] Refactored to eliminate circular imports
  - [x] Minimize cross-package dependencies:
    - [x] Review imports between packages
    - [x] Implemented dependency injection for REPL instantiation
    - [x] Added factory pattern for REPL creation
    - [x] Created documentation for dependency management strategies in docs/technical/dependency-management.md
    - [x] Document intentional coupling points in the codebase

### 4.10 Manual test suite for cmd line (cli and repl both) ✅ (Completed)
  - [x] Create a set of integration tests using command line directly for cli and repl
    - [x] Created test.config.yaml and test.mock.config.yaml template files
    - [x] Implemented TestEnv structure for consistent test setup and environment
    - [x] Developed helpers for running commands and interactive sessions
    - [x] Added support for conditional testing based on storage backend
  - [x] Implement tests for core CLI functionality
    - [x] Created test files for basic commands (version, help, config)
    - [x] Added tests for ask command with various options
    - [x] Implemented tests for chat mode and REPL commands
    - [x] Added tests for session management and history
  - [x] Implement tests for more complex features
    - [x] Session branching and merging tests
    - [x] Provider fallback and error handling tests
    - [x] File attachment tests
    - [x] Command alias tests
  - [x] Support both storage backends in tests
    - [x] Added ForEachStorageType test helper to run tests on multiple backends
    - [x] Implemented conditional testing based on storage type
    - [x] Created proper cleanup routines to prevent test pollution
  - [x] Document potential issues and improvements
    - [x] Created test_issues.md with detailed findings
    - [x] Documented edge cases and areas for future improvement
    - [x] Identified missing test coverage to address in future work