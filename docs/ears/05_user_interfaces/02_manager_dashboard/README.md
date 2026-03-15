# Manager Dashboard

The Manager web dashboard enables team administration, provider configuration, and analytics viewing.

## Purpose

Provide managers with a comprehensive interface for managing their team's AI tool configuration and usage.

## Functional Requirements (EARS Format)

### 1. Dashboard Access

**Purpose:** Control access to the manager dashboard.

#### Ubiquitous Requirements

- **MD-05-001:** `The system shall require authentication to access the manager dashboard.`
- **MD-05-002:** `The system shall restrict dashboard access to users with the manager role.`

#### Event-Driven Requirements (Authentication)

- **MD-05-101:** `When a user accesses the dashboard URL, the system shall prompt for authentication credentials.`
- **MD-05-102:** `When a user authenticates with the manager role, the system shall display the dashboard.`
- **MD-05-103:** `When a user authenticates with the member role, the system shall deny access to the dashboard.`

#### Unwanted Behaviour Requirements (Access Denial)

- **MD-05-201:** `If a member attempts to access the dashboard, the system shall display an error message indicating manager access is required.`

---

### 2. Dashboard Overview

**Purpose:** Display team usage summary at a glance.

#### Event-Driven Requirements (Overview Display)

- **MD-05-301:** `When a manager views the dashboard, the system shall display total requests for the selected time range.`
- **MD-05-302:** `When a manager views the dashboard, the system shall display total tokens used.`
- **MD-05-303:** `When a manager views the dashboard, the system shall display total estimated costs.`
- **MD-05-304:** `When a manager views the dashboard, the system shall display a request trend chart.`
- **MD-05-305:** `When a manager views the dashboard, the system shall display top providers by usage.`
- **MD-05-306:** `When a manager views the dashboard, the system shall display top users by usage.`

#### Event-Driven Requirements (Time Range Selection)

- **MD-05-307:** `When a manager selects a time range (Today, Last 7 days, Last 30 days, Custom), the system shall update all dashboard metrics for the selected period.`

---

### 3. User Management

**Purpose:** Enable user management operations.

#### Event-Driven Requirements (User Operations)

- **MD-05-401:** `When a manager adds a new user, the system shall send an invitation email.`
- **MD-05-402:** `When a manager deactivates a user, the system shall preserve the user's data and prevent login.`
- **MD-05-403:** `When a manager assigns a role to a user, the system shall update the user's role immediately.`
- **MD-05-404:** `When a manager views the user list, the system shall display user status (active/inactive).`
- **MD-05-405:** `When a manager views per-user usage, the system shall display usage statistics for the selected user.`

---

### 4. Provider Configuration

**Purpose:** Enable provider setup and management.

#### Event-Driven Requirements (Provider Operations)

- **MD-05-501:** `When a manager adds a provider, the system shall display the provider configuration form.`
- **MD-05-502:** `When a manager saves a provider, the system shall validate and store the provider configuration.`
- **MD-05-503:** `When a manager edits a provider, the system shall allow modification of provider settings.`
- **MD-05-504:** `When a manager deletes a provider, the system shall remove the provider configuration.`
- **MD-05-505:** `When a manager tests a provider, the system shall verify connectivity and display the result.`

---

### 5. Team Analytics

**Purpose:** Provide detailed team usage insights.

#### Event-Driven Requirements (Analytics Display)

- **MD-05-601:** `When a manager views team analytics, the system shall display usage data aggregated across all users.`
- **MD-05-602:** `When a manager views team analytics, the system shall display cost breakdown by provider, model, and user.`
- **MD-05-603:** `When a manager views team analytics, the system shall display user rankings by usage and cost.`
- **MD-05-604:** `When a manager filters team analytics, the system shall update the display based on the selected filters.`

---

### 6. License Management

**Purpose:** Enable license viewing and activation.

#### Event-Driven Requirements (License Operations)

- **MD-05-701:** `When a manager views license status, the system shall display the current license type and expiration.`
- **MD-05-702:** `When a manager activates a commercial license, the system shall validate and apply the license key.`
- **MD-05-703:** `When a license expires, the system shall display renewal options and feature limitations.`

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

---

**Related:** [05.01 Member App](../01_member_app/) | [05.03 Accessibility](../03_accessibility/)
