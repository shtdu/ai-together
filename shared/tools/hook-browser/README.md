# Hook Browser

A terminal user interface (TUI) for browsing and searching through hook event data collected by the AI Together platform. Built with [tview](https://github.com/rivo/tview) featuring a modern purple-themed design.

## Features

### Core Functionality
- **Event Browsing**: View hook events from sessions in a clean two-pane layout
- **Tree View**: Hierarchical view showing event lifecycle relationships (toggle with `t`)
- **Session Management**: Switch between multiple sessions easily
- **JSON Display**: View complete event data with proper formatting
- **Clipboard Integration**: Copy event contents to system clipboard

### Enhanced Features
- **Search & Filter**: Incremental search through events with match highlighting
- **Navigation Progress**: See current position in event list (e.g., "Event 5/42")
- **Help Screen**: Quick reference for all keyboard shortcuts
- **Notifications**: Color-coded toast messages for actions and errors
  - Green: Success
  - Red: Errors
  - Yellow: Warnings
  - Blue: Info

## Installation

```bash
# Build from source
go build -o hook-browser

# Or use the root Makefile
make build
```

## Usage

### Browse Default Bucket

```bash
# Run with default database
./hook-browser

# Run with custom database path
./hook-browser --db-path /path/to/hook-events.sqlite
```

### Browse Tool-Specific Bucket

Use the `--tool-name` flag to browse events from a specific tool's bucket:

```bash
# Browse Claude Code events (from sessions.claude-code bucket)
./hook-browser --tool-name claude-code

# Browse Codex events (from sessions.codex bucket)
./hook-browser --tool-name codex
```

This allows you to browse events collected from different tools without mixing them together.

### Generate Summary Report

Use the `-summary` flag to generate a JSON summary report for a specific day:

```bash
# Generate summary report for a specific date (all tools)
./hook-browser -summary -date 2026-01-15

# Generate summary report for a specific date and tool
./hook-browser -summary -date 2026-01-15 -tool-name claude-code

# With custom database path
./hook-browser -summary -date 2026-01-15 -db-path /path/to/hook-events.sqlite
```

**Flags**:
- `-summary`: Generate summary report as JSON (output to stdout, then exit)
- `-date YYYY-MM-DD`: **Required** - Target date for the report (format: YYYY-MM-DD, local timezone)
- `-tool-name NAME`: Optional - Filter to specific tool (claude-code, codex, etc.)

**JSON Output Format**:

```json
{
  "date": "2026-01-15",
  "timezone": "EST",
  "generated_at": "2026-02-05T10:30:00-05:00",
  "tools": [
    {
      "tool_name": "claude-code",
      "details": {
        "session_stats": {
          "session_count": 3,
          "event_count": 45,
          "unique_sessions": [
            {
              "session_id": "session-id-1",
              "begin_time": "2026-01-15T09:00:00-05:00",
              "last_time": "2026-01-15T09:15:30-05:00",
              "duration_ms": 930000
            },
            {
              "session_id": "session-id-2",
              "begin_time": "2026-01-15T10:00:00-05:00",
              "last_time": "2026-01-15T10:05:00-05:00",
              "duration_ms": 300000
            },
            {
              "session_id": "session-id-3",
              "begin_time": "2026-01-15T14:20:00-05:00",
              "last_time": "2026-01-15T16:45:30-05:00",
              "duration_ms": 8730000
            }
          ]
        },
        "event_type_counts": {
          "UserPromptSubmit": 5,
          "PreToolUse": 15,
          "PostToolUse": 15,
          "Stop": 5,
          "SessionEnd": 2
        },
        "tool_usage_stats": [
          {"tool_name": "Read", "count": 8},
          {"tool_name": "Write", "count": 5},
          {"tool_name": "Glob", "count": 2}
        ],
        "prompt_stats": {
          "total_prompts": 5,
          "average_prompt_length": 245.5,
          "longest_prompt": 1250,
          "shortest_prompt": 12
        }
      }
    }
  ]
}
```

**Key Points**:
- Report represents a **single natural day** (midnight to midnight in local timezone)
- `tool_usage_stats` refers to tools used WITHIN events (Read, Write, Glob, etc.), not hook-browser tools
- Tools with no events for the target day are omitted from the report
- All statistics are filtered at SQL level for performance
- Prompt length is calculated as UTF-8 character count

## Bucket Naming Convention

- **Default bucket** (no `--tool-name`): `sessions`
- **Tool-specific bucket** (with `--tool-name`): `sessions.{tool_name}`
  - Example: `--tool-name claude-code` → `sessions.claude-code`
  - Example: `--tool-name codex` → `sessions.codex`

## Key Bindings

### Navigation
| Key | Action |
|-----|--------|
| ↑/↓ or j/k | Navigate events one by one |
| J (Shift+j) | Page down (skip 10 events) |
| K (Shift+k) | Page up (skip 10 events) |
| PageUp/Down | Scroll in JSON details |

### Actions
| Key | Action |
|-----|--------|
| t | Toggle between List view and Tree view |
| c | Copy current event to clipboard |
| r | Refresh events from database |
| ? | Show help screen |
| q | Quit application |

### Session Management
| Key | Action |
|-----|--------|
| s | Open session picker |
| Enter | Select session |
| Esc | Close session picker / Clear search |

### Search
| Key | Action |
|-----|--------|
| / | Open search bar |
| n | Next search match |
| N | Previous search match |
| Esc | Clear search |

### System
| Key | Action |
|-----|--------|
| Ctrl+C | Force quit |

## Layout

### List View Mode (Default)

```
┌─────────────────────────────────────────────────────────┐
│ ┌─────────────┬───────────────────────────────────────┐ │
│ │ Events (25%)│ Event Details (75%)                   │ │
│ │             │                                       │ │
│ │ 1 - UserPromptSubmit│ • JSON content display        │ │
│ │ 2 - PreToolUse  │ • Scrollable                      │ │
│ │ 3 - PostToolUse │ • Formatted output                │ │
│ │ 4 - Stop        │                                   │ │
│ │ [↑/↓ scroll]   │                                   │ │
│ └─────────────┴───────────────────────────────────────┘ │
│ Session: abc123 | Events: 42 | Mode: List | t:tree | ...│
└─────────────────────────────────────────────────────────┘
```

### Tree View Mode (Press `t` to toggle)

```
┌─────────────────────────────────────────────────────────┐
│ ┌─────────────┬───────────────────────────────────────┐ │
│ │ Tree (35%)  │ Event Details (65%)                   │ │
│ │             │                                       │ │
│ │ ▾ SessionStart│ • JSON content display              │ │
│ │   ▾ Prompt #1│ • Scrollable                        │ │
│ │     ▾ Agentic Loop│ • Formatted output              │ │
│ │       ├─ PreToolUse│                               │ │
│ │       ├─ PostToolUse│                              │ │
│ │       └─ Stop│                                     │ │
│ │   ▸ Prompt #2│                                     │ │
│ └─────────────┴───────────────────────────────────────┘ │
│ Session: abc123 | Events: 42 | Mode: Tree | t:list | ...│
└─────────────────────────────────────────────────────────┘
```

**Tree Structure**:
- `SessionStart` → Root node with session timestamp
- `Prompt #N` → Each user prompt creates a new branch
- `Agentic Loop` → Groups tool execution events (PreToolUse, PostToolUse, etc.)
- `Stop` → Marks end of response for each prompt
- `PreCompact` / `SessionEnd` → Terminal events under root
- `Notification (async)` → Async events linked to root

**Tree Navigation**:
- `Enter` on branch nodes → Expand/collapse
- `Enter` on leaf nodes → View event details
- `t` → Toggle back to list view

## Color Scheme

The interface uses a carefully chosen purple theme:
- **Primary Purple** (#7D56F4): Borders and headers
- **Highlight Pink** (#EE6FF8): Selected items and emphasis
- **Dark Purple** (#3E387D): Status bar background
- **Off-White** (#FAFAFA): Normal text
- **Gray** (#6B6B8A): Dimmed text

## Session Management

When multiple sessions exist in the database, the session picker is displayed automatically:
- Sessions are sorted by most recently used
- Each session shows: ID, time ago, event count
- Format: `abc123  2h ago  42 events`

## Error Handling

The application provides clear feedback for errors:
- Clipboard failures show red error notification
- Database errors are displayed with context
- Invalid operations show warning messages
- All notifications auto-dismiss after 5 seconds

## Data Source

The hook-browser reads from a SQLite database containing hook event data. The default database path is:
- **Default**: `$HOME/.code-together/hook-events.sqlite`
- **Custom**: `--db-path /path/to/hook-events.db`

## Troubleshooting

### "No sessions found" Error
- Verify database path with `--db-path` flag
- Check database contains events for your tool
- Ensure proper file permissions

### Clipboard Not Working
- macOS: Requires no additional setup
- Linux: Install `xclip`: `sudo apt install xclip`
- Windows: Built-in support

### Terminal Issues
- Ensure terminal supports UTF-8
- Use a modern terminal emulator (iTerm2, Terminal.app, gnome-terminal)
- Verify `$TERM` is set correctly

## Performance

- Efficient rendering with tview
- Incremental search for instant feedback
- Optimized for sessions with hundreds of events

## Development

### Project Structure
```
internal/model/
├── app.go           # App state and initialization
├── layout.go        # UI layout setup
├── handlers.go      # Input handling
├── styles.go        # Color definitions
├── types.go         # Type definitions and formatters
├── notifications.go # Toast notification system
└── search.go        # Search and filter functionality
```

### Adding New Features

1. **New Key Bindings**: Add to `handlers.go`
2. **UI Components**: Extend `layout.go`
3. **Color Changes**: Update `styles.go`
4. **Data Operations**: Modify `app.go`

## Contributing

When adding features:
1. Follow existing code patterns
2. Maintain the purple color scheme
3. Add keyboard shortcuts to help screen
4. Test with large event sets (100+ events)
5. Ensure cross-platform compatibility

## License

Part of the AI Together project. See main repository LICENSE.
