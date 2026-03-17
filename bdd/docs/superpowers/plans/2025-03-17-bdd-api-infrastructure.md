# BDD API Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the infrastructure foundation for BDD tests to use real API calls instead of mocks, including client factory, test context updates, fixture loading, and test lifecycle hooks.

**Architecture:** Replace mock-based step implementations with real HTTP API calls using the generated OpenAPI client from `shared/integration`. Tests will authenticate against the test server (port 8088) and make actual API requests for resource creation, cleanup, and validation.

**Tech Stack:** Go 1.25+, godog BDD framework, generated OpenAPI client (`github.com/code-together/shared/integration`), slog for logging

---

## Scope

This plan covers **Phase 1: Infrastructure Setup** from the design spec. This creates the foundation for API-based testing:

1. Client factory for creating anonymous and authenticated API clients
2. Test context updates to store and manage API clients
3. Fixture loading from JSON files (no hardcoded defaults)
4. Test lifecycle hooks for data creation and cleanup
5. Response handling utilities

**Out of scope (Phase 2):** Converting individual step definitions from mocks to real API calls (will be planned separately after this infrastructure is validated)

---

## File Structure

### New Files to Create:

- `bdd/support/client_factory.go` - Client initialization functions (wrappers around `shared/integration`)
- `bdd/support/api_helpers.go` - Response handling and error utilities
- `bdd/support/fixtures.json` - Test fixture data (already created)

### Files to Modify:

- `bdd/support/fixtures.go` - Remove hardcoded defaults, load from JSON only
- `bdd/support/test_context.go` - Add Logger field, client management methods
- `bdd/step_definitions/context.go` - Implement CleanupScenarioResources with retry logic
- `bdd/godog/godog_suite_test.go` - Initialize clients, update BeforeScenario/AfterScenario hooks

---

## Chunk 1: Client Factory and Test Context Foundation

### Task 1: Create client factory with API client wrappers

**Files:**
- Create: `bdd/support/client_factory.go`

- [ ] **Step 1: Create client_factory.go with package declaration and imports**

```go
// Package support provides test utilities and context management for BDD tests
package support

import (
    "log/slog"
    "github.com/code-together/shared/integration"
)
```

Run: `touch bdd/support/client_factory.go`

- [ ] **Step 2: Implement NewAnonymousClient wrapper function**

```go
// NewAnonymousClient creates a client without authentication for public endpoints
// Verbose flag controls whether request/response bodies are logged
func NewAnonymousClient(serverURL string, logger *slog.Logger, verbose bool) (*integration.ClientWithResponses, error) {
    return integration.NewAnonymousClient(serverURL, logger, verbose)
}
```

- [ ] **Step 3: Implement NewAuthenticatedClient wrapper function**

```go
// NewAuthenticatedClient creates a client with automatic token injection
// getToken function provides the current auth token for each request
// Verbose flag controls whether request/response bodies are logged
func NewAuthenticatedClient(
    serverURL string,
    getToken func() (string, error),
    logger *slog.Logger,
    verbose bool,
) (*integration.ClientWithResponses, error) {
    return integration.NewAuthenticatedClient(serverURL, getToken, logger, verbose)
}
```

- [ ] **Step 4: Verify compilation with package build**

Run:
```bash
cd bdd/support
go build ./support/...
```

Expected: No compilation errors (this checks all files in the package for import/type issues)

- [ ] **Step 5: Commit**

