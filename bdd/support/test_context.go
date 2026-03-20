// Package support provides test utilities and context management for BDD tests
package support

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/code-together/shared/integration"
)

// BDDTestContext holds shared state across BDD scenarios.
// IMPORTANT: Not thread-safe for parallel scenario execution.
// Each scenario should have its own context instance.
// The mutex protects concurrent access within a scenario.
type BDDTestContext struct {
	mu sync.Mutex // Protects all fields below

	// Server connection (immutable after setup)
	ServerURL string
	TestDBURL string

	// API clients (immutable after setup)
	AnonymousClient *integration.ClientWithResponses
	Client          *integration.ClientWithResponses
	ManagerClient   *integration.ClientWithResponses

	// Logger for client creation and operations
	Logger *slog.Logger

	// Authentication (per-scenario)
	AdminToken  string
	MemberToken string
	CurrentUser *UserInfo

	// Test data storage (per-scenario, for sharing between steps)
	LastProviderID     int64
	LastUserID         string
	LastTeamID         int64
	LastLicenseID      string
	CreatedResourceIDs map[string]string

	// Response storage (per-scenario, for assertions)
	LastStatusCode    int
	LastResponse      interface{}
	LastErrorResponse string

	// Resource tracking for cleanup
	createdProviders []int64
	createdUsers     []string
	createdTeams     []int64
}

// UserInfo represents user information for scenarios
type UserInfo struct {
	Email    string
	Password string
	Name     string
	Role     string // "admin" or "member"
	Token    string
}

// SetLastResponse stores the last response in a thread-safe manner
func (ctx *BDDTestContext) SetLastResponse(code int, resp interface{}, errMsg string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.LastStatusCode = code
	ctx.LastResponse = resp
	ctx.LastErrorResponse = errMsg
}

// GetLastResponse retrieves the last response in a thread-safe manner
func (ctx *BDDTestContext) GetLastResponse() (int, interface{}, string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.LastStatusCode, ctx.LastResponse, ctx.LastErrorResponse
}

// TrackCreatedResource adds a resource to the cleanup map
func (ctx *BDDTestContext) TrackCreatedResource(resourceType, id string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.CreatedResourceIDs == nil {
		ctx.CreatedResourceIDs = make(map[string]string)
	}
	ctx.CreatedResourceIDs[resourceType] = id
}

// GetCreatedResource retrieves a tracked resource ID
func (ctx *BDDTestContext) GetCreatedResource(resourceType string) (string, bool) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	id, ok := ctx.CreatedResourceIDs[resourceType]
	return id, ok
}

// TrackProvider adds a provider to the cleanup list
func (ctx *BDDTestContext) TrackProvider(providerID int64) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.createdProviders = append(ctx.createdProviders, providerID)
}

// TrackUser adds a user to the cleanup list
func (ctx *BDDTestContext) TrackUser(userID string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.createdUsers = append(ctx.createdUsers, userID)
}

// TrackTeam adds a team to the cleanup list
func (ctx *BDDTestContext) TrackTeam(teamID int64) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.createdTeams = append(ctx.createdTeams, teamID)
}

// GetCreatedProviders returns a copy of the tracked provider IDs
func (ctx *BDDTestContext) GetCreatedProviders() []int64 {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	providers := make([]int64, len(ctx.createdProviders))
	copy(providers, ctx.createdProviders)
	return providers
}

// GetCreatedUsers returns a copy of the tracked user IDs
func (ctx *BDDTestContext) GetCreatedUsers() []string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	users := make([]string, len(ctx.createdUsers))
	copy(users, ctx.createdUsers)
	return users
}

// GetCreatedTeams returns a copy of the tracked team IDs
func (ctx *BDDTestContext) GetCreatedTeams() []int64 {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	teams := make([]int64, len(ctx.createdTeams))
	copy(teams, ctx.createdTeams)
	return teams
}

// ClearCreatedResources clears all tracked resource lists
func (ctx *BDDTestContext) ClearCreatedResources() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.createdProviders = nil
	ctx.createdUsers = nil
	ctx.createdTeams = nil
	ctx.CreatedResourceIDs = make(map[string]string)
}

// Reset clears all scenario state
func (ctx *BDDTestContext) Reset() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Reset response data
	ctx.LastStatusCode = 0
	ctx.LastResponse = nil
	ctx.LastErrorResponse = ""

	// Reset test data
	ctx.LastProviderID = 0
	ctx.LastUserID = ""
	ctx.LastTeamID = 0
	ctx.LastLicenseID = ""

	// Reset resource tracking
	ctx.createdProviders = nil
	ctx.createdUsers = nil
	ctx.createdTeams = nil
	ctx.CreatedResourceIDs = nil

	// Reset authentication
	ctx.AdminToken = ""
	ctx.MemberToken = ""
	ctx.CurrentUser = nil
}

