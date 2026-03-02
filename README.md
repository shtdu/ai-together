# Code Together

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/code-together/code-together)](https://goreportcard.com/report/github.com/code-together/code-together)

A provider proxy platform for centralized management of AI coding tools (Claude Code, Codex, OpenCode), offering transparent request routing, intelligent failover, and team collaboration features.

## Key Features

- Seamlessly switch between different providers without restarting AI tools
- Support for automatic failover across multiple providers to ensure service continuity
- Request-level usage statistics for clear cost tracking
- Team collaboration and centralized configuration management
- Cross-platform desktop application built with [Wails 3](https://v3.wails.io)

## System Architecture

Code Together consists of four main components:

- **Member Client** (`member/`) - Desktop client running a local HTTP proxy service
- **Server** (`server/`) - Backend API service providing centralized management and collaboration features
- **Manager UI** (`manager/`) - Administrative interface for team analytics and user management
- **Integration Tests** (`integration/`) - Independent integration test module

### How It Works

The application creates an HTTP proxy server on local port 18100 at startup and automatically updates AI tool configurations to point to this proxy. The proxy only exposes key compatible endpoints:

- `/v1/messages` - Routes to configured Claude provider
- `/responses` - Routes to Codex provider
- `/v1/chat/completions` - Routes to OpenCode provider

Requests are dynamically routed by the proxy handler based on current provider priority and enabled status, with automatic failover on failures.

## Download and Installation

[macOS](https://github.com/code-together/code-together/releases) | [Windows](https://github.com/code-together/code-together/releases)

## User Manual

For complete documentation, see: [**Member Client User Manual**](docs/manuals/member.md)

The manual includes:
- Quick start guide
- Core feature explanations
- Configuration and settings
- User roles and permissions
- Frequently asked questions

## Development Environment

### Required Tools
- Go 1.24+
- Node.js 18+ and npm/pnpm
- Wails 3 CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`

### Optional Tools
- PostgreSQL 14+ (required only for server development)
- Air (hot reload): `go install github.com/cosmtrek/air@latest`
- mingw-w64 (Windows cross-compilation on macOS): `brew install mingw-w64`

## Development

### Quick Start

```bash
# View all available commands
make help

# Build all components
make build

# Run in development mode (server + member)
make dev

# Run tests
make test

# Clean build artifacts
make clean
```

### Module Development

Each module has its own Makefile with detailed commands:

```bash
# Server module
cd server && make help      # View all server commands
make server-dev             # Run server in development mode from root

# Member client
cd member && make help      # View all member commands
make member-dev             # Run member in development mode from root

# Manager UI
cd manager && make help     # View all manager commands
make manager-dev            # Run manager in development mode from root

# Integration tests
cd integration && make help # View all integration commands
make integration-test       # Run integration tests from root
```

## Build Process

### Quick Build

```bash
# Build all components (current platform)
make build

# Build release version (current platform)
make release

# Build release versions for all platforms
make release-all
```

### Individual Builds

```bash
# Server
make server-build
make server-linux        # Cross-compile for Linux

# Member client
make member-build
make member-release      # Package .app / .exe

# Manager UI
make manager-build
```

### Windows Cross-Compilation (macOS)

```bash
# Install cross-compilation tools
brew install mingw-w64

# Build Windows version
make member-win
```

For detailed build commands, refer to each module's Makefile:
- `cd server && make help`
- `cd member && make help`
- `cd manager && make help`

## Common Issues

- **".app cannot be opened"**: Run `cd member && make update-assets` before building
- **macOS cross-compilation**: Terminal needs full disk access permissions
- **Server startup**: Configure `DATABASE_URL` in `server/.env` file

For more issues, refer to module-specific documentation.

## Development Documentation

- [CLAUDE.md](CLAUDE.md) - Project overview and architecture
- [member/CLAUDE.md](member/CLAUDE.md) - Member client development
- [server/CLAUDE.md](server/CLAUDE.md) - Server development
- [server/README.md](server/README.md) - Server RBAC permissions
- [manager/CLAUDE.md](manager/CLAUDE.md) - Manager UI development
- [integration/INTEGRATION_TEST_SETUP.md](integration/INTEGRATION_TEST_SETUP.md) - Integration test setup
- [integration/integration.design.md](integration/integration.design.md) - Integration test design

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Development setup
- Code style
- Pull request process
- Testing requirements
- Commit message conventions

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Security

For security policies and vulnerability reporting, see [SECURITY.md](SECURITY.md).

## Chinese Documentation

中文文档请参考: [README_CN.md](README_CN.md)
