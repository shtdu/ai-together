# Hook Schema Analyzer

A tool for parsing Claude Code hook log files and generating Go struct definitions and JSON Schema v7 specifications for each event type found in the logs.

## What It Does

This analyzer:
1. Parses JSON-formatted log files from Claude Code hooks
2. Groups events by `hook_event_name`
3. Infers field types from actual log data
4. Generates Go struct files with proper JSON tags
5. Generates JSON Schema v7 (`draft/2020-12`) files
6. Properly identifies required fields (those present in 100% of instances)

## Installation

The tool is located at `shared/tools/hook-schema-analyzer/`. It uses the workspace's Go module - no separate installation needed.

## Usage

### Generate Schemas

```bash
cd shared/tools/hook-schema-analyzer
go run main.go <log-file>
```

Example:
```bash
go run main.go ../../../docs/hooks/job_done.log
```

This will generate:
- Go struct files: `<event-name>.go`
- JSON Schema files: `<event-name>.json`

All output goes to `../../events/` (i.e., `shared/events/`).

### Validate Required Fields

Check if the "required" fields in schemas match the actual log data:

```bash
go run main.go validate <log-file> <schema-dir>
```

Example:
```bash
go run main.go validate ../../../docs/hooks/job_done.log ../../events
```

This validates each schema's required fields against the log data and reports:
- Which fields are truly required (present in 100% of instances)
- Which fields marked as required are optional (appear in <100% of instances)
- Optional fields with high presence (≥95%) that might need to be considered

## Output Format

### Go Struct Files

```go
package events

// Runs before tool calls (can block them)
type PreToolUse struct {
    // Current working directory
    Cwd string `json:"cwd"`
    // Permission mode (e.g., 'default', 'acceptEdits')
    PermissionMode string `json:"permission_mode"`
    // Unique identifier for session
    SessionId string `json:"session_id"`
    // Input parameters for the tool
    ToolInput map[string]interface{} `json:"tool_input"`
    // Name of the tool being called
    ToolName string `json:"tool_name"`
    // Unique identifier for this tool use
    ToolUseId string `json:"tool_use_id"`
    // Path to the transcript file
    TranscriptPath string `json:"transcript_path"`
}
```

### JSON Schema Files

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "description": "Runs before tool calls (can block them)",
  "properties": {
    "cwd": {
      "description": "Current working directory",
      "type": "string"
    },
    "permission_mode": {
      "description": "Permission mode (e.g., 'default', 'acceptEdits')",
      "type": "string"
    },
    "session_id": {
      "description": "Unique identifier for session",
      "type": "string"
    },
    "tool_input": {
      "description": "Input parameters for the tool",
      "type": "object"
    },
    "tool_name": {
      "description": "Name of the tool being called",
      "type": "string"
    },
    "tool_use_id": {
      "description": "Unique identifier for this tool use",
      "type": "string"
    },
    "transcript_path": {
      "description": "Path to the transcript file",
      "type": "string"
    }
  },
  "required": [
    "cwd",
    "permission_mode",
    "session_id",
    "tool_input",
    "tool_name",
    "tool_use_id",
    "transcript_path"
  ],
  "title": "PreToolUse",
  "type": "object"
}
```

## Features

### Type Inference

The tool infers Go and JSON types from actual log data:
- `string` → `string`
- `number` → `float64` / `number`
- `boolean` → `bool` / `boolean`
- `object` → `map[string]interface{}` / `object`
- `array` → `[]interface{}` / `array`
- `null` → `interface{}` / `null`

Fields with multiple possible types (e.g., `string | null`) become `interface{}` in Go.

### Field Name Handling

- **snake_case** → **PascalCase**: `file_path` → `FilePath`
- **CLI flags**: Mapped to descriptive names
  - `-i` → `CaseInsensitive`
  - `-A` → `AfterContext`
  - `-B` → `BeforeContext`
  - `-C` → `ContextLines`
  - `-n` → `ShowLineNumbers`
- **Reserved keywords**: `type` → `TypeField`
- **Duplicates**: Automatically suffixed (`FilePath1`, `FilePath2`, etc.)

### Required Fields

A field is marked as required only if it appears in 100% of the log instances for that event type. This is validated against actual log data.

### Field Descriptions

The tool includes documentation for known fields from the Claude Code hooks guide. These appear as:
- Go struct comments
- JSON Schema `description` properties

## Supported Event Types

The analyzer can process any hook event type. From the current logs:

- **PreToolUse**: Runs before tool calls (can block them)
- **PostToolUse**: Runs after tool calls complete
- **Stop**: Runs when Claude Code finishes responding
- **SubagentStop**: Runs when subagent tasks complete
- **SessionEnd**: Runs when Claude Code session ends
- **UserPromptSubmit**: Runs when user submits a prompt, before Claude processes it
- **PreCompact**: Runs before Claude Code is about to run a compact operation

## Log File Format

The tool expects line-by-line JSON objects, separated by blank lines:

```json
{
  "hook_event_name": "PreToolUse",
  "session_id": "abc123",
  "cwd": "/path/to/project",
  "tool_name": "Bash",
  "tool_use_id": "toolu_01A1B2C3",
  "tool_input": {
    "command": "ls",
    "description": "List files"
  },
  "permission_mode": "default",
  "transcript_path": "/path/to/transcript.md"
}

{
  "hook_event_name": "PostToolUse",
  "session_id": "abc123",
  ...
}
```

## Troubleshooting

### "No events found in log file"
- Check that the log file contains valid JSON objects
- Ensure objects have the `hook_event_name` field

### "duplicate field" compilation errors
- The tool automatically adds numeric suffixes to duplicate Go field names
- Regenerate schemas after changes

### Required fields seem incorrect
- Run the validation tool: `go run main.go validate <log-file> <schema-dir>`
- Regenerate schemas with the latest log data

## Building for Production

```bash
cd shared/tools/hook-schema-analyzer
go build -o hook-schema-analyzer .
./hook-schema-analyzer <log-file>
```

## License

Part of the Code Together project.
