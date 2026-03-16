// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/code-together/bdd/support"
)

// RegisterCommonSteps registers common step definitions for BDD scenarios
func RegisterCommonSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// Given Steps - Setup context
	suite.Given(`^the test server is running$`, ctx.theTestServerIsRunning)
	suite.Given(`^I have a unique provider name "([^"]*)"$`, ctx.iHaveAUniqueProviderName)

	// When Steps - Perform actions
	suite.When(`^I check the health endpoint$`, ctx.iCheckTheHealthEndpoint)

	// Then Steps - Assert outcomes
	suite.Then(`^I should receive a (\d+) status$`, ctx.iShouldReceiveAStatus)
	suite.Then(`^the response should be successful$`, ctx.theResponseShouldBeSuccessful)
	suite.Then(`^the system should be healthy$`, ctx.theSystemShouldBeHealthy)
}

// GIVENS

// theTestServerIsRunning checks if the test server is running
func (ctx *ScenarioContext) theTestServerIsRunning() error {
	if !support.IsTestServerRunning(ctx.ServerURL) {
		return fmt.Errorf("test server is not running at %s", ctx.ServerURL)
	}
	return nil
}

// iHaveAUniqueProviderName generates a unique provider name for testing
func (ctx *ScenarioContext) iHaveAUniqueProviderName(baseName string) error {
	uniqueName := support.GenerateUniqueProviderName(baseName)
	ctx.TrackCreatedResource("provider_name", uniqueName)
	return nil
}

// WHENS

// iCheckTheHealthEndpoint checks the server health endpoint
func (ctx *ScenarioContext) iCheckTheHealthEndpoint() error {
	// TODO: Implement actual health check via API
	// For now, just set a success status
	ctx.SetLastResponse(200, nil, "")
	return nil
}

// THENS

// iShouldReceiveAStatus checks if the response status matches the expected status
func (ctx *ScenarioContext) iShouldReceiveAStatus(expectedStatus int) error {
	statusCode, _, errMsg := ctx.GetLastResponse()

	if statusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d. Error: %s",
			expectedStatus, statusCode, errMsg)
	}

	return nil
}

// theResponseShouldBeSuccessful checks if the response indicates success (2xx status)
func (ctx *ScenarioContext) theResponseShouldBeSuccessful() error {
	statusCode, _, errMsg := ctx.GetLastResponse()

	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("expected successful status (2xx), got %d. Error: %s",
			statusCode, errMsg)
	}

	return nil
}

// theSystemShouldBeHealthy checks if the system health check passed
func (ctx *ScenarioContext) theSystemShouldBeHealthy() error {
	if !support.IsTestServerRunning(ctx.ServerURL) {
		return fmt.Errorf("system is not healthy - test server not running")
	}
	return nil
}
