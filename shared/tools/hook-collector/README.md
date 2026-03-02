# Hook Event Collector

A CLI tool that collects Claude Code hook events via stdin, stores them in a local BoltDB database, and provides export/cleanup capabilities.

## Overview

The hook event collector is designed to work seamlessly with Claude Code hooks. It follows the Unix philosophy - each invocation processes one event and exits. Events are stored as raw JSON without parsing or validation, keeping the tool simple and fast.

## Installation

```bash
cd shared/tools/hook-collector
go build -o hook-collector
```

Move the binary to your PATH:

```bash
sudo mv hook-collector ~/.code-together/hook-collector
```

## Quick Start

### 1. Collect an Event (Default Bucket)

The `collect` command reads a JSON event from stdin and stores it in the database:

```bash
echo '{"session_id":"test-123","hook_event_name":"UserPromptSubmit","prompt":"hello"}' | \
  hook-collector collect
```

### 2. Collect an Event (Tool-Specific Bucket)

Use the `--tool-name` flag to store events in a tool-specific bucket:

```bash
echo '{"session_id":"test-123","hook_event_name":"UserPromptSubmit","prompt":"hello"}' | \
  hook-collector --tool-name claude collect
```

This stores events in the `sessions.claude` bucket instead of the default `sessions` bucket.

### 3. Export Events

Export all events from the default bucket to stdout:

```bash
hook-collector export
```

Export events from a specific tool bucket:

```bash
hook-collector --tool-name claude export
```

Export all events to a file:

```bash
hook-collector export events.jsonl
```

Export events from the last 7 days:

```bash
hook-collector --since 7d export events.jsonl
```

Export to stdout with filtering:

```bash
hook-collector --since 7d export | jq '.hook_event_name'
```

### 4. Clean Old Events

Remove events older than 90 days from the default bucket:

```bash
hook-collector --older-than 90d clean
```

Remove events from a specific tool bucket:

```bash
hook-collector --tool-name claude --older-than 90d clean
```

## Claude Code Hook Setup

Configure Claude Code hooks in `~/.claude/settings.json` or project `.claude/settings.json`:

### Default Bucket Configuration

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector collect"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector collect"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector collect"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector collect"
          }
        ]
      }
    ]
  }
}
```

### Tool-Specific Bucket Configuration

To isolate events for different tools, use the `--tool-name` flag:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector --tool-name claude collect"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector --tool-name claude collect"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector --tool-name claude collect"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq -cM | hook-collector --tool-name claude collect"
          }
        ]
      }
    ]
  }
}
```

