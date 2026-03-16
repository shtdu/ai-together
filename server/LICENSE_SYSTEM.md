# License System

## Overview

The AI Together license system controls access to features based on configurable limits for user seats and AI provider configurations. The system uses Ed25519 cryptographic signing to ensure license integrity and supports multiple license tiers with increasing capabilities.

### Key Concepts

- **Tier**: License version (e.g., "0.0", "1.0", "2.0", "3.0") that determines feature set
- **Seats**: Maximum number of users allowed in a tenant
- **Provider Limits**: Per-kind limits on AI providers (claude, codex, opencode)
- **Default License**: Fallback tier 0.0 (3 seats, 2 providers per kind) when no valid license exists

### Architecture

```
┌─────────────────┐      ┌──────────────────┐      ┌─────────────────┐
│   License       │      │   License        │      │   License       │
│   Handler       │─────▶│   Service        │─────▶│   Repository    │
│   (HTTP)        │      │   (Business      │      │   (Database)    │
│                 │      │    Logic)        │      │                 │
└─────────────────┘      └──────────────────┘      └─────────────────┘
         │                        │                         │
         │                   ┌───┴────┐                  │
         │                   │ Cache  │                  │
         │                   │ (Load  │                  │
         │                   │ at     │                  │
         │                   │Startup)│                  │
         │                   └───┬────┘                  │
         │                        │                         │
         │                        ▼                         │
         │                 ┌──────────────┐               │
         │                 │   go-license │               │
         │                 │   (Crypto)   │               │
         └─────────────────┴──────────────┘               │
              Middleware Injection                           │
                       │                                    │
                       └────────────────────────────────────┘
                               All licenses loaded once at startup
```

### In-Memory License Cache

The license system uses an in-memory cache for optimal performance:

**Startup Loading:**
- All tenant licenses are loaded from database at service initialization
- Each license is decoded and validated once using Ed25519 signature verification
- Decoded values (tier, seats) from the PEM are stored as the source of truth
- Invalid or expired licenses are skipped with warnings logged

**Cache Structure:**
```go
type LicenseService struct {
    licenseRepo   repository.LicenseRepositoryInterface
    publicKey     ed25519.PublicKey
    licenses      map[int64]*models.License  // tenantID → decoded license
    licensesMutex sync.RWMutex               // Thread-safe concurrent access
}
```

**Performance Benefits:**
- **Before:** 1 DB query + 1 PEM decode + 1 signature validation per request
- **After:** 1 mutex RLock/RUnlock per request (nanoseconds)
- All cryptographic operations performed once at startup
- Zero database queries for license checks on cache hits

**Thread Safety:**
- Concurrent reads use `RLock()` (multiple goroutines can read simultaneously)
- Writes use `Lock()` (exclusive access during license activation)
- Race detector verified with `-race` flag (50+ tests passing)

**Cache Invalidation:**
- Cache is refreshed immediately when a new license is activated
- License expiration checked on every read (expired licenses return defaults)
- Database modifications are detected by `VerifyLicenseIntegrity()`

**Source of Truth:**
- Decoded values from license PEM are trusted
- Database columns (tier, seats) are storage only, not used for validation
- Prevents database tampering (changing tier/seats directly in DB has no effect)

## License Tiers

### Tier Comparison Table

| Tier | Name         | Max Users | Providers Per Kind | Features                             |
| ---- | ------------ | --------- | ------------------ | ------------------------------------ |
| 0.0  | Community    | 3         | 2                  | Basic provider management            |
| 1.0  | Standard     | 10        | Unlimited (-1)     | Unlimited providers, basic analytics |
| 2.0  | Professional | 25        | Unlimited (-1)     | Full analytics, team management      |
| 3.0  | Enterprise   | 100       | Unlimited (-1)     | SSO integration, audit logs          |

### Feature Matrix

