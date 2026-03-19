# BDD Test Gap Implementation Plan

**Created:** 2026-03-19
**Status:** Ready for Implementation
**Scope:** Server API Gaps (A-G from EARS Review)

---

## Executive Summary

This plan adds ~30 new BDD scenarios to cover critical gaps identified in the EARS requirements review. The focus is on server API functionality only—proxy behavior (request routing, failover) is out of scope.

**Estimated Effort:** 2-3 sessions
**Files Modified:** 2 feature files, 3 step definition files

---

## Phase 1: Identity & Access Gaps

**Target File:** `features/00_identity_and_access.feature`

### 1.1 User Registration Workflow (6 scenarios)

**EARS Reference:** UA-01-101 to UA-01-106

```gherkin
Rule: User Registration

  Scenario: Register new organization successfully
    Given I am not authenticated
    And I have a unique email "newuser@example.com"
    And I have a strong password "StrongPass123!"
    When I register a new account
    Then the response status code should be 201
    And a new organization should be created
    And I should receive a valid authentication token
    And I should be granted the Manager role

  Scenario: Registration creates unique organization name
    Given I am not authenticated
    And I have a unique email "orgtest@example.com"
    When I register a new account
    Then the organization name should be auto-generated

  Scenario: Registration validates password requirements
    Given I am not authenticated
    And I have a unique email "weakpass@example.com"
    And I have a weak password "short"
    When I register a new account
    Then I should receive a 400 error
    And the error message should contain "password"

  Scenario Outline: Registration password requirements
    Given I am not authenticated
    And I have a unique email "passtest@example.com"
    And I have password "<password>"
    When I register a new account
    Then I should receive a 400 error

    Examples:
      | password |
      | short |
      | nouppercase123! |
      | NOLOWERCASE123! |
      | NoNumbers! |
      | NoSpecial123 |

  Scenario: Registration rejects duplicate email
    Given a user exists with email "existing@example.com"
    And I am not authenticated
    When I register with email "existing@example.com"
    Then I should receive a 400 error
    And the error message should contain "email already exists"

  Scenario: Registration auto-logs in new user
    Given I am not authenticated
    And I have a unique email "autologin@example.com"
    And I have a strong password "StrongPass123!"
    When I register a new account
    Then I should be able to access protected endpoints
```

### 1.2 Team Invitation Workflow (8 scenarios)

**EARS Reference:** UA-01-201 to UA-01-208, UA-01-301 to UA-01-304

```gherkin
Rule: Team Invitation

  Scenario: Manager sends team invitation
    Given I am logged in as a manager
    And I have a unique invitee email "invited@example.com"
    When I send a team invitation
    Then the response status code should be 201
    And an invitation record should be created
    And the invitation should have a unique token

  Scenario: Invitation expires after 7 days
    Given I am logged in as a manager
    And there is an invitation created 8 days ago
    When the invitee attempts to accept the invitation
    Then the invitation should be expired

  Scenario: Accept invitation with new user
    Given I am not authenticated
    And there is a pending invitation for "newuser@example.com"
    When I accept the invitation with password "NewUser123!"
    Then a new user should be created
    And the user should have the Member role
    And I should receive a valid authentication token

  Scenario: Accept invitation sets up password
    Given I am not authenticated
    And there is a pending invitation for "setup@example.com"
    When I accept the invitation with password "SetupPass123!"
    Then the user password should be set
    And the user should be able to login

  Scenario: Manager cancels pending invitation
    Given I am logged in as a manager
    And there is a pending invitation for "cancel@example.com"
    When I cancel the invitation
    Then the invitation should be invalidated
    And the invitee cannot accept the invitation

  Scenario: Duplicate email in organization
    Given I am logged in as a manager
    And a user exists with email "duplicate@example.com"
    When I invite "duplicate@example.com"
    Then I should receive a 400 error
    And the error message should contain "already exists"

  Scenario: Resend invitation email
    Given I am logged in as a manager
    And there is a pending invitation for "resend@example.com"
    When I resend the invitation
    Then the response status code should be 200
    And the invitation token should remain valid

  Scenario: Member cannot send invitations
    Given I am logged in as a member
    When I attempt to send a team invitation
    Then I should receive a 403 error
```

### 1.3 Account Lockout (4 scenarios)

**EARS Reference:** UA-01-801 to UA-01-803

