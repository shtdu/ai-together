# Integration Tests

This module contains black-box integration tests for the Code Together server API.

## Prerequisites

1. **PostgreSQL running** with test database
2. **Test server running** on port 8088 (via `./test-server.sh`)
3. **Initial server setup completed** (tenant created)

## Quick Start

### 1. Setup Test Database

```bash
# Create test database
createdb codetogether_test

# Or from psql:
psql -c "CREATE DATABASE codetogether_test;"
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.test.example .env.test

# Edit if needed (default should work)
```

### 3. Start Test Server

**For Automated Testing with Coverage:**

```bash
cd integration
./integration-test.sh
```

This script runs the complete workflow automatically (server + tests + coverage).

**For Manual Testing/Development:**

```bash
cd integration
./test-server.sh
```

This starts the server and keeps it running for manual testing in a separate terminal.

**Important:** Use `test-server.sh` when you want to:
- Run tests manually while the server is running
- Debug test failures interactively
- Run specific tests repeatedly without restarting the server
- Develop and iterate on tests

### 4. Initialize Tenant (First Time Only)

The server needs an initial admin user and tenant. Call the setup endpoint:

```bash
curl -X POST http://localhost:8088/api/v1/setup/admin \
  -H "Content-Type: application/json" \
  -d '{
    "organization_name": "Test Organization",
    "admin_email": "admin@example.com",
    "admin_name": "Test Admin",
    "admin_password": "TestPassword123!"
  }'
```

### 5. Run Integration Tests

**Automated with Coverage (Recommended):**

```bash
cd integration
./integration-test.sh
```

**Manual (if server already running):**

```bash
cd integration

# Load environment and run tests
set -a && source .env.test && set +a
go test -v github.com/code-together/integration

# Run specific test suite
go test -v -run "TestAuth" github.com/code-together/integration

# Run without cache
go test -count=1 -v github.com/code-together/integration
```

### 6. View Coverage Report

After running tests with `integration-test.sh` or stopping `test-server.sh`:

```bash
open server/coverage.html
```

The coverage report shows which server code was executed during integration tests.

## Test Structure

- **integration_test.go**: Main test suite with setup/teardown
- **auth_test.go**: 13 authentication tests
- **permission_test.go**: 8 permission tests
- **helpers.go**: Test helper functions and fixtures
- **cleanup.go**: Database setup and cleanup utilities
- **fixtures.go**: Fixture data loader

## Key Principles

1. **Black-box testing**: Only uses `shared/integration` API client
2. **API-only fixtures**: All test data created via API calls
3. **API-only assertions**: Verify results using API responses
4. **Database isolation**: Separate test database, cleanup between tests

## Troubleshooting

### "violates foreign key constraint users_tenant_id_fkey"

**Cause:** No tenant exists in the database.

**Solution:** Run the initial setup via `/api/v1/setup/admin` endpoint (see step 4 above).

### "Server did not become ready within timeout"

**Cause:** Test server not running on port 8088.

**Solution:** Start the server before running tests:
```bash
cd integration
./test-server.sh
```

### Database cleanup issues

If tests fail due to database state, manually clean:
```bash
psql codetogether_test -c "TRUNCATE TABLE request_log, team_usage_summary, team_members, teams, providers, users, licenses CASCADE;"
-- Note: Don't truncate 'tenants' - it's preserved for setup
```

## Current Test Status

- ✅ Test infrastructure complete
- ✅ Auth tests (13 tests)
- ✅ Permission tests (8 tests)
- ✅ Provider tests (30 tests)
- ✅ License tests (40 tests)
- ✅ Usage tests (18 tests)
- ✅ User tests (10 tests)

**Total Implemented:** 119 tests across 6 test suites

## Test Files

- `integration_test.go` - Main test suite with setup/teardown
- `auth_test.go` - Authentication tests (13)
- `permission_test.go` - Permission/RBAC tests (8)
- `provider_test.go` - Provider management tests (30)
- `license_test.go` - License management tests (40)
- `usage_test.go` - Usage tracking tests (18)
- `user_test.go` - User management tests (10)
- `helpers.go` - Test helper functions
- `cleanup.go` - Database setup and cleanup
- `fixtures.go` - Fixture data loader
