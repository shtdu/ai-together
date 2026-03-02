# Multi-Tenancy

Each Code Together customer (organization) is a separate tenant with completely isolated data. Users from one organization never see or interact with data from another organization.

## Purpose

Ensure complete data isolation between organizations for security, compliance, and privacy.

## User Personas

- **Organization Leader** - Wants assurance their team's data is private and secure
- **Member** - Wants to know their usage data is only visible within their organization
- **Compliance Officer** - Needs assurance that data cannot leak between organizations

## User Stories

- As an **organization leader**, I want to be sure my team's data is private and secure
- As a **user**, I want to know that my usage data is only visible to my organization
- As a **compliance officer**, I want assurance that data cannot leak between organizations

## What Is Isolated

Each organization (tenant) has completely separate:
- User accounts
- Provider configurations
- Usage data and analytics
- Teams and team settings
- License information

## User Experience

### Organizational Boundaries

Users only ever see:
- Their own organization's users
- Their own organization's providers
- Their own organization's usage data
- Their own organization's analytics

Users never see:
- Other organizations' names
- Other organizations' users
- Other organizations' data
- Any indication of other organizations existing

### Cross-Organization Behavior

| Action | Behavior |
|--------|----------|
| User from Org A logs in | Only sees Org A data |
| Manager from Org A tries to access Org B | Access denied |
| Manager attempts to invite existing user from another org | System shows email already belongs to another organization |
| User leaves Org A, joins Org B | Org A data inaccessible, Org B data visible |

## Functional Requirements

- **FR-001:** Users must be authenticated to an organization before accessing any data
- **FR-002:** All data queries must be scoped to the user's organization
- **FR-003:** Users cannot switch organizations without separate authentication
- **FR-004:** Organization names must be unique (for identification purposes)

## Business Rules

- **BR-001:** Each organization has exactly one unique identifier
- **BR-002:** Users belong to exactly one organization and belong to one team within their organization
- **BR-003:** Organization data cannot be merged or split
- **BR-004:** When an organization is deleted, all its data is deleted

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