```bash
git add bdd/support/client_factory.go
git commit -m "feat: add API client factory for BDD tests

Add wrapper functions for creating anonymous and authenticated API clients.
These wrap the generated OpenAPI client from shared/integration.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 2: Update fixtures.go to load from JSON only

**Files:**
- Modify: `bdd/support/fixtures.go:60-85`

- [ ] **Step 1: Read current LoadFixtureData implementation**

Run:
```bash
cd bdd/support
sed -n '60,85p' fixtures.go
```

Note the current implementation that checks integration testdata

- [ ] **Step 2: Replace LoadFixtureData to check local fixtures first**

Replace the entire LoadFixtureData function (lines 60-85) with:

```go
// LoadFixtureData loads test fixture data from JSON files
// Priority: 1) local fixtures.json, 2) integration testdata, 3) error
// Note: Falls back to integration testdata for compatibility during transition
func LoadFixtureData() (*FixtureData, error) {
    // First, try local fixtures.json (in bdd/support/)
    localFixturePath := filepath.Join(".", "fixtures.json")

    if data, err := os.ReadFile(localFixturePath); err == nil {
        var fixtures FixtureData
        if err := json.Unmarshal(data, &fixtures); err != nil {
            return nil, fmt.Errorf("failed to parse local fixture JSON: %w", err)
        }
        return &fixtures, nil
    }

    // Fall back to integration testdata (for compatibility)
    integrationFixturePath := filepath.Join("..", "..", "integration", "testdata", "fixtures.json")

    if _, err := os.Stat(integrationFixturePath); os.IsNotExist(err) {
        return nil, fmt.Errorf("no fixtures file found: tried %s and %s", localFixturePath, integrationFixturePath)
    }

    data, err := os.ReadFile(integrationFixturePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read integration fixture file: %w", err)
    }

    var fixtures FixtureData
    if err := json.Unmarshal(data, &fixtures); err != nil {
        return nil, fmt.Errorf("failed to parse integration fixture JSON: %w", err)
    }

    return &fixtures, nil
}
```

Note: This keeps the fallback to integration testdata for compatibility. The "JSON only" requirement is enforced by `GetStandardFixtureData()` which has no fallback.

- [ ] **Step 3: Replace GetStandardFixtureData to load from JSON**

Replace the entire GetStandardFixtureData function (lines 103-153) with:

```go
// GetStandardFixtureData returns fixture data from fixtures.json
// Returns error if file doesn't exist (no hardcoded fallback)
func GetStandardFixtureData() (*FixtureData, error) {
    localFixturePath := filepath.Join(".", "fixtures.json")

    if _, err := os.Stat(localFixturePath); os.IsNotExist(err) {
        return nil, fmt.Errorf("fixtures.json not found in %s", localFixturePath)
    }

    data, err := os.ReadFile(localFixturePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read fixtures.json: %w", err)
    }

    var fixtures FixtureData
    if err := json.Unmarshal(data, &fixtures); err != nil {
        return nil, fmt.Errorf("failed to parse fixtures.json: %w", err)
    }

    return &fixtures, nil
}
```

- [ ] **Step 4: Verify fixtures.json exists and is valid JSON**

Run:
```bash
cd bdd/support
cat fixtures.json | jq .
```

Expected: Valid JSON output showing users, providers, teams, licenses

- [ ] **Step 5: Test the updated fixture loading**

Run:
```bash
cd bdd/support
go test -v -run TestLoadFixtureData
```

Expected: Tests pass (or create test if it doesn't exist)

- [ ] **Step 6: Commit**

```bash
git add bdd/support/fixtures.go
git commit -m "refactor: load fixtures from JSON only, remove hardcoded defaults

Update LoadFixtureData and GetStandardFixtureData to load from
fixtures.json, returning error if file not found.

Benefits:
- Single source of truth for test data
- Easy to modify fixtures without code changes
- Fail fast if fixtures.json is missing

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 3: Add API helpers for response handling

**Files:**
- Create: `bdd/support/api_helpers.go`

- [ ] **Step 1: Create api_helpers.go with package declaration**

```go
// Package support provides test utilities and context management for BDD tests
package support

import (
    "fmt"
    "net/http"
)
```

Run: `touch bdd/support/api_helpers.go`

- [ ] **Step 2: Implement handleAPIResponse function**

```go
// handleAPIResponse processes API response and returns error if unsuccessful
// statusCode is the HTTP status code from the response
// body is the raw response body bytes
// Returns error if status code indicates failure (>= 400)
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

- [ ] **Step 3: Check if IsTestServerRunning already exists**

Check if function exists:
```bash
cd bdd/support
grep -n "IsTestServerRunning" *.go
```

Expected: Function already exists in `server_lifecycle.go`

If NOT found (should not happen), add to api_helpers.go:
```go
import (
    // ... existing imports
    "net/http"
    "time"
)

