# CLAUDE.md - Member CLI Tool

Headless CLI mode for containers, servers, and CI/CD pipelines.

**Module:** `codeswitch` | **Path:** `member/cmd/cli/`

## Overview

The CLI tool provides the same core functionality as the GUI member client without requiring a display. It loads providers directly from the server in-memory and enables/disables tool-specific proxies.

## Build & Run

```bash
# Build CLI tool
make cli-build
# Binary: bin/member-cli (26MB)

# Run with existing .code-together settings
./bin/member-cli -tool-name claude

# Run with explicit credentials
./bin/member-cli -tool-name claude -auth <token> -server http://server:9080

# View help
./bin/member-cli -h
```

## CLI Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `-tool-name` | Yes | Tool: `claude`, `codex`, or `opencode` |
| `-auth` | Conditional | JWT token (must pair with `-server`) |
| `-server` | Conditional | Server URL (must pair with `-auth`) |
| `-verbose` | No | Enable debug logging |

**Rule**: Both `-auth` AND `-server` must be provided together, or neither (uses settings).

## Configuration

### Settings Mode (Default)

Uses existing `.code-together/config.json` and `.code-together/auth.json`:

```bash
./bin/member-cli -tool-name claude
```

### Override Mode

Provide both `-auth` and `-server` to override settings:

```bash
./bin/member-cli -tool-name claude -auth <token> -server http://server:9080
```

### Invalid Usage

```bash
# ERROR: Cannot override just one
./bin/member-cli -tool-name claude -server http://localhost:9080
./bin/member-cli -tool-name claude -auth <token>
```

## Architecture

The CLI tool reuses all services from the GUI client:

```
cmd/cli/
├── main.go         # Entry point, signal handling
├── config.go       # Configuration resolution
├── providers.go    # In-memory provider loading
└── main_test.go    # Unit tests
```

### Key Differences from GUI Mode

| Feature | GUI Mode | CLI Mode |
|---------|----------|----------|
| Entry Point | `main.go` (Wails) | `cmd/cli/main.go` |
| Provider Storage | Local files | In-memory from server |
| Settings | GUI dialogs | CLI flags or settings files |
| Lifecycle | Desktop app (system tray) | Foreground process |
| Platform | Desktop (macOS, Windows, Linux) | Headless (containers, servers) |

## Startup Sequence

1. **Configuration Resolution** - Load from CLI flags or `.code-together` settings
2. **Provider Loading** - Fetch from server API (in-memory)
3. **Settings Backup** - Back up current tool settings internally
4. **Proxy Enable** - Configure tool to use proxy (port 18100)
5. **Start Relay** - Begin accepting and forwarding requests
6. **Signal Handling** - Wait for SIGTERM/SIGINT

## Graceful Shutdown

On `SIGTERM` or `SIGINT`:

1. Stop accepting new requests
2. Wait for pending requests (100ms)
3. Disable proxy (restores from backup internally)
4. Close hook service
5. Exit cleanly

## Proxy Details

- **Port**: 18100
- **Endpoints**:
  - `/v1/messages` → Claude provider
  - `/responses` → Codex provider
  - `/chat/completions` → OpenCode provider
  - `/collect/{tool_name}` → Hook event collection

## Container Usage

### Dockerfile (mounted settings)

```dockerfile
FROM alpine:latest
VOLUME ["/root/.code-together"]
COPY bin/member-cli /usr/local/bin/
CMD ["member-cli", "-tool-name", "claude"]
```

### Dockerfile (environment credentials)

```dockerfile
FROM alpine:latest
ENV AUTH_TOKEN=eyJhbGc...
ENV SERVER_URL=http://server:9080
COPY bin/member-cli /usr/local/bin/
CMD ["sh", "-c", "member-cli -tool-name claude -auth $AUTH_TOKEN -server $SERVER_URL"]
```

### Kubernetes

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: member-cli
    image: your-org/member-cli
    args: ["-tool-name=claude"]
    env:
      - name: AUTH_TOKEN
        valueFrom:
          secretKeyRef:
            name: member-auth
            key: token
      - name: SERVER_URL
        value: "http://member-server:9080"
```

## Development

```bash
make cli-build            # Build CLI tool
make cli-test             # Run CLI tests
make cli-clean             # Clean CLI binary
```

## Testing

```bash
make cli-test
# ✓ PASS: All 7 test suites (28 tests)
```

Test suites:
- `TestIsValidTool` - Tool name validation
- `TestResolveConfig_AuthServerPairing` - Auth+server pairing
- `TestResolveConfig_InvalidTool` - Invalid tool handling
- `TestInMemoryProviderService` - In-memory provider loading
- `TestGetSettingsService` - Settings service factory
- `TestSettingsWrappers` - Settings wrapper implementations
- `TestSetupLogger` - Logger configuration

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "both -auth and -server must be specified together" | Provide both flags together, or neither to use settings |
| "server URL not found in settings" | Run GUI client once, or provide both flags |
| "auth token not found in settings" | Run GUI client once, or provide both flags |
| "failed to load providers from server" | Check server URL, verify token, run with `-verbose` |
| Port 18100 already in use | `lsof -ti:18100 \| xargs kill -9` |

## Files

- `main.go` - CLI entry point, signal handling
- `config.go` - Configuration resolution
- `providers.go` - In-memory provider loading from server
- `main_test.go` - Unit tests
- `README.md` - User documentation

## See Also

- [README.md](README.md) - Complete CLI usage guide
- [Issue #117](https://github.com/shtdu/code-together/issues/117) - Implementation
- [../CLAUDE.md](../CLAUDE.md) - Member client overview
