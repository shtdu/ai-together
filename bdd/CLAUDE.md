# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Behavior-Driven Development (BDD) test suite** for the AI Together platform, built using the [godog](https://github.com/cucumber/godog) framework. The suite converts integration tests into Gherkin scenarios that serve as executable specifications.

**Key characteristic:** This is an independent Go module with local module replacements for development, located in a git worktree for isolated feature development.

## Module Architecture

### Local Module Replacements

The `go.mod` file uses local module replacements to import sibling modules during development:

```go
replace github.com/code-together/shared => ../shared
replace github.com/code-together/integration => ../integration
```

**Important:** These replacements assume a specific directory structure relative to the main monorepo. When working in this worktree, the relative paths point to the parent directories.

### Directory Structure

```
bdd/
├── features/                   # Gherkin .feature files by domain
│   ├── hello_world.feature     # Initial infrastructure validation
│   ├── 00_identity_and_access.feature    # Auth, users, roles
│   ├── 01_provider_management.feature    # Provider CRUD operations
│   ├── 03_usage_insights.feature         # Analytics scenarios
│   ├── 04_user_interfaces.feature        # Dashboard scenarios
│   └── 05_system_behaviors.feature       # Health checks
├── step_definitions/           # Step implementations (Go)
│   ├── context.go              # ScenarioContext wrapper
│   ├── common_steps.go         # Generic steps
│   ├── auth_steps.go           # Authentication steps
│   ├── permission_steps.go     # RBAC steps
│   ├── provider_steps.go       # Provider management
│   ├── license_steps.go        # License management
│   ├── usage_steps.go          # Usage analytics
│   ├── dashboard_steps.go      # Dashboard operations
│   ├── health_steps.go         # Health checks
│   └── resource_tracking.go    # Cleanup utilities
├── support/                    # Test infrastructure
│   ├── test_context.go         # BDDTestContext (state management)
│   ├── fixtures.go             # Test data fixtures
│   ├── helpers.go              # Utility functions
│   └── server_lifecycle.go     # Server health/start/stop
├── godog/                      # Test suite entry point
│   ├── godog_suite_test.go     # Scenario initialization
│   └── godog_test.go           # Test runner
└── bdd-test.sh                 # Test execution script
```

## Test Execution Commands

### Common Commands

```bash
# Run all BDD tests (auto-starts server if needed)
./bdd-test.sh

# Run with specific tags
./bdd-test.sh --tags "@smoke"

# Run for CI/CD (JUnit output)
./bdd-test.sh --format junit

# Run without auto-starting server (must be running already)
./bdd-test.sh --no-server

# Run single feature file
cd godog && go test -v -godog.paths="../features/hello_world.feature"
```

### Test Server Management

```bash
# Check server health
curl http://localhost:8088/health

# View server logs
cat /tmp/test-server.log
```

### Development Commands

```bash
go fmt ./...          # Format Go code
go vet ./...          # Vet Go code
go mod download       # Download dependencies
go mod tidy           # Tidy dependencies
```

## Core Architecture Patterns

### 1. Test Context Hierarchy

**BDDTestContext** (`support/test_context.go`):
- Thread-safe context using `sync.Mutex`
- Stores API clients, authentication state, test data
- Tracks created resources for cleanup
- Each scenario gets its own instance

**ScenarioContext** (`step_definitions/context.go`):
- Wraps `BDDTestContext` with scenario-specific methods
- Provides `ResetScenarioState()` and `CleanupScenarioResources()`
- Acts as the receiver for all step definition methods

### 2. Step Definition Registration Pattern

All step definitions follow this pattern:

```go
// Register function called from godog_suite_test.go
func RegisterAuthSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
    suite.Given(`^I am logged in as a manager$`, ctx.iAmLoggedInAsAManager)
    suite.When(`^I login with email "([^"]*)"$`, ctx.iLoginWithCredentials)
    suite.Then(`^the response status code should be (\d+)$`, ctx.responseStatusCodeShouldBe)
}

// Implementation method on ScenarioContext
func (ctx *ScenarioContext) iAmLoggedInAsAManager() error {
    // Implementation
    return nil
}
```

**Registration happens in `godog/godog_suite_test.go`:**
```go
step_definitions.RegisterCommonSteps(testContext, ctx)
step_definitions.RegisterAuthSteps(testContext, ctx)
step_definitions.RegisterPermissionSteps(testContext, ctx)
// ... etc
```

### 3. Scenario Lifecycle Hooks

**Before Scenario** (`godog_suite_test.go`):
```go
ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    testContext.ResetScenarioState()
    // Verify test server is running
    // Load fixture data
    return ctx, nil
})
```

**After Scenario:**
```go
ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
    testContext.CleanupScenarioResources()  // Clean up even on failure
    return ctx, err
})
```

### 4. State Management Pattern

**Thread-safe access:**
```go
// Store response
ctx.SetLastResponse(200, responseData, "")

// Retrieve response
statusCode, resp, errMsg := ctx.GetLastResponse()

// Track resources for cleanup
ctx.TrackProvider(providerID)
ctx.TrackUser(userID)
ctx.TrackTeam(teamID)
```

**State isolation:** Each scenario has its own `BDDTestContext` instance, preventing state leakage between scenarios.

### 5. Resource Tracking and Cleanup

Resources are tracked during scenarios and cleaned up automatically:

```go
// In step definitions
ctx.TrackProvider(providerID)

