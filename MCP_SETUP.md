# PostgreSQL MCP Server Setup

## Overview

This project has PostgreSQL MCP (Model Context Protocol) servers configured to allow Claude Code to directly interact with your databases.

## Configuration

**File:** `.mcp.json` (project root)

Two MCP servers are configured:

### 1. **postgres-test** (Test Database)
- **Database:** `codetogether_test`
- **Purpose:** Integration test database, schema inspection, test data exploration
- **Connection:** `postgresql://localhost:5432/codetogether_test?sslmode=disable`

### 2. **postgres-dev** (Development Database)
- **Database:** `codetogether` (main development database)
- **Purpose:** Development schema inspection, query debugging
- **Connection:** `postgresql://localhost:5432/codetogether?sslmode=disable`

## Installation

The MCP server uses `@5shuang/mcp-postgres-server` which runs via `npx` - no installation needed!

## Usage

### With Claude Code CLI

When you start Claude Code in this project, the MCP servers will be automatically available. You can:

1. **Query the database:**
   ```
   "Show me all providers in the test database"
   "List all users with role 'manager'"
   ```

2. **Inspect schema:**
   ```
   "What tables exist in the test database?"
   "Show me the schema for the providers table"
   ```

3. **Debug test data:**
   ```
   "How many license records exist in the test database?"
   "Show me the last 10 request_log entries"
   ```

### Database Schemas

**Test Database (`codetogether_test`):**
- `providers` - AI provider configurations
- `request_log` - Detailed usage tracking
- `users` - User accounts
- `tenants` - Organizations
- `licenses` - License records
- `teams`, `team_members` - Team management
- `team_usage_summary` - Aggregated statistics

## Common Use Cases

### 1. Verify Test Data After Running Tests
```bash
# Run integration tests
go test -v

# Then ask Claude:
"Show me the providers created during the test"
"Verify that all test users were cleaned up"
```

### 2. Inspect Schema During Development
```
"What columns does the providers table have?"
"Show me the foreign key relationships"
```

### 3. Debug Migration Issues
```
"Did the last migration run successfully?"
"Show me the current schema version"
```

### 4. Validate Multi-Tenant Data Isolation
```
"Show me tenants and their counts"
"Verify all providers have valid tenant_id"
```

## Troubleshooting

### MCP Server Not Connecting

1. **Check PostgreSQL is running:**
   ```bash
   brew services list | grep postgresql
   # Or
   psql -h localhost -d postgres -c "SELECT version();"
   ```

2. **Verify database exists:**
   ```bash
   psql -h localhost -d postgres -c "\l" | grep codetogether
   ```

3. **Test connection manually:**
   ```bash
   psql -h localhost -d codetogether_test -c "\dt"
   ```

4. **Check MCP configuration:**
   ```bash
   cat .mcp.json
   ```

### Authentication Issues

If your PostgreSQL requires a password, update `.mcp.json`:

```json
{
  "mcpServers": {
    "postgres-test": {
      "command": "npx",
      "args": ["-y", "@5shuang/mcp-postgres-server"],
      "env": {
        "DATABASE_URL": "postgresql://username:password@localhost:5432/codetogether_test?sslmode=disable"
      }
    }
  }
}
```

### MCP Server Not Responding

1. **Restart Claude Code** to reload MCP configuration
2. **Check npm is available:** `which npx`
3. **Test MCP server directly:**
   ```bash
   DATABASE_URL="postgresql://localhost:5432/codetogether_test?sslmode=disable" \
     npx -y @5shuang/mcp-postgres-server
   ```

## Security Notes

⚠️ **Important:** `.mcp.json` is in `.gitignore` - it may contain sensitive database credentials.

- Never commit database passwords to version control
- Use environment variables for sensitive data
- The current configuration uses peer authentication (no password)
- For production, use read-only database users

## Advanced Configuration

### Custom Database Connection

To add a new database connection, edit `.mcp.json`:

```json
{
  "mcpServers": {
    "postgres-production": {
      "command": "npx",
      "args": ["-y", "@5shuang/mcp-postgres-server"],
      "env": {
        "DATABASE_URL": "postgresql://user:pass@prod-host:5432/codetogether_prod?sslmode=require"
      }
    }
  }
}
```

### Read-Only User (Recommended)

Create a read-only user for safer MCP access:

```sql
-- Create read-only user
CREATE USER mcp_readonly WITH PASSWORD 'secure_password';

-- Grant read-only access
GRANT CONNECT ON DATABASE codetogether_test TO mcp_readonly;
GRANT USAGE ON SCHEMA public TO mcp_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO mcp_readonly;

-- Automatically grant select on new tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO mcp_readonly;
```

Then update `.mcp.json` to use the read-only user.

## References

- **MCP Protocol:** [What Is MCP? The 2026 Guide](https://generect.com/blog/what-is-mcp/)
- **PostgreSQL MCP Server:** [@5shuang/mcp-postgres-server on npm](https://www.npmjs.com/package/@5shuang/mcp-postgres-server)
- **Alternative Implementation:** [ahmedmustahid/postgres-mcp-server](https://github.com/ahmedmustahid/postgres-mcp-server)
- **Integration Guide:** [Integrating PostgreSQL MCP Server with Docker and Claude Desktop](https://dasroot.net/posts/2026/01/integrating-postgresql-mcp-server-docker-claude-desktop/)
- **Chinese Tutorial:** [Claude Code 配置MCP 完全指南](https://gaccode.store/post/claude-code-mcp-configuration)
