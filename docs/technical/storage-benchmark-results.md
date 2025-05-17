# Storage Backend Performance Benchmark Results

This document compares the performance characteristics of the FileSystem and SQLite storage backends.

## Test Environment

- Platform: Darwin (macOS)
- Architecture: ARM64 (Apple M1 Ultra)
- Go Version: 1.24.3
- SQLite: Without FTS5 support

## Benchmark Results

### Session Save Performance

| Backend | Messages | Ops/sec | ns/op | B/op | allocs/op |
|---------|----------|---------|-------|------|-----------|
| FileSystem | 10 | 8,590 | 138,269 | 15,822 | 34 |
| FileSystem | 50 | 4,687 | 238,365 | 65,559 | 87 |
| FileSystem | 100 | 3,157 | 362,949 | 130,519 | 154 |
| SQLite | 10 | 1,388 | 958,669 | 21,324 | 429 |
| SQLite | 50 | 843 | 1,895,231 | 93,637 | 1,820 |
| SQLite | 100 | 583 | 2,613,101 | 185,119 | 3,579 |

### Key Findings

1. **FileSystem is faster for saves**: FileSystem backend is approximately 6-7x faster than SQLite for save operations.

2. **Memory efficiency**: FileSystem uses fewer allocations and less memory per operation.

3. **Scalability**: Both backends show linear performance degradation with message count, but SQLite's overhead is more pronounced.

4. **SQLite advantages** (not reflected in raw performance):
   - Multi-user support with isolation
   - Full-text search capabilities (when FTS5 is available)
   - ACID properties and crash recovery
   - Concurrent access safety
   - Better query capabilities

## Recommendations

1. **Use FileSystem backend when**:
   - Single-user scenarios
   - Performance is critical
   - Simple session storage is sufficient
   - File-based workflows are preferred

2. **Use SQLite backend when**:
   - Multi-user support is needed
   - Search functionality is important
   - Data consistency is critical
   - Concurrent access is required
   - Future migration to other databases is planned

## Performance Optimization Tips

### For FileSystem Backend
- Keep session files in directories with good file system performance
- Consider SSD storage for better I/O
- Implement file-based indexing for search if needed
- Use compression for large sessions

### For SQLite Backend
- Enable WAL mode for better concurrency
- Use prepared statements (already implemented)
- Consider connection pooling for high-load scenarios
- Build with FTS5 support for better search performance
- Tune page size and cache size based on workload

## Future Improvements

1. Add connection pooling to SQLite backend
2. Implement batch operations for bulk saves
3. Add caching layer for frequently accessed sessions
4. Optimize JSON marshaling/unmarshaling
5. Consider using SQLite's JSON1 extension for structured data