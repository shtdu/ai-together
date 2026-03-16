# Roles & Permissions

Code Together has two user roles with different permissions. This ensures users can access appropriate features based on their responsibilities.

## Purpose

Define what actions each user role can perform within the system.

## User Personas

- **Manager** - Team administrator who manages providers, users, and team settings
- **Member** - Team member who uses AI tools and views personal usage data

## Functional Requirements (EARS Format)

### 1. Permission Enforcement System

**Purpose:** Ensure all operations are properly authorized based on user roles.

#### Ubiquitous Requirements

- **RP-01-001:** `The system shall check user permissions before every protected action.`
- **RP-01-002:** `The system shall enforce permission checks at the API level (server-side).`
- **RP-01-003:** `The system shall apply role changes immediately without requiring re-login.`
- **RP-01-004:** `The system shall hide features from the UI based on user role.`
- **RP-01-005:** `The system shall ensure each organization has at least one Manager at all times.`

---

### 2. Provider Management

**Purpose:** Control who can manage AI service provider configurations.

#### Event-Driven Requirements (Manager Capabilities)

- **RP-01-101:** `When a manager requests to view the provider list, the system shall display the provider list.`
- **RP-01-102:** `When a manager requests to view provider status, the system shall display the provider status.`
- **RP-01-103:** `When a manager requests to add a provider, the system shall allow the provider addition operation.`
- **RP-01-104:** `When a manager requests to edit a provider, the system shall allow the provider edit operation.`
- **RP-01-105:** `When a manager requests to delete a provider, the system shall allow the provider deletion operation.`
- **RP-01-106:** `When a manager requests to enable or disable a provider, the system shall allow the provider state change operation.`
- **RP-01-107:** `When a manager requests to set provider priority, the system shall allow the priority setting operation.`
- **RP-01-108:** `When a manager requests to test a provider connection, the system shall execute the connection test.`

#### Event-Driven Requirements (Member Read-Only Access)

- **RP-01-109:** `When a member requests to view the provider list, the system shall display the provider list.`
- **RP-01-110:** `When a member requests to view provider status, the system shall display the provider status.`

#### Unwanted Behaviour Requirements (Member Restrictions)

- **RP-01-201:** `If a member attempts to add a provider, then the system shall deny the provider addition operation.`
- **RP-01-202:** `If a member attempts to edit a provider, then the system shall deny the provider edit operation.`
- **RP-01-203:** `If a member attempts to delete a provider, then the system shall deny the provider deletion operation.`
- **RP-01-204:** `If a member attempts to enable or disable a provider, then the system shall deny the provider state change operation.`
- **RP-01-205:** `If a member attempts to set provider priority, then the system shall deny the priority setting operation.`
- **RP-01-206:** `If a member attempts to test a provider connection, then the system shall deny the connection test operation.`

---

### 3. User Management

**Purpose:** Control who can manage user accounts and roles.

#### Event-Driven Requirements (Manager Capabilities)

- **RP-01-301:** `When a manager requests to view the user list, the system shall display the user list.`
- **RP-01-302:** `When a manager requests to add a user, the system shall allow the user addition operation.`
- **RP-01-303:** `When a manager requests to deactivate a user, the system shall allow the user deactivation operation.`
- **RP-01-304:** `When a manager requests to assign a role, the system shall allow the role assignment operation.`

#### Unwanted Behaviour Requirements (Member Restrictions)

- **RP-01-401:** `If a member attempts to view the user list, then the system shall deny the user list access.`
- **RP-01-402:** `If a member attempts to add a user, then the system shall deny the user addition operation.`
- **RP-01-403:** `If a member attempts to deactivate a user, then the system shall deny the user deactivation operation.`
- **RP-01-404:** `If a member attempts to assign a role, then the system shall deny the role assignment operation.`

---

### 4. Analytics & Reporting

**Purpose:** Control access to usage analytics and reporting features.

#### Event-Driven Requirements (Personal Analytics - All Users)

- **RP-01-501:** `When a user requests to view personal usage statistics, the system shall display the personal usage data.`
- **RP-01-502:** `When a user requests to export usage data, the system shall generate the usage data export.`

#### Event-Driven Requirements (Team Analytics - Manager Only)

- **RP-01-503:** `When a manager requests to view team analytics, the system shall display the team analytics data.`
- **RP-01-504:** `When a manager requests to view other users' usage data, the system shall display the requested usage data.`

#### Unwanted Behaviour Requirements (Member Restrictions)

- **RP-01-601:** `If a member attempts to view team analytics, then the system shall deny the team analytics access.`
- **RP-01-602:** `If a member attempts to view other users' usage data, then the system shall deny the access to other users' data.`

---

### 5. Configuration Management

**Purpose:** Control who can push and pull configuration changes.

#### Event-Driven Requirements (Pull Configuration - All Users)

- **RP-01-701:** `When a user requests to pull configuration from the server, the system shall execute the configuration pull operation.`

#### Event-Driven Requirements (Push Configuration - Manager Only)

- **RP-01-702:** `When a manager requests to push configuration to the server, the system shall execute the configuration push operation.`

#### Unwanted Behaviour Requirements (Member Restrictions)

- **RP-01-801:** `If a member attempts to push configuration to the server, then the system shall deny the configuration push operation.`

---

### 6. License Management

**Purpose:** Control who can manage license activation and viewing.

#### Event-Driven Requirements (View License - All Users)

- **RP-01-901:** `When a user requests to view license status, the system shall display the license status.`

#### Event-Driven Requirements (Activate License - Manager Only)

- **RP-01-902:** `When a manager requests to activate a commercial license, the system shall allow the license activation operation.`

#### Unwanted Behaviour Requirements (Member Restrictions)

- **RP-01-1001:** `If a member attempts to activate a commercial license, then the system shall deny the license activation operation.`

---

### 7. Role Change Behavior

**Purpose:** Define how permission changes take effect.

#### Complex Requirements (Role Change Propagation)

- **RP-01-1101:** `When a user's role is changed while the user has active sessions, the system shall apply the new permissions on the next action.`

---

### 8. Manager Self-Action Restrictions

**Purpose:** Prevent managers from accidentally locking themselves out.

#### Unwanted Behaviour Requirements (Last Manager Protection)

- **RP-01-1201:** `If the last manager attempts to demote themselves to Member, then the system shall deny the role change operation.`
- **RP-01-1202:** `If the last manager attempts to deactivate themselves, then the system shall deny the deactivation operation.`

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

## Business Rules

- **BR-02-001:** Organization must have at least one Manager at all times
- **BR-02-002:** Only Managers can change user roles
- **BR-02-003:** Only Managers can deactivate users
- **BR-02-004:** Users can only see data for their own organization
- **BR-02-005:** Last Manager cannot be demoted or deactivated

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
