# BDD Test Refactoring: Mocks → Real API Calls

**Status:** Draft
**Date:** 2025-03-17
**Author:** Claude Code
**Related:** BDD test suite improvement

## Executive Summary

Convert BDD test step definitions from mock implementations to real HTTP API calls against the test server. This refactoring will start with smoke-tagged scenarios and expand to all 215 scenarios incrementally.

**Key Changes:**
- Replace mock responses with real API client calls
- Initialize API clients using `github.com/code-together/shared/integration`
- Implement API-based test data lifecycle (create/cleanup via API)
- Fail fast on API errors to ensure test quality

## Problem Statement

The BDD test suite currently has 215 scenarios passing with 100% success rate, but all step implementations use mock data instead of real API calls. This means:

1. **False confidence:** Tests pass even when API endpoints are broken
2. **No integration validation:** Real API contracts aren't tested
3. **Manual testing required:** Real server behavior must be tested separately
4. **Maintenance burden:** Mock implementations must stay in sync with real APIs

The `bdd-test.sh` script already manages a real test server on port 8088, but the BDD scenarios don't use it.

## Proposed Solution

### Architecture Overview

**Current State (Mock-based):**
```
Scenario → Step Definition → ctx.SetLastResponse(status, mockData, "")
                                      ↓
                              Mock response stored in context
```

**Target State (API-based):**
```
Scenario → Step Definition → API Client → HTTP Request → Test Server (port 8088)
                                      ↓
                              Real response stored in context
```

### Components

#### 1. Client Factory (`support/client_factory.go` - NEW)

Creates and configures API client instances.

**Responsibilities:**
- Create anonymous client for auth endpoints (login, register)
- Create authenticated client for protected endpoints
- Manage token injection for authenticated requests
- Configure base URL, timeouts, and retry logic

**Interface:**
```go
package support

import (
    "log/slog"
    "github.com/code-together/shared/integration"
)

// NewAnonymousClient creates a client without authentication
func NewAnonymousClient(serverURL string, logger *slog.Logger) (*integration.ClientWithResponses, error) {
    return integration.NewAnonymousClient(serverURL, logger, false)
}

// NewAuthenticatedClient creates a client with automatic token injection
func NewAuthenticatedClient(serverURL string, getToken func() (string, error), logger *slog.Logger) (*integration.ClientWithResponses, error) {
    return integration.NewAuthenticatedClient(serverURL, getToken, logger, false)
}
```

**Note:** The `verbose` parameter defaults to `false` for minimal logging. Set to `true` for debugging.

#### 2. Test Context Updates (`support/test_context.go` - MODIFY)

Replace interface{} types with actual API client types.

**Changes:**
- `AnonymousClient interface{}` → `*integration.ClientWithResponses` (for unauthenticated requests)
- `Client interface{}` → `*integration.ClientWithResponses` (authenticated with current user token)
- `ManagerClient interface{}` → `*integration.ClientWithResponses` (authenticated with admin/manager token)

**Client Architecture:**
The three client types correspond to different authentication contexts:
- **AnonymousClient**: Used for login, registration, and public endpoints
- **Client**: Authenticated with the current scenario user's token (for member-level operations)
- **ManagerClient**: Authenticated with admin/manager token (for administrative operations)

All three are the same underlying type (`*integration.ClientWithResponses`) but are created with different token getters to match the authentication context required by different scenarios.

**New Methods:**
```go
// InitializeClients sets up all API clients
func (ctx *BDDTestContext) InitializeClients(serverURL string, logger *slog.Logger) error

// GetAuthToken returns the current authentication token
func (ctx *BDDTestContext) GetAuthToken() (string, error)

// GetAuthenticatedClient returns a client with automatic token injection
func (ctx *BDDTestContext) GetAuthenticatedClient() (*integration.ClientWithResponses, error)
```

#### 3. Step Definition Refactoring

Replace mock implementations with real API calls across all step definition files.

**Files to Update:**
- `auth_steps.go` - Login, logout, refresh, verify tokens
- `common_steps.go` - Health checks
- `provider_steps.go` - CRUD operations
- `license_steps.go` - Activation, status checks
- `usage_steps.go` - Statistics retrieval
- `dashboard_steps.go` - Analytics data
- `permission_steps.go` - RBAC verification
- `health_steps.go` - System health
- `resource_tracking.go` - Cleanup via API

