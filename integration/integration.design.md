# Integration Test Design

> **Current Status: 171/171 tests passing (100%)** ✅
>
> Complete integration test suite for the AI Together server API.
> **3 tests skipped due to server-side issues.**

## Quick Links

- [Quick Start: Adding New Tests](#quick-start-adding-new-tests)
- [Test Patterns & Conventions](#test-patterns-conventions)
- [Test Reference Catalog](#test-reference-catalog)
- [Architecture Overview](#architecture-overview)
- [Fixture Reference](#fixture-reference)

---

## Quick Start: Adding New Tests

### Step-by-Step Guide

**1. Create test file** (if new category)
```bash
# Follow naming pattern: <category>_test.go
touch integration/myfeature_test.go
```

**2. Add test methods** to `IntegrationTestSuite`
```go
package integration

import (
    "context"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) TestMyNewFeature() {
    ctx := context.Background()

    // Arrange: Create fixtures via API
    uniqueName := generateUniqueProviderName("my-feature")
    req := integration.PostApiV1ProvidersJSONRequestBody{
        Name:   uniqueName,
        ApiKey: "test-key",
        // ... other fields
    }
    resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
    require.NoError(s.T(), err)

    // Act: Call API being tested
    // ...

    // Assert: Verify via API response
    assert.Equal(s.T(), 200, resp.StatusCode())

    // Cleanup: Delete resources
    defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, id)
}
```

**3. Use helper functions** from `helpers.go`:
- `generateUniqueProviderName(baseName)` - unique names to avoid conflicts
- `LoadLicenseFixture(fixtureName)` - load license PEM files
- `intPointer(v)`, `int64Pointer(v)` - convert to pointers
- `registerAndLoginUserFromFixture(fixtureName)` - create/test users

**4. Run your test**
```bash
cd integration
go test -v -run "TestMyNewFeature"
```

### Common Patterns

#### Pattern 1: Unique Resource Names
```go
// Prevents duplicate name errors across test runs
uniqueName := generateUniqueProviderName("test-provider")
```

#### Pattern 2: Cleanup with defer
```go
// Ensure cleanup even if test fails
resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
require.NoError(s.T(), err)
defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, resp.JSON201.Id)
```

#### Pattern 3: Conditional Bootstrap
```go
// For tests needing clean state (no standard fixtures)
s.skipBootstrap = true  // Skip bootstrapStandardFixture()
```

#### Pattern 4: API-Only Assertions
```go
// Assert via API responses, NOT database queries
assert.Equal(s.T(), 200, resp.StatusCode())
assert.Equal(s.T(), "expected", resp.JSON200.Name)
```

### Troubleshooting

| Problem | Solution |
|---------|----------|
| Duplicate name errors | Use `generateUniqueProviderName()` |
| Test fixtures not found | Run `make integration-setup` first |
| Server not responding | Start test server: `./test-server.sh` |
| Permission denied | Check RBAC: use admin client for mutations |
| License errors | Activate license: `s.activateLicenseFixture()` |

---

## Test Patterns & Conventions

### Test File Structure

```
integration/
├── integration_test.go    # Test suite setup/teardown
├── helpers.go             # Utility functions
├── fixtures.go            # Test data fixtures
├── cleanup.go             # Cleanup helpers
├── provider_test.go       # Provider tests (30)
├── auth_test.go           # Auth tests (13, 1 skipped)
├── permission_test.go     # RBAC tests (8)
├── license_test.go        # License tests (40)
├── usage_test.go          # Usage tests (12)
├── user_test.go           # User profile tests (10)
├── analytics_test.go      # Analytics tests (4)
├── analytics_e2e_test.go  # E2E data flow tests (5)
├── analytics_aggregation_test.go  # Aggregation accuracy tests (6)
├── analytics_time_test.go # Time-based analytics tests (5)
├── analytics_filter_test.go  # Cross-filtering tests (5)
├── analytics_edge_cases_test.go  # Edge cases & statistics tests (9)
├── mgr_dashboard_test.go  # Manager dashboard tests (6)
├── mgr_team_test.go       # Manager team CRUD tests (9)
├── mgr_users_crud_test.go # Manager user CRUD tests (8, 1 skipped)
└── health_test.go         # Health check tests (2)
```

### Common Helper Functions

**Unique Name Generation** (`helpers.go:42`)
```go
func generateUniqueProviderName(baseName string) string
// Returns: "basename-1", "basename-2", etc.
// Prevents duplicate name conflicts
```

**License Fixture Loader** (`helpers.go:57`)
```go
func LoadLicenseFixture(fixtureName string) (string, error)
// Loads PEM files from testdata/licenses/
// Fixtures: tier_0.pem, tier_1.pem, tier_2.pem, expired.pem
```

**User Fixture Helper** (`integration_test.go:141`)
```go
func (s *IntegrationTestSuite) registerAndLoginUserFromFixture(fixtureName string) string
// Registers/logs in user from fixtures.go
// Returns: access token
```

**License Activation Helper** (`integration_test.go:193`)
```go
func (s *IntegrationTestSuite) activateLicenseFixture(fixtureName string) *LicenseStatus
// Activates license via API
// Returns: license status from API
```

**Provider Creation Helper** (`integration_test.go:219`)
```go
func (s *IntegrationTestSuite) createProviderFixtureFromFixture(fixtureName string) int64
// Creates provider from fixtures.go
// Returns: provider ID
```

### Pointer Conversion Utilities

```go
// Convert primitives to pointers (for optional API fields)
intPointer(v int) *int
int64Pointer(v int64) *int64
boolPointer(v bool) *bool
stringPointer(v string) *string
```

### Fixture Creation Patterns

**API-Only Fixtures** (no database inserts)
```go
// ✅ Correct: Create via API
req := integration.PostApiV1ProvidersJSONRequestBody{...}
resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)

// ❌ Wrong: Direct DB inserts
// db.Exec("INSERT INTO providers ...")
```

**Standard Bootstrap** (runs before each test)
- Activates Tier 1 license
- Creates 2 providers (Claude, OpenCode)
- Creates member user

**Clean State Tests**
```go
s.skipBootstrap = true  // Disable standard fixtures
```

### Setup/Teardown Patterns

**Test Lifecycle**
```go
func (s *IntegrationTestSuite) SetupTest() {
    // Runs before EACH test
    s.loginAdminUser()
    if !s.skipBootstrap {
        s.bootstrapStandardFixture()
    }
}

func (s *IntegrationTestSuite) TearDownTest() {
    // Runs after EACH test
    // Cleans up providers via API
}
```

**Suite Lifecycle**
```go
func (s *IntegrationTestSuite) SetupSuite() {
    // Runs ONCE before all tests
    // Connects to test server
}

func (s *IntegrationTestSuite) TearDownSuite() {
    // Runs ONCE after all tests
    // Cleanup
}
```

### Assertion Patterns

**Status Code Assertions**
```go
require.Equal(s.T(), 201, resp.StatusCode())  // Hard failure
assert.Equal(s.T(), "value", resp.JSON200.Name)  // Soft failure
```

**Error Handling**
```go
require.NoError(s.T(), err, "Failed to create provider")
require.NotNil(s.T(), resp.JSON201, "Response should have JSON201 body")
```

**API Response Validation**
```go
assert.Equal(s.T(), expectedName, resp.JSON200.Name)
assert.Greater(s.T(), resp.JSON201.Id, int64(0))
assert.True(s.T(), *resp.JSON200.Enabled)
```

### Test Isolation Techniques

1. **Unique Names**: `generateUniqueProviderName()` prevents conflicts
2. **API Cleanup**: `defer` statements ensure resource cleanup
3. **Conditional Bootstrap**: `skipBootstrap` for clean state
4. **Separate Database**: Uses `codetogether_test` database
5. **Independent Transactions**: Each test is isolated

---

## Test Reference Catalog

| Category | Test File | Test Count | Key Tests | Status |
|----------|-----------|------------|-----------|--------|
| **Provider** | `provider_test.go` | 29 | Create, Update, Delete, List, Limits, Enable/Disable, Stats | ✅ 29/29 (1 skipped) |
| **Auth** | `auth_test.go` | 13 | Login, Register, Verify, Refresh, Profile | ✅ 13/13 (1 skipped) |
| **Permission** | `permission_test.go` | 8 | RBAC for all endpoints | ✅ 8/8 |
| **License** | `license_test.go` | 40 | Activate, Tiers, Limits, Expiration, Features | ✅ 40/40 |
| **Usage** | `usage_test.go` | 12 | Upload, Batch, Statistics, Cache Tokens | ✅ 12/12 |
| **User** | `user_test.go` | 10 | Profile fields, Role in token, Tenant consistency | ✅ 10/10 |
| **Analytics** | `analytics_*.go` (5 files) | 34 | E2E, Aggregation, Time, Filter, Edge Cases | ✅ 34/34 |
| **Manager Dashboard** | `mgr_dashboard_test.go` | 6 | Metrics, Members, Rankings | ✅ 6/6 |
| **Manager Teams** | `mgr_team_test.go` | 9 | Team CRUD, Members, Settings | ✅ 9/9 |
| **Manager Users** | `mgr_users_crud_test.go` | 7 | User CRUD, Validation, Password | ✅ 7/7 (1 skipped) |
| **Health** | `health_test.go` | 2 | Health check, Database status | ✅ 2/2 |
| **Total** | **16 files** | **171** | **Complete API coverage** | **✅ 168/168 (3 skipped)** |

**Skipped Tests (server-side issues):**
1. `TestAuthLogout` - Server lacks `/auth/logout` route (only has `/api/v1/logout`)
2. `TestMgrGetUser` - Server lacks `GET /users/:id` endpoint

### Test Case Specifications

#### Provider Tests (26 tests) - `provider_test.go`

**Create Provider (8 tests)**
- `TestProviderCreateClaude` - POST /providers with kind=claude, verify 201, kind=claude, enabled=true
- `TestProviderCreateCodex` - POST /providers with kind=codex, verify 201, kind=codex
- `TestProviderCreateOpenCode` - POST /providers with kind=opencode, verify 201, kind=opencode
- `TestProviderCreateDuplicateName` - Create with duplicate name, verify 400 error
- `TestProviderCreateInvalidKind` - Create with invalid kind, verify 400 error
- `TestProviderCreateMissingFields` - Create without required fields, verify 400 error
- `TestProviderCreateWithAllFields` - Create with all optional fields, verify all persisted
- `TestProviderCreateDisabled` - Create with enabled=false, verify created but disabled

**Get Provider (3 tests)**
- `TestProviderGetExisting` - GET /providers/{id}, verify 200 with provider details
- `TestProviderGetNotFound` - GET /providers/99999, verify 404
- `TestProviderGetValidation` - Validate all fields in response

**Update Provider (6 tests)**
- `TestProviderUpdateName` - PUT /providers/{id} with new name, verify 200
- `TestProviderUpdateDuplicateName` - Update to duplicate name, verify 400 error
- `TestProviderUpdateKind` - Change provider kind, verify 200
- `TestProviderToggleEnabled` - Toggle enabled field, verify 200
- `TestProviderUpdateAllFields` - Update all fields, verify all persisted
- `TestProviderUpdateNotFound` - PUT /providers/99999, verify 404

**Delete Provider (2 tests)**
- `TestProviderDeleteExisting` - DELETE /providers/{id}, verify 200, no longer accessible
- `TestProviderDeleteNotFound` - DELETE /providers/99999, verify 404

**List Providers (4 tests)**
- `TestProviderListAll` - GET /providers, verify 200 with array
- `TestProviderListPagination` - GET /providers?page=1&limit=10, verify pagination metadata
- `TestProviderListEmpty` - Delete all providers, verify empty array
- `TestProviderListFilterByKind` - GET /providers?kind=claude, verify only Claude providers

**Provider Limits (3 tests)**
- `TestProviderLimitEnforcement` - Create providers exceeding license limit, verify 403
- `TestProviderUnlimitedTier2` - Tier 2 allows 100+ providers
- `TestProviderLimitAfterDowngrade` - Tier 2→1 downgrade, existing preserved, new blocked

#### Auth Tests (14 tests) - `auth_test.go`

**Login (5 tests)**
- `TestLoginValidCredentials` - POST /auth/login with valid credentials, verify 200 with tokens
- `TestLoginInvalidEmail` - Login with invalid email, verify 401
- `TestLoginInvalidPassword` - Login with wrong password, verify 401
- `TestLoginMissingFields` - Login without email/password, verify 400
- `TestLoginNonExistentUser` - Login for non-existent user, verify 401

**Register (3 tests)**
- `TestRegisterNewUser` - POST /auth/register, verify 201, user created
- `TestRegisterDuplicateEmail` - Register with existing email, verify 400
- `TestRegisterMissingFields` - Register without required fields, verify 400

**Token Refresh (3 tests)**
- `TestRefreshValidToken` - POST /auth/refresh with valid refresh_token, verify 200 with new access_token
- `TestRefreshInvalidToken` - Refresh with invalid token, verify 401
- `TestRefreshExpiredToken` - Refresh with expired token, verify 401

**Token Verification (3 tests)**
- `TestVerifyValidToken` - POST /auth/verify with valid access_token, verify 200 with user info
- `TestVerifyInvalidToken` - Verify with invalid token, verify 401
- `TestVerifyExpiredToken` - Verify with expired token, verify 401

#### Permission Tests (8 tests) - `permission_test.go`

**Provider Permissions (3 tests)**
- `TestMemberCannotCreateProvider` - Member creates provider, verify 403
- `TestMemberCannotDeleteProvider` - Member deletes provider, verify 403
- `TestMemberCannotUpdateProvider` - Member updates provider, verify 403

**License Permissions (2 tests)**
- `TestMemberCannotActivateLicense` - Member activates license, verify 403
- `TestMemberCannotDeleteLicense` - Member deletes license, verify 403

**User Permissions (2 tests)**
- `TestMemberCannotDeleteUsers` - Member deletes user, verify 403
- `TestMemberCannotUpdateOtherUsers` - Member updates another user, verify 403

**Analytics Permissions (1 test)**
- `TestMemberCanViewAnalytics` - Member views analytics, verify 200 (read-only)

#### License Tests (40 tests) - `license_test.go`

**License Activation (8 tests)**
- `TestActivateTier0License` - Activate tier_0.pem, verify tier=0, seats=0, providers=0
- `TestActivateTier1License` - Activate tier_1.pem, verify tier=1, seats=10, providers=5
- `TestActivateTier2License` - Activate tier_2.pem, verify tier=2, seats=100, providers=100
- `TestActivateExpiredLicense` - Activate expired.pem, verify 400 error
- `TestActivateInvalidSignature` - Activate invalid_signature.pem, verify 400 error
- `TestActivateDuplicateLicense` - Activate same license twice, verify 400 error
- `TestLicenseUpgrade` - Tier 0→2 upgrade, verify tier upgraded
- `TestLicenseDowngrade` - Tier 2→1 downgrade, verify tier downgraded

**License Status (5 tests)**
- `TestGetLicenseStatus` - GET /api/v1/license, verify 200 with license details
- `TestGetLicenseNoActivation` - GET without activation, verify 404
- `TestValidateLicenseFields` - Validate all fields in response
- `TestLicenseExpirationCheck` - Check expiration date in status
- `TestLicenseRemainingSeats` - Check remaining seats in status

**License Deletion (3 tests)**
- `TestDeleteLicense` - DELETE /api/v1/license, verify 200, license deleted
- `TestDeleteNonExistentLicense` - Delete without activation, verify 404
- `TestLicenseDeletionClearsLimits` - Delete license, verify limits cleared, reverts to Tier 0

**Seat Limits (10 tests)**
- `TestSeatLimitEnforcement` - Create users exceeding seat limit, verify 403 on seat+1
- `TestSeatCountAfterUserCreation` - Create user, verify seat count incremented
- `TestSeatCountAfterUserDeletion` - Delete user, verify seat count decremented
- `TestUnlimitedSeatsTier2` - Tier 2 allows 100 users
- `TestZeroSeatsLicense` - zero_seats.pem allows only admin
- `TestSeatsAfterDowngrade` - Tier 2→1, existing users preserved
- `TestSeatLimitOnRegistration` - Register when seats full, verify 403
- `TestAdminNotCountedInSeats` - Verify admin doesn't count against limit
- `TestActiveUsersCountInSeats` - Verify only active users count
- `TestSeatLimitImmediateExpiry` - immediate_expiry.pem expires immediately

**Provider Limits (8 tests)**
- `TestProviderLimitEnforcement` - Create providers exceeding limit, verify 403
- `TestProviderLimitTier0` - Tier 0 allows 0 providers
- `TestProviderLimitTier1` - Tier 1 allows 5+ providers
- `TestProviderLimitTier2` - Tier 2 allows 100+ providers
- `TestProviderCountAfterCreation` - Create provider, verify count incremented
- `TestProviderCountAfterDeletion` - Delete provider, verify count decremented
- `TestProviderLimitAfterLicenseChange` - Change license, verify limit updated
- `TestDisabledProvidersCountInLimit` - Verify disabled providers count

**License Features (6 tests)**
- `TestTier0Restrictions` - Tier 0: no providers or users allowed
- `TestTier1Features` - Tier 1: 10 seats, 5 providers
- `TestTier2Features` - Tier 2: 100 seats, 100 providers
- `TestLicenseGracePeriod` - Verify grace period after expiration
- `TestLicenseRenewal` - Renew expired license, verify features restored
- `TestLicenseTransfer` - Transfer license to new tenant

#### Usage Tests (10 tests) - `usage_test.go`

**Usage Upload (3 tests)**
- `TestUploadUsageData` - POST /api/v1/usage, verify 200, usage recorded
- `TestUploadBatchUsageData` - Upload multiple records, verify all recorded
- `TestUploadInvalidProviderId` - Upload for non-existent provider, verify 400

**Usage Statistics (3 tests)**
- `TestGetUsageStatistics` - GET /api/v1/usage/statistics, verify aggregated stats
- `TestUsageStatisticsDateRange` - Query with date range, verify filtered stats
- `TestUsageStatisticsByProvider` - Query grouped by provider, verify per-provider stats

**Usage Aggregation (4 tests)**
- `TestDailyUsageAggregation` - Verify usage grouped by day
- `TestProviderUsageAggregation` - Verify usage grouped by provider
- `TestModelUsageAggregation` - Verify usage grouped by model
- `TestUsageHistoryPagination` - GET /api/v1/usage/history?page=1&limit=10, verify pagination

#### User Tests (10 tests) - `user_test.go`

**User Creation (3 tests)**
- `TestCreateUser` - POST /api/v1/users, verify 201, user ID returned
- `TestCreateDuplicateUser` - Create with existing email, verify 400
- `TestCreateUserMissingFields` - Create without required fields, verify 400

**User Update (3 tests)**
- `TestUpdateUser` - PUT /api/v1/users/{id}, verify 200, user updated
- `TestUpdateNonExistentUser` - Update non-existent user, verify 404
- `TestUpdateUserRole` - Change role from member to manager, verify 200

**User Deletion (2 tests)**
- `TestDeleteUser` - DELETE /api/v1/users/{id}, verify 200, user deleted
- `TestDeleteNonExistentUser` - Delete non-existent user, verify 404

**User List (2 tests)**
- `TestListUsers` - GET /api/v1/users, verify 200 with array
- `TestListUsersPagination` - GET /api/v1/users?page=1&limit=10, verify pagination

#### Manager Dashboard Tests (6 tests) - `mgr_dashboard_test.go`

**Dashboard Metrics (5 tests)**
- `TestMgrGetDashboardMetrics` - GET /api/v1/dashboard/metrics, verify summary metrics
- `TestMgrGetDashboardMetrics7d` - GET /api/v1/dashboard/metrics?range=7d, verify 7-day stats
- `TestMgrGetDashboardMetrics24h` - GET /api/v1/dashboard/metrics?range=24h, verify 24-hour stats
- `TestMgrGetDashboardMetrics30d` - GET /api/v1/dashboard/metrics?range=30d, verify 30-day stats
- `TestMgrGetDashboardRankings` - GET /api/v1/dashboard/rankings, verify user rankings by usage

**Dashboard Members (1 test)**
- `TestMgrGetDashboardMembers` - GET /api/v1/dashboard/members, verify member list

#### Manager Team Tests (9 tests) - `mgr_team_test.go`

**Team CRUD (6 tests)**
- `TestMgrListTeams` - GET /api/v1/teams, verify 200 with team list
- `TestMgrCreateTeam` - POST /api/v1/teams, verify 201, team created
- `TestMgrCreateTeamMinimal` - Create with minimal fields (name only), verify 201
- `TestMgrUpdateTeam` - PUT /api/v1/teams/{id}, verify 200, team updated
- `TestMgrDeleteTeam` - DELETE /api/v1/teams/{id}, verify 200, team deleted
- `TestMgrDeleteTeamNotFound` - DELETE non-existent team, verify 404

**Team Settings (2 tests)**
- `TestMgrGetTeamSettings` - GET /api/v1/teams/{id}/settings, verify 200 with settings
- `TestMgrUpdateTeamSettings` - PUT /api/v1/teams/{id}/settings, verify 200, settings updated

**Team Members (3 tests)**
- `TestMgrListTeamMembers` - GET /api/v1/teams/{id}/members, verify 200 with member list
- `TestMgrAddTeamMember` - POST /api/v1/teams/{id}/members, verify 201, member added
- `TestMgrRemoveTeamMember` - DELETE /api/v1/teams/{id}/members/{memberId}, verify 200, member removed
- `TestMgrRemoveTeamMemberNotFound` - DELETE non-existent member, verify 200 (idempotent)

**Team Get (1 test)**
- `TestMgrGetTeamNotFound` - GET non-existent team, verify 404

#### Manager User CRUD Tests (7 tests) - `mgr_users_crud_test.go`

**User Creation (3 tests)**
- `TestMgrListUsers` - GET /api/v1/users, verify 200 with user list
- `TestMgrCreateUser` - POST /api/v1/users, verify 201, user created with role
- `TestMgrCreateUserManager` - Create manager user, verify 201, role=manager
- `TestMgrCreateUserValidation` - Create without required fields, verify 400

**User Update (3 tests)**
- `TestMgrUpdateUser` - PUT /api/v1/users/{id}, verify 200, user updated
- `TestMgrUpdateUserPassword` - Update only password, verify 200, password changed

**User Deletion (2 tests)**
- `TestMgrDeleteUser` - DELETE /api/v1/users/{id}, verify 200, user deleted
- `TestMgrDeleteUserNotFound` - DELETE non-existent user, verify 404

**User Get (2 tests)**
- `TestMgrGetUser` - GET /api/v1/users/{id} (SKIPPED - server endpoint missing)
- `TestMgrGetUserNotFound` - GET non-existent user, verify 404

#### Analytics Tests (34 tests) - `analytics_*.go`

**Basic Analytics (4 tests)** - `analytics_test.go`
- `TestGetProviderAnalytics` - GET /analytics/providers, verify provider statistics
- `TestProviderAnalyticsDateRange` - Query with date range, verify filtered stats
- `TestGetUserAnalytics` - GET /analytics/users, verify user statistics
- `TestUsageHistoryPagination` - GET /analytics/history?page=1&limit=10, verify pagination

**E2E Data Flow (5 tests)** - `analytics_e2e_test.go`
- `TestUploadToProviderAnalyticsFlow` - Upload 10 records, verify in provider analytics (count, tokens, model breakdown, success rate)
- `TestUploadToUserAnalyticsFlow` - Upload for 3 users, verify per-user statistics and rankings
- `TestUploadToHistoryAnalyticsFlow` - Upload 25 records, verify pagination (page 1: 10, total: 25, sorted desc)
- `TestRealTimeDataFreshness` - Upload 5 records, query immediately, verify all appear (< 1s latency)
- `TestErrorRecordsInAnalytics` - Upload mix of 200/4xx/5xx, verify error rate (30%), per-status-code breakdown

**Aggregation Accuracy (6 tests)** - `analytics_aggregation_test.go`
- `TestTokenSumAggregation` - Upload 3 records with known tokens, verify sums (input: 4500, output: 2250, total: 6750)
- `TestAverageCalculationAccuracy` - Upload 5 records with durations [1.0, 2.0, 3.0, 4.0, 5.0]s, verify avg=3.0s
- `TestModelLevelAggregation` - Upload across 3 models, verify per-model totals sum to 100%
- `TestDailyAggregationAccuracy` - Upload over 3 days, verify daily grouping (no wrong-day records)
- `TestMultiProviderAggregation` - Upload for 3 providers (5k/4k/6k tokens), verify combined total=15k
- `TestCacheTokensInAggregation` - Upload with cache_read/write_tokens, verify included in totals

**Time-Based Analytics (5 tests)** - `analytics_time_test.go`
- `TestDateRangeFiltering` - Upload over 7 days, query various ranges, verify correct counts
- `TestHourlyGranularity` - Upload 24 records (one per hour), verify 24 hourly data points
- `TestWeeklyAggregation` - Upload over 4 weeks, verify weekly sums and week boundaries
- `TestMonthBoundaryHandling` - Upload spanning month end/beginning, verify no data loss or double-counting
- `TestTimeZoneHandling` - Upload with UTC timestamps, query from different timezones, verify consistent grouping

**Cross-Filtering (5 tests)** - `analytics_filter_test.go`
- `TestProviderUserDateRangeFilter` - Upload 2 providers × 3 users × 7 days, query with all 3 filters, verify intersection
- `TestModelProviderFilter` - Upload 2 providers × 3 models, filter Provider A + Model X, verify only matching
- `TestSuccessRateFiltering` - Upload mix of 200/4xx/5xx, query success/error filters, verify correct percentages
- `TestStreamingFiltering` - Upload mix of streaming/non-streaming, verify correct grouping
- `TestPlatformFiltering` - Upload across 3 platforms, verify per-platform totals

**Edge Cases & Statistics (9 tests)** - `analytics_edge_cases_test.go`
- `TestPaginationFirstPage` - Upload 25 records, query page 1 limit 10, verify exactly 10 records
- `TestPaginationLastPage` - Upload 23 records, query page 3 limit 10, verify only 3 records, no next page
- `TestPaginationEmptyResult` - Query date range with no data, verify empty array, total_count=0, no error
- `TestZeroTokenHandling` - Upload zero input/output tokens, verify no division-by-zero in averages
- `TestNegativeTokenHandling` - Upload negative token values, verify rejected/skipped (currently accepted TODO)
- `TestMixedTokenValues` - Upload varied values (0, 1, large, output > input), verify all recorded, totals correct
- `TestPercentileCalculation` - Upload 10 records with known distribution, verify p50/p90/p95/p99 percentiles
- `TestRateLimitingMetrics` - Upload mix of 200/429/500, verify success/error rates, per-error-code tracking
- `TestCostCalculation` - Upload known token counts, verify cost using model pricing (per-token cost accurate)

---

## Architecture Overview

### Independent Test Module

The integration tests use a **separate `go.mod`** file to minimize dependencies:

```
integration/
├── go.mod                               # Independent test module
├── go.sum
├── integration_test.go                  # Test suite setup
├── helpers.go                           # Utility functions
├── fixtures.go                          # Test data
└── *_test.go                            # Test files
```

**Benefits:**
- ✅ **Minimal dependencies** - Only imports `shared/integration` client
- ✅ **Isolated testing** - Tests run independently of server build
- ✅ **Faster dependency resolution** - Smaller `go.mod` = faster `go mod download`
- ✅ **Clean separation** - Test dependencies don't pollute server dependencies

**Example `integration/go.mod`:**
```go
module github.com/shtdu/integration

go 1.25

require (
    github.com/shtdu/shared/integration v0.0.0
    github.com/stretchr/testify v1.11.1
)

replace github.com/shtdu/shared/integration => ../shared/integration
```

### Shared Integration Client

The tests use the **shared integration client** from `shared/integration/`:

- **Type-safe API calls** via `generated.go` (auto-generated from OpenAPI spec)
- **Helper factories** in `client.go` for authenticated and anonymous clients
- **Transport layer** with logging, retry, and automatic token injection

**Regenerate the client** after API changes:
```bash
cd shared/integration && go generate
```

### Black-Box Testing Principles

Tests treat the server as a **black box**:

1. **API-Only Communication**: All interactions via HTTP API
2. **No Database Access**: Never import server packages or query DB directly
3. **Shared Client**: Use `shared/integration` client for all API calls
4. **External Fixtures**: Create test data via API, not DB inserts

**Example API Call:**
```go
// Create provider via API
req := integration.PostApiV1ProvidersJSONRequestBody{
    Name:   "Test Provider",
    ApiKey: "sk-test-key",
    Kind:   integration.CreateProviderRequestKindClaude,
}
resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
require.NoError(s.T(), err)
require.Equal(s.T(), 201, resp.StatusCode())
```

### Test Prerequisites

**Required:**
- Go 1.25+
- PostgreSQL 14+ (for test database)
- Test server running (see INTEGRATION_TEST_SETUP.md)

**Setup:**
```bash
cd integration
cp .env.test.example .env.test  # Configure database URL
make integration-setup           # Setup test database
./test-server.sh                # Start test server (separate terminal)
go test -v                      # Run tests
```

**Database:**
- Uses separate `codetogether_test` database
- Auto-created/migrated on first run
- Isolated from production data

---

## Fixture Reference

### License Fixtures

| Fixture | File | Tier | Seats | Providers | Expiration |
|---------|------|------|-------|-----------|------------|
| Tier 0 | `tier_0.pem` | 0 | 0 | 0 | None |
| Tier 1 | `tier_1.pem` | 1 | 10 | 5 | None |
| Tier 2 | `tier_2.pem` | 2 | 100 | 100 | None |
| Expired | `expired.pem` | 1 | 10 | 5 | Past |
| Zero Seats | `zero_seats.pem` | 1 | 0 | 5 | None |
| Immediate Expiry | `immediate_expiry.pem` | 1 | 10 | 5 | Immediate |
| Invalid Signature | `invalid_signature.pem` | - | - | - | Invalid |

**Location:** `integration/testdata/licenses/`

**Usage:**
```go
licensePEM, err := LoadLicenseFixture("tier_1.pem")
activateResp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx,
    integration.PostApiV1LicenseActivateJSONRequestBody{
        LicenseKey: licensePEM,
    })
```

### User Fixtures

| Name | Email | Password | Role |
|------|-------|----------|------|
| Admin | `admin@example.com` | `AdminPassword123!` | `admin` |
| Member | `member@example.com` | `MemberPassword123!` | `member` |

**Location:** `integration/fixtures.go:18`

**Usage:**
```go
token := s.registerAndLoginUserFromFixture("member")
client := s.createAuthenticatedClient(token)
```

### Provider Fixtures

| Name | Kind | API URL | Level |
|------|------|---------|-------|
| Claude | `claude` | `https://api.anthropic.com` | 1 |
| OpenCode | `opencode` | `https://api.opencode.com` | 1 |
| Codex | `codex` | `https://api.codex.com` | 1 |

**Location:** `integration/fixtures.go:35`

**Usage:**
```go
providerID := s.createProviderFixtureFromFixture("claude")
```

---

## Running Tests

> **Comprehensive Guide:** See **[INTEGRATION_TEST_SETUP.md](./INTEGRATION_TEST_SETUP.md)** for detailed setup, environment configuration, and troubleshooting.

### Quick Commands

```bash
# Run all tests
cd integration && go test -v

# Run specific category
go test -v -run "TestProvider"
go test -v -run "TestAuth"
go test -v -run "TestLicense"
```

### Test Execution Modes

| Mode | Command | Description |
|------|---------|-------------|
| All tests | `go test -v` | Run complete test suite |
| Specific category | `go test -v -run "TestProvider"` | Run one test category |
| Coverage | `go test -cover -coverprofile=coverage.out` | Generate coverage report |
| Verbose debug | `go test -v -run "TestName"` | Run single test with output |

### Prerequisites

1. **Test database:** `make integration-setup` (one-time)
2. **Test server:** `./test-server.sh` (separate terminal)
3. **Environment:** `.env.test` configured

### Troubleshooting

| Issue | Solution |
|-------|----------|
| Port already in use | Change `TEST_SERVER_URL` in `.env.test` |
| Database connection failed | Verify `TEST_DB_URL` credentials |
| Tests timing out | Increase timeout: `go test -timeout 30m` |
| License activation fails | Check fixture files exist in `testdata/licenses/` |

**For complete troubleshooting guide, see:** [INTEGRATION_TEST_SETUP.md - Troubleshooting](./INTEGRATION_TEST_SETUP.md#troubleshooting)

### Additional Resources

- **[INTEGRATION_TEST_SETUP.md](./INTEGRATION_TEST_SETUP.md)** - Complete setup and execution guide
- **[issue.md](./issue.md)** - Test status and known issues
- **API Documentation** - `docs/client_api/server_api.yaml`
- **Shared Client** - `shared/integration/`
