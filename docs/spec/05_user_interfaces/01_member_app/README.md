# Member App

The Member desktop application runs on developers' computers, providing AI tool routing and personal usage insights.

## Purpose

Give members a simple, unobtrusive way to use AI tools with team-configured providers.

## User Personas

- **Member** - Software developer using AI coding tools as part of daily work
- **Individual Contributor** - Team member who uses Claude Code, Codex, or OpenCode

## User Stories

- As a **member**, I want the app to run in the background so it doesn't interrupt my work
- As a **member**, I want to see which providers are available and their status
- As a **member**, I want to see my AI usage so I can track my consumption
- As a **member**, I want to connect to my team server to get configuration
- As a **member**, I want to see my request history
- As a **member**, I want to view AI agent behavior reports to understand what actions were taken

## Platform Support

| Platform | Minimum Version | Notes |
|----------|-----------------|-------|
| **Windows** | Windows 10 | 64-bit required |
| **macOS** | macOS 12 (Monterey) | Intel and Apple Silicon |
| **Linux** | Ubuntu 20.04+ | Other distros may work |

## Key Features

### 1. Background Operation

**Description:** App runs unobtrusively in the background while routing AI requests.

**Behavior:**
- App minimizes to system tray
- Continues routing AI requests when minimized
- Shows status via tray icon
- Can be quit from system tray menu

### 2. Provider Status

**Description:** Shows current status of all configured providers.

**Displayed Information:**
- Provider names
- Online/offline status
- Today's request count
- Success rate
- Average response time

**Status Indicators:**
- 🟢 Green: Healthy (no recent failures)
- 🔴 Red: Unhealthy (3+ consecutive failures or single 429 response)

### 3. Personal Usage

**Description:** Shows personal AI usage statistics.

**Displayed Information:**
- Today's request count
- Today's token usage
- Today's estimated cost
- Usage heatmap (GitHub-style activity graph)

### 4. Server Connection

**Description:** Manages connection to team server for configuration sync.

**Displayed Information:**
- Connection status (connected/disconnected/error)
- Server URL
- Last sync time
- Pending local changes (if applicable)

**Actions:**
- Connect to server
- Pull configuration
- Login/logout

### 5. AI Agent Behavior Reports

**Description:** Shows AI agent behavior analysis and activity reports.

**KPI Metrics:**
- Total sessions per day
- Agent actions taken (PostToolUse events)
- Total time spent in AI sessions
- Autonomy ratio (actions per user prompt)

**Visualizations:**
- Timeline Gantt chart showing session activity over time
- Interaction breakdown chart (permission requests vs prompts)
- Prompt statistics (average, max, min length)
- Session table with detailed information

**Features:**
- Daily report with date navigation (within retention period)
- Data retention: 7 days (Open Source), 90 days (Commercial)

## User Interface Overview

### Main Window

```
┌─────────────────────────────────────────┐
│ AI Together    [Server] [Settings] [Log] │
├─────────────────────────────────────────┤
│ [Usage Heatmap - GitHub style]          │
├─────────────────────────────────────────┤
│ [Claude] [Codex] [OpenCode]             │
├─────────────────────────────────────────┤
│ Provider Cards:                          │
│ ┌─────────────────────────────────────┐ │
│ │ Anthropic        [ON]  1,250 tokens  │ │
│ │ 98% success  ─────  450ms avg       │ │
│ └─────────────────────────────────────┘ │
│ ┌─────────────────────────────────────┐ │
│ │ Team OpenAI       [ON]    840 tokens  │ │
│ │ 95% success  ─────  380ms avg       │ │
│ └─────────────────────────────────────┘ │
├─────────────────────────────────────────┤
│ Today: 125 requests | 45k tokens | $0.90 │
└─────────────────────────────────────────┘
```

### Server View

```
┌─────────────────────────────────────────┐
│ ← Back                             [Login] │
├─────────────────────────────────────────┤
│ Connection: http://server:8080          │
│ Status: ✅ Connected                    │
├─────────────────────────────────────────┤
│ User: alex@example.com                  │
│ Role: Manager                           │
├─────────────────────────────────────────┤
│ [Pull Configuration]  [Push Configuration] │
│ Last sync: 2 minutes ago                │
└─────────────────────────────────────────┘
```

### System Tray Menu

- Show AI Together
- Enable/disable Claude Code proxy
- Enable/disable Codex proxy
- Enable/disable OpenCode proxy
- Open main window
- Quit

## Functional Requirements

- **FR-001:** App must install with standard installer for platform
- **FR-002:** App must auto-start on option
- **FR-003:** App must minimize to system tray
- **FR-004:** Tray icon must indicate proxy status
- **FR-005:** App must show notifications for important events

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl+T` | Toggle proxy on/off |
| `Cmd/Ctrl+L` | Show logs |
| `Cmd/Ctrl+,` | Open settings |
| `Cmd/Ctrl+Q` | Quit app |

## Installation

### Windows

1. Download `CodeTogether-setup.exe`
2. Run installer
3. App installs to `Program Files`
4. Desktop shortcut created
5. Auto-start option available

### macOS

1. Download `CodeTogether.dmg`
2. Open disk image
3. Drag AI Together to Applications
4. App installs to `/Applications`
5. Launch from Launchpad or Applications

### Linux

1. Download `code-together.AppImage`
2. Make executable: `chmod +x code-together.AppImage`
3. Move to `/usr/local/bin/` or run from location

## Success Criteria

- App installs and launches successfully
- App runs reliably in background
- Proxy routes requests correctly
- Status indicators accurately reflect state
- Personal usage displays accurately

---

**Related:** [05.02 Manager Dashboard](../02_manager_dashboard/) | [05.03 Accessibility](../03_accessibility/) | [Domain 05 Overview](../README.md)
