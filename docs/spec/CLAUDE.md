# spec — MECE Product Documentation

This directory contains **product-focused documentation** for AI Together, structured using the **MECE** methodology: **M**utually **E**xclusive and **C**ollectively **E**xhaustive.

## Purpose

- **Audience:** Product managers, stakeholders, and engineering teams who need to understand product requirements and features.
- **Scope:** The entire product feature set is covered exactly once—no overlap between sections, no missing areas.
- **Use:** Feature planning, requirement gathering, stakeholder communication, and acceptance criteria definition.

## Naming Convention

**Level 1 (Domains):** `NN_domain_name/` - e.g., `01_identity_and_access/`
**Level 2 (Subdomains):** `NN_subdomain_name/` - e.g., `01_user_accounts/`

This creates numbered paths like `01_identity_and_access/01_user_accounts/` for easy cross-referencing.

## MECE in Practice

**Mutually exclusive**
Categories do not overlap. Each feature belongs to exactly one top-level domain.

**Collectively exhaustive**
Categories together cover everything. Every product capability maps to one of the six domains.

## Product Feature Taxonomy (MECE)

The product is decomposed into **six domains**:

| Domain | Covers | Examples |
|--------|--------|----------|
| **1. Identity & access** | Who can use the system and what they can do | Login, roles, permissions, licenses, multi-tenant isolation |
| **2. Provider management** | AI service configuration and request routing | Provider setup, failover, model mapping |
| **3. Configuration distribution** | How settings flow to team members | Server sync, team settings, offline mode |
| **4. Usage insights** | What data is collected and how it's presented | Token tracking, cost calculation, analytics dashboards |
| **5. User interfaces** | How people interact with the product | Desktop app, web dashboard, accessibility |
| **6. System behaviors** | Cross-cutting behaviors and constraints | Privacy, performance, security policies |

## Document Template

Each feature document should follow this structure:

```markdown
# [Feature Name]

## Purpose
[What problem does this solve? For whom? Why does it exist?]

## User Personas
[Who are the users affected by this feature?]

## User Stories
- As a [persona], I want [feature], so that [benefit]

## Functional Requirements
- FR-XXX: The system shall [action] when [condition]
- FR-XXX: The system shall not [action] when [condition]

## Business Rules
- BR-XXX: [Constraint or rule that affects behavior]

## User Workflows
[Step-by-step description from user's perspective]

## Edge Cases & Error Handling
[What happens when things go wrong?]

## Success Criteria
[How do we know this feature works correctly?]
```

## Writing Guidelines

### DO:
- Describe WHAT the system does, not HOW it's implemented
- Focus on user-visible behavior and experiences
- Include user stories and acceptance criteria
- Define business rules and constraints
- Describe edge cases and error handling
- Use plain language that non-engineers can understand

### DON'T:
- Include code snippets
- Show database schemas
- List file paths or implementation details
- Describe internal algorithms
- Include API endpoint specifications
- Reference technical components (unless necessary for clarity)

## Document Map

| Domain | Folder | Subdomains |
|--------|--------|------------|
| Product overview | `00_overview/` | - |
| Identity & access | `01_identity_and_access/` | 01_user_accounts, 02_roles_permissions, 03_multi_tenancy, 04_licensing |
| Provider management | `02_provider_management/` | 01_provider_configuration, 02_request_routing, 03_model_mapping |
| Configuration distribution | `03_configuration_sync/` | 01_distribution, 02_teams, 03_offline_mode |
| Usage insights | `04_usage_insights/` | 01_data_collection, 02_personal_analytics, 03_team_analytics, 04_cost_tracking, 05_data_retention |
| User interfaces | `05_user_interfaces/` | 01_member_app, 02_manager_dashboard, 03_accessibility |
| System behaviors | `06_system_behaviors/` | 01_privacy, 02_security, 03_performance, 04_compliance |

## Related Documentation

- **Technical implementation:** `../epic_0/architecture.md` - System architecture
- **Module development:** `../../member/CLAUDE.md`, `../../server/CLAUDE.md` - Technical guides
- **Design principles:** `../epic_0/constitution.md` - Design principles and technology choices

---

**Remember:** This directory is for PRODUCT documentation. Technical implementation details belong in module-specific CLAUDE.md files or architecture docs.
