# BDD Deduplication Matrix - 00_identity_and_access.feature

**Current Status**: 345 scenarios
**Target**: ~150-200 scenarios (40-50% reduction)

## Summary of Actions

| Action | Count | Description |
|--------|-------|-------------|
| KEEP | ~120 | Canonical scenarios representing unique requirements |
| MERGE | ~80 | Convert to Scenario Outlines or combine duplicates |
| DELETE | ~145 | Coverage push, repeat tests, true duplicates |

---

## Section: AUTHENTICATION (Login)

### Login Happy Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Login with valid credentials as manager | IA-01-001 | KEEP | Canonical happy path |
| Login with valid credentials as member | IA-01-002 | MERGE → Outline | Convert to parameterized test |
| Login returns valid access token | IA-01-047 | DELETE | Duplicate of IA-01-001 |
| Login returns user profile data | IA-01-048 | DELETE | Covered by IA-01-001 |
| Login with correct email and password | IA-01-068 | DELETE | Duplicate |
| Login returns both access and refresh tokens | IA-01-069 | MERGE → IA-01-001 | Add assertion to canonical |
| Login success returns status 200 | IA-01-083 | DELETE | Covered by IA-01-001 |
| Login with valid credentials returns tokens | IA-01-055 | DELETE | Duplicate |
| Login as manager and verify response | IA-01-138 | DELETE | Duplicate |
| Login as member and verify response | IA-01-139 | DELETE | Duplicate |
| Complete login flow | IA-01-143 | DELETE | Covered by other tests |
| Member login and profile access | IA-01-144 | DELETE | Covered by profile tests |
| Login with valid credentials returns tokens | IA-01-104 | DELETE | Duplicate |
| Login returns proper response structure | IA-01-150 | DELETE | Duplicate |

### Login Error Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Login with invalid email | IA-01-003 | KEEP | User not found |
| Login with invalid password | IA-01-004 | KEEP | Wrong password |
| Login non-existent user | IA-01-005 | DELETE | Duplicate of IA-01-003 |
| Login with empty credentials | IA-01-039 | DELETE | Edge case, low value |
| Login with missing password fails | IA-01-090 | DELETE | Validation edge case |
| Login with non-existent email fails | IA-01-088 | DELETE | Duplicate |
| Login with wrong password fails | IA-01-089 | DELETE | Duplicate |

### Login Edge Cases
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Login with uppercase email | IA-01-070 | KEEP | Case insensitivity test |
| Login validates both email and password | IA-01-042 | DELETE | Covered by happy path |
| Login multiple times | IA-01-140 | DELETE | Repeat test |
| Login five times consecutively | IA-01-157 | DELETE | Repeat test |
| Multiple successful logins | IA-01-153 | DELETE | Repeat test |

---

## Section: REFRESH TOKEN

### Refresh Happy Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Refresh valid authentication token | IA-01-007 | KEEP | Canonical happy path |
| Refresh token shortly before expiration | IA-01-010 | KEEP | Important edge case |
| Refresh token returns new access token | IA-01-020 | DELETE | Duplicate |
| Refresh token includes user info in response | IA-01-021 | DELETE | Covered by canonical |
| Refresh token generates new access token | IA-01-058 | DELETE | Duplicate |
| Refresh token returns new tokens | IA-01-079 | DELETE | Duplicate |
| Refresh token returns new access token | IA-01-091 | DELETE | Duplicate |
| Refresh token returns user data | IA-01-092 | DELETE | Duplicate |
| Refresh token success status | IA-01-093 | DELETE | Covered by canonical |
| Refresh with valid token returns new tokens | IA-01-126 | DELETE | Duplicate |
| Refresh response contains user info | IA-01-127 | DELETE | Duplicate |
| Refresh after token expiration | IA-01-125 | DELETE | Edge case |
| Refresh token returns new access token (reprise) | IA-01-058 | DELETE | Duplicate |

