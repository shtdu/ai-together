# CLAUDE.md - Member Client

Wails 3 desktop app running HTTP proxy on port 18100. Routes AI tool requests (Claude Code, Codex, OpenCode) to configured providers with automatic failover.

**Module:** `codeswitch` | **Path:** `member/`

## Build First

**Always verify compilation after changes:**
```bash
make member-build          # Build production binary (41MB)
make member-test           # Run tests
```

**Testing Rule:** All tests MUST use temporary directories, NEVER `~/.code-together` live data.

## Quick Start

```bash
# First time: install frontend deps
cd member/frontend && npm install && npm run build

# Development (hot reload)
make member-dev            # From repo root
# or: wails3 task dev      # From member/ directory
```

## Project Structure

```
member/
├── main.go              # Entry point, service registration (GUI)
├── cmd/cli/             # CLI tool (headless mode) - see cmd/cli/CLAUDE.md
├── services/            # Business logic (exposed to frontend)
│   ├── providerservice.go    # Provider CRUD
│   ├── providerrelay.go      # HTTP proxy (port 18100)
│   ├── hookservice.go        # Hook event collection
│   └── *_test.go             # Unit/integration tests
├── internal/
│   ├── api/             # Generated OpenAPI client
│   ├── db/              # App database (SQLite)
│   └── hookdb/          # Hook events database (SQLite)
└── frontend/            # Vue 3 UI (embedded via //go:embed)
```

## Key Services

| Service | Purpose |
|---------|---------|
| ProviderService | Provider CRUD operations |
| ProviderRelay | HTTP proxy with failover |
| HookService | Hook event collection API |
| AuthService | JWT authentication |
| ConfigSyncService | Sync providers with server |
| UsageSyncService | Report usage statistics |

## HTTP Proxy (port 18100)

```
POST /v1/messages           → Claude provider (Anthropic API compatible)
POST /responses             → Codex provider
POST /chat/completions      → OpenCode provider
POST /collect/{tool_name}   → Hook event collection
```

Automatic failover by provider priority.

## Development

```bash
make member-build          # Build production binary
make member-dev            # Hot reload development
make member-test           # Run tests

# Frontend (in member/frontend/)
npm run dev                # Vite dev server (port 9245)
npm run build              # Production build
```

## Testing

```bash
go test ./... -v                    # All tests
go test -run TestHookService ./...  # Specific test suite
```

## Common Issues

| Problem | Solution |
|---------|----------|
| Port 18100 in use | `lsof -ti:18100 \| xargs kill -9` |
| Frontend not updating | `cd frontend && npm run build` |
| Build fails | Ensure frontend built first |
| Database locked | Close all app instances |
| Import errors | Run `go mod tidy` |

## Key Files

- `main.go` - App entry point, service registration
- `services/hookservice.go` - Hook event collection
- `services/providerrelay.go` - HTTP proxy with failover
- `services/providerservice.go` - Provider management

## Related Docs

- **[CLI Documentation](cmd/cli/CLAUDE.md)** - CLI tool for headless environments
- **[Server Integration Guide](docs/server_integration.md)** - API client patterns, OpenAPI