// String returns a string representation of the context (for debugging)
func (ctx *BDDTestContext) String() string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	return fmt.Sprintf(
		"BDDTestContext{ServerURL: %s, LastStatusCode: %d, CreatedProviders: %d, CreatedUsers: %d, CreatedTeams: %d}",
		ctx.ServerURL,
		ctx.LastStatusCode,
		len(ctx.createdProviders),
		len(ctx.createdUsers),
		len(ctx.createdTeams),
	)
}

// InitializeClients sets up all API clients for testing
// Called once during test suite initialization (not per scenario)
func (ctx *BDDTestContext) InitializeClients(serverURL string, logger *slog.Logger) error {
	// Store logger for client creation
	ctx.Logger = logger

	// Create anonymous client (no authentication required)
	anonClient, err := NewAnonymousClient(serverURL, false)
	if err != nil {
		return fmt.Errorf("failed to create anonymous client: %w", err)
	}
	ctx.AnonymousClient = anonClient

	// Authenticated clients will be created per-scenario after user login
	ctx.Client = nil
	ctx.ManagerClient = nil

	return nil
}

// GetAnonymousClient returns the anonymous client for unauthenticated requests
func (ctx *BDDTestContext) GetAnonymousClient() *integration.ClientWithResponses {
	return ctx.AnonymousClient
}

// GetAuthToken returns the current authentication token for the scenario
// Checks AdminToken first, then MemberToken
func (ctx *BDDTestContext) GetAuthToken() (string, error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.AdminToken != "" {
		return ctx.AdminToken, nil
	}
	if ctx.MemberToken != "" {
		return ctx.MemberToken, nil
	}
	return "", fmt.Errorf("no authentication token available")
}

// GetAuthenticatedClient returns a client with automatic token injection
// Creates or reuses the authenticated client for the current scenario
// Uses ManagerClient for admin/manager operations, Client for member operations
func (ctx *BDDTestContext) GetAuthenticatedClient() (*integration.ClientWithResponses, error) {
	// Get current token
	_, err := ctx.GetAuthToken()
	if err != nil {
		return nil, fmt.Errorf("cannot create authenticated client without token: %w", err)
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Create or reuse authenticated client based on user role
	// Use ManagerClient for admin/manager roles, Client for member roles
	if ctx.CurrentUser != nil && (ctx.CurrentUser.Role == "admin" || ctx.CurrentUser.Role == "manager") {
		if ctx.ManagerClient == nil {
			client, err := NewAuthenticatedClient(ctx.ServerURL, ctx.GetAuthToken, false)
			if err != nil {
				return nil, err
			}
			ctx.ManagerClient = client
		}
		return ctx.ManagerClient, nil
	}

	// Use Client for non-admin/non-manager users
	if ctx.Client == nil {
		client, err := NewAuthenticatedClient(ctx.ServerURL, ctx.GetAuthToken, false)
		if err != nil {
			return nil, err
		}
		ctx.Client = client
	}
	return ctx.Client, nil
}

// UpdateAuthenticatedClients refreshes the authenticated clients with a new token
// Called after login/logout to update token injection
func (ctx *BDDTestContext) UpdateAuthenticatedClients(token string) error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Create token getter that returns the provided token
	tokenGetter := func() (string, error) {
		return token, nil
	}

	client, err := NewAuthenticatedClient(ctx.ServerURL, tokenGetter, false)
	if err != nil {
		return err
	}

	// Use ManagerClient for admin/manager roles, Client for member roles
	if ctx.CurrentUser != nil && (ctx.CurrentUser.Role == "admin" || ctx.CurrentUser.Role == "manager") {
		ctx.ManagerClient = client
	} else {
		ctx.Client = client
	}
	return nil
}

// SetTestCredentials stores test credentials for scenario use
// Used by lockout and other auth testing scenarios that need to reuse credentials
func (ctx *BDDTestContext) SetTestCredentials(email, password string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.CreatedResourceIDs == nil {
		ctx.CreatedResourceIDs = make(map[string]string)
	}
	ctx.CreatedResourceIDs["test_email"] = email
	ctx.CreatedResourceIDs["test_password"] = password
}

// GetTestCredentials retrieves stored test credentials
// Returns empty strings if not set
func (ctx *BDDTestContext) GetTestCredentials() (email, password string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.CreatedResourceIDs == nil {
		return "", ""
	}
	email, _ = ctx.CreatedResourceIDs["test_email"]
	password, _ = ctx.CreatedResourceIDs["test_password"]
	return email, password
}

// GetCreatedResources returns a copy of all tracked resources
// Used for verification and scenario flow validation
func (ctx *BDDTestContext) GetCreatedResources() map[string]string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	resources := make(map[string]string)
	for k, v := range ctx.CreatedResourceIDs {
		resources[k] = v
	}
	return resources
}