```gherkin
Rule: Account Lockout

  Scenario: Account locked after 5 failed attempts
    Given a user exists with email "locktest@example.com"
    When I fail to login 5 times with wrong password
    Then the account should be locked
    And I should receive an "account_locked" error

  Scenario: Locked account cannot login with correct password
    Given a user exists with email "locked@example.com"
    And the account is locked
    When I login with correct credentials
    Then I should receive an "account_locked" error
    And the response status code should be 401

  Scenario: Account unlocks after 15 minutes
    Given a user exists with email "timed@example.com"
    And the account was locked 16 minutes ago
    When I login with correct credentials
    Then I should receive a valid authentication token

  Scenario: Successful login resets failed attempt counter
    Given a user exists with email "reset@example.com"
    And I have failed to login 4 times
    When I login with correct credentials
    Then the failed attempt counter should be reset to 0
```

### 1.4 Password Reset Workflow (4 scenarios)

**EARS Reference:** UA-01-902 to UA-01-905

```gherkin
Rule: Password Reset

  Scenario: Request password reset email
    Given a user exists with email "reset@example.com"
    When I request a password reset
    Then the response status code should be 200
    And a reset token should be generated
    And the reset token should expire in 1 hour

  Scenario: Reset password with valid token
    Given a user exists with email "validreset@example.com"
    And I have a valid reset token
    When I reset password to "NewPassword123!"
    Then the response status code should be 200
    And I can login with the new password

  Scenario: Reset password with expired token
    Given a user exists with email "expired@example.com"
    And I have an expired reset token
    When I attempt to reset password
    Then I should receive a 400 error
    And the error message should contain "expired"

  Scenario: Reset password with invalid token
    Given I have an invalid reset token
    When I attempt to reset password
    Then I should receive a 400 error
    And the error message should contain "invalid"
```

### 1.5 Last Manager Protection (3 scenarios)

**EARS Reference:** RP-01-1201, RP-01-1202, UA-01-501, UA-01-502

```gherkin
Rule: Last Manager Protection

  Scenario: Last manager cannot demote themselves
    Given I am logged in as the only manager
    When I attempt to change my role to member
    Then I should receive a 403 error
    And the error message should contain "last manager"

  Scenario: Last manager cannot be deactivated by another manager
    Given I am logged in as a manager
    And there is only one other manager
    When I attempt to deactivate the other manager
    Then I should receive a 403 error
    And the error message should contain "last manager"

  Scenario: Manager can be demoted when other managers exist
    Given I am logged in as a manager
    And there are at least 2 managers in the organization
    When I demote another manager to member
    Then the operation should succeed
    And the user should have the Member role
```

---

## Phase 2: Licensing & Usage Gaps

### 2.1 Team Limit Enforcement (3 scenarios)

**EARS Reference:** LC-01-503, LC-01-406

**Target File:** `features/00_identity_and_access.feature`

```gherkin
Rule: Team Limit Enforcement

  Scenario: Open Source license allows only 1 team
    Given I have an Open Source license
    And I have created 1 team
    When I attempt to create another team
    Then I should receive a 403 error
    And the error message should contain "team limit"

  Scenario: Commercial license allows unlimited teams
    Given I have a Commercial license
    And I have created 5 teams
    When I create another team
    Then the operation should succeed

  Scenario: Expired Commercial license reverts to 1 team limit
    Given I have an expired Commercial license
    And I have 3 existing teams
    When I attempt to create a new team
    Then I should receive a 403 error
    And the error message should contain "team limit"
```

### 2.2 Data Retention by License (5 scenarios)

**EARS Reference:** UI-04-401 to UI-04-403

**Target File:** `features/03_usage_insights.feature`

```gherkin
Rule: Data Retention by License

  Scenario: Open Source license retains data for 7 days
    Given I have an Open Source license
    And I have usage data from 8 days ago
    When I query usage statistics
    Then I should not see data older than 7 days

  Scenario: Commercial license retains data for 90 days
    Given I have a Commercial license
    And I have usage data from 60 days ago
    When I query usage statistics
    Then I should see data from 60 days ago

  Scenario: Data is automatically deleted after retention period
    Given I have an Open Source license
    And I have usage data from 10 days ago
    When the retention cleanup job runs
    Then the old data should be deleted

  Scenario: License upgrade extends retention period
    Given I have usage data from 30 days ago
    And I upgrade from Open Source to Commercial license
    When I query usage statistics
    Then I should see data from 30 days ago

  Scenario: License downgrade does not delete existing data immediately
    Given I have a Commercial license
    And I have usage data from 60 days ago
    When I downgrade to Open Source license
    Then the existing data should be retained
    But new data should follow 7-day retention
```

---

## Phase 3: Step Definitions

### 3.1 New Steps in `auth_steps.go`

| Step Pattern | Used By |
|--------------|---------|
| `^I register a new account$` | Registration |
| `^I have a strong password "([^"]*)"$` | Registration |
| `^I have a weak password "([^"]*)"$` | Registration |
| `^I have password "([^"]*)"$` | Password validation |
| `^a new organization should be created$` | Registration |
| `^I should be granted the Manager role$` | Registration |
| `^the failed attempt counter should be reset to 0$` | Lockout |
| `^I should be able to access protected endpoints$` | Registration |

