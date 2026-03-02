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

> **For detailed field specifications and technical implementation,** see [Data Collection](../../04_usage_insights/01_data_collection/).

## Why This Approach?

### Protecting User Data

AI prompts and responses often contain:
- Proprietary code
- Trade secrets
- Confidential information
- Sensitive business logic

Storing this data would create:
- **Security risks** - Data breaches could expose customer intellectual property
- **Legal risks** - Compliance violations with data protection regulations
- **Privacy violations** - Erosion of user trust

### Compliance

Our approach supports:
- **GDPR** - Data minimization and privacy by design principles
- **SOC 2** - Security principles (future certification planned)
- **Customer policies** - No sensitive data leaves user's environment without explicit consent

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
| **Retention** | Retained per license type policy (see [Data Retention](../../04_usage_insights/05_data_retention/)) |
| **Deletion** | Automatically deleted after retention period expires |

## Data Access Policies

### User Access Rights

Users have the following rights under our privacy policy:

| Right | Description | Implementation |
|-------|-------------|----------------|
| **Right to Access** | View all metadata collected about them | See [Data Collection - Export](../../04_usage_insights/01_data_collection/#data-export) |
| **Right to Export** | Export personal usage data | See [Data Collection - Export](../../04_usage_insights/01_data_collection/#data-export) |
| **Right to Deletion** | Request account deletion with data removal | See [Data Deletion](#data-deletion) below |

### Manager Access

**Managers Can:**
- View team usage metadata
- See per-user usage statistics
- Export organization-wide usage data

**Managers Cannot:**
- See prompt/response content (because we don't store it)
- Access data from other organizations
- View data outside their tenant isolation

### Data Isolation

- Each organization (tenant) is completely isolated
- Users see only their organization's data
- No cross-tenant data access is technically possible

## Data Sharing

**We Never:**
- Sell user data to third parties
- Share data with advertisers
- Share data between organizations
- Use data for purposes other than analytics and billing

## Data Deletion

### Upon Request

Users can request deletion of their data:

| Right | Policy |
|-------|--------|
| **GDPR "Right to be Forgotten"** | Personal usage data deleted within 30 days |
| **Account Deletion** | Account and identity data deleted per organization policy |
| **Compliance** | Deletion complies with GDPR requirements |

### Automatic Deletion

- Data automatically deleted after retention period expires
- See [Data Retention](../../04_usage_insights/05_data_retention/) for details per license type

## Certifications (Planned)

> **DEFERRED** - Certifications are planned for a future release.

We plan to pursue:
- **SOC 2 Type II** - Security and privacy controls
- **GDPR Compliance** - European data protection
- **ISO 27001** - Information security management

## Privacy Policy

Our full privacy policy will be available at: `https://code-together.com/privacy` (placeholder)

## Questions?

For privacy questions or concerns:
- Email: privacy@code-together.com
- Documentation: See [Data Collection](../../04_usage_insights/01_data_collection/) for implementation details

---

**Related:** [06.02 Security](../02_security/) | [06.04 Compliance](../04_compliance/)
