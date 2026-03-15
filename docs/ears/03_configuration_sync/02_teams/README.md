# Teams

> **DEFERRED** - This feature is not currently implemented. The system supports a single default team per organization. Multi-team support is planned for a future release.

## Purpose

Enable organizations to structure their AI tool usage by teams, with separate configurations and membership.

## Current Implementation

**Single Team Mode:**
- Each organization has exactly one team (default team)
- All organization members belong to this team
- Provider configurations apply organization-wide
- No team management UI is exposed

## Functional Requirements (EARS Format) - Current Implementation

### 1. Single Team Mode

**Purpose:** Define the current single-team behavior.

#### Ubiquitous Requirements

- **TM-03-001:** `The system shall provide exactly one default team per organization.`
- **TM-03-002:** `The system shall assign all organization members to the default team.`
- **TM-03-003:** `The system shall apply provider configurations organization-wide.`
- **TM-03-004:** `The system shall not expose team management UI in the current implementation.`

---

## Functional Requirements (EARS Format) - Planned Features

### 2. Multi-Team Support (Planned)

**Purpose:** Enable organizations to create multiple teams with separate configurations.

#### Event-Driven Requirements (Team Creation)

- **TM-03-101:** `When a manager creates a new team, the system shall assign a unique team identifier.`
- **TM-03-102:** `When a manager creates a new team, the system shall associate the team with the organization.`
- **TM-03-103:** `When a manager creates a new team, the system shall allow the manager to assign team members.`
- **TM-03-104:** `When a manager creates a new team, the system shall allow the manager to configure providers for the team.`

#### Event-Driven Requirements (Team Membership)

- **TM-03-201:** `When a manager assigns a user to a team, the system shall add the user to the team.`
- **TM-03-202:** `When a manager removes a user from a team, the system shall remove the user from the team.`
- **TM-03-203:** `When a user belongs to multiple teams, the system shall allow the user to access providers from all their teams.`

#### Unwanted Behaviour Requirements (Team Limits)

- **TM-03-301:** `If an Open Source license attempts to create more than 1 team, then the system shall display an error indicating the team limit has been reached.`
- **TM-03-302:** `If a Commercial license attempts to create teams, then the system shall allow unlimited team creation.`

---

### 3. Team Provider Configuration (Planned)

**Purpose:** Enable separate provider configurations per team.

#### Ubiquitous Requirements

- **TM-03-401:** `The system shall associate providers with specific teams.`
- **TM-03-402:** `The system shall allow each team to have its own set of providers.`
- **TM-03-403:** `The system shall allow members to access providers from all teams they belong to.`
- **TM-03-404:** `The system shall allow managers to view all teams and their providers.`

---

### 4. Team Analytics (Planned)

**Purpose:** Enable usage analytics viewing per team.

#### Event-Driven Requirements (Analytics Access)

- **TM-03-501:** `When a manager views team analytics, the system shall display usage data filtered by team.`
- **TM-03-502:** `When a member views personal usage, the system shall show usage across all teams the member belongs to.`

## Team Limits by License Type

| License Type | Max Teams |
|--------------|-----------|
| Open Source | 1 |
| Commercial | Unlimited |

## Planned User Stories

- As a **manager**, I want to create teams so different groups can have different providers
- As a **manager**, I want to assign users to teams so they get the right configuration
- As a **member**, I want to see providers from all my teams
- As a **manager**, I want to organize users by team for better management

## Related Documentation

- **Distribution:** [`03.01 Distribution`](../01_distribution/) - How configuration syncs
- **Offline Mode:** [`03.03 Offline Mode`](../03_offline_mode/) - Working without server
