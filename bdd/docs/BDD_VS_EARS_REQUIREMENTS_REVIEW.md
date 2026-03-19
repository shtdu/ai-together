# BDD Test Coverage vs EARS Requirements Review

**Review Date:** 2026-03-19
**BDD Feature Files:** 5 files (~175 scenarios)
**EARS Domains:** 6 domains with detailed requirements

## Executive Summary

The BDD test suite provides good coverage for API-level operations but has significant gaps in:

| Category | Gap Count | Priority |
|----------|-----------|----------|
| Critical (Core Functionality) | 8 | High |
| Important (Business Rules) | 6 | Medium |
| Nice-to-have (Performance) | 4 | Low |

**Overall Assessment:** BDD tests cover ~60% of EARS functional requirements.

---

## Domain 01: Identity & Access

### BDD File: `00_identity_and_access.feature` (51 scenarios)

### Coverage Analysis

| EARS Subdomain | Requirements | BDD Coverage | Status |
|----------------|--------------|--------------|--------|
| User Accounts | 25+ FRs | ~40% | ⚠️ Partial |
| Roles & Permissions | 20+ FRs | ~50% | ⚠️ Partial |
| Multi-Tenancy | 15+ FRs | ~80% | ✅ Good |
| Licensing | 20+ FRs | ~70% | ✅ Good |

### Detailed Gap Analysis

#### 1.1 User Accounts - GAPS

| Missing Requirement | EARS ID | Priority |
|---------------------|---------|----------|
| User registration workflow | UA-01-101 to UA-01-106 | **HIGH** |
| Team invitation workflow | UA-01-201 to UA-01-208 | **HIGH** |
| Password requirements validation | UA-01-001 to UA-01-006 | **MEDIUM** |
| Password reset workflow | UA-01-902 to UA-01-905 | **MEDIUM** |
| Failed login attempts (5 attempts, 15min lock) | UA-01-801 to UA-01-803 | **MEDIUM** |
| Session management (24h expiry) | UA-01-701 to UA-01-704 | **MEDIUM** |
| "Remember me" extended session | UA-01-604 | LOW |
| Invitation resend | UA-01-302 | LOW |
| Invitation expiration (7 days) | UA-01-204, UA-01-303 | LOW |

**Suggested Scenarios:**
```gherkin
Scenario: Register new organization
  When I register with email "newuser@example.com" and password "StrongPass123!"
  Then a new organization should be created
  And I should be granted the Manager role

Scenario: Account lockout after failed attempts
  Given a user exists with email "user@example.com"
  When I fail to login 5 times with wrong password
  Then the account should be locked for 15 minutes

Scenario: Password requirements validation
  When I register with password "weak"
  Then I should receive a 400 error
  And the error should indicate password requirements
```

#### 1.2 Roles & Permissions - GAPS

| Missing Requirement | EARS ID | Priority |
|---------------------|---------|----------|
| Member can view provider list (read-only) | RP-01-109 | **MEDIUM** |
| Member cannot test provider connection | RP-01-206 | **MEDIUM** |
| Last manager cannot be demoted | RP-01-1201 | **HIGH** |
| Last manager cannot be deactivated | RP-01-1202 | **HIGH** |
| Role change takes effect immediately | RP-01-1101 | **MEDIUM** |
| Member cannot view user list | RP-01-401 | **MEDIUM** |

**Suggested Scenarios:**
```gherkin
Scenario: Last manager cannot be demoted
  Given I am logged in as the only manager
  When I attempt to change my role to member
  Then I should receive a 403 error
  And the error message should contain "last manager"

Scenario: Member can view but not test providers
  Given I am logged in as a member
  When I list all providers
  Then the operation should succeed
  But when I test provider connectivity
  Then I should receive a 403 error
```

#### 1.3 Multi-Tenancy - GAPS

| Missing Requirement | EARS ID | Priority |
|---------------------|---------|----------|
| Email already in another org error | MT-01-301 | **MEDIUM** |
| User organization transition | MT-01-302, MT-01-303 | LOW |
| Organization deletion grace period | MT-01-503, MT-01-701, MT-01-702 | LOW |
| Data export cross-org prevention | MT-01-601, MT-01-602 | **MEDIUM** |

#### 1.4 Licensing - GAPS

