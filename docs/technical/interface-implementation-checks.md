# Interface Implementation Checks

This document identifies interfaces that require compile-time implementation checks in the form of `var _ Interface = (*Implementation)(nil)`.

## Current Status

| Interface | Implementation | Has Check | Location | Action Needed |
|-----------|---------------|-----------|----------|---------------|
| `command.Interface` | `command.SimpleCommand` | ❌ No | pkg/command/interface.go | Add check |
| `command.Interface` | `mockCommand` | ✅ Yes | (test file) | None |
| `command.discoverer` | `packageDiscoverer` | ❌ No | pkg/command/discovery.go | Add check |
| `command.discoverer` | `builderDiscoverer` | ❌ No | pkg/command/discovery.go | Add check |
| `command.discoverer` | `reflectionDiscoverer` | ❌ No | pkg/command/discovery.go | Add check |
| `domain.SessionRepository` | `storage.Backend` | ❌ No | pkg/storage/backend.go | Add check |
| `storage.Backend` | `filesystem.Backend` | ❌ No | pkg/storage/filesystem/filesystem.go | Add check |
| `storage.Backend` | `sqlite.Backend` | ❌ No | pkg/storage/sqlite/sqlite.go | Add check |
| `llm.Provider` | `providerAdapter` | ✅ Yes | pkg/llm/provider.go | None |
| `llm.DomainProvider` | `domainProviderAdapter` | ❌ No | pkg/llm/domain_provider.go | Add check |

## Recommended Checks to Add

### Command Package

For `command.SimpleCommand`:
```go
// Add to pkg/command/interface.go after SimpleCommand declaration
var _ Interface = (*SimpleCommand)(nil)
```

For `packageDiscoverer`:
```go
// Add to pkg/command/discovery.go after packageDiscoverer declaration
var _ discoverer = (*packageDiscoverer)(nil)
```

For `builderDiscoverer`:
```go
// Add to pkg/command/discovery.go after builderDiscoverer declaration
var _ discoverer = (*builderDiscoverer)(nil)
```

For `reflectionDiscoverer`:
```go
// Add to pkg/command/discovery.go after reflectionDiscoverer declaration
var _ discoverer = (*reflectionDiscoverer)(nil)
```

### Storage Package

For `filesystem.Backend`:
```go
// Add to pkg/storage/filesystem/filesystem.go
var _ storage.Backend = (*Backend)(nil)
```

For `sqlite.Backend`:
```go
// Add to pkg/storage/sqlite/sqlite.go
var _ storage.Backend = (*Backend)(nil)
```

For `Backend` implementing `domain.SessionRepository`:
```go
// Add to pkg/storage/backend.go
var _ domain.SessionRepository = (Backend)(nil) // Note: Interface, not pointer
```

### LLM Package

For `domainProviderAdapter`:
```go
// Add to pkg/llm/domain_provider.go
var _ DomainProvider = (*domainProviderAdapter)(nil)
```

## Implementation Process

1. Add the checks one file at a time
2. Compile after each addition to ensure correctness
3. Address any incompatibilities revealed by the checks
4. Update interface documentation during the process
5. Run tests to ensure no regressions