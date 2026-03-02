# CLAUDE.md - Integration Test Module

Guidance for Claude Code when working with the integration test module.

## Module Overview

**Location:** `integration/` (project root)

**Purpose:** Independent Go test module for server API integration testing with black-box coverage.

**Status:** 141/141 tests passing (100%)

### Architecture

- **Separate `go.mod`** (module: `github.com/code-together/integration`)
- **Only imports** `shared/integration` client (NOT server codebase)
- **Black-box tests** - Treats server as external dependency
- **Regenerate client:** `cd shared/integration && go generate`

---

## Quick Start

### Automated Testing (Recommended)

```bash
cd integration
./integration-test.sh
```

### Manual Testing (Development)

```bash
# Terminal 1: Start server
cd integration
./test-server.sh

# Terminal 2: Run tests
set -a && source .env.test && set +a
go test -v github.com/code-together/integration
```

---

## Key Documentation

| Document | Purpose |
|----------|---------|
| **[integration.design.md](./integration.design.md)** | Test specifications, patterns, adding new tests |
| **[INTEGRATION_TEST_SETUP.md](./INTEGRATION_TEST_SETUP.md)** | Setup, environment, troubleshooting |
| **[issue.md](./issue.md)** | Test status, known issues |

---

## Test Files

| Category | File | Tests |
|----------|------|-------|
| Provider | `provider_test.go` | 26 |
| Auth | `auth_test.go` | 14 |
| Permission | `permission_test.go` | 8 |
| License | `license_test.go` | 40 |
| Usage | `usage_test.go` | 10 |
| User | `user_test.go` | 10 |
| Analytics | `analytics_*.go` (5 files) | 34 |

**Total:** 142 tests across 12 files (integration_test.go contains suite setup)

---

## Common Commands

```bash
# Run all tests
go test -v github.com/code-together/integration

# Run specific category
go test -v -run "TestProvider" github.com/code-together/integration
go test -v -run "TestAuth" github.com/code-together/integration
go test -v -run "TestLicense" github.com/code-together/integration

# Run without cache
go test -count=1 -v github.com/code-together/integration

# View server coverage (after automated test)
open server/coverage.html
```

---

## Key Helper Functions

**From `helpers.go`:**
- `generateUniqueProviderName(baseName)` - Unique names to avoid conflicts
- `LoadLicenseFixture(fixtureName)` - Load license PEM files
- `intPointer(v)`, `int64Pointer(v)` - Convert to pointers

**From `integration_test.go`:**
- `registerAndLoginUserFromFixture(fixtureName)` - Create/test users
- `activateLicenseFixture(fixtureName)` - Activate license via API
- `createProviderFixtureFromFixture(fixtureName)` - Create provider
- `bootstrapStandardFixture()` - Setup standard test environment

---

## Adding New Tests

1. Read **[integration.design.md](./integration.design.md)** for patterns
2. Add test to appropriate `*_test.go` file
3. Run: `go test -v -run "TestNewFeature"`

### Test Template

```go
func (s *IntegrationTestSuite) TestNewFeature() {
    ctx := context.Background()

    // Arrange: Create fixtures via API
    uniqueName := generateUniqueProviderName("new-feature")
    req := integration.PostApiV1ProvidersJSONRequestBody{
        Name:   uniqueName,
        ApiKey: "test-key",
    }
    resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
    require.NoError(s.T(), err)

    // Act: Call API being tested
    // ...

    // Assert: Verify via API response
    assert.Equal(s.T(), 200, resp.StatusCode())

    // Cleanup: Delete resources
    defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, id)
}
```

---

## Architecture & Principles

### Black-Box Testing

1. **API-Only Communication:** All interactions via HTTP API
2. **No Database Access:** Never import server packages or query DB directly
3. **Shared Client:** Use `shared/integration` client for all API calls
4. **External Fixtures:** Create test data via API, not DB inserts

### Test Isolation

1. **Unique Names:** `generateUniqueProviderName()` prevents conflicts
2. **API Cleanup:** `defer` statements ensure resource cleanup
3. **Conditional Bootstrap:** `skipBootstrap` for clean state
4. **Separate Database:** Uses `codetogether_test` database

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Database does not exist | `createdb codetogether_test` |
| Connection refused | `brew services start postgresql` |
| Relation does not exist | Run migrations (start test server) |
| Missing licenses | `cd integration/testdata/licenses && ./generate.sh` |
| `setup_required: true` | Run tenant initialization (one-time setup) |

**Full troubleshooting:** [INTEGRATION_TEST_SETUP.md](./INTEGRATION_TEST_SETUP.md#troubleshooting)

---

## References

### Module Documentation
- **[integration.design.md](./integration.design.md)** - Test specifications and patterns
- **[INTEGRATION_TEST_SETUP.md](./INTEGRATION_TEST_SETUP.md)** - Setup and execution guide
- **[issue.md](./issue.md)** - Test status and known issues

### External Documentation
- **[../shared/integration/](../shared/integration/)** - API client implementation
- **[../server/CLAUDE.md](../server/CLAUDE.md)** - Server module documentation
- **[../CLAUDE.md](../CLAUDE.md)** - Project-level overview

---

**Last Updated:** 2025-02-11
**Test Status:** 141/141 passing (100%)