| Missing Requirement | EARS ID | Priority |
|---------------------|---------|----------|
| Team limit enforcement (1 team for OSS) | LC-01-503 | **HIGH** |
| Commercial license required message | LC-01-501, LC-01-502 | **MEDIUM** |
| License renewal workflow | (implicit) | LOW |

**Note:** Current BDD has `@wip` on license kind limits scenario, but EARS spec says provider limits are NOT enforced.

---

## Domain 02: Provider Management

### BDD File: `01_provider_management.feature` (27 scenarios)

### Coverage Analysis

| EARS Subdomain | Requirements | BDD Coverage | Status |
|----------------|--------------|--------------|--------|
| Provider Configuration | 5 FRs | 100% | ✅ Complete |
| Request Routing | 5 FRs | 0% | ❌ Missing |
| Model Mapping | 5 FRs | 0% | ❌ Missing |
| Provider Health | 8 FRs | 0% | ❌ Missing |
| Failover & Reliability | 4 FRs | 0% | ❌ Missing |

### Critical Gaps

| Missing Requirement | EARS ID | Priority |
|---------------------|---------|----------|
| Request routing by priority | PM-02-101 | **HIGH** |
| Automatic failover on failure | PM-02-102 | **HIGH** |
| Max 3 provider attempts | PM-02-103 | **HIGH** |
| Failover notification | PM-02-104 | **MEDIUM** |
| Disabled providers not used | PM-02-105 | **MEDIUM** |
| Model name mapping | PM-02-201 to PM-02-205 | **MEDIUM** |
| Provider health marking | PM-02-301 to PM-02-304 | **MEDIUM** |
| Unhealthy provider cooldown | PM-02-401, PM-02-402 | **MEDIUM** |
| Request timeout (30s) | PM-02-502 | **MEDIUM** |

**Note:** These gaps exist because BDD tests are **API-only** (server endpoints). Request routing, failover, and model mapping are **proxy behaviors** that happen in the member client, not the server API.

**Recommendation:** Consider:
1. Adding integration tests for proxy behavior
2. Or adding API endpoints to configure/test routing behavior

---

## Domain 04: Usage Insights

### BDD File: `03_usage_insights.feature` (58 scenarios)

### Coverage Analysis

| EARS Subdomain | Requirements | BDD Coverage | Status |
|----------------|--------------|--------------|--------|
| Data Collection | 4 FRs | 50% | ⚠️ Partial |
| Personal Analytics | 3 FRs | 100% | ✅ Complete |
| Team Analytics | 3 FRs | 100% | ✅ Complete |
| Cost Tracking | 3 FRs | 100% | ✅ Complete |
| Data Retention | 3 FRs | 0% | ❌ Missing |

### Gaps

| Missing Requirement | EARS ID | Priority |
|---------------------|---------|----------|
| Offline metadata queuing | UI-04-004 | **MEDIUM** |
| 30-second batch upload | UI-04-003 | LOW |
| Data retention by license (7d/90d) | UI-04-401 to UI-04-403 | **HIGH** |
| Automatic data deletion | UI-04-403 | **MEDIUM** |

**Suggested Scenarios:**
```gherkin
Scenario: Data retention for Open Source license
  Given I have an Open Source license
  And usage data from 8 days ago
  When I query usage statistics
  Then I should not see data older than 7 days

Scenario: Data retention for Commercial license
  Given I have a Commercial license
  And usage data from 60 days ago
  When I query usage statistics
  Then I should see data from 60 days ago
```

---

## Domain 05: User Interfaces (Dashboard)

### BDD File: `04_user_interfaces.feature` (27 scenarios)

### Coverage Analysis

| EARS Subdomain | Requirements | BDD Coverage | Status |
|----------------|--------------|--------------|--------|
| Manager Dashboard | N/A (UI spec) | ~80% | ✅ Good |
| Team Management | N/A | ~90% | ✅ Good |
| User Management | N/A | ~90% | ✅ Good |

### Gaps

| Missing Area | Priority |
|--------------|----------|
| Default team cannot be deleted | Already covered (Scenario 51) |
| Member cannot create users | Already covered (@wip) |
| Member cannot manage teams | Already covered (@wip) |

**Note:** Most scenarios are well-covered. The `@wip` tags indicate incomplete step definitions, not missing scenarios.

---

## Domain 06: System Behaviors

### BDD File: `05_system_behaviors.feature` (6 scenarios)

### Coverage Analysis

