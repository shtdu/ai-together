# Licensing

AI Together uses a 2-type license system: Open Source (free) and Commercial (paid). The system gracefully degrades when no license is active.

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

**Eligibility:** For open source projects or personal non-commercial use.

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

## License Status

### Active License

**Definition:** License is valid and not expired.

**Behavior:**
- All type-specific features available
- No seat or provider limits enforced
- Team limit enforced based on license type (1 for Open Source, unlimited for Commercial)

### Expired Commercial License

**Definition:** Commercial license was active but has passed expiration date.

**Behavior:**
- Existing users can still log in
- Existing configurations still work
- Analytics and historical data remain accessible
- **New team creation blocked** (reverts to 1 team limit)
- **Advanced analytics features disabled**
- Manager sees renewal prompt

### No License (Default to Open Source)

**Definition:** Organization has never activated a license.

**Behavior:**
- System functions as Open Source type
- All Open Source features available
- 1 team limit enforced

## User Stories

### License Activation

- As a **manager**, I want to activate a commercial license so that my team can access advanced analytics features
- As a **manager**, I want to view our license status so I know when it expires
- As a **user**, I want to understand which features are available based on our license type

### Open Source Usage

- As an **open source maintainer**, I want to use the software for free with unlimited contributors
- As an **individual developer**, I want to use the software for personal projects without restrictions

## Functional Requirements

### License Activation

- **FR-001:** Managers can activate a commercial license by entering license key
- **FR-002:** System validates license signature before activation
- **FR-003:** License features unlock immediately upon activation
- **FR-004:** License status visible to all users

### License Type Enforcement

- **FR-005:** Open Source license limited to 1 team
- **FR-006:** Commercial license supports unlimited teams
- **FR-007:** System shows current license type to all users
- **FR-008:** System prevents creating teams beyond license type limit

### Feature Availability

- **FR-009:** Team analytics dashboard requires commercial license
- **FR-010:** Usage data export requires commercial license
- **FR-011:** Data retention respects license type limits (7 days Open Source, 90 days Commercial)
- **FR-012:** Cost calculation features require commercial license

## Business Rules

- **BR-001:** Graceful degradation - system always functions, features may lock
- **BR-002:** Expired commercial licenses preserve data access but disable advanced features
- **BR-003:** No seat limits - both license types support unlimited team members
- **BR-004:** No provider limits - both license types support unlimited providers per kind
- **BR-005:** License activation takes effect immediately

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
