# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Design Constitution

**IMPORTANT:** Before making any architectural or technical decisions, read **`docs/design/constitution.md`**.

The constitution captures the foundational principles that govern all technical decisions:

| Principle | Priority | Description |
|-----------|----------|-------------|
| Privacy-First | ⭐⭐⭐ | Never access/store prompt/response data; metadata only |
| Multi-Tenant Isolation | ⭐⭐⭐ | Strict data segregation at database level |
| Stateless Authentication | ⭐⭐ | JWT tokens, no server-side sessions |
| API-First Design | ⭐ | RESTful with OpenAPI spec |
| Graceful Degradation | ⭐ | Member clients function offline |

**Technology Stack** (fixed - changes require approval):
- Backend: Go 1.24+, Gin, PostgreSQL 14+, sqlc, JWT, Casbin, ed25519
- Frontend: React 18 + TypeScript (Manager), Wails v3 + Vue 3 (Member)

**Architectural Patterns:**
- Three-tier: Handlers → Services → Repositories
- Repository pattern with interfaces
- Middleware chain: Logging → CORS → Recovery → Auth → RBAC → License → Handler

**Violations** to core principles require explicit justification and team approval.

**See also:** [`docs/design/architecture.md`](docs/design/architecture.md) for detailed system design, data flows, security architecture, and deployment patterns.

## Product Requirements (Single Source of Truth)

**IMPORTANT:** All product requirements, user stories, and feature specifications are maintained in **`vibe_doc/`**.

The `vibe_doc/` directory contains MECE-structured (Mutually Exclusive, Collectively Exhaustive) product documentation organized by domain. This is the **single source of truth** for:
- What features the system should have
- User personas and use cases
- Functional requirements and acceptance criteria
- Business rules and constraints
- Privacy, security, and compliance requirements

### Product Domains (`vibe_doc/`)

| ID | Domain | Folder | Key Topics |
|----|--------|--------|------------|
| 00 | Product Overview | `vibe_doc/00_overview/` | Value proposition, user personas, capabilities overview |
| 01 | Identity & Access | `vibe_doc/01_identity_and_access/` | User accounts, roles/permissions, multi-tenancy, licensing |
| 02 | Provider Management | `vibe_doc/02_provider_management/` | Provider configuration, request routing, model mapping |
| 03 | Configuration Sync | `vibe_doc/03_configuration_sync/` | Distribution to members, teams, offline mode |
| 04 | Usage Insights | `vibe_doc/04_usage_insights/` | Data collection, personal/team analytics, cost tracking, retention |
| 05 | User Interfaces | `vibe_doc/05_user_interfaces/` | Member desktop app, manager dashboard, accessibility |
| 06 | System Behaviors | `vibe_doc/06_system_behaviors/` | Privacy, security, performance, compliance |

**See:** [`vibe_doc/README.md`](vibe_doc/README.md) for complete documentation map.

### When to Reference `vibe_doc/`

- **Before implementing features:** Check `vibe_doc/` for functional requirements and acceptance criteria
- **When defining user stories:** Reference user persona definitions in `00_overview/`
- **For permission checks:** See `01_identity_and_access/02_roles_permissions/`
- **For license-dependent features:** See `01_identity_and_access/04_licensing/`
- **For data retention policies:** See `04_usage_insights/05_data_retention/`

### Documentation Structure

This project uses a three-tier documentation structure:

| Tier | Location | Purpose | Audience |
|------|----------|---------|----------|
| **Product Requirements** | `vibe_doc/` | WHAT to build (features, user stories, acceptance criteria) | PMs, stakeholders, engineers |
| **Technical Design** | `docs/design/` | HOW to architect (principles, patterns, system design) | Engineers, architects |
| **Implementation Guides** | `*/CLAUDE.md` | HOW to code (module-specific patterns, workflows) | Engineers |

**Workflow:** When implementing features, start with product requirements (`vibe_doc/`), reference technical design (`docs/design/`) for architectural guidance, and consult module guides (`*/CLAUDE.md`) for implementation specifics.

## Project Overview

AI Together is a team collaboration platform for managing AI coding tools (Claude Code, Codex, and OpenCode). It provides centralized provider management, team-level settings, usage statistics, and automatic configuration distribution across team members.

