# EARS Requirements Syntax Guide

This directory contains the **EARS (Easy Approach to Requirements Syntax)** methodology for writing precise, testable software requirements. EARS imposes a strict, rule-based temporal logic on textual requirements, ensuring that software behaviors are deterministic and precisely mapped to system execution states.

## Purpose

- **Audience:** Product managers, stakeholders, and engineering teams who need to write precise, unambiguous software requirements
- **Goal:** Eliminate logic omissions, algorithmic ambiguity, and untestability from requirements documentation
- **Use:** Writing functional requirements that are deterministic, testable, and machine-parsable

## EARS Core Syntax

The architectural integrity of a software requirement in EARS relies on an immutable grammatical sequence:

```
<optional preconditions> <optional trigger> the <system name> shall <system response>
```

- **`<system name>`**: The computational actor (e.g., the specific microservice, class, or operating system)
- **`<system response>`**: The mandatory algorithmic execution or output
- **Important**: Terms like "and/or" must be explicitly separated into exclusive conditions to avoid combinatorial explosion

## EARS Pattern Taxonomy

EARS categorizes software behaviors into **five distinct patterns**. All requirements must use one of these patterns:

### 1. Ubiquitous Software Requirements

**Purpose:** Defines fundamental, global software properties, baseline constraints, or continuous background services that do not require an external stimulus to execute.

**Syntax:** `The <system name> shall <system response>.`

**Examples:**
- `The software shall be written in Go.`
- `The software package shall include an installer.`
- `The system shall allow the admin to verify new providers.`

### 2. Event-Driven Software Requirements

**Purpose:** Governs discrete, strictly defined API invocations, asynchronous message receptions, or user interface interactions. A specific system response is mandated *when and only when* the triggering event is detected.

**Syntax:** `When <trigger>, the <system name> shall <system response>.`

**Examples:**
- `When an order is shipped and order terms are not 'Prepaid', the system shall create an invoice.`
- `When the user selects the caller count from the menu, the software shall display a count of the number of participants in the audio call in the UI.`
- `When the driver has accepted the package request, the system shall send the package request to the selected driver, and the rider will be notified.`

### 3. State-Driven Software Requirements

**Purpose:** Dictates modal software behaviors, session states, or specific runtime environments. The software response remains actively enforced exclusively for the duration that the operational state evaluates to true.

**Syntax:** `While <system state>, the <system name> shall <system response>.`

**Examples:**
- `While the mute button is depressed, the software shall mute the microphone.`
- `While in manufacturing mode, the software shall boot without user intervention.`
- `While the license is expired, the system shall deny access to premium features.`

### 4. Optional Feature Software Requirements

**Purpose:** Critical for Software Product Line Engineering (SPLE), this pattern manages modular configurations and feature flags, decoupling variable logic from the core application codebase.

**Syntax:** `Where <feature is included>, the <system name> shall <system response>.`

**Examples:**
- `Where a thesaurus is part of the software package, the installer shall prompt the user before installing the thesaurus.`
- `Where cash on delivery is available, the system shall allow the customer to select the payment method.`
- `Where multi-tenancy is enabled, the system shall isolate tenant data at the database level.`

### 5. Unwanted Behaviour Software Requirements

**Purpose:** Establishes the foundational logic for exception handling, cybersecurity fault tolerance, and data validation. It articulates the precise computational mitigation required when the software encounters malformed inputs or backend failures.

**Syntax:** `If <trigger>, then the <system name> shall <system response>.`

**Examples:**
- `If an invalid credit card number is entered, then the website shall display 'please re-enter credit card details'.`
- `If the memory checksum is invalid, then the software shall display an error message.`
- `If the alarm software detects that a sensor has malfunctioned, then the alarm software shall phone the Alarm Company to report the malfunction.`

### 6. Complex Software Requirements

**Purpose:** Synthesizes dense Boolean logic, fusing specific runtime states with triggering events or exceptions.

**Syntax:** `<Multiple conditions>, the <system name> shall <system response>.`

**Examples:**
- `While on DC power, if the software detects an error, then the software shall cache the error message instead of writing the error message to disk.`
- `When more than 3 incorrect login attempts occur for a single user ID within a 30 minute period, the software shall lock the account associated with that user ID.`

## MECE Product Documentation Structure

EARS requirements are organized within the **MECE** (Mutually Exclusive, Collectively Exhaustive) product documentation structure:

### Naming Convention

