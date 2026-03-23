# BDD Test Suite

Behavioral-Driven Development (BDD) tests for the AI Together platform using the [godog](https://github.com/cucumber/godog) framework.

## Overview

This BDD test suite converts existing integration tests into Gherkin scenarios that serve as executable specifications. The tests are organized by domain and align with the EARS requirements structure.

## Prerequisites

- Go 1.25+
- PostgreSQL 14+ (for test database)
- Test server running or auto-start capability

## Quick Start

### Run All BDD Tests

```bash
# The script will automatically build and start the test server if needed
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

The BDD tests require a running test server. The script manages the server independently.

### Automatic Server Management (Default)

```bash
./bdd-test.sh
```

This will:
- Build the test server binary if needed (from `../server`)
- Check if test server is already running
- Start the server automatically if not running
- Wait for server to be ready
- Execute BDD tests
- Stop the server automatically if we started it
- Handle Ctrl+C gracefully to cleanup

### Manual Server Management

If you prefer to manage the server manually:

```bash
# Terminal 1: Start server manually
cd ../server
PORT=8088 ./bin/codetogether_test.cover -test.run TestCoverageServer

# Terminal 2: Run tests with --no-server flag
cd ../bdd
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
| `TEST_SERVER_PORT` | Test server port | `8088` |
| `TEST_DATABASE_URL` | Test database URL | `postgres://test:test@localhost:5432/codetogether_test?sslmode=disable` |
| `TEST_SERVER_LOG_PATH` | Test server log file | `/tmp/bdd-test-server.log` |

## Troubleshooting

### Test Server Not Running

```bash
# Check server status
curl http://localhost:8088/health

# View BDD test server logs
cat /tmp/bdd-test-server.log

# The script will automatically start the server, but if you want to do it manually:
cd ../server
PORT=8088 ./bin/codetogether_test.cover -test.run TestCoverageServer
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

## Test Coverage

| Domain | Scenarios |
|--------|-----------|
| Authentication & Users | 79 |
| Provider Management | 24 |
| Usage Analytics | 63 |
| Dashboard & Permissions | 28 |
| System Behaviors | 6 |
| Infrastructure | 2 |
| **Total** | **212** |

**Current Status:** 198 passing (93%), 13 failing (backend gaps), 24 skipped (@wip)

## Documentation

- **[CLAUDE.md](./CLAUDE.md)** - Development guide for extending the test suite
- **[Integration Tests](../integration/)** - Related integration test module

## External References

- [Godog Framework](https://github.com/cucumber/godog)
- [Gherkin Syntax](https://cucumber.io/docs/gherkin/)