### Refresh Error Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Refresh invalid authentication token | IA-01-008 | KEEP | Invalid token error |
| Refresh expired authentication token | IA-01-009 | KEEP | Expired token error |
| Refresh token without authentication | IA-01-012 | KEEP | Missing token |
| Refresh token with malformed token | IA-01-013 | DELETE | Covered by IA-01-008 |
| Refresh with invalid refresh token | IA-01-061 | DELETE | Duplicate of IA-01-008 |
| Refresh with empty refresh token | IA-01-062 | DELETE | Covered by IA-01-012 |
| Refresh with malformed JWT | IA-01-063 | DELETE | Covered by IA-01-008 |
| Refresh with expired token | IA-01-064 | DELETE | Duplicate of IA-01-009 |
| Refresh with empty token fails | IA-01-094 | DELETE | Duplicate |
| Refresh with invalid format fails | IA-01-095 | DELETE | Duplicate |
| Refresh with very long token fails | IA-01-096 | DELETE | Edge case, delete |
| Refresh with special characters fails | IA-01-097 | DELETE | Edge case, delete |
| Refresh with invalid JSON fails | IA-01-098 | KEEP | JSON validation |
| Refresh with empty request fails | IA-01-099 | DELETE | Covered by IA-01-098 |
| Refresh with empty refresh token | IA-01-116 | DELETE | Duplicate |
| Refresh with null refresh token | IA-01-117 | DELETE | Duplicate |
| Refresh with malformed JWT structure | IA-01-118 | DELETE | Duplicate |
| Refresh with corrupted signature | IA-01-119 | DELETE | Duplicate |
| Refresh without refresh_token field | IA-01-120 | DELETE | Duplicate |
| Refresh with invalid JSON payload | IA-01-121 | DELETE | Duplicate |
| Refresh token with malformed token | IA-01-033 | DELETE | Duplicate |

### Refresh Request Validation
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Refresh token with empty request body | IA-01-015 | MERGE → IA-01-098 | Combine JSON validation |
| Refresh token with malformed JSON | IA-01-032 | MERGE → IA-01-098 | Combine JSON validation |
| Refresh token with missing refresh_token field | IA-01-034 | DELETE | Covered by IA-01-098 |
| Refresh token with empty refresh_token value | IA-01-034 | DELETE | Duplicate |
| Refresh token with expired user context | IA-01-034 | KEEP | Edge case - deleted user |

### Refresh Repeat Tests (DELETE ALL)
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Refresh token multiple times | IA-01-011 | DELETE | Coverage push |
| Refresh token multiple times consecutively | IA-01-045 | DELETE | Repeat test |
| Refresh token multiple times consecutively (2) | IA-01-159 | DELETE | Repeat test |
| Refresh token multiple times consecutively (3) | IA-01-168 | DELETE | Repeat test |
| Refresh token multiple times consecutively (4) | IA-01-189 | DELETE | Repeat test |
| All refresh operations | IA-01-196 | DELETE | Coverage push |
| Refresh repeated operations | IA-01-202 | DELETE | Coverage push |
| Auth refresh multiple | IA-01-189 | DELETE | Coverage push |
| Multiple refresh operations (various) | IA-01-xxx | DELETE | All repeat tests |

---

## Section: VERIFY TOKEN

### Verify Happy Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Verify valid authentication token | IA-01-014 | KEEP | Canonical happy path |
| Verify token returns user profile | IA-01-029 | DELETE | Duplicate |
| Verify token returns user information | IA-01-053 | DELETE | Duplicate |
| Verify token with valid access token | IA-01-049 | DELETE | Duplicate |
| Verify token with valid token succeeds | IA-01-084 | DELETE | Duplicate |
| Verify token returns valid response | IA-01-044 | DELETE | Duplicate |
| Verify token without authentication fails | IA-01-076 | DELETE | Covered by error paths |
| Verify token success returns user info | IA-01-105 | DELETE | Duplicate |
| Verify endpoint validates token | IA-01-151 | DELETE | Duplicate |

