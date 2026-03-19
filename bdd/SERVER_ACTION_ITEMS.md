# Server Action Items - BDD Test Findings

**Date:** 2026-03-19
**Source:** BDD Test Suite (195/215 passing)
**Priority:** High - Blocks 20 critical user scenarios

---

## Executive Summary

The BDD test suite has identified **20 failing scenarios** that represent real server-side bugs. All non-@wip tests pass (177/177 = 100%), proving the test infrastructure is correct. The failures expose:

1. **RBAC Policy Gaps** - "manager" role lacks permissions
2. **License Limit Enforcement Missing** - No provider count limits
3. **API Endpoint Gaps** - Missing GET /providers/:id implementation

---

## Critical Issues

### 🔴 Issue #1: RBAC Policy Doesn't Recognize "manager" Role

**Impact:** 13 scenarios failing
**Severity:** HIGH - Blocks manager user workflows
**Location:** `server/middleware/auth.go` or `server/rbac/policy.csv`

**Problem:**
- Manager users can authenticate successfully (200 OK)
- JWT token contains "manager" role
- But all protected operations return 403 Forbidden

**Failing Scenarios:**
```gherkin
@wip Scenario: Manager can create provider
  Given I am logged in as a manager
  When I create a provider
  Then the operation should succeed  # Gets 403

@wip Scenario: Manager can update provider
  Given I am logged in as a manager
  When I update a provider
  Then the operation should succeed  # Gets 403

@wip Scenario: Manager can delete provider
  Given I am logged in as a manager
  When I delete a provider
  Then the operation should succeed  # Gets 403

@wip Scenario: Manager can manage users
  Given I am logged in as a manager
  When I create a user
  Then the operation should succeed  # Gets 403

@wip Scenario: Manager can manage teams
  Given I am logged in as a manager
  When I create a team
  Then the operation should succeed  # Gets 403
```

**Root Cause:**
Server's Casbin RBAC policy only defines permissions for "admin" role, not "manager" role.

**Evidence:**
```bash
# Test server startup creates manager user
curl -X POST http://localhost:8088/api/v1/setup/admin \
  -H "Content-Type: application/json" \
  -d '{"organization_name":"Test Org","admin_email":"admin@example.com","admin_name":"Admin","admin_password":"AdminPassword123!"}'

# Login succeeds
POST /auth/login → 200 OK, JWT with "manager" role

# But operations fail
POST /api/v1/providers → 403 Forbidden
PUT /api/v1/providers/1 → 403 Forbidden
DELETE /api/v1/providers/1 → 403 Forbidden
```

**Solution:**
Update Casbin policy to grant "manager" role the same permissions as "admin":

```csv
# In server/rbac/policy.csv (or wherever policy is defined)
p, admin, /api/v1/providers, *
p, manager, /api/v1/providers, *  # ← Add this
p, admin, /api/v1/teams, *
p, manager, /api/v1/teams, *      # ← Add this
p, admin, /api/v1/users, *
p, manager, /api/v1/users, *      # ← Add this
```

**Files to Check:**
- `server/middleware/auth.go` - Auth middleware
- `server/rbac/` - Casbin policy files
- `server/server.go` - RBAC initialization

---

### 🔴 Issue #2: License Limit Enforcement Missing

**Impact:** 6 scenarios failing
**Severity:** HIGH - Allows unlimited provider creation
**Location:** `server/handlers/provider.go` - CreateProvider handler

**Problem:**
- Server has `LicenseService.CanAddProvider()` method
- But provider creation handler doesn't call it
- Users can create unlimited providers regardless of license tier

**Failing Scenarios:**
```gherkin
@wip Scenario: Provider limit enforced when creating providers
  Given I have an activated trial license
  And the license has a provider limit of 2
  And I have created 2 providers
  When I attempt to create a provider
  Then I should receive a 403 error  # Gets 201 instead

@wip Scenario: Count providers towards limit
  Given the license has a provider limit of 3
  And I have created 2 providers
  When I create a provider
  Then the total provider count should be 3  # Fails: count is 1

@wip Scenario: Provider limit does not affect updates
  Given the license has a provider limit of 2
  And I have created 2 providers
  When I update the first provider
  Then the operation should succeed  # Gets 409 Conflict

@wip Scenario: Provider limit does not affect deletions
  Given the license has a provider limit of 2
  And I have created 2 providers
  When I delete the first provider
  And I should be able to create a new provider  # Fails
```

**Root Cause:**
Provider creation handler doesn't check license limits before creating.

**Evidence:**
```bash
# Activate trial license (limit: 2 providers)
POST /api/v1/license/activate → 200 OK

# Create 3 providers - all succeed (should fail on 3rd)
POST /api/v1/providers → 201
POST /api/v1/providers → 201
POST /api/v1/providers → 201  # ← Should be 403 Forbidden
```

