# Shared Code

Common Go packages and utilities shared across AI Together modules.

**Module Path:** `github.com/code-together/shared`

This directory contains independent Go packages that are imported by server, member, and integration modules.

---

## Directory Overview

| Directory | Purpose | Used By |
|-----------|---------|---------|
| **`integration/`** | Member API client (OpenAPI-generated) | Member, Integration Tests |
| **`integration_manager/`** | Manager API client (OpenAPI-generated) | Integration Tests |
| **`provider/`** | AI provider routing & streaming parsers | Server, Member |
| **`streaming/`** | SSE client for Claude streaming responses | Member |
| **`events/`** | Claude Code hook event type definitions | Hook Tools |
| **`hook-common/`** | Shared code for hook event tools | Hook Tools |
| **`hook-utils/`** | Git/CWD utilities for hooks | Hook Tools |
| **`tools/`** | Standalone CLI tools for hook events | Development |

---

## API Clients

### `integration/` — Member API Client

**Generated from:** `docs/client_api/server_api.yaml` (not yet created)

Type-safe client for the AI Together server API. Used by member client to communicate with the central server.

**Features:**
- Authentication (login, register, refresh, verify)
- Provider management (CRUD for managers)
- Read-only team information
- Usage tracking and analytics
- User profile management

**Usage:**
```go
import "github.com/code-together/shared/integration"

client, _ := integration.NewClient("http://localhost:9080")
resp, _ := client.PostApiV1AuthLoginWithResponse(ctx, req)
```

**Regeneration:**
```bash
cd shared/integration && go generate
# Or from root: make shared-api-gen
```

---

### `integration_manager/` — Manager API Client

**Generated from:** `docs/client_api/server_api_manager.yaml` (not yet created)

Type-safe client for the AI Together Manager admin API. Used by integration tests for Manager API coverage.

**Usage:**
```go
import integration_manager "github.com/code-together/integration_manager"

client, _ := integration_manager.NewClient("http://localhost:9080")
resp, _ := client.PostApiV1ManagerUsersWithResponse(ctx, req)
```

**See also:** `integration/mgr_*.go` test files

---

## Provider & Streaming

### `provider/` — AI Provider Routing

Handles configuration and request routing to AI providers (Claude, Codex, OpenCode).

**Key Types:**
- `ProviderConfig` - Provider settings (name, API endpoint, auth, models)
- `ModelMapping` - Maps model aliases to provider-specific models
- `StreamToken` - Unified token type for streaming responses

**Files:**
- `config.go` - Provider configuration types
- `model_helpers.go` - Model mapping utilities
- `builder.go` - Request builders for different providers

---

### `streaming/` — SSE Streaming Client

Server-Sent Events (SSE) client for consuming streaming responses from AI providers.

**Key Types:**
- `SSEClient` - HTTP client for SSE streams
- `TokenParser` - Parse tokens from Claude/Codex responses

**Used by:** Member client's provider relay service

---

## Hook Events Tooling

### `events/` — Hook Event Types

Go struct definitions for Claude Code hook events.

**Event Types:**
- `Stop` - Claude Code finishes responding
- `PreToolUse` / `PostToolUse` - Before/after tool calls
- `SessionEnd` - Session ends
- `SubagentStop` - Subagent tasks complete
- `UserPromptSubmit` - User submits prompt

**Generated from:** `hook-schema-analyzer` tool

---

### `hook-common/` — Hook Event Storage

Shared database operations for storing and retrieving hook events using SQLite with WAL mode.

**Components:**
- `storage/` - SQLite database operations (read, write, types)
- `config/` - Configuration and path resolution

**Features:**
- Concurrent read/write access (WAL mode)
- Tool isolation (claude-code, codex, opencode)
- Session and event management

**Used by:** `hook-collector`, `hook-browser`

**See also:** `shared/hook-common/README.md`

---

### `hook-utils/` — Hook Utilities

Utility functions for hook event tools.

**Utilities:**
- `git.go` - Git repository operations
- `cwd.go` - Current working directory detection

---

### `tools/` — Hook Event CLI Tools

Standalone CLI tools for working with Claude Code hook events.

| Tool | Description |
|------|-------------|
| `hook-schema-analyzer` | Generate Go structs from hook logs |
| `hook-events-tool` | Split hook events into organized JSON |
| `hook-collector` | Collect and store events in SQLite |
| `hook-browser` | TUI browser for exploring events |

**See also:** `shared/tools/README.md`

---

## Go Workspace

This directory uses Go 1.25+ workspace feature (`go.work`) to coordinate development across modules:

```go
use (
    .
    ./tools/hook-browser
    ./tools/hook-collector
)
```

---

## Dependencies

Shared packages have minimal external dependencies:

- `oapi-codegen/runtime` - OpenAPI client runtime
- `modernc.org/sqlite` - Pure Go SQLite (hook-common)

---

## Related Documentation

- **`../server/CLAUDE.md`** - Server module documentation
- **`../member/CLAUDE.md`** - Member client documentation
- **`../integration/CLAUDE.md`** - Integration test documentation
- **`../docs/design/constitution.md`** - Architecture principles

---

**Last Updated:** 2025-03-17