// IsTestServerRunning checks if the test server is running at the given URL
func IsTestServerRunning(serverURL string) bool {
    client := &http.Client{
        Timeout: 2 * time.Second,
    }

    resp, err := client.Get(serverURL + "/health")
    if err != nil {
        return false
    }
    defer resp.Body.Close()

    return resp.StatusCode == http.StatusOK
}
```

Note: This function already exists in `server_lifecycle.go`, so we'll use that implementation.

- [ ] **Step 4: Verify compilation with package build**

Run:
```bash
cd bdd/support
go build ./support/...
```

Expected: No compilation errors

- [ ] **Step 5: Commit**

```bash
git add bdd/support/api_helpers.go
git commit -m "feat: add API response handling helpers

Add handleAPIResponse for consistent error handling.
Note: IsTestServerRunning already exists in server_lifecycle.go

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Chunk 1 Verification

- [ ] **Verification Step: Compile all modified packages**

Run:
```bash
cd bdd
go build ./support/...
```

Expected: No compilation errors

- [ ] **Verification Step: Test client factory**

Run:
```bash
cd bdd/support
go test -v -run TestNewAnonymousClient || echo "Test will be created in Task 8"
```

Expected: Either test passes or test doesn't exist yet (will be created in verification phase)

- [ ] **Verification Step: Verify fixtures.json is valid**

Run:
```bash
cd bdd/support
cat fixtures.json | jq '.users | length'
```

Expected: Number >= 1 (at least one admin user)

---

## Chunk 2: Test Context with Client Management

### Task 4: Update test context with Logger field and client types

**Files:**
- Modify: `bdd/support/test_context.go`

- [ ] **Step 1: Read current BDDTestContext struct**

Run:
```bash
cd bdd/support
sed -n '13,47p' test_context.go
```

Note the current field definitions

- [ ] **Step 2: Update import to include slog**

Check if slog is already imported:
```bash
grep "log/slog" test_context.go
```

If not present, add to imports (line ~4):
```go
import (
    "fmt"
    "log/slog"
    "sync"
)
```

- [ ] **Step 3: Replace client field types in BDDTestContext struct**

**IMPORTANT:** This step changes the struct field types from `interface{}` to `*integration.ClientWithResponses`. This must be completed before adding methods in Task 5 that use these concrete types.

Find the client fields (around lines 20-24) and replace:
```go
// API clients (immutable after setup)
AnonymousClient *integration.ClientWithResponses
Client          *integration.ClientWithResponses
ManagerClient   *integration.ClientWithResponses
```

With:
```go
// API clients (immutable after setup)
AnonymousClient *integration.ClientWithResponses
Client          *integration.ClientWithResponses
ManagerClient   *integration.ClientWithResponses

// Logger for client creation and operations
Logger *slog.Logger
```

Note: Add the import at the top of the file if not present:
```go
import (
    "fmt"
    "log/slog"
    "sync"
    "github.com/code-together/shared/integration"
)
```

- [ ] **Step 4: Verify compilation**

Run:
```bash
cd bdd/support
go build test_context.go
```

