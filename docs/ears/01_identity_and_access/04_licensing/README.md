# Licensing

Code Together uses a 2-type license system: Open Source (free) and Commercial (paid). The system gracefully degrades when no license is active.

## Purpose

Support open source usage while enabling sustainable commercial development through feature differentiation.

## User Personas

- **Open Source User** - Individual or team using the software for open source projects
- **Commercial User** - Organization using the software for commercial purposes
- **Manager** - Team administrator who activates licenses and manages the subscription

## License Types Overview

| Type | Price | Seats | Providers | Teams | Data Retention |
|------|-------|-------|-----------|-------|----------------|
| **Open Source** | Free | Unlimited | Unlimited | 1 | 7 days |
| **Commercial** | Paid | Unlimited | Unlimited | Unlimited | 90 days |

## Functional Requirements (EARS Format)

### 1. License System Foundation

**Purpose:** Define the core license types and system behavior.

#### Ubiquitous Requirements

- **LC-01-001:** `The system shall support two license types: Open Source and Commercial.`
- **LC-01-002:** `The system shall function with Open Source features when no license is activated.`
- **LC-01-003:** `The system shall display the current license type and status to all users.`
- **LC-01-004:** `The system shall support unlimited team members for both license types.`
- **LC-01-005:** `The system shall support unlimited providers per kind for both license types.`

---

### 2. License Activation

**Purpose:** Handle license activation workflow and validation.

#### Event-Driven Requirements (License Activation Workflow)

- **LC-01-101:** `When a manager enters a commercial license key, the system shall validate the license signature.`
- **LC-01-102:** `When a manager enters a valid commercial license key, the system shall activate the commercial license.`
- **LC-01-103:** `When a commercial license is activated, the system shall unlock commercial features immediately.`
- **LC-01-104:** `When a manager activates a commercial license, the system shall update the license status for all users.`

#### Unwanted Behaviour Requirements (Activation Errors)

- **LC-01-201:** `If a manager enters an invalid license key, then the system shall display an error message indicating the license key is invalid.`
- **LC-01-202:** `If a manager attempts to activate an expired license key, then the system shall display an error message indicating the license has expired.`

---

### 3. Active Commercial License Features

**Purpose:** Define features available when commercial license is active.

#### State-Driven Requirements (Commercial Features)

- **LC-01-301:** `While a commercial license is active, the system shall allow unlimited team creation.`
- **LC-01-302:** `While a commercial license is active, the system shall allow access to the team analytics dashboard.`
- **LC-01-303:** `While a commercial license is active, the system shall allow usage data export.`
- **LC-01-304:** `While a commercial license is active, the system shall retain usage data for 90 days.`
- **LC-01-305:** `While a commercial license is active, the system shall enable cost calculation features.`
- **LC-01-306:** `While a commercial license is active, the system shall enable advanced analytics with filtering.`
- **LC-01-307:** `While a commercial license is active, the system shall enable request history with pagination.`
- **LC-01-308:** `While a commercial license is active, the system shall enable cost breakdown by provider, model, and user.`

---

### 4. Open Source License Features

**Purpose:** Define features available with Open Source license (default).

#### State-Driven Requirements (Open Source Features)

- **LC-01-401:** `While an Open Source license is active, the system shall allow basic provider configuration.`
- **LC-01-402:** `While an Open Source license is active, the system shall allow personal usage tracking.`
- **LC-01-403:** `While an Open Source license is active, the system shall retain usage data for 7 days.`
- **LC-01-404:** `While an Open Source license is active, the system shall allow unlimited providers per kind.`
- **LC-01-405:** `While an Open Source license is active, the system shall allow unlimited team members.`
- **LC-01-406:** `While an Open Source license is active, the system shall limit team creation to 1 team.`

#### Unwanted Behaviour Requirements (Open Source Feature Restrictions)

- **LC-01-501:** `If a user with an Open Source license attempts to access the team analytics dashboard, then the system shall deny the access and display a message indicating a commercial license is required.`
- **LC-01-502:** `If a user with an Open Source license attempts to export usage data, then the system shall deny the export operation and display a message indicating a commercial license is required.`
- **LC-01-503:** `If a manager with an Open Source license attempts to create a second team, then the system shall deny the team creation and display an error message indicating the team limit has been reached.`

