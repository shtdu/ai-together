# BDD Test Fix Reference - Quick Implementation Guide

**Last Updated:** 2026-03-18
**Status:** Ready for implementation
**Expected Result:** 212-213/215 scenarios passing (96-98%)

## 🎯 Implementation Overview

| Phase | Focus | Scenarios Fixed | Complexity | Time Estimate |
|-------|-------|----------------|------------|---------------|
| **Phase 1** | Authentication & Client Strategy | 9 | Medium | 2-3 hours |
| **Phase 2** | License Management | 14 | Medium | 3-4 hours |
| **Phase 3** | Edge Cases | 3 | Low | 1-2 hours |
| **Total** | All @wip scenarios | 22+ | - | 6-9 hours |

---

## 📋 Phase 1: Authentication & Client Strategy (9 scenarios)

### 1.1 Fix User Registration (2 scenarios)

**File:** `step_definitions/auth_steps.go`

**Add new function:**
```go
// Line ~314 (after userExists function)

// ensureUserExistsViaPublicRegistration follows integration test pattern
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
            return fmt.Errorf("login after duplicate user failed")
        }
    } else if regResp.StatusCode() != 201 {
        return fmt.Errorf("registration failed with status %d", regResp.StatusCode())
    }

    // 5. Login after successful registration
    loginResp, err = ctx.AnonymousClient.PostAuthLoginWithResponse(reqCtx, loginReq)
    if err != nil || loginResp.StatusCode() != 200 {
        return fmt.Errorf("login after registration failed")
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

**Update existing function:**
```go
// Line ~123 (iAmLoggedInAsAMember function)

func (ctx *ScenarioContext) iAmLoggedInAsAMember() error {
    // Load fixtures to get member credentials
    fixtures, err := support.LoadFixtureData()
    if err != nil {
        return fmt.Errorf("failed to load fixtures: %w", err)
    }

    if len(fixtures.Users) < 2 {
        return fmt.Errorf("not enough users in fixtures (need at least 2)")
    }

    // Get member user (second user should be member)
    memberUser := fixtures.Users[1]
    if memberUser.Role != "member" {
        return fmt.Errorf("second user in fixtures is not member, got role: %s", memberUser.Role)
    }

    // ✅ FIXED: Use public registration approach
    if err := ctx.ensureUserExistsViaPublicRegistration(memberUser.Email, memberUser.Password, memberUser.Name); err != nil {
        return fmt.Errorf("failed to ensure member user exists: %w", err)
    }

    return nil
}
```

**Test:**
```bash
./bdd-test.sh --tags "@wip" 2>&1 | grep -E "(Get own profile as member|Profile has correct tenant ID)"
```

### 1.2 Fix Client Selection (7 scenarios)

**File:** `step_definitions/context.go`

**Update GetAuthenticatedClient:**
```go
// Find the GetAuthenticatedClient function and update it

func (ctx *ScenarioContext) GetAuthenticatedClient() (*integration.ClientWithResponses, error) {
    // If we already have an authenticated client, return it
    if ctx.AuthenticatedClient != nil {
        return ctx.AuthenticatedClient, nil
    }

    // Determine which token to use (prioritize admin token)
    var token string
    if ctx.AdminToken != "" {
        token = ctx.AdminToken
    } else if ctx.MemberToken != nil && ctx.MemberToken.Token != "" {
        token = ctx.MemberToken.Token
    } else {
        return nil, fmt.Errorf("cannot create authenticated client without token: no authentication token available")
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

**File:** `step_definitions/provider_steps.go`

**Add ManagerClient support:**
```go
// Add to ScenarioContext struct (line ~20)

type ScenarioContext struct {
    *BDDTestContext
    // ... existing fields ...
    ManagerClient *integration_manager.ClientWithResponses  // Add this line
}
```

**Update team creation:**
```go
// Find iCreateATeam function and update

func (ctx *ScenarioContext) iCreateATeam() error {
    // ✅ FIXED: Use ManagerClient for team operations
    if ctx.ManagerClient == nil {
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

**Test:**
```bash
./bdd-test.sh --tags "@wip" 2>&1 | grep -E "(Manager can create|Member can view own usage)"
```

---

## 📋 Phase 2: License Management (14 scenarios)

### 2.1 Add License Helpers

**File:** `support/fixtures.go` (or create `support/license.go`)

```go
package support

import (
    "fmt"
    "os"
)

// LoadLicenseFixture loads a license PEM file from testdata/licenses
func LoadLicenseFixture(fixtureName string) (string, error) {
    licensePath := fmt.Sprintf("testdata/licenses/%s.pem", fixtureName)
    pemContent, err := os.ReadFile(licensePath)
    if err != nil {
        return "", fmt.Errorf("failed to read license fixture %s: %w", licensePath, err)
    }
    return string(pemContent), nil
}
```

### 2.2 Implement License Activation Steps

**File:** `step_definitions/license_steps.go`

```go
// Replace mock implementation with real API calls

func (ctx *ScenarioContext) iActivateTheLicense() error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        ctx.SetLastResponse(401, nil, "unauthorized: authentication required")
        return nil
    }

    // Get license key from context
    licenseKey, hasKey := ctx.GetCreatedResource("license_key")
    if !hasKey {
        ctx.SetLastResponse(400, nil, "no license key provided")
        return nil
    }

    req := integration.PostApiV1LicenseActivateJSONRequestBody{
        LicenseKey: licenseKey,
    }

    resp, err := client.PostApiV1LicenseActivateWithResponse(context.Background(), req)
    if err != nil {
        ctx.SetLastResponse(500, nil, fmt.Sprintf("API request failed: %v", err))
        return nil
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
        ctx.SetLastResponse(resp.StatusCode(), nil, "")
    }

    return nil
}

