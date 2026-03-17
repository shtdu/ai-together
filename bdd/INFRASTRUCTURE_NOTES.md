# BDD API Infrastructure - Implementation Notes

## Completed (Phase 1)

### Client Infrastructure
- [x] **Client factory created** (`support/client_factory.go`)
  - `NewAnonymousClient()` - Creates unauthenticated API client
  - `NewAuthenticatedClient()` - Creates client with automatic token injection

### Test Context
- [x] **Test context updated** (`support/test_context.go`)
  - Client types changed from `interface{}` to `*integration.ClientWithResponses`
  - Added `Logger` field for client operations
  - Client management methods:
    - `InitializeClients()` - One-time client setup
    - `GetAnonymousClient()` - Access anonymous client
    - `GetAuthToken()` - Retrieve authentication token
    - `GetAuthenticatedClient()` - Get/reuse authenticated client
    - `UpdateAuthenticatedClients()` - Refresh after login/logout

### Data Management
- [x] **Fixture loading from JSON** (`support/fixtures.go` + `fixtures.json`)
  - `LoadFixtureData()` - Loads from local fixtures.json
  - `GetStandardFixtureData()` - JSON-only loading (no hardcoded defaults)
  - Fixtures: 3 users, 3 providers, 1 team, 3 licenses

### API Helpers
- [x] **Response handling** (`support/api_helpers.go`)
  - `handleAPIResponse()` - Consistent API error handling
  - `IsTestServerRunning()` - Server health check (from `server_lifecycle.go`)

### Test Lifecycle
- [x] **Resource cleanup** (`step_definitions/context.go`)
  - `CleanupScenarioResources()` - Logs and clears tracked resources
  - TODO: API deletion will be added during step conversion

- [x] **Test suite hooks** (`godog/godog_suite_test.go`)
  - API clients initialized during suite setup
  - Test server health verification before scenarios
  - Fixture data loading in BeforeScenario hook
  - Cleanup call in AfterScenario hook

## Completed (Phase 2)

### Health Check Steps
- [x] **Converted to real API calls** (`step_definitions/common_steps.go`)
  - `iCheckTheHealthEndpoint()` - Makes real GET request to `/health` endpoint
  - `theSystemShouldBeHealthy()` - Validates actual response status code
  - Robust response parsing (JSON200, JSON503, fallback)
  - Error handling with descriptive messages

### Auth Login Steps
- [x] **Converted to real API calls** (`step_definitions/auth_steps.go`)
  - `iLoginWithCredentials()` - Makes real POST request to `/auth/login` endpoint
  - Token extraction and storage (AdminToken/MemberToken based on role)
  - User profile parsing and storage in CurrentUser
  - Client refresh via `UpdateAuthenticatedClients()`
  - Comprehensive error handling (400, 401, 403 responses)

- [x] **Helper steps updated**:
  - `iAmLoggedInAsAManager()` - Loads fixtures, gets admin credentials, calls login
  - `iAmLoggedInAsAMember()` - Loads fixtures, gets member credentials, calls login

- [x] **Fixtures path resolution** (`support/fixtures.go`)
  - Fixed to use `runtime.Caller()` for absolute path resolution
  - Works correctly in worktree structure

### Test Validation
- [x] **Tests execute and make real API calls**
  - Verified via `./bdd-test.sh`
  - Logs show real HTTP requests to `/auth/login`
  - API responses captured correctly (401 for non-existent users, 200 for successful)
  - Fixtures loading correctly (3 users, 3 providers, 1 team, 3 licenses)

## Completed (Phase 3)

### User Creation Steps
- [x] **Implemented user creation via manager API** (`step_definitions/auth_steps.go`)
  - `userExists()` - Creates users via `/api/v1/users` endpoint (not `/auth/register`)
  - Admin authentication to get manager token
  - Authenticated manager client for user creation
  - Special handling for admin/manager users created by test server setup
  - Name extraction from email (user@example.com → User Example)
  - Resource tracking for cleanup (user ID conversion to string)
  - Comprehensive error handling (400, 401, 403, 409 responses)
  - Handles 500 duplicate key errors as success (user already exists)

### Test Validation
- [x] **Login scenarios now passing end-to-end**
  - "Login with valid credentials as manager" - ✅ Passing
  - "Login with valid credentials as member" - ✅ Passing
  - Users created dynamically before login attempts
  - Admin user verification via login attempt

## Completed (Phase 4)

### Profile Operations
- [x] **Converted profile retrieval to real API calls** (`step_definitions/auth_steps.go`)
  - `iGetUserProfile()` - Uses `GET /api/v1/user/profile` endpoint
  - Authenticated client with token injection
  - Handles JSON200/JSON401 responses
  - Error handling with descriptive messages

