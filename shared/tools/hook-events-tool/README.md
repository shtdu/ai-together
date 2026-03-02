# Hook Events Tool

A tool to split Claude Code hook events from the log file into organized JSON files by session and sequence.

## Usage

```bash
# Build
go build -o hook-events-tool

# Run with default output (current directory)
./hook-events-tool

# Run with custom output folder
./hook-events-tool -o /path/to/output
```

## Input

Log file location: `~/.claude/job_done.log`

## Output Structure

```
{output_folder}/
└── <session_id>/
    ├── 01_PreToolUse.json
    ├── 02_PostToolUse.json
    ├── 03_Stop.json
    └── ...
```

**Filename format:** `{sequence_id:02d}_{hook_event_name}.json`

- `sequence_id`: Two-digit sequence order (01, 02, 03, ...) of events within the same session
- `hook_event_name`: Event type name (e.g., PreToolUse, PostToolUse, Stop, SessionEnd)

## Event Counting

Use `count_events.sh` to count event types per session:

```bash
# Count events for all sessions in a folder
./count_events.sh {output_folder}

# Example output:
# === Session: 1234-abcd-5678 ===
#      5 PostToolUse
#      3 PreToolUse
#      1 SessionEnd
```

## Technical Requirements

- Uses only Go standard library
- Parses JSON objects from log file
- Maintains event order by sequence within each session

## Notes:

7 Events full session ids:

```
fa64ed95-f485-4050-be58-17b41000d105
a1348ff0-0f7e-4cae-9752-2bdbec828912
74e58d8e-5dca-4347-933d-9b64e5717968 (small)
```








