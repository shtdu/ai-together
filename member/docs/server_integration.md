# Server Integration Guide

This guide covers the HTTP client architecture used for server communication in the member client.

## Overview

The member client uses a type-safe, generated HTTP client for all server interactions. This architecture:
- Eliminates runtime type errors
- Provides centralized logging, authentication, and retry logic
- Ensures API contracts match the OpenAPI specification

## OpenAPI Code Generation

All server API types and client code are generated from the OpenAPI specification:

| Item | Location |
|------|----------|
| **Source spec** | `docs/client_api/server_api.yaml` (member client specific) |
| **Generated output** | `shared/integration/generated.go` (~3,172 lines) |
| **Code generator** | oapi-codegen v2.5.1 |

The member client API spec is a filtered subset of the full server API spec (`docs/phase_1/server_api.yaml`) optimized for member client integration. It excludes:
- Manager-only endpoints (team write operations, proxy endpoints)
- Unused schemas (Invitation, configuration sync types)

### Regenerating the Client

```bash
make shared-api-gen    # From repository root
cd shared/integration && go generate  # Direct
```

This regenerates all types and client methods from the OpenAPI spec. The generated code is committed to git for stability.

## HTTP Transport Middleware Chain

The client uses a layered transport middleware pattern:

```
┌─────────────────────────────────────┐
│   RetryTransport (exponential backoff) │
├─────────────────────────────────────┤
│   AuthTransport (JWT injection)      │
├─────────────────────────────────────┤
│   LoggingTransport (detailed logs)   │
├─────────────────────────────────────┤
│   http.DefaultTransport              │
└─────────────────────────────────────┘
```

### Transport Components

**1. RetryTransport** (`internal/api/transport.go:203-274`)
- Retries failed requests with exponential backoff
- Retries on: 408, 429, 500, 502, 503, 504
- Max retries: 3
- Backoff: 500ms × attempt number

**2. AuthTransport** (`internal/api/transport.go:164-201`)
- Injects JWT tokens via `Authorization: Bearer <token>` header
- Token obtained via `TokenGetter` function
- Skips auth if token getter is nil

**3. LoggingTransport** (`shared/integration/transport.go:17-162`)
- Logs all requests and responses with structured logging (log/slog)
- Captures and prettifies JSON bodies
- Sanitizes auth tokens in logs
- Outputs bodies directly to stderr (unescaped)

## Client Factory Functions

The `shared/integration/client.go` package provides factory functions for creating configured clients:

### Anonymous Client

Used for auth operations (login, register, refresh token). No JWT token required.

```go
import "github.com/code-together/shared/integration"

client, err := integration.NewAnonymousClientImpl(serverURL, logger)
```

### Authenticated Client

Used for all other operations. JWT token automatically injected via `AuthTransport`.

```go
getToken := func() (string, error) {
    return authService.GetAccessToken()
}
client, err := integration.NewAuthenticatedClientImpl(serverURL, getToken, logger)
```

Both functions return `integration.ClientWithResponsesInterface`, which can be mocked for testing.

## Service Integration Pattern

All server integration services follow the same pattern:

1. **Accept `ClientWithResponsesInterface` in constructor**
2. **Use generated request/response types**
3. **Handle pointer types for optional fields**

### Example: UsageSyncService

```go
type UsageSyncService struct {
    authService *AuthService
    apiClient   api.ClientWithResponsesInterface  // Generated client interface
}

func NewUsageSyncService(authService *AuthService, apiClient api.ClientWithResponsesInterface) *UsageSyncService {
    return &UsageSyncService{
        authService: authService,
        apiClient:   apiClient,
    }
}

func (us *UsageSyncService) SyncUsageStats() (int, error) {
    // Use generated client method
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    resp, err := us.apiClient.PostApiV1UsageBatchWithResponse(ctx, serverRecords)
    if err != nil {
        return 0, fmt.Errorf("failed to send request: %w", err)
    }

    if resp.StatusCode() != http.StatusOK {
        return 0, fmt.Errorf("server returned status %d", resp.StatusCode())
    }

    // Use generated response type
    batchResp := resp.JSON200
    // ... process response
}
```

## Type Conversion Patterns

### Email Type

The OpenAPI generator uses `openapi_types.Email` for email fields. Convert between string and Email:

```go
import openapi_types "github.com/oapi-codegen/runtime/types"

// Request: string → Email
resp, err := client.PostAuthLoginWithResponse(ctx, api.PostAuthLoginJSONRequestBody{
    Email:    openapi_types.Email(email),  // Convert string to Email
    Password: password,
})

// Response: Email → string
user := &User{
    Email: string(userResp.User.Email),  // Convert Email to string
}
```

### Optional Fields (Pointer Types)

Generated types use pointers for optional fields. Only set them if non-zero:

```go
// Required fields - always set
record := api.UsageRecord{
    Platform:  log.Platform,
    Model:     log.Model,
    TenantId:  &tenantID,    // Use & for pointer types
    UserId:    &userID,
}

// Optional fields - only set if non-zero
if log.InputTokens > 0 {
    record.InputTokens = &log.InputTokens
}
if log.OutputTokens > 0 {
    record.OutputTokens = &log.OutputTokens
}
```

