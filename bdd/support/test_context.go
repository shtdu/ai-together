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
	LastProviderID    int64
	LastUserID        string
	LastTeamID        int64
	LastLicenseID     string
	CreatedResourceIDs map[string]string

	// Response storage (per-scenario, for assertions)
	LastStatusCode    int
	LastResponse      interface{}
	LastErrorResponse string

	// Resource tracking for cleanup
	createdProviders []int64
	createdUsers      []string
	createdTeams      []int64
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
