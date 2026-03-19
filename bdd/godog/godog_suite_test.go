// Package godog provides the test suite entry point for BDD tests
package godog

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/code-together/bdd/step_definitions"
	"github.com/code-together/bdd/support"
	"github.com/cucumber/godog"
)

// TestGodog runs the godog test suite
func TestGodog(t *testing.T) {
	// Read configuration from environment variables
	format := "pretty"
	if f := os.Getenv("GODOG_FORMAT"); f != "" {
		format = f
	}

	tags := ""
	if tg := os.Getenv("GODOG_TAGS"); tg != "" {
		tags = tg
	}

	suite := godog.TestSuite{
		Name:                 "bdd",
		TestSuiteInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:        format,
			Paths:         []string{"../features"},
			Tags:          tags,
			Strict:        true,
			StopOnFailure: false,
			NoColors:      false,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("godog suite failed")
	}
}

// InitializeScenario sets up the scenario context and registers step definitions
func InitializeScenario(suite *godog.TestSuiteContext) {
	// Create logger for client operations
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create test context with server configuration
	testContext := &step_definitions.ScenarioContext{
		BDDTestContext: &support.BDDTestContext{
			ServerURL: support.GetTestServerURL(),
			TestDBURL: support.GetTestDatabaseURL(),
		},
	}

	// Initialize API clients
	if err := testContext.InitializeClients(testContext.ServerURL, logger); err != nil {
		log.Printf("WARNING: Failed to initialize API clients: %v", err)
		// Continue anyway - scenarios may handle this
	}

	// Get the scenario context from the suite
	ctx := suite.ScenarioContext()

	// Register step definitions
	step_definitions.RegisterCommonSteps(testContext, ctx)
	step_definitions.RegisterAuthSteps(testContext, ctx)
	step_definitions.RegisterPasswordResetSteps(testContext, ctx)
	step_definitions.RegisterPermissionSteps(testContext, ctx)
	step_definitions.RegisterProviderSteps(testContext, ctx)
	step_definitions.RegisterLicenseSteps(testContext, ctx)
	step_definitions.RegisterUsageSteps(testContext, ctx)
	step_definitions.RegisterDashboardSteps(testContext, ctx)
	step_definitions.RegisterHealthSteps(testContext, ctx)

	// Before scenario hook - reset state and setup test environment
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		log.Printf("=== Starting Scenario: %s ===", sc.Name)

		// Reset scenario state
		testContext.ResetScenarioState()

		// Ensure test server is running
		if !support.IsTestServerRunning(testContext.ServerURL) {
			log.Printf("WARNING: Test server is not running at %s", testContext.ServerURL)
			return ctx, fmt.Errorf("test server not running at %s", testContext.ServerURL)
		}

		// Load fixtures and create test data via API
		fixtures, err := support.LoadFixtureData()
		if err != nil {
			return ctx, fmt.Errorf("failed to load fixtures: %w", err)
		}

		// TODO: Create test data from fixtures via API
		// This will be implemented when CreateTestDataFromFixtures is added
		log.Printf("Loaded %d users, %d providers, %d teams, %d licenses from fixtures",
			len(fixtures.Users), len(fixtures.Providers), len(fixtures.Teams), len(fixtures.Licenses))

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
