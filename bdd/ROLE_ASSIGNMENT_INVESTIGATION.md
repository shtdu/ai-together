# Role Assignment Investigation

**Date:** 2026-03-19
**Purpose:** Investigate how integration tests handle role assignment and apply learnings to BDD tests

---

## Executive Summary

The investigation revealed that **role assignment works correctly** in both integration and BDD tests. The 20 failing @wip scenarios are caused by **server-side RBAC policy gaps**, not test configuration issues.

**Key Finding:** Managers authenticate successfully with "manager" role, but the server's RBAC system doesn't grant them the necessary permissions.

---

## How Role Assignment Works

### Integration Test Flow

```
┌─────────────────────────────────────────────────────────────┐
│ test-server.sh (startup)                                    │
│                                                             │
│  1. Start server on port 8088                               │
│  2. Wait for health check                                   │
│  3. Call /api/v1/setup/admin endpoint                       │
│     POST {                                                  │
│       "organization_name": "Test Org",                      │
│       "admin_email": "admin@example.com",                   │
│       "admin_name": "Admin",                                │
│       "admin_password": "AdminPassword123!"                 │
│     }                                                       │
│  4. Server creates:                                         │
│     - Tenant (ID: 1)                                        │
│     - Admin user with "manager" role                        │
│     - Default team                                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ Integration Tests                                           │
│                                                             │
│  - Login as manager: admin@example.com                      │
│  - Create member users via /auth/register (gets "member")   │
│  - Test RBAC with properly authenticated users              │
└─────────────────────────────────────────────────────────────┘
```

### BDD Test Flow

```
┌─────────────────────────────────────────────────────────────┐
│ bdd-test.sh (startup)                                       │
│                                                             │
│  1. Check if server running                                 │
│  2. If not, call ../integration/test-server.sh              │
│     (This creates manager user via setup endpoint)          │
│  3. Run BDD scenarios                                       │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ BDD Scenarios                                               │
│                                                             │
│  iAmLoggedInAsAManager():                                   │
│    1. Load fixture: admin@example.com / AdminPassword123!   │
│    2. Call ensureUserExistsViaPublicRegistration()          │
│    3. Try login first (succeeds - user exists from setup)   │
│    4. Store token with "manager" role                       │
│                                                             │
│  iAmLoggedInAsAMember():                                    │
│    1. Load fixture: member@example.com                      │
│    2. Call ensureUserExistsViaPublicRegistration()          │
│    3. Try login (may fail)                                  │
│    4. Register via /auth/register (gets "member" role)      │
│    5. Login and store token                                 │
└─────────────────────────────────────────────────────────────┘
```

---

## Role Assignment Endpoints

### `/api/v1/setup/admin` (Setup Endpoint)
- **Purpose:** One-time tenant initialization
- **Creates:** User with "manager" role
- **Called by:** `test-server.sh` during test startup
- **Request:**
  ```json
  {
    "organization_name": "Test Org",
    "admin_email": "admin@example.com",
    "admin_name": "Admin",
    "admin_password": "AdminPassword123!"
  }
  ```

### `/auth/register` (Public Registration)
- **Purpose:** Public user registration
- **Creates:** User with "member" role (hardcoded in server)
- **Server code** (`server/handlers/auth_handler.go`):
  ```go
  user, err := h.userService.CreateUser(
      req.Email,
      req.Password,
      req.Name,
      "member",  // ← Hardcoded role
      tenantID,
  )
  ```

---

## Investigation Results

### What Was Fixed

1. **Fixture Configuration** (`bdd/testdata/fixtures/users.json`):
   ```json
   {
     "admin": {
       "email": "admin@example.com",
       "password": "AdminPassword123!",
       "name": "Test Admin",
       "role": "manager"  // ← Added this field
     },
     "member": {
       "email": "member@example.com",
       "password": "MemberPass123!",
       "name": "Test Member",
       "role": "member"  // ← Added this field
     }
   }
   ```

2. **Role Validation** (`bdd/step_definitions/auth_steps.go`):
   ```go
   // Before: Only accepted "admin" role
   if adminUser.Role != "admin" { ... }

   // After: Accept both "admin" and "manager"
   if adminUser.Role != "admin" && adminUser.Role != "manager" { ... }
   ```