**Example Conversion (auth_steps.go):**

**Before (Mock):**
```go
func (ctx *ScenarioContext) iLoginWithCredentials(email, password string) error {
    if email == "nonexistent@example.com" {
        ctx.SetLastResponse(401, nil, "user_not_found")
        return nil
    }
    // Mock success response
    ctx.SetLastResponse(200, map[string]string{
        "token": fmt.Sprintf("mock-token-%s", email),
        "email": email,
    }, "")
    return nil
}
```

**After (Real API):**
```go
func (ctx *ScenarioContext) iLoginWithCredentials(email, password string) error {
    client := ctx.GetAnonymousClient()

    req := integration.PostAuthLoginJSONRequestBody{
        Email:    email,
        Password: password,
    }

    resp, err := client.PostAuthLoginWithResponse(context.Background(), req)
    if err != nil {
        return fmt.Errorf("login request failed: %w", err)
    }

    // Store response for assertions
    ctx.SetLastResponse(resp.StatusCode(), resp.JSON200, "")

    // Store token if successful
    if resp.StatusCode() == 200 && resp.JSON200 != nil {
        ctx.AdminToken = *resp.JSON200.Token
        ctx.CurrentUser = &support.UserInfo{
            Email: email,
            Token: *resp.JSON200.Token,
        }
    }

    return nil
}
```

#### 4. Test Data Lifecycle (`godog_suite_test.go` - MODIFY)

Implement API-based test data creation and cleanup.

**BeforeScenario Hook:**
```go
ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    testContext.ResetScenarioState()

    // Ensure test server is running
    if !support.IsTestServerRunning(testContext.ServerURL) {
        return ctx, fmt.Errorf("test server not running at %s", testContext.ServerURL)
    }

    // Load fixtures and create real data via API
    fixtures, err := support.LoadFixtureData()
    if err != nil {
        return ctx, fmt.Errorf("failed to load fixtures: %w", err)
    }

    // Create test data via API
    if err := support.CreateTestDataFromFixtures(testContext, fixtures); err != nil {
        return ctx, fmt.Errorf("failed to create test data: %w", err)
    }

    return ctx, nil
})
```

**CreateTestDataFromFixtures Implementation:**

This function creates test resources via API calls based on fixture data.

```go
// CreateTestDataFromFixtures creates test resources from fixture data via API
func CreateTestDataFromFixtures(ctx *BDDTestContext, fixtures *FixtureData) error {
    client := ctx.GetAnonymousClient()

    // Create users and track IDs
    for _, user := range fixtures.Users {
        // Skip if user already exists (check by email)
        // Register/login user via API
        req := integration.PostAuthRegisterJSONRequestBody{
            Email:    user.Email,
            Password: user.Password,
            Name:     user.Name,
        }

        resp, err := client.PostAuthRegisterWithResponse(context.Background(), req)
        if err != nil {
            return fmt.Errorf("failed to create user %s: %w", user.Email, err)
        }

        if resp.StatusCode() == 201 || resp.StatusCode() == 200 {
            if resp.JSON200 != nil && resp.JSON200.ID != nil {
                ctx.TrackUser(*resp.JSON200.ID)
            }
        } else if resp.StatusCode() != 409 { // 409 = already exists, which is OK
            return fmt.Errorf("unexpected status creating user %s: %d", user.Email, resp.StatusCode())
        }
    }

    // Create providers and track IDs
    authenticatedClient := ctx.GetAuthenticatedClient()
    for _, provider := range fixtures.Providers {
        req := integration.PostApiV1ProvidersJSONRequestBody{
            Name:  provider.Name,
            Kind:  provider.Kind,
            ApiKey: provider.APIKey,
        }

        resp, err := authenticatedClient.PostApiV1ProvidersWithResponse(context.Background(), req)
        if err != nil {
            return fmt.Errorf("failed to create provider %s: %w", provider.Name, err)
        }

        if resp.StatusCode() == 201 && resp.JSON201 != nil {
            ctx.TrackProvider(resp.JSON201.ID)
        }
    }

    // Create teams and track IDs (skip default team ID 1)
    for _, team := range fixtures.Teams {
        if team.ID == 1 {
            continue // Default team already exists
        }

        req := integration.PostApiV1TeamsJSONRequestBody{
            Name: team.Name,
        }

        resp, err := authenticatedClient.PostApiV1TeamsWithResponse(context.Background(), req)
        if err != nil {
            return fmt.Errorf("failed to create team %s: %w", team.Name, err)
        }

        if resp.StatusCode() == 201 && resp.JSON201 != nil {
            ctx.TrackTeam(resp.JSON201.ID)
        }
    }

    return nil
}
```

