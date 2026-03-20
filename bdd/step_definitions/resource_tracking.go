// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"
	"log"
)

// TrackScenarioOutlineProvider tracks providers created in scenario outlines
// This helps with cleanup and verification
func (ctx *ScenarioContext) TrackScenarioOutlineProvider(kind string, providerID int64) {
	ctx.TrackProvider(providerID)
	log.Printf("Tracked %s provider with ID %d from scenario outline", kind, providerID)
}

// GetProviderCount returns the current number of tracked providers
func (ctx *ScenarioContext) GetProviderCount() int {
	providers := ctx.GetCreatedProviders()
	return len(providers)
}

// CleanupAllProviders removes all tracked providers
// This is useful for scenario outlines that create multiple providers
func (ctx *ScenarioContext) CleanupAllProviders() error {
	providers := ctx.GetCreatedProviders()
	log.Printf("Cleaning up %d providers from scenario outline", len(providers))

	for _, providerID := range providers {
		log.Printf("Cleaning up provider %d from scenario outline", providerID)
		// TODO: Delete provider via API
		// ctx.Client.DeleteApiV1ProvidersProviderIdWithResponse(context.Background(), providerID)
	}

	ctx.ClearCreatedResources()
	return nil
}

// VerifyProviderCount verifies that the expected number of providers exist
func (ctx *ScenarioContext) VerifyProviderCount(expectedCount int) error {
	actualCount := ctx.GetProviderCount()
	if actualCount != expectedCount {
		return fmt.Errorf("expected %d providers, found %d", expectedCount, actualCount)
	}
	return nil
}