---

### 5. Expired Commercial License

**Purpose:** Define system behavior when commercial license expires.

#### State-Driven Requirements (Expiration Behavior)

- **LC-01-601:** `While a commercial license is expired, the system shall allow existing users to log in.`
- **LC-01-602:** `While a commercial license is expired, the system shall preserve access to existing configurations.`
- **LC-01-603:** `While a commercial license is expired, the system shall preserve access to analytics and historical data.`
- **LC-01-604:** `While a commercial license is expired, the system shall block new team creation (reverts to 1 team limit).`
- **LC-01-605:** `While a commercial license is expired, the system shall disable advanced analytics features.`
- **LC-01-606:** `While a commercial license is expired, the system shall display a renewal prompt to managers.`

#### Complex Requirements (License Expiration During Operation)

- **LC-01-701:** `When a commercial license expires while the system is running, the system shall lock advanced features on the next operation while preserving core functionality.`

---

### 6. Feature Access Control

**Purpose:** Control access to features based on license type.

#### Event-Driven Requirements (Feature Access Requests)

- **LC-01-801:** `When a user requests access to a commercial feature with an active commercial license, the system shall allow the feature access.`
- **LC-01-802:** `When a user requests access to a commercial feature without a commercial license, the system shall display a message indicating the feature requires a commercial license.`
- **LC-01-803:** `When a user requests access to a core feature (available in both license types), the system shall allow the feature access regardless of license type.`

## Feature Availability

### Core Features (Both Types)

- AI provider management (Claude, Codex, OpenCode)
- Transparent request routing with failover
- Local configuration management
- Personal usage statistics
- Desktop applications (Windows, macOS, Linux)
- Unlimited providers per AI tool kind
- Unlimited team members

### Open Source (Free)

**Features:**
- ✅ Basic provider configuration
- ✅ Personal usage tracking
- ✅ 7-day usage data retention
- ✅ Unlimited providers per kind
- ✅ Unlimited team members
- ❌ Team analytics dashboard
- ❌ Cost calculation
- ❌ Usage data export
- ❌ Multi-team support

**Use Case:** Open source projects, individuals, and non-commercial use.

### Commercial (Paid)

**Features:**
- ✅ All Open Source features
- ✅ Team analytics dashboard
- ✅ Provider performance metrics
- ✅ User rankings and activity
- ✅ Cost calculation
- ✅ 90-day data retention
- ✅ Usage data export
- ✅ Unlimited teams
- ✅ Advanced analytics with filtering
- ✅ Request history with pagination
- ✅ Cost breakdown by provider/model/user

**Use Case:** Commercial organizations requiring team collaboration, analytics, and extended data retention.

## Business Rules

- **BR-04-001:** Graceful degradation - system always functions, features may lock
- **BR-04-002:** Expired commercial licenses preserve data access but disable advanced features
- **BR-04-003:** No seat limits - both license types support unlimited team members
- **BR-04-004:** No provider limits - both license types support unlimited providers per kind
- **BR-04-005:** License activation takes effect immediately

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Trying to create team beyond limit | Error: "Team limit reached. Commercial license required for multiple teams." |
| Commercial license expires while system running | Advanced features lock on next operation, core functionality preserved |
| Manager tries to activate invalid key | Error: "Invalid license key" |
| Accessing commercial feature without license | Message: "This feature requires a commercial license" |

## Success Criteria

- License activation completes within 10 seconds
- Team limits enforced before team creation
- License type and status accurately displayed
- Graceful degradation maintains core functionality
- Expired licenses don't lose data
- Open Source usage requires no license activation

## Pricing Summary

| Type | Price (Monthly) |
|------|-----------------|
| Open Source | Free |
| Commercial | Contact Sales |

---

**Related:** [01.01 User Accounts](../01_user_accounts/) | [01.02 Roles & Permissions](../02_roles_permissions/) | [04.05 Data Retention](../../04_usage_insights/05_data_retention/) | [Domain 01 Overview](../README.md)
