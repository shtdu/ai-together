// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"log"

	"github.com/code-together/bdd/support"
)

// ScenarioContext wraps BDDTestContext with scenario-specific methods
type ScenarioContext struct {
	*support.BDDTestContext
}

// ResetScenarioState resets all scenario state before execution
func (ctx *ScenarioContext) ResetScenarioState() {
	ctx.Reset()
	log.Printf("Scenario state reset")
}

// CleanupScenarioResources cleans up all resources created during the scenario
func (ctx *ScenarioContext) CleanupScenarioResources() error {
	log.Printf("Cleaning up scenario resources...")

	// Clean up providers
	providers := ctx.GetCreatedProviders()
	for _, providerID := range providers {
		log.Printf("Cleaning up provider %d", providerID)
		// TODO: Delete provider via API
		// ctx.Client.DeleteApiV1ProvidersProviderIdWithResponse(context.Background(), providerID)
	}

	// Clean up users
	users := ctx.GetCreatedUsers()
	for _, userID := range users {
		log.Printf("Cleaning up user %s", userID)
		// TODO: Delete user via API
		// ctx.ManagerClient.DeleteApiV1UsersUserIdWithResponse(context.Background(), userID)
	}

	// Clean up teams
	teams := ctx.GetCreatedTeams()
	for _, teamID := range teams {
		log.Printf("Cleaning up team %d", teamID)
		// Skip default team (ID 1)
		if teamID == 1 {
			continue
		}
		// TODO: Delete team via API
		// ctx.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(context.Background(), teamID)
	}

	// Clear resource lists
	ctx.ClearCreatedResources()

	log.Printf("Cleanup complete")
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
