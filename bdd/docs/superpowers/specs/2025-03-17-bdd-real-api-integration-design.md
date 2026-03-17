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
- Add `Logger *slog.Logger` field for logging in client creation

**Client Architecture:**
The three client types correspond to different authentication contexts:
- **AnonymousClient**: Used for login, registration, and public endpoints
- **Client**: Authenticated with the current scenario user's token (for member-level operations)
- **ManagerClient**: Authenticated with admin/manager token (for administrative operations)

All three are the same underlying type (`*integration.ClientWithResponses`) but are created with different token getters to match the authentication context required by different scenarios.

**New Methods:**
```go
// InitializeClients sets up all API clients
// Called once during test suite initialization (not per scenario)
func (ctx *BDDTestContext) InitializeClients(serverURL string, logger *slog.Logger) error {
    // Store logger for client creation
    ctx.Logger = logger

    // Create anonymous client (no authentication required)
    anonClient, err := NewAnonymousClient(serverURL, logger)
    if err != nil {
        return fmt.Errorf("failed to create anonymous client: %w", err)
    }
    ctx.AnonymousClient = anonClient

    // Authenticated clients will be created per-scenario after user login
    // They use a token getter that reads from the context's current token
    ctx.Client = nil // Will be set in scenario after login
    ctx.ManagerClient = nil // Will be set in scenario after admin login

    return nil
}

// GetAnonymousClient returns the anonymous client
func (ctx *BDDTestContext) GetAnonymousClient() *integration.ClientWithResponses {
    return ctx.AnonymousClient
}

// GetAuthToken returns the current authentication token for the scenario
func (ctx *BDDTestContext) GetAuthToken() (string, error) {
    if ctx.AdminToken != "" {
        return ctx.AdminToken, nil
    }
    if ctx.MemberToken != "" {
        return ctx.MemberToken, nil
    }
    return "", fmt.Errorf("no authentication token available")
}

// GetAuthenticatedClient returns a client with automatic token injection
// Creates or reuses the authenticated client for the current scenario
func (ctx *BDDTestContext) GetAuthenticatedClient() (*integration.ClientWithResponses, error) {
    // Get current token
    token, err := ctx.GetAuthToken()
    if err != nil {
        return nil, fmt.Errorf("cannot create authenticated client without token: %w", err)
    }

    // Create or reuse authenticated client based on user role
    if ctx.CurrentUser != nil && ctx.CurrentUser.Role == "manager" {
        if ctx.ManagerClient == nil {
            client, err := NewAuthenticatedClient(ctx.ServerURL, func() (string, error) {
                return ctx.GetAuthToken()
            }, ctx.Logger)
            if err != nil {
                return nil, err
            }
            ctx.ManagerClient = client
        }
        return ctx.ManagerClient, nil
    }

    // Use Client for non-manager users
    if ctx.Client == nil {
        client, err := NewAuthenticatedClient(ctx.ServerURL, func() (string, error) {
            return ctx.GetAuthToken()
        }, ctx.Logger)
        if err != nil {
            return nil, err
        }
        ctx.Client = client
    }
    return ctx.Client, nil
}

// UpdateAuthenticatedClients refreshes the authenticated clients with a new token
// Called after login/logout to update token injection
func (ctx *BDDTestContext) UpdateAuthenticatedClients(token string) error {
    client, err := NewAuthenticatedClient(ctx.ServerURL, func() (string, error) {
        return token, nil
    }, ctx.Logger)
    if err != nil {
        return err
    }

    if ctx.CurrentUser != nil && ctx.CurrentUser.Role == "manager" {
        ctx.ManagerClient = client
    } else {
        ctx.Client = client
    }
    return nil
}
```

**Client Lifecycle:**
- **AnonymousClient**: Created once during test suite initialization, reused for all unauthenticated requests
- **Client/ManagerClient**: Created per-scenario after user login, recreated when token changes
- **Token Management**: Tokens stored in context (`AdminToken`, `MemberToken`) and injected via closure
- **No token refresh**: BDD scenarios are short-lived, tokens don't expire during scenario execution
- **Cleanup**: Clients are garbage collected when scenario context is reset

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

