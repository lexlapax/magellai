# Test Utilities

This package provides shared testing utilities for the Magellai project.

## Structure

- **mocks/**: Mock implementations of common interfaces
- **fixtures/**: Test data factories for consistent test objects
- **helpers/**: Utility functions for common testing operations

## Usage

### Using Mocks

```go
// Create a mock backend
backend := mocks.NewMockBackend()
backend.SetError(errors.New("test error"))

// Create a mock provider
provider := mocks.NewMockProvider()
provider.SetResponse(&llm.Response{
    Content: "test response",
    Model:   "test-model",
})
```

### Using Fixtures

```go
// Create test sessions
session := fixtures.CreateTestSession("test-id")
sessionWithMessages := fixtures.CreateTestSessionWithMessages("test-id", 5)
sessionTree, branches := fixtures.CreateTestSessionTree()

// Create test messages
userMsg := fixtures.CreateTestUserMessage("Hello")
assistantMsg := fixtures.CreateTestAssistantMessage("Hi there!")

// Create test attachments
textAttachment := fixtures.CreateTextAttachment("file.txt", "content")
imageAttachment := fixtures.CreateImageAttachment("photo.jpg")
```

### Using Helpers

```go
// I/O helpers
tempDir := helpers.CreateTempDir(t, "test")
content := helpers.ReadFile(t, path)
helpers.AssertFileContains(t, path, "expected content")

// Comparison helpers
helpers.CompareSessions(t, expected, actual)
helpers.CompareMessages(t, expected, actual)
helpers.AssertErrorContains(t, err, "expected error")
helpers.AssertNoError(t, err)

// Context helpers
ctx := helpers.TestContext(t)
ctx = helpers.TestContextWithTimeout(t, 5*time.Second)
helpers.AssertContextDone(t, ctx, time.Second)
```

## Naming Conventions

- Mock types: `Mock<Interface>` (e.g., `MockProvider`, `MockBackend`)
- Fixture functions: `Create<Type>` (e.g., `CreateTestSession`, `CreateTestMessage`)
- Helper functions: `<Action><Object>` (e.g., `CompareMessages`, `AssertNoError`)

## Best Practices

1. Use mocks for testing interface implementations
2. Use fixtures to create consistent test data across tests
3. Use helpers to reduce boilerplate in test code
4. Keep test-specific utilities in the test file unless they're reusable
5. Always use `t.Helper()` in helper functions for better error reporting

## Adding New Utilities

When adding new test utilities:

1. Place mocks in the `mocks/` directory
2. Place test data factories in the `fixtures/` directory
3. Place helper functions in the `helpers/` directory
4. Add appropriate documentation and examples
5. Follow the established naming conventions