The system consists of four main components:
1. **Server** (`server/`) - Backend API for centralized management, authentication, and usage tracking
2. **Member Client** (`member/`) - Individual user desktop application with auto-configuration
3. **Manager Client** (`manager/`) - Administrative web interface for team analytics and user management
4. **Integration Tests** (`integration/`) - Independent integration test module with minimal dependencies

## Architecture

### Monorepo Structure

This is a multi-module monorepo with the following structure:

- **Member Module** (`codeswitch`)
  - Contains the member client code in `member/` directory
  - Independent Go module with its own `go.mod`
  - Desktop application built with Wails 3

- **Server Module** (`switch-server`)
  - Located in `server/` directory
  - Independent Go module with its own `go.mod`
  - Backend API server built with Gin framework

- **Manager Module** (`code-together-ui`)
  - Located in `manager/` directory
  - React 18 + TypeScript + Vite web application
  - Administrative interface for team analytics and management
  - Communicates with server via REST API

- **Integration Module** (`integration`)
  - Located in `integration/` directory at project root
  - Independent Go test module with its own `go.mod`
  - Minimal dependencies - only imports `shared/integration` client
  - Uses external PostgreSQL database for testing

**Key Point:** The Go modules (server, member, integration) share the root `Makefile` for unified build orchestration. The manager module uses npm/package.json scripts but is integrated into the Makefile for convenience.

### System Overview

The member client runs an HTTP proxy on port 18100 that transparently routes AI tool requests to configured providers, enabling seamless provider switching without restarting Claude Code or Codex. The member client can optionally connect to a central server for team collaboration, usage tracking, and centralized provider management.

**Key proxy endpoints:**
- `/v1/messages` - Routes to configured Claude provider
- `/responses` - Routes to Codex provider

**Server Integration:**
- Member clients sync provider configurations with the server
- Usage statistics are aggregated from member clients
- Team-level settings are distributed to all members

## Build Requirements

**Required:**
- Go 1.24+
- Node.js 18+ and pnpm (for member client frontend and manager UI)
- Wails 3 CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`

**Optional:**
- PostgreSQL 14+ for server development
- Air for live reload: `go install github.com/cosmtrek/air@latest`
- mingw-w64 for Windows cross-compilation on macOS: `brew install mingw-w64`

## Root Makefile Commands

The root Makefile provides orchestration commands for working with both modules. Run `make help` to see all available commands.

### Quick Start
```bash
make build                         # Build server, member, and manager
make test                          # Run all tests
make server                        # Show server-specific commands
make member                        # Show member-specific commands
make manager                       # Show manager-specific commands
```

### Module Commands
Each module has its own help menu with available commands:

```bash
make server                        # Show all server commands
make member                        # Show all member commands
make manager                       # Show all manager commands
```

### Common Tasks
```bash
make clean                         # Remove all build artifacts
make fmt                           # Format all Go code
make vet                           # Vet all Go code
make tidy                          # Run go mod tidy everywhere
make deps                          # Download all Go dependencies
```

#### Server-specific commands (run `make server` to see all):
```bash
make server-build                  # Build server binary
make server-dev                    # Run in dev mode with live reload (air)
make server-test                   # Run server tests
make server-itest                  # Run integration tests (requires PostgreSQL)
```

#### Integration-specific commands (run `make integration` to see all):
```bash
make integration-test              # Run integration tests (requires PostgreSQL)
make integration-setup             # Setup test database and environment
make integration-clean             # Clean test database
```

#### Member-specific commands (run `make member` to see all):
```bash
make member-build                  # Build member app
make member-dev                    # Run in dev mode (hot reload)
make member-test                   # Run member tests
```

#### Manager-specific commands (run `make manager` to see all):
```bash
make manager-build                 # Build manager UI
make manager-dev                   # Run dev server with hot reload
make manager-lint                  # Lint TypeScript/React code
```

## Module-Specific Documentation

For detailed information about each module, refer to their respective CLAUDE.md files:

### Server (`server/CLAUDE.md`)
- Backend API architecture (Gin framework, PostgreSQL)
- Multi-tenant data isolation
- API endpoints reference
- Database schema and migrations
- Three-tier architecture (Handlers → Services → Repository)
- Server-specific development patterns

### Integration Tests (`integration/INTEGRATION_TEST_SETUP.md`)
- Independent test module with minimal dependencies
- External PostgreSQL database setup
- Integration test execution and debugging
- Test fixtures and license generation
- Environment configuration

### Member Client (`member/CLAUDE.md`)
- Wails 3 desktop application architecture
- Service layer pattern and registration
- HTTP proxy implementation (port 18100)
- Vue 3 + TypeScript frontend
- AI tool configuration (Claude Code, Codex, OpenCode)
- Server integration services
- Frontend development workflow
- Cross-platform builds

### Manager Client (`manager/CLAUDE.md`)
- React 18 + TypeScript web application
- Material UI (MUI) component library
- React Router v7 for navigation
- TanStack React Query for server state
- Vite for build tooling and dev server
- API client architecture with Axios
- Authentication and RBAC
- Data visualization with Recharts
- Responsive layout patterns

## Quick Start

### Initial Setup

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

```bash
# Terminal 1: Server (with live reload via air)
make server-dev                    # Run server in dev mode with hot reload

