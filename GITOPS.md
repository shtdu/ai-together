# Git Operations & Pull Request Checklist

This document defines the validation process for all changes before submitting a pull request.

## Quick Reference

```bash
# Run all validation checks
make test                           # Run all Go tests
cd manager && pnpm lint            # Lint TypeScript code

# Format code
(cd server && go fmt ./...)
(cd member && go fmt ./...)
(cd integration && go fmt ./...)

# Static analysis
(cd server && go vet ./...)
(cd member && go vet ./...)
(cd integration && go vet ./...)
```

## Pre-Commit Checklist

### Required Checks

| Check | Command | Expected Result |
|-------|---------|-----------------|
| **Go Formatting** | `go fmt ./...` (in each module) | No output = already formatted |
| **Go Vet** | `go vet ./...` (in each module) | No output = no issues |
| **Build** | `go build ./...` (in each module) | Binary compiled successfully |
| **Tests** | `make test` | All tests pass |
| **Manager Lint** | `cd manager && pnpm lint` | 0 errors, minimal warnings |
| **Integration Tests** | `make integration-test` | All tests pass |

### Module-Specific Commands

```bash
# Server
cd server
go fmt ./...
go vet ./...
go test ./...

# Member
cd member
go fmt ./...
go vet ./...
go test ./...

# Integration
cd integration
go fmt ./...
go vet ./...
go test ./...

# Manager UI
cd manager
pnpm lint
```

## PR Template Checklist

When creating a PR, complete the `.github/PULL_REQUEST_TEMPLATE.md` checklist:

- [ ] **Code follows style guidelines** - All `go fmt` and linting checks pass
- [ ] **Self-review performed** - Reviewed own code for logic errors and improvements
- [ ] **Comments in hard-to-understand areas** - Complex code has explanatory comments
- [ ] **Documentation updated** - Updated relevant CLAUDE.md files if behavior changed
- [ ] **No new warnings** - No new compiler or linter warnings introduced
- [ ] **Tested locally** - All tests pass locally before pushing
- [ ] **Semantic commit messages** - Commits follow `type(scope): description` format

## Semantic Commit Convention

Use these commit types:

| Type | When to Use | Example |
|------|-------------|---------|
| `feat` | New feature | `feat(server): add OAuth2 authentication` |
| `fix` | Bug fix | `fix(member): resolve proxy timeout issue` |
| `docs` | Documentation only | `docs: update API endpoint documentation` |
| `style` | Code style changes (formatting) | `style: format Go code with go fmt` |
| `refactor` | Code refactoring | `refactor(server): simplify provider routing logic` |
| `perf` | Performance improvement | `perf(database): add index to request_log table` |
| `test` | Adding or updating tests | `test(integration): add e2e test for user registration` |
| `chore` | Maintenance tasks | `chore: update dependencies to latest versions` |
| `ci` | CI/CD changes | `ci: add GitHub Actions workflow for linting` |
| `build` | Build system changes | `build: upgrade Go version to 1.24` |

## Module-Specific Validation

### Server (`server/`)

Additional checks for server changes:
- [ ] Database migrations added if schema changed
- [ ] API documentation updated if endpoints changed
- [ ] RBAC permissions verified if new routes added
- [ ] sqlc queries regenerated if SQL files modified

See `server/CLAUDE.md` for server-specific validation.

### Member (`member/`)

Additional checks for member changes:
- [ ] Frontend assets rebuilt if UI modified: `cd frontend && npm run build`
- [ ] Service registration updated in `main.go` if new services added
- [ ] Proxy endpoints tested if routing changed
- [ ] SQLite database migrations considered

See `member/CLAUDE.md` for member-specific validation.

### Manager (`manager/`)

Additional checks for manager changes:
- [ ] TypeScript types updated if API contracts changed
- [ ] API client regenerated if OpenAPI spec changed
- [ ] UI components tested in development mode
- [ ] Responsive design verified on mobile breakpoints

See `manager/CLAUDE.md` for manager-specific validation.

### Integration (`integration/`)

Additional checks for integration tests:
- [ ] Test database setup: `make integration-setup`
- [ ] Test fixtures updated if schema changed
- [ ] Environment variables configured in `.env.test`
- [ ] License generation scripts tested

See `integration/INTEGRATION_TEST_SETUP.md` for integration-specific validation.

## Pre-Push Verification

Before pushing to remote:

```bash
# Verify clean working directory
git status

# Run full test suite
make test && make integration-test

# Check commit history
git log --oneline -5

# Push to remote
git push
```

## Common Validation Issues

### Formatting Issues

```bash
# Auto-fix formatting
(cd server && go fmt ./...)
(cd member && go fmt ./...)
(cd integration && go fmt ./...)
```

### Import Issues

```bash
# Clean up imports
(cd server && go mod tidy)
(cd member && go mod tidy)
(cd integration && go mod tidy)
```

### Test Failures

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestHandler ./handlers/
```

## Continuous Integration (CI)

The project uses GitHub Actions for automated checks. See `.github/workflows/` for:
- **Linting** - Automated code style checks
- **Testing** - Automated test execution
- **Build** - Verification that all modules build successfully

All PRs must pass CI checks before merge.
