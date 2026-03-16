# Usage Insights

The system tracks AI tool usage and presents insights to users. Managers see team-wide data, while members see personal statistics.

## Purpose

Provide visibility into AI usage patterns, costs, and trends for both individuals and teams.

## Subdomains

| Subdomain | Status | Description |
|-----------|--------|-------------|
| **Data Collection** | ✅ Implemented | What data is collected and how |
| **Personal Analytics** | ✅ Implemented | Individual user insights |
| **Team Analytics** | ✅ Implemented | Organization-wide insights (Commercial) |
| **Cost Tracking** | ✅ Implemented | How costs are calculated and displayed |
| **Data Retention** | ✅ Implemented | How long data is kept by license type |

## Quick Facts

- **Privacy-first:** Only metadata collected, no prompt/response content
- **Collected data:** Model name, tokens, provider, timestamp, duration, status
- **Personal analytics:** Available to all users
- **Team analytics:** Available to managers (Commercial)
- **Cost calculation:** Estimates based on provider pricing models

## Update Intervals

Analytics dashboards update at different intervals based on scope:

| Analytics Type | Update Interval | Rationale |
|----------------|-----------------|-----------|
| **Personal Analytics** | 5 seconds | Individual data, low volume, immediate feedback valuable |
| **Team Analytics** | 30 seconds | Aggregated data, higher volume, balanced freshness and load |

## Functional Requirements (EARS Format)

### 1. Data Collection

**Purpose:** Track usage for analytics and billing while respecting user privacy.

#### Ubiquitous Requirements

- **UI-04-001:** `The system shall collect only metadata about AI tool requests.`
- **UI-04-002:** `The system shall never store prompt or response content.`
- **UI-04-003:** `The system shall upload metadata to the server in batches every 30 seconds.`
- **UI-04-004:** `The system shall queue metadata locally when offline.`

---

### 2. Personal Analytics

**Purpose:** Enable users to view their own usage statistics.

#### Event-Driven Requirements

- **UI-04-101:** `When a user views personal analytics, the system shall display usage statistics in real-time (within 5 seconds).`
- **UI-04-102:** `When a user exports personal data, the system shall generate CSV/JSON files.`
- **UI-04-103:** `When a user filters personal analytics, the system shall display data matching the filters.`

---

### 3. Team Analytics (Commercial)

**Purpose:** Provide managers with organization-wide usage visibility.

#### Event-Driven Requirements

- **UI-04-201:** `When a manager views team analytics, the system shall display aggregated statistics for all users (Commercial only).`
- **UI-04-202:** `When a manager views team analytics, the system shall update data every 30 seconds.`
- **UI-04-203:** `When a manager views user rankings, the system shall display users ranked by usage metrics.`

---

### 4. Cost Tracking

**Purpose:** Calculate and display estimated costs for AI usage.

#### Event-Driven Requirements

- **UI-04-301:** `When an AI tool request completes, the system shall calculate estimated cost based on token usage and provider pricing.`
- **UI-04-302:** `When cost analytics are displayed, the system shall show costs are approximate.`
- **UI-04-303:** `When managers view team costs, the system shall display total and per-user costs.`

---

### 5. Data Retention

**Purpose:** Define how long usage data is stored.

#### State-Driven Requirements

- **UI-04-401:** `While an organization has an Open Source license, the system shall retain usage metadata for 7 days.`
- **UI-04-402:** `While an organization has a Commercial license, the system shall retain usage metadata for 90 days.`
- **UI-04-403:** `While the retention period expires, the system shall automatically delete expired data.`

---

## User Stories

- As a **member**, I want to see my token usage so I can track my consumption
- As a **manager**, I want to see team costs so I can manage budget
- As a **manager**, I want to know which providers are most used so I can optimize
- As a **member**, I want to export my data so I can analyze it further

## Related Documentation

- **Data Collection:** [`04.01 Data Collection`](01_data_collection/) - Privacy and what's tracked
- **Personal Analytics:** [`04.02 Personal Analytics`](02_personal_analytics/) - Individual insights
- **Team Analytics:** [`04.03 Team Analytics`](03_team_analytics/) - Organization insights
- **Cost Tracking:** [`04.04 Cost Tracking`](04_cost_tracking/) - Cost calculation
- **Data Retention:** [`04.05 Data Retention`](05_data_retention/) - Retention policies

---

**Next:** [User Interfaces](../05_user_interfaces/)