**Note:** `CleanupScenarioResources()` should be added to `step_definitions/context.go` as a method on `ScenarioContext`. The existing `resource_tracking.go` file may not exist yet or may only contain helper functions.

**Example Conversions:**

**Example 1: Simple Login (auth_steps.go)**

**Before (Mock):**
```go
func (ctx *ScenarioContext) iLoginWithCredentials(email, password string) error {
    if email == "nonexistent@example.com" {
        ctx.SetLastResponse(401, nil, "user_not_found")
        return nil
    }
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
    client := ctx.BDDTestContext.GetAnonymousClient()

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

**Example 2: Multiple API Calls (provider_steps.go)**

**Scenario:** Create provider, then verify it was created

**Before (Mock):**
```go
func (ctx *ScenarioContext) iCreateAProviderWithKindAndAPIKey(kind, apiKey string) error {
    providerID := int64(123)
    ctx.TrackProvider(providerID)
    ctx.LastProviderID = providerID
    ctx.SetLastResponse(201, map[string]interface{}{
        "id": providerID,
        "kind": kind,
    }, "")
    return nil
}

func (ctx *ScenarioContext) theProviderShouldBeCreated() error {
    if ctx.LastProviderID == 0 {
        return fmt.Errorf("provider was not created")
    }
    ctx.SetLastResponse(200, map[string]interface{}{
        "id": ctx.LastProviderID,
        "name": "test-provider",
    }, "")
    return nil
}
```

**After (Real API):**
```go
func (ctx *ScenarioContext) iCreateAProviderWithKindAndAPIKey(kind, apiKey string) error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return err
    }

    req := integration.PostApiV1ProvidersJSONRequestBody{
        Name:  support.GenerateUniqueProviderName("test"),
        Kind:  kind,
        ApiKey: apiKey,
    }

    resp, err := client.PostApiV1ProvidersWithResponse(context.Background(), req)
    if err != nil {
        return fmt.Errorf("create provider failed: %w", err)
    }

    ctx.SetLastResponse(resp.StatusCode(), resp.JSON201, "")

    // Track provider ID for cleanup and subsequent steps
    if resp.StatusCode() == 201 && resp.JSON201 != nil {
        ctx.TrackProvider(resp.JSON201.ID)
        ctx.LastProviderID = resp.JSON201.ID
    }

    return nil
}

func (ctx *ScenarioContext) theProviderShouldBeCreated() error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return err
    }

    if ctx.LastProviderID == 0 {
        return fmt.Errorf("no provider ID available - was it created?")
    }

    resp, err := client.GetApiV1ProvidersProviderIdWithResponse(
        context.Background(), ctx.LastProviderID)
    if err != nil {
        return fmt.Errorf("get provider failed: %w", err)
    }

    ctx.SetLastResponse(resp.StatusCode(), resp.JSON200, "")
    return nil
}
```

**Example 3: Response Data Extraction (usage_steps.go)**

**Scenario:** Create usage record, then verify stats include it

**Before (Mock):**
```go
func (ctx *ScenarioContext) iGetMyUsageStatistics() error {
    ctx.SetLastResponse(200, map[string]interface{}{
        "total_tokens": 1000,
    }, "")
    return nil
}

func (ctx *ScenarioContext) theStatisticsShouldIncludeField(field string) error {
    _, resp, _ := ctx.GetLastResponse()
    stats, ok := resp.(map[string]interface{})
    if !ok {
        return fmt.Errorf("response is not a map")
    }
    if _, exists := stats[field]; !exists {
        return fmt.Errorf("field %s not found in stats", field)
    }
    return nil
}
```

**After (Real API):**
```go
func (ctx *ScenarioContext) iGetMyUsageStatistics() error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return err
    }

    resp, err := client.GetApiV1UsageStatsWithResponse(context.Background())
    if err != nil {
        return fmt.Errorf("get usage stats failed: %w", err)
    }

    ctx.SetLastResponse(resp.StatusCode(), resp.JSON200, "")
    return nil
}