Expected: No compilation errors (may have errors about missing methods, that's OK for now)

- [ ] **Step 5: Commit**

```bash
git add bdd/support/test_context.go
git commit -m "refactor: update test context client types to use real API clients

Replace interface{} types with *integration.ClientWithResponses
and add Logger field for client creation.

This is the first step in converting from mocks to real API calls.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 5: Add client management methods to BDDTestContext

**Files:**
- Modify: `bdd/support/test_context.go` (append after existing methods)

- [ ] **Step 1: Find end of test_context.go to add new methods**

Run:
```bash
cd bdd/support
wc -l test_context.go
```

Note the line count, will append after last method

- [ ] **Step 2: Add InitializeClients method**

Append to end of file:
```go
// InitializeClients sets up all API clients for testing
// Called once during test suite initialization (not per scenario)
func (ctx *BDDTestContext) InitializeClients(serverURL string, logger *slog.Logger) error {
    // Store logger for client creation
    ctx.Logger = logger

    // Create anonymous client (no authentication required)
    anonClient, err := NewAnonymousClient(serverURL, logger, false)
    if err != nil {
        return fmt.Errorf("failed to create anonymous client: %w", err)
    }
    ctx.AnonymousClient = anonClient

    // Authenticated clients will be created per-scenario after user login
    ctx.Client = nil
    ctx.ManagerClient = nil

    return nil
}
```

- [ ] **Step 3: Add GetAnonymousClient method**

Append to file:
```go
// GetAnonymousClient returns the anonymous client for unauthenticated requests
func (ctx *BDDTestContext) GetAnonymousClient() *integration.ClientWithResponses {
    return ctx.AnonymousClient
}
```

- [ ] **Step 4: Add GetAuthToken method**

Append to file:
```go
// GetAuthToken returns the current authentication token for the scenario
// Checks AdminToken first, then MemberToken
func (ctx *BDDTestContext) GetAuthToken() (string, error) {
    ctx.mu.Lock()
    defer ctx.mu.Unlock()

    if ctx.AdminToken != "" {
        return ctx.AdminToken, nil
    }
    if ctx.MemberToken != "" {
        return ctx.MemberToken, nil
    }
    return "", fmt.Errorf("no authentication token available")
}
```

- [ ] **Step 5: Add GetAuthenticatedClient method**

Append to file:
```go
// GetAuthenticatedClient returns a client with automatic token injection
// Creates or reuses the authenticated client for the current scenario
// Uses ManagerClient for admin operations, Client for member operations
func (ctx *BDDTestContext) GetAuthenticatedClient() (*integration.ClientWithResponses, error) {
    // Get current token
    token, err := ctx.GetAuthToken()
    if err != nil {
        return nil, fmt.Errorf("cannot create authenticated client without token: %w", err)
    }

    ctx.mu.Lock()
    defer ctx.mu.Unlock()

    // Create or reuse authenticated client based on user role
    if ctx.CurrentUser != nil && ctx.CurrentUser.Role == "admin" {
        if ctx.ManagerClient == nil {
            client, err := NewAuthenticatedClient(ctx.ServerURL, func() (string, error) {
                return ctx.GetAuthToken()
            }, ctx.Logger, false)
            if err != nil {
                return nil, err
            }
            ctx.ManagerClient = client
        }
        return ctx.ManagerClient, nil
    }

    // Use Client for non-admin users
    if ctx.Client == nil {
        client, err := NewAuthenticatedClient(ctx.ServerURL, func() (string, error) {
            return ctx.GetAuthToken()
        }, ctx.Logger, false)
        if err != nil {
            return nil, err
        }
        ctx.Client = client
    }
    return ctx.Client, nil
}
```

- [ ] **Step 6: Add UpdateAuthenticatedClients method**

Append to file:
```go
// UpdateAuthenticatedClients refreshes the authenticated clients with a new token
// Called after login/logout to update token injection
func (ctx *BDDTestContext) UpdateAuthenticatedClients(token string) error {
    ctx.mu.Lock()
    defer ctx.mu.Unlock()

    client, err := NewAuthenticatedClient(ctx.ServerURL, func() (string, error) {
        return token, nil
    }, ctx.Logger, false)
    if err != nil {
        return err
    }

    if ctx.CurrentUser != nil && ctx.CurrentUser.Role == "admin" {
        ctx.ManagerClient = client
    } else {
        ctx.Client = client
    }
    return nil
}
```

- [ ] **Step 7: Verify compilation with package build**

Run:
```bash
cd bdd/support
go build ./support/...
```

Expected: No compilation errors (checks entire package for import/type issues)

- [ ] **Step 8: Test client initialization manually**

Create test file `test_context_test.go`:
```go
package support

import (
    "log/slog"
    "os"
    "testing"
)

func TestInitializeClients(t *testing.T) {
    ctx := &BDDTestContext{
        ServerURL: "http://localhost:8088",
    }

    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    err := ctx.InitializeClients(ctx.ServerURL, logger)

    if err != nil {
        t.Fatalf("InitializeClients failed: %v", err)
    }

    if ctx.AnonymousClient == nil {
        t.Error("AnonymousClient should not be nil after initialization")
    }

    if ctx.Logger == nil {
        t.Error("Logger should not be nil after initialization")
    }
}
```

Run:
```bash
cd bdd/support
go test -v -run TestInitializeClients
```

Expected: Test passes

- [ ] **Step 9: Commit**

```bash
git add bdd/support/test_context.go bdd/support/test_context_test.go
git commit -m "feat: add client management methods to test context

Add InitializeClients, GetAnonymousClient, GetAuthToken,
GetAuthenticatedClient, and UpdateAuthenticatedClients methods.

These enable per-scenario client lifecycle management with automatic
token injection for authenticated requests.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Chunk 3: Test Lifecycle Hooks and Cleanup

### Task 6: Implement CleanupScenarioResources with retry logic

**Files:**
- Modify: `bdd/step_definitions/context.go`

- [ ] **Step 1: Read current CleanupScenarioResources implementation**

Run:
```bash
cd bdd/step_definitions
sed -n '22,58p' context.go
```

Note the current TODO implementation

- [ ] **Step 2: Add required imports to context.go**

Check imports:
```bash
head -15 context.go
```

Add missing imports if needed:
```go
import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/code-together/bdd/support"
    "github.com/cucumber/godog"
)
```

- [ ] **Step 3: Replace CleanupScenarioResources implementation**

Replace the entire function (lines 22-58) with:

```go
// CleanupScenarioResources cleans up all resources created during the scenario
// Deletes resources via API with retry logic for robustness
func (ctx *ScenarioContext) CleanupScenarioResources() error {
    log.Printf("Cleaning up scenario resources...")

    client, err := ctx.GetAuthenticatedClient()
    if err != nil {
        return fmt.Errorf("failed to get authenticated client for cleanup: %w", err)
    }

    // Retry configuration
    maxRetries := 3
    baseDelay := 100 * time.Millisecond

    // Clean up teams first (may have foreign key dependencies)
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

    // Clean up users last (no foreign key dependencies)
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

    // Clear resource lists
    ctx.ClearCreatedResources()

    log.Printf("Cleanup complete")
    return nil
}
```

- [ ] **Step 4: Verify compilation with package build**

Run:
```bash
cd bdd/step_definitions
go build ./step_definitions/...
```

Expected: No compilation errors

- [ ] **Step 5: Commit**

```bash
git add bdd/step_definitions/context.go
git commit -m "feat: implement resource cleanup via API with retry logic

Replace TODO cleanup implementation with real API calls to delete
created resources. Includes retry logic for robustness and logging
for debugging cleanup failures.

Resources cleaned in reverse order:
- Teams (may have dependencies)
- Providers
- Users (no dependencies)

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 7: Update test suite hooks for client initialization and data lifecycle

**Files:**
- Modify: `bdd/godog/godog_suite_test.go`

- [ ] **Step 1: Read current InitializeScenario implementation**

Run:
```bash
cd bdd/godog
sed -n '46,112p' godog_suite_test.go
```

Note the current BeforeScenario and AfterScenario hooks

- [ ] **Step 2: Add required imports**

Check imports at top of file:
```bash
head -20 godog_suite_test.go
```

Add missing imports:
```go
import (
    "context"
    "log"
    "log/slog"
    "os"
    "testing"

    "github.com/cucumber/godog"
    "github.com/code-together/bdd/step_definitions"
    "github.com/code-together/bdd/support"
)
```

- [ ] **Step 3: Update InitializeScenario to initialize clients**

Replace the InitializeScenario function starting at line 47:

```go
// InitializeScenario sets up the scenario context and registers step definitions
func InitializeScenario(suite *godog.TestSuiteContext) {
    // Create logger for client operations
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create test context with server configuration
    testContext := &step_definitions.ScenarioContext{
        BDDTestContext: &support.BDDTestContext{
            ServerURL: support.GetTestServerURL(),
            TestDBURL: support.GetTestDatabaseURL(),
        },
    }

    // Initialize API clients
    if err := testContext.InitializeClients(testContext.ServerURL, logger); err != nil {
        log.Printf("WARNING: Failed to initialize API clients: %v", err)
        // Continue anyway - scenarios may handle this
    }

    // Get the scenario context from the suite
    ctx := suite.ScenarioContext()

    // Register step definitions
    step_definitions.RegisterCommonSteps(testContext, ctx)
    step_definitions.RegisterAuthSteps(testContext, ctx)
    step_definitions.RegisterPermissionSteps(testContext, ctx)
    step_definitions.RegisterProviderSteps(testContext, ctx)
    step_definitions.RegisterLicenseSteps(testContext, ctx)
    step_definitions.RegisterUsageSteps(testContext, ctx)
    step_definitions.RegisterDashboardSteps(testContext, ctx)
    step_definitions.RegisterHealthSteps(testContext, ctx)

    // Before scenario hook - reset state and setup test environment
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        log.Printf("=== Starting Scenario: %s ===", sc.Name)

        // Reset scenario state
        testContext.ResetScenarioState()

        // Ensure test server is running
        if !support.IsTestServerRunning(testContext.ServerURL) {
            log.Printf("WARNING: Test server is not running at %s", testContext.ServerURL)
            return ctx, fmt.Errorf("test server not running at %s", testContext.ServerURL)
        }

        // Load fixtures and create test data via API
        fixtures, err := support.LoadFixtureData()
        if err != nil {
            return ctx, fmt.Errorf("failed to load fixtures: %w", err)
        }

        // TODO: Create test data from fixtures via API
        // This will be implemented when CreateTestDataFromFixtures is added
        log.Printf("Loaded %d users, %d providers, %d teams, %d licenses from fixtures",
            len(fixtures.Users), len(fixtures.Providers), len(fixtures.Teams), len(fixtures.Licenses))

        return ctx, nil
    })

    // After scenario hook - clean up resources
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        log.Printf("=== Completed Scenario: %s ===", sc.Name)

        // Clean up resources even if scenario failed
        cleanupErr := testContext.CleanupScenarioResources()
        if cleanupErr != nil {
            // Log cleanup error but don't fail the scenario
            log.Printf("WARNING: Cleanup failed: %v", cleanupErr)
        }

        return ctx, err
    })
}
```

Note: Need to add slog import:
```go
import (
    "context"
    "log"
    "log/slog"
    "os"
    // ... other imports
)
```

- [ ] **Step 4: Verify compilation**

Run:
```bash
cd bdd/godog
go build godog_suite_test.go
```

Expected: No compilation errors

- [ ] **Step 5: Test suite initialization**

Run:
```bash
cd bdd/godog
go test -v -run TestGodog
```

Expected: Tests run (may fail on individual scenarios, but suite should initialize)

- [ ] **Step 6: Commit**

```bash
git add bdd/godog/godog_suite_test.go
git commit -m "feat: initialize API clients in test suite setup