# Terminal 2: Member app (with live reload via wails3)
make member-dev                    # Run member app in dev mode with hot reload

# Terminal 3: Manager UI (with live reload via Vite)
make manager-dev                   # Run manager UI in dev mode (port 3000)
                                    # Proxies /api and /auth to server port 9080
```

See module-specific CLAUDE.md files for detailed development workflows and patterns.

## Git Operations & PR Validation

**Before pushing changes**, follow the validation checklist in **[`GITOPS.md`](GITOPS.md)**.

The GITOPS.md guide covers:
- Pre-commit validation checks (formatting, vetting, testing)
- PR template checklist
- Semantic commit conventions
- Module-specific validation requirements
- CI/CD requirements

**Quick validation:**
```bash
make test              # Run all tests
cd manager && pnpm lint  # Lint TypeScript
```

## Common Issues

- **".app cannot be opened"**: See `member/CLAUDE.md` for Wails build asset fixes
- **"Migration errors"**: See `server/CLAUDE.md` for PostgreSQL troubleshooting
- **"Frontend not updating"**: See `member/CLAUDE.md` for frontend build instructions
- **"Port conflicts"**: Server uses port 9080 (internal), external access via 8080, member proxy uses port 18100, manager UI uses port 3000

## Related Documentation

### Module Documentation
- `server/CLAUDE.md` - Server module documentation
- `member/CLAUDE.md` - Member client documentation
- `manager/CLAUDE.md` - Manager UI documentation
- `server/README.md` - Server RBAC permissions and setup

### Product Requirements
- `vibe_doc/README.md` - **Product documentation index (single source of truth)**
- `vibe_doc/00_overview/` - Product overview, user personas, value proposition
- `vibe_doc/01_identity_and_access/` - User accounts, roles, multi-tenancy, licensing
- `vibe_doc/02_provider_management/` - Provider configuration and routing
- `vibe_doc/03_configuration_sync/` - Configuration distribution and teams
- `vibe_doc/04_usage_insights/` - Analytics and cost tracking
- `vibe_doc/05_user_interfaces/` - Member app and manager dashboard specs
- `vibe_doc/06_system_behaviors/` - Privacy, security, performance, compliance

### Technical Design (Architecture & Constitution)

**Start Here:**
- `docs/design/constitution.md` - **Design principles, technology stack, patterns, anti-patterns (READ FIRST)**
- `docs/design/architecture.md` - **System architecture, data flows, security, deployment, scalability**

**Implementation Guides:**
- `docs/phase_*/` - Implementation phase documentation
- `docs/epic_*/` - Epic feature documentation
- `integration/integration.design.md` - Integration test specifications
- `integration/INTEGRATION_TEST_SETUP.md` - Integration test setup guide

**Project Overview:**
- `README.md` - Project overview (Chinese)

### External Resources
- Wails 3 docs: https://v3.wails.io
- Gin docs: https://gin-gonic.com/docs/
- Casbin docs: https://casbin.org/docs/overview
