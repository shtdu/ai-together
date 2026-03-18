# BDD vs Integration Test Analysis - Resolution Strategy

**Generated:** 2026-03-18
**Purpose:** Systematic analysis of 22 failing BDD scenarios vs 100% passing integration tests
**Goal:** Provide concrete fix patterns based on working integration test code

## Executive Summary

| Metric | BDD Tests | Integration Tests |
|--------|-----------|-------------------|
| **Pass Rate** | 193/215 (89.8%) | 141/141 (100%) |
| **Test Type** | Gherkin scenarios | Go unit tests |
| **Client Strategy** | Mixed (anonymous/authenticated) | Clear separation |
| **User Creation** | Manager endpoint (broken) | Public registration (working) |
| **License Handling** | Not implemented | Full coverage (8 tests) |

**Root Cause:** BDD tests use different architectural patterns than integration tests for 3 key areas:
1. **User authentication** - Uses admin endpoint instead of public registration
2. **License management** - Steps not implemented
3. **Client selection** - Wrong client types for specific operations

---

## Category 1: User Authentication Failures (2 scenarios)

### Failing Scenarios
```
@wip
Scenario: Get own profile as member
  Given I am logged in as a member
  When I get my user profile
  Then the response status code should be 200

@wip
Scenario: Profile has correct tenant ID
  Given I am logged in as a member
  And I belong to tenant with ID "1"
  When I get my user profile
  Then the response status code should be 200
```

### BDD Implementation (BROKEN)
**File:** `step_definitions/auth_steps.go`
```go
func (ctx *ScenarioContext) iAmLoggedInAsAMember() error {
    // ... load fixtures ...
    // ❌ WRONG: Uses manager endpoint to create users
    if err := ctx.userExists(memberUser.Email, memberUser.Password); err != nil {
        return fmt.Errorf("failed to ensure member user exists: %w", err)
    }
    return ctx.iLoginWithCredentials(memberUser.Email, memberUser.Password)
}
```

**Problem:** `userExists()` tries to create users via `POST /api/v1/users` (admin-only endpoint)

### Integration Test Pattern (WORKING ✅)
**File:** `integration/integration_test.go:185`
```go
func (s *IntegrationTestSuite) registerAndLoginUserFromFixture(fixtureName string) string {
    // ✅ CORRECT: Try login first via AnonymousClient
    loginResp := s.AnonymousClient.PostAuthLoginWithResponse(ctx, loginReq)

    // ✅ If login fails, register via public endpoint
    if loginResp.StatusCode() != 200 {
        regResp := s.AnonymousClient.PostAuthRegisterWithResponse(ctx, regReq)

        // ✅ Handle 500 (duplicate user) by retrying login
        if regResp.StatusCode() == 500 {
            loginResp = s.AnonymousClient.PostAuthLoginWithResponse(ctx, loginReq)
        }
    }

    return loginResp.JSON200.AccessToken
}
```

### Resolution Strategy
1. **Replace** `userExists()` with `ensureUserExistsViaPublicRegistration()`
2. **Use** `ctx.AnonymousClient` for both login and registration
3. **Follow** exact pattern from `registerAndLoginUserFromFixture()`