### Verify Error Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Verify invalid authentication token | IA-01-011 | KEEP | Invalid token |
| Verify expired authentication token | IA-01-012 | KEEP | Expired token |
| Verify token without authentication | IA-01-030 | DELETE | Covered by IA-01-011 |
| Verify with invalid token format | IA-01-031 | DELETE | Covered by IA-01-011 |
| Verify token with empty string | IA-01-032 | DELETE | Covered by IA-01-011 |
| Verify token with whitespace only | IA-01-033 | DELETE | Edge case |
| Verify token with empty token | IA-01-051 | DELETE | Duplicate |
| Verify token with malformed JWT | IA-01-052 | DELETE | Covered by IA-01-011 |
| Verify token with invalid token format | IA-01-077 | DELETE | Duplicate |

### Verify Repeat Tests (DELETE ALL)
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Multiple verify requests succeed | IA-01-054 | DELETE | Repeat test |
| Verify token three times | IA-01-160 | DELETE | Repeat test |
| Verify token multiple times succeeds | IA-01-078 | DELETE | Repeat test |
| All verify operations | IA-01-197 | DELETE | Coverage push |

---

## Section: LOGOUT

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Logout successfully | IA-01-006 | KEEP | Only logout test |

---

## Section: GET USER PROFILE

### Profile Happy Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Get user profile as authenticated manager | IA-01-016 | KEEP | Manager canonical |
| Get user profile as authenticated member | IA-01-017 | KEEP | Member canonical |
| Get own profile as manager | IA-01-027 | DELETE | Duplicate of IA-01-016 |
| Get own profile as member | IA-01-028 | DELETE | Duplicate of IA-01-017 |
| Get profile as manager | IA-01-056 | DELETE | Duplicate |
| Get profile as member | IA-01-057 | DELETE | Duplicate |
| Get profile returns user data | IA-01-046 | DELETE | Duplicate |
| Get profile after successful login | IA-01-050 | DELETE | Duplicate |
| Profile access returns consistent data | IA-01-217 | DELETE | Duplicate |
| Profile contains user ID | IA-01-218 | DELETE | Covered by field validation |
| Profile contains correct fields | IA-01-030 | DELETE | Covered by field validation |
| Profile has correct tenant ID | IA-01-031 | DELETE | Covered by field validation |
| Profile contains created timestamp | IA-01-220 | DELETE | Covered by field validation |
| Get profile returns user ID | IA-01-085 | DELETE | Duplicate |
| Get profile returns name | IA-01-086 | DELETE | Duplicate |
| Get profile returns role | IA-01-087 | DELETE | Duplicate |
| Get profile contains email | IA-01-108 | DELETE | Duplicate |
| Get profile contains name | IA-01-109 | DELETE | Duplicate |
| Get profile contains role | IA-01-110 | DELETE | Duplicate |
| Get profile contains tenant ID | IA-01-111 | DELETE | Duplicate |
| Get profile contains tenant ID (2) | IA-01-065 | DELETE | Duplicate |
| Get profile contains name field | IA-01-066 | DELETE | Duplicate |
| Get profile returns all fields | IA-01-096 | DELETE | Duplicate |
| Get profile as member (2) | IA-01-097 | DELETE | Duplicate |
| Get profile returns all required fields | IA-01-106 | DELETE | Duplicate |
| Get profile multiple times | IA-01-098 | DELETE | Repeat test |
| Profile contains creation timestamp | IA-01-099 | DELETE | Duplicate |
| Profile contains last update timestamp | IA-01-100 | DELETE | Duplicate |
| Get profile multiple times consecutively | IA-01-131 | DELETE | Repeat test |
| Get profile as member (3) | IA-01-132 | DELETE | Duplicate |
| Get profile contains user ID (2) | IA-01-133 | DELETE | Duplicate |
| Get profile multiple times in sequence | IA-01-134 | DELETE | Repeat test |
| Login and then get profile | IA-01-135 | DELETE | Duplicate |
| Profile contains all expected fields | IA-01-223 | DELETE | Duplicate |
| Profile access before and after refresh | IA-01-225 | DELETE | Duplicate |

### Profile Field Validation (MERGE INTO ONE)
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Get profile includes created and updated timestamps | IA-01-029 | MERGE | Create comprehensive field test |
| Get profile includes team information | IA-01-030 | MERGE | Create comprehensive field test |
| Profile contains all required data | IA-01-231 | MERGE | Create comprehensive field test |