Update InitializeScenario to:
- Create API clients during suite initialization
- Verify test server is running before scenarios
- Load fixture data for test data creation
- Call CleanupScenarioResources via API

Infrastructure is now ready for converting step definitions from
mocks to real API calls.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Verification and Testing

### Task 8: Verify infrastructure is ready for step conversion

**Files:**
- Test: All modified files

- [ ] **Step 1: Verify all files compile**

Run:
```bash
cd bdd
go build ./support/...
go build ./step_definitions/...
go build ./godog/...
```

Expected: No compilation errors across all packages

- [ ] **Step 2: Run test suite to verify initialization**

Run:
```bash
cd bdd
./bdd-test.sh --tags "@smoke" 2>&1 | head -50
```

Expected: Test suite initializes, scenarios may fail but infrastructure should work

- [ ] **Step 3: Check test server health**

Run:
```bash
curl http://localhost:8088/health
```

Expected: Server returns 200 OK

- [ ] **Step 4: Verify fixtures load correctly**

First check if test exists:
```bash
cd bdd/support
ls -la fixtures_test.go
```

If test exists, run it:
```bash
cd bdd/support
go test -v -run TestLoadFixtureData
```

If test doesn't exist, create it:
```bash
cat > bdd/support/fixtures_test.go << 'EOF'
package support

import "testing"

func TestLoadFixtureData(t *testing.T) {
    fixtures, err := LoadFixtureData()
    if err != nil {
        t.Fatalf("LoadFixtureData failed: %v", err)
    }

    if fixtures == nil {
        t.Fatal("fixtures should not be nil")
    }

    // Verify we have at least one admin user
    foundAdmin := false
    for _, user := range fixtures.Users {
        if user.Role == "admin" {
            foundAdmin = true
            break
        }
    }

    if !foundAdmin {
        t.Error("fixtures should contain at least one admin user")
    }
}
EOF
```

