# Configuration Distribution

Provider configurations flow from managers to team members automatically through background synchronization.

## Purpose

Ensure all team members have consistent, up-to-date provider configurations without manual effort.

## User Stories

- As a **manager**, I want my provider changes to automatically reach my team
- As a **member**, I want to receive provider updates without doing anything
- As a **manager**, I want to push critical updates immediately

## How Distribution Works

### Automatic Sync (Background)

**Behavior:**
- Member clients check for updates every 5 minutes
- If server has newer configuration, download automatically
- Configuration is applied and proxy restarts

**Condition:** Only happens if no local changes exist

### Manual Pull

**User Story:** As a member, I want to get latest configuration now.

**Workflow:**
1. Member clicks "Pull Configuration" button
2. System checks server for updates
3. If updates exist, preview is shown
4. Member confirms: "Pull Now"
5. Configuration downloads and applies
6. Proxy restarts with new configuration

### Manager Push

**User Story:** As a manager, I want to send my changes to the team immediately.

**Workflow:**
1. Manager makes provider configuration changes
2. Manager clicks "Push to Server" button
3. Configuration uploaded to server
4. Server stores as latest version
5. All members auto-pull within 5 minutes

## Sync Status Display

Members see current sync status:

| Status | Indicator | Meaning | Allowed Actions |
|--------|-----------|---------|-----------------|
| **Synced** | ✅ Green | Local matches server | Pull (refresh), Push (managers only) |
| **Offline** | 🔌 Gray | Server unreachable | View cached config only |

## Configuration Flow

```
Manager Action:
1. Manager edits providers in Manager UI or Member app
2. Manager clicks "Push to Server"
3. Server stores configuration with UTC timestamp

Member Reception:
1. Background timer triggers (every 5 min)
2. Member checks server timestamp vs local
3. If server newer: download configuration
4. Apply configuration and restart proxy

Edge Cases:
- Member offline during push → Update received when member reconnects
- Multiple managers push → Last push wins (based on server UTC timestamp)
```

## Functional Requirements

- **FR-001:** Members must check for updates every 5 minutes (configurable per organization)
- **FR-002:** Configuration download must complete within 10 seconds or fail with timeout error
- **FR-003:** Configuration must apply immediately after successful download
- **FR-004:** Managers can push configuration to server
- **FR-005:** Sync status must be visible to users

## Business Rules

- **BR-001:** Server configuration is always the source of truth (server wins)
- **BR-002:** Manager push to server overwrites server configuration
- **BR-003:** Members always pull and apply server configuration (no conflict resolution)
- **BR-004:** Configuration timestamps use UTC to determine which version is newer

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

## Configuration Validation

Before a configuration is distributed to members, it is validated to prevent problems.

### Validation Rules

**Server-Side Validation (on push):**

| Check | Description | Failure Action |
|-------|-------------|----------------|
| **Required Fields** | All provider required fields present | Reject push, show missing fields |
| **API Key Format** | API key matches expected format | Reject push, show format error |
| **Endpoint URL** | Valid URL format | Reject push, show URL error |
| **At Least One Provider** | Each tool has ≥1 enabled provider | Reject push, show which tool needs provider |
| **Unique Names** | Provider names unique within organization | Reject push, show duplicate name |
| **Priority Values** | Priority is positive integer | Reject push, show priority error |

**Client-Side Validation (on pull):**

| Check | Description | Failure Action |
|-------|-------------|----------------|
| **Schema Validation** | Configuration matches expected schema | Reject pull, show validation error |
| **Version Compatibility** | Configuration version supported by client | Reject pull, show incompatibility warning |
| **Proxy Restart Test** | Proxy can restart with new config | Reject pull, keep previous config |

### Validation Feedback

**On Validation Failure:**
- Clear error message explaining what's wrong
- Specific field or section highlighted
- Suggestion for how to fix (when possible)
- Configuration not saved/pushed until valid

**Example Validation Error:**
```
Configuration Validation Failed

Error: At least one provider must be enabled for the "Claude" tool.

Affected Providers:
- Anthropic (Disabled)
- Backup Provider (Disabled)

Fix: Enable at least one provider or add a new provider.
```

## Configuration Rollback

Managers can revert to a previous configuration if problems occur after an update.

### Rollback Scenarios

**Automatic Rollback:**
- Proxy fails to start after configuration update
- Client detects invalid configuration after pull

**Manual Rollback:**
- Manager identifies issues with current configuration
- Provider causing problems needs to be removed

### Rollback Process

**Manual Rollback Workflow:**

1. **View History**
   - Manager accesses configuration history
   - Shows past 10 configurations with timestamps
   - Each entry shows who made the change

2. **Select Previous Version**
   - Manager clicks "Rollback" on desired version
   - System shows preview of changes

3. **Confirm Rollback**
   - Manager confirms rollback
   - Previous configuration pushed to server
   - All members receive rollback within 5 minutes

### Configuration History

**Stored Information:**
- Configuration version (timestamp-based)
- Who made the change
- What changed (diff summary)
- Validation result

**Retention:**
- Last 10 configurations retained
- Older versions automatically deleted
- History cleared when organization deleted

### Rollback Limitations

**Constraints:**
- Cannot rollback if provider was deleted (API key not retained)
- Cannot rollback beyond history retention (10 versions)
- Rollback overwrites current configuration (not merged)

## Configuration Templates

Managers can use templates to quickly set up providers for common scenarios.

### Available Templates

| Template | Description | Use Case |
|----------|-------------|----------|
| **Anthropic Only** | Single Anthropic provider | Teams using only Claude |
| **Primary + Backup** | Two providers with failover | Teams wanting redundancy |
| **Cost Optimized** | Lower-cost provider prioritized | Budget-conscious teams |
| **Custom** | User-defined template | Teams with specific needs |

### Template Workflow

1. **Select Template**
   - Manager chooses from template list
   - Preview shows what will be configured

2. **Customize**
   - Manager fills in required fields (API keys, endpoints)
   - Can modify provider priorities
   - Can add or remove providers

3. **Apply**
   - Configuration validated
   - Pushed to server
   - Distributed to members

### Custom Templates

**Manager-Created Templates:**
- Managers can save current configuration as template
- Templates stored per organization
- Can be shared with other managers in same organization

**Template Use:**
- "Save as Template" button in configuration screen
- Template name and description required
- Templates exclude sensitive data (API keys)

---

**Related:** [03.02 Teams](../02_teams/) | [03.03 Offline Mode](../04_offline_mode/)
