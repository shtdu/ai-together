# Code Together Server

## Air (Live Reload)

From `server/`, run:

```bash
# install air
# go install github.com/cosmtrek/air@latest
air
```

This builds to `server/tmp/server` and runs `server/run-server.sh`.

## Gin Proxy Warning

If you see:

```
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe.
```

Set trusted proxies explicitly in `server/server.go` after `router := gin.New()`:

```go
if err := router.SetTrustedProxies(nil); err != nil {
	log.Fatal("Failed to set trusted proxies:", err)
}
```

If the server runs behind a reverse proxy, set the proxy IP/CIDR instead of `nil`.

## Role-Based Access Control (RBAC)

The server implements role-based permissions using Casbin:

| Resource | Action | Member | Manager |
|----------|--------|--------|---------|
| providers | read | ✅ | ✅ |
| providers | write | ❌ | ✅ |
| config | read | ✅ | ✅ |
| config | sync | ✅ | ✅ |
| config | write | ❌ | ✅ |
| license | read | ✅ | ✅ |
| license | write | ❌ | ✅ |

### Endpoint Permissions

**Read-only (members & managers):**
- `GET /api/v1/providers` - List providers
- `GET /api/v1/providers/:id/stats` - Get provider statistics
- `GET /api/v1/user/profile` - Get current user profile

**Write operations (managers only):**
- `POST /api/v1/providers` - Create provider
- `PUT /api/v1/providers/:id` - Update provider
- `DELETE /api/v1/providers/:id` - Delete provider
- `POST /api/v1/providers/:id/test` - Test provider connectivity
- `POST /api/v1/providers/:id/enable` - Enable provider
- `DELETE /api/v1/providers/:id/disable` - Disable provider

**Public endpoints (no authentication):**
- `GET /health` - Health check
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Refresh access token
- `POST /auth/verify` - Verify token validity

### License System

The Code Together server includes a license management system that controls user seat limits and AI provider configuration limits based on license tier. When no valid license is active, the system defaults to Tier 0.0 (Community) with basic limits.

#### License Impact

**User Seats**: Each license tier limits the maximum number of users per tenant
- Tier 0.0 (default): 3 users
- Tier 1.0+: Up to 100+ users depending on tier

**Provider Limits**: Controls how many AI provider configurations can be created per provider kind (claude, codex, opencode)
- Tier 0.0 (default): 2 providers per kind
- Tier 1.0+: Unlimited providers

**Key Behavior**:
- System gracefully degrades to default license when no valid license exists
- Existing users and providers remain functional when license expires
- New user/provider creation may be blocked if limits are exceeded

#### Environment Variables

- `LICENSE_PUBLIC_KEY`: Base64-encoded Ed25519 public key for license validation
  - Generate keypair with: `go run server/scripts/keygen/main.go`

For complete license system documentation including architecture, tier details, API endpoints, and troubleshooting, see [LICENSE_SYSTEM.md](LICENSE_SYSTEM.md).
