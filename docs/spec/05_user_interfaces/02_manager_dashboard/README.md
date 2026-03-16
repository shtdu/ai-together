# Manager Dashboard

The Manager web dashboard enables team administration, provider configuration, and analytics viewing.

## Purpose

Provide managers with a comprehensive interface for managing their team's AI tool configuration and usage.

## User Personas

- **Manager** - Team administrator responsible for AI tool configuration and team management
- **Tech Lead** - Technical lead who manages providers and monitors team usage
- **Team Administrator** - Anyone responsible for team AI tool management

## User Stories

- As a **manager**, I want to see team usage at a glance
- As a **manager**, I want to add and manage team members
- As a **manager**, I want to configure AI providers for my team
- As a **manager**, I want detailed analytics on team usage
- As a **manager**, I want to view and manage our license

## Access

- **URL:** `http://your-server:8080` (or custom domain)
- **Authentication:** Required (manager role only)
- **Browsers:** Chrome 100+, Safari 15+, Firefox 100+, Edge 100+

## Key Features

### 1. Dashboard Overview

**Description:** Shows team usage summary at a glance.

**Displayed Information:**
- Total requests (selected time range)
- Total tokens
- Total cost
- Request trend chart
- Top providers
- Top users

**Time Range Options:**
- Today
- Last 7 days
- Last 30 days
- Custom range

### 2. User Management

**Description:** Add and manage team members.

**Capabilities:**
- Add new users (by email invitation)
- Deactivate users (preserves data)
- Assign roles (manager/member)
- View user status (active/inactive)
- View per-user usage

**User List Columns:**
- Name
- Email
- Role
- Status (active/inactive)
- Last active
- Actions (edit, deactivate)

### 3. Provider Management

**Description:** Configure AI providers for the team.

**Capabilities:**
- Add new provider
- Edit provider configuration
- Delete provider
- Test provider connectivity
- Enable/disable provider
- Set provider priority
- View provider statistics

**Provider Card Shows:**
- Provider name and type
- API endpoint (masked)
- Status (enabled/disabled)
- Priority level
- Request count
- Success rate
- Average response time

### 4. Analytics

**Description:** Detailed analytics on team usage (Commercial license required).

**Available Views:**
- Provider analytics (usage, performance, cost)
- User analytics (usage by person)
- Request history (detailed log with filters)
- Cost breakdown (by provider, model, user)

**License Requirement:** Team analytics requires Commercial license. Open Source users see personal analytics only.

**Filters:**
- Date range
- Provider
- Model
- User
- Team

### 5. License Management

**Description:** View and manage license information.

**Displayed Information:**
- License type (Open Source/Commercial)
- License status (active/expired)
- Expiration date
- Days remaining
- Feature availability

**Actions:**
- Activate new license key
- Contact sales (upgrade)

## User Interface Overview

### Navigation

```
┌─────────────────────────────────────────────┐
│ ☰  AI Together        [Admin] [Logout]    │
├─────────────────────────────────────────────┤
│ Dashboard                                     │
│ Users                                        │
│ Providers                                    │
│ Analytics                                    │
│ License                                      │
└─────────────────────────────────────────────┘
```

### Dashboard Page

```
┌─────────────────────────────────────────────┐
│ Dashboard              [Last 7 days ▼]       │
├─────────────────────────────────────────────┤
│ ┌──────────┐ ┌──────────┐ ┌─────────────┐  │
│ │ Requests │ │  Tokens  │ │    Cost     │  │
│ │  12,450  │ │  4.5M    │ │   $125      │  │
│ └──────────┘ └──────────┘ └─────────────┘  │
├─────────────────────────────────────────────┤
│ [Request Trend Chart - line graph]          │
├─────────────────────────────────────────────┤
│ Top Providers          Top Users             │
│ 1. Anthropic (45%)     1. alice@...         │
│ 2. OpenAI (35%)        2. bob@...           │
│ 3. Custom (20%)        3. carol@...         │
└─────────────────────────────────────────────┘
```

## User Workflows

### Adding a New Provider

1. Navigate to Providers page
2. Click "Add Provider"
3. Select provider type (Claude/Codex/OpenCode)
4. Enter provider name
5. Enter API endpoint URL
6. Enter API key
7. Configure model mappings (optional)
8. Set priority level
9. Click "Test" to verify
10. Click "Save"

### Inviting a User

1. Navigate to Users page
2. Click "Add User"
3. Enter email address
4. Select role (manager/member)
5. Click "Send Invitation"
6. User receives invitation email

### Viewing Analytics

1. Navigate to Analytics page
2. Select desired view (providers/users/requests)
3. Choose time range
4. Apply filters as needed
5. View results and charts
6. Export data (CSV/JSON) if needed

## Functional Requirements

- **FR-001:** Dashboard must load within 3 seconds
- **FR-002:** All actions must complete within 10 seconds
- **FR-003:** Data must update in real-time (within 30 seconds)
- **FR-004:** Charts must be interactive (zoom, pan, hover)
- **FR-005:** Export must handle up to 100,000 records

## Business Rules

- **BR-001:** Only managers can access the dashboard
- **BR-002:** Team analytics requires Commercial license
- **BR-003:** Last manager cannot be deactivated (must have at least one)

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Provider test fails | Show error: "Connection failed. Check configuration." |
| No analytics data | Show: "No data available for selected time range." |
| License expired | Show warning banner: "License expired. Renew to continue full access." |

## Success Criteria

- Managers can complete all intended actions
- Dashboard loads quickly
- Data is accurate and up-to-date
- Error messages are clear and actionable
- Interface works on supported browsers

---

**Related:** [05.01 Member App](../01_member_app/) | [Domain 05 Overview](../README.md)
