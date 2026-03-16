# Data Collection

The system collects metadata about AI tool requests. No prompt or response content is ever stored.

## Purpose

Track usage for analytics and billing while respecting user privacy.

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

## Functional Requirements (EARS Format)

### 1. Privacy Protection

**Purpose:** Ensure user privacy by never storing content.

#### Ubiquitous Requirements

- **DC-04-001:** `The system shall never store prompt content from AI tool requests.`
- **DC-04-002:** `The system shall never store response content from AI tool replies.`
- **DC-04-003:** `The system shall never inspect or analyze prompt or response content.`
- **DC-04-004:** `The system shall only collect metadata about AI tool requests.`

---

### 2. Metadata Collection

**Purpose:** Define what metadata is collected for each request.

#### Event-Driven Requirements (Data Extraction)

- **DC-04-101:** `When an AI tool request completes, the system shall extract the request ID.`
- **DC-04-102:** `When an AI tool request completes, the system shall extract the user who made the request.`
- **DC-04-103:** `When an AI tool request completes, the system shall extract the platform (AI tool type).`
- **DC-04-104:** `When an AI tool request completes, the system shall extract the model name used.`
- **DC-04-105:** `When an AI tool request completes, the system shall extract the provider name used.`
- **DC-04-106:** `When an AI tool request completes, the system shall extract input token count.`
- **DC-04-107:** `When an AI tool request completes, the system shall extract output token count.`
- **DC-04-108:** `When an AI tool request completes, the system shall extract cache token count.`
- **DC-04-109:** `When an AI tool request completes, the system shall extract request duration.`
- **DC-04-110:** `When an AI tool request completes, the system shall extract HTTP status code.`
- **DC-04-111:** `When an AI tool request completes, the system shall extract timestamp of request.`

#### State-Driven Requirements (Data Storage)

- **DC-04-201:** `While metadata is extracted, the system shall store the data locally before upload to server.`
- **DC-04-202:** `While metadata extraction occurs, the system shall not access prompt or response content.`

---

### 3. Data Upload and Sync

**Purpose:** Handle batch upload of metadata to server.

#### Event-Driven Requirements (Batch Upload)

- **DC-04-301:** `When the batch upload interval elapses (30 seconds), the system shall upload pending metadata to the server.`
- **DC-04-302:** `When metadata is successfully uploaded to the server, the system shall delete the local copy.`
- **DC-04-303:** `When the server receives metadata, the system shall store the data per organization (tenant).`

#### State-Driven Requirements (Offline Behavior)

- **DC-04-401:** `While the system is offline, the system shall queue metadata locally for later upload.`
- **DC-04-402:** `While the system is offline, the system shall continue collecting and storing metadata locally.`

---

### 4. Data Access and Privacy

**Purpose:** Control who can access collected data.

#### Event-Driven Requirements (User Data Access)

- **DC-04-501:** `When a user requests to view their collected data, the system shall display all metadata stored about the user.`
- **DC-04-502:** `When a manager requests to view team usage data, the system shall display aggregated metadata for all users.`

#### Unwanted Behaviour Requirements (Access Restrictions)

- **DC-04-601:** `If a user attempts to view another user's data, then the system shall deny access.`
- **DC-04-602:** `If a member attempts to view team usage data, then the system shall deny access.`

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

### Storage Behavior

| Location | Behavior |
|----------|----------|
| **Local (Member Client)** | Transient buffer before upload to server. Deleted after successful sync. Queues data when offline. |
| **Server** | Metadata only (no content). Stored per organization (tenant). Retained per license type policy. |

## Business Rules

- **BR-04-001:** Metadata collection is automatic and cannot be disabled
- **BR-04-002:** Collected data belongs to the organization
- **BR-04-003:** Users can view their own data but not others' (except managers)
- **BR-04-004:** Data is retained according to license type (see [Data Retention](../05_data_retention/))

---

**Related:** [04.02 Personal Analytics](../02_personal_analytics/) | [04.03 Team Analytics](../03_team_analytics/) | [04.05 Data Retention](../05_data_retention/)
