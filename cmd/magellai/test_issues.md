# CLI Test Issues and Findings

This document outlines the potential edge cases, issues, and areas for improvement identified during the creation of the CLI test suite.

## Implementation Issues Found

1. **Configuration Flag Name**
   - The tests were using `--config` but the actual flag is `--config-file`
   - All test utilities were updated to use the correct flag name

2. **CLI Output Format Changes**
   - The expected output formats in tests don't match the current CLI output
   - Need to update test assertions to match the actual output structure
   - Specific issues:
     - Help output format doesn't contain "Commands:" but has category-based grouping
     - Version --json flag may not be supported
     - Config show output doesn't contain "Configuration:" prefix
     - Model list output doesn't match expected format

## Potential Issues

1. **Storage Backend Initialization**
   - When switching between storage backends in tests, proper cleanup is critical
   - SQLite database files should be properly closed to avoid file locking issues
   - Consider adding a dedicated test helper for storage type switching

2. **Interactive Test Timeout Handling**
   - Current interactive tests have a fixed 10-second timeout
   - May not be sufficient for tests with complex interactions or network calls
   - Consider making timeout configurable based on test type

3. **Error Propagation in Interactive Mode**
   - Error handling in REPL mode might not propagate errors correctly in some cases
   - Tests should verify error conditions are properly communicated to the user
   - Add dedicated tests for error output formatting

4. **Provider Fallback Edge Cases**
   - Tests for provider fallback with mixed transient and permanent failures needed
   - Special handling required for rate limit errors vs. authentication failures
   - Consider adding more granular fallback tests with specific error types

5. **File Path Handling Cross-Platform Compatibility**
   - Current tests use OS-specific path separators and commands
   - May not work correctly on all platforms (especially Windows)
   - Add more robust path handling in test utilities

## Improvement Areas

1. **Test Configuration Management**
   - Current approach creates a new config file for each test
   - Consider caching configurations for test suites with similar requirements
   - Add support for overriding specific config sections without regenerating the entire file

2. **Mock Provider Enhancements**
   - Current mock providers have limited functionality
   - Add support for simulating different response patterns and error types
   - Implement more realistic streaming behavior in mocks

3. **Cross-Platform Test Support**
   - Improve handling of platform-specific commands (e.g., file manipulation)
   - Use filepath package consistently for path operations
   - Replace direct bash commands with Go utilities where possible

4. **Test Isolation**
   - Some tests may have interdependencies or side effects
   - Improve cleanup to ensure complete isolation between test runs
   - Consider adding transaction-based setup/teardown for database tests

5. **Session State Verification**
   - Add more thorough verification of session state after operations
   - Directly inspect storage backend to verify expected state when appropriate
   - Compare actual and expected state structures rather than relying on output text

## Missing Test Coverage

1. **Concurrent Session Access**
   - Tests for multiple simultaneous sessions accessing the same storage backend
   - Locking and concurrency handling in SQLite mode

2. **Network Error Resilience**
   - Simulate network interruptions during streaming responses
   - Verify partial response handling and recovery

3. **Large Message Handling**
   - Test performance and memory usage with very large messages
   - Verify proper chunking and streaming of large responses

4. **Command Line Argument Edge Cases**
   - Test handling of unusual or extreme argument values
   - Verify proper escaping of special characters in arguments

5. **Environment Variable Override Tests**
   - More comprehensive tests for environment variable precedence
   - Tests for invalid or conflicting environment settings

## Next Steps

1. Prioritize and address the identified issues
2. Expand test coverage for missing areas
3. Implement more realistic mock providers for advanced testing
4. Add performance and stress tests for critical functionality
5. Improve cross-platform compatibility of the test suite