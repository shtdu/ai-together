# Member CLI Tool

A CLI-only version of the member client for headless environments (containers, Linux servers, CI/CD pipelines).

## Overview

The CLI tool provides the same core functionality as the GUI member client without requiring a display. It:
- Loads providers directly from the server (in-memory, no file pollution)
- Can reuse existing `.code-together` settings as an alternative to the GUI
- Enables/disables tool-specific proxies (Claude, Codex, OpenCode)
- Supports graceful shutdown with settings restoration

## Quick Start

### Build

```bash
make cli-build
```

Binary location: `bin/member-cli`

### Run

```bash
# Use existing .code-together settings (default)
./bin/member-cli -tool-name claude

# Override with explicit credentials
./bin/member-cli -tool-name claude -auth <token> -server http://server:9080
```

## Usage

```bash
member-cli -tool-name <tool> [options]
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `-tool-name` | Yes | Tool to configure: `claude`, `codex`, or `opencode` |
| `-auth` | Conditional | JWT auth token (must be paired with `-server`) |
| `-server` | Conditional | Server URL (must be paired with `-auth`) |
| `-verbose` | No | Enable verbose logging |

### Configuration Modes

**Settings Mode (Default)**
Uses existing `.code-together/config.json` and `.code-together/auth.json`:

```bash
./bin/member-cli -tool-name claude
```

**Override Mode**
Provide both `-auth` and `-server` to override settings:

```bash
./bin/member-cli -tool-name claude -auth <token> -server http://localhost:9080
```

**Invalid Usage** (both flags must be provided together):

```bash
# ERROR: Cannot override just one setting
./bin/member-cli -tool-name claude -server http://localhost:9080
./bin/member-cli -tool-name claude -auth <token>
```

## How It Works

### Startup Sequence

1. **Configuration Resolution**: Load from CLI flags or `.code-together` settings
2. **Provider Loading**: Fetch providers from server API (in-memory)
3. **Settings Backup**: Back up current tool settings internally
4. **Proxy Enable**: Configure tool to use the proxy (port 18100)
5. **Start Relay**: Begin accepting and forwarding requests
6. **Signal Handling**: Wait for SIGTERM/SIGINT

### Graceful Shutdown

On `SIGTERM` or `SIGINT`:

1. Stop accepting new requests
2. Wait for pending requests to complete (100ms)
3. Disable proxy (restores from backup internally)
4. Close hook service
5. Exit cleanly

### Proxy Details

- **Port**: 18100
- **Endpoints**:
  - `/v1/messages` → Claude provider
  - `/responses` → Codex provider
  - `/chat/completions` → OpenCode provider
  - `/collect/{tool_name}` → Hook event collection

## Examples

### Basic Usage

```bash
# Run Claude tool proxy
./bin/member-cli -tool-name claude

# Run with verbose logging
./bin/member-cli -tool-name claude -verbose
```

### Container Usage

**Dockerfile** (mounted settings):

```dockerfile
FROM alpine:latest
# Mount existing .code-together settings
VOLUME ["/root/.code-together"]
COPY bin/member-cli /usr/local/bin/
CMD ["member-cli", "-tool-name", "claude"]
```

**Dockerfile** (environment credentials):

```dockerfile
FROM alpine:latest
ENV AUTH_TOKEN=eyJhbGc...
ENV SERVER_URL=http://server:9080
COPY bin/member-cli /usr/local/bin/
CMD ["sh", "-c", "member-cli -tool-name claude -auth $AUTH_TOKEN -server $SERVER_URL"]
```

**Docker Compose**:

```yaml
services:
  member-cli:
    image: your-org/member-cli
    environment:
      - TOOL_NAME=claude
      # Or mount settings:
    volumes:
      - ~/.code-together:/root/.code-together:ro
