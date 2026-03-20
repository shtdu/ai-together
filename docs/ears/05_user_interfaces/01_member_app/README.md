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

## Functional Requirements (EARS Format)

### 1. Background Operation

**Purpose:** Enable unobtrusive background operation.

#### State-Driven Requirements (Minimized State)

- **MA-05-001:** `While the member app is minimized to the system tray, the system shall continue routing AI requests.`
- **MA-05-002:** `While the member app is running in the background, the system shall display status via the tray icon.`
- **MA-05-003:** `While the member app is running, the system shall allow quitting from the system tray menu.`

---

### 2. Provider Status Display

**Purpose:** Show current status of configured providers.

#### Event-Driven Requirements (Status Display)

- **MA-05-101:** `When a member views provider status, the system shall display provider names and their current status.`
- **MA-05-102:** `When a provider is healthy (no recent failures), the system shall display a green 🟢 indicator.`
- **MA-05-103:** `When a provider is unhealthy (3+ consecutive failures or single 429 response), the system shall display a red 🔴 indicator.`
- **MA-05-104:** `When a member views provider status, the system shall display today's request count for each provider.`
- **MA-05-105:** `When a member views provider status, the system shall display success rate for each provider.`
- **MA-05-106:** `When a member views provider status, the system shall display average response time for each provider.`

---

### 3. Personal Usage Analytics

**Purpose:** Display personal AI usage statistics.

#### Event-Driven Requirements (Usage Display)

- **MA-05-201:** `When a member views personal usage, the system shall display token usage statistics.`
- **MA-05-202:** `When a member views personal usage, the system shall display estimated costs.`
- **MA-05-203:** `When a member views request history, the system shall display a list of recent requests with timestamps, models, and token counts.`
- **MA-05-204:** `When a member filters personal analytics by date range, the system shall display data for the selected period.`

---

### 4. Server Connection

**Purpose:** Enable connection to team server for configuration sync.

#### Event-Driven Requirements (Connection Workflow)

- **MA-05-301:** `When a member connects to the team server, the system shall authenticate the user with email and password.`
- **MA-05-302:** `When a member connects to the team server, the system shall download the latest provider configuration.`
- **MA-05-303:** `When a member disconnects from the team server, the system shall function in standalone mode with local configuration.`
- **MA-05-304:** `When the member app detects configuration updates from the server, the system shall automatically download and apply the updates.`

---

### 5. Configuration Sync

**Purpose:** Handle configuration synchronization with server.

#### Event-Driven Requirements (Sync Operations)

- **MA-05-401:** `When a member clicks "Pull Configuration", the system shall check for server updates.`
- **MA-05-402:** `When configuration updates are available, the system shall download and apply the new configuration.`
- **MA-05-403:** `When configuration is applied, the system shall restart the proxy with the new configuration.`

---

### 6. Platform Support

**Purpose:** Define supported platforms and requirements.

#### Ubiquitous Requirements

- **MA-05-501:** `The system shall support Windows 10 (64-bit) or later.`
- **MA-05-502:** `The system shall support macOS 12 (Monterey) or later (Intel and Apple Silicon).`
- **MA-05-503:** `The system shall support Ubuntu 20.04 or later.`

---

### 7. AI Agent Behavior Reports

**Purpose:** Display AI agent behavior analysis and activity reports to help users understand agent actions.

#### Event-Driven Requirements (KPI Metrics Display)

- **MA-05-601:** `When a member views AI agent behavior reports, the system shall display total sessions per day.`
- **MA-05-602:** `When a member views AI agent behavior reports, the system shall display agent actions taken (PostToolUse events).`
- **MA-05-603:** `When a member views AI agent behavior reports, the system shall display total time spent in AI sessions.`
- **MA-05-604:** `When a member views AI agent behavior reports, the system shall display autonomy ratio (actions per user prompt).`

#### Event-Driven Requirements (Visualizations)

- **MA-05-701:** `When a member views AI agent behavior reports, the system shall display a timeline Gantt chart showing session activity over time.`
- **MA-05-702:** `When a member views AI agent behavior reports, the system shall display an interaction breakdown chart (permission requests vs prompts).`
- **MA-05-703:** `When a member views AI agent behavior reports, the system shall display prompt statistics (average, max, min length).`
- **MA-05-704:** `When a member views AI agent behavior reports, the system shall display a session table with detailed information.`