**Data Structures:**

```go
// FixtureData represents test fixture data loaded from JSON
type FixtureData struct {
    Users     []TestFixtureUser    `json:"users"`
    Providers []TestFixtureProvider `json:"providers"`
    Teams     []TestFixtureTeam    `json:"teams"`
    Licenses  []TestFixtureLicense `json:"licenses"`
}

type TestFixtureUser struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
    Role     string `json:"role"`
}

type TestFixtureProvider struct {
    Name  string `json:"name"`
    Kind  string `json:"kind"`
    APIKey string `json:"api_key"`
}

type TestFixtureTeam struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

type TestFixtureLicense struct {
    Key  string `json:"key"`
    Tier string `json:"tier"`
}
```

**AfterScenario Hook:**
```go
ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
    // Clean up resources via API
    cleanupErr := testContext.CleanupScenarioResources()
    if cleanupErr != nil {
        log.Printf("WARNING: Cleanup failed: %v", cleanupErr)
    }

    return ctx, err
})
```

**Resource Tracking (`resource_tracking.go`):**

The `BDDTestContext` tracks created resources using slices:
- `createdProviders []int64` - Provider IDs
- `createdUsers []string` - User IDs (UUID strings)
- `createdTeams []int64` - Team IDs

```go
// CleanupScenarioResources deletes all tracked resources via API
func (ctx *ScenarioContext) CleanupScenarioResources() error {
    client := ctx.GetAuthenticatedClient()

    // Clean up providers (returns []int64)
    for _, providerID := range ctx.GetCreatedProviders() {
        _, err := client.DeleteApiV1ProvidersProviderIdWithResponse(
            context.Background(), providerID)
        if err != nil {
            log.Printf("WARNING: Failed to delete provider %d: %v", providerID, err)
        }
    }

    // Clean up users (returns []string - UUIDs)
    for _, userID := range ctx.GetCreatedUsers() {
        // Note: User deletion may require admin privileges
        _, err := client.DeleteApiV1UsersUserIdWithResponse(
            context.Background(), userID)
        if err != nil {
            log.Printf("WARNING: Failed to delete user %s: %v", userID, err)
        }
    }

    // Clean up teams (returns []int64, skip default team ID 1)
    for _, teamID := range ctx.GetCreatedTeams() {
        if teamID == 1 {
            continue
        }
        _, err := client.DeleteApiV1TeamsTeamIdWithResponse(
            context.Background(), teamID)
        if err != nil {
            log.Printf("WARNING: Failed to delete team %d: %v", teamID, err)
        }
    }

    ctx.ClearCreatedResources()
    return nil
}
```

### Implementation Phases

#### Phase 1: Infrastructure Setup

**Tasks:**
1. Create `support/client_factory.go`
2. Update `support/test_context.go` with real client types
3. Add `support/api_helpers.go` for response handling
4. Update `godog_suite_test.go` to initialize clients
5. Add error handling helpers

**Deliverable:** Foundation for API-based testing

#### Phase 2: Smoke Test Conversion

**Tasks:**
1. Identify smoke test scenarios (see methodology below)
2. Convert health check steps (simplest - no auth required)
3. Convert authentication steps (enables authenticated requests)
4. Convert remaining smoke test step definitions
5. Update smoke test scenarios if needed
6. Run smoke tests and fix failures

**Deliverable:** All smoke tests using real API calls

**Smoke Test Identification:**

Find all smoke test scenarios using:
```bash
# Grep for @smoke tag
grep -r "@smoke" bdd/features/

# Expected results:
# - features/hello_world.feature (all scenarios - infrastructure validation)
# - Scenarios tagged @smoke in other feature files
```

