# Error Handling Standardization

This document summarizes the error handling standardization completed in Phase 4.9.4.

## Overview

We standardized error handling across the codebase by:
1. Creating package-specific error definitions
2. Using sentinel errors for common error cases
3. Implementing consistent error wrapping patterns
4. Removing duplicate error strings

## Package-Specific Error Files Created

1. **pkg/storage/errors.go** - Storage backend errors
   - `ErrInvalidBackend`
   - `ErrSessionNotFound`
   - `ErrSessionExists`
   - `ErrCorruptedData`
   - `ErrStorageFull`
   - `ErrPermission`
   - `ErrBackendNotAvailable`
   - `ErrTransactionFailed`
   - `ErrBranchNotFound`
   - `ErrInvalidBranch`
   - `ErrMergeConflict`

2. **pkg/repl/errors.go** - REPL operation errors
   - `ErrInvalidCommand`
   - `ErrSessionNotInitialized`
   - `ErrNoActiveSession`
   - `ErrCommandFailed`
   - `ErrInvalidAttachment`
   - `ErrAttachmentNotFound`
   - `ErrInvalidSystemPrompt`
   - `ErrExportFailed`
   - `ErrInvalidExportFormat`
   - `ErrBranchOperationFailed`
   - `ErrMergeOperationFailed`
   - `ErrInvalidMetadataKey`
   - `ErrRecoveryFailed`
   - `ErrCommandNotFound`

3. **pkg/config/errors.go** - Configuration errors
   - `ErrConfigNotFound`
   - `ErrInvalidConfig`
   - `ErrConfigLoadFailed`
   - `ErrConfigSaveFailed`
   - `ErrProfileNotFound`
   - `ErrInvalidProfile`
   - `ErrAliasNotFound`
   - `ErrInvalidAlias`
   - `ErrSettingNotFound`
   - `ErrInvalidSettingValue`
   - `ErrMergeConflict`
   - `ErrValidationFailed`
   - `ErrPermission`

4. **pkg/llm/errors.go** - LLM provider errors
   - `ErrProviderNotFound`
   - `ErrModelNotFound`
   - `ErrInvalidProvider`
   - `ErrInvalidModel`
   - `ErrAPIKeyMissing`
   - `ErrInvalidAPIKey`
   - `ErrProviderUnavailable`
   - `ErrContextLengthExceeded`
   - `ErrRateLimitExceeded`
   - `ErrTokenLimitExceeded`
   - `ErrStreamingNotSupported`
   - `ErrInvalidResponse`
   - `ErrPartialResponse`
   - `ErrProviderTimeout`
   - `ErrProviderError`

## Error Wrapping Pattern

We standardized on using Go's error wrapping with `%w` format:

```go
// Old pattern
return fmt.Errorf("session not found: %s", id)

// New pattern
return fmt.Errorf("%w: %s", storage.ErrSessionNotFound, id)
```

This allows proper error testing with `errors.Is()`:

```go
if errors.Is(err, storage.ErrSessionNotFound) {
    // Handle session not found
}
```

## Code Updates

### Storage Package
- Updated filesystem backend to use `storage.ErrSessionNotFound`
- Updated sqlite backend to use `storage.ErrSessionNotFound`

### Config Package
- Updated profile functions to use `config.ErrProfileNotFound`

### LLM Package
- Updated model lookup to use `llm.ErrModelNotFound`

## Benefits

1. **Consistent Error Handling**: All packages now follow the same pattern
2. **Error Testing**: Callers can test for specific errors using `errors.Is()`
3. **Better Error Context**: Wrapping provides additional context while preserving the error type
4. **Reduced Duplication**: Common errors are defined once and reused
5. **Maintainability**: Easier to update error messages in one place

## Best Practices Going Forward

1. Always use sentinel errors for common error conditions
2. Wrap errors with context using `fmt.Errorf("%w: extra context", err)`
3. Test for errors using `errors.Is()` instead of string comparison
4. Create package-specific error files when needed
5. Document error conditions in godoc comments