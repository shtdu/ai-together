# Identity & Access

This domain covers **who can use the system** and **what they're allowed to do**. It includes user accounts, roles, permissions, multi-tenant isolation, and license-based feature limits.

## Purpose

Enable secure access control, user management, and feature differentiation based on roles and license types.

## User Personas

- **Manager** - Team administrator who manages users, providers, and settings
- **Member** - Team member who uses AI tools and views personal data
- **Organization Creator** - First user who sets up a new organization

## Overview

AI Together supports two types of users with different permissions. Organizations (tenants) are completely isolated from each other. License types determine which features are available.

## Subdomains

| ID | Subdomain | Description |
|----|-----------|-------------|
| **01.01** | [User Accounts](01_user_accounts/) | Creating accounts, logging in, session management, user deactivation |
| **01.02** | [Roles & Permissions](02_roles_permissions/) | What managers vs members can do |
| **01.03** | [Multi-Tenancy](03_multi_tenancy/) | How organizations are isolated |
| **01.04** | [Licensing](04_licensing/) | Feature limits by license tier |

## Functional Requirements (EARS Format)

### 1. User Role System

**Purpose:** Define the two fundamental user roles with distinct permission levels.

#### Ubiquitous Requirements

- **IA-01-001:** `The system shall support two user roles: Manager and Member.`
- **IA-01-002:** `The system shall enforce role-based permissions for all operations.`
- **IA-01-003:** `The system shall associate each user with exactly one organization.`

---

### 2. Manager Capabilities

**Purpose:** Define what managers can do within the system.

#### Event-Driven Requirements (User Management)

- **IA-01-101:** `When a manager requests to add a user, the system shall allow the user creation operation.`
- **IA-01-102:** `When a manager requests to deactivate a user, the system shall allow the user deactivation operation.`
- **IA-01-103:** `When a manager requests to assign a role, the system shall allow the role assignment operation.`

#### Event-Driven Requirements (Provider Management)

- **IA-01-104:** `When a manager requests to configure AI providers, the system shall allow the provider configuration operation.`

#### Event-Driven Requirements (Analytics & Configuration)

- **IA-01-105:** `When a manager requests to view team-wide analytics, the system shall display the team analytics data.`
- **IA-01-106:** `When a manager requests to manage the license, the system shall allow the license management operation.`
- **IA-01-107:** `When a manager requests to push configuration to team members, the system shall execute the configuration push operation.`

#### Unwanted Behaviour Requirements (Manager Restrictions)

- **IA-01-201:** `If a manager attempts to impersonate another user, then the system shall deny the impersonation operation.`
- **IA-01-202:** `If a manager attempts to modify historical usage data, then the system shall deny the modification operation.`
- **IA-01-203:** `If a manager attempts to access another tenant's data, then the system shall deny the access operation.`

---

### 3. Member Capabilities

**Purpose:** Define what members can do within the system.

#### Event-Driven Requirements (Provider Usage)

- **IA-01-301:** `When a member requests to use configured AI providers, the system shall allow the provider usage operation.`

#### Event-Driven Requirements (Personal Data Access)

- **IA-01-302:** `When a member requests to view personal usage statistics, the system shall display the personal usage data.`
- **IA-01-303:** `When a member requests to view personal request history, the system shall display the personal request history data.`
- **IA-01-304:** `When a member requests to export personal usage data, the system shall generate the personal usage data export.`

#### Event-Driven Requirements (Configuration Sync)

- **IA-01-305:** `When a member requests to pull configuration from server, the system shall execute the configuration pull operation.`

#### Unwanted Behaviour Requirements (Member Restrictions)

- **IA-01-401:** `If a member attempts to modify team provider settings, then the system shall deny the modification operation.`
- **IA-01-402:** `If a member attempts to view other users' data, then the system shall deny the access operation.`
- **IA-01-403:** `If a member attempts to push configuration to server, then the system shall deny the push operation.`
- **IA-01-404:** `If a member attempts to access team analytics, then the system shall deny the access operation.`

---

### 4. Common Capabilities

**Purpose:** Define capabilities available to all authenticated users regardless of role.

#### State-Driven Requirements (Authenticated User Access)

- **IA-01-501:** `While a user is logged in with a valid role, the system shall allow the user to use AI providers.`
- **IA-01-502:** `While a user is logged in with a valid role, the system shall allow the user to view personal usage.`
- **IA-01-503:** `While a user is logged in with a valid role, the system shall allow the user to pull configuration from server.`

---

### 5. Multi-Tenant Isolation

**Purpose:** Ensure complete data separation between organizations.

#### Ubiquitous Requirements

- **IA-01-002:** `The system shall isolate each organization's data from other organizations.`

## Permission Matrix

| Capability | Member | Manager |
|------------|--------|---------|
| Use AI providers | ✅ | ✅ |
| View personal usage | ✅ | ✅ |
| View team analytics | ❌ | ✅ |
| Configure providers | ❌ | ✅ |
| Add/deactivate users | ❌ | ✅ |
| Assign roles | ❌ | ✅ |
| Push configuration | ❌ | ✅ |
| Pull configuration | ✅ | ✅ |

## Business Rules

- **BR-01-001:** Each user must belong to exactly one organization
- **BR-01-002:** Each organization must have at least one Manager
- **BR-01-003:** User permissions are determined by their role (Manager or Member)
- **BR-01-004:** Organization data is completely isolated between tenants

## Related Documentation

- **User Accounts:** [`01.01 User Accounts`](01_user_accounts/) - Creating accounts and logging in
- **Roles & Permissions:** [`01.02 Roles & Permissions`](02_roles_permissions/) - Detailed permission matrix
- **Multi-Tenancy:** [`01.03 Multi-Tenancy`](03_multi_tenancy/) - Organization isolation
- **Licensing:** [`01.04 Licensing`](04_licensing/) - License tiers and feature limits

---

**Next:** [02 Provider Management](../02_provider_management/)