### 3.2 New Steps in `invitation_steps.go` (New File)

| Step Pattern | Used By |
|--------------|---------|
| `^I have a unique invitee email "([^"]*)"$` | Invitation |
| `^I send a team invitation$` | Invitation |
| `^an invitation record should be created$` | Invitation |
| `^there is a pending invitation for "([^"]*)"$` | Invitation |
| `^I accept the invitation with password "([^"]*)"$` | Invitation |
| `^I cancel the invitation$` | Invitation |
| `^I resend the invitation$` | Invitation |
| `^the invitation should be expired$` | Invitation |
| `^there is an invitation created (\d+) days? ago$` | Invitation |

### 3.3 New Steps in `license_steps.go`

| Step Pattern | Used By |
|--------------|---------|
| `^I have an Open Source license$` | Team limits, Retention |
| `^I have a Commercial license$` | Team limits, Retention |
| `^I have an expired Commercial license$` | Team limits |
| `^I upgrade from Open Source to Commercial license$` | Retention |
| `^I downgrade to Open Source license$` | Retention |

### 3.4 New Steps in `usage_steps.go`

| Step Pattern | Used By |
|--------------|---------|
| `^I have usage data from (\d+) days? ago$` | Retention |
| `^I query usage statistics$` | Retention |
| `^I should (not )?see data (older than|from) (\d+) days? ago$` | Retention |
| `^the retention cleanup job runs$` | Retention |
| `^the old data should be deleted$` | Retention |

### 3.5 New Steps in `user_steps.go` (Extend existing)

| Step Pattern | Used By |
|--------------|---------|
| `^I am logged in as the only manager$` | Last manager |
| `^there is only one other manager$` | Last manager |
| `^there are at least (\d+) managers in the organization$` | Last manager |
| `^I attempt to change my role to member$` | Last manager |
| `^I demote another manager to member$` | Last manager |
| `^the account is locked$` | Lockout |
| `^the account was locked (\d+) minutes? ago$` | Lockout |
| `^I fail to login (\d+) times with wrong password$` | Lockout |
| `^I have failed to login (\d+) times$` | Lockout |

### 3.6 New Steps in `password_reset_steps.go` (New File)

| Step Pattern | Used By |
|--------------|---------|
| `^I request a password reset$` | Password reset |
| `^a reset token should be generated$` | Password reset |
| `^I have a valid reset token$` | Password reset |
| `^I have an expired reset token$` | Password reset |
| `^I have an invalid reset token$` | Password reset |
| `^I reset password to "([^"]*)"$` | Password reset |
| `^I attempt to reset password$` | Password reset |
| `^I can login with the new password$` | Password reset |

---

## Phase 4: Implementation Sequence

### Week 1: Core Authentication Gaps

| Task | File | Est. Time |
|------|------|-----------|
| 1. Add registration scenarios | `00_identity_and_access.feature` | 30 min |
| 2. Implement registration steps | `auth_steps.go` | 1 hr |
| 3. Add account lockout scenarios | `00_identity_and_access.feature` | 20 min |
| 4. Implement lockout steps | `auth_steps.go` | 45 min |
| 5. Add password reset scenarios | `00_identity_and_access.feature` | 20 min |
| 6. Create `password_reset_steps.go` | New file | 1 hr |
| 7. Run tests and fix issues | All | 1 hr |

### Week 2: Team Management Gaps

| Task | File | Est. Time |
|------|------|-----------|
| 1. Add invitation scenarios | `00_identity_and_access.feature` | 30 min |
| 2. Create `invitation_steps.go` | New file | 1.5 hr |
| 3. Add last manager scenarios | `00_identity_and_access.feature` | 20 min |
| 4. Implement last manager steps | `user_steps.go` | 45 min |
| 5. Add team limit scenarios | `00_identity_and_access.feature` | 15 min |
| 6. Implement team limit steps | `license_steps.go` | 45 min |
| 7. Run tests and fix issues | All | 1 hr |

### Week 3: Usage & Retention Gaps

| Task | File | Est. Time |
|------|------|-----------|
| 1. Add retention scenarios | `03_usage_insights.feature` | 30 min |
| 2. Implement retention steps | `usage_steps.go` | 1 hr |
| 3. Register all new steps | `godog_suite_test.go` | 15 min |
| 4. Full test suite run | All | 30 min |
| 5. Documentation update | `CLAUDE.md`, `README.md` | 30 min |
| 6. Final review and cleanup | All | 1 hr |