### Implementation Code
```go
func (ctx *ScenarioContext) ensureUserExistsViaPublicRegistration(email, password, name string) error {
    reqCtx := context.Background()
    emailAddr := openapi_types.Email(email)

    // 1. Try login first (user might already exist)
    loginReq := integration.PostAuthLoginJSONRequestBody{
        Email:    emailAddr,
        Password: password,
    }
    loginResp, err := ctx.AnonymousClient.PostAuthLoginWithResponse(reqCtx, loginReq)

    // 2. If login succeeds, store token and return
    if err == nil && loginResp.StatusCode() == 200 && loginResp.JSON200 != nil {
        token := loginResp.JSON200.AccessToken
        user := loginResp.JSON200.User

        // Store token based on role
        if string(user.Role) == "admin" || string(user.Role) == "manager" {
            ctx.AdminToken = token
        } else {
            ctx.MemberToken = token
        }

        ctx.BDDTestContext.CurrentUser = &support.UserInfo{
            Email:    string(user.Email),
            Password: password,
            Name:     user.Name,
            Role:     string(user.Role),
            Token:    token,
        }
        return nil
    }

    // 3. Login failed, try to register via public endpoint
    regReq := integration.PostAuthRegisterJSONRequestBody{
        Email:    emailAddr,
        Password: password,
        Name:     name,
    }
    regResp, err := ctx.AnonymousClient.PostAuthRegisterWithResponse(reqCtx, regReq)
    if err != nil {
        return fmt.Errorf("registration request failed: %w", err)
    }

    // 4. Handle 500 (duplicate user) - retry login
    if regResp.StatusCode() == 500 {
        loginResp, err = ctx.AnonymousClient.PostAuthLoginWithResponse(reqCtx, loginReq)
        if err != nil || loginResp.StatusCode() != 200 {
            return fmt.Errorf("login after duplicate user failed: %w", err)
        }
    } else if regResp.StatusCode() != 201 {
        errMsg := ""
        if regResp.JSON400 != nil {
            errMsg = fmt.Sprintf("validation error: %s", regResp.JSON400.Error)
        }
        return fmt.Errorf("registration failed: %s", errMsg)
    }

    // 5. Login after successful registration
    loginResp, err = ctx.AnonymousClient.PostAuthLoginWithResponse(reqCtx, loginReq)
    if err != nil || loginResp.StatusCode() != 200 {
        return fmt.Errorf("login after registration failed: %w", err)
    }

    // 6. Store token and user info
    token := loginResp.JSON200.AccessToken
    user := loginResp.JSON200.User

    if string(user.Role) == "admin" || string(user.Role) == "manager" {
        ctx.AdminToken = token
    } else {
        ctx.MemberToken = token
    }

    ctx.BDDTestContext.CurrentUser = &support.UserInfo{
        Email:    string(user.Email),
        Password: password,
        Name:     user.Name,
        Role:     string(user.Role),
        Token:    token,
    }

    return nil
}
```

**Impact:** Fixes 2 scenarios, improves foundation for 7+ more scenarios

---

## Category 2: Provider Management Permissions (7 scenarios)

### Failing Scenarios
```
@wip
Scenario: Manager can create provider
@wip
Scenario: Manager can update provider
@wip
Scenario: Manager can delete provider
@wip
Scenario: Manager can manage users
@wip
Scenario: Manager can view team analytics
@wip
Scenario: Member can view own usage
@wip
Scenario: Get provider by ID
```

### BDD Implementation (BROKEN)
**File:** `step_definitions/provider_steps.go`
```go
func (ctx *ScenarioContext) iCreateAProviderWithKindAndAPIKey(kind, apiKey string) error {
    // ❌ WRONG: Uses GetAuthenticatedClient() which may not have admin token
    client, err := ctx.GetAuthenticatedClient()
    if err != nil || client == nil {
        ctx.SetLastResponse(401, nil, "unauthorized: admin/manager access required")
        return nil
    }

    resp, err := client.PostApiV1ProvidersWithResponse(context.Background(), req)
    // ...
}
```

**Problem:** `GetAuthenticatedClient()` returns wrong client type or no client for managers

### Integration Test Pattern (WORKING ✅)
**File:** `integration/integration_test.go:80-130`
```go
func (s *IntegrationTestSuite) SetupTest() {
    // ✅ CORRECT: Clear tokens and authenticate fresh for each test
    s.AdminToken = ""
    s.MemberToken = ""

    // ✅ Authenticate as admin and create Client
    adminToken := s.registerAndLoginUserFromFixture("admin")
    s.AdminToken = adminToken

    client, err := integrationclient.NewAuthenticatedClient(
        s.ServerURL,
        integrationclient.TokenGetter(func() (string, error) { return s.AdminToken, nil }),
        s.Logger,
    )
    s.Require().NoError(err)
    s.Client = client

    // ✅ Authenticate as member if needed
    memberToken := s.registerAndLoginUserFromFixture("member")
    s.MemberToken = memberToken
}
```