func (ctx *ScenarioContext) theStatisticsShouldIncludeField(field string) error {
    _, resp, _ := ctx.GetLastResponse()

    stats, ok := resp.(*integration.GetApiV1UsageStatsResponse)
    if !ok || stats.JSON200 == nil {
        return fmt.Errorf("invalid response format")
    }

    // Use reflection or type assertions to check field existence
    // Example for specific fields:
    switch field {
    case "total_tokens":
        if stats.JSON200.TotalTokens == nil {
            return fmt.Errorf("field %s not found in response", field)
        }
    case "total_requests":
        if stats.JSON200.TotalRequests == nil {
            return fmt.Errorf("field %s not found in response", field)
        }
    default:
        return fmt.Errorf("unknown field: %s", field)
    }

    return nil
}
```

**Example 4: Async Operations (license_steps.go)**

**Scenario:** Activate license and poll for status

**Before (Mock):**
```go
func (ctx *ScenarioContext) iActivateALicenseWithKey(key string) error {
    ctx.SetLastResponse(200, map[string]interface{}{
        "status": "active",
        "key": key,
    }, "")
    return nil
}
```

**After (Real API):**
```go
func (ctx *ScenarioContext) iActivateALicenseWithKey(key string) error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return err
    }

    req := integration.PostApiV1LicenseActivateJSONRequestBody{
        LicenseKey: key,
    }

    resp, err := client.PostApiV1LicenseActivateWithResponse(context.Background(), req)
    if err != nil {
        return fmt.Errorf("license activation failed: %w", err)
    }

    ctx.SetLastResponse(resp.StatusCode(), resp.JSON200, "")

    // If activation is async, poll for status
    if resp.StatusCode() == 202 { // Accepted - processing
        return ctx.waitForLicenseActivation(key, 30*time.Second)
    }

    return nil
}

func (ctx *ScenarioContext) waitForLicenseActivation(key string, timeout time.Duration) error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return err
    }

    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for time.Now().Before(deadline) {
        select {
        case <-ticker.C:
            resp, err := client.GetApiV1LicenseWithResponse(context.Background())
            if err != nil {
                continue
            }

            if resp.StatusCode() == 200 && resp.JSON200 != nil {
                if resp.JSON200.Status == "active" {
                    ctx.SetLastResponse(200, resp.JSON200, "")
                    return nil
                }
            }
        }
    }

    return fmt.Errorf("license activation timed out after %v", timeout)
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

**Execution Order (to avoid circular dependency):**
1. Create users using anonymous client (no auth required)
2. Login as first manager user to get auth token
3. Create authenticated client with token
4. Create providers/teams/licenses using authenticated client

```go
// CreateTestDataFromFixtures creates test resources from fixture data via API
func CreateTestDataFromFixtures(ctx *BDDTestContext, fixtures *FixtureData) error {
    anonClient := ctx.AnonymousClient
    var authToken string

    // Step 1: Create users using anonymous client (no auth required)
    for _, user := range fixtures.Users {
        req := integration.PostAuthRegisterJSONRequestBody{
            Email:    user.Email,
            Password: user.Password,
            Name:     user.Name,
        }

        resp, err := anonClient.PostAuthRegisterWithResponse(context.Background(), req)
        if err != nil {
            return fmt.Errorf("failed to create user %s: %w", user.Email, err)
        }

        // Handle 409 (already exists) - track that user exists but don't fail
        if resp.StatusCode() == 409 {
            // User already exists, try to get ID via login
            loginReq := integration.PostAuthLoginJSONRequestBody{
                Email:    user.Email,
                Password: user.Password,
            }
            loginResp, err := anonClient.PostAuthLoginWithResponse(context.Background(), loginReq)
            if err != nil {
                return fmt.Errorf("failed to login existing user %s: %w", user.Email, err)
            }
            if loginResp.StatusCode() == 200 && loginResp.JSON200 != nil && loginResp.JSON200.ID != nil {
                ctx.TrackUser(*loginResp.JSON200.ID)
            }
            continue
        }

        // Track user ID from response
        if resp.StatusCode() == 201 && resp.JSON201 != nil && resp.JSON201.ID != nil {
            ctx.TrackUser(*resp.JSON201.ID)
        } else if resp.StatusCode() == 200 && resp.JSON200 != nil && resp.JSON200.ID != nil {
            ctx.TrackUser(*resp.JSON200.ID)
        } else if resp.StatusCode() != 409 {
            return fmt.Errorf("unexpected status creating user %s: %d", user.Email, resp.StatusCode())
        }
    }

    // Step 2: Login as first manager user to get auth token
    // Find first manager user in fixtures
    var managerUser *TestFixtureUser
    for i := range fixtures.Users {
        if fixtures.Users[i].Role == "manager" {
            managerUser = &fixtures.Users[i]
            break
        }
    }

    if managerUser == nil {
        return fmt.Errorf("no manager user in fixtures for authentication")
    }

    loginReq := integration.PostAuthLoginJSONRequestBody{
        Email:    managerUser.Email,
        Password: managerUser.Password,
    }

    loginResp, err := anonClient.PostAuthLoginWithResponse(context.Background(), loginReq)
    if err != nil {
        return fmt.Errorf("failed to login manager user: %w", err)
    }

    if loginResp.StatusCode() != 200 || loginResp.JSON200 == nil {
        return fmt.Errorf("manager login failed with status %d", loginResp.StatusCode())
    }

    authToken = *loginResp.JSON200.Token
    ctx.AdminToken = authToken

    // Step 3: Create authenticated client
    authenticatedClient, err := NewAuthenticatedClient(ctx.ServerURL, func() (string, error) {
        return authToken, nil
    }, ctx.Logger)
    if err != nil {
        return fmt.Errorf("failed to create authenticated client: %w", err)
    }
    ctx.ManagerClient = authenticatedClient

    // Step 4: Create providers using authenticated client
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
        } else if resp.StatusCode() != 409 {
            return fmt.Errorf("unexpected status creating provider %s: %d", provider.Name, resp.StatusCode())
        }
    }

    // Step 5: Create teams using authenticated client (skip default team ID 1)
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
        } else if resp.StatusCode() != 409 {
            return fmt.Errorf("unexpected status creating team %s: %d", team.Name, resp.StatusCode())
        }
    }

    return nil
}
```

