# Contributing to Code Together

Thank you for your interest in contributing to Code Together! We appreciate your help in making this project better.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Commit Message Conventions](#commit-message-conventions)

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites

- **Go 1.24+** - Download from [go.dev](https://go.dev/dl/)
- **Node.js 18+ and pnpm** - For member client frontend and manager UI
- **Wails 3 CLI** - `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- **PostgreSQL 14+** - For server development
- **Make** - For build orchestration

### Initial Setup

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/ai-together.git
   cd ai-together
   ```

3. Install dependencies:
   ```bash
   # Server setup
   cd server
   cp .env.example .env
   # Edit .env with your PostgreSQL credentials

   # Member client setup
   cd ../member/frontend
   npm install
   npm run build  # Build frontend assets once
   cd ../..

   # Manager UI setup
   cd ../manager
   pnpm install
   cd ..
   ```

### Development Workflow

We use a systematic development workflow to ensure code quality:

```bash
# Terminal 1: Server (with live reload via air)
make server-dev

# Terminal 2: Member app (with live reload via wails3)
make member-dev

# Terminal 3: Manager UI (with live reload via Vite)
make manager-dev
```

## Code Style Guidelines

### Go Code

- Follow standard Go conventions described in [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` to format code: `make fmt`
- Run `go vet` to check for issues: `make vet`
- Keep functions focused and small
- Use descriptive names for variables and functions
- Add comments for exported functions and complex logic

### JavaScript/TypeScript Code

- Use ESLint for linting: `make manager-lint`
- Follow the existing code style in the project
- Use meaningful variable and function names
- Add JSDoc comments for complex functions

### Documentation

- Update README.md if you change user-facing behavior
- Add comments to code explaining complex logic
- Keep module-specific CLAUDE.md files up to date

## Testing Requirements

### Running Tests

```bash
# Run all tests
make test

# Run server tests
make server-test

# Run member tests
make member-test

# Run integration tests (requires PostgreSQL)
make integration-test
```

### Writing Tests

- Write unit tests for new functions and methods
- Maintain test coverage above 80%
- Use table-driven tests for multiple test cases
- Mock external dependencies
- Test error conditions, not just success paths

### Test Naming

- Use descriptive test names: `TestCalculatePrice_WithDiscount_ReturnsCorrectPrice`
- Group related tests using subtests

## Pull Request Process

### Before Submitting

1. **Update documentation** - Ensure README and CLAUDE.md files are updated
2. **Add tests** - All new features must have tests
3. **Run tests** - Ensure all tests pass: `make test`
4. **Format code** - Run `make fmt` and `make vet`
5. **Build** - Verify the build succeeds: `make build`

### Submitting a Pull Request

1. Create a descriptive branch name:
   ```bash
   git checkout -b feature/add-provider-management
   ```

2. Commit your changes with clear messages (see [Commit Message Conventions](#commit-message-conventions))

3. Push to your fork:
   ```bash
   git push origin feature/add-provider-management
   ```

4. Create a pull request from your fork to the main repository

5. Fill out the pull request template with:
   - Description of changes
   - Type of change (bug fix, feature, docs, etc.)
   - Testing performed
   - Related issue numbers

### Pull Request Review Process

- All PRs must be reviewed by at least one maintainer
- Address review comments promptly
- Keep the discussion focused and constructive
- Squash commits if requested by reviewers

### After Merge

- Delete your branch after merge
- Update your fork regularly to stay in sync

## Commit Message Conventions

We follow semantic commit messages to make the project history easier to understand:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, etc.)
- **refactor**: Code change that neither fixes a bug nor adds a feature
- **test**: Adding or updating tests
- **chore**: Changes to build process or auxiliary tools

### Examples

```
feat(member): add support for custom provider endpoints

Implement support for custom API endpoints in provider configuration.
This allows users to specify alternative endpoints for providers
that use non-standard URLs.

Closes #123
```

```
fix(server): resolve race condition in usage tracking

The usage tracking service had a race condition where concurrent
requests could cause duplicate records. Added proper locking
mechanism to prevent this issue.

Fixes #456
```

```
docs(readme): update installation instructions

Clarified the installation process and added troubleshooting
section for common setup issues.
```

### Best Practices

- Use the imperative mood ("add" not "added" or "adds")
- Limit the first line to 72 characters or less
- Reference issues in the footer
- Explain what and why, not how

## Module-Specific Guidelines

### Server Module

See [server/CLAUDE.md](server/CLAUDE.md) for:
- Backend API architecture
- Database schema conventions
- Three-tier architecture patterns

### Member Client

See [member/CLAUDE.md](member/CLAUDE.md) for:
- Wails 3 desktop application patterns
- Service layer pattern
- HTTP proxy implementation

### Manager UI

See [manager/CLAUDE.md](manager/CLAUDE.md) for:
- React 18 + TypeScript patterns
- Material UI component usage
- API client architecture

### Integration Tests

See [integration/INTEGRATION_TEST_SETUP.md](integration/INTEGRATION_TEST_SETUP.md) for:
- Integration test patterns
- Black-box testing principles
- Test data fixtures

## Questions?

- Check existing [GitHub Issues](https://github.com/shtdu/ai-together/issues)
- Start a [Discussion](https://github.com/shtdu/ai-together/discussions)
- Read the [Documentation](README.md)

Thank you for contributing to Code Together!
