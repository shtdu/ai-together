# BDD Test Suite

Behavioral-Driven Development (BDD) tests for the AI Together platform using the [godog](https://github.com/cucumber/godog) framework.

## Overview

This BDD test suite converts existing integration tests into Gherkin scenarios that serve as executable specifications. The tests are organized by domain and align with the EARS requirements structure.

## Prerequisites

- Go 1.24+
- PostgreSQL 14+ (for test database)
- Access to AI Together codebase
- Test server running or auto-start capability

## Quick Start

### Run All BDD Tests

```bash
make test-bdd
```

### Run Specific Feature

```bash
cd godog && go test -v -godog.paths="../features/hello_world.feature"
```

### Run with Specific Tags

```bash
make test-bdd-tags tags="@smoke"
```

### Run with Coverage

```bash
make test-bdd-coverage
```

## Project Structure

```
bdd/
├── features/                    # Gherkin feature files by domain
│   ├── hello_world.feature      # First test scenario
│   ├── 00_identity_and_access.feature    # Auth, users, roles, licensing
│   ├── 01_provider_management.feature    # Provider CRUD, limits, connectivity
│   ├── 03_usage_insights.feature         # Usage analytics
│   ├── 04_user_interfaces.feature        # Manager dashboard
│   └── 05_system_behaviors.feature       # Health checks
├── step_definitions/            # Step implementations
│   ├── context.go               # Scenario context wrapper
│   ├── common_steps.go          # Common Given/When/Then steps
│   ├── auth_steps.go            # Authentication steps ✓
│   ├── permission_steps.go      # RBAC steps ✓
│   ├── provider_steps.go        # Provider management steps ✓
│   ├── license_steps.go         # License management steps ✓
│   ├── usage_steps.go           # Usage analytics steps ✓
│   ├── dashboard_steps.go       # Dashboard steps ✓
│   ├── health_steps.go          # Health check steps ✓
│   └── resource_tracking.go     # Resource tracking utilities ✓
├── support/                     # Test utilities
│   ├── test_context.go          # Shared test context
│   ├── fixtures.go              # Test data fixtures
│   ├── helpers.go               # Utility functions
│   └── server_lifecycle.go      # Server lifecycle management
├── godog/                       # Test suite entry points
│   ├── godog_suite_test.go      # Test suite initializer
│   └── godog_test.go            # Main test runner
├── Makefile                     # Test execution commands
└── README.md                    # This file
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make test-bdd` | Run all BDD tests |
| `make test-bdd-ci` | Run with JUnit output for CI/CD |
| `make test-bdd-coverage` | Run with coverage report |
| `make test-bdd-tags tags="@smoke"` | Run with specific tags |
| `make ensure-test-server` | Ensure test server is running |
| `make clean` | Clean test artifacts |
| `make deps` | Download Go dependencies |
| `make fmt` | Format Go code |
| `make vet` | Vet Go code |
| `make lint` | Run all linters |
| `make test-server-logs` | Show test server logs |
| `make test-server-health` | Check test server health |

## Writing New Scenarios

### 1. Create or Edit Feature File

Create a new `.feature` file in `features/` or edit an existing one:

```gherkin
Feature: Provider Management
  As a manager
  I want to configure AI service providers
  So that my team can use AI tools

  Scenario: Create a new provider
    Given I am logged in as a manager
    And I have a unique provider name "my-provider"
    When I create a Claude provider with API key "sk-test-123"
    Then the provider should be created successfully
    And the provider kind should be "claude"
```

### 2. Implement Step Definitions

Add step implementations in `step_definitions/`:

```go
suite.When(`^I create a ([^"]*) provider with API key "([^"]*)"$`,
    ctx.iCreateAProviderWithKindAndAPIKey)
```

### 3. Run the Scenario

```bash
make test-bdd
```

## Test Server

The BDD tests require a running test server. The Makefile will automatically start the server if it's not running.

### Manual Server Start

```bash
cd ../integration
./test-server.sh
```

### Check Server Status

```bash
curl http://localhost:8088/health
```

## Test Data

### Fixtures

Test fixtures are loaded from `../integration/testdata/fixtures.json`.

### Unique Data Generation

Use helper functions to generate unique test data:

```go
support.GenerateUniqueProviderName("test-provider")  // "test-provider-1648123456789"
support.GenerateUniqueEmail("user")                  // "user-a1b2c3d4@example.com"
```

## Concurrency Safety

The `BDDTestContext` uses `sync.Mutex` to protect concurrent access within a scenario. Each scenario gets its own context instance.

**Important:** The test context is not thread-safe for parallel scenario execution. Parallel execution is not currently supported.

## Cleanup

Resources created during scenarios are automatically cleaned up:

- Providers are deleted via API
- Users are deleted via API
- Teams are deleted via API (except default team ID 1)

Cleanup failures are logged but don't fail the scenario.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TEST_SERVER_URL` | Test server URL | `http://localhost:8088` |
| `TEST_DATABASE_URL` | Test database URL | `postgres://test:test@localhost:5432/codetogether_test?sslmode=disable` |
| `TEST_SERVER_LOG_PATH` | Test server log file | `/tmp/test-server.log` |

## Troubleshooting

### Test Server Not Running

```bash
# Check server status
make test-server-health

# View server logs
make test-server-logs

# Manually start server
cd ../integration && ./test-server.sh
```

### Cleanup Failures

If cleanup fails, resources may accumulate in the test database. Reset the test database:

```bash
cd ../integration
make integration-clean
make integration-setup
```

### Dependency Issues

```bash
# Download and tidy dependencies
make deps
```

## Implementation Status

**All 8 Phases Complete! ✅**

✅ **Phase 1: Foundation Setup** - Module, test context, server lifecycle, first scenario
✅ **Phase 2: Authentication & Users** - 30 scenarios (Auth, Profiles, RBAC, Multi-tenant)
✅ **Phase 3: Provider Management** - 31 scenarios (CRUD, Limits, Connectivity)
✅ **Phase 4: License Management** - 29 license scenarios added
✅ **Phase 5: Usage Analytics** - 58 scenarios (Upload, Aggregation, Filtering, Costs)
✅ **Phase 6: Dashboard & Permissions** - 28 scenarios (Teams, Users, Dashboard, RBAC)
✅ **Phase 7: System Behaviors** - 5 scenarios (Health Checks, Readiness)
✅ **Phase 8: Documentation & Polish** - Complete documentation

**Total: ~181 BDD scenarios implemented across 8 phases**

## References

- [Design Document](../docs/superpowers/specs/2025-03-16-bdd-test-implementation-design.md)
- [Implementation Plan](../docs/superpowers/specs/2025-03-16-bdd-implementation-plan.md)
- [Godog Framework](https://github.com/cucumber/godog)
- [Gherkin Syntax](https://cucumber.io/docs/gherkin/)
- [Integration Tests](../integration/)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for BDD contribution guidelines.

## License

[Add license information]
