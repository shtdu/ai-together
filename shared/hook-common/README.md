# hook-common

Shared code for hook event tools (`hook-collector` and `hook-browser`).

## Overview

This package provides common database operations, configuration, and data types for storing and retrieving Claude Code hook events using SQLite with WAL mode for concurrent access.

## Architecture

### Components

- **`storage/`** - SQLite database operations
  - `db.go` - Database connection management
  - `read.go` - Read operations (GetAllSessionIDs, GetSessionInfo, GetSessionEvents)
  - `write.go` - Write operations (StoreEvent, DeleteEventRange, DeleteSession)
  - `types.go` - Shared data types (Event, SessionInfo)
  - `constants.go` - Bucket names and tool isolation helpers
  - `sqlite.go` - SQLite-specific implementation with WAL mode

- **`config/`** - Configuration and path resolution
  - `path.go` - Workspace root detection, database path resolution

## Database Schema

```sql
-- Sessions table (tracks sessions per tool)
CREATE TABLE sessions (
    session_id TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    created_at INTEGER,
    PRIMARY KEY (session_id, tool_name)
);

-- Events table (stores hook events with auto-incremented IDs)
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    event_name TEXT,
    event_data TEXT,
    ts INTEGER,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- Indexes for common queries
CREATE INDEX idx_sessions_tool_created ON sessions(tool_name, created_at DESC);
CREATE INDEX idx_events_session_tool ON events(session_id, tool_name);
CREATE INDEX idx_events_tool_ts ON events(tool_name, ts);
CREATE INDEX idx_events_id ON events(id);
```

**Tool Isolation:** Each tool (e.g., "claude-code", "codex") has its own data space using the `tool_name` column. Queries filter by `tool_name` to ensure data separation.

## Usage

### Opening the Database

```go
import "github.com/code-together/shared/hook-common/storage"

// For reading (hook-browser)
// Note: SQLite WAL mode allows concurrent reads/writes, so readOnly is less critical
db, err := storage.OpenDB("/path/to/hook-events.db", true, "claude-code")
defer db.Close()

// For writing (hook-collector)
db, err := storage.OpenDB("/path/to/hook-events.db", false, "claude-code")
defer db.Close()
```

### Writing Events

```go
// Store an event (id is auto-incremented)
eventJSON := []byte(`{"hook_event_name": "test", "data": "value"}`)
err := storage.StoreEvent(db, "session-123", "claude-code", 0, eventJSON)
// The 4th parameter (seq) is ignored - SQLite uses auto-incremented id
```

### Reading Events

```go
// Get all sessions for a tool
sessions, err := storage.GetAllSessionIDs(db, "claude-code")
if err != nil {
    // handle error
}

// Get session info with event counts
sessionInfos, err := storage.GetSessionInfo(db, "claude-code")
for _, info := range sessionInfos {
    fmt.Printf("Session %s has %d events\n", info.ID, info.Count)
}

// Get events for a specific session
events, err := storage.GetSessionEvents(db, "session-123", "claude-code")
for _, event := range events {
    fmt.Printf("Event %d: %s\n", event.ID, event.EventName)
}
```

## SQLite Concurrency Model

### WAL Mode Benefits

SQLite with WAL (Write-Ahead Logging) mode provides excellent concurrency:

- **Multiple readers** can access simultaneously without blocking
- **Writers don't block readers** - reads can continue while writes happen
- **Multiple writers** can access concurrently (SQLite handles locking at the row level)

This is a major improvement over traditional BoltDB file-level locking.

### Real-World Usage Pattern

For `hook-collector` and `hook-browser` to work together:

1. **hook-collector** can keep DB open and write continuously
2. **hook-browser** can open DB at any time and read current state
3. **No turn-taking required** - both can operate concurrently

**Both tools can use long-lived connections:**

```go
// ✅ RIGHT: Keep DB open for continuous access
db, err := storage.OpenDB(path, false, "claude-code")
defer db.Close()

// Writer can write continuously
for {
    eventJSON := getNextEvent()
    StoreEvent(db, sessionID, "claude-code", 0, eventJSON)
}

// Reader can read anytime (even while writer is active)
events, err := GetSessionEvents(db, sessionID, "claude-code")
```

### Connection Pool Settings

The database is configured with sensible defaults:
- Max open connections: 25
- Max idle connections: 5
- Connection lifetime: 5 minutes
- WAL timeout: 5000ms

These can be adjusted in `storage/sqlite.go` if needed for your workload.

### Multiple Tools

Different tools can store data in the same database without interference:

```go
// claude-code writes to its own space
StoreEvent(db, sessionID, "claude-code", 0, eventJSON)

// codex writes to its own space
StoreEvent(db, sessionID, "codex", 0, eventJSON)

// Each tool only sees its own data
events, _ := GetSessionEvents(db, sessionID, "claude-code")
```

## Test Coverage

The package includes comprehensive tests:

### Basic Tests
- Database opening
- Session management (create, read, delete)
- Event operations (store, retrieve, delete)
- Tool isolation

### Concurrent Access Tests (3 tests)

1. **TestConcurrentAccessPattern**
   - Demonstrates real-world concurrent access
   - hook-collector writes events
   - hook-browser reads events
   - Both operate concurrently without blocking

2. **TestShortLivedConcurrentAccess**
   - hook-collector writes in batches with short-lived connections
   - hook-browser reads periodically
   - No lock errors with WAL mode
   - Both tools run concurrently

3. **TestMultipleReaders**
   - Multiple hook-browser instances read simultaneously
   - All can access DB while writer is active
   - Demonstrates SQLite's multi-reader capability

All tests use `/tmp/hook-common-test/` for test databases (no production data impact).

Run tests:
```bash
cd shared/hook-common
go test -v ./storage ./config
```

## Development

### Adding New Storage Operations

1. Add read-only operations to `storage/read.go`
2. Add write operations to `storage/write.go`
3. Add types to `storage/types.go`
4. Update tests in `storage/*_test.go`

### Testing Concurrency

When adding features that modify data:
- Test with single writer
- Test with multiple concurrent readers
- Test with multiple concurrent writers
- Verify tool isolation (different `tool_name` values)
- Ensure transactions are used correctly for consistency

### Tool Isolation

Always pass the correct `toolName` parameter to ensure data isolation:
- `claude-code` for Claude Code hooks
- `codex` for Codex hooks
- `opencode` for OpenCode hooks

## Migration from BoltDB

This package was originally written using BoltDB (bbolt). The migration to SQLite:

- **Improved concurrency** - No more file-level locking
- **Simpler API** - No bucket management
- **Better querying** - SQL for complex queries
- **Tool isolation** - Explicit `tool_name` column instead of bucket names
- **Auto-incrementing IDs** - No need to manage sequences manually

Key API changes:
- `OpenDB()` now takes a `toolName` parameter
- `StoreEvent()` auto-generates IDs (seq parameter ignored)
- Removed bucket-based functions (kept for compatibility but return nil)

## Dependencies

- `modernc.org/sqlite` v1.44.3 - Pure Go SQLite implementation (no CGO required)

## License

Part of the AI Together project.