**Data Structures (Existing - in bdd/support/fixtures.go):**

```go
// FixtureData represents test fixture data loaded from JSON
type FixtureData struct {
    Users     []UserFixture     `json:"users"`
    Providers []ProviderFixture `json:"providers"`
    Teams     []TeamFixture     `json:"teams"`
    Licenses  []LicenseFixture  `json:"licenses"`
}

// UserFixture defines a test user
// ID can be specified in fixtures or populated after API creation
type UserFixture struct {
    ID             string `json:"id"`                        // Optional: populated after creation
    Email          string `json:"email"`                     // Required
    Name           string `json:"name"`                      // Required
    Password       string `json:"password"`                  // Required
    Role           string `json:"role"`                      // Required: "admin" or "member"
    TenantID       int64  `json:"tenant_id"`                 // Required
    ProviderUserID string `json:"provider_user_id,omitempty"` // Optional
}

// ProviderFixture defines a test provider
type ProviderFixture struct {
    ID        int64  `json:"id"`          // Optional: populated after creation
    Name      string `json:"name"`        // Required
    Kind      string `json:"kind"`        // Required: "claude", "codex", "opencode"
    APIKey    string `json:"api_key"`     // Required
    Priority  int    `json:"priority"`    // Optional
    Enabled   bool   `json:"enabled"`     // Optional
    TenantID  int64  `json:"tenant_id"`   // Required
    CreatedAt string `json:"created_at,omitempty"` // Optional
}

// TeamFixture defines a test team
type TeamFixture struct {
    ID          int64  `json:"id"`                       // Optional: use to specify known team (e.g., default team ID 1)
    Name        string `json:"name"`                     // Required
    Description string `json:"description,omitempty"`    // Optional
    TenantID    int64  `json:"tenant_id"`                // Required
    CreatedAt   string `json:"created_at,omitempty"`     // Optional
}

// LicenseFixture defines a test license
type LicenseFixture struct {
    ID        string `json:"id"`                       // Optional: populated after creation
    Key       string `json:"key"`                      // Required
    Tier      string `json:"tier"`                     // Required: "trial", "professional", "enterprise"
    ExpiresAt string `json:"expires_at,omitempty"`     // Optional
    CreatedAt string `json:"created_at,omitempty"`     // Optional
}
```

