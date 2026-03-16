# BDD Test Implementation Design

**Date:** 2025-03-16
**Status:** Design Approved
**Priority:** High

## Executive Summary

This design document outlines the implementation of a comprehensive Behavior-Driven Development (BDD) test suite for the AI Together platform using the [godog](https://github.com/cucumber/godog) framework. The BDD suite will convert the existing 171 integration tests into ~247 Gherkin scenarios organized by domain, providing better test documentation, improved maintainability, and enhanced collaboration between technical and non-technical stakeholders.

**Key Objectives:**
- Convert 171 integration tests to BDD scenarios using godog
- Align test organization with EARS domain structure from `docs/ears/`
- Reuse existing integration test infrastructure (API clients, fixtures, test server)
- Provide executable specifications in plain language (Gherkin)
- Enable scenario-based testing with clear Given-When-Then structure

## Table of Contents

1. [Architecture](#architecture)
2. [Project Structure](#project-structure)
3. [Test Mapping Strategy](#test-mapping-strategy)
4. [Test Context & State Management](#test-context--state-management)
5. [Step Definitions](#step-definitions)
6. [Feature Files](#feature-files)
7. [Integration with Existing Infrastructure](#integration-with-existing-infrastructure)
8. [Test Execution](#test-execution)
9. [Implementation Plan](#implementation-plan)

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         BDD Test Suite                      │
│                      (godog + Go)                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Shared Integration Module                │
│              (github.com/code-together/shared)              │
│  • API Clients (Anonymous, Authenticated, Manager)          │
│  • Request/Response Types                                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Test Server (Port 8088)                 │
│                      (server module)                        │
│  • REST API Endpoints                                       │
│  • PostgreSQL Database                                      │
│  • JWT Authentication                                       │
└─────────────────────────────────────────────────────────────┘
```

### Design Principles

1. **Black-Box Testing** - Tests interact only through API endpoints
2. **Domain-Driven Organization** - Aligns with EARS requirements structure
3. **Reusable Components** - Leverages existing integration test infrastructure
4. **Clear Documentation** - Gherkin scenarios serve as executable specifications
5. **Independent Execution** - Separate Go module with own dependencies

## Project Structure

```
bdd/
├── go.mod                                    # BDD module (depends on integration module)
├── go.sum
├── Makefile                                  # Test execution commands
├── README.md                                 # BDD testing guide
├── features/                                 # Gherkin feature files by EARS domain
│   ├── 00_identity_and_access.feature        # Auth, users, roles, licensing, multi-tenancy
│   ├── 01_provider_management.feature        # Provider CRUD, testing, limits, connectivity
│   ├── 02_configuration_sync.feature         # Teams, config distribution (if tests exist)
│   ├── 03_usage_insights.feature             # Usage upload, analytics, aggregation
│   ├── 04_user_interfaces.feature            # Manager dashboard operations
│   └── 05_system_behaviors.feature           # Health checks, system behavior
├── step_definitions/                         # Godog step implementations
│   ├── context.go                            # Shared test context & state
│   ├── common_steps.go                       # Common steps (Given/When/Then)
│   ├── auth_steps.go                         # Authentication steps
│   ├── provider_steps.go                     # Provider management steps
│   ├── license_steps.go                      # Licensing steps
│   ├── usage_steps.go                        # Usage analytics steps
│   ├── permission_steps.go                   # RBAC steps
│   └── dashboard_steps.go                    # Manager dashboard steps
├── support/                                  # Test utilities
│   ├── test_context.go                       # Setup/teardown, server connection
│   ├── fixtures.go                           # Test data fixtures
│   └── helpers.go                            # Utility functions
└── godog/                                    # Test suite entry points
    ├── godog_suite_test.go                   # Testify suite wrapper
    └── godog_test.go                         # Main test runner
```

## Test Mapping Strategy

### Conversion Pattern

Each integration test method converts to one or more Gherkin scenarios:

**Integration Test:**
```go
func (s *IntegrationTestSuite) TestProviderCreateSuccess() {
    // Arrange: Setup admin client, unique provider name
    uniqueName := generateUniqueProviderName("test-provider")
    req := integration.PostApiV1ProvidersJSONRequestBody{
        Name:   uniqueName,
        Kind:   "claude",
        ApiKey: "sk-test-123",
    }

    // Act: Create provider via API
    resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
    require.NoError(s.T(), err)

    // Assert: Verify response
    assert.Equal(s.T(), 201, resp.StatusCode())
    assert.Equal(s.T(), uniqueName, resp.JSON201.Name)

    // Cleanup
    defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, resp.JSON201.Id)
}
```

**BDD Scenario:**
```gherkin
Scenario: Create a new Claude provider
    Given I am logged in as a manager
    And I have a unique provider name "test-provider-12345"
    When I create a Claude provider with:
        | name        | kind   | api_key      |
        | test-provider-12345 | claude | sk-test-123 |
    Then the provider should be created successfully
    And the provider name should be "test-provider-12345"
    And the provider kind should be "claude"
```

### Test Coverage Mapping

| Integration Test File | Feature File | Test Count | Scenario Count |
|----------------------|--------------|------------|----------------|
| `auth_test.go` | `00_identity_and_access.feature` | 16 | ~20 |
| `provider_test.go` | `01_provider_management.feature` | 20 | ~25 |
| `license_test.go` | `00_identity_and_access.feature` | 33 | ~40 |
| `user_test.go` | `00_identity_and_access.feature` | 7 | ~10 |
| `usage_test.go` | `03_usage_insights.feature` | 24 | ~30 |
| `analytics_*.go` | `03_usage_insights.feature` | 47 | ~60 |
| `permission_test.go` | `00_identity_and_access.feature` | 8 | ~10 |
| `mgr_*.go` | `04_user_interfaces.feature` | 20 | ~25 |
| `health_test.go` | `05_system_behaviors.feature` | 2 | ~3 |
| **Total** | | **171** | **~223** |

## Test Context & State Management

### Shared Test Context

```go
// step_definitions/context.go
package step_definitions

type BDDTestContext struct {
    // Server connection
    ServerURL       string
    TestDBURL       string

    // API clients
    AnonymousClient *integrationclient.ClientWithResponses
    Client          *integrationclient.ClientWithResponses
    ManagerClient   *integration_manager.ClientWithResponses

    // Authentication
    AdminToken      string
    MemberToken     string
    CurrentUser     *UserInfo

    // Test data storage (for sharing between steps)
    LastProviderID    int64
    LastUserID        string
    LastTeamID        int64
    LastLicenseID     string
    CreatedResourceIDs map[string]string

    // Response storage (for assertions)
    LastStatusCode    int
    LastResponse      interface{}
    LastErrorResponse string
}

type UserInfo struct {
    Email    string
    Password string
    Name     string
    Role     string
    Token    string
}
```

### Scenario Lifecycle

**Before Scenario:**
1. Reset test context state
2. Login as default user (manager)
3. Load standard fixtures

**After Scenario:**
1. Clean up created resources via API
2. Reset context to initial state

## Step Definitions

### Step Organization

Steps are organized by domain in separate files:

```go
// step_definitions/auth_steps.go
func RegisterAuthSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
    // Given steps - Setup context
    suite.Given(`^I am logged in as a manager$`, ctx.iAmLoggedInAsAManager)
    suite.Given(`^I am logged in as a member$`, ctx.iAmLoggedInAsAMember)

    // When steps - Perform actions
    suite.When(`^I login with email "([^"]*)" and password "([^"]*)"$`,
               ctx.iLoginWithCredentials)
    suite.When(`^I logout$`, ctx.iLogout)

    // Then steps - Assert outcomes
    suite.Then(`^I should receive a valid authentication token$`,
               ctx.iShouldReceiveValidToken)
    suite.Then(`^the response status code should be (\d+)$`,
               ctx.responseStatusCodeShouldBe)
}
```

### Step Implementation Pattern

```go
func (ctx *ScenarioContext) iCreateAProviderWithKindAndAPIKey(kind, apiKey string) error {
    // Generate unique name
    uniqueName := generateUniqueProviderName("test-provider")

    // Build request
    req := integrationclient.PostApiV1ProvidersJSONRequestBody{
        Name:  uniqueName,
        Kind:  kind,
        ApiKey: apiKey,
        Priority: intPointer(1),
        Enabled: boolPointer(true),
    }

    // Execute API call
    resp, err := ctx.Client.PostApiV1ProvidersWithResponse(
        context.Background(), req)
    if err != nil {
        return fmt.Errorf("failed to create provider: %w", err)
    }

    // Store response for subsequent assertions
    ctx.LastStatusCode = resp.StatusCode()
    if resp.JSON201 != nil {
        ctx.LastResponse = resp.JSON201
        ctx.LastProviderID = resp.JSON201.Id
    }

    return nil
}
```

## Feature Files

### Feature File Structure

Each feature file follows Gherkin standard with:
- Feature description (As a... I want... So that...)
- Background (common setup)
- Rules (grouping related scenarios)
- Scenarios (Given-When-Then steps)
- Examples (data tables for Scenario Outlines)

### Example Feature File

```gherkin
# features/01_provider_management.feature
Feature: Provider Management
  As a manager
  I want to configure and manage AI service providers
  So that my team can use AI tools with automatic failover

  Background:
    Given the test server is running
    And I am logged in as a manager
    And the standard fixtures are loaded

  Rule: Provider Creation

    Scenario Outline: Create provider with valid data
      Given I have a unique provider name
      When I create a <kind> provider with API key "sk-test-123"
      Then the provider should be created successfully
      And the provider kind should be "<kind>"

      Examples:
        | kind    |
        | claude  |
        | codex   |
        | opencode |

    Scenario: Create provider with duplicate name should fail
      Given I have a unique provider name "my-provider"
      And I create a claude provider with name "my-provider"
      When I create another claude provider with name "my-provider"
      Then I should receive a 400 error
```

### Tagging Strategy

- `@smoke` - Critical happy path tests
- `@critical` - Must-pass tests for production
- `@security` - Security and permission tests
- `@auth` - Authentication scenarios
- `@provider` - Provider management
- `@license` - License management
- `@usage` - Usage analytics
- `@dashboard` - Manager dashboard
- `@wip` - Work in progress (don't run in CI)

## Integration with Existing Infrastructure

### Module Dependencies

```go
// bdd/go.mod
module github.com/shtdu/bdd

go 1.24

require (
    github.com/cucumber/godog v0.14.1
    github.com/stretchr/testify v1.9.0
    github.com/code-together/shared v1.0.0
)

// Uses existing integration module
replace github.com/code-together/shared => ../shared
```

### Fixture Reuse

Fixtures are loaded from `integration/testdata/`:

```go
// support/fixtures.go
func LoadFixtureData() (*FixtureData, error) {
    fixturePath := filepath.Join("..", "integration",
                                  "testdata", "fixtures.json")
    data, err := os.ReadFile(fixturePath)
    if err != nil {
        return nil, err
    }

    var fixtures FixtureData
    if err := json.Unmarshal(data, &fixtures); err != nil {
        return nil, err
    }

    return &fixtures, nil
}
```

### Test Server Integration

The BDD suite uses the same test server as integration tests:

```bash
# Start test server (reused from integration/)
cd ../integration && ./test-server.sh

# Run BDD tests
cd bdd && make test-bdd
```

## Test Execution

### Command-Line Interface

```bash
# Run all BDD tests
make test-bdd

# Run specific feature file
make test-bdd features/01_provider_management.feature

# Run with specific tags
make test-bdd-tags tags="@smoke && @critical"

# Run with coverage
make test-bdd-coverage

# Run in CI mode (JUnit output)
make test-bdd-ci
```

### CI/CD Integration

```yaml
# .github/workflows/bdd-tests.yml
name: BDD Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  bdd-tests:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_DB: codetogether_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test

    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run BDD tests
        run: |
          cd bdd
          make test-bdd-ci
```

## Implementation Plan

### Phase 1: Foundation Setup (Week 1)

**Goals:** Establish project structure, tooling, and basic infrastructure

**Tasks:**
- [ ] Initialize BDD Go module with godog dependency
- [ ] Create directory structure (features/, step_definitions/, support/, godog/)
- [ ] Implement shared test context (`step_definitions/context.go`)
- [ ] Adapt fixtures and helpers from integration/
- [ ] Create Makefile with test commands
- [ ] Set up test server integration
- [ ] Write README with documentation
- [ ] Create first "hello world" scenario

**Deliverables:**
- Working `go test github.com/shtdu/bdd/...`
- Test server starts and connects
- First scenario runs successfully

### Phase 2: Core Authentication & User Management (Week 2)

**Goals:** Convert auth and user tests (~27 scenarios)

**Tasks:**
- [ ] Create `features/00_identity_and_access.feature`
- [ ] Implement `step_definitions/auth_steps.go`
- [ ] Implement `step_definitions/common_steps.go`
- [ ] Convert 16 auth integration tests → BDD scenarios
- [ ] Convert 7 user integration tests → BDD scenarios
- [ ] Add background scenarios for setup
- [ ] Implement scenario hooks (before/after)
- [ ] Verify all auth scenarios pass

**Scenarios:** ~27 scenarios covering login, logout, token refresh, user profiles, permissions

### Phase 3: Provider Management (Week 3)

**Goals:** Convert provider tests (~35 scenarios)

**Tasks:**
- [ ] Create `features/01_provider_management.feature`
- [ ] Implement `step_definitions/provider_steps.go`
- [ ] Convert 20 provider integration tests → BDD scenarios
- [ ] Add data table examples for different provider types
- [ ] Implement CRUD scenarios
- [ ] Add connectivity testing scenarios
- [ ] Verify all provider scenarios pass

**Scenarios:** ~35 scenarios covering provider CRUD, validation, limits, connectivity

### Phase 4: License Management (Week 4)

**Goals:** Convert license tests (~50 scenarios)

**Tasks:**
- [ ] Extend `features/00_identity_and_access.feature` with license section
- [ ] Implement `step_definitions/license_steps.go`
- [ ] Convert 33 license integration tests → BDD scenarios
- [ ] Add license activation scenarios
- [ ] Add tier-based limit scenarios
- [ ] Add expiration and renewal scenarios
- [ ] Verify all license scenarios pass

**Scenarios:** ~50 scenarios covering license activation, tier limits, expiration

### Phase 5: Usage Analytics (Week 5)

**Goals:** Convert usage and analytics tests (~85 scenarios)

**Tasks:**
- [ ] Create `features/03_usage_insights.feature`
- [ ] Implement `step_definitions/usage_steps.go`
- [ ] Convert 24 usage integration tests → BDD scenarios
- [ ] Convert 47 analytics integration tests → BDD scenarios
- [ ] Add aggregation scenarios
- [ ] Add filtering and query scenarios
- [ ] Add cost calculation scenarios
- [ ] Verify all usage scenarios pass

**Scenarios:** ~85 scenarios covering usage upload, aggregation, filtering, costs

### Phase 6: Manager Dashboard & Permissions (Week 6)

**Goals:** Convert dashboard and permission tests (~45 scenarios)

**Tasks:**
- [ ] Create `features/04_user_interfaces.feature`
- [ ] Implement `step_definitions/dashboard_steps.go`
- [ ] Implement `step_definitions/permission_steps.go`
- [ ] Convert 20 manager integration tests → BDD scenarios
- [ ] Convert 8 permission integration tests → BDD scenarios
- [ ] Add team management scenarios
- [ ] Add user management scenarios
- [ ] Verify all dashboard scenarios pass

**Scenarios:** ~45 scenarios covering teams, users, dashboard, permissions

### Phase 7: System Behaviors & Health Checks (Week 7)

**Goals:** Convert health check tests (~5 scenarios)

**Tasks:**
- [ ] Create `features/05_system_behaviors.feature`
- [ ] Convert 2 health integration tests → BDD scenarios
- [ ] Add system behavior scenarios
- [ ] Verify all system behavior scenarios pass

**Scenarios:** ~5 scenarios covering health checks, system readiness

### Phase 8: Documentation & Polish (Week 8)

**Goals:** Complete documentation, examples, and CI/CD integration

**Tasks:**
- [ ] Write comprehensive README
- [ ] Create contributing guide
- [ ] Add example scenarios for each domain
- [ ] Set up CI/CD pipeline (.github/workflows/bdd-tests.yml)
- [ ] Add coverage reporting
- [ ] Performance optimization
- [ ] Code review and refinement

**Deliverables:**
- Complete documentation
- CI/CD pipeline running
- Coverage reports
- Example scenarios

### Implementation Timeline

| Phase | Duration | Scenarios | Focus |
|-------|----------|-----------|-------|
| 1 | 1 week | Foundation | Project setup |
| 2 | 1 week | ~27 | Auth & Users |
| 3 | 1 week | ~35 | Providers |
| 4 | 1 week | ~50 | Licenses |
| 5 | 1 week | ~85 | Usage & Analytics |
| 6 | 1 week | ~45 | Dashboard & Permissions |
| 7 | 1 week | ~5 | System Behaviors |
| 8 | 1 week | - | Documentation & Polish |
| **Total** | **8 weeks** | **~247** | **Complete BDD suite** |

## Success Criteria

The BDD test implementation will be considered successful when:

1. ✅ All 171 integration tests have corresponding BDD scenarios
2. ✅ All BDD scenarios pass consistently
3. ✅ Test coverage matches or exceeds integration test coverage
4. ✅ Documentation is complete and clear
5. ✅ CI/CD pipeline runs BDD tests on every push
6. ✅ Team can run BDD tests locally with simple commands
7. ✅ Scenarios serve as clear, executable specifications

## References

- [Godog Framework](https://github.com/cucumber/godog)
- [Gherkin Syntax](https://cucumber.io/docs/gherkin/)
- [EARS Requirements](docs/ears/)
- [Integration Tests](integration/)
- [Product Requirements](vibe_doc/)

---

**Document Status:** Design Approved
**Next Step:** Invoke writing-plans skill to create detailed implementation plan