---

## Phase 5: Server API Prerequisites

Before implementing certain scenarios, verify these server endpoints exist:

### Required Endpoints

| Scenario Group | Endpoint | Method | Status |
|----------------|----------|--------|--------|
| Registration | `/api/v1/auth/register` | POST | ⚠️ Verify |
| Invitations | `/api/v1/invitations` | POST, DELETE | ⚠️ Verify |
| Password Reset | `/api/v1/auth/reset-password` | POST | ⚠️ Verify |
| Account Lockout | (Built into login) | - | ⚠️ Verify |
| Last Manager | (Business logic in user update) | - | ⚠️ Verify |
| Team Limits | (Business logic in team create) | - | ⚠️ Verify |
| Data Retention | (Background job) | - | ⚠️ Verify |

### Action Items

1. **Audit server code** for each endpoint/feature
2. **Mark scenarios as `@wip`** if API not yet implemented
3. **Document API gaps** in `SERVER_ACTION_ITEMS.md`
4. **Prioritize server work** alongside BDD implementation

---

## Success Criteria

### Completion Checklist

- [ ] All 30 scenarios added to feature files
- [ ] All step definitions implemented
- [ ] All new steps registered in `godog_suite_test.go`
- [ ] `make bdd-test` passes (or failures are documented `@wip`)
- [ ] `bdd/CLAUDE.md` updated with new scenarios
- [ ] `bdd/README.md` updated with coverage stats
- [ ] Server API gaps documented

### Quality Gates

- [ ] Each scenario follows Gherkin best practices
- [ ] Scenario names clearly describe the behavior
- [ ] Step definitions are reusable across scenarios
- [ ] No duplicate step patterns
- [ ] Error messages are asserted for negative scenarios

---

## File Changes Summary

| File | Action | Lines Added |
|------|--------|-------------|
| `features/00_identity_and_access.feature` | Extend | +120 |
| `features/03_usage_insights.feature` | Extend | +35 |
| `step_definitions/auth_steps.go` | Extend | +50 |
| `step_definitions/invitation_steps.go` | **Create** | +80 |
| `step_definitions/password_reset_steps.go` | **Create** | +60 |
| `step_definitions/license_steps.go` | Extend | +30 |
| `step_definitions/usage_steps.go` | Extend | +40 |
| `step_definitions/user_steps.go` | Extend | +40 |
| `godog/godog_suite_test.go` | Extend | +10 |
| `CLAUDE.md` | Update | +20 |
| `README.md` | Update | +10 |

**Total:** ~495 lines added across 11 files

---

## Appendix: EARS Requirement Traceability

| EARS ID | Requirement | BDD Scenario |
|---------|-------------|--------------|
| UA-01-101 | Registration form display | `Register new organization successfully` |
| UA-01-102 | Password validation | `Registration password requirements` |
| UA-01-103 | Organization creation | `Register new organization successfully` |
| UA-01-104 | Manager role assignment | `I should be granted the Manager role` |
| UA-01-105 | Auto-login after registration | `Registration auto-logs in new user` |
| UA-01-201 | Send invitation email | `Manager sends team invitation` |
| UA-01-204 | Invitation expires 7 days | `Invitation expires after 7 days` |
| UA-01-205 | Password setup form | `Accept invitation with new user` |
| UA-01-206 | Member role assignment | `Accept invitation with new user` |
| UA-01-801 | Lock after 5 attempts | `Account locked after 5 failed attempts` |
| UA-01-802 | Lock message | `Locked account cannot login` |
| UA-01-803 | Reset counter on success | `Successful login resets counter` |
| UA-01-902 | Reset email sent | `Request password reset email` |
| UA-01-903 | Reset expires 1 hour | `Reset password with expired token` |
| UA-01-904 | Reset form display | `Reset password with valid token` |
| RP-01-1201 | Last manager no demote | `Last manager cannot demote themselves` |
| RP-01-1202 | Last manager no deactivate | `Last manager cannot be deactivated` |
| LC-01-406 | OSS 1 team limit | `Open Source license allows only 1 team` |
| LC-01-503 | OSS team limit error | `Open Source license allows only 1 team` |
| UI-04-401 | OSS 7-day retention | `Open Source license retains data for 7 days` |
| UI-04-402 | Commercial 90-day retention | `Commercial license retains data for 90 days` |
| UI-04-403 | Auto-delete expired data | `Data is automatically deleted after retention` |

---

## Related Documents

- [BDD vs EARS Requirements Review](./BDD_VS_EARS_REQUIREMENTS_REVIEW.md)
- [EARS Requirements](../../docs/ears/)
- [Server Action Items](../SERVER_ACTION_ITEMS.md)
