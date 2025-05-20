# Interface Method Signature Analysis

This document analyzes method signature inconsistencies across interfaces in the Magellai codebase.

## Repository Interfaces

### SessionRepository vs. Backend

| Operation | SessionRepository | Storage Backend | Notes |
|-----------|-------------------|----------------|-------|
| Create | `Create(session *Session) error` | `NewSession(name string) *domain.Session` | Different approach: Backend creates and returns, Repository takes and persists |
| Read | `Get(id string) (*Session, error)` | `LoadSession(id string) (*domain.Session, error)` | Naming inconsistency (Get vs Load) |
| Update | `Update(session *Session) error` | `SaveSession(session *domain.Session) error` | Naming inconsistency (Update vs Save) |
| Delete | `Delete(id string) error` | `DeleteSession(id string) error` | Backend adds type name to method |
| List | `List() ([]*SessionInfo, error)` | `ListSessions() ([]*domain.SessionInfo, error)` | Backend adds type name to method |
| Search | `Search(query string) ([]*SearchResult, error)` | `SearchSessions(query string) ([]*domain.SearchResult, error)` | Backend adds type name to method |

The Storage Backend interface also includes additional methods not in SessionRepository:
- `ExportSession(id string, format domain.ExportFormat, w io.Writer) error`
- `MergeSessions(targetID, sourceID string, options domain.MergeOptions) (*domain.MergeResult, error)`
- `Close() error`

### Recommendations for Repository Interfaces

1. **Align Method Names**: 
   - Standardize on `Get/Create/Update/Delete` (without type names) in both interfaces
   - Or standardize on more descriptive `GetSession/CreateSession/etc.` in both

2. **Method Parameter Consistency**:
   - Both should use either pass-by-value or pass-by-reference consistently
   - Both should use consistent parameter naming

3. **Context Parameter**:
   - Consider adding context parameters to both interfaces for cancellation support

4. **Feature Parity**:
   - Either add export/merge operations to SessionRepository
   - Or create a separate interface for these extended operations

## Provider Interfaces

### Provider vs. DomainProvider

| Operation | Provider | DomainProvider | Notes |
|-----------|----------|----------------|-------|
| Message Generation | `GenerateMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (*Response, error)` | `GenerateDomainMessage(ctx context.Context, messages []*domain.Message, options ...ProviderOption) (*domain.Message, error)` | Inconsistent use of slice of structs vs slice of pointers |
| Message Streaming | `StreamMessage(ctx context.Context, messages []domain.Message, options ...ProviderOption) (<-chan StreamChunk, error)` | `StreamDomainMessage(ctx context.Context, messages []*domain.Message, options ...ProviderOption) (<-chan *domain.Message, error)` | Inconsistent use of slice of structs vs slice of pointers |

### Recommendations for Provider Interfaces

1. **Consistent Pointer Usage**:
   - Standardize on either value types or pointer types for both interfaces
   - Recommend using pointer types for complex structures like messages

2. **Return Type Consistency**:
   - Provider returns custom `*Response` type while DomainProvider returns `*domain.Message`
   - Consider using wrapper or adapter functions to standardize on domain types

3. **Method Naming**:
   - Provider uses `GenerateMessage` while DomainProvider uses `GenerateDomainMessage`
   - Consider a more consistent naming scheme that emphasizes the difference

## Command Interface

The `command.Interface` appears consistent internally, but some observations:

1. **Context Handling**:
   - Uses `context.Context` as a parameter rather than embedding it in the ExecutionContext
   - Consider consolidating to simplify method signatures

2. **Error Handling**:
   - All methods return errors as the last parameter (good practice)
   - Consider standardizing error types and wrapping

## General Recommendations

Based on the analysis of interfaces across the codebase, the following recommendations apply:

1. **Consistent Context Handling**:
   - For all interfaces that could benefit from cancellation/timeouts, add context as first parameter
   - Example: `Get(ctx context.Context, id string) (*Session, error)`

2. **Error Wrapping**:
   - Ensure all interfaces consistently use error wrapping when appropriate
   - Use domain-specific error types when possible

3. **Method Naming Conventions**:
   - For CRUD operations: Use `Create`, `Get`, `Update`, `Delete`
   - For list operations: Use `List` or `ListXxx` consistently
   - For search operations: Use `Search` or `SearchXxx` consistently

4. **Parameter Ordering**:
   - Context (if applicable) should always be first parameter
   - Primary identifiers (IDs, names) should come next
   - Complex structures should come next
   - Options/flags should come last

5. **Return Values**:
   - Single item operations should return `(Item, error)`
   - Multiple item operations should return `([]Items, error)`
   - Operations that don't return data should return just `error`

By applying these recommendations consistently, the interfaces will be more intuitive, easier to use, and more maintainable.