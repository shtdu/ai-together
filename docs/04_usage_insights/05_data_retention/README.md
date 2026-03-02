# Data Retention

Usage data is retained for different periods depending on license type. Older data is automatically deleted.

## Purpose

Balance storage costs with user needs for historical data, while providing appropriate retention for each license type.

## Retention Periods by License Type

| License Type | Retention Period | What It Means |
|--------------|------------------|---------------|
| **Open Source** | 7 days | Data older than 7 days is auto-deleted |
| **Commercial** | 90 days | Data older than 90 days is auto-deleted |

## What Gets Deleted

**Automatic Deletion:**
- Request records older than retention period
- Associated token counts and metadata
- Analytics aggregates for expired periods

**Never Deleted (unless account deleted):**
- User accounts
- Provider configurations
- Team settings
- License information

## Deletion Behavior

### Automatic Purge

**Timing:**
- Open Source: Daily purge of data older than 7 days
- Commercial: Weekly purge of data older than 90 days

**User Impact:**
- Old data disappears from analytics
- Historical charts show available period only
- Exports are only possible within retention period

## Functional Requirements

- **FR-001:** System must automatically delete data older than retention period
- **FR-002:** Deletion must happen on schedule (daily/weekly/monthly)
- **FR-003:** Users must be able to export data before deletion
- **FR-004:** Analytics must only show data within retention period

## Business Rules

- **BR-001:** Retention period is enforced at the database level
- **BR-002:** Deleted data cannot be recovered
- **BR-003:** Retention period is determined by license type at time of purge
- **BR-004:** Upgrading license extends retention going forward (does not restore deleted data)

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| License downgrade with older data | Keep data until original retention expires |
| License upgrade | New retention applies from upgrade date forward |
| Export fails during purge window | Retry purge in next cycle |
| User requests deleted data | Show error: "Data not available (retention period exceeded)" |

## Data Lifecycle

```
Request Created
    ↓
Stored in local database (member client)
    ↓
Uploaded to server (suggested: within 30 seconds)
    ↓
Stored in server database
    ↓
Available in analytics
    ↓
Retention period expires
    ↓
Data automatically deleted
```

## Compliance

### GDPR Support

Retention policies support GDPR requirements:
- **Right to be forgotten:** Data deleted after retention period
- **Data minimization:** Only kept as long as necessary
- **Purpose limitation:** Used only for analytics during retention

### Data Backup

**Note:** System does not automatically backup data before deletion. Users who need long-term retention should:
- Use Commercial license for extended retention
- Export data regularly for offline storage

## Success Criteria

- Data is deleted on schedule according to license type
- Analytics only show available data
- Export functionality works for data within retention period
- Deleted data cannot be recovered

---

**Related:** [04.01 Data Collection](../01_data_collection/) | [04.03 Team Analytics](../03_team_analytics/)
