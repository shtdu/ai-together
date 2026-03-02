# Shared Tools

Utility tools for working with Claude Code hook events.

## Tools

| Tool                     | Description                                                     |
| ------------------------ | --------------------------------------------------------------- |
| **hook-schema-analyzer** | Generate Go structs and JSON schemas from hook logs             |
| **hook-events-tool**     | Split hook events from log files into organized JSON by session |
| **hook-collector**       | Collect and store hook events in BoltDB with export/cleanup     |
| **hook-browser**         | TUI browser for exploring collected hook events                 |

See each tool's directory for detailed documentation.

## Building

```bash
# Build individual tools
cd hook-collector && go build .
cd hook-browser && go build .
cd hook-events-tool && go build .
cd hook-schema-analyzer && go build .
```
