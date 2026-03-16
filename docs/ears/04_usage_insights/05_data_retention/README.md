# Data Retention

The system retains usage metadata for different durations based on license type.

## Purpose

Define how long usage data is stored and when it is deleted.

## Functional Requirements (EARS Format)

### 1. Retention Periods by License Type

**Purpose:** Define data retention duration for each license tier.

#### State-Driven Requirements (Open Source License)

- **DR-04-001:** `While an organization has an Open Source license, the system shall retain usage metadata for 7 days.`
- **DR-04-002:** `While the 7-day retention period expires, the system shall delete usage metadata older than 7 days.`

#### State-Driven Requirements (Commercial License)

- **DR-04-003:** `While an organization has a Commercial license, the system shall retain usage metadata for 90 days.`
- **DR-04-004:** `While the 90-day retention period expires, the system shall delete usage metadata older than 90 days.`

---

### 2. Data Deletion

**Purpose:** Handle automatic deletion of expired data.

#### Event-Driven Requirements (Deletion Trigger)

- **DR-04-101:** `When the retention period expires for a usage record, the system shall delete the record.`
- **DR-04-102:** `When data is deleted due to retention policy, the system shall permanently remove the record from storage.`
- **DR-04-103:** `When an organization is deleted, the system shall delete all associated usage data.`

---

### 3. Data Availability in Analytics

**Purpose:** Define what data is visible in analytics based on retention.

#### State-Driven Requirements (Analytics Data Range)

- **DR-04-201:** `While personal analytics are displayed, the system shall only show data within the retention period.`
- **DR-04-202:** `While team analytics are displayed, the system shall only show data within the retention period.`
- **DR-04-203:** `While data export is requested, the system shall only include records within the retention period.`

---

### 4. License Changes and Retention

**Purpose:** Define retention behavior when license changes.

#### Event-Driven Requirements (License Upgrade)

- **DR-04-301:** `When an organization upgrades from Open Source to Commercial, the system shall begin retaining data for 90 days from the upgrade date.`
- **DR-04-302:** `When an organization upgrades to Commercial, the system shall not delete existing data that is within the new 90-day retention period.`

#### Event-Driven Requirements (License Downgrade)

- **DR-04-303:** `When an organization downgrades from Commercial to Open Source, the system shall begin retaining data for 7 days from the downgrade date.`
- **DR-04-304:** `When an organization downgrades to Open Source, the system shall delete data older than 7 days from the downgrade date.`

#### Complex Requirements (Transition Behavior)

- **DR-04-305:** `When a license changes, the system shall apply the new retention policy going forward while preserving existing data within the new retention window.`

---

### 5. Data Retention Policies

**Purpose:** Define comprehensive retention rules.

#### Ubiquitous Requirements

- **DR-04-401:** `The system shall enforce data retention based on the organization's current license type.`
- **DR-04-402:** `The system shall not retain prompt or response content (only metadata).`
- **DR-04-403:** `The system shall automatically delete expired data based on retention policy.`
- **DR-04-404:** `The system shall not provide manual data recovery options after deletion.`

## Retention Periods

| License Type | Data Retention |
|--------------|----------------|
| **Open Source** | 7 days |
| **Commercial** | 90 days |

## Data Deletion Behavior

| Trigger | Behavior |
|---------|----------|
| Retention period expires | Automatic deletion of expired records |
| Organization deleted | Immediate deletion of all data |
| License downgrade | Data older than new retention period deleted |

## Business Rules

- **BR-04-001:** Data retention is determined by current license type
- **BR-04-002:** Deleted data cannot be recovered
- **BR-04-003:** License changes affect retention going forward (with transition window)
- **BR-04-004:** Only metadata is retained (no content)

---

**Related:** [04.01 Data Collection](../01_data_collection/) | [04.02 Personal Analytics](../02_personal_analytics/) | [04.03 Team Analytics](../03_team_analytics/) | [01.04 Licensing](../../01_identity_and_access/04_licensing/)