### Profile Error Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Get profile without authentication | IA-01-018 | KEEP | Unauthorized |
| Get profile with invalid token | IA-01-019 | DELETE | Covered by IA-01-018 |
| Get profile without authentication (2) | IA-01-031 | DELETE | Duplicate |
| Get profile without authentication fails | IA-01-067 | DELETE | Duplicate |
| Get profile without authentication (3) | IA-01-122 | KEEP | Duplicate - keep as variant |
| Get profile with invalid token (2) | IA-01-123 | DELETE | Duplicate |
| Get profile with expired token | IA-01-124 | DELETE | Covered by IA-01-018 |
| Profile access with expired token | IA-01-219 | DELETE | Duplicate |

### Profile Repeat Tests (DELETE ALL - DOZENS OF THESE)
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Get profile five times | IA-01-158 | DELETE | Repeat test |
| Multiple profile requests (many variations) | IA-01-xxx | DELETE | All repeat tests |
| Profile access multiple times | IA-01-221 | DELETE | Repeat test (5x) |
| Profile access as member multiple times | IA-01-222 | DELETE | Repeat test (3x) |
| Multiple profile requests (2) | IA-01-173 | DELETE | Repeat test |
| Multiple profile checks (many) | IA-01-xxx | DELETE | All repeat tests |
| Profile repeated access | IA-01-201 | DELETE | Repeat test (7x) |
| Profile access repeated many times | IA-01-230 | DELETE | Repeat test (10x) |
| Profile verify profile sequence | IA-01-227 | DELETE | Repeat test |
| Profile verify refresh profile sequence | IA-01-229 | DELETE | Repeat test |
| Profile access after multiple refreshes | IA-01-232 | DELETE | Repeat test |
| All other "get profile N times" tests | IA-01-xxx | DELETE | DELETE ALL |

---

## Section: REGISTRATION

### Registration Happy Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Register new organization successfully | IA-01-017 | KEEP | Canonical happy path |
| Registration creates unique organization name | IA-01-018 | KEEP | Organization name generation |
| Registration auto-logs in new user | IA-01-022 | DELETE | Covered by IA-01-017 |
| Registration with minimal valid data | IA-01-212 | DELETE | Covered by IA-01-017 |
| Register with valid email and password | IA-01-071 | DELETE | Duplicate |
| Register with valid email and strong password | IA-01-072 | DELETE | Duplicate |
| Register returns user profile data | IA-01-073 | DELETE | Covered by IA-01-017 |
| Register with different valid email | IA-01-112 | DELETE | Duplicate |
| Register returns success with user data | IA-01-113 | DELETE | Duplicate |
| Register with valid data succeeds | IA-01-107 | DELETE | Duplicate |
| Register with valid password succeeds | IA-01-114 | DELETE | Duplicate |
| Register creates user account | IA-01-115 | DELETE | Duplicate |
| Register requires email and password | IA-01-043 | DELETE | Covered by validation |
| Register with valid data creates user | IA-01-059 | DELETE | Duplicate |
| Register with lowercase email | IA-01-074 | DELETE | Edge case |
| Register with mixed case email | IA-01-075 | DELETE | Edge case |
| Multiple registration attempts | IA-01-213 | DELETE | Repeat test |

### Registration Password Validation
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Registration validates password requirements | IA-01-019 | DELETE | Covered by outline |
| Registration password requirements (Outline) | IA-01-020 | KEEP | Already an outline |
| Registration with very short password | IA-01-128 | DELETE | Covered by outline |
| Registration password too short | IA-01-214 | DELETE | Covered by outline |
| Registration password missing number | IA-01-215 | DELETE | Covered by outline |
| Registration password missing special char | IA-01-216 | DELETE | Covered by outline |

