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

## Planned Features (Future)

### User Stories

- As a **manager**, I want to create teams so different groups can have different providers
- As a **manager**, I want to assign users to teams so they get the right configuration
- As a **member**, I want to see providers from all my teams
- As a **manager**, I want to organize users by team for better management

### Team Concepts

#### What is a Team?

A team is a group of users who share:
- Provider configurations
- Settings and policies
- Usage analytics (can be viewed per team)

#### Team Limits by License Type

| License Type | Max Teams |
|--------------|-----------|
| Open Source | 1 |
| Commercial | Unlimited |

#### Providers and Teams

- Providers belong to a specific team
- Each team has its own set of providers
- Members see providers from all teams they belong to
- Managers can view all teams and providers

### User Workflows

#### Creating a Team

**User Story:** As a manager, I want to create a new team.

**Workflow:**
1. Manager navigates to Teams section
2. Clicks "Create Team"
3. Enters team name and description
4. Saves team
5. Team appears in team list
6. Manager can now add providers and users to the team

#### Adding Users to a Team

**User Story:** As a manager, I want to add users to a team.

**Workflow:**
1. Manager opens team details
2. Clicks "Add Members"
3. Selects users from organization
4. Assigns role for each user (manager/member for the team)
5. Saves changes
6. Users receive notification of team assignment
7. Users see new providers on next sync

#### Configuring Team Providers

**User Story:** As a manager, I want to set up providers for a team.

**Workflow:**
1. Manager opens team details
2. Navigates to Providers section
3. Adds, edits, or removes providers
4. Sets provider priority
5. Saves configuration
6. Configuration syncs to all team members

### Multi-Team Membership

Users can belong to multiple teams:
- User sees providers from all their teams
- Usage is attributed to the team associated with the provider used
- Managers can see which teams a user belongs to

### Functional Requirements (Future)

- **FR-001:** Managers can create up to license type limit of teams
- **FR-002:** Each team has separate provider configurations
- **FR-003:** Users can belong to multiple teams
- **FR-004:** Members see providers from all their teams
- **FR-005:** Managers can add/remove users from teams

### Business Rules (Future)

- **BR-001:** Default team created when organization is created
- **BR-002:** Deleting a team deletes its providers and team membership
- **BR-003:** Team name must be unique within organization
- **BR-004:** Users must belong to at least one team
- **BR-005:** Last user cannot be removed from last team

### Edge Cases & Error Handling (Future)

| Scenario | System Behavior |
|----------|----------------|
| Trying to create team beyond limit | Error: "Team limit reached. Commercial license required for multiple teams." |
| Deleting team with active users | Warning: "Team has X members. Reassign or remove them first." |
| Removing user from all teams | Error: "User must belong to at least one team." |
| Duplicate team name | Error: "Team with this name already exists." |

### Success Criteria (Future)

- Managers can create and manage teams
- Team providers sync correctly to members
- Multi-team members see all their providers
- Team limits are enforced
- Usage is correctly attributed to teams

---

**Related:** [03.01 Distribution](../01_distribution/) | [03.03 Offline Mode](../03_offline_mode/)