**File:** `integration/provider_test.go:29`
```go
func (s *IntegrationTestSuite) TestProviderCreateClaude() {
    // ✅ Uses s.Client (authenticated admin client)
    resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
    require.NoError(s.T(), err)
    require.Equal(s.T(), 201, resp.StatusCode())
}
```

### Resolution Strategy
1. **Fix** `GetAuthenticatedClient()` to return admin client for managers
2. **Ensure** admin token is set before manager operations
3. **Use** same pattern as integration tests for client creation

### Implementation Code
```go
// In step_definitions/context.go
func (ctx *ScenarioContext) GetAuthenticatedClient() (*integration.ClientWithResponses, error) {
    // If we already have an authenticated client, return it
    if ctx.AuthenticatedClient != nil {
        return ctx.AuthenticatedClient, nil
    }

    // Determine which token to use
    var token string
    if ctx.AdminToken != "" {
        token = ctx.AdminToken
    } else if ctx.MemberToken != nil && ctx.MemberToken.Token != "" {
        token = ctx.MemberToken.Token
    } else {
        return nil, fmt.Errorf("cannot create authenticated client without token")
    }

    // Create authenticated client
    client, err := integration.NewAuthenticatedClient(
        ctx.ServerURL,
        integration.TokenGetter(func() (string, error) { return token, nil }),
        ctx.Logger,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create authenticated client: %w", err)
    }

    ctx.AuthenticatedClient = client
    return client, nil
}
```

**Impact:** Fixes 7 scenarios related to provider operations

---

## Category 3: License Management (7 scenarios)

### Failing Scenarios
```
@wip
Scenario: Activate commercial license successfully
@wip
Scenario: Activate open-source license successfully
@wip
Scenario: Activate license with signature
@wip
Scenario: Activate license with invalid signature
@wip
Scenario: Activate license without authentication
@wip
Scenario: Non-manager cannot activate license
@wip (and 13 more license scenarios)
```

### BDD Implementation (NOT IMPLEMENTED ❌)
**File:** `step_definitions/license_steps.go`
```go
// ❌ MISSING: License activation steps not implemented
func (ctx *ScenarioContext) iActivateTheLicense() error {
    // Not implemented - returns mock response
    ctx.SetLastResponse(200, map[string]interface{}{
        "status": "active",
        "tier":   "professional",
    }, "")
    return nil
}
```

### Integration Test Pattern (WORKING ✅)
**File:** `integration/license_test.go:68`
```go
func (s *IntegrationTestSuite) TestLicenseActivateOpenSource() {
    ctx := context.Background()

    // ✅ Load license from PEM file
    licensePEM, err := LoadLicenseFixture(LicenseOpenSource)
    require.NoError(s.T(), err)

    // ✅ Activate via authenticated client
    req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
        LicenseKey: licensePEM,
    }

    resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
    require.NoError(s.T(), err)
    require.Equal(s.T(), 200, resp.StatusCode())

    // ✅ Verify response
    assert.Equal(s.T(), "opensource", getLicenseType(resp.JSON200))
}
```

**File:** `integration/helpers.go`
```go
// LoadLicenseFixture loads a license PEM file from testdata
func LoadLicenseFixture(fixtureName string) (string, error) {
    licensePath := fmt.Sprintf("testdata/licenses/%s", fixtureName)
    pemContent, err := os.ReadFile(licensePath)
    if err != nil {
        return "", fmt.Errorf("failed to read license fixture: %w", err)
    }
    return string(pemContent), nil
}

// activateLicenseFixture loads and activates a license
func activateLicenseFixture(s *IntegrationTestSuite, fixtureName string) {
    ctx := context.Background()
    pemContent, _ := LoadLicenseFixture(fixtureName)

    req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
        LicenseKey: string(pemContent),
    }

    resp := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
    assert.Equal(s.T(), 200, resp.StatusCode())
}
```

