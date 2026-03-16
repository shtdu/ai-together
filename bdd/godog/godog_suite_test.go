// Package godog provides the test suite entry point for BDD tests
package godog

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/code-together/bdd/step_definitions"
	"github.com/code-together/bdd/support"
)

// TestGodog runs the godog test suite
func TestGodog(t *testing.T) {
	suite := godog.TestSuite{
		Name:                 "bdd",
		TestSuiteInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:      "pretty",
			Paths:       []string{"../features"},
			Strict:      true,
			StopOnFailure: false,
			NoColors:    false,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("godog suite failed")
	}
}

// InitializeScenario sets up the scenario context and registers step definitions
func InitializeScenario(ctx *godog.ScenarioContext) {
	// Create test context with server configuration
	testContext := &step_definitions.ScenarioContext{
		BDDTestContext: &support.BDDTestContext{
			ServerURL: support.GetTestServerURL(),
			TestDBURL: support.GetTestDatabaseURL(),
		},
	}

	// Register step definitions
	step_definitions.RegisterCommonSteps(testContext, ctx)
	step_definitions.RegisterAuthSteps(testContext, ctx)
	step_definitions.RegisterPermissionSteps(testContext, ctx)
	step_definitions.RegisterProviderSteps(testContext, ctx)
	step_definitions.RegisterLicenseSteps(testContext, ctx)
	step_definitions.RegisterUsageSteps(testContext, ctx)
	step_definitions.RegisterDashboardSteps(testContext, ctx)
	step_definitions.RegisterHealthSteps(testContext, ctx)

	// Register additional step definitions as they are implemented
	// step_definitions.RegisterProviderSteps(testContext, ctx)
	// step_definitions.RegisterLicenseSteps(testContext, ctx)
	// step_definitions.RegisterUsageSteps(testContext, ctx)
	// step_definitions.RegisterDashboardSteps(testContext, ctx)
	// step_definitions.RegisterHealthSteps(testContext, ctx)

	// Before scenario hook - reset state and setup test environment
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		log.Printf("=== Starting Scenario: %s ===", sc.Name)

		// Reset scenario state
		testContext.ResetScenarioState()

		// Ensure test server is running
		if !support.IsTestServerRunning(testContext.ServerURL) {
			log.Printf("WARNING: Test server is not running at %s", testContext.ServerURL)
			// Continue anyway - scenarios may handle this appropriately
		}

		// Load standard fixtures
		_, err := support.LoadFixtureData()
		if err != nil {
			log.Printf("WARNING: Could not load fixture data: %v (using defaults)", err)
		}

		return ctx, nil
	})

	// After scenario hook - clean up resources
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		log.Printf("=== Completed Scenario: %s ===", sc.Name)

		// Clean up resources even if scenario failed
		cleanupErr := testContext.CleanupScenarioResources()
		if cleanupErr != nil {
			// Log cleanup error but don't fail the scenario
			log.Printf("WARNING: Cleanup failed: %v", cleanupErr)
		}

		return ctx, err
	})
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
