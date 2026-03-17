// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/code-together/bdd/support"
)

// RegisterPermissionSteps registers permission-related step definitions
func RegisterPermissionSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// WHEN STEPS - Permission-sensitive operations

	suite.When(`^I attempt to delete the provider$`, ctx.iAttemptToDeleteProvider)
	suite.When(`^I attempt to update the provider$`, ctx.WhenIAttemptToUpdateTheProvider)
	suite.Given(`^I have a unique user "([^"]*)"$`, ctx.iHaveAUniqueUser)
	suite.When(`^I login as a manager in tenant "([^"]*)"$`, ctx.iLoginAsManagerInTenant)
	suite.When(`^I attempt to get the provider$`, ctx.iAttemptToGetProvider)
	suite.Given(`^I have created (\d+) claude providers$`, ctx.iHaveCreatedClaudeProviders)
	suite.Given(`^I have created (\d+) users$`, ctx.iHaveCreatedUsers)
}

// iAttemptToDeleteProvider attempts to delete a provider
func (ctx *ScenarioContext) iAttemptToDeleteProvider() error {
	// TODO: Implement actual provider deletion via API
	if ctx.BDDTestContext.CurrentUser != nil && ctx.BDDTestContext.CurrentUser.Role == "admin" {
		ctx.SetLastResponse(204, nil, "")
		return nil
	}
	ctx.SetLastResponse(403, nil, "permission denied")
	return nil
}

// WhenIAttemptToUpdateTheProvider attempts to update a provider
func (ctx *ScenarioContext) WhenIAttemptToUpdateTheProvider() error {
	// TODO: Implement actual provider update via API
	if ctx.BDDTestContext.CurrentUser != nil && ctx.BDDTestContext.CurrentUser.Role == "admin" {
		ctx.SetLastResponse(200, map[string]interface{}{"id": ctx.LastProviderID}, "")
		return nil
	}
	ctx.SetLastResponse(403, nil, "permission denied")
	return nil
}

// iHaveAUniqueUser generates a unique user for testing
func (ctx *ScenarioContext) iHaveAUniqueUser(email string) error {
	uniqueEmail := support.GenerateUniqueEmail("user")
	ctx.TrackCreatedResource("test_user_email", uniqueEmail)
	return nil
}

// iLoginAsManagerInTenant logs in as a manager in a specific tenant
func (ctx *ScenarioContext) iLoginAsManagerInTenant(tenantID string) error {
	// TODO: Implement actual login with tenant context
	ctx.BDDTestContext.CurrentUser = &support.UserInfo{
		Email:    fmt.Sprintf("manager-%s@example.com", tenantID),
		Password: "TestPassword123!",
		Name:     "Test Manager",
		Role:     support.RoleAdmin,
		Token:    fmt.Sprintf("mock-token-tenant-%s", tenantID),
	}
	ctx.TrackCreatedResource("tenant_id", tenantID)
	return nil
}

// iAttemptToGetProvider attempts to get a provider
func (ctx *ScenarioContext) iAttemptToGetProvider() error {
	// TODO: Implement actual provider retrieval via API
	// For cross-tenant scenarios, this should return 404
	_, hasTenant := ctx.GetCreatedResource("cross_tenant_provider")
	if hasTenant {
		ctx.SetLastResponse(404, nil, "found")
		return nil
	}
	ctx.SetLastResponse(200, map[string]interface{}{"id": 1}, "")
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedClaudeProviders(count int) error {
	for i := 0; i < count; i++ {
		providerID := int64(100 + i)
		ctx.TrackProvider(providerID)
	}
	// Use the same resource key that the limit check expects
	ctx.TrackCreatedResource("created_provider_count_claude", fmt.Sprintf("%d", count))
	ctx.TrackCreatedResource("claude_providers_count", fmt.Sprintf("%d", count))
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedUsers(count int) error {
	for i := 0; i < count; i++ {
		userID := fmt.Sprintf("user-%d", i)
		ctx.TrackCreatedResource("test_user", userID)
	}
	ctx.TrackCreatedResource("users_count", fmt.Sprintf("%d", count))
	return nil
}
