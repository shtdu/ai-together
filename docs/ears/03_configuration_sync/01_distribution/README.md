# Configuration Distribution

Provider configurations flow from managers to team members automatically through background synchronization.

## Purpose

Ensure all team members have consistent, up-to-date provider configurations without manual effort.

## Functional Requirements (EARS Format)

### 1. Automatic Background Sync

**Purpose:** Enable members to receive configuration updates automatically.

#### State-Driven Requirements (Sync Interval)

- **CD-03-001:** `While a member client is connected to the server, the system shall check for configuration updates every 5 minutes.`
- **CD-03-002:** `While a member client checks for updates, if the server has a newer configuration timestamp, then the system shall download the new configuration.`
- **CD-03-003:** `While a member client downloads new configuration, the system shall apply the configuration and restart the proxy.`

#### Event-Driven Requirements (Sync Trigger)

- **CD-03-101:** `When the background sync timer triggers, the system shall compare the server configuration timestamp with the local timestamp.`
- **CD-03-102:** `When the server configuration is newer, the system shall download the configuration.`
- **CD-03-103:** `When the configuration download completes successfully, the system shall apply the configuration immediately.`

---

### 2. Manual Configuration Pull

**User Story:** As a member, I want to get latest configuration now.

#### Event-Driven Requirements (Manual Pull Workflow)

- **CD-03-201:** `When a member clicks "Pull Configuration", the system shall check the server for configuration updates.`
- **CD-03-202:** `When configuration updates exist on the server, the system shall display a preview of the changes.`
- **CD-03-203:** `When a member confirms "Pull Now", the system shall download and apply the configuration.`
- **CD-03-204:** `When the configuration is applied, the system shall restart the proxy with the new configuration.`

---

### 3. Manager Configuration Push

**User Story:** As a manager, I want to send my changes to the team immediately.

#### Event-Driven Requirements (Push Workflow)

- **CD-03-301:** `When a manager makes provider configuration changes, the system shall prepare the configuration for upload.`
- **CD-03-302:** `When a manager clicks "Push to Server", the system shall upload the configuration to the server.`
- **CD-03-303:** `When the server receives the configuration, the system shall store it as the latest version with a UTC timestamp.`
- **CD-03-304:** `When the server stores the new configuration, the system shall mark it available for all members to sync.`

#### Event-Driven Requirements (Push Validation)

- **CD-03-305:** `When a manager pushes configuration to the server, the system shall validate all required fields are present.`
- **CD-03-306:** `When a manager pushes configuration, the system shall validate API key formats.`
- **CD-03-307:** `When a manager pushes configuration, the system shall validate endpoint URL formats.`
- **CD-03-308:** `When a manager pushes configuration, the system shall validate each tool has at least one enabled provider.`
- **CD-03-309:** `When a manager pushes configuration, the system shall validate provider names are unique within the organization.`
- **CD-03-310:** `When a manager pushes configuration, the system shall validate priority values are positive integers.`

#### Unwanted Behaviour Requirements (Validation Failures)

- **CD-03-401:** `If a configuration validation fails, then the system shall reject the push and display a specific error message.`
- **CD-03-402:** `If a configuration has no enabled providers for a tool, then the system shall reject the push and display which tool needs a provider.`
- **CD-03-403:** `If a configuration has duplicate provider names, then the system shall reject the push and display the duplicate name error.`

---

### 4. Sync Status Display

**Purpose:** Show users the current synchronization state.

#### Event-Driven Requirements (Status Display)

- **CD-03-501:** `When a member client checks sync status, the system shall display the current status indicator.`
- **CD-03-502:** `When configuration is synced (local matches server), the system shall display a green ✅ indicator.`
- **CD-03-503:** `When the server is unreachable, the system shall display a gray 🔌 indicator showing offline status.`
- **CD-03-504:** `When sync status is displayed, the system shall show allowed actions based on current state.`

#### State-Driven Requirements (Status Actions)

- **CD-03-601:** `While sync status shows "Synced", the system shall allow Pull (refresh) and Push (managers only) actions.`
- **CD-03-602:** `While sync status shows "Offline", the system shall only allow viewing cached configuration.`

---

### 5. Configuration Download & Application

**Purpose:** Handle the technical aspects of downloading and applying configuration.

#### Event-Driven Requirements (Download Process)

- **CD-03-701:** `When a member downloads configuration, the system shall complete the download within 10 seconds.`
- **CD-03-702:** `When a member downloads configuration, the system shall validate the configuration schema matches the expected format.`
- **CD-03-703:** `When a member downloads configuration, the system shall validate the configuration version is compatible with the client.`
- **CD-03-704:** `When a configuration download is successful, the system shall test that the proxy can restart with the new configuration.`

#### Unwanted Behaviour Requirements (Download Failures)