**Smoke Test Scope:**
- All scenarios in `features/hello_world.feature` (health check, basic connectivity)
- Any scenario in other feature files tagged with `@smoke`
- These validate core infrastructure is working before running full test suite

#### Phase 3: Full Conversion (Future)

**Tasks:**
1. Convert remaining step definitions by domain:
   - Identity & Access (`features/00_identity_and_access.feature`)
   - Provider Management (`features/01_provider_management.feature`)
   - Usage Insights (`features/03_usage_insights.feature`)
   - User Interfaces (`features/04_user_interfaces.feature`)
   - System Behaviors (`features/05_system_behaviors.feature`)
2. Remove all mock implementations
3. Update CLAUDE.md documentation
4. Verify all 215 scenarios pass with real APIs

**Deliverable:** Complete API-based BDD test suite

### Error Handling Strategy

**Fail Fast Approach:**
- API errors (4xx, 5xx) immediately fail the step
- Clear error messages with request context
- Test run stops on first failure for smoke tests
- Full test run continues but reports all failures

**Response Handler:**
```go
// handleAPIResponse processes API response and returns error if unsuccessful
func handleAPIResponse(statusCode int, body []byte, apiError *integration.APIError) error {
    if apiError != nil {
        return fmt.Errorf("API call failed: %s (code: %s, status: %d)",
            apiError.Message, apiError.ErrorCode, apiError.StatusCode)
    }

    if statusCode >= 400 {
        errMsg := string(body)
        if errMsg == "" {
            errMsg = http.StatusText(statusCode)
        }
        return fmt.Errorf("API returned error %d: %s", statusCode, errMsg)
    }

    return nil
}
```

**Usage Example:**
```go
resp, err := client.PostAuthLoginWithResponse(context.Background(), req)

// Check for network/transport errors
if err != nil {
    return fmt.Errorf("login request failed: %w", err)
}

// Check for API errors
if err := handleAPIResponse(resp.StatusCode(), resp.Body, resp.HTTPResponse); err != nil {
    return err
}

// Success - store response
ctx.SetLastResponse(resp.StatusCode(), resp.JSON200, "")
```

### Test Data Management

**Creation Strategy:**

Test data is created **per scenario** (not per test run or per feature file). This ensures:
- Each scenario has isolated test data
- No state leakage between scenarios
- Scenarios can run in any order
- Failed scenarios don't affect others

**Performance Considerations:**
- Per-scenario creation has more API overhead but ensures isolation
- If performance becomes an issue, consider:
  - Using test database transactions for rollback (faster than API deletes)
  - Caching commonly used resources (e.g., default team)
  - Parallel scenario execution (requires thread-safe context)

**Fixture Loading:**
- Load test data from `bdd/support/fixtures.json`
- Parse fixture definitions (users, providers, teams, licenses)
- Create resources via API in BeforeScenario hook
- Track created resource IDs in context

**Cleanup Strategy:**
- Delete resources in reverse creation order
- Skip default resources (team ID 1, system users)
- Log cleanup failures but don't fail scenario
- Clear tracking maps after cleanup

**Example Fixture Data:**
```json
{
  "users": [
    {
      "email": "test-manager@example.com",
      "password": "TestPassword123!",
      "role": "manager",
      "name": "Test Manager"
    }
  ],
  "providers": [
    {
      "name": "test-claude-provider",
      "kind": "claude",
      "api_key": "sk-test-claude-123"
    }
  ]
}
```

## Trade-offs

### Chosen Approach: Refactor In-Place

**Advantages:**
- ✅ No code duplication during migration
- ✅ Clear migration path
- ✅ Single source of truth after completion
- ✅ Forces immediate correctness (no crutches)

**Disadvantages:**
- ❌ Risk of breaking tests during conversion
- ❌ Cannot compare mock vs real behavior
- ❌ Requires careful step-by-step migration

### Alternative 1: Parallel Implementations

**Approach:** Create new files like `auth_steps_real.go` alongside existing mocks.

**Advantages:**
- ✅ Safer migration (can run both versions)
- ✅ Easy to compare mock vs real behavior
- ✅ Rollback is trivial

**Disadvantages:**
- ❌ Code duplication during migration
- ❌ Maintenance burden
- ❌ Confusing to have both implementations
- ❌ Need to update step registration