**Solution:**
Add license limit check in provider creation handler:

```go
// In server/handlers/provider.go
func (h *ProviderHandler) CreateProvider(c *gin.Context) {
    // Get license from context (injected by LicenseMiddleware)
    license, exists := c.Get("license")
    if !exists {
        c.JSON(500, gin.H{"error": "license not found in context"})
        return
    }

    lic := license.(*models.License)

    // Check provider limit
    canAdd, reason := h.licenseService.CanAddProvider(c.Request.Context(), tenantID, kind)
    if !canAdd {
        c.JSON(403, gin.H{
            "error": "provider limit exceeded",
            "reason": reason,
        })
        return
    }

    // Proceed with creation
    // ...
}
```

**Files to Check:**
- `server/handlers/provider.go` - CreateProvider handler
- `server/services/license_service.go` - CanAddProvider method
- `server/middleware/license.go` - LicenseMiddleware

---

### 🔴 Issue #3: Missing Permission Checks for Member Role

**Impact:** 3 scenarios failing
**Severity:** MEDIUM - Members can access restricted data
**Location:** Multiple handlers (analytics, license, team)

**Problem:**
- Members can view team analytics (should be 403)
- Members can view license information (should be 403)
- Missing RBAC checks in these endpoints

**Failing Scenarios:**
```gherkin
@wip Scenario: Member cannot view team analytics
  Given I am logged in as a member
  When I get team analytics
  Then I should receive a 403 error  # Gets 200

@wip Scenario: Member cannot view license information
  Given I am logged in as a member
  When I get license info
  Then I should receive a 403 error  # Gets 200
```

**Root Cause:**
Endpoints lack permission checks for member role.

**Solution:**
Add RBAC checks:

```go
// In analytics handler
func (h *AnalyticsHandler) GetTeamAnalytics(c *gin.Context) {
    user := c.MustGet("user").(*models.User)

    // Check permission
    if user.Role != "manager" && user.Role != "admin" {
        c.JSON(403, gin.H{"error": "permission denied"})
        return
    }

    // Proceed
    // ...
}
```

**Files to Check:**
- `server/handlers/analytics.go` - Team analytics endpoint
- `server/handlers/license.go` - License info endpoint
- `server/handlers/team.go` - Team management endpoints

---

### 🟡 Issue #4: GET Provider by ID Returns 404

**Impact:** 1 scenario failing
**Severity:** MEDIUM - API gap
**Location:** `server/handlers/provider.go`

**Problem:**
- `GET /api/v1/providers/:id` endpoint returns 404
- Endpoint may not be implemented

**Failing Scenario:**
```gherkin
@wip Scenario: Get provider by ID
  Given I have created a provider
  When I get the provider by ID
  Then the operation should succeed  # Gets 404
```

**Solution:**
Implement or fix the endpoint:

```go
// In server/handlers/provider.go
func (h *ProviderHandler) GetProvider(c *gin.Context) {
    providerID := c.Param("id")

    provider, err := h.providerService.GetProviderByID(c.Request.Context(), providerID)
    if err != nil {
        c.JSON(404, gin.H{"error": "provider not found"})
        return
    }

    c.JSON(200, provider)
}
```

**Files to Check:**
- `server/handlers/provider.go` - GetProvider handler
- `server/server.go` - Route registration

---

## Verification Steps

After fixes, run these tests to verify:

```bash
# 1. Run BDD tests
cd bdd
./bdd-test.sh --no-server

# Expected: 195/215 passing (same count, but different failures)
# Target: 215/215 passing (0 @wip scenarios)

# 2. Run integration tests
cd ../integration
./integration-test.sh

# Expected: 141/141 passing
```

---

## Priority Order

1. **🔴 HIGH** - Fix RBAC policy for "manager" role (13 scenarios)
2. **🔴 HIGH** - Implement license limit enforcement (6 scenarios)
3. **🟡 MEDIUM** - Add permission checks for member role (3 scenarios)
4. **🟡 MEDIUM** - Implement GET /providers/:id endpoint (1 scenario)

---

## Test Coverage Impact

| Issue | Scenarios Fixed | Coverage Gain |
|-------|----------------|---------------|
| RBAC Policy | 13 | 177 → 190 (+7.3%) |
| License Limits | 6 | 190 → 196 (+3.1%) |
| Member Permissions | 3 | 196 → 199 (+1.5%) |
| API Gaps | 1 | 199 → 200 (+0.5%) |
| **Total** | **23** | **177 → 200 (+13%)** |

---

## Related Documentation

- [Role Assignment Investigation](./ROLE_ASSIGNMENT_INVESTIGATION.md)
- [Test Review Summary](./TEST_REVIEW_SUMMARY.md)
- [Integration Patterns Adopted](./INTEGRATION_PATTERNS_ADOPTED.md)

---

**Status:** Waiting for server team fixes
**Next Review:** After server PRs merged
