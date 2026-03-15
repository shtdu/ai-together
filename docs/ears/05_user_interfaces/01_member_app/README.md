# Member App

The Member desktop application runs on developers' computers, providing AI tool routing and personal usage insights.

## Purpose

Give members a simple, unobtrusive way to use AI tools with team-configured providers.

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

---

**Related:** [05.02 Manager Dashboard](../02_manager_dashboard/) | [05.03 Accessibility](../03_accessibility/)