- **Level 1 (Domains):** `NN_domain_name/` - e.g., `01_identity_and_access/`
- **Level 2 (Subdomains):** `NN_subdomain_name/` - e.g., `01_user_accounts/`

This creates numbered paths like `01_identity_and_access/01_user_accounts/` for easy cross-referencing.

### MECE Principles

**Mutually Exclusive**: Categories do not overlap. Each feature belongs to exactly one top-level domain.

**Collectively Exhaustive**: Categories together cover everything. Every product capability maps to one of the six domains.

### Product Feature Taxonomy

| Domain | Covers | Examples |
|--------|--------|----------|
| **1. Identity & access** | Who can use the system and what they can do | Login, roles, permissions, licenses, multi-tenant isolation |
| **2. Provider management** | AI service configuration and request routing | Provider setup, failover, model mapping |
| **3. Configuration distribution** | How settings flow to team members | Server sync, team settings, offline mode |
| **4. Usage insights** | What data is collected and how it's presented | Token tracking, cost calculation, analytics dashboards |
| **5. User interfaces** | How people interact with the product | Desktop app, web dashboard, accessibility |
| **6. System behaviors** | Cross-cutting behaviors and constraints | Privacy, performance, security policies |

## EARS Document Template

Each feature document should follow this structure, using EARS syntax for all functional requirements:

```markdown
# [Feature Name]

## Purpose
[What problem does this solve? For whom? Why does it exist?]

## User Personas
[Who are the users affected by this feature?]

## User Stories
- As a [persona], I want [feature], so that [benefit]

## Functional Requirements (EARS Format)

### Ubiquitous Requirements
- FR-XXX: `The <system name> shall <system response>.`

### Event-Driven Requirements
- FR-XXX: `When <trigger>, the <system name> shall <system response>.`

### State-Driven Requirements
- FR-XXX: `While <system state>, the <system name> shall <system response>.`

### Optional Feature Requirements
- FR-XXX: `Where <feature is included>, the <system name> shall <system response>.`

### Unwanted Behaviour Requirements
- FR-XXX: `If <trigger>, then the <system name> shall <system response>.`

### Complex Requirements
- FR-XXX: `<Multiple conditions>, the <system name> shall <system response>.`

## Business Rules
- BR-XXX: [Constraint or rule that affects behavior]

## User Workflows
[Step-by-step description from user's perspective]

## Edge Cases & Error Handling
[What happens when things go wrong? Use EARS unwanted behaviour pattern]

## Success Criteria
[How do we know this feature works correctly?]
```

## Requirement Metadata Integration

Each EARS requirement must be embedded within a structured metadata record to ensure full lifecycle traceability, impact analysis, and verification. A rigorous software requirement record must encompass:

1. **Name:** A discrete, programmatic identifier (e.g., `Create_Invoice`)
2. **Requirement:** The strictly formatted EARS textual syntax
3. **Rationale:** The business or architectural logic dictating the function
4. **Priority & Criticality:** Boolean or hierarchical metrics dictating implementation sequencing
5. **Traceability:** Bidirectional linkages to source logic, peer dependencies, and downstream tests

## Writing Guidelines

### DO:
- Use EARS patterns for ALL functional requirements
- Describe WHAT the system does, not HOW it's implemented
- Focus on user-visible behavior and experiences
- Include user stories and acceptance criteria
- Define business rules and constraints
- Describe edge cases and error handling using the unwanted behaviour pattern
- Use plain language that non-engineers can understand
- Ensure each requirement uses one of the five EARS patterns
- Mark quoted syntax with backticks: `` `The system shall...` ``

### DON'T:
- Write ambiguous requirements like "The system should handle errors"
- Use "and/or" - separate into explicit conditions
- Include code snippets (use backticks for syntax only)
- Show database schemas
- List file paths or implementation details
- Describe internal algorithms
- Include API endpoint specifications
- Reference technical components (unless necessary for clarity)

## Advanced EARS (Adv-EARS)

For automated software verification and toolchain integration, EARS requirements can be extended to Adv-EARS:

- **Lexical Mapping:** Maps grammatical elements to Object-Oriented paradigms
  - `<system name>` → `<entity>` (actors or software classes)
  - `<system response>` → `<functionality>` (executable use cases)
- **CFG Parsing:** Context-Free Grammar can programmatically parse requirements to identify system entities and synthesize use case models

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

**Remember:** EARS requirements eliminate ambiguity by enforcing strict grammatical patterns. Always use one of the five EARS patterns when writing functional requirements.