### Registration Error Paths
| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Registration rejects duplicate email | IA-01-021 | KEEP | Duplicate check |
| Registration with duplicate email fails | IA-01-211 | DELETE | Duplicate |
| Registration with empty email | IA-01-025 | KEEP | Validation |
| Registration with invalid email format | IA-01-026 | KEEP | Validation |
| Registration with empty password | IA-01-027 | DELETE | Covered by password outline |
| Registration with missing name | IA-01-028 | KEEP | Validation |
| Registration with missing name (2) | IA-01-130 | DELETE | Duplicate |
| Registration with very long email | IA-01-037 | DELETE | Edge case |
| Registration with email without domain | IA-01-038 | DELETE | Covered by IA-01-026 |
| Register without email field | IA-01-129 | DELETE | Covered by IA-01-025 |

---

## Section: PASSWORD RESET

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Request password reset email | IA-01-023 | KEEP | Happy path |
| Reset password with valid token | IA-01-024 | KEEP | Happy path |
| Reset password with expired token | IA-01-025 | KEEP | Error path |
| Reset password with invalid token | IA-01-026 | KEEP | Error path |

---

## Section: ACCOUNT LOCKOUT

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Account locked after 5 failed attempts | IA-01-013 | KEEP | Core behavior |
| Locked account cannot login with correct password | IA-01-014 | KEEP | Core behavior |
| Account unlocks after 15 minutes | IA-01-015 | KEEP | Time-based behavior |
| Successful login resets failed attempt counter | IA-01-016 | KEEP | Reset behavior |

---

## Section: ROLE-BASED ACCESS CONTROL

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Manager can create provider | IA-02-001 | KEEP | RBAC test |
| Member cannot create provider | IA-02-002 | KEEP | RBAC test |
| Manager can view team analytics | IA-02-003 | KEEP | RBAC test |
| Manager can manage users | IA-02-004 | KEEP | RBAC test |
| Member cannot manage users | IA-02-005 | KEEP | RBAC test |
| Manager can delete provider | IA-02-006 | KEEP | RBAC test |
| Member can view own usage | IA-02-007 | KEEP | RBAC test |
| Manager can update provider | IA-02-008 | KEEP | RBAC test |
| Member cannot update provider | IA-02-009 | KEEP | RBAC test |

---

## Section: MULTI-TENANT ISOLATION

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Users from different tenants cannot access each other's data | IA-03-001 | KEEP (p0) | Critical isolation |
| Manager can only see users from own tenant | IA-03-002 | KEEP | Isolation test |
| Profile contains correct tenant ID | IA-03-003 | DELETE | Covered by profile tests |
| Resources are isolated by tenant | IA-03-004 | KEEP | Isolation test |

---

## Section: LICENSE ACTIVATION

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Activate commercial license successfully | IA-04-001 | KEEP | Happy path |
| Activate open-source license successfully | IA-04-002 | KEEP | Happy path |
| Activate license with signature (Outline) | IA-04-003 | KEEP | Already outline |
| Activate license with invalid signature | IA-04-004 | KEEP | Error path |
| Activate license with expired signature | IA-04-005 | KEEP | Error path |
| Activate license without authentication | IA-04-006 | KEEP | Auth check |
| Non-manager cannot activate license | IA-04-007 | KEEP | RBAC check |
| Activate Open Source license successfully | IA-04-019 | DELETE | Duplicate of IA-04-002 |
| Activate Commercial license successfully | IA-04-020 | DELETE | Duplicate of IA-04-001 |
| Activate expired license fails | IA-04-021 | KEEP | Error path |
| Activate invalid license fails | IA-04-022 | DELETE | Covered by IA-04-004 |
| License with immediate expiry | IA-04-023 | DELETE | Edge case |

---

## Section: LICENSE FEATURES

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Commercial license enables provider management | IA-04-008 | KEEP | Feature test |
| Trial license has limited features | IA-04-009 | KEEP | Feature test |
| Enterprise license enables all features | IA-04-010 | KEEP | Feature test |
| License feature flags are correct (Outline) | IA-04-011 | KEEP | Already outline |
| Check license status returns correct information | IA-04-012 | DELETE | Covered by license status |
| Open Source license feature flags | IA-04-030 | DELETE | Covered by outline |
| Commercial license feature flags | IA-04-031 | DELETE | Covered by outline |

---

