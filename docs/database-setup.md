# Database Storage Support

Magellai supports database-based session storage as an optional feature. This allows for better session management, multi-user support, and full-text search capabilities.

## Supported Databases

Currently supported:
- SQLite (local database)

Planned:
- PostgreSQL (remote database)
- Redis (in-memory cache)
- S3 (object storage)

## Building with Database Support

Database support must be enabled at compile time using build tags:

```bash
# Build with SQLite support only
make build-db

# Build with all database support (future)
make build-full

# Or manually:
go build -tags="sqlite db" -o magellai ./cmd/magellai
```

## Configuration

### SQLite Configuration

Add to your Magellai configuration file:

```yaml
session:
  storage:
    type: sqlite
    settings:
      path: ~/.config/magellai/sessions.db
      # Optional: specify user ID (defaults to system username)
      user_id: myusername
```

Or set via environment variables:

```bash
export MAGELLAI_SESSION_STORAGE_TYPE=sqlite
export MAGELLAI_SESSION_STORAGE_SETTINGS_PATH=/path/to/sessions.db
```

## Features

### Multi-User Support

The database storage backends support multi-tenant usage:
- Sessions are isolated by user ID
- Defaults to the current system username
- Can be overridden in configuration

### Full-Text Search

SQLite storage includes full-text search capabilities:
- Search across all session messages
- FTS5 engine for efficient searching
- Highlighted search results

### Performance Considerations

Database storage offers:
- Better concurrent access
- Efficient searching and filtering
- Automatic indexing
- Transaction support

However, for simple single-user scenarios, filesystem storage may be sufficient.

## Migration

To migrate from filesystem to database storage:

1. Build with database support
2. Update your configuration
3. Use the built-in migration tool (coming soon)

## Schema

The SQLite database uses the following schema:

```sql
-- Sessions table
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT,
    created TIMESTAMP NOT NULL,
    updated TIMESTAMP NOT NULL,
    metadata TEXT,
    conversation TEXT NOT NULL
);

-- Messages table  
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    attachments TEXT,
    timestamp TIMESTAMP NOT NULL,
    sequence INTEGER NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

-- Full-text search index
CREATE VIRTUAL TABLE messages_fts USING fts5(
    session_id UNINDEXED,
    user_id UNINDEXED,
    role UNINDEXED,
    content
);
```

## Troubleshooting

### "SQLite storage not available" Error

This occurs when trying to use SQLite storage without the proper build tags:

```bash
# Wrong - built without SQLite support
go build -o magellai ./cmd/magellai

# Correct - built with SQLite support  
go build -tags="sqlite db" -o magellai ./cmd/magellai
```

### Permission Errors

Ensure the database file and directory have proper permissions:

```bash
mkdir -p ~/.config/magellai
chmod 700 ~/.config/magellai
```

### Performance Issues

For large databases, consider:
- Adding appropriate indexes
- Using VACUUM periodically
- Monitoring query performance
- Using PostgreSQL for larger deployments