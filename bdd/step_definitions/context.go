// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"
	"log"

	"github.com/code-together/bdd/support"
	integration_manager "github.com/code-together/integration_manager"
)

// ScenarioContext wraps BDDTestContext with scenario-specific methods
type ScenarioContext struct {
	*support.BDDTestContext
	ManagerClient *integration_manager.ClientWithResponses // For manager-specific operations
}

// ResetScenarioState resets all scenario state before execution
func (ctx *ScenarioContext) ResetScenarioState() {
	ctx.Reset()
	ctx.ManagerClient = nil // Clear manager client so it gets recreated with new token
	log.Printf("Scenario state reset")
}

// CleanupScenarioResources cleans up all resources created during the scenario
// TODO: Will delete resources via API during step definition conversion
func (ctx *ScenarioContext) CleanupScenarioResources() error {
	log.Printf("Cleaning up scenario resources...")

	// Log resources that would be cleaned up
	teams := ctx.GetCreatedTeams()
	providers := ctx.GetCreatedProviders()
	users := ctx.GetCreatedUsers()

	if len(teams) > 0 {
		log.Printf("Would clean up %d teams: %v", len(teams), teams)
	}
	if len(providers) > 0 {
		log.Printf("Would clean up %d providers: %v", len(providers), providers)
	}
	if len(users) > 0 {
		log.Printf("Would clean up %d users: %v", len(users), users)
	}

	// Clear resource lists
	ctx.ClearCreatedResources()

	log.Printf("Cleanup complete (API deletion will be added during step conversion)")
	return nil
}

// LoginAsManager logs in as a manager user
func (ctx *ScenarioContext) LoginAsManager() error {
	// TODO: Implement login logic
	// This will be implemented when auth steps are added
	log.Printf("TODO: Login as manager")
	return nil
}

// LoginAsMember logs in as a member user
func (ctx *ScenarioContext) LoginAsMember() error {
	// TODO: Implement login logic
	// This will be implemented when auth steps are added
	log.Printf("TODO: Login as member")
	return nil
}

// GetAuthenticatedManagerClient returns a manager client with automatic token injection
// Creates or reuses the manager client for the current scenario
func (ctx *ScenarioContext) GetAuthenticatedManagerClient() (*integration_manager.ClientWithResponses, error) {
	// Get current token
	token, err := ctx.GetAuthToken()
	if err != nil {
		return nil, fmt.Errorf("cannot create manager client without token: %w", err)
	}

	// Create or reuse manager client
	if ctx.ManagerClient == nil {
		client, err := integration_manager.NewAuthenticatedClient(
			ctx.ServerURL,
			integration_manager.TokenGetter(func() (string, error) { return token, nil }),
			ctx.Logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create manager client: %w", err)
		}
		ctx.ManagerClient = client
	}

	return ctx.ManagerClient, nil
}
