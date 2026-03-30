# Team Analytics

Managers can view organization-wide AI usage statistics and patterns (Commercial license only).

## Purpose

Provide managers with visibility into team-wide AI usage, costs, and trends for budget management and optimization.

## Functional Requirements (EARS Format)

### 1. Team Analytics Access

**Purpose:** Control access to team analytics features.

#### Ubiquitous Requirements

- **TA-04-001:** `The system shall restrict team analytics to managers with Commercial licenses.`
- **TA-04-002:** `The system shall not allow members to access team analytics.`

#### Event-Driven Requirements (Manager Access)

- **TA-04-101:** `When a manager with a Commercial license views team analytics, the system shall display organization-wide usage statistics.`
- **TA-04-102:** `When a manager views team analytics, the system shall update the data every 30 seconds.`

#### Unwanted Behaviour Requirements (Access Restrictions)

- **TA-04-201:** `If a member attempts to view team analytics, then the system shall deny access and display a message indicating a Commercial license is required.`
- **TA-04-202:** `If an Open Source license attempts to view team analytics, then the system shall deny access and display a message indicating the feature requires a Commercial license.`

---

### 2. Aggregated Statistics

**Purpose:** Display organization-wide usage metrics.

#### Event-Driven Requirements (Statistics Display)

- **TA-04-301:** `When a manager views team analytics, the system shall display total requests across all users.`
- **TA-04-302:** `When a manager views team analytics, the system shall display total tokens used across all users.`
- **TA-04-303:** `When a manager views team analytics, the system shall display total estimated costs across all users.`
- **TA-04-304:** `When a manager views team analytics, the system shall display success rate across all users.`

#### Event-Driven Requirements (Time Period Selection)

- **TA-04-305:** `When a manager selects a time period, the system shall display statistics for the selected period across all users.`

---

### 3. User Breakdown and Rankings

**Purpose:** Show per-user usage metrics.

#### Event-Driven Requirements (User Analytics)

- **TA-04-401:** `When a manager views user breakdown, the system shall display usage statistics for each user.`
- **TA-04-402:** `When a manager views user rankings, the system shall display users ranked by token usage.`
- **TA-04-403:** `When a manager views user rankings, the system shall display users ranked by cost.`
- **TA-04-404:** `When a manager views user rankings, the system shall display users ranked by request count.`

---

### 4. Provider and Model Analytics

**Purpose:** Show which providers and models are most used.

#### Event-Driven Requirements (Provider Analytics)

- **TA-04-501:** `When a manager views provider analytics, the system shall display usage breakdown by provider.`
- **TA-04-502:** `When a manager views model analytics, the system shall display usage breakdown by model.`
- **TA-04-503:** `When a manager views tool analytics, the system shall display usage breakdown by AI tool (Claude/Codex/OpenCode).` ✅ **Implemented**

---

### 5. Team Visualizations

**Purpose:** Provide visual insights into team usage patterns.

#### Event-Driven Requirements (Visualization Display)

- **TA-04-601:** `When a manager views team analytics, the system shall display team usage trends over time.`
- **TA-04-602:** `When a manager views team analytics, the system shall display provider usage distribution.`
- **TA-04-603:** `When a manager views team analytics, the system shall display model usage distribution.`
- **TA-04-604:** `When a manager views team analytics, the system shall display tool usage distribution visualization (pie chart and table).` ✅ **Implemented**

## Data Availability

| Tier | Team Data Available |
|------|-------------------|
| Open Source | ❌ Not available |
| Commercial | ✅ Full team analytics |

Team analytics are available only to managers with Commercial licenses.

## Business Rules

- **BR-04-001:** Team analytics require a Commercial license
- **BR-04-002:** Only managers can view team analytics
- **BR-04-003:** Team data includes all users in the organization
- **BR-04-004:** Data is retained per license type (see Data Retention)

---

**Related:** [04.01 Data Collection](../01_data_collection/) | [04.02 Personal Analytics](../02_personal_analytics/) | [04.04 Cost Tracking](../04_cost_tracking/)