## Section: LICENSE STATUS

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Get license information | IA-04-022 | KEEP | Happy path |
| License status includes limits | IA-04-023 | DELETE | Covered by field validation |
| License status includes features | IA-04-024 | DELETE | Covered by feature tests |
| Check license without authentication | IA-04-025 | KEEP | Auth check |
| Manager can view license information | IA-04-026 | DELETE | Covered by IA-04-022 |
| Member can view license information | IA-04-027 | KEEP | RBAC test |
| Get current license status | IA-04-027 | DELETE | Duplicate |
| License status includes usage statistics | IA-04-028 | DELETE | Covered by field validation |
| License status for non-admin users | IA-04-029 | DELETE | Duplicate of IA-04-027 |

---

## Section: LICENSE LIMITS

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| License kind limits | IA-04-013 | DELETE | Per spec note - NOT enforced |
| Open Source license allows only 1 team | IA-04-014 | KEEP | Team limit |
| Commercial license allows unlimited teams | IA-04-015 | KEEP | Team limit |
| Expired Commercial license reverts to 1 team limit | IA-04-016 | KEEP | Team limit |
| Seat limit is informational only | IA-04-032 | KEEP | Explicit test |
| Team limit enforcement by Open Source license | IA-04-033 | DELETE | Duplicate of IA-04-014 |
| Commercial license has no team limit | IA-04-034 | DELETE | Duplicate of IA-04-015 |

---

## Section: LICENSE EXPIRATION

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| License expires after end date | IA-04-017 | DELETE | Coverage test |
| License expiration warnings (Outline) | IA-04-018 | KEEP | Already outline |
| Renew expired license | IA-04-019 | KEEP | Renewal test |
| License grace period after expiration | IA-04-020 | KEEP | Grace period |
| License suspended after grace period | IA-04-021 | KEEP | Suspension |
| License expiration date calculation | IA-04-035 | DELETE | Coverage test |
| License expiration warning levels | IA-04-036 | DELETE | Covered by outline |
| Expired license behavior | IA-04-037 | DELETE | Covered by other tests |

---

## Section: LICENSE UPGRADES/DOWNGRADES

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Upgrade from Open Source to Commercial license | IA-04-024 | KEEP | Upgrade test |
| Downgrade from Commercial to Open Source license | IA-04-025 | KEEP | Downgrade test |
| License reactivation preserves settings | IA-04-026 | DELETE | Edge case |

---

## Section: LICENSE DATA RETENTION

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Open Source license 7-day retention | IA-04-038 | KEEP | Retention test |
| Commercial license 90-day retention | IA-04-039 | KEEP | Retention test |

---

## Section: LICENSE TIERS

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Get license tiers without authentication | IA-04-028 (p0) | KEEP | Public endpoint |
| Get license tiers as manager | IA-04-029 | DELETE | Covered by IA-04-028 |
| Get license tiers as member | IA-04-030 | DELETE | Covered by IA-04-028 |
| License tiers response contains valid data | IA-04-031 | DELETE | Covered by IA-04-028 |
| Get license tiers multiple times | IA-04-032 | DELETE | Repeat test |
| License tiers are consistent | IA-04-033 | DELETE | Repeat test |

---

## Section: TEAM INVITATIONS

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Manager sends team invitation | IA-05-001 | KEEP | Happy path |
| Invitation expires after 7 days | IA-05-002 | KEEP | Expiration test |
| Accept invitation with new user | IA-05-003 | KEEP | Happy path |
| Accept invitation sets up password | IA-05-004 | DELETE | Covered by IA-05-003 |
| Manager cancels pending invitation | IA-05-005 | KEEP | Cancellation |
| Duplicate email in organization | IA-05-006 | KEEP | Duplicate check |
| Resend invitation email | IA-05-007 | KEEP | Resend test |
| Member cannot send invitations | IA-05-008 | KEEP | RBAC test |

---

## Section: LAST MANAGER PROTECTION

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| Last manager cannot demote themselves | IA-06-001 | KEEP | Core protection |
| Last manager cannot be deactivated by another manager | IA-06-002 | KEEP | Core protection |
| Manager can be demoted when other managers exist | IA-06-003 | KEEP | Allow when safe |

