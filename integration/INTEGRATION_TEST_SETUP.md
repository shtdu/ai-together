# Integration Test Setup Guide

Quick setup guide for running integration tests from `integration/`.

## Quick Start

**Automated with Coverage (Recommended):**

```bash
# 1. Setup test database
createdb codetogether_test

# 2. Configure environment
cd integration
cp .env.test.example .env.test
# Edit .env.test with your database credentials

# 3. Run complete workflow with coverage
./integration-test.sh
```

This script:
- Starts server with coverage instrumentation
- Runs all integration tests
- Generates coverage report at `../server/coverage.html`
- Cleans up automatically

**Manual Workflow:**

```bash
# 1. Setup test database
createdb codetogether_test

# 2. Configure environment
cd integration
cp .env.test.example .env.test
# Edit .env.test with your database credentials

# 3. Start test server
./test-server.sh

# 4. Initialize tenant (one-time setup - script does this automatically)
curl -X POST http://localhost:8088/api/v1/setup/admin \
  -H "Content-Type: application/json" \
  -d '{"organization_name":"Test Org","admin_email":"admin@example.com","admin_name":"Admin","admin_password":"AdminPassword123!"}'

# 5. Run tests (in a new terminal)
cd integration
set -a && source .env.test && set +a
go test -v github.com/shtdu/integration

# 6. Stop server (Ctrl+C in terminal with test-server.sh)
# Coverage report will be generated automatically
```

## Environment Variables

Create `.env.test` in `integration/` (see `.env.test.example` for full template):

```bash
# Test server URL
TEST_SERVER_URL=http://localhost:8088

# Database (test database for both tests and server)
TEST_DATABASE_URL=postgres://user:pass@localhost:5432/codetogether_test?sslmode=disable
DATABASE_URL=postgres://user:pass@localhost:5432/codetogether_test?sslmode=disable

# Security
JWT_SECRET=default_secret_key_for_development
LICENSE_PUBLIC_KEY=+ftu10SQxIREoOQ3lHpJl0tr2pnPtuBVJCUbcRMDQKw=

# Server configuration
PORT=8088
DEBUG=true
```

**Generate keys:** `cd server && go run scripts/keygen/main.go`

**Note:** `.env.test` is auto-ignored by git.

## Test Server Options

| Mode | Command | Use Case |
|------|---------|----------|
| **Automated** | `./integration-test.sh` | Quick verification, CI/CD, coverage runs |
| **Manual** | `./test-server.sh` | Development, debugging, repeated test runs |
| **No Coverage** | `go run .` (from server/) | Quick testing without coverage overhead |

### Automated Testing
```bash
cd integration
./integration-test.sh
# Coverage report generated at: ../server/coverage.html
```

### Manual Testing
```bash
cd integration
./test-server.sh  # Terminal 1: Starts server on port 8088
# Server runs until stopped (Ctrl+C), generates coverage on exit
# Run tests in separate terminal (see below)
```

### Manual Start (No Coverage)
```bash
cd integration
if [ -f .env.test ]; then
  set -a && source .env.test && set +a
fi
cd ../server && go run .
```

## Running Tests

### All Tests
```bash
cd integration
set -a && source .env.test && set +a
go test -v github.com/shtdu/integration
```

### Specific Suite
```bash
go test -v -run "TestProvider"   # Provider tests (26)
go test -v -run "TestAuth"       # Auth tests (14)
go test -v -run "TestPermission" # Permission tests (8)
go test -v -run "TestLicense"    # License tests (40)
go test -v -run "TestUsage"      # Usage tests (10)
go test -v -run "TestUser"       # User tests (10)
go test -v -run "TestAnalytics"  # Analytics tests (34)
```

### Priority-Based
```bash
# P0 only (provider, auth, permission, license)
go test -v -run "(TestProvider|TestAuth|TestPermission|TestLicense)"

# P1 only (usage, user)
go test -v -run "(TestUsage|TestUser)"
```

### Coverage

**Automated (Recommended):**
```bash
cd integration
./integration-test.sh
# Coverage report: ../server/coverage.html
```