### User Management Improvements
- [x] **Ensure users exist before login** (`step_definitions/auth_steps.go`)
  - `iAmLoggedInAsAManager()` calls `userExists()` first
  - `iAmLoggedInAsAMember()` calls `userExists()` first
  - Prevents 401 errors when users don't exist in database

### Test Results
- **Overall:** 215 scenarios, 199 passed (92.5%), 16 failed
- **Passing areas:**
  - ✅ Health checks
  - ✅ Authentication and login
  - ✅ User creation via manager API
  - ✅ Profile retrieval
- **Failing areas** (not yet converted to real API calls):
  - ❌ Provider management (403 errors)
  - ❌ License activation (403 errors)
  - ❌ Dashboard operations (403 errors)

## Completed (Phase 5)

### Provider Management - Core CRUD
- [x] **Converted provider creation to real API calls** (`step_definitions/provider_steps.go`)
  - `iCreateAProviderWithKindAndAPIKey()` - Uses `POST /api/v1/providers` endpoint
  - Requires: name, kind, api_key, api_url (all required fields)
  - Provider kinds: claude, codex, opencode
  - API URLs mapped per kind:
    - claude → https://api.anthropic.com
    - codex → https://api.github.com
    - opencode → https://api.opencode.com
  - Handles JSON201/JSON400/JSON401/JSON403 responses
  - Tracks created providers for cleanup

- [x] **Converted provider listing to real API calls**
  - `iListAllProvidersFromProvider()` - Uses `GET /api/v1/providers` endpoint
  - Handles JSON200/JSON401 responses

- [x] **Converted provider retrieval to real API calls**
  - `iGetProviderByID()` - Uses `GET /api/v1/providers/{id}` endpoint
  - Handles JSON200/JSON401/JSON404 responses

- [x] **Converted provider deletion to real API calls**
  - `iDeleteProvider()` - Uses `DELETE /api/v1/providers/{id}` endpoint
  - Handles JSON200/JSON401/JSON404 responses

- [x] **Converted provider update to real API calls**
  - `iUpdateProviderName()` - Uses `PUT /api/v1/providers/{id}` endpoint
  - Handles JSON200/JSON400/JSON401/JSON404 responses

### Testing Improvements
- [x] **Added @wip tags for focused testing**
  - Tag work-in-progress scenarios with `@wip`
  - Test only WIP scenarios: `./bdd-test.sh --tags "@wip"`
  - Faster iteration during development

### Test Validation
- [x] **Provider creation scenarios passing**
  - Create provider with valid data (claude) - ✅ Passing
  - Create provider with valid data (codex) - ✅ Passing
  - Create provider with valid data (opencode) - ✅ Passing
  - All use real API calls to create providers

### Test Results
- **@wip scenarios:** 3/3 passing (provider creation)
- **Overall:** 215 scenarios, 190 passed (88%), 25 failed
- **Note:** Some scenarios now fail because they relied on mock behavior
  - Need to decide: convert to real API or keep mocks for specific validation tests

## Completed (Phase 6)

### License Management - Core Operations
- [x] **Converted license activation to real API calls** (`step_definitions/license_steps.go`)
  - `iActivateTheLicense()` - Uses `POST /api/v1/license/activate` endpoint
  - Requires authenticated client (admin/manager)
  - Handles JSON200/JSON400/JSON401/JSON403 responses
  - **Validates license keys** - Correctly rejects malformed/invalid keys
  - Error messages: "malformed license", "invalid signature", etc.

- [x] **Converted license information retrieval to real API calls**
  - `iGetLicenseInformation()` - Uses `GET /api/v1/license` endpoint
  - Handles JSON200/JSON401/JSON500 responses
  - Returns real license status from database

### Test Validation
- [x] **License activation API working correctly**
  - API calls are being made successfully ✅
  - License validation is enforced ✅
  - Invalid keys are properly rejected ✅
  - Returns correct HTTP status codes (400 for invalid keys)

### Test Results
- **@wip scenarios:** 2/2 fail with "malformed license" (expected)
- **Cause:** Fixture license keys are fake, not real licenses
- **This is correct behavior:** API should validate license format
- **Note:** To make tests pass, need valid test license keys or test helper endpoint

## Next Steps (Phase 7+)

### Convert Usage Analytics
- [ ] Usage statistics (`usage_steps.go`)
- [ ] Team usage analytics
- [ ] Personal usage tracking
- [ ] Cost calculation endpoints

### Convert Dashboard Operations
- [ ] Dashboard summaries (`dashboard_steps.go`)
- [ ] Team statistics
- [ ] Team management (create, update, delete)

### Convert Additional Provider Operations
- [ ] Provider enable/disable (still mock)
- [ ] Provider connectivity testing (still mock)
- [ ] Provider statistics (still mock)
- [ ] Provider filtering by kind (still mock)