### What Was Verified

✅ Manager user is created with "manager" role (via setup endpoint)
✅ BDD tests can authenticate as manager successfully
✅ JWT token contains "manager" role
✅ Authentication succeeds (200 OK with valid token)
✅ Role is correctly stored in test context

### Root Cause of Failures

The 403 Forbidden errors are **NOT** caused by missing roles. They're caused by:

**Server RBAC Policy Gaps:**
- Casbin policy doesn't grant "manager" role access to:
  - Provider management endpoints (create/update/delete)
  - Team management endpoints
  - User management endpoints
- Permission checks missing for member role on:
  - Team analytics endpoints
  - License information endpoints

**License Limit Enforcement:**
- `LicenseService.CanAddProvider()` not called in provider creation handler
- Server allows unlimited provider creation regardless of license tier

---

## Test Results

### Current Status: 195/215 passing (90.7%)

**Non-@wip tests:** 177/177 (100%) ✅
**@wip tests:** 18/38 (47%) - These expose real server bugs

### Failure Categories

| Category | Scenarios | Root Cause |
|----------|-----------|------------|
| RBAC Permission Gaps | 13 | Server's Casbin policy doesn't recognize "manager" role |
| License Limits | 6 | Provider creation doesn't enforce license limits |
| API Gaps | 1 | GET /providers/:id endpoint missing |

---

## Lessons Learned

### 1. Test Server Setup is Critical
- Both integration and BDD tests depend on `test-server.sh`
- Setup endpoint creates the manager user with correct role
- Tests should NOT try to create manager via public registration

### 2. Real API Testing Catches Real Bugs
- With mocks: Tests passed but hid RBAC bugs
- With real API: Tests fail, exposing real permission gaps
- **This is the correct behavior!** Tests should fail when server has bugs

### 3. Role Hierarchy
```
manager (from /api/v1/setup/admin)
   └─ Can manage providers, teams, users
   └─ Created by setup endpoint

member (from /auth/register)
   └─ Can only view own data
   └─ Created by public registration
```

### 4. Authentication vs Authorization
- **Authentication** ✅ Working: Users can log in, get valid JWT tokens
- **Authorization** ❌ Broken: RBAC doesn't grant proper permissions

---

## Recommendations

### For BDD Tests
1. ✅ **DONE** - Use fixture with "manager" role
2. ✅ **DONE** - Accept both "admin" and "manager" roles
3. ✅ **DONE** - Keep @wip tags on failing scenarios (they expose real bugs)

### For Server Team
1. **URGENT** - Fix RBAC policy for "manager" role:
   - Grant provider management permissions
   - Grant team management permissions
   - Grant user management permissions

2. **URGENT** - Implement license limit enforcement:
   - Call `LicenseService.CanAddProvider()` before creating provider
   - Return 403 if limit exceeded

3. **TODO** - Add permission checks for member role:
   - Block access to team analytics
   - Block access to license information

### For Integration Tests
- ✅ Already using correct pattern (setup endpoint for manager)
- ✅ Tests correctly expose RBAC bugs

---

## Files Modified

| File | Change |
|------|--------|
| `bdd/testdata/fixtures/users.json` | Added `role` field to users |
| `bdd/step_definitions/auth_steps.go` | Accept "manager" role in addition to "admin" |
| `bdd/TEST_REVIEW_SUMMARY.md` | Updated failure analysis with RBAC findings |
| `bdd/INTEGRATION_PATTERNS_ADOPTED.md` | Added Pattern 8: Role Assignment |
| `bdd/ROLE_ASSIGNMENT_INVESTIGATION.md` | This document |

---

## References

- Integration Test Setup: `../integration/test-server.sh`
- Setup Endpoint: `server/handlers/setup_handler.go:137`
- Auth Handler: `server/handlers/auth_handler.go`
- RBAC Configuration: `server/middleware/auth.go`
- BDD Fixtures: `bdd/testdata/fixtures/users.json`

---

**Investigation Status:** ✅ Complete
**Next Steps:** Server team to fix RBAC policies and license enforcement