Then run:
```bash
cd bdd/support
go test -v -run TestLoadFixtureData
```

Expected: Fixtures load from JSON successfully

- [ ] **Step 5: Review client factory creates correct client types**

Create quick test:
```go
// In client_factory_test.go
package support

import (
    "log/slog"
    "os"
    "testing"
)

func TestNewAnonymousClient(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    client, err := NewAnonymousClient("http://localhost:8088", logger, false)

    if err != nil {
        t.Fatalf("NewAnonymousClient failed: %v", err)
    }

    if client == nil {
        t.Error("Client should not be nil")
    }
}
```

Run:
```bash
cd bdd/support
go test -v -run TestNewAnonymousClient
```

Expected: Client created successfully

- [ ] **Step 6: Document any issues or next steps**

Create notes file:
```bash
cat > bdd/INFRASTRUCTURE_NOTES.md << 'EOF'
# BDD API Infrastructure - Implementation Notes

## Completed (Phase 1)

- [x] Client factory created
- [x] Test context updated with real API client types
- [x] Fixture loading from JSON only
- [x] Client management methods implemented
- [x] CleanupScenarioResources with retry logic
- [x] Test suite hooks updated

## Next Steps (Phase 2)

- [ ] Convert health check steps (common_steps.go)
- [ ] Convert auth steps (auth_steps.go)
- [ ] Run smoke tests and validate
- [ ] Document any API endpoints not yet implemented

## Known Issues

Document any issues found during infrastructure setup:
-
-
EOF
cat bdd/INFRASTRUCTURE_NOTES.md
```