| Feature                   | Community (0.0) | Standard (1.0) | Professional (2.0) | Enterprise (3.0) |
| ------------------------- | --------------- | -------------- | ------------------ | ---------------- |
| Max Users                 | 3               | 10             | 25                 | 100              |
| Providers per Kind        | 2               | Unlimited      | Unlimited          | Unlimited        |
| Basic Provider Management | ✅               | ✅              | ✅                  | ✅                |
| Usage Statistics          | ✅               | ✅              | ✅                  | ✅                |
| Basic Analytics           | ❌               | ✅              | ✅                  | ✅                |
| Full Analytics            | ❌               | ❌              | ✅                  | ✅                |
| Team Management           | ❌               | ❌              | ✅                  | ✅                |
| SSO Integration           | ❌               | ❌              | ❌                  | ✅                |
| Audit Logs                | ❌               | ❌              | ❌                  | ✅                |

### Use Cases by Tier

**Community (0.0)** - Small teams evaluating the platform
- Up to 3 users
- Limited provider configurations (2 per kind)
- Basic provider management

**Standard (1.0)** - Growing teams needing flexibility
- Up to 10 users
- Unlimited provider configurations
- Basic usage analytics

**Professional (2.0)** - Organizations requiring advanced features
- Up to 25 users
- Unlimited provider configurations
- Full analytics and team management

**Enterprise (3.0)** - Large organizations with security requirements
- Up to 100 users
- Unlimited provider configurations
- SSO integration and comprehensive audit logging

## License Key Format

### Cryptographic Signing