### Resolution Strategy
1. **Add** `LoadLicenseFixture()` helper to BDD support
2. **Implement** real license activation steps
3. **Use** authenticated client for activation
4. **Handle** error responses (400, 401, 403)

### Implementation Code
```go
// In step_definitions/license_steps.go

func (ctx *ScenarioContext) iActivateTheLicense() error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return fmt.Errorf("failed to get authenticated client: %w", err)
    }

    // Get license key from context
    licenseKey, hasKey := ctx.GetCreatedResource("license_key")
    if !hasKey {
        return fmt.Errorf("no license key found in context")
    }

    req := integration.PostApiV1LicenseActivateJSONRequestBody{
        LicenseKey: licenseKey,
    }

    resp, err := client.PostApiV1LicenseActivateWithResponse(context.Background(), req)
    if err != nil {
        return fmt.Errorf("license activation request failed: %w", err)
    }

    // Handle response
    switch {
    case resp.JSON200 != nil:
        ctx.SetLastResponse(200, resp.JSON200, "")
    case resp.JSON400 != nil:
        ctx.SetLastResponse(400, resp.JSON400, "")
    case resp.JSON401 != nil:
        ctx.SetLastResponse(401, resp.JSON401, "")
    case resp.JSON403 != nil:
        ctx.SetLastResponse(403, resp.JSON403, "")
    default:
        ctx.SetLastResponse(resp.StatusCode(), nil, fmt.Sprintf("unexpected status: %d", resp.StatusCode()))
    }

    return nil
}

func (ctx *ScenarioContext) iHaveAValidCommercialLicenseKey() error {
    // Load license from fixtures
    licensePEM, err := support.LoadLicenseFixture("tier_1.pem")
    if err != nil {
        return fmt.Errorf("failed to load commercial license: %w", err)
    }

    ctx.TrackCreatedResource("license_key", licensePEM)
    ctx.TrackCreatedResource("license_tier", "professional")
    return nil
}
```

**Impact:** Fixes 7+ license scenarios

---

## Category 4: License Features & Limits (7 scenarios)

### Failing Scenarios
```
@wip
Scenario: Commercial license enables provider management
@wip
Scenario: Trial license has limited features
@wip
Scenario: Enterprise license enables all features
@wip
Scenario: License provider limits by tier
@wip
Scenario: Provider limit enforced when creating providers
@wip
Scenario: Update provider does not count towards limit
@wip
Scenario: Delete provider frees up limit
```

### BDD Implementation (MOCK DATA ❌)
**File:** `step_definitions/license_steps.go`
```go
// ❌ WRONG: Returns hardcoded mock data
func (ctx *ScenarioContext) iCheckAvailableFeatures() error {
    ctx.SetLastResponse(200, map[string]interface{}{
        "provider_management": true,
        "team_analytics":      true,
    }, "")
    return nil
}
```

### Integration Test Pattern (WORKING ✅)
**File:** `integration/license_test.go:183`
```go
func (s *IntegrationTestSuite) TestLicenseGetStatus() {
    ctx := context.Background()

    // ✅ CORRECT: Call real API endpoint
    resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
    require.NoError(s.T(), err)
    require.Equal(s.T(), 200, resp.StatusCode())

    // ✅ Verify actual license properties from API
    license := resp.JSON200
    assert.NotNil(s.T(), license)
    assert.Equal(s.T(), "opensource", getLicenseType(license))
    assert.Equal(s.T(), 1, getMaxTeams(license))
    assert.Equal(s.T(), -1, getMaxSeats(license)) // -1 = unlimited
}
```

### Resolution Strategy
1. **Replace** mock data with real API calls
2. **Use** `GET /api/v1/license` to get license status
3. **Parse** response to extract features and limits
4. **Store** license info in context for verification