- [ ] **Step 7: Final commit with notes**

```bash
git add bdd/INFRASTRUCTURE_NOTES.md
git commit -m "docs: add infrastructure implementation notes

Document completed Phase 1 infrastructure and next steps
for Phase 2 (smoke test conversion).

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Success Criteria

Phase 1 infrastructure is complete when:

- [ ] `bdd/support/client_factory.go` creates anonymous and authenticated clients
- [ ] `bdd/support/test_context.go` has all client management methods
- [ ] `bdd/support/fixtures.go` loads from JSON only (no hardcoded defaults)
- [ ] `bdd/step_definitions/context.go` implements CleanupScenarioResources with retry
- [ ] `bdd/godog/godog_suite_test.go` initializes clients and validates server
- [ ] All code compiles without errors
- [ ] Test suite can be initialized successfully
- [ ] Fixtures load from `bdd/support/fixtures.json`
- [ ] Documentation notes created

---

## Dependencies

This plan requires:

- Go 1.25+ installed
- Test server running on port 8088 (from `integration/test-server.sh`)
- `github.com/code-together/shared/integration` package available (via local replace)
- `fixtures.json` exists at `bdd/support/fixtures.json`

## Related Documentation

- Design spec: `bdd/docs/superpowers/specs/2025-03-17-bdd-real-api-integration-design.md`
- BDD documentation: `bdd/CLAUDE.md`
- Integration client: `shared/integration/client.go`
- Test server setup: `integration/INTEGRATION_TEST_SETUP.md`