| EARS Subdomain | Requirements | BDD Coverage | Status |
|----------------|--------------|--------------|--------|
| Health Checks | 2 FRs | 100% | ✅ Complete |
| System Readiness | 2 FRs | 100% | ✅ Complete |
| Performance | 3 FRs | 0% | ❌ Missing |
| Privacy | 4 FRs | 0% | ❌ Missing |
| Security | 6 FRs | 0% | ❌ Missing |
| Compliance | 4 FRs | 0% | ❌ Missing |

### Gaps

| Missing Requirement | EARS ID | Priority |
|---------------------|---------|----------|
| Request overhead < 100ms | SB-06-201 | LOW (perf) |
| Failover switch < 5s | SB-06-202 | LOW (perf) |
| Config distribution < 5min | SB-06-203 | LOW (perf) |
| Password hashing (bcrypt) | SB-06-102 | LOW (impl detail) |
| TLS 1.3 encryption | SB-06-103 | LOW (infra) |
| API key encryption at rest | SB-06-104 | LOW (impl detail) |

**Note:** Security and performance requirements are typically verified through:
- Security: Penetration testing, code review, not BDD
- Performance: Load testing, benchmarks, not BDD
- Compliance: Audit, certification, not BDD

These are appropriately NOT covered by BDD tests.

---

## Summary

### Priority 1: Critical Gaps (Add These)

| # | Gap | Domain | EARS IDs |
|---|-----|--------|----------|
| 1 | User registration workflow | 01 | UA-01-101 to 106 |
| 2 | Team invitation workflow | 01 | UA-01-201 to 208 |
| 3 | Last manager protection | 01 | RP-01-1201, 1202 |
| 4 | Team limit enforcement (OSS) | 01 | LC-01-503 |
| 5 | Request routing & failover | 02 | PM-02-101 to 105 |
| 6 | Data retention by license | 04 | UI-04-401 to 403 |
| 7 | Account lockout | 01 | UA-01-801 to 803 |
| 8 | Password reset workflow | 01 | UA-01-902 to 905 |

### Priority 2: Important Gaps (Consider Adding)

| # | Gap | Domain |
|---|-----|--------|
| 1 | Member read-only provider access | 01 |
| 2 | Password requirements validation | 01 |
| 3 | Session expiry (24h) | 01 |
| 4 | Model mapping | 02 |
| 5 | Provider health monitoring | 02 |
| 6 | Offline metadata queuing | 04 |

### Priority 3: Out of Scope (Don't Add)

- Performance timing tests (use benchmarks)
- Security implementation tests (use code review/pen testing)
- Infrastructure tests (TLS, encryption at rest)

---

## Recommendations

### 1. Add Missing Critical Scenarios

Create new feature file or extend existing:

**`00_identity_and_access.feature` - Add:**
- Registration scenarios
- Invitation scenarios
- Account lockout scenarios
- Password reset scenarios
- Last manager protection scenarios

**`01_provider_management.feature` - Note:**
- Request routing/failover requires proxy testing (out of scope for API BDD)

**`03_usage_insights.feature` - Add:**
- Data retention scenarios by license type

### 2. Review @wip Scenarios

Current `@wip` count: ~15 scenarios

These need step definitions completed:
- Member login scenarios
- RBAC permission scenarios
- License activation scenarios

### 3. Consider Separate Test Categories

| Category | Tool | Coverage |
|----------|------|----------|
| API BDD | godog | Server endpoints |
| Proxy Integration | Go test | Routing, failover, model mapping |
| Performance | benchmarks | Timing requirements |
| Security | pen testing | Security requirements |

---

## Appendix: Scenario Count by Domain

| Domain | EARS FRs | BDD Scenarios | Coverage % |
|--------|----------|---------------|------------|
| 01 Identity & Access | ~80 | 51 | ~60% |
| 02 Provider Management | ~25 | 27 | ~40%* |
| 03 Configuration Sync | ~15 | 0 | 0%** |
| 04 Usage Insights | ~15 | 58 | ~80% |
| 05 User Interfaces | N/A | 27 | N/A |
| 06 System Behaviors | ~20 | 6 | ~30% |

*Provider Management lower due to proxy behaviors not testable via API
**Configuration Sync is member client behavior, not server API

---

**Next Steps:**
1. Review and prioritize gap list with team
2. Add critical scenarios to BDD suite
3. Complete `@wip` step definitions
4. Consider proxy integration tests for routing/failover