// Cleanup happens in AfterScenario hook
func (ctx *ScenarioContext) CleanupScenarioResources() error {
    providers := ctx.GetCreatedProviders()
    for _, providerID := range providers {
        // Delete via API
    }
    ctx.ClearCreatedResources()
    return nil
}
```

**Important:** The default team (ID: 1) is never deleted during cleanup.

## Test Server Lifecycle

### Server Configuration

- **Default URL:** `http://localhost:8088`
- **Health endpoint:** `/health`
- **Log file:** `/tmp/test-server.log`
- **Startup script:** `../integration/test-server.sh`

### Health Check Pattern

```go
// Check if server is running
if !support.IsTestServerRunning(ctx.ServerURL) {
    // Handle server not running
}

// Wait for server with timeout
err := support.WaitForTestServer(serverURL, 30*time.Second)
```

### Server Auto-start

The bdd-test.sh script automatically starts the test server if it's not running:

```bash
# Default behavior: auto-start server if needed
./bdd-test.sh

# Skip auto-start if server is already managed manually
./bdd-test.sh --no-server
```

The script follows the same pattern as integration-test.sh:
- Checks if server is running before starting
- Starts server in background with PID tracking
- Waits for server to be ready (up to 30 seconds)
- Gracefully shuts down server on exit (SIGTERM → SIGKILL)
- Handles Ctrl+C for cleanup

## Writing New Scenarios

### 1. Add Feature File

Create or edit a `.feature` file in `features/`:

```gherkin
Feature: Provider Management
  As a manager
  I want to configure AI service providers
  So that my team can use AI tools

  @smoke
  Scenario: Create a new provider
    Given I am logged in as a manager
    And I have a unique provider name "test-provider"
    When I create a Claude provider with API key "sk-test-123"
    Then the provider should be created successfully
    And the provider kind should be "claude"
```

### 2. Implement Step Definitions

Add implementations in the appropriate `step_definitions/*.go` file:

```go
suite.When(`^I create a ([^"]*) provider with API key "([^"]*)"$`,
    ctx.iCreateAProviderWithKindAndAPIKey)

func (ctx *ScenarioContext) iCreateAProviderWithKindAndAPIKey(kind, apiKey string) error {
    // Call API
    // Track resource for cleanup
    // Store response for assertions
    return nil
}
```

### 3. Register Steps

Add the registration call to `godog/godog_suite_test.go`:

```go
step_definitions.RegisterProviderSteps(testContext, ctx)
```

## Test Data Utilities

### Unique Data Generation

```go
support.GenerateUniqueProviderName("test")  // "test-1648123456789"
support.GenerateUniqueEmail("user")         // "user-a1b2c3d4@example.com"
support.GenerateUniqueTeamName("team")      // "team-20250317-150405"
support.GenerateUniqueLicenseKey()          // "test-license-uuid"
```

### Test Constants

Defined in `support/helpers.go`:

```go
const (
    TestAPIKeyClaude   = "sk-test-claude-123"
    TestPasswordStrong = "TestPassword123!"
    ProviderKindClaude = "claude"
    RoleAdmin = "admin"
    TierProfessional = "professional"
)
```

### Validation Functions

```go
support.IsValidProviderKind(kind)   // bool
support.IsValidRole(role)           // bool
support.IsValidTier(tier)           // bool
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TEST_SERVER_URL` | Test server URL | `http://localhost:8088` |
| `TEST_DATABASE_URL` | Test database URL | `postgres://test:test@localhost:5432/codetogether_test?sslmode=disable` |
| `TEST_SERVER_LOG_PATH` | Test server log file | `/tmp/test-server.log` |

## Concurrency and Thread Safety

**Critical limitation:** The `BDDTestContext` is **NOT thread-safe for parallel scenario execution**.

- Each scenario gets its own context instance (safe)
- Within a scenario, `sync.Mutex` protects concurrent access
- Parallel scenario execution is **not currently supported**

## Common Workflows

### Adding a New Domain

1. Create `features/XX_domain_name.feature`
2. Create `step_definitions/domain_steps.go`
3. Implement `RegisterDomainSteps()` function
4. Register in `godog/godog_suite_test.go`

### Debugging Test Failures

```bash
# Check server health
curl http://localhost:8088/health

# View server logs
cat /tmp/test-server.log

# Run single scenario with tags
./bdd-test.sh --tags "@smoke and @failing"

# Run single feature file
cd godog && go test -v -godog.paths="../features/hello_world.feature"
```

### Resetting Test Database

If cleanup fails or resources accumulate:

```bash
cd ../integration
make integration-clean    # Drop test database
make integration-setup    # Recreate test database
```

## Implementation Status

**All 8 phases complete** (~181 scenarios):
- ✅ Phase 1: Foundation Setup
- ✅ Phase 2: Authentication & Users (30 scenarios)
- ✅ Phase 3: Provider Management (31 scenarios)
- ✅ Phase 4: License Management (29 scenarios)
- ✅ Phase 5: Usage Analytics (58 scenarios)
- ✅ Phase 6: Dashboard & Permissions (28 scenarios)
- ✅ Phase 7: System Behaviors (5 scenarios)
- ✅ Phase 8: Documentation & Polish

## Dependencies

**Go version:** 1.25.0+

**Key dependencies:**
- `github.com/cucumber/godog` - BDD test framework
- `github.com/cucumber/gherkin/go/v26` - Gherkin parser
- `github.com/google/uuid` - UUID generation for test data

**Local modules:**
- `github.com/code-together/shared` - Shared types and utilities
- `github.com/code-together/integration` - Integration test infrastructure

## Related Documentation

- [BDD Test Suite README](README.md) - User-facing documentation
- [Integration Test Setup](../integration/INTEGRATION_TEST_SETUP.md) - Test database and server setup
- [BDD Implementation Design](../docs/superpowers/specs/2025-03-16-bdd-test-implementation-design.md) - Design document
- [Godog Framework](https://github.com/cucumber/godog) - Official documentation