- **CD-03-801:** `If a configuration download exceeds 10 seconds, then the system shall show a timeout notification and keep the previous configuration.`
- **CD-03-802:** `If a downloaded configuration is invalid, then the system shall show an error and keep the previous configuration.`
- **CD-03-803:** `If a proxy restart fails with new configuration, then the system shall automatically rollback to the previous configuration.`

---

### 6. Conflict Resolution

**Purpose:** Define how configuration conflicts are handled.

#### Ubiquitous Requirements

- **CD-03-901:** `The system shall use the server configuration as the source of truth.`
- **CD-03-902:** `The system shall use UTC timestamps to determine which configuration version is newer.`
- **CD-03-903:** `The system shall not support conflict resolution—server always wins.`

#### Event-Driven Requirements (Multiple Managers)

- **CD-03-1001:** `When multiple managers push simultaneously, the system shall accept the last push based on server UTC timestamp.`
- **CD-03-1002:** `When a manager pushes configuration to the server, the system shall overwrite the existing server configuration.`

---

### 7. Offline Mode Behavior

**Purpose:** Define how the system behaves when server is unreachable.

#### State-Driven Requirements (Offline State)

- **CD-03-1101:** `While the server is unreachable during a scheduled sync, the system shall use the cached configuration.`
- **CD-03-1102:** `While the server is unreachable, the system shall retry the sync every 5 minutes.`
- **CD-03-1103:** `While a member is offline during a manager push, the system shall receive the update when the member reconnects to the server.`

---

### 8. Configuration Rollback

**Purpose:** Allow managers to revert to previous configurations.

#### Event-Driven Requirements (Rollback Workflow)

- **CD-03-1201:** `When a manager accesses configuration history, the system shall display the past 10 configurations with timestamps.`
- **CD-03-1202:** `When a manager views configuration history, the system shall show who made each change.`
- **CD-03-1203:** `When a manager selects a previous version and clicks "Rollback", the system shall show a preview of the changes.`
- **CD-03-1204:** `When a manager confirms the rollback, the system shall push the previous configuration to the server.`
- **CD-03-1205:** `When a rollback is complete, the system shall distribute the configuration to all members within 5 minutes.`

#### Unwanted Behaviour Requirements (Rollback Limitations)

- **CD-03-1301:** `If a manager attempts to rollback to a version that included a deleted provider, then the system shall display an error indicating the rollback is not possible.`
- **CD-03-1302:** `If a manager attempts to rollback beyond history retention (10 versions), then the system shall display an error indicating the version is not available.`

#### Ubiquitous Requirements (History Retention)

- **CD-03-1401:** `The system shall retain the last 10 configurations.`
- **CD-03-1402:** `The system shall automatically delete configuration versions older than the last 10.`
- **CD-03-1403:** `The system shall clear configuration history when an organization is deleted.`

---

### 9. Configuration Templates

**Purpose:** Provide pre-configured templates for common scenarios.

#### Event-Driven Requirements (Template Application)

- **CD-03-1501:** `When a manager selects a template, the system shall display a preview of what will be configured.`
- **CD-03-1502:** `When a manager customizes a template, the system shall allow modification of required fields (API keys, endpoints).`
- **CD-03-1503:** `When a manager applies a template, the system shall validate the configuration before pushing.`
- **CD-03-1504:** `When a template is applied, the system shall push the configuration to the server and distribute to members.`

#### Event-Driven Requirements (Custom Template Creation)

- **CD-03-1505:** `When a manager saves current configuration as a template, the system shall store the template per organization.`
- **CD-03-1506:** `When a manager creates a template, the system shall exclude sensitive data (API keys) from the template.`
- **CD-03-1507:** `When a manager creates a template, the system shall require a template name and description.`

## Business Rules

- **BR-03-001:** Server configuration is always the source of truth (server wins)
- **BR-03-002:** Manager push to server overwrites server configuration
- **BR-03-003:** Members always pull and apply server configuration (no conflict resolution)
- **BR-03-004:** Configuration timestamps use UTC to determine which version is newer

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Server unreachable during sync | Use cached config, retry in 5 minutes |
| Configuration invalid after download | Show error, keep previous configuration |
| Download timeout (>10 seconds) | Show timeout notification, keep previous config, retry in 5 minutes |
| Member offline during manager push | Member receives update when back online |
| Multiple managers push simultaneously | Last push wins (based on server UTC timestamp) |

## Success Criteria

- Configuration changes reach all members within 5 minutes
- Manual push completes within 10 seconds
- Sync status accurately reflects state
- Conflicts are detected and prevented
- Offline mode uses cached configuration

---

**Related:** [03.02 Teams](../02_teams/) | [03.03 Offline Mode](../03_offline_mode/)