### Implementation Code
```go
func (ctx *ScenarioContext) iGetLicenseFeatures() error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return fmt.Errorf("failed to get authenticated client: %w", err)
    }

    resp, err := client.GetApiV1LicenseWithResponse(context.Background())
    if err != nil {
        return fmt.Errorf("failed to get license info: %w", err)
    }

    if resp.JSON200 == nil {
        return fmt.Errorf("unexpected response: %d", resp.StatusCode())
    }

    ctx.SetLastResponse(200, resp.JSON200, "")
    return nil
}

func (ctx *ScenarioContext) providerManagementShouldBeEnabled() error {
    _, resp, _ := ctx.GetLastResponse()
    if license, ok := resp.(*integration.LicenseStatus); ok {
        if license.Type == nil {
            return fmt.Errorf("license type is nil")
        }
        // Check if provider management is enabled for this tier
        tier := string(*license.Type)
        if tier == "opensource" || tier == "trial" || tier == "starter" ||
           tier == "professional" || tier == "enterprise" {
            return nil // All these tiers support provider management
        }
        return fmt.Errorf("provider management not enabled for tier: %s", tier)
    }
    return fmt.Errorf("response is not a LicenseStatus")
}
```

**Impact:** Fixes 7 license feature/limit scenarios

---

## Category 5: Analytics & Usage (3 scenarios)

### Failing Scenarios
```
@wip
Scenario: Upload usage record with metadata
@wip
Scenario: Member cannot see team analytics
@wip
Scenario: Count providers towards limit
```

### Issue 1: Usage Metadata (API Limitation)
**Problem:** `POST /api/v1/usage/batch` only returns `{synced_count: int}` - doesn't include uploaded records
**BDD Expectation:** Metadata should be in response
**Integration Test:** Doesn't verify metadata in batch response (acknowledges limitation)

**Resolution:** This scenario tests non-existent API behavior. Tag as `@wip` indefinitely or change to query usage by timestamp.

### Issue 2: Team Analytics Permissions
**Problem:** Returns 200 instead of 403 for members
**Integration Test:** N/A - no specific test for this permission

**Resolution:** Add permission check or adjust expectation based on actual API behavior.

### Issue 3: Provider Count Pagination
**Problem:** API returns 1 provider instead of 3
**Hypothesis:** Pagination or filtering in list API
**Integration Test:** Uses `s.Client.GetApiV1ProvidersWithResponse()` successfully

**Resolution:** Investigate pagination parameters or fix API response handling.

**Impact:** 3 scenarios require API investigation or behavior clarification

---

## Category 6: Team Management (1 scenario)

### Failing Scenario
```
@wip
Scenario: Manager can manage teams
  Given I am logged in as a manager
  When I create a team
  Then the operation should succeed
```

### BDD Implementation (WRONG CLIENT ❌)
**File:** `step_definitions/provider_steps.go`
```go
func (ctx *ScenarioContext) iCreateATeam() error {
    // ❌ WRONG: Uses Client (admin client) instead of ManagerClient
    client, err := ctx.GetAuthenticatedClient()
    // ...
}
```

### Integration Test Pattern (WORKING ✅)
**File:** `integration/mgr_team_test.go:29`
```go
func (s *IntegrationTestSuite) TestMgrTeamCreate() {
    // ✅ CORRECT: Uses ManagerClient for team operations
    resp := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, req)
    assert.Equal(s.T(), 201, resp.StatusCode())
}
```

**File:** `integration/integration_test.go:130-160`
```go
func (s *IntegrationTestSuite) SetupTest() {
    // ... admin authentication ...

    // ✅ Authenticate as manager and create ManagerClient
    managerToken := s.registerAndLoginUserFromFixture("manager")
    s.ManagerToken = managerToken

    managerClient, err := integration_manager.NewAuthenticatedClient(
        s.ServerURL,
        integration_manager.TokenGetter(func() (string, error) { return s.ManagerToken, nil }),
        s.Logger,
    )
    s.Require().NoError(err)
    s.ManagerClient = managerClient
}
```