---

## Section: USER MANAGEMENT

| Scenario | Requirement | Action | Notes |
|----------|-------------|--------|-------|
| List all users as manager | IA-01-146 | KEEP | Happy path |
| List users multiple times | IA-01-147 | DELETE | Repeat test |
| List users as member returns forbidden | IA-01-148 | KEEP | RBAC test |
| List users without authentication returns unauthorized | IA-01-149 | KEEP | Auth test |

---

## Section: COVERAGE PUSH SCENARIOS (DELETE ALL)

**Rule**: These scenarios exist solely to increase coverage numbers. They test the same behavior repeatedly.

| Rule | Scenarios to DELETE |
|------|---------------------|
| Simple Authentication Flows | IA-01-039 through IA-01-043 |
| More Authentication Variations | IA-01-044 through IA-01-046 |
| Extended Authentication Flows | IA-01-047 through IA-01-050 |
| Additional Authentication Paths | IA-01-055 through IA-01-060 |
| Token Verify Extended Scenarios | IA-01-051 through IA-01-054 |
| Additional Authentication Flows (2) | IA-01-081 through IA-01-083 |
| Login Extended Variations | IA-01-068 through IA-01-070 |
| Registration Extended Variations | IA-01-071 through IA-01-075 |
| Verify Token Extended Scenarios | IA-01-076 through IA-01-078 |
| Token Refresh Extended Scenarios | IA-01-079 through IA-01-080 |
| Login Extended Variations (2) | IA-01-104 through IA-01-107 |
| Profile Extended Queries | IA-01-108 through IA-01-111 |
| Registration Additional Variations | IA-01-112 through IA-01-115 |
| Profile Extended Paths | IA-01-096 through IA-01-100 |
| Refresh Token Error Paths | IA-01-116 through IA-01-121 |
| Profile Error Paths | IA-01-122 through IA-01-124 |
| Additional Refresh Scenarios | IA-01-125 through IA-01-127 |
| Registration Validation Extended | IA-01-128 through IA-01-130 |
| Profile Extended Scenarios | IA-01-131 through IA-01-133 |
| Profile Variations | IA-01-134 through IA-01-137 |
| Login Variations | IA-01-138 through IA-01-142 |
| Simple Auth Flows | IA-01-143 through IA-01-145 |
| Authentication Basic Flows | IA-01-150 through IA-01-153 |
| Simple Auth Flows (2) | IA-01-154 through IA-01-156 |
| Authentication Repeated Operations | IA-01-157 through IA-01-160 |
| Provider Simple Operations | IA-01-161 through IA-01-163 |
| Additional Simple Scenarios | IA-01-164 through IA-01-177 |
| Final Coverage Push | IA-01-178 through IA-01-191 |
| Extended Coverage Scenarios | IA-01-192 through IA-01-200 |
| Maximum Coverage Push | IA-01-201 through IA-01-210 |
| Additional Extended Scenarios | IA-01-207 through IA-01-210 |
| Additional Profile Scenarios | IA-01-221 through IA-01-232 |

---

## Section: MISC / MOVE TO OTHER FEATURES

| Scenario | Requirement | Action | Target Feature |
|----------|-------------|--------|----------------|
| Dashboard metrics accessible to managers | IA-01-060 | MOVE | 04_user_interfaces.feature |
| Dashboard operations (various) | IA-01-xxx | MOVE | 04_user_interfaces.feature |
| Provider list operations | IA-01-xxx | MOVE | 01_provider_management.feature |
| Health check operations | IA-01-xxx | MOVE | 05_system_behaviors.feature |
| Usage statistics operations | IA-01-xxx | MOVE | 03_usage_insights.feature |

---

## Reduction Strategy by Section

### 1. AUTHENTICATION (Login/Logout/Verify/Refresh)
- **Current**: ~90 scenarios
- **Target**: ~25 scenarios
- **Reduction**: 65 scenarios (72%)
- **Keep**:
  - 1-2 login happy paths (convert to outline for roles)
  - 3-4 login error paths
  - 1 logout
  - 2-3 refresh happy paths
  - 3-4 refresh error paths
  - 1 verify happy path
  - 2-3 verify error paths