```

### Kubernetes

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: member-cli
    image: your-org/member-cli
    args:
      - "-tool-name=claude"
    # Option 1: Use ConfigMap for settings
    # Option 2: Environment variables for auth
    env:
      - name: AUTH_TOKEN
        valueFrom:
          secretKeyRef:
            name: member-auth
            key: token
      - name: SERVER_URL
        value: "http://member-server:9080"
```

## Configuration Files

The CLI tool reads from the same `.code-together` directory as the GUI client:

### `~/.code-together/config.json`
```json
{
  "app": {
    "show_heatmap": false,
    "auto_start": false
  },
  "server": {
    "server_url": "http://server:9080"
  },
  "sync": {
    "last_sync_time": "2026-02-12T11:25:28.426831+08:00"
  }
}
```

### `~/.code-together/auth.json`
```json
{
  "tokens": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "expires_at": "2026-02-13T11:25:28.426831+08:00"
  },
  "user": {
    "id": 123,
    "email": "user@example.com",
    "name": "User",
    "role": "member"
  }
}
```

## Troubleshooting

### "both -auth and -server must be specified together"

**Cause**: You provided only one of `-auth` or `-server`.

**Solution**: Provide both together, or neither to use settings:

```bash
# Correct: Both flags
./bin/member-cli -tool-name claude -auth <token> -server http://server:9080

# Correct: Neither flag (uses settings)
./bin/member-cli -tool-name claude
```

### "server URL not found in settings"

**Cause**: No `.code-together/config.json` exists or `server.server_url` is empty.

**Solution**: Run GUI member client once to log in, or provide both flags:

```bash
./bin/member-cli -tool-name claude -auth <token> -server http://server:9080
```

### "auth token not found in settings"

**Cause**: No `.code-together/auth.json` exists or token is empty.

**Solution**: Run GUI member client once to log in, or provide both flags.

### "failed to load providers from server"

**Cause**: Server connection failed, authentication failed, or server returned error.

**Solutions**:
- Check server URL is correct
- Verify auth token is valid (not expired)
- Check network connectivity
- Run with `-verbose` to see API request/response details

### Port 18100 already in use

**Cause**: Another member CLI or GUI instance is running.

**Solution**: Stop the existing instance first:

```bash
# Find process using port 18100
lsof -ti:18100 | xargs kill -9

# Or use graceful shutdown
kill -TERM $(pgrep member-cli)
```

## Development

### Build

```bash
make cli-build
```

### Test

```bash
make cli-test
```

### Clean

```bash
make cli-clean
```

## Architecture

```
member/
├── cmd/cli/
│   ├── main.go         # Entry point, signal handling
│   ├── config.go       # Configuration resolution
│   ├── providers.go    # In-memory provider loading
│   ├── main_test.go    # Unit tests
│   └── README.md       # This file
├── services/           # Shared with GUI client
│   ├── providerservice.go
│   ├── providerrelay.go
│   ├── claudesettings.go
│   ├── codexsettings.go
│   └── ...
└── Makefile           # Build targets
```

## Key Differences from GUI Client

| Feature | GUI Client | CLI Tool |
|---------|-----------|----------|
| **Provider Storage** | Local files | In-memory from server |
| **Settings Persistence** | Yes (internal backup) | Yes (internal backup) |
| **UI** | Wails desktop app | None (headless) |
| **Lifecycle** | Long-running with system tray | Foreground process |
| **Configuration** | GUI dialogs | CLI flags or settings files |

## Design Decisions

1. **In-Memory Providers**: No file pollution, always fresh from server
2. **Auth+Server Pairing**: Tokens are server-specific, must be provided together
3. **Reuse Settings Files**: Seamless switching between GUI and CLI
4. **Graceful Shutdown**: Proper SIGTERM/SIGINT handling
5. **No Daemon Mode**: PID management by external scripts/orchestration

## See Also

- [Member Client Documentation](../CLAUDE.md)
- [Provider Configuration](../../docs/provider_management/)
- [Issue #117](https://github.com/shtdu/code-together/issues/117)