**Manual:**
```bash
# Terminal 1: Start server with coverage
cd integration && ./test-server.sh

# Terminal 2: Run tests
set -a && source .env.test && set +a
go test -v github.com/shtdu/integration

# Terminal 1: Stop server (Ctrl+C)
# Coverage report: ../server/coverage.html
```

**View coverage:**
```bash
open server/coverage.html
# Or check percentage:
cd server && go tool cover -func=coverage.out | grep total
```

## Database

### Create Test Database
```bash
createdb codetogether_test
# Or: psql postgres -c "CREATE DATABASE codetogether_test OWNER your_username;"
```

### Manual Cleanup
```bash
psql codetogether_test <<SQL
TRUNCATE TABLE
  request_log, team_usage_summary, team_members,
  teams, providers, users, tenants, licenses
CASCADE;
SQL
```

### Run Migrations
```bash
cd server
DATABASE_URL="postgres://user:pass@localhost:5432/codetogether_test?sslmode=disable" \
  go run .
```

## Tenant Initialization

**IMPORTANT:** Before running tests for the first time, initialize the tenant using the setup API.

### Check Setup Status
```bash
curl http://localhost:8088/api/v1/setup/status
# Response: {"setup_required":true}
```

### Initialize Tenant (One-Time Setup)
```bash
curl -X POST http://localhost:8088/api/v1/setup/admin \
  -H "Content-Type: application/json" \
  -d '{
    "organization_name": "Test Organization",
    "admin_email": "admin@example.com",
    "admin_name": "Test Admin",
    "admin_password": "AdminPassword123!"
  }'
```

This creates:
- Tenant: "Test Organization"
- Admin user: `admin@example.com` / `AdminPassword123!`
- Default team

**Note:** The admin user persists across test runs. Only test data is truncated between tests.

## Test Data

### Generate License Fixtures
```bash
cd integration/testdata/licenses
./generate.sh
```

### Generate Test Keys
```bash
cd server
go run scripts/keygen/main.go
# Copy public key to .env.test LICENSE_PUBLIC_KEY
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `database does not exist` | `createdb codetogether_test` |
| `connection refused` | `brew services start postgresql` |
| `relation does not exist` | Run migrations (start test server) |
| `permission denied` | `psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE codetogether_test TO your_username;"` |
| Missing licenses | `cd integration/testdata/licenses && ./generate.sh` |
| Missing public key | `cd server && go run scripts/keygen/main.go`, copy to `.env.test` |
| `setup_required: true` | Run tenant initialization (see above) |
| `401 Unauthorized` | Check admin credentials: `admin@example.com` / `AdminPassword123!` |
| `connection refused` on port 8088 | Start test server: `cd integration && ./test-server.sh` |
| Tests fail with "admin not found" | Initialize tenant first (one-time setup) |

## Debugging

```bash
# Verbose
go test -v -run "TestFailingTest"

# With race detector
go test -race -v -run "TestFailingTest"

# Save output
go test -v -run "TestFailingTest" 2>&1 | tee test.log

# Detailed trace
GOTRACEBACK=2 go test -v -run "TestFailingTest"
```

## CI/CD Example

```yaml
- name: Setup test database
  run: |
    psql -U postgres -c "CREATE DATABASE codetogether_test;"

- name: Run integration tests
  run: |
    cd integration
    TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/codetogether_test?sslmode=disable" \
      JWT_SECRET=test_secret \
      go test -v -cover
```

## Adding Tests

For comprehensive guidance on writing new tests (patterns, templates, conventions), see **[integration.design.md](./integration.design.md)**.

Quick start:
1. Add test to appropriate `*_test.go` file
2. Update `integration.design.md` with test specification
3. Run: `go test -v -run "TestNewTest"`

## Reference

- **[integration.design.md](./integration.design.md)** - Test specifications and fixtures
- **[.env.test.example](./.env.test.example)** - Environment template
- **[test-server.sh](./test-server.sh)** - Test server startup script with coverage
- **[integration-test.sh](./integration-test.sh)** - Automated test runner with coverage
- **[Go Testing](https://golang.org/pkg/testing/)** - Official Go testing documentation
