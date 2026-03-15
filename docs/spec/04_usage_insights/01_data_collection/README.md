# Data Collection

The system collects metadata about AI tool requests. No prompt or response content is ever stored.

## Purpose

Track usage for analytics and billing while respecting user privacy.

> **For privacy policy, compliance, and storage policies,** see [Privacy](../../06_system_behaviors/01_privacy/).

## Privacy Guarantee

**We never store:**
- Prompt content (what users ask AI)
- Response content (what AI replies)
- Code snippets or user data
- Sensitive information

**We only store:**
- Model name used
- Token counts (input, output, cache)
- Provider name
- Request duration
- HTTP status code
- Timestamp

## What Is Tracked

### Per Request

Every AI tool request generates a record with:

| Field | Description | Example |
|-------|-------------|---------|
| Request ID | Unique identifier | `req_abc123` |
| User | Who made the request | `user@example.com` |
| Platform | Which AI tool | `claude`, `codex`, `opencode` |
| Model | Which model was used | `claude-3-5-sonnet` |
| Provider | Which provider was used | `Anthropic` |
| Input Tokens | Tokens in the request | 1000 |
| Output Tokens | Tokens in the response | 500 |
| Cache Tokens | Cached read tokens | 100 |
| Duration | How long the request took | 1500ms |
| Status | HTTP status code | 200 |
| Timestamp | When the request occurred | 2025-02-18 10:00:00 |

## Data Flow

```
1. User's AI tool makes request
2. Request passes through Code Together proxy
3. Proxy extracts metadata (no content inspection)
4. Response returns to user
5. Proxy extracts token counts from response
6. Metadata stored locally
7. Metadata uploaded to server (batch, suggested interval: 30 seconds)
```

### Storage Behavior

| Location | Behavior |
|----------|----------|
| **Local (Member Client)** | Transient buffer before upload to server. Deleted after successful sync. Queues data when offline. |
| **Server** | Metadata only (no content). Stored per organization (tenant). Retained per license type policy. |

## Functional Requirements

- **FR-001:** System must never store prompt or response content
- **FR-002:** System must extract metadata without inspecting prompt or response content
- **FR-003:** Users must be able to view all data collected about them
- **FR-004:** Data upload must happen in batches (every 30 seconds)
- **FR-005:** Local storage must queue data when offline

## Business Rules

- **BR-001:** Metadata collection is automatic and cannot be disabled
- **BR-002:** Collected data belongs to the organization
- **BR-003:** Users can view their own data but not others' (except managers)
- **BR-004:** Data is retained according to license type (see [Data Retention](../05_data_retention/))

## Data Visibility

| Role | Visibility |
|------|------------|
| **Members** | See only their own usage data |
| **Managers** | Can see all team members' usage data |
| **All Users** | Never see other organizations' data |

> **For data access policies and user rights under privacy law,** see [Privacy](../../06_system_behaviors/01_privacy/#data-access-policies).

## Data Export

Users can export their collected data for external analysis or record-keeping.

### Export Functionality

**Available to:** All users (Members export personal data, Managers export organization data)

### Export Specifications

| Specification | Value |
|---------------|-------|
| **Supported Formats** | CSV, JSON |
| **Maximum Size** | 100,000 records per export |
| **Generation Time** | Within 60 seconds for typical exports |
| **Timeout** | 5 minutes maximum generation time |
| **Availability Window** | Download link valid for 24 hours |
| **Date Range** | Maximum 365 days per export |

### Export Process

**Workflow:**
1. User selects export format and date range
2. System validates access permissions
3. System generates export file
4. Download link provided in UI and emailed to user
5. Link expires after 24 hours
6. Export action logged in audit trail

### Export Contents

**Personal Export (Members):**
- All personal usage records within date range
- Fields: timestamp, platform, model, provider, tokens, duration, status

**Organization Export (Managers):**
- All organization usage records within date range
- Fields: user, timestamp, platform, model, provider, tokens, duration, status
- Aggregated statistics by user, model, provider

### Data Format Examples

**CSV Format:**
```csv
timestamp,user,platform,model,provider,input_tokens,output_tokens,cache_tokens,duration_ms,status
2025-02-19T10:00:00Z,user@example.com,claude,claude-3-5-sonnet,Anthropic,1000,500,100,1500,200
```

**JSON Format:**
```json
{
  "export_type": "personal_usage",
  "organization": "org_abc123",
  "date_range": {
    "start": "2025-02-01T00:00:00Z",
    "end": "2025-02-19T23:59:59Z"
  },
  "records": [
    {
      "timestamp": "2025-02-19T10:00:00Z",
      "platform": "claude",
      "model": "claude-3-5-sonnet",
      "provider": "Anthropic",
      "input_tokens": 1000,
      "output_tokens": 500,
      "cache_tokens": 100,
      "duration_ms": 1500,
      "status": 200
    }
  ]
}
```

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Token counts not available | Use estimate, mark as approximate |
| Server unreachable during upload | Queue locally, retry when online |
| Export exceeds 100,000 records | Show error: "Date range too large. Please select a shorter range." |
| Export generation times out (>5 minutes) | Show error: "Export too large. Try a shorter date range." |
| Export download link expired | Show option to regenerate export |

## Success Criteria

- No prompt/response content is ever stored
- Metadata is accurate and complete
- Data uploads happen reliably
- Users can access their own data
- Privacy guarantee is maintained

---

**Related:** [04.02 Personal Analytics](../02_personal_analytics/) | [04.05 Data Retention](../05_data_retention/)
