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

## User Roles

### Manager

**Who:** Team administrators, tech leads, engineering managers

**Can do:**
- Add and deactivate users
- Assign roles (manager/member)
- Configure AI providers
- View team-wide analytics
- Manage license
- Push configuration to team members

**Cannot do:**
- Impersonate other users
- Modify historical usage data
- Access other tenants' data

### Member

**Who:** Individual developers, end users

**Can do:**
- Use configured AI providers
- View personal usage statistics
- View personal request history
- Pull configuration from server
- Export personal usage data

**Cannot do:**
- Modify team provider settings
- View other users' data
- Push configuration to server
- Access team analytics

## Quick Reference

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

## Related Documentation

- **User Accounts:** [`01.01 User Accounts`](01_user_accounts/) - Creating accounts and logging in
- **Roles & Permissions:** [`01.02 Roles & Permissions`](02_roles_permissions/) - Detailed permission matrix
- **Multi-Tenancy:** [`01.03 Multi-Tenancy`](03_multi_tenancy/) - Organization isolation
- **Licensing:** [`01.04 Licensing`](04_licensing/) - License tiers and feature limits

---

**Next:** [02 Provider Management](../02_provider_management/)