#### Event-Driven Requirements (Date Navigation)

- **MA-05-801:** `When a member navigates AI agent behavior reports by date, the system shall display data for the selected date.`
- **MA-05-802:** `When a member navigates AI agent behavior reports, the system shall limit date selection to within the retention period.`

#### State-Driven Requirements (Data Retention)

- **MA-05-901:** `While an Open Source license is active, the system shall retain AI agent behavior data for 7 days.`
- **MA-05-902:** `While a Commercial license is active, the system shall retain AI agent behavior data for 90 days.`

---

### 8. Installation

**Purpose:** Provide platform-specific installation capabilities.

#### Ubiquitous Requirements (Platform Support)

- **MA-05-1001:** `The system shall provide a standard installer for Windows (exe).`
- **MA-05-1002:** `The system shall provide a DMG installer for macOS.`
- **MA-05-1003:** `The system shall provide an AppImage for Linux.`

#### Event-Driven Requirements (Windows Installation)

- **MA-05-1101:** `When a user runs the Windows installer, the system shall install the app to Program Files.`
- **MA-05-1102:** `When a user runs the Windows installer, the system shall create a desktop shortcut.`
- **MA-05-1103:** `When a user runs the Windows installer, the system shall offer an auto-start option.`

#### Event-Driven Requirements (macOS Installation)

- **MA-05-1201:** `When a user opens the macOS DMG, the system shall display the app for drag-to-Applications.`
- **MA-05-1202:** `When a user drags the app to Applications, the system shall install the app to /Applications.`

#### Event-Driven Requirements (Linux Installation)

- **MA-05-1301:** `When a user downloads the AppImage, the system shall provide the file as executable.`
- **MA-05-1302:** `When a user runs the AppImage, the system shall launch the app without installation.`

---

### 9. Auto-Start and Background Operation

**Purpose:** Enable the app to start automatically and run unobtrusively.

#### Event-Driven Requirements (Auto-Start)

- **MA-05-1401:** `When a user enables auto-start, the system shall launch the app on system boot.`
- **MA-05-1402:** `When a user disables auto-start, the system shall not launch the app on system boot.`
- **MA-05-1403:** `When the app starts automatically, the system shall start minimized to system tray.`

#### State-Driven Requirements (Tray Operation)

- **MA-05-1501:** `While the app is running in the tray, the system shall display the proxy status via tray icon.`
- **MA-05-1502:** `While the app is running in the tray, the system shall allow toggling each proxy from the tray menu.`
- **MA-05-1503:** `While the app is running in the tray, the system shall allow opening the main window from the tray menu.`
- **MA-05-1504:** `While the app is running in the tray, the system shall allow quitting from the tray menu.`

---

### 10. Keyboard Shortcuts

**Purpose:** Provide keyboard shortcuts for common actions.

#### Event-Driven Requirements (Shortcuts)

- **MA-05-1601:** `When a user presses Cmd/Ctrl+T, the system shall toggle the proxy on/off.`
- **MA-05-1602:** `When a user presses Cmd/Ctrl+L, the system shall display the logs window.`
- **MA-05-1603:** `When a user presses Cmd/Ctrl+,, the system shall open the settings.`
- **MA-05-1604:** `When a user presses Cmd/Ctrl+Q, the system shall quit the app.`

---

### 11. Notifications

**Purpose:** Notify users of important events.

#### Event-Driven Requirements (Notification Triggers)

- **MA-05-1701:** `When an important event occurs, the system shall display a notification to the user.`
- **MA-05-1702:** `When a proxy fails to start, the system shall display a notification indicating the failure.`
- **MA-05-1703:** `When a configuration update is applied, the system shall display a notification indicating the update.`
- **MA-05-1704:** `When the server connection status changes, the system shall display a notification.`
- **MA-05-1705:** `When a user disables notifications, the system shall not display notifications.`

#### Unwanted Behaviour Requirements

- **MA-05-1801:** `If a notification fails to display, then the system shall log the event without interrupting operation.`

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
- AI agent behavior reports display correctly

---

**Related:** [05.02 Manager Dashboard](../02_manager_dashboard/) | [05.03 Accessibility](../03_accessibility/)