### Optional: Test License Management
- [ ] Create test license generation endpoint
- [ ] Generate valid test license keys for BDD scenarios
- [ ] Or use real trial license keys for testing

## Known Issues

### Resource Cleanup
**Status:** Infrastructure ready, API deletion pending
- Current cleanup logs resources and clears tracking lists
- Actual API deletion will be implemented during step conversion
- This allows step definitions to verify resources exist before deletion
- Retry logic will be added when API delete methods are confirmed

### Client Factory
**Status:** Working correctly
- Uses `slog.Default()` internally (logger parameter not needed)
- Client initialization verified: anonymous and authenticated clients created successfully

### Authentication Flow
**Status:** ✅ Complete
- Token tracking fields exist (`AdminToken`, `MemberToken`)
- Client management supports token refresh via `UpdateAuthenticatedClients()`
- Login steps call `UpdateAuthenticatedClients()` after successful auth
- User creation uses manager endpoint (not `/auth/register` which is intentionally hidden)

### Module Structure
**Status:** ✅ Clarified
- `integration` package: `/auth/login`, `/health` endpoints (anonymous access)
- `integration_manager` package: `/api/v1/users`, `/api/v1/providers`, etc. (requires authentication)
- go.mod uses local replace directives for both packages
- User creation requires admin login → manager token → authenticated manager client

## Infrastructure Verification

✅ **Compilation:** All packages compile without errors
```bash
go build ./support/...
go build ./step_definitions/...
go build ./godog/...
```

✅ **Test Server:** Running and healthy at http://localhost:8088
```bash
curl http://localhost:8088/health
# Returns: {"version":"v0.2.0","status":"ok",...}
```

✅ **Fixtures:** Load correctly from JSON
```bash
go test -v -run TestLoadFixtureData
# Result: 4 PASS, 1 SKIP (expected)
```

✅ **Client Factory:** Creates valid API clients
```go
// Verified in test_context_test.go
client, err := NewAnonymousClient("http://localhost:8088", false)
// Result: Client created successfully
```

## Implementation Commits

### Phase 1 - Infrastructure
1. `20fb530` feat: add API client factory for BDD tests
2. `9ce884d` refactor: load fixtures from JSON only, remove hardcoded defaults
3. `dd4e60d` feat: add API response handling helpers
4. `b5a5af4` test: add comprehensive tests for handleAPIResponse
5. `4e2c87f` refactor: update test context client types to use real API clients
6. `4e5f2db` feat: add client management methods to test context
7. `30dc597` feat: implement resource cleanup infrastructure
8. `ed703f2` feat: initialize API clients in test suite setup
9. `0cb835c` docs: add infrastructure implementation notes

### Phase 2 - Smoke Test Conversion
10. `53fd616` feat: convert health check steps to use real API calls
11. `5238ace` feat: convert auth login steps to use real API calls
12. `797ba4a` fix: resolve fixture data path for worktree structure

### Phase 3 - User Creation
13. `33040fc` fix: resolve fixture data path for worktree structure
14. `786b218` test: achieve 100% BDD test success rate (215/215 scenarios)
15. `688d2f5` fix: start test server on port 8088 instead of 9080
16. `f817fa0` feat: implement missing BDD step definitions for usage analytics
17. `cdeb928` fix: resolve function signature mismatches and validation issues
18. `a8a6fa6` feat: implement user creation step via manager API

## Design Decisions

### Logger Strategy
- Client factory uses `slog.Default()` internally
- Logger stored in `BDDTestContext.Logger` for future use
- Keeps client factory signatures simple

### Cleanup Approach
- Infrastructure logs and tracks resources without API deletion
- Deletion will be added during step conversion when API methods are confirmed
- Prevents premature implementation that might block infrastructure

### Client Lifecycle
- Anonymous client created once during suite initialization
- Authenticated clients created per-scenario after login
- Token refresh via `UpdateAuthenticatedClients()` after login/logout

## Testing Strategy

### Unit Tests
- ✅ Fixture loading (4 test cases)
- ✅ API response handling (5 test cases)
- ✅ Client initialization (1 test case)

### Integration Tests
- ⏳ Smoke test conversion (next phase)
- ⏳ Step definition API conversion (subsequent phases)

### Test Execution
```bash
# Run all BDD tests
./bdd-test.sh

# Run smoke tests only
./bdd-test.sh --tags "@smoke"

# Run with specific format
./bdd-test.sh --format junit
```

---

**Last Updated:** 2025-03-17
**Status:** Phase 3 complete (user creation implemented). 199/215 scenarios passing (92.5%). Login and authentication scenarios fully working. Ready for Phase 4 (profile, provider, license, and dashboard step conversions).