func (ctx *ScenarioContext) iHaveAValidCommercialLicenseKey() error {
    licensePEM, err := support.LoadLicenseFixture("tier_1")
    if err != nil {
        return fmt.Errorf("failed to load commercial license: %w", err)
    }

    ctx.TrackCreatedResource("license_key", licensePEM)
    ctx.TrackCreatedResource("license_tier", "professional")
    return nil
}

func (ctx *ScenarioContext) iHaveAValidOpenSourceLicenseKey() error {
    licensePEM, err := support.LoadLicenseFixture("opensource")
    if err != nil {
        return fmt.Errorf("failed to load open-source license: %w", err)
    }

    ctx.TrackCreatedResource("license_key", licensePEM)
    ctx.TrackCreatedResource("license_tier", "trial")
    return nil
}
```

### 2.3 Implement License Feature Steps

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

    // LicenseStatus has a Type field
    if license, ok := resp.(*integration.LicenseStatus); ok {
        if license.Type == nil {
            return fmt.Errorf("license type is nil")
        }

        tier := string(*license.Type)
        // All tiers support provider management
        if tier == "opensource" || tier == "trial" || tier == "starter" ||
           tier == "professional" || tier == "enterprise" {
            return nil
        }
        return fmt.Errorf("provider management not enabled for tier: %s", tier)
    }

    return fmt.Errorf("response is not a LicenseStatus")
}
```

**Test:**
```bash
./bdd-test.sh --tags "@wip" 2>&1 | grep -E "(Activate.*license|License.*feature)"
```

---

## 📋 Phase 3: Edge Cases (3 scenarios)

### 3.1 Investigate Provider Count Issue

**File:** `step_definitions/provider_steps.go`

**Add debugging to list providers:**
```go
func (ctx *ScenarioContext) iListAllProvidersFromProvider() error {
    client, err := ctx.GetAuthenticatedClient()
    if err != nil || client == nil {
        ctx.SetLastResponse(401, nil, "unauthorized: authentication required")
        return nil
    }

    // Call API to list providers
    resp, err := client.GetApiV1ProvidersWithResponse(context.Background())
    if err != nil {
        ctx.SetLastResponse(0, nil, err.Error())
        return fmt.Errorf("provider list request failed: %w", err)
    }

    // ✅ DEBUG: Log response details
    if resp.JSON200 != nil {
        fmt.Printf("DEBUG: Listed %d providers\n", len(*resp.JSON200))
        for i, p := range *resp.JSON200 {
            fmt.Printf("  [%d] ID=%d Name=%s\n", i, p.Id, p.Name)
        }
    }

    // ... rest of function
}
```

**Test:**
```bash
./bdd-test.sh --tags "@wip" 2>&1 | grep -A 10 "Count providers towards limit"
```

### 3.2 Review Remaining @wip Scenarios

**Usage metadata:** API limitation - keep @wip or remove scenario
**Team analytics permissions:** Investigate if API returns 403 or 200

---

## ✅ Testing Strategy

### Before Each Phase
```bash
# Establish baseline
./bdd-test.sh --tags "~@wip"
# Expected: 171/171 passing
```

### After Each Phase
```bash
# Test all scenarios
./bdd-test.sh
# Track progress: 193 → 202 → 216/215 (target)
```

### Final Validation
```bash
# Run integration tests to ensure no regression
cd ../integration && go test -v ./...
# Expected: 141/141 passing
```

---

## 🔧 Troubleshooting

### Issue: "no authentication token available"
**Fix:** Ensure `ensureUserExistsViaPublicRegistration()` is called before operations

### Issue: "provider limit reached"
**Fix:** Check if cleanup is working - providers may be accumulating

### Issue: "unexpected response type"
**Fix:** Add response type checking using reflection (see `totalProviderCountShouldBe`)

---

## 📊 Success Metrics

| Phase | Target | Verification |
|-------|--------|--------------|
| Phase 1 | 202/215 (94%) | `./bdd-test.sh` |
| Phase 2 | 213/215 (99%) | `./bdd-test.sh` |
| Phase 3 | 215/215 (100%) | `./bdd-test.sh` |

**Integration Tests:** Must remain 141/141 (100%)

---

## 📚 Reference Documents

- `BDD_VS_INTEGRATION_ANALYSIS.md` - Detailed analysis with integration test code
- `BDD_FAILURES_ANALYSIS.md` - Root cause analysis
- `BDD_STATUS.md` - Overall status and roadmap
- `../integration/` - Working reference implementation (100% pass rate)
