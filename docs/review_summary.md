# Requirements Convention Comparison: `spec/` vs `ears/`

**Review Date:** 2026-03-19
**Reviewer:** Product Analysis

---

## Executive Summary

This document compares two requirements documentation conventions used in this project:
- **`docs/spec/`** - Traditional MECE product documentation
- **`docs/ears/`** - EARS (Easy Approach to Requirements Syntax) methodology

**Recommendation:** Adopt `docs/ears/` as the primary requirements specification for functional requirements, while using `docs/spec/` for stakeholder communication and product overviews.

---

## Overview

| Aspect | `docs/spec/` | `docs/ears/` |
|--------|--------------|--------------|
| **Methodology** | Traditional MECE product documentation | EARS (Easy Approach to Requirements Syntax) |
| **Maturity Level** | Basic/Intermediate | Advanced/Formal |
| **Primary Focus** | Communication & planning | Precision & testability |
| **Target Audience** | PMs, stakeholders, engineers | Engineers, testers, auditors |

---

## Quality Analysis

### 1. Requirements Precision

**`docs/spec/`** - Traditional approach:
```
FR-001: Managers can activate a commercial license by entering license key
FR-002: System validates license signature before activation
```
- ✅ Simple, readable
- ❌ Ambiguous - doesn't specify exact conditions or responses
- ❌ Missing temporal context (when exactly does this happen?)

**`docs/ears/`** - EARS approach:
```
LC-01-101: `When a manager enters a commercial license key, the system shall validate the license signature.`
LC-01-102: `When a manager enters a valid commercial license key, the system shall activate the commercial license.`
```
- ✅ Precise trigger and response
- ✅ Deterministic behavior
- ✅ Maps directly to test cases

**Winner:** `docs/ears/` - significantly more precise

---

### 2. Coverage Completeness

**`docs/spec/`** Licensing example:
- 12 functional requirements
- 5 business rules
- ~170 lines

**`docs/ears/`** Licensing example:
- 28 functional requirements (organized by EARS pattern)
- 5 business rules
- ~200 lines

The EARS convention forces you to think about:
- **Ubiquitous** requirements (always true)
- **Event-driven** requirements (trigger → response)
- **State-driven** requirements (while in state X, do Y)
- **Unwanted behavior** requirements (if error, then handle)
- **Complex** requirements (combinations)

**Winner:** `docs/ears/` - more comprehensive coverage

---

### 3. Testability

| Metric | `docs/spec/` | `docs/ears/` |
|--------|--------------|--------------|
| Direct test mapping | Poor | Excellent |
| Acceptance criteria clarity | Medium | High |
| Edge case coverage | Ad-hoc | Systematic |

EARS patterns map directly to test patterns:
- `When X, the system shall Y` → Integration test with trigger X, assert Y
- `While X, the system shall Y` → State-based test
- `If X, then the system shall Y` → Error handling test

**BDD Alignment Example:**

```gherkin
# EARS requirement:
# LC-01-501: If a user with an Open Source license attempts to access
# the team analytics dashboard, then the system shall deny the access
# and display a message indicating a commercial license is required.

Scenario: Open source user cannot access team analytics
  Given the organization has an Open Source license
  When the user attempts to access the team analytics dashboard
  Then the system shall deny access
  And display "This feature requires a commercial license"
```

---

### 4. Traceability

**`docs/spec/`:**
```
FR-009: Team analytics dashboard requires commercial license
```
- Generic IDs (FR-001, FR-002...)
- Hard to trace to domain/component

**`docs/ears/`:**
```
LC-01-501: `If a user with an Open Source license attempts to access
the team analytics dashboard, then the system shall deny the access
and display a message indicating a commercial license is required.`
```
- Domain-prefixed IDs (`LC-01` = Licensing, domain 01)
- Easy to trace requirement origin
- Enables bidirectional traceability (requirement ↔ test)

**ID Convention in EARS:**
| Prefix | Domain |
|--------|--------|
| `LC-01-XXX` | Licensing (01_identity_and_access/04_licensing) |
| `PC-02-XXX` | Provider Configuration (02_provider_management/01_provider_configuration) |
| `PR-06-XXX` | Privacy (06_system_behaviors/01_privacy) |

---

### 5. Learning Curve & Maintenance

| Aspect | `docs/spec/` | `docs/ears/` |
|--------|--------------|--------------|
| Learning curve | Low | Medium-High |
| Writing speed | Fast | Slower |
| Review efficiency | Good | Requires EARS knowledge |
| Maintenance overhead | Low | Medium |

---

## EARS Pattern Reference

The EARS methodology uses five distinct patterns:

### 1. Ubiquitous Software Requirements
**Purpose:** Global properties that don't require external stimulus.
**Syntax:** `The <system name> shall <system response>.`
**Example:** `The system shall support two license types: Open Source and Commercial.`

### 2. Event-Driven Software Requirements
**Purpose:** Discrete API invocations, UI interactions.
**Syntax:** `When <trigger>, the <system name> shall <system response>.`
**Example:** `When a manager enters a valid commercial license key, the system shall activate the commercial license.`

### 3. State-Driven Software Requirements
**Purpose:** Modal behaviors, session states, runtime environments.
**Syntax:** `While <system state>, the <system name> shall <system response>.`
**Example:** `While a commercial license is active, the system shall allow unlimited team creation.`

