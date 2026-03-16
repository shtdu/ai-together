# Offline Mode

> **DEFERRED** - This feature is not currently implemented. The system currently requires an active server connection. Offline support with local caching and queued sync is planned for a future release.

## Purpose

Ensure developers can continue their work even when the team server is unavailable.

## Current Implementation

**Online-Only Mode:**
- System requires active server connection to function
- If server is unreachable, the member app cannot operate
- No local caching of configuration
- No offline usage tracking

## Planned Features (Future)

### User Stories

- As a **member**, I want to keep working if the server is down
- As a **member**, I want my usage data saved even when offline
- As a **member**, I want to know when I'm in offline mode
- As a **member**, I want my data to sync when connection is restored

### What Works Offline

#### Fully Functional
- AI tool routing (uses cached provider configuration)
- Usage tracking (stored locally)
- Provider failover (works as configured)
- Personal usage statistics (view from local cache)

#### Not Available Offline
- Pulling latest configuration from server
- Pushing configuration changes to server
- Uploading usage data to server
- Team analytics (requires server data)
- Adding/managing users (requires server)

### Offline Indicators

Users see clear offline indicators:

| Location | Indicator |
|----------|-----------|
| System tray | Gray/offline icon |
| Main window | Banner: "Offline - using cached configuration" |
| Server view | Status: "Disconnected" |
| Sync status | "Last sync: [time] ago" |

### Offline Behavior

#### Entering Offline Mode

**Triggers:**
- Network connection lost
- Server unreachable
- Server maintenance

**System Response:**
- Detect offline state within 30 seconds
- Show offline indicator
- Continue using cached configuration
- Queue usage data locally

#### While Offline

**Configuration:**
- Last synced configuration is used
- Configuration changes are discarded (server-wins policy)

**Usage Data:**
- All usage tracked locally
- Stored in local database
- Marked as "not synced"

**Sync Behavior:**
- Auto-sync attempts continue every 5 minutes
- Failed attempts are logged (not shown to user)
- No error messages unless user tries manual sync

#### Returning Online

**Automatic Recovery:**
1. Connection restored
2. Next auto-sync attempt succeeds
3. Usage data uploaded (up to 100 records per batch)
4. Local records marked as synced
5. Pull latest server configuration
6. Status changes to "Connected"

### Manual Sync While Offline

**User Story:** As a member, I want to manually try to sync when offline.

**Workflow:**
1. User clicks "Sync Now"
2. System attempts connection to server
3. If still offline: show "Server unavailable. Retrying..."
4. User can retry or wait for auto-recovery

### Data Safety

#### While Offline

**Safe:**
- Cached configuration is used
- All usage data is recorded locally
- No data is lost

**Considerations:**
- Configuration may be stale
- Storage limited by local disk space

#### When Returning Online

**Automatic Actions:**
- Upload pending usage data (up to 500 records)
- Pull latest server configuration (server-wins)

### Functional Requirements (Future)

- **FR-001:** System must detect offline state within 30 seconds
- **FR-002:** System must use cached configuration when offline
- **FR-003:** System must queue usage data locally when offline
- **FR-004:** System must automatically sync when connection restored
- **FR-005:** System must show clear offline indicators

### Business Rules (Future)

- **BR-001:** Cached configuration is used when offline
- **BR-002:** Usage data never lost when offline
- **BR-003:** Auto-sync resumes when connection restored
- **BR-004:** Pending uploads complete within 1 minute of reconnection
- **BR-005:** Server configuration always wins when returning online

### Edge Cases & Error Handling (Future)

| Scenario | System Behavior |
|----------|----------------|
| Long-term offline (weeks) | Continue working, usage accumulates locally |
| Local disk full while offline | Show error, pause usage tracking |
| Server data deleted while offline | Next sync pulls latest config, keeps local usage data |

### Success Criteria (Future)

- Offline mode activates within 30 seconds of server loss
- Users can continue working with cached configuration
- All usage data is preserved
- Automatic recovery when connection restored
- Clear indicators show online/offline state

---

**Related:** [03.01 Distribution](../01_distribution/) | [03.02 Teams](../02_teams/)