### 2. USER PROFILE
- **Current**: ~80 scenarios
- **Target**: ~8 scenarios
- **Reduction**: 72 scenarios (90%)
- **Keep**:
  - 1 manager profile
  - 1 member profile
  - 1 comprehensive field validation
  - 2-3 error paths

### 3. REGISTRATION
- **Current**: ~25 scenarios
- **Target**: ~8 scenarios
- **Reduction**: 17 scenarios (68%)
- **Keep**:
  - 1 happy path
  - 1 password outline
  - 4-5 validation errors

### 4. PASSWORD RESET
- **Current**: 4 scenarios
- **Target**: 4 scenarios
- **Reduction**: 0 (already lean)

### 5. ACCOUNT LOCKOUT
- **Current**: 4 scenarios
- **Target**: 4 scenarios
- **Reduction**: 0 (already lean)

### 6. RBAC
- **Current**: 9 scenarios
- **Target**: 9 scenarios
- **Reduction**: 0 (unique tests)

### 7. MULTI-TENANT
- **Current**: 4 scenarios
- **Target**: 3 scenarios
- **Reduction**: 1 (merge duplicate)

### 8. LICENSE MANAGEMENT
- **Current**: ~50 scenarios
- **Target**: ~30 scenarios
- **Reduction**: 20 scenarios (40%)
- **Keep**:
  - Activation tests (merge duplicates)
  - Feature flags (keep outline)
  - Status checks (merge duplicates)
  - Limits/enforcement (keep unique)
  - Expiration (keep outline)
  - Upgrades (keep)

### 9. TEAM INVITATIONS
- **Current**: 8 scenarios
- **Target**: 7 scenarios
- **Reduction**: 1 (merge duplicate)

### 10. LAST MANAGER PROTECTION
- **Current**: 3 scenarios
- **Target**: 3 scenarios
- **Reduction**: 0 (critical tests)

### 11. USER MANAGEMENT
- **Current**: 4 scenarios
- **Target**: 3 scenarios
- **Reduction**: 1 (remove repeat)

### 12. LICENSE TIERS
- **Current**: 6 scenarios
- **Target**: 1 scenario
- **Reduction**: 5 (merge into single test)

---

## Final Count

| Section | Before | After | Reduction |
|---------|--------|-------|-----------|
| Authentication | 90 | 25 | -65 (-72%) |
| User Profile | 80 | 8 | -72 (-90%) |
| Registration | 25 | 8 | -17 (-68%) |
| Password Reset | 4 | 4 | 0 |
| Account Lockout | 4 | 4 | 0 |
| RBAC | 9 | 9 | 0 |
| Multi-Tenant | 4 | 3 | -1 |
| License Mgmt | 50 | 30 | -20 (-40%) |
| Team Invitations | 8 | 7 | -1 |
| Last Manager Protection | 3 | 3 | 0 |
| User Management | 4 | 3 | -1 |
| License Tiers | 6 | 1 | -5 |
| Misc (Dashboard/Health/etc) | 58 | 0 | -58 (MOVE) |
| **TOTAL** | **345** | **105** | **-240 (-70%)** |

**Note**: After moving dashboard/health/provider/misc scenarios to appropriate features, the identity_and_access.feature will have ~105 scenarios, well within the 150-200 target.

---

## Recommended Split Structure

After reduction, split 00_identity_and_access.feature into:

1. **auth.feature** (~35 scenarios)
   - Login, logout, refresh, verify
   - Password reset
   - Account lockout

2. **profile.feature** (~15 scenarios)
   - Get profile
   - Profile field validation
   - Profile errors

3. **invitations.feature** (~7 scenarios)
   - Team invitations
   - Invitation acceptance
   - Invitation cancellation

4. **licensing.feature** (~45 scenarios)
   - License activation
   - License features
   - License status
   - License limits
   - License expiration
   - License tiers

5. **rbac.feature** (~20 scenarios)
   - Role-based access control
   - Multi-tenant isolation
   - Last manager protection
   - User management
