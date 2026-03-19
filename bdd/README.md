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
# First, start the test server in another terminal
cd ../integration && ./test-server.sh

# Then run BDD tests in the bdd directory
./bdd-test.sh
```

### Run Specific Feature

```bash
cd godog && go test -v ./...
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
│   ├── auth_steps.go            # Authentication & registration steps ✓
│   ├── permission_steps.go      # RBAC steps ✓
│   ├── provider_steps.go        # Provider management steps ✓
│   ├── license_steps.go         # License management & team limits ✓
│   ├── usage_steps.go           # Usage analytics & data retention ✓
│   ├── dashboard_steps.go       # Dashboard steps ✓
│   ├── health_steps.go          # Health check steps ✓
│   ├── invitation_steps.go      # Team invitation steps ✓
│   ├── password_reset_steps.go  # Password reset steps ✓
│   ├── user_steps.go            # User management & last manager protection ✓
│   └── resource_tracking.go     # Resource tracking utilities ✓
├── support/                     # Test utilities
│   ├── test_context.go          # Shared test context
│   ├── fixtures.go              # Test data fixtures
│   ├── helpers.go               # Utility functions
│   └── server_lifecycle.go      # Server lifecycle management
├── godog/                       # Test suite entry points
│   ├── godog_suite_test.go      # Test suite initializer
│   └── godog_test.go            # Main test runner
├── bdd-test.sh                  # Test execution script
└── README.md                    # This file
```

## Test Script Commands

| Command | Description |
|---------|-------------|
| `./bdd-test.sh` | Run all BDD tests (auto-starts server if needed) |
| `./bdd-test.sh --format junit` | Run with JUnit output for CI/CD |
| `./bdd-test.sh --tags "@smoke"` | Run scenarios with specific tags |
| `./bdd-test.sh --no-server` | Don't auto-start server (must be running) |
| `./bdd-test.sh --help` | Show all available options |

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
./bdd-test.sh
```

## Test Server

The BDD tests require a running test server. The script will automatically start the server if it's not running.

### Automatic Server Management (Default)

```bash
./bdd-test.sh
```

This will:
- Check if test server is running
- Start the server automatically if not running (via `../integration/test-server.sh`)
- Wait for server to be ready
- Execute BDD tests
- Stop the server automatically if we started it
- Handle Ctrl+C gracefully to cleanup

### Manual Server Management

If you prefer to manage the server manually:

```bash
# Terminal 1: Start server manually
cd ../integration
./test-server.sh

# Terminal 2: Run tests with --no-server flag
./bdd-test.sh --no-server
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
curl http://localhost:8088/health

# View server logs
cat /tmp/test-server.log

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

✅ **Phase 1: Foundation Setup** - Module, test context, server lifecycle, first scenario (2 scenarios)
✅ **Phase 2: Authentication & Users** - 79 scenarios (Auth, Registration, Lockout, Password Reset, Invitations, Last Manager, Team Limits)
✅ **Phase 3: Provider Management** - 24 scenarios (CRUD, Limits, Connectivity)
✅ **Phase 4: License Management** - Included in Phase 2 (team limits)
✅ **Phase 5: Usage Analytics** - 63 scenarios (Upload, Aggregation, Filtering, Costs, Data Retention)
✅ **Phase 6: Dashboard & Permissions** - 28 scenarios (Teams, Users, Dashboard, RBAC)
✅ **Phase 7: System Behaviors** - 6 scenarios (Health Checks, Readiness)
✅ **Phase 8: Documentation & Polish** - Complete documentation

**Total: 202 BDD scenarios across 8 phases**

### New Scenarios Added (Latest Implementation)

Added ~30 new scenarios addressing critical EARS requirements gaps:

| Domain | Feature | Scenarios | Status |
|--------|---------|-----------|--------|
| Identity & Access | User Registration | 6 | ✅ Complete |
| Identity & Access | Account Lockout | 4 | ✅ Complete |
| Identity & Access | Password Reset | 4 | ✅ Complete |
| Identity & Access | Team Invitations | 8 | ✅ Complete |
| Identity & Access | Last Manager Protection | 3 | ✅ Complete |
| License | Team Limits | 3 | ✅ Complete |
| Usage Insights | Data Retention | 5 | ✅ Complete |

### Current Test Results

| Metric | Status |
|--------|--------|
| **Total Scenarios** | 202 |
| **Passing** | ~179 (88.6%) |
| **Pending (@wip)** | ~23 (11.4%) |
| **Non-@wip Passing** | ~179/179 (100%) ✅ |

**Note:** The 23 @wip scenarios require backend API implementations:
- User registration endpoint
- Account lockout tracking and enforcement
- Password reset token generation and validation
- Team invitation email delivery
- Last manager protection validation
- Team limit enforcement by license tier
- Data retention cleanup jobs

See [BDD_VS_EARS_REQUIREMENTS_REVIEW.md](./docs/BDD_VS_EARS_REQUIREMENTS_REVIEW.md) for detailed coverage analysis.

## Documentation

### For Developers
- **[CLAUDE.md](./CLAUDE.md)** - Development guide for working with BDD tests
- **[README.md](./README.md)** - This file (quick start, structure, commands)

### For Test Engineers
- **[TEST_REVIEW_SUMMARY.md](./TEST_REVIEW_SUMMARY.md)** - Comprehensive test suite review and comparison with integration tests
- **[BDD_VS_INTEGRATION_COMPARISON.md](./BDD_VS_INTEGRATION_COMPARISON.md)** - Detailed coverage comparison
- **[INTEGRATION_PATTERNS_ADOPTED.md](./INTEGRATION_PATTERNS_ADOPTED.md)** - Patterns learned from integration tests

### For Server Team
- **[SERVER_ACTION_ITEMS.md](./SERVER_ACTION_ITEMS.md)** - Server bugs discovered by BDD tests (RBAC, license limits)
- **[ROLE_ASSIGNMENT_INVESTIGATION.md](./ROLE_ASSIGNMENT_INVESTIGATION.md)** - Role assignment mechanism investigation

### Historical
- **[BDD_FAILURES_ANALYSIS.md](./BDD_FAILURES_ANALYSIS.md)** - Analysis of test failures
- **[BDD_FIX_REFERENCE.md](./BDD_FIX_REFERENCE.md)** - Reference guide for fixing common issues

## External References

- [Design Document](../docs/superpowers/specs/2025-03-16-bdd-test-implementation-design.md)
- [Implementation Plan](../docs/superpowers/specs/2025-03-16-bdd-implementation-plan.md)
- [Godog Framework](https://github.com/cucumber/godog)
- [Gherkin Syntax](https://cucumber.io/docs/gherkin/)
- [Integration Tests](../integration/)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for BDD contribution guidelines.

## License

[Add license information]