This stores events in the `sessions.claude` bucket, allowing you to collect events from multiple tools without mixing them.
```

## Configuration

The collector uses a configuration file at `.code-together/hook-collector.json` in your workspace root:

```json
{
  "database_path": "./.code-together/hook-events.db",
  "retention_days": 90
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `database_path` | string | `./.code-together/hook-events.db` | Path to the BoltDB database file |
| `retention_days` | int | `90` | Number of days to retain events (used by clean command) |

## Commands

### `collect` - Collect Event from Stdin

Collects a single JSON event from stdin and stores it in the database.

```bash
cat event.json | hook-collector collect [flags]
```

**Flags:**
- `--overwrite` - Overwrite existing events (default: skip duplicates)
- `--verbose` - Show detailed processing information

### `export` - Export Events

Exports events from the database to a file or stdout.

```bash
hook-collector [flags] export [output-file]
```

**Note:** Flags must come before the command. The output file is optional - if not provided, exports to stdout.

**Flags:**
- `--format <format>` - Export format: `json`, `jsonl`, `csv` (default: `jsonl`)
- `--since <time>` - Export events since this time (e.g., `7d`, `30d`, `2024-01-01`, `2024-01-01T10:00:00Z`)
- `--until <time>` - Export events until this time
- `--session-id <id>` - Export only this session
- `--event-type <type>` - Export only this event type
- `--compress` - Compress output with gzip

**Examples:**

```bash
# Export all events to stdout
hook-collector export

# Export all events to file
hook-collector export events.jsonl

# Export last 7 days, compressed
hook-collector --since 7d --compress export events.jsonl.gz

# Export specific session to JSON
hook-collector --session-id 74e58d8e-5dca-4347-933d-9b64e5717968 --format json export session.json

# Export specific event type to stdout
hook-collector --event-type UserPromptSubmit export | jq '.prompt'

# Export from a specific date
hook-collector --since 2024-01-01 export

# Export from a specific datetime
hook-collector --since "2024-01-01T10:00:00Z" export
```

### `clean` - Clean Old Events

Removes old events from the database based on retention policy.

```bash
hook-collector clean [flags]
```

**Flags:**
- `--older-than <time>` - Remove events older than (e.g., `30d`, `1y`). Defaults to `retention_days` from config
- `--dry-run` - Show what would be deleted without deleting
- `--compact` - Compact database after cleanup

**Examples:**

```bash
# Remove events older than 90 days (default)
hook-collector clean

# Remove events older than 30 days and compact database
hook-collector --older-than 30d --compact clean

# Dry run to see what would be deleted
hook-collector --older-than 30d --dry-run clean
```

## Global Flags

These flags apply to all commands:

- `--db-path <path>` - Path to BoltDB database
- `--config <path>` - Path to config file
- `--tool-name <name>` - Tool name for bucket isolation (e.g., `claude`, `codex`). Uses default bucket if not specified.
- `--verbose` - Enable verbose logging
- `--quiet` - Suppress non-error output

## Storage Schema

The tool uses BoltDB with a nested bucket structure. When using the `--tool-name` flag, events are stored in tool-specific buckets:

### Default Bucket (no --tool-name flag)

```
[sessions] bucket
  └── session_id → nested bucket
      ├── 1 → event JSON (with ts field added)
      ├── 2 → event JSON
      └── ...
```

### Tool-Specific Bucket (with --tool-name flag)

```
[sessions.claude] bucket
  └── session_id → nested bucket
      ├── 1 → event JSON (with ts field added)
      ├── 2 → event JSON
      └── ...

[sessions.codex] bucket
  └── session_id → nested bucket
      ├── 1 → event JSON (with ts field added)
      ├── 2 → event JSON
      └── ...
```

This allows you to collect and manage events from multiple tools without mixing them.

**Key Design Decisions:**

1. **No metadata bucket** - Session info computed on-demand:
   - Event count: `sessionBucket.Stats().KeyN` (instant, no iteration)
   - First/last timestamps: Read first event (sequence 1) and last event (max sequence)
2. **No parsing** - Events stored as raw JSON bytes
3. **Minimal processing** - Only extract `session_id` for bucket lookup and add `ts` timestamp
4. **Bucket isolation** - Each tool can have its own bucket, preventing event mixing

## Data Format

Events are stored as raw JSON with an added `ts` field:

```json
{
  "session_id": "74e58d8e-5dca-4347-933d-9b64e5717968",
  "hook_event_name": "UserPromptSubmit",
  "cwd": "/path/to/workspace",
  "prompt": "Help me write a function",
  "ts": "2024-01-25T12:34:56.789Z"
}
```

## Supported Event Types

- `UserPromptSubmit` - User submits a prompt
- `PreToolUse` - Before tool execution
- `PostToolUse` - After tool execution
- `Stop` - Claude Code finishes responding
- `SessionEnd` - Session ends
- `PreCompact` - Before compact operation
- `SubagentStop` - Subagent task completes

## Testing

Run unit tests:

```bash
go test -v ./shared/tools/hook-collector/internal/...
```

Run tests with coverage:

```bash
go test -cover ./shared/tools/hook-collector/internal/...
```

## Troubleshooting

### Error Logging

When the collector encounters an error (such as missing session_id, invalid JSON, or database issues), it automatically logs the error and the original payload to `~/.code-together/collector.log`. This helps with debugging hook configurations and identifying problematic events.

**Log format:**
```
[2026-01-27T13:19:29Z] ERROR: missing session_id in event
Payload:
{"test": "data"}
```

**Common errors logged:**
- `missing session_id in event` - Event JSON doesn't contain a `session_id` field
- `session_id is not a string` - The `session_id` field is not a string type
- `failed to parse JSON` - Event is not valid JSON
- `event too large` - Event exceeds 10MB size limit

### "failed to open database"

Ensure the `.code-together` directory exists in your workspace root:

```bash
mkdir -p .code-together
```

### "missing session_id in event"

The event JSON must contain a `session_id` field. Ensure your hook configuration includes valid events. Check the collector log for the actual payload:

```bash
tail -20 ~/.code-together/collector.log
```

### Database file locked

The BoltDB file may be locked by another process. Wait a moment and try again, or close other applications that may be accessing the file.

## Development

### Project Structure

```
shared/tools/hook-collector/
├── main.go                    # CLI entry point
├── commands/                   # Command implementations
│   ├── collect.go             # Collect command
│   ├── export.go              # Export command
│   └── clean.go               # Clean command
├── internal/                   # Internal packages
│   ├── storage/
│   │   ├── storage.go          # BoltDB operations
│   │   └── storage_test.go     # Storage tests
│   ├── collector/
│   │   ├── collector.go        # Event collection logic
│   │   ├── collector_test.go   # Collector tests
│   │   ├── logger.go           # Error logging to ~/.code-together/collector.log
│   │   └── logger_test.go      # Logger tests
│   └── config/
│       ├── config.go           # Configuration management
│       └── config_test.go      # Config tests
├── go.mod
└── README.md
```

### Dependencies

- `go.etcd.io/bbolt` - BoltDB database library
- Standard library only for CLI, JSON, logging

## License

MIT

## Contributing

Contributions are welcome! Please run tests before submitting:

```bash
go test ./shared/tools/hook-collector/internal/...
go build ./shared/tools/hook-collector
```