**Note:** The fixture types already exist in `bdd/support/fixtures.go`. The spec uses these existing types rather than creating new ones. User IDs can be specified in fixtures or populated from API responses after creation.

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

**CleanupScenarioResources implementation:**
See "Cleanup Error Handling" section above for the full implementation with retry logic. The method should be added to `step_definitions/context.go` as a method on `ScenarioContext`.

### Implementation Phases

#### Phase 1: Infrastructure Setup

**Tasks:**
1. Create `support/client_factory.go` with client initialization functions
2. Update `support/test_context.go` with real client types and methods
3. Add `support/api_helpers.go` for response handling utilities
4. Update `godog_suite_test.go` to initialize clients in suite setup
5. Add error handling helpers and response wrapper functions
6. Create `bdd/support/fixtures.json` with test data structure

**Deliverable:** Foundation for API-based testing

**Acceptance Criteria:**
- [ ] Client factory creates anonymous and authenticated clients
- [ ] Test context initializes clients successfully
- [ ] Unit tests pass for client factory and helper functions
- [ ] `fixtures.json` exists with at least one manager user
- [ ] Health check endpoint can be called via anonymous client
- [ ] No compilation errors in support package
- [ ] All new code is documented with godoc comments

**Verification:**
```bash
# Test client factory
cd bdd/support
go test -v -run TestNewAnonymousClient
go test -v -run TestNewAuthenticatedClient

# Test health check via client
curl http://localhost:8088/health

# Verify fixtures exist
cat bdd/support/fixtures.json | jq .
```

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
   - Identity & Access (`features/00_identity_and_access.feature`) - ~30 scenarios
   - Provider Management (`features/01_provider_management.feature`) - ~31 scenarios
   - Usage Insights (`features/03_usage_insights.feature`) - ~58 scenarios
   - User Interfaces (`features/04_user_interfaces.feature`) - ~28 scenarios
   - System Behaviors (`features/05_system_behaviors.feature`) - ~5 scenarios
2. Remove all mock implementations from all step files
3. Update CLAUDE.md documentation to reflect API-based approach
4. Verify all 215 scenarios pass with real APIs
5. Update bdd-test.sh to require server for all tests (not just smoke)

**Deliverable:** Complete API-based BDD test suite

**Total Scenarios:** 215 (across all feature files)

**Conversion Order:** Recommended order based on dependencies:
1. System Behaviors (health checks, infrastructure) - ~5 scenarios
2. Identity & Access (auth, users) - ~30 scenarios
3. Provider Management - ~31 scenarios
4. License Management - ~29 scenarios
5. Usage Insights - ~58 scenarios
6. User Interfaces (dashboard, analytics) - ~28 scenarios

### Error Handling Strategy

**Fail Fast Approach:**
- API errors (4xx, 5xx) immediately fail the step
- Clear error messages with request context
- Test run stops on first failure for smoke tests
- Full test run continues but reports all failures

**Response Handler:**
```go
// handleAPIResponse processes API response and returns error if unsuccessful
func handleAPIResponse(statusCode int, body []byte) error {
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

// Check for API errors (status code)
if err := handleAPIResponse(resp.StatusCode(), resp.Body); err != nil {
    return err
}

// Success - store response
ctx.SetLastResponse(resp.StatusCode(), resp.JSON200, "")
```

### Test Data Management

**Available Test Data:**

The BDD module has test data available in `bdd/testdata/` (copied from `integration/testdata/`):

- `testdata/fixtures/users.json` - User credentials (admin, member, member2)
- `testdata/fixtures/providers.json` - Provider configs (claude, codex, opencode)
- `testdata/licenses/*.pem` - License files for different tiers

**Format Note:** The existing test data files use a different format than `FixtureData` expects. We have two options:

1. **Convert to FixtureData format** - Create `bdd/support/fixtures.json` matching the expected array format
2. **Adapt loader** - Update `LoadFixtureData()` to handle both formats

**Recommended:** Create `bdd/support/fixtures.json` in the expected format for clarity and simplicity.

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
- Delete resources in reverse creation order (teams → providers → users)
- Skip default resources (team ID 1, system users)
- Log cleanup failures but don't fail scenario
- Clear tracking maps after cleanup
- Retry failed deletions up to 3 times with exponential backoff

**Cleanup Error Handling:**

