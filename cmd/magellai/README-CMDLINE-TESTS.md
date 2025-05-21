# Command Line Integration Tests for Magellai

This directory contains integration tests that specifically focus on testing the command line interface (CLI) functionality by actually executing the CLI binary.

## Test Organization

The CLI tests are organized under the build tag `cmdline` instead of the standard `integration` tag used by other integration tests. This separation allows for:

1. Running CLI tests separately from other integration tests
2. Controlling when these tests run, as they may be more brittle due to environment dependencies
3. Fine-tuning test expectations based on CLI output format changes

## Test Files

The following files contain CLI-specific integration tests:

- `cli_test_utils.go` - Utilities for CLI testing
- `cli_basic_test.go` - Basic CLI functionality tests (help, version, etc.)
- `cli_ask_test.go` - Tests for the "ask" command
- `cli_chat_test.go` - Tests for the "chat" command and REPL
- `cli_session_test.go` - Tests for session management features
- `provider_fallback_test.go` - Tests for provider fallback mechanisms
- `config_precedence_integration_test.go` - Tests for config loading behavior
- `integration_test.go` - General CLI integration tests
- `session_branching_integration_test.go` - Tests for session branching

## Running the Tests

You can run the CLI tests specifically using the Makefile target:

```bash
make test-cmdline
```

This will run only the tests with the `cmdline` build tag.

To run all tests, including CLI tests:

```bash
make test-all
```

## Test Utilities

The `cli_test_utils.go` file provides several helpers for CLI testing:

- `TestEnv` - Test environment setup with temporary directories
- `ForEachStorageType` - Run tests with each available storage backend
- `WithMockEnv` - Run tests with mock providers
- `WithLiveEnv` - Run tests with real providers (requires API keys)

## Test Maintenance

When the CLI interface changes, these tests may need to be updated. Common changes that require test updates:

1. Flag name changes
2. Output format changes
3. Command structure changes
4. Error message format changes

When updating these tests, consider:

- Making assertions more flexible to handle minor format changes
- Using pattern matching where appropriate rather than exact string matches
- Adding conditional logic to handle different CLI versions or configurations
- Using `t.Skip()` for tests that are temporarily incompatible

## Notes for Test Authors

- Prefer using the test utilities over direct command execution
- Ensure proper cleanup of temporary files and directories
- Use unique identifiers for session names to avoid conflicts
- Consider environment-specific behavior (SQLite availability, API keys, etc.)
- Be aware of timeout settings for interactive tests