### 4. Optional Feature Software Requirements
**Purpose:** Feature flags, modular configurations (SPLE).
**Syntax:** `Where <feature is included>, the <system name> shall <system response>.`
**Example:** `Where multi-tenancy is enabled, the system shall isolate tenant data at the database level.`

### 5. Unwanted Behaviour Software Requirements
**Purpose:** Exception handling, error cases, validation failures.
**Syntax:** `If <trigger>, then the <system name> shall <system response>.`
**Example:** `If a manager enters an invalid license key, then the system shall display an error message indicating the license key is invalid.`

### 6. Complex Software Requirements
**Purpose:** Dense Boolean logic combining states, events, and exceptions.
**Syntax:** `<Multiple conditions>, the <system name> shall <system response>.`
**Example:** `When a commercial license expires while the system is running, the system shall lock advanced features on the next operation while preserving core functionality.`

---

## Maturity Assessment

### `docs/spec/` - **Level 2 (Managed)**

Characteristics:
- Structured approach with MECE organization
- Good for stakeholder communication
- Suitable for early-stage products
- Lacks formal verification support

### `docs/ears/` - **Level 4 (Quantitatively Managed)**

Characteristics:
- Industry-standard methodology (EARS is used in aerospace, automotive, medical devices)
- Formal syntax enables automated validation
- Supports safety-critical development practices
- Machine-parsable for toolchain integration
- Adv-EARS supports Context-Free Grammar parsing for automated model generation

---

## Recommendation

### For this project, adopt `docs/ears/` as the primary specification because:

1. **Multi-tenant complexity** - State-driven requirements (`While X...`) are critical for correct isolation behavior

2. **License enforcement** - EARS patterns force systematic thinking through all combinations (active/expired/no license)

3. **Security & privacy** - The unwanted behavior pattern (`If X, then...`) systematically covers error handling

4. **BDD alignment** - Gherkin scenarios map naturally to EARS patterns

5. **Scalability** - Formal structure prevents requirement entropy as product grows

### Hybrid Approach

| Use `docs/spec/` for | Use `docs/ears/` for |
|---------------------|---------------------|
| Product overviews | Functional requirements |
| User stories | Business rules |
| Stakeholder presentations | Test specifications |
| Marketing documentation | API contracts |
| Executive summaries | Compliance documentation |

---

## Summary Score

| Criterion | `docs/spec/` | `docs/ears/` |
|-----------|:------------:|:------------:|
| Precision | 6/10 | 9/10 |
| Completeness | 7/10 | 9/10 |
| Testability | 5/10 | 10/10 |
| Traceability | 6/10 | 9/10 |
| Readability | 9/10 | 7/10 |
| **Overall Maturity** | **6.6/10** | **8.8/10** |

---

## Action Items

1. [x] ~~Migrate active development to use `docs/ears/` specifications~~
2. [x] ~~Add missing Data Export requirements to `docs/ears/04_usage_insights/01_data_collection/`~~
3. [x] ~~Add AI Agent Behavior Reports to `docs/ears/05_user_interfaces/01_member_app/`~~
4. [x] ~~Add Installation, Keyboard Shortcuts, and Notifications to Member App~~
5. [x] ~~Add invitation delivery time to User Accounts~~
6. [ ] Keep `docs/spec/` for stakeholder-facing documentation
7. [ ] Update CLAUDE.md to reference `docs/ears/` as the primary requirements source
8. [ ] Train team on EARS patterns and conventions
9. [ ] Consider tooling for EARS requirement validation

---

## Gap Analysis Results (2026-03-19)

The following gaps were identified and resolved:

| Document | Gap Severity | Missing Requirements | Status |
|----------|--------------|---------------------|--------|
| **04_usage_insights/01_data_collection** | 🔴 Critical | ~15 requirements (entire Export feature) | ✅ Fixed |
| **05_user_interfaces/01_member_app** | 🟡 Moderate | ~20 requirements (Agent Reports, Shortcuts, Installation) | ✅ Fixed |
| **01_identity_and_access/01_user_accounts** | 🟢 Minor | ~1 requirement (invitation delivery time) | ✅ Fixed |

### Added Requirements Summary

**Data Collection (DC-04-701 to DC-04-1004):**
- Data export workflow (7 requirements)
- Export limits and constraints (4 requirements)
- Export error handling (3 requirements)
- Export formats and specifications (4 requirements)

**Member App (MA-05-601 to MA-05-1801):**
- AI Agent Behavior Reports KPIs (4 requirements)
- AI Agent Behavior visualizations (4 requirements)
- Date navigation for reports (2 requirements)
- Data retention for reports (2 requirements)
- Installation requirements (7 requirements)
- Auto-start and tray operation (7 requirements)
- Keyboard shortcuts (4 requirements)
- Notifications (6 requirements)

**User Accounts (UA-01-202):**
- Invitation email delivery within 30 seconds

---

## Related Documentation

- [EARS Syntax Guide](ears/spec.md)
- [EARS CLAUDE.md](ears/CLAUDE.md)
- [Spec CLAUDE.md](spec/CLAUDE.md)
- [Design Constitution](design/constitution.md)

---

**Conclusion:** `docs/ears/` is the more mature and production-ready requirements specification. The EARS methodology provides the formal structure needed for complex multi-tenant systems with licensing, security, and privacy requirements.