```go
// CleanupScenarioResources deletes all tracked resources via API
func (ctx *ScenarioContext) CleanupScenarioResources() error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return fmt.Errorf("failed to get authenticated client for cleanup: %w", err)
    }

    // Retry configuration
    maxRetries := 3
    baseDelay := 100 * time.Millisecond

    // Clean up teams (first - may have foreign key dependencies)
    for _, teamID := range ctx.GetCreatedTeams() {
        if teamID == 1 {
            continue // Skip default team
        }

        var lastErr error
        for attempt := 0; attempt < maxRetries; attempt++ {
            _, err := client.DeleteApiV1TeamsTeamIdWithResponse(context.Background(), teamID)
            if err == nil {
                lastErr = nil
                break
            }
            lastErr = err
            if attempt < maxRetries-1 {
                time.Sleep(baseDelay * time.Duration(1<<attempt))
            }
        }

        if lastErr != nil {
            log.Printf("WARNING: Failed to delete team %d after %d attempts: %v", teamID, maxRetries, lastErr)
        }
    }

    // Clean up providers
    for _, providerID := range ctx.GetCreatedProviders() {
        var lastErr error
        for attempt := 0; attempt < maxRetries; attempt++ {
            _, err := client.DeleteApiV1ProvidersProviderIdWithResponse(context.Background(), providerID)
            if err == nil {
                lastErr = nil
                break
            }
            lastErr = err
            if attempt < maxRetries-1 {
                time.Sleep(baseDelay * time.Duration(1<<attempt))
            }
        }

        if lastErr != nil {
            log.Printf("WARNING: Failed to delete provider %d after %d attempts: %v", providerID, maxRetries, lastErr)
        }
    }

    // Clean up users (last - no foreign key dependencies)
    for _, userID := range ctx.GetCreatedUsers() {
        var lastErr error
        for attempt := 0; attempt < maxRetries; attempt++ {
            _, err := client.DeleteApiV1UsersUserIdWithResponse(context.Background(), userID)
            if err == nil {
                lastErr = nil
                break
            }
            lastErr = err
            if attempt < maxRetries-1 {
                time.Sleep(baseDelay * time.Duration(1<<attempt))
            }
        }

        if lastErr != nil {
            log.Printf("WARNING: Failed to delete user %s after %d attempts: %v", userID, maxRetries, lastErr)
        }
    }

    ctx.ClearCreatedResources()
    return nil
}
```

**Manual Cleanup Procedure:**

If cleanup fails consistently (e.g., server down during test run):

```bash
# Option 1: Restart test database (cleans all data)
cd ../integration
make integration-clean
make integration-setup

# Option 2: Manual cleanup via API
curl -X DELETE http://localhost:8088/api/v1/providers/{id}
curl -X DELETE http://localhost:8088/api/v1/users/{id}
curl -X DELETE http://localhost:8088/api/v1/teams/{id}

# Option 3: Direct database access (last resort)
psql $TEST_DATABASE_URL -c "DELETE FROM providers WHERE name LIKE 'test-%';"
psql $TEST_DATABASE_URL -c "DELETE FROM users WHERE email LIKE '%@example.com';"
```

**Example Fixture Data:**
```json
{
  "users": [
    {
      "email": "test-manager@example.com",
      "password": "TestPassword123!",
      "role": "manager",
      "name": "Test Manager"
    },
    {
      "email": "test-member@example.com",
      "password": "TestPassword123!",
      "role": "member",
      "name": "Test Member"
    }
  ],
  "providers": [
    {
      "name": "test-claude-provider",
      "kind": "claude",
      "api_key": "sk-test-claude-123"
    }
  ],
  "teams": [
    {
      "id": 1,
      "name": "Default Team"
    }
  ]
}
```

**Fixture File Requirements:**

- **Required file:** `bdd/support/fixtures.json`
- **Required content:** At least one user with `role: "admin"` (for authentication)
- **Optional content:** Users, providers, teams, licenses
- **File format:** Valid JSON matching `FixtureData` structure
- **Single source of truth:** All test fixtures loaded from JSON files
- **No hardcoded defaults:** Both `LoadFixtureData()` and `GetStandardFixtureData()` load from JSON

**Fixture Loading Strategy:**