**Why Not Chosen:** The test suite already has 100% success rate with mocks, so parallel implementations don't provide value. The refactoring can be done incrementally by feature file.

### Alternative 2: Adapter Pattern

**Approach:** Create a backend interface that can swap between mock and real implementations.

**Advantages:**
- ✅ Most flexible approach
- ✅ Can switch via config/env var
- ✅ Useful for local development without server

**Disadvantages:**
- ❌ Adds complexity (new interface layer)
- ❌ Overkill for this use case
- ❌ Indirection makes debugging harder
- ❌ More code to maintain

**Why Not Chosen:** The `bdd-test.sh` script already manages the test server lifecycle. Having real tests always use the real server is the goal. Mocks were only a temporary measure.

## Success Criteria

### Phase 1 (Infrastructure)
- [ ] Client factory created and tested
- [ ] Test context updated with real client types
- [ ] Helper functions for response handling
- [ ] Documentation updated

### Phase 2 (Smoke Tests)
- [ ] All smoke test scenarios identified
- [ ] Health check steps use real API
- [ ] Authentication steps use real API
- [ ] All smoke tests pass with real server
- [ ] Zero mock implementations in smoke test steps
- [ ] Test server required for smoke test execution

### Phase 3 (Full Conversion - Future)
- [ ] All 215 scenarios use real API calls
- [ ] Zero mock implementations in codebase
- [ ] Test data lifecycle fully API-based
- [ ] CLAUDE.md documentation updated
- [ ] All scenarios pass with real server

## Risks and Mitigations

### Risk 1: API Endpoints Not Implemented

**Risk:** Some BDD scenarios test endpoints that don't exist yet in the server.

**Mitigation:**
- Start with smoke tests (most likely to be implemented)
- Fail fast and fix server implementation before proceeding
- Document which endpoints are missing

### Risk 2: Test Data Cleanup Failures

**Risk:** Resources aren't cleaned up properly, causing test pollution.

**Mitigation:**
- Log cleanup failures prominently
- Implement retry logic for cleanup
- Provide manual cleanup script if needed
- Consider using test database transactions for rollback

### Risk 3: Flaky Tests Due to Timing

**Risk:** Tests fail intermittently due to server startup or network issues.

**Mitigation:**
- Implement proper health checks in BeforeScenario
- Use adequate timeouts for API calls
- Retry transient failures with exponential backoff
- Log detailed timing information for debugging

### Risk 4: Breaking Existing Functionality

**Risk:** Refactoring breaks currently passing tests.

**Mitigation:**
- Start with smoke tests (smaller surface area)
- Run tests after each step definition file conversion
- Keep git history clean for easy revert
- Validate against real server behavior

## Implementation Notes

### Dependencies

The BDD module already has the required dependency:
```go
require github.com/code-together/shared => ../shared
```

The `shared/integration` package provides:
- `ClientWithResponses` - Generated OpenAPI client type
- `integration.NewAnonymousClient(baseURL, logger, verbose bool)` - Client factory for auth
- `integration.NewAuthenticatedClient(baseURL, getToken, logger, verbose bool)` - Client factory with token injection
- `APIError` - Error type with `Message`, `ErrorCode`, `StatusCode` fields

### Environment Variables

No new environment variables required. Existing variables will be used:
- `TEST_SERVER_URL` - Default: `http://localhost:8088`
- `TEST_DATABASE_URL` - For test data setup if needed

### Testing the Conversion

To validate the conversion works:

```bash
# Start test server
cd ../integration && ./test-server.sh

# Run smoke tests only
cd ../bdd
./bdd-test.sh --tags "@smoke"

# Run single feature for debugging
cd godog && go test -v -godog.paths="../features/hello_world.feature"
```

## Related Documentation

- `bdd/CLAUDE.md` - BDD test suite documentation
- `bdd/bdd-test.sh` - Test execution script
- `shared/integration/client.go` - API client factory
- `integration/INTEGRATION_TEST_SETUP.md` - Test server setup

## Next Steps

1. Implement Phase 1 (Infrastructure Setup)
2. Convert and validate smoke tests (Phase 2)
3. Document findings and adjust approach if needed
4. Plan full conversion (Phase 3) based on learnings
