# Privacy

Code Together is designed to be privacy-first. We never access or store prompt or response content from AI tools.

## Purpose

Protect user privacy by collecting only necessary metadata, ensuring that sensitive information never leaves the user's computer.

## Privacy Guarantee

**We Never Store:**
- Prompt content (what users ask AI tools)
- Response content (what AI tools reply)
- Code snippets
- File contents
- Any sensitive information from AI tool interactions

**We Only Store:**
- Metadata about requests (model name, token counts, timestamps)
- Provider information
- Usage statistics (aggregated)

## Functional Requirements (EARS Format)

### 1. Data Minimization

**Purpose:** Ensure only necessary metadata is collected.

#### Ubiquitous Requirements

- **PR-06-001:** `The system shall never store prompt content from AI tool requests.`
- **PR-06-002:** `The system shall never store response content from AI tool replies.`
- **PR-06-003:** `The system shall never inspect or analyze prompt or response content.`
- **PR-06-004:** `The system shall only collect metadata about AI tool requests.`
- **PR-06-005:** `The system shall never store code snippets or file contents from AI tool interactions.`

---

### 2. Data Storage Policies

**Purpose:** Define where and how metadata is stored.

#### State-Driven Requirements (Local Storage)

- **PR-06-101:** `While metadata is stored locally on the member client, the system shall use it as a transient buffer before upload to server.`
- **PR-06-102:** `While metadata is successfully uploaded to the server, the system shall delete the local copy.`
- **PR-06-103:** `While the system is offline, the system shall queue metadata locally for later upload.`

#### State-Driven Requirements (Server Storage)

- **PR-06-104:** `While metadata is stored on the server, the system shall store only metadata (no content).`
- **PR-06-105:** `While metadata is stored on the server, the system shall maintain complete isolation per organization (tenant).`
- **PR-06-106:** `While the retention period expires, the system shall automatically delete metadata according to license type policy.`

---

### 3. User Data Rights

**Purpose:** Implement user rights under GDPR and privacy policies.

#### Event-Driven Requirements (Data Access)

- **PR-06-201:** `When a user requests to view their collected data, the system shall display all metadata stored about the user.`
- **PR-06-202:** `When a user exports their personal usage data, the system shall generate a CSV/JSON file with all their metadata.`
- **PR-06-203:** `When a user requests account deletion, the system shall delete the user's account and all associated metadata.`

---

### 4. Manager Access Limitations

**Purpose:** Define what managers can access.

#### Ubiquitous Requirements

- **PR-06-301:** `The system shall allow managers to view aggregated team usage data.`
- **PR-06-302:** `The system shall not allow managers to access prompt or response content (because it is never stored).`
- **PR-06-303:** `The system shall allow managers to view individual user metadata for team management purposes.`

## Data Access Policies

### User Access Rights

Users have the following rights under our privacy policy:

| Right | Description | Implementation |
|-------|-------------|----------------|
| **Right to Access** | View all metadata collected about them | Users can view their personal analytics |
| **Right to Export** | Export personal usage data | Users can export data as CSV/JSON |
| **Right to Deletion** | Request account deletion with data removal | Users can request account deletion |

## Data Storage Policies

### Local Storage (Member Client)

| Aspect | Policy |
|--------|--------|
| **Purpose** | Transient buffer before upload to server |
| **Retention** | Deleted after successful sync |
| **Offline Handling** | Queued locally when offline, uploaded when reconnected |

### Server Storage

| Aspect | Policy |
|--------|--------|
| **Data Type** | Metadata only (no content) |
| **Isolation** | Stored per organization (tenant) with complete isolation |
| **Retention** | Retained per license type policy (7 days Open Source, 90 days Commercial) |
| **Deletion** | Automatically deleted after retention period expires |

## Related Documentation

- **Data Collection:** [`04.01 Data Collection`](../../04_usage_insights/01_data_collection/) - Detailed field specifications
- **Data Retention:** [`04.05 Data Retention`](../../04_usage_insights/05_data_retention/) - Retention policies by license type
- **Compliance:** [`06.04 Compliance`](../04_compliance/) - Regulatory compliance details

---

**Related:** [06.02 Security](../02_security/) | [06.04 Compliance](../04_compliance/)