```go
// LoadFixtureData loads fixtures with this priority:
// 1. bdd/support/fixtures.json (local BDD fixtures)
// 2. ../integration/testdata/fixtures.json (shared integration fixtures)
// 3. Return empty FixtureData with error if neither exists
//
// GetStandardFixtureData() loads with this priority:
// 1. bdd/support/fixtures.json (local BDD fixtures)
// 2. Return error if file doesn't exist (no hardcoded fallback)
```

**Rationale:**
- **Maintainability:** All fixtures in JSON, easy to modify without code changes
- **DRY principle:** Single source of truth, no duplication between code and JSON
- **Flexibility:** Easy to add new fixtures or modify existing ones
- **Error visibility:** Missing fixtures fail early rather than using hidden defaults
- **Consistency:** Same fixture data used across BDD, integration, and manual testing

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

### Test Server Requirements

**Minimum server version:** Development build (latest commit)

**Required API endpoints for Phase 1 (Infrastructure + Smoke Tests):**

| Endpoint | Method | Purpose | Auth Required |
|----------|--------|---------|---------------|
| `/health` | GET | Health check validation | No |
| `/auth/login` | POST | User authentication | No |
| `/auth/register` | POST | User registration | No |
| `/auth/refresh` | POST | Token refresh | No |
| `/api/v1/user/profile` | GET | User profile retrieval | Yes |
| `/api/v1/providers` | GET | List providers | Yes |
| `/api/v1/providers` | POST | Create provider | Yes |
| `/api/v1/providers/{id}` | GET | Get provider | Yes |
| `/api/v1/providers/{id}` | PUT | Update provider | Yes |
| `/api/v1/providers/{id}` | DELETE | Delete provider | Yes |

**Handling Missing Endpoints:**

If an endpoint is not implemented yet:
1. Scenario will fail with clear error: "API returned 404: endpoint not implemented"
2. Fix strategy: Implement endpoint in server OR skip scenario with tag `@not-implemented`
3. Never add mock implementations to work around missing endpoints

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

## Implementation Checklist

### Files to CREATE:

- [ ] `bdd/support/client_factory.go` - Client initialization functions
- [ ] `bdd/support/api_helpers.go` - Response handling utilities (optional, can be inline)
- [ ] `bdd/support/fixtures.json` - Test data fixtures (optional, uses defaults if missing)

### Files to MODIFY:

- [ ] `bdd/support/fixtures.go` - Remove hardcoded defaults:
  - Update `GetStandardFixtureData()` to load from `fixtures.json`
  - Remove hardcoded fixture data (use JSON only)
  - Return error if JSON file not found (fail fast)

- [ ] `bdd/support/test_context.go` - Add Logger field and 5 new methods:
  - `InitializeClients(serverURL string, logger *slog.Logger) error`
  - `GetAnonymousClient() *integration.ClientWithResponses`
  - `GetAuthToken() (string, error)`
  - `GetAuthenticatedClient() (*integration.ClientWithResponses, error)`
  - `UpdateAuthenticatedClients(token string) error`

- [ ] `bdd/step_definitions/context.go` - Add/update method:
  - `CleanupScenarioResources() error` - Implement with retry logic

- [ ] `bdd/godog/godog_suite_test.go` - Update hooks:
  - Initialize clients in suite setup
  - Update BeforeScenario to call CreateTestDataFromFixtures
  - Update AfterScenario to call CleanupScenarioResources

- [ ] `bdd/step_definitions/*.go` - Convert mock implementations to real API calls:
  - `auth_steps.go` - Login, logout, refresh, verify
  - `common_steps.go` - Health checks
  - `provider_steps.go` - CRUD operations
  - `license_steps.go` - Activation, status
  - `usage_steps.go` - Statistics
  - `dashboard_steps.go` - Analytics
  - `permission_steps.go` - RBAC
  - `health_steps.go` - System health

### Existing Types (already in bdd/support/fixtures.go):

- `FixtureData` - Container for all fixtures
- `UserFixture` - User test data
- `ProviderFixture` - Provider test data
- `TeamFixture` - Team test data
- `LicenseFixture` - License test data

**Note:** No new fixture types need to be created. Use existing types from `fixtures.go`.
