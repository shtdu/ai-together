# Testing Guide

Quick guide for running tests in the AI Together codebase.

## Overview

| Test Type | Location | Purpose | Dependencies |
|-----------|----------|---------|--------------|
| **Unit Tests** | `server/`, `member/`, `shared/` | Test individual components in isolation | No external services |
| **Integration Tests** | `integration/` | Test full API stack and workflows | PostgreSQL required |

---

## Quick Commands

```bash
# Run all unit tests
make test

# Run server tests only
make server-test

# Run member tests only
make member-test

# Run integration tests (requires PostgreSQL)
cd integration && ./integration-test.sh

# Run specific test
go test -v -run TestProviderCreate
```

---

## Unit Tests

### Running Unit Tests

```bash
# Run all unit tests (server + member)
make test

# Run server unit tests only
cd server && make test
# or: make server-test

# Run member unit tests only
cd member && make test
# or: make member-test

# Run specific package tests
cd server && go test ./handlers -v
cd member && go test ./services -v

# Run with coverage
cd server && make cover

# Run without cache
go test -count=1 ./...
```

### Server Unit Tests

Located in `server/` with `*_test.go` files.

**Test Categories:**
- `handlers/` - HTTP endpoint handlers
- `middleware/` - Authentication and authorization middleware
- `services/` - Business logic services
- `models/` - Data models and validation
- `config/` - Configuration loading

**Example:**
```bash
cd server
go test -v ./handlers -run TestAnalyticsHandler
```

### Member Unit Tests

Located in `member/` with `*_test.go` files.

**Test Categories:**
- `services/` - Provider, relay, and sync services
- `cmd/cli/` - CLI tool tests

**Example:**
```bash
cd member
go test -v ./services -run TestProviderService
```

### Shared Library Tests

Located in `shared/` with `*_test.go` files.

**Example:**
```bash
cd shared/hook-common/storage
go test -v
```

---

## Integration Tests

**See:** [`integration/INTEGRATION_TEST_SETUP.md`](integration/INTEGRATION_TEST_SETUP.md) for complete integration test documentation.

### Quick Start

```bash
# 1. Create test database (one-time setup)
createdb codetogether_test

# 2. Configure environment
cd integration
cp .env.test.example .env.test
# Edit .env.test with your database credentials

# 3. Run automated tests with coverage
./integration-test.sh
```

### Manual Testing

For development and debugging:

```bash
# Terminal 1: Start test server
cd integration && ./test-server.sh

# Terminal 2: Run tests
cd integration
set -a && source .env.test && set +a
go test -v github.com/shtdu/integration
```

### Integration Test Documentation

| Document | Purpose |
|----------|---------|
| **[integration/INTEGRATION_TEST_SETUP.md](integration/INTEGRATION_TEST_SETUP.md)** | Setup, environment, troubleshooting |
| **[integration/integration.design.md](integration/integration.design.md)** | Test specifications and patterns |
| **[integration/CLAUDE.md](integration/CLAUDE.md)** | Module documentation |

---

## Coverage Reports

### Server Coverage

```bash
cd server

# Terminal coverage
make cover

# HTML coverage (opens in browser)
make cover-html

# Coverage by function
make cover-func
```

### Integration Test Coverage

```bash
cd integration

# Run automated tests with coverage
./integration-test.sh

# View coverage report
open ../server/coverage.html
```

---

## Common Issues

| Problem | Solution |
|---------|----------|
| `go test` fails with "no matching files found" | Build frontend first: `make member-build` |
| Database connection refused | Ensure PostgreSQL is running |
| `relation does not exist` | Run migrations: Start test server |
| `setup_required: true` | See [integration/INTEGRATION_TEST_SETUP.md](integration/INTEGRATION_TEST_SETUP.md#troubleshooting) |

### Member Frontend Test Issues

If member tests fail with frontend errors:

```bash
# Rebuild member with frontend
make member-build
```

---

## Additional Resources

### Module Documentation
- **[server/CLAUDE.md](server/CLAUDE.md)** - Server module documentation
- **[member/CLAUDE.md](member/CLAUDE.md)** - Member client documentation
- **[integration/CLAUDE.md](integration/CLAUDE.md)** - Integration test documentation

### Integration Test Documentation
- **[integration/INTEGRATION_TEST_SETUP.md](integration/INTEGRATION_TEST_SETUP.md)** - Setup and execution guide
- **[integration/integration.design.md](integration/integration.design.md)** - Test specifications and patterns

### External Resources
- **[Testify Documentation](https://github.com/stretchr/testify)** - Assertion library
- **[Go Testing Guide](https://golang.org/pkg/testing/)** - Official Go testing guide

---

**Last Updated:** 2025-03-04