The license system uses [go-license](https://github.com/vitalvas/go-license) with Ed25519 signatures:

- **Algorithm**: Ed25519 (EdDSA)
- **Key Size**: 32 bytes public key, 64 bytes signature
- **Format**: PEM-encoded license file

### License PEM Format

```pem
-----BEGIN LICENSE KEY-----
<base64-encoded license data>
-----END LICENSE KEY-----
```

### Embedded License Data

Each license contains a JSON payload with custom data:

```json
{
  "tier": "1.0",
  "seats": 10
}
```

### License Fields

| Field        | Type   | Description                                   |
| ------------ | ------ | --------------------------------------------- |
| `id`         | string | Unique license identifier (auto-generated UUID) |
| `customer`   | string | Customer name/description                     |
| `tier`       | string | License tier (e.g., "0.0", "1.0", "2.0")      |
| `seats`      | int    | Maximum number of users allowed               |
| `issued_at`  | int64  | Unix timestamp when license was issued        |
| `expires_at` | int64  | Unix timestamp when license expires           |
| `data`       | bytes  | Custom JSON payload with tier/seats           |

### License Identifiers

- **`id`**: Auto-generated UUID that uniquely identifies each license
  - Generated automatically by the keygen script using `uuid.New()`
  - Stored in the database as `license_id`
  - Used for license tracking and identification
  - Format: standard UUID (e.g., `550e8400-e29b-41d4-a716-446655440000`)

- **`customer`**: Human-readable customer description
  - Provided via `-customer` flag during license generation
  - Examples: `"Daniel"`, `"Acme Corp"`, `"Shanghai Branch"`
  - Stored within the signed license for reference

**Design Rationale**: For on-premise deployments, the auto-generated UUID ensures:
- Each license has a globally unique identifier
- No manual identifier management required
- Easy tracking and reference in support systems

### Generating License Keys

The server includes a key generation utility:

```bash
# Generate new keypair
cd server
go run scripts/keygen/main.go

# Output:
# Public key (base64): <base64-encoded-public-key>
# Private key (base64): <base64-encoded-private-key>
```

**Important**: Store the private key securely. It is needed to generate licenses but should never be deployed to production servers.

## Database Schema

### License Fields in Tenants Table

```sql
ALTER TABLE tenants ADD COLUMN (
  license_id            TEXT,
  license_tier          TEXT,
  license_seats         INTEGER,
  license_key           TEXT,
  license_issued_at     TIMESTAMP,
  license_expires_at    TIMESTAMP
);
```

### Field Descriptions

| Column               | Type      | Nullable | Description                  |
| -------------------- | --------- | -------- | ---------------------------- |
| `license_id`         | TEXT      | Yes      | Unique license identifier    |
| `license_tier`       | TEXT      | Yes      | License tier (e.g., "1.0")   |
| `license_seats`      | INTEGER   | Yes      | Maximum users allowed        |
| `license_key`        | TEXT      | Yes      | Full PEM-encoded license key |
| `license_issued_at`  | TIMESTAMP | Yes      | When the license was issued  |
| `license_expires_at` | TIMESTAMP | Yes      | When the license expires     |

### NULL Handling

- All license fields are nullable
- When `license_id` is NULL or `license_expires_at` is in the past, the system uses default license
- Partial data (e.g., only `license_tier` set) is treated as invalid and falls back to defaults

### Migration

License system was added in migration `0004_add_license.up.sql`:

```sql
-- Add license fields to tenants table
ALTER TABLE tenants ADD COLUMN license_id TEXT;
ALTER TABLE tenants ADD COLUMN license_tier TEXT;
ALTER TABLE tenants ADD COLUMN license_seats INTEGER;
ALTER TABLE tenants ADD COLUMN license_key TEXT;
ALTER TABLE tenants ADD COLUMN license_issued_at TIMESTAMP;
ALTER TABLE tenants ADD COLUMN license_expires_at TIMESTAMP;
```

## API Documentation

### GET /api/v1/license

Get current license status, usage statistics, and tier information.

**Authentication**: Required (Bearer token)

**Permissions**: Members and Managers

**Response**:
```json
{
  "license": {
    "license_id": "CL-2024-001",
    "tier": "1.0",
    "tier_name": "Standard",
    "seats": 10,
    "max_providers_per_kind": -1,
    "issued_at": "2024-01-15T00:00:00Z",
    "expires_at": "2025-01-15T00:00:00Z"
  },
  "usage": {
    "current_users": 5,
    "seats_remaining": 5,
    "provider_counts": {
      "claude": 3,
      "codex": 1
    }
  },
  "status": {
    "has_active_license": true,
    "days_remaining": 180,
    "can_add_user": true,
    "can_add_provider": {
      "claude": true,
      "codex": true,
      "opencode": true
    }
  }
}
```

**Response with Default License**:
```json
{
  "license": {
    "license_id": null,
    "tier": "0.0",
    "tier_name": "Community",
    "seats": 3,
    "max_providers_per_kind": 2,
    "issued_at": null,
    "expires_at": null
  },
  "usage": {
    "current_users": 1,
    "seats_remaining": 2,
    "provider_counts": {
      "claude": 1
    }
  },
  "status": {
    "has_active_license": false,
    "is_using_defaults": true,
    "days_remaining": 0,
    "can_add_user": true,
    "can_add_provider": {
      "claude": true,
      "codex": true,
      "opencode": true
    }
  }
}
```

### POST /api/v1/license/activate

Activate or upgrade a license with a license key.

**Authentication**: Required (Bearer token)

**Permissions**: Managers only

**Request**:
```json
{
  "license_key": "-----BEGIN LICENSE KEY-----\nMIIB...\n-----END LICENSE KEY-----"
}
```

**Success Response** (200):
```json
{
  "message": "License activated successfully"
}
```

**Error Responses**:

- 400 Bad Request - Invalid license key:
```json
{
  "error": "invalid license key: signature verification failed"
}
```

- 400 Bad Request - Expired license:
```json
{
  "error": "license has expired"
}
```

- 400 Bad Request - Invalid license data:
```json
{
  "error": "invalid license data: tier is required"
}
```

### GET /api/v1/license/tiers

Get information about available license tiers (public endpoint).

**Authentication**: Not required

**Response**:
```json
{
  "tiers": [
    {
      "tier": "0.0",
      "name": "Community",
      "max_providers_per_kind": 2,
      "features": ["Basic provider management"]
    },
    {
      "tier": "1.0",
      "name": "Standard",
      "max_providers_per_kind": -1,
      "features": ["Unlimited providers", "Basic analytics"]
    },
    {
      "tier": "2.0",
      "name": "Professional",
      "max_providers_per_kind": -1,
      "features": ["Unlimited providers", "Full analytics", "Team management"]
    },
    {
      "tier": "3.0",
      "name": "Enterprise",
      "max_providers_per_kind": -1,
      "features": ["Unlimited providers", "Full analytics", "SSO integration", "Audit logs"]
    }
  ]
}
```

## License Validation Flow

### Startup Flow (Service Initialization)

```
┌──────────────────┐
│ Service Created  │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Load All Licenses│
│ from DB (once)   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ For Each License │
└────────┬─────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌──────┐  ┌──────┐
│ Decode│  │ Skip │
│ PEM   │  │ (no  │
│ &     │  │ key) │
│ Verify│  └──────┘
└──┬───┘
   │
   ▼
┌──────────────────┐
│ Extract tier/    │
│ seats from PEM   │
│ (source of truth)│
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Store in Cache   │
│ (map[tenantID])  │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Ready to Serve   │
│ Requests         │
└──────────────────┘
```

### Request Processing

```
┌──────────────┐    Auth Middleware    ┌──────────────┐
│   Incoming   │──────────────────────▶│   License    │
│   Request    │                       │ Middleware   │
└──────────────┘                       └──────────────┘
                                               │
                                               ▼
                                        ┌──────────────┐
                                        │ Get License  │
                                        │ from Cache   │
                                        │ (RLock - nanos)│
                                        └──────────────┘
                                               │
                                    ┌──────────┴──────────┐
                                    │                     │
                               Valid License        Invalid/Null
                               (not expired)        (or expired)
                                    │                     │
                                    ▼                     ▼
                            ┌──────────────┐    ┌──────────────┐
                            │ Use License  │    │  Use Default │
                            │ from Cache   │    │  (Tier 0.0)  │
                            └──────────────┘    └──────────────┘
                                    │                     │
                                    └──────────┬──────────┘
                                               ▼
                                        ┌──────────────┐
                                        │ Inject into  │
                                        │ Request Ctx  │
                                        └──────────────┘
                                               │
                                               ▼
                                        ┌──────────────┐
                                        │   Handler    │
                                        │  Processing  │
                                        └──────────────┘
```

### Middleware Injection

The `LicenseMiddleware` injects the effective license into every request context:

```go
func LicenseMiddleware(licenseService LicenseServiceInterface) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.Get("user")
        authenticatedUser := user.(*models.User)

        // Get effective license (or default)
        license := licenseService.GetEffectiveLicense(ctx, tenantID)
        c.Set("license", license)
        c.Set("has_active_license", licenseService.HasActiveLicense(ctx, tenantID))

        c.Next()
    }
}
```

**Important**: The middleware never blocks requests. It always provides a license (either the stored license or default fallback).

### Handler License Checks

Handlers can access the license from context:

```go
license, exists := c.Get("license")
if exists {
    lic := license.(*models.License)
    if lic.MajorTier() == 0 {
        // Enforce provider limits for tier 0
    }
}
```

### Default License Behavior

When no valid license exists, the system falls back to default values:

| Constant              | Value | Description              |
| --------------------- | ----- | ------------------------ |
| `DefaultLicenseTier`  | "0.0" | Community tier           |
| `DefaultLicenseSeats` | 3     | Max 3 users              |
| `DefaultMaxProviders` | 2     | Max 2 providers per kind |

**When Defaults Are Used**:
1. No license record exists in database
2. License has expired (`expires_at` < now)
3. License data is invalid (missing required fields)

## Default License Behavior

### Graceful Degradation

The license system is designed to never block access due to missing or invalid licenses. Instead, it applies tier 0.0 restrictions:

- **Max Users**: 3 seats
- **Max Providers per Kind**: 2 (for each provider type: claude, codex, opencode)

### Enforcement Points

License limits are enforced at:

1. **User Creation**: `CanAddUser()` checks seat limits before creating new users
2. **Provider Creation**: `CanAddProvider()` checks provider limits per kind
3. **Team Management**: Indirectly enforced through user limits

### Upgrading from Defaults

To upgrade from the default license:

1. Obtain a license key (generated with private key)
2. Call `POST /api/v1/license/activate` with the license key
3. System validates signature and expiration
4. License is stored in database
5. All subsequent requests use the new license

## For Developers

For service layer patterns, handler usage, and enforcement patterns in daily development, see [CLAUDE.md](CLAUDE.md#service-layer-patterns).

### License Cache Behavior

**Important:** The license service maintains an in-memory cache that is loaded at startup:

```go
// Cache is populated during service initialization
service := NewLicenseService(repo, publicKey)

// All subsequent reads use cache (no DB queries)
license := service.GetEffectiveLicense(ctx, tenantID)
```

**Implications for development:**

1. **Hot reload doesn't reload licenses** - Changes to database require service restart
2. **Cache is thread-safe** - Multiple goroutines can read simultaneously using RLock()
3. **License activation updates cache** - New licenses are immediately available
4. **Expired licenses are filtered** - Cache expiration check happens on every read

**Testing with cached licenses:**

```go
// Use helper to create service with pre-loaded licenses
service, mockRepo, license := newMockServiceWithLicenses(
    t, privateKey, publicKey, tenantID, "2.0", 25, 365,
)

// License is already in cache - no DB query needed
effective := service.GetEffectiveLicense(ctx, tenantID)
assert.Equal(t, "2.0", effective.Tier)
```

### Performance Considerations

**Cache Performance:**
- Read operations: ~50-100ns (mutex RLock + map lookup)
- No database queries
- No cryptographic operations
- Thread-safe concurrent reads

**Startup Cost:**
- One-time load of all tenant licenses
- PEM decode + signature validation for each license
- Amortized over thousands of requests
- Typically completes in milliseconds

**Memory Usage:**
- Bounded by number of tenants (typically small)
- Approx. 1KB per tenant in cache
- Negligible memory overhead

### Adding New License Tiers

To add a new license tier:

1. **Update tier definitions** in `services/license_service.go`:

```go
func (s *LicenseService) GetTiers() map[string]interface{} {
    return map[string]interface{}{
        "tiers": []map[string]interface{}{
            {
                "tier":                  "4.0",
                "name":                  "Ultra",
                "max_providers_per_kind": -1,
                "features":              []string{"All features", "Priority support"},
            },
            // ... existing tiers
        },
    }
}
```

2. **Update documentation** in `LICENSE_SYSTEM.md` and `README.md`
3. **Generate license keys** with the new tier
4. **Test tier behavior** with integration tests

### Testing License Scenarios

See `server/docs/license_testing_guide.md` for comprehensive testing instructions.

Quick test:

```bash
cd server
go test ./services/... -run TestLicense
go test ./handlers/... -run TestLicenseHandler

# Run with race detector to verify thread safety
go test ./services/... -run TestLicense -race
```

**Thread Safety Testing:**

The license cache is designed for concurrent access. Always test with the race detector:

```bash
# Run all license tests with race detection
go test ./services -race -v -run TestLicenseService

# Expected: All tests pass with no race warnings
# 50+ test cases covering cache reads, writes, and concurrent access
```

**Cache Behavior in Tests:**

- Tests use `newMockServiceWithLicenses()` to preload cache with encoded licenses
- Invalid/expired licenses are tested by preloading them into cache
- Cache miss scenarios tested by not preloading licenses (returns defaults)
- Thread safety verified with `-race` flag (all tests passing)

### Debugging License Issues

Enable debug logging to trace license validation:

```bash
# Set debug mode in .env
DEBUG=true

# Check logs for license validation messages
tail -f /var/log/codetogether/server.log
```

**Cache-related debugging**:

```bash
# Check server startup logs for cache loading
grep "License loading complete" /var/log/codetogether/server.log

# Verify cache is populated (should show tenant count)
grep "loaded.*licenses for.*tenants" /var/log/codetogether/server.log

# Check for license decode warnings
grep "WARN.*Invalid license key" /var/log/codetogether/server.log

# License activation confirms cache refresh
grep "License activated for tenant" /var/log/codetogether/server.log
```

**Common debugging commands**:

```sql
-- Check tenant license in database
SELECT id, name, license_id, license_tier, license_seats,
       license_issued_at, license_expires_at
FROM tenants WHERE id = 1;

-- Count current users
SELECT COUNT(*) FROM users WHERE tenant_id = 1;

-- Count providers by kind
SELECT kind, COUNT(*) FROM providers p
JOIN teams t ON p.team_id = t.id
WHERE t.tenant_id = 1
GROUP BY kind;
```

### Extending License System

The license system can be extended to support:

- **Feature flags**: Add boolean flags to `LicenseData` for granular feature control
- **Time-limited trials**: Short-term licenses with automatic expiration
- **Concurrent user limits**: Instead of total users, track active sessions
- **Usage quotas**: Limit API calls or token usage per tier
- **Custom tiers**: Per-tenant negotiated limits

Example extension:

```go
type LicenseData struct {
    Tier      string   `json:"tier"`
    Seats     int      `json:"seats"`
    Features  []string `json:"features"`  // Custom feature flags
    Quota     int64    `json:"quota"`     // API call quota
}
```

## Troubleshooting

### Common Errors

#### "invalid license key: signature verification failed"

**Cause**: License key was not signed with the matching private key, or the public key is misconfigured.

**Solutions**:
1. Verify `LICENSE_PUBLIC_KEY` environment variable is set correctly
2. Ensure the license was signed with the corresponding private key
3. Check for whitespace or encoding issues in the license key

**Debug**:
```bash
echo $LICENSE_PUBLIC_KEY | base64 -d | od -A x -t x1 -v
```

#### "license has expired"

**Cause**: The license's expiration date has passed.

**Solutions**:
1. Generate a new license with a future expiration date
2. Contact license provider for renewal
3. System will fall back to default license tier

#### "invalid license data: tier is required"

**Cause**: The custom data embedded in the license is malformed.

**Solutions**:
1. Verify license generation includes valid JSON data
2. Check that `tier` and `seats` fields are present
3. Ensure data is valid JSON before signing

### License Activation Failures

**Problem**: License activation returns 400 error

**Debug steps**:

1. **Verify license format**:
```bash
# Check PEM format
cat license.key | head -1
# Should output: -----BEGIN LICENSE KEY-----
```

2. **Check public key configuration**:
```bash
# In server/.env
grep LICENSE_PUBLIC_KEY .env
```

3. **Test with validation endpoint**:
```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d @license.json \
  http://localhost:9080/api/v1/license/activate
```

### Expiration Handling

**Behavior when license expires**:
- System falls back to default tier 0.0 (3 seats, 2 providers per kind)
- Existing users and providers remain functional
- New users/providers may be blocked if limits exceeded
- Graceful degradation - no service interruption

**To renew**:
1. Generate new license with extended expiration
2. Activate via `POST /api/v1/license/activate`
3. System immediately uses new license

**Expiration warnings**:
The `/api/v1/license` endpoint returns `days_remaining` for UI warnings:

```json
{
  "status": {
    "days_remaining": 7,
    "has_active_license": true
  }
}
```

### Integration Issues

**Problem**: License changes in database not taking effect

**Cause**: Licenses are cached at startup and changes require cache refresh.

**Solutions**:
1. Restart the server to reload all licenses from database
2. Use `POST /api/v1/license/activate` to refresh cache immediately
3. Call `VerifyLicenseIntegrity()` to detect and correct database tampering

**Debug**:
```bash
# Check when licenses were loaded
grep "License loading complete" server.log

# Verify cache has expected license
curl -H "Authorization: Bearer <token>" \
  http://localhost:9080/api/v1/license | jq '.license.tier'

# Activate license to refresh cache
curl -X POST http://localhost:9080/api/v1/license/activate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"license_key": "..."}'
```

**Problem**: License middleware not injecting license

**Debug**:
```go
// In handler, check if license exists
license, exists := c.Get("license")
if !exists {
    log.Error("License not found in context")
}
```

**Solutions**:
1. Verify `LicenseMiddleware` is registered in `server.go`
2. Check middleware order (should be after AuthMiddleware)
3. Ensure JWT authentication succeeds

**Problem**: Limits not enforced

**Debug**:
```bash
# Check current usage
curl -H "Authorization: Bearer <token>" \
  http://localhost:9080/api/v1/license

# Verify can_add_user and can_add_provider fields
```

**Solutions**:
1. Check `CanAddUser()` and `CanAddProvider()` are called before creation
2. Verify service layer uses license checks
3. Review middleware is properly injecting license

## Related Files

### Core License Files

- `models/license.go` - License data structures and validation
- `services/license_service.go` - Business logic for license operations, in-memory cache implementation
  - `loadAllLicenses()` - Startup loading of all tenant licenses
  - `GetEffectiveLicense()` - Cache read with concurrent access
  - `ActivateLicense()` - License activation with cache refresh
  - `licenses` map - In-memory cache of decoded licenses
  - `licensesMutex` - Thread-safe read/write lock
- `handlers/license_handler.go` - HTTP endpoints for license management
- `middleware/license_middleware.go` - Request context injection
- `repository/license_repository.go` - Database operations
  - `GetAllLicenses()` - Bulk license loading for cache initialization

### Database

- `internal/db/license.sql` - SQL queries for license operations
- `internal/db/license.sql.go` - Generated Go code (sqlc)
- `migrations/0004_add_license.up.sql` - License schema migration

### Testing

- `services/license_service_test.go` - Unit tests for license service
- `handlers/license_handler_mock_test.go` - Handler tests with mocks
- `license_integration_test.go` - Integration tests
- `scripts/keygen/main.go` - License key generation utility

### Scripts

- `scripts/import_license.sh` - License import utility script

## License Import Script

The server includes a utility script for importing licenses via the HTTP API. This script handles authentication, license key reading, and activation in a single command.

### Script Location

`server/scripts/import_license.sh`

### Prerequisites

Before using the import script:
1. Server must be running and accessible
2. An admin/manager user account must exist (set up during initial installation)
3. License file (.pem) must be available

### Usage

```bash
# Via command-line arguments
./scripts/import_license.sh admin@example.com password /path/to/license.pem

# Via environment variables
export LICENSE_EMAIL="admin@example.com"
export LICENSE_PASSWORD="password"
export LICENSE_FILE="/path/to/license.pem"
./scripts/import_license.sh

# With custom server URL
./scripts/import_license.sh admin@example.com password /path/to/license.pem --server-url https://api.example.com

# Force overwrite without confirmation (when existing valid license exists)
./scripts/import_license.sh admin@example.com password /path/to/license.pem --force
```

### Arguments

| Argument       | Description                                          |
| -------------- | ---------------------------------------------------- |
| `EMAIL`        | Admin/manager user email address                     |
| `PASSWORD`     | User password                                        |
| `LICENSE_FILE` | Path to license file (.pem format)                   |
| `--server-url` | Optional server URL (default: http://localhost:8080) |
| `--force`      | Skip confirmation when overwriting a valid license  |

### Environment Variables

| Variable           | Description                                          |
| ------------------ | ---------------------------------------------------- |
| `LICENSE_EMAIL`    | User email (alternative to first argument)           |
| `LICENSE_PASSWORD` | User password (alternative to second argument)       |
| `LICENSE_FILE`     | Path to license file (alternative to third argument) |
| `SERVER_URL`       | Server URL (alternative to --server-url)             |

### How It Works

1. **Authentication**: Logs in via `POST /auth/login` to obtain JWT token
2. **Current License Check**: Fetches current license status via `GET /api/v1/license`
   - If a valid license exists with days remaining, displays current license details
   - Prompts for confirmation before proceeding (unless `--force` is specified)
3. **License Reading**: Reads license key from specified file
4. **Activation**: Calls `POST /api/v1/license/activate` with JWT token and license key
5. **Verification**: Displays success/error messages with colored output

### License Validation Behavior

Before activating a new license, the script checks the current license status:

- **No active license**: Proceeds directly to activation
- **Expired license**: Proceeds directly to activation with a message
- **Valid license detected**: Displays current license details and prompts for confirmation
  - Shows: Customer name, Tier, Expiration date, Days remaining
  - Asks: "Do you want to proceed? (y/N)"
  - Use `--force` flag to skip this confirmation and proceed automatically

### Requirements

- `curl` - For making HTTP requests
- `jq` - Optional, for JSON parsing (script has fallback to grep/sed)

### Example Workflow

```bash
# 1. Ensure server is running
make server-dev

# 2. Import the license (using existing admin account)
cd server/scripts
./import_license.sh admin@example.com adminpass /path/to/license.pem

# If a valid license exists, you'll see:
# ℹ Current License Status:
#   Customer: Acme Corp
#   Tier: Enterprise
#   Expires: 2027-02-03T17:30:37Z
#   Days Remaining: 364
# WARNING: You are about to overwrite a valid license!
# Do you want to proceed? (y/N):

# 3. Verify license was activated
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"adminpass"}' | jq -r '.access_token')

curl -s http://localhost:8080/api/v1/license \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Troubleshooting

#### "Login failed"
- Verify email and password are correct
- Ensure admin user exists (check with `/auth/verify` endpoint)
- Confirm server is accessible at specified URL

#### "License file not found"
- Verify file path is correct
- Ensure file has `.pem` extension
- Check file permissions

#### "License activation failed"
- Verify license key format (should start with `-----BEGIN LICENSE KEY-----`)
- Check license hasn't expired
- Ensure `LICENSE_PUBLIC_KEY` environment variable is set correctly on server
- Verify license was signed with matching private key

#### Script shows "unexpected response"
- Check if jq is installed: `command -v jq`
- If jq missing, script falls back to grep/sed
- Enable debug mode: Add `set -x` after line 3 in script

## Sample License

```
# Generate keypair (one-time setup)
go run scripts/keygen/main.go -genkeys

# Get the public key (for LICENSE_PUBLIC_KEY env):
cat public.key | base64

# Generate a license (license ID is auto-generated as UUID)
go run scripts/keygen/main.go -customer daniel -days 365 -seats 100 -tier 3.0 > license.pem

# The script outputs the license ID to stderr:
# License ID: 550e8400-e29b-41d4-a716-446655440000

# Import the license:
./scripts/import_license.sh admin@example.com adminpass /path/to/license.pem
```

## Additional Resources

- [go-license library](https://github.com/vitalvas/go-license) - License cryptography
- [Ed25519 specification](https://ed25519.cr.yp.to/) - Signature algorithm details
- `README.md` - Quick reference for license endpoints
- `CLAUDE.md` - Developer patterns for license system
