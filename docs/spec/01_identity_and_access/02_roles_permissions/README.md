# Roles & Permissions

AI Together has two user roles with different permissions. This ensures users can access appropriate features based on their responsibilities.

## Purpose

Define what actions each user role can perform within the system.

## User Personas

- **Manager** - Team administrator who manages providers, users, and team settings
- **Member** - Team member who uses AI tools and views personal usage data

## User Roles

### Manager

**Who:** Team administrators, tech leads, engineering managers

**User Stories:**
- As a **manager**, I want to configure providers so that my team can use AI tools
- As a **manager**, I want to add team members so they can access our configuration
- As a **manager**, I want to deactivate users who leave our organization
- As a **manager**, I want to view team analytics so I can understand usage patterns
- As a **manager**, I want to push configuration changes so everyone gets updates immediately

**Can do:**
- Add and deactivate users
- Assign roles (manager/member)
- Configure AI providers (add, edit, delete, test)
- Set provider priority and enable/disable
- View team-wide analytics and usage
- Push configuration to team members
- Activate commercial license
- View all team members' usage data

**Cannot do:**
- Impersonate other users
- Modify historical usage data
- Access other organizations' data

### Member

**Who:** Individual developers, end users

**User Stories:**
- As a **member**, I want to view available providers so I know what's configured
- As a **member**, I want to pull latest configuration so I have provider updates
- As a **member**, I want to view my usage so I can track my personal consumption
- As a **member**, I want to export my data so I can analyze it further

**Can do:**
- Use configured AI providers
- View provider list and status
- View personal usage statistics
- View personal request history
- Pull configuration from server
- Export personal usage data

**Cannot do:**
- Add, edit, or delete providers
- Push configuration to server
- View other users' data
- Access team analytics dashboard
- Manage users or activate commercial license

## Permission Matrix

| Level | Action | Member | Manager |
|-------|--------|--------|---------|
| **1** | **Provider Management** | | |
| 1.1 | View provider list | ✅ | ✅ |
| 1.2 | View provider status | ✅ | ✅ |
| 1.3 | Add provider | ❌ | ✅ |
| 1.4 | Edit provider | ❌ | ✅ |
| 1.5 | Delete provider | ❌ | ✅ |
| 1.6 | Enable/disable provider | ❌ | ✅ |
| 1.7 | Set provider priority | ❌ | ✅ |
| 1.8 | Test provider connection | ❌ | ✅ |
| **2** | **User Management** | | |
| 2.1 | View user list | ❌ | ✅ |
| 2.2 | Add user | ❌ | ✅ |
| 2.3 | Deactivate user | ❌ | ✅ |
| 2.4 | Assign roles | ❌ | ✅ |
| **3** | **Analytics** | | |
| 3.1 | View personal usage | ✅ | ✅ |
| 3.2 | View team analytics | ❌ | ✅ |
| 3.3 | View other users' usage | ❌ | ✅ |
| 3.4 | Export usage data | ✅ | ✅ |
| **4** | **Configuration** | | |
| 4.1 | Pull config from server | ✅ | ✅ |
| 4.2 | Push config to server | ❌ | ✅ |
| **5** | **License** | | |
| 5.1 | View license status | ✅ | ✅ |
| 5.2 | Activate commercial license | ❌ | ✅ |

## Functional Requirements

- **FR-001:** System must check user permissions before every protected action
- **FR-002:** System must deny member access to manager-only features
- **FR-003:** Permission checks must happen at API level (server-side enforcement)
- **FR-004:** UI must hide/show features based on user role
- **FR-005:** Role changes must take effect immediately (no login required)
- **FR-006:** Last Manager cannot be demoted or deactivated

## Business Rules

- **BR-001:** Organization must have at least one Manager at all times
- **BR-002:** Only Managers can change user roles
- **BR-003:** Only Managers can deactivate users
- **BR-004:** Users can only see data for their own organization
- **BR-005:** Last Manager cannot be demoted or deactivated

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Member tries to access manager-only feature | Shows "Contact your manager" message |
| Last manager tries to demote self | System prevents action with error |
| Last manager tries to deactivate self | System prevents action with error |
| Manager tries to deactivate self as last manager | System prevents action with error |
| Member tries to deactivate another member | Button disabled or shows error |
| Member tries to push config | Button disabled or shows error |
| Role change while user has active sessions | New permissions apply on next action |

## Success Criteria

- Permission checks happen before every protected action
- Error messages clearly explain what's needed to access feature
- Role changes take effect immediately
- Members cannot accidentally perform manager actions

---

**Related:** [01.01 User Accounts](../01_user_accounts/) | [01.03 Multi-Tenancy](../03_multi_tenancy/) | [Domain 01 Overview](../README.md)