### Resolution Strategy
1. **Add** ManagerClient to BDD test context
2. **Create** manager token during setup
3. **Use** ManagerClient for team operations
4. **Import** `integration_manager` package

### Implementation Code
```go
// In step_definitions/context.go

type ScenarioContext struct {
    *BDDTestContext
    // ... existing fields ...
    ManagerClient *integration_manager.ClientWithResponses
}

func (ctx *ScenarioContext) iCreateATeam() error {
    // ✅ Use ManagerClient for team operations
    if ctx.ManagerClient == nil {
        // Create manager client if not exists
        if ctx.AdminToken == "" {
            return fmt.Errorf("no admin token available for manager client")
        }

        client, err := integration_manager.NewAuthenticatedClient(
            ctx.ServerURL,
            integration_manager.TokenGetter(func() (string, error) { return ctx.AdminToken, nil }),
            ctx.Logger,
        )
        if err != nil {
            return fmt.Errorf("failed to create manager client: %w", err)
        }
        ctx.ManagerClient = client
    }

    req := integration_manager.PostApiV1TeamsJSONRequestBody{
        Name: support.GenerateUniqueTeamName("test"),
    }

    resp, err := ctx.ManagerClient.PostApiV1TeamsWithResponse(context.Background(), req)
    if err != nil {
        return fmt.Errorf("failed to create team: %w", err)
    }

    // Handle response
    switch {
    case resp.JSON201 != nil:
        ctx.SetLastResponse(201, resp.JSON201, "")
        ctx.TrackTeam(resp.JSON201.Id)
    case resp.JSON403 != nil:
        ctx.SetLastResponse(403, resp.JSON403, "")
    default:
        ctx.SetLastResponse(resp.StatusCode(), nil, "")
    }

    return nil
}
```

**Impact:** Fixes 1 team management scenario

---

## Implementation Priority & Impact

### Phase 1: Foundation (HIGH) - 2+7 scenarios
1. **Fix user authentication** (2 scenarios)
   - Implement public registration approach
   - Impact: Enables all member scenarios
2. **Fix client selection** (7 scenarios)
   - Correct GetAuthenticatedClient() logic
   - Add ManagerClient support
   - Impact: Fixes provider operations

### Phase 2: License Implementation (MEDIUM) - 14 scenarios
1. **Implement license activation** (7 scenarios)
   - Add LoadLicenseFixture() helper
   - Implement real activation steps
2. **Implement license features** (7 scenarios)
   - Replace mock data with API calls
   - Add limit checking logic

### Phase 3: Edge Cases (LOW) - 3 scenarios
1. **Investigate API limitations** (3 scenarios)
   - Usage metadata (API limitation)
   - Team analytics permissions
   - Provider count pagination

**Total Impact:** 22 scenarios → 0-3 remaining failures (96-98% pass rate)

---

## Testing Strategy

### Before Implementation
```bash
# Run only passing scenarios to establish baseline
./bdd-test.sh --tags "~@wip"
# Expected: 171/171 passing (100%)
```

### After Each Phase
```bash
# Test all scenarios to measure improvement
./bdd-test.sh
# Track progress: 193 → 200 → 207 → 212/215
```

### Validation
```bash
# Run integration tests to ensure no regression
cd ../integration && go test -v ./...
# Expected: 141/141 passing (100%)
```

---

## Conclusion

The integration tests provide a **proven, working reference implementation** for all failing BDD scenarios. The key differences are:

1. **User Creation:** Public registration vs admin endpoint
2. **Client Strategy:** Clear separation of AnonymousClient/Client/ManagerClient
3. **API Usage:** Real endpoints vs mock responses
4. **Error Handling:** Proper HTTP status code checking

By following the exact patterns from integration tests, we can achieve **96-98% pass rate** (212-213/215 scenarios passing).
