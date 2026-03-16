# Multi-Tenancy

Each Code Together customer (organization) is a separate tenant with completely isolated data. Users from one organization never see or interact with data from another organization.

## Purpose

Ensure complete data isolation between organizations for security, compliance, and privacy.

## User Personas

- **Organization Leader** - Wants assurance their team's data is private and secure
- **Member** - Wants to know their usage data is only visible within their organization
- **Compliance Officer** - Needs assurance that data cannot leak between organizations

## Functional Requirements (EARS Format)

### 1. Data Isolation System

**Purpose:** Ensure complete separation of data between organizations.

#### Ubiquitous Requirements

- **MT-01-001:** `The system shall authenticate users to an organization before accessing any data.`
- **MT-01-002:** `The system shall scope all data queries to the user's organization.`
- **MT-01-003:** `The system shall ensure organization names are unique.`
- **MT-01-004:** `The system shall assign each organization a unique identifier.`
- **MT-01-005:** `The system shall enforce data isolation at all system layers.`

---

### 2. Authentication & Data Access

**Purpose:** Control which organization's data a user can access.

#### Event-Driven Requirements (Login)

- **MT-01-101:** `When a user from organization A logs in, the system shall display only organization A data.`
- **MT-01-102:** `When a user logs in, the system shall authenticate the user to the user's organization.`

#### Unwanted Behaviour Requirements (Cross-Organization Access)

- **MT-01-201:** `If a user from organization A attempts to access organization B data, then the system shall deny the access operation.`
- **MT-01-202:** `If a manager from organization A attempts to access organization B analytics, then the system shall deny the access operation.`
- **MT-01-203:** `If a user attempts to access another organization's data, then the system shall display an access denied error.`

#### Event-Driven Requirements (Organization Switching)

- **MT-01-103:** `When a user attempts to switch organizations, the system shall require separate authentication.`

---

### 3. Data Visibility

**Purpose:** Ensure users only see their organization's data.

#### State-Driven Requirements (Active User Session)

- **MT-01-401:** `While a user is logged in, the system shall only display the user's organization's users.`
- **MT-01-402:** `While a user is logged in, the system shall only display the user's organization's providers.`
- **MT-01-403:** `While a user is logged in, the system shall only display the user's organization's usage data.`
- **MT-01-404:** `While a user is logged in, the system shall only display the user's organization's analytics.`
- **MT-01-405:** `While a user is logged in, the system shall not display any indication of other organizations existing.`

---

### 4. User Invitation & Organization Boundaries

**Purpose:** Handle user invitations across organization boundaries.

#### Event-Driven Requirements (User Invitation)

- **MT-01-301:** `When a manager attempts to invite a user whose email exists in another organization, then the system shall display a message indicating the email already belongs to another organization.`

#### Complex Requirements (User Organization Transition)

- **MT-01-302:** `When a user leaves organization A and joins organization B, the system shall make organization A data inaccessible to the user.`
- **MT-01-303:** `When a user leaves organization A and joins organization B, the system shall make organization B data visible to the user.`

---

### 5. Data Portability & Export

**Purpose:** Allow users to export their organization's data while maintaining isolation.

#### Event-Driven Requirements (Data Export)

- **MT-01-501:** `When a user requests to export personal usage data, the system shall generate an export containing only the user's personal data.`
- **MT-01-502:** `When a manager requests to export organization-wide data, the system shall generate an export containing only the organization's data.`

#### Unwanted Behaviour Requirements (Cross-Organization Export Prevention)

- **MT-01-601:** `If a user attempts to export data from another organization, then the system shall deny the export operation.`
- **MT-01-602:** `If a member attempts to export organization-wide data, then the system shall deny the export operation.`

---

### 6. Organization Deletion

**Purpose:** Handle organization deletion with proper data lifecycle management.

#### Event-Driven Requirements (Deletion Workflow)

- **MT-01-503:** `When an organization is deleted, the system shall schedule all organization data for deletion after a 30-day grace period.`

#### Complex Requirements (Grace Period Behavior)

- **MT-01-701:** `When an organization is deleted, while the 30-day grace period is active, the system shall preserve the organization data for potential restoration.`
- **MT-01-702:** `When the 30-day grace period expires after organization deletion, the system shall permanently remove all organization data.`

## Business Rules

- **BR-03-001:** Each organization has exactly one unique identifier
- **BR-03-002:** Users belong to exactly one organization and belong to one team within their organization
- **BR-03-003:** Organization data cannot be merged or split
- **BR-03-004:** When an organization is deleted, all its data is deleted

## Data Ownership

### Who Owns What?

| Data Type | Owner | Accessible By |
|-----------|-------|---------------|
| User account | Organization | User, managers |
| Provider config | Organization | All members, managers |
| Usage data | Organization | User (personal), managers (all) |
| Team settings | Organization | All members, managers |

### Data Portability

**User Stories:**
- As a **user**, I want to export my personal usage data before leaving an organization
- As an **organization**, I want to export our data before closing our account

**Behavior:**
- Users can export their personal usage data at any time
- Managers can export organization-wide data
- Upon organization deletion, data is permanently removed after 30-day grace period

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| User tries to access another org's data | Access denied with error |
| Manager attempts to view another org's analytics | Access denied with error |
| Organization is deleted | All data scheduled for deletion after 30 days |
| Same email invited to different organization | System shows email already registered to another org |

## Security Implications

Multi-tenancy provides:
- **Compliance** - GDPR-ready data segregation
- **Privacy** - No cross-organization data leakage
- **Security** - Compromised credentials affect only one organization
- **Auditability** - Clear data ownership per organization

## Success Criteria

- Users never see data from other organizations
- All data queries are properly scoped
- Organization isolation is enforced at all system layers
- Data exports contain only the user's organization data

---

**Related:** [01.01 User Accounts](../01_user_accounts/) | [01.02 Roles & Permissions](../02_roles_permissions/) | [06.01 Privacy](../../06_system_behaviors/01_privacy/) | [Domain 01 Overview](../README.md)