### Response with Optional Fields

Check optional fields before dereferencing:

```go
userResp := resp.JSON200
user := &User{
    ID:        userResp.User.Id,
    Email:     string(userResp.User.Email),
    Name:      userResp.User.Name,
    UpdatedAt: time.Time{},  // Default value
}

if userResp.User.UpdatedAt != nil {  // Check optional field
    user.UpdatedAt = *userResp.User.UpdatedAt  // Dereference
}
```

## Dual Client Pattern

Services that need both anonymous and authenticated access use the dual client pattern:

| Client Type | Use Case | Endpoints | Token Required |
|-------------|----------|-----------|----------------|
| **Anonymous** | AuthService | Login, register, refresh token | No |
| **Authenticated** | Other services | Config sync, usage sync, profile | Yes (auto-injected) |

### Implementation in main.go

```go
// Get server URL for client creation
serverURL, err := serverConfigService.GetServerURL()
if err != nil {
    log.Printf("Warning: failed to get server URL: %v", err)
    serverURL = ""
}

var authService *services.AuthService
var configSyncService *services.ConfigSyncService
var usageSyncService *services.UsageSyncService

// Only create server integration services if server URL is configured
if serverURL != "" {
    // Create anonymous API client for auth operations
    anonClient, err := api.NewAnonymousClientImpl(serverURL, logger)
    if err != nil {
        log.Printf("Warning: failed to create anonymous API client: %v", err)
    } else {
        // Create auth service with anonymous client
        authService = services.NewAuthService(anonClient, serverConfigService)

        // Create authenticated API client for other operations
        authClient, err := api.NewAuthenticatedClientImpl(serverURL, authService.GetAccessToken, logger)
        if err != nil {
            log.Printf("Warning: failed to create authenticated API client: %v", err)
        } else {
            // Create other services that need the authenticated client
            configSyncService = services.NewConfigSyncService(authService, authClient, providerService, importService)
            usageSyncService = services.NewUsageSyncService(authService, authClient)
        }
    }
}

// Add services to app if they were created
if authService != nil {
    servicesList = append(servicesList, application.NewService(authService))
}
if configSyncService != nil {
    servicesList = append(servicesList, application.NewService(configSyncService))
}
```

This pattern ensures:
1. The app works offline without server integration
2. Server connection failures don't crash the app
3. Services are only registered if successfully initialized

## Method Names After API Path Changes

When the OpenAPI spec paths change, oapi-codegen generates different method names. For example:
- Path `/user/profile` → Method `GetUserProfileWithResponse()`
- Path `/api/v1/user/profile` → Method `GetApiV1UserProfileWithResponse()`

After updating OpenAPI paths:
1. Regenerate the client: `cd member/internal/api && go generate`
2. Update all service files to use new method names
3. Rebuild and test

## Testing

The transport middleware has comprehensive tests (`internal/api/transport_test.go`):

```bash
cd member
go test -v ./internal/api/...
```

| Test | Purpose |
|------|---------|
| `TestLoggingTransport` | Verifies request/response logging |
| `TestAuthTransport` | Verifies JWT token injection |
| `TestRetryTransport_Success` | Verifies normal request flow |
| `TestRetryTransport_RetryOn500` | Verifies retry logic |
| `TestTransportChain` | Verifies transport chain composition |
| `TestNewAnonymousClient` | Verifies anonymous client creation |
| `TestNewAuthenticatedClient` | Verifies authenticated client creation |
| `TestPrettifyJSON_*` | Verifies JSON prettification |

All tests pass.

## Common Tasks

### Adding a New Server Integration Service

1. **Create service** that accepts `ClientWithResponsesInterface`:
   ```go
   type MyService struct {
       apiClient api.ClientWithResponsesInterface
   }
   ```

2. **Use generated types** for requests and responses

3. **Register in main.go** using the dual client pattern

4. **Test** with a mock implementation of `ClientWithResponsesInterface`

### Updating OpenAPI Spec

The member client uses its own API spec optimized for client integration:

1. For member client specific changes: Edit `docs/client_api/server_api.yaml`
2. For full API changes: Update `docs/phase_1/server_api.yaml`, then filter changes to `docs/client_api/server_api.yaml`
3. Regenerate client: `cd member/internal/api && go generate`
4. Update service files with new method names
5. Run tests: `go test ./...`

## Key Files

| File | Purpose |
|------|---------|
| `docs/client_api/server_api.yaml` | Member client OpenAPI spec (source of truth) |
| `internal/api/doc.go` | OpenAPI code generation directive (`go:generate`) |
| `internal/api/generated.go` | Generated types and client (from OpenAPI spec) |
| `internal/api/transport.go` | HTTP middleware (retry, auth, logging) |
| `internal/api/client.go` | Client factory functions |
| `internal/api/transport_test.go` | Transport middleware tests |
| `services/authservice.go` | JWT authentication (uses anonymous client) |
| `services/configsyncservice.go` | Config sync (uses authenticated client) |
| `services/usagesyncservice.go` | Usage sync (uses authenticated client) |
| `services/serverconfigservice.go` | Server settings (uses authenticated client) |
| `main.go` | Dependency wiring for both client types |
