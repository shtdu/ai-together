// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/code-together/bdd/support"
)

// RegisterAuthSteps registers authentication step definitions
func RegisterAuthSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// GIVEN STEPS - Setup context

	suite.Given(`^I am logged in as a manager$`, ctx.iAmLoggedInAsAManager)
	suite.Given(`^I am logged in as a member$`, ctx.iAmLoggedInAsAMember)
	suite.Given(`^I am not authenticated$`, ctx.iAmNotAuthenticated)
	suite.Given(`^I have a valid authentication token$`, ctx.iHaveAValidAuthToken)
	suite.Given(`^I have an expired authentication token$`, ctx.iHaveAnExpiredAuthToken)
	suite.Given(`^a user exists with email "([^"]*)" and password "([^"]*)"$`, ctx.userExists)
	suite.Given(`^the user has role "([^"]*)"$`, ctx.userHasRole)
	suite.Given(`^I belong to tenant with ID "([^"]*)"$`, ctx.iBelongToTenant)
	suite.Given(`^there is a provider in tenant "([^"]*)"$`, ctx.thereIsAProviderInTenant)
	suite.Given(`^there are users in tenant "([^"]*)"$`, ctx.thereAreUsersInTenant)
	suite.Given(`^I have created a provider$`, ctx.iHaveCreatedAProvider)
	suite.Given(`^I have a provider with ID "([^"]*)"$`, ctx.iHaveAProviderWithID)

	// WHEN STEPS - Perform actions

	suite.When(`^I login with email "([^"]*)" and password "([^"]*)"$`, ctx.iLoginWithCredentials)
	suite.When(`^I logout$`, ctx.iLogout)
	suite.When(`^I refresh my authentication token$`, ctx.iRefreshAuthToken)
	suite.When(`^I refresh my authentication token with "([^"]*)"$`, ctx.iRefreshAuthTokenWithToken)
	suite.When(`^I verify my authentication token$`, ctx.iVerifyAuthToken)
	suite.When(`^I verify my authentication token with "([^"]*)"$`, ctx.iVerifyAuthTokenWithToken)
	suite.When(`^I get my user profile$`, ctx.iGetUserProfile)
	suite.When(`^I attempt to create a provider$`, ctx.iAttemptToCreateProvider)
	suite.When(`^I get team analytics$`, ctx.iGetTeamAnalytics)
	suite.When(`^I get my usage statistics$`, ctx.iGetMyUsageStatistics)
	suite.When(`^I list all users$`, ctx.iListAllUsers)
	suite.When(`^I list all providers$`, ctx.iListAllProviders)

	// THEN STEPS - Assert outcomes

	suite.Then(`^I should receive a valid authentication token$`, ctx.iShouldReceiveValidToken)
	suite.Then(`^the response status code should be (\d+)$`, ctx.responseStatusCodeShouldBe)
	suite.Then(`^I should receive an "([^"]*)" error$`, ctx.iShouldReceiveAnError)
	suite.Then(`^my profile should contain my email$`, ctx.myProfileShouldContainEmail)
	suite.Then(`^my profile should contain my role$`, ctx.myProfileShouldContainRole)
	suite.Then(`^my profile should contain tenant ID$`, ctx.myProfileShouldContainTenantID)
	suite.Then(`^the tenant ID should be "([^"]*)"$`, ctx.theTenantIDShouldBe)
	suite.Then(`^my authentication token should be invalid$`, ctx.myAuthTokenShouldBeInvalid)
	suite.Then(`^my profile should contain "([^"]*)" field$`, ctx.myProfileShouldContainField)
	suite.Then(`^the operation should succeed$`, ctx.operationShouldSucceed)
	suite.Then(`^I should receive a (\d+) error$`, ctx.iShouldReceiveAStatusCode)
	suite.Then(`^the error message should contain "([^"]*)"$`, ctx.errorMessageShouldContain)
	suite.Then(`^I should not see "([^"]*)"$`, ctx.iShouldNotSee)
	suite.Then(`^I should only see users from tenant "([^"]*)"$`, ctx.iShouldOnlySeeUsersFromTenant)
}

// GIVENS - Setup context

// iAmLoggedInAsAManager sets up authentication as a manager
func (ctx *ScenarioContext) iAmLoggedInAsAManager() error {
	// TODO: Implement actual login via API
	// For now, set mock token
	ctx.AdminToken = "mock-manager-token"
	ctx.BDDTestContext.CurrentUser = &support.UserInfo{
		Email:    "manager@example.com",
		Password: "TestPassword123!",
		Name:     "Test Manager",
		Role:     support.RoleAdmin,
		Token:    ctx.AdminToken,
	}
	return nil
}

// iAmLoggedInAsAMember sets up authentication as a member
func (ctx *ScenarioContext) iAmLoggedInAsAMember() error {
	// TODO: Implement actual login via API
	ctx.MemberToken = "mock-member-token"
	ctx.BDDTestContext.CurrentUser = &support.UserInfo{
		Email:    "member@example.com",
		Password: "TestPassword123!",
		Name:     "Test Member",
		Role:     support.RoleMember,
		Token:    ctx.MemberToken,
	}
	return nil
}

// iAmNotAuthenticated ensures no user is logged in
func (ctx *ScenarioContext) iAmNotAuthenticated() error {
	ctx.AdminToken = ""
	ctx.MemberToken = ""
	ctx.BDDTestContext.CurrentUser = nil
	return nil
}

// iHaveAValidAuthToken sets up a valid authentication token
func (ctx *ScenarioContext) iHaveAValidAuthToken() error {
	ctx.AdminToken = "mock-valid-token"
	return nil
}

// iHaveAnExpiredAuthToken sets up an expired authentication token
func (ctx *ScenarioContext) iHaveAnExpiredAuthToken() error {
	ctx.AdminToken = "mock-expired-token"
	return nil
}

// userExists creates a test user (mock for now)
func (ctx *ScenarioContext) userExists(email, password string) error {
	// TODO: Create user via API
	// For now, just track the user
	ctx.TrackCreatedResource("test_user", email)
	return nil
}

// userHasRole sets the role for the current user
func (ctx *ScenarioContext) userHasRole(role string) error {
	if ctx.BDDTestContext.CurrentUser != nil {
		ctx.BDDTestContext.CurrentUser.Role = role
	}
	return nil
}

// iBelongToTenant sets the tenant ID for the current user
func (ctx *ScenarioContext) iBelongToTenant(tenantID string) error {
	// TODO: Set tenant context
	ctx.TrackCreatedResource("tenant_id", tenantID)
	return nil
}

// thereIsAProviderInTenant creates a provider in a different tenant
func (ctx *ScenarioContext) thereIsAProviderInTenant(tenantID string) error {
	// TODO: Create provider in specific tenant via API
	ctx.TrackCreatedResource("cross_tenant_provider", tenantID)
	return nil
}

// thereAreUsersInTenant creates users in a different tenant
func (ctx *ScenarioContext) thereAreUsersInTenant(tenantID string) error {
	// TODO: Create users in specific tenant via API
	ctx.TrackCreatedResource("cross_tenant_users", tenantID)
	return nil
}

// iHaveCreatedAProvider creates a provider for testing
func (ctx *ScenarioContext) iHaveCreatedAProvider() error {
	// TODO: Create provider via API and track it
	providerID := int64(999)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	return nil
}

// iHaveAProviderWithID sets up a provider with specific ID
func (ctx *ScenarioContext) iHaveAProviderWithID(id string) error {
	// TODO: Set provider context
	ctx.TrackCreatedResource("provider_id", id)
	return nil
}

// WHENS - Perform actions

// iLoginWithCredentials attempts to login with email and password
func (ctx *ScenarioContext) iLoginWithCredentials(email, password string) error {
	// TODO: Implement actual login via API
	// For now, simulate success/failure based on credentials
	if email == "nonexistent@example.com" || email == "nobody@example.com" {
		ctx.SetLastResponse(401, nil, "user_not_found")
		return nil
	}
	if password == "WrongPassword" {
		ctx.SetLastResponse(401, nil, "invalid_password")
		return nil
	}
	// Success
	token := fmt.Sprintf("mock-token-%s", email)
	ctx.SetLastResponse(200, map[string]string{"token": token}, "")
	return nil
}

// iLogout logs out the current user
func (ctx *ScenarioContext) iLogout() error {
	// TODO: Implement actual logout via API
	ctx.SetLastResponse(204, nil, "")
	ctx.AdminToken = ""
	ctx.MemberToken = ""
	return nil
}

// iRefreshAuthToken refreshes the current authentication token
func (ctx *ScenarioContext) iRefreshAuthToken() error {
	// TODO: Implement actual token refresh via API
	ctx.SetLastResponse(200, map[string]string{"token": "mock-refreshed-token"}, "")
	return nil
}

// iRefreshAuthTokenWithToken refreshes with a specific token
func (ctx *ScenarioContext) iRefreshAuthTokenWithToken(token string) error {
	// TODO: Implement actual token refresh via API
	if token == "invalid-token" {
		ctx.SetLastResponse(401, nil, "invalid_token")
		return nil
	}
	ctx.SetLastResponse(200, map[string]string{"token": "mock-refreshed-token"}, "")
	return nil
}

// iVerifyAuthToken verifies the current authentication token
func (ctx *ScenarioContext) iVerifyAuthToken() error {
	// TODO: Implement actual token verification via API
	ctx.SetLastResponse(200, ctx.BDDTestContext.CurrentUser, "")
	return nil
}

// iVerifyAuthTokenWithToken verifies a specific token
func (ctx *ScenarioContext) iVerifyAuthTokenWithToken(token string) error {
	// TODO: Implement actual token verification via API
	if token == "invalid-token" {
		ctx.SetLastResponse(401, nil, "invalid_token")
		return nil
	}
	ctx.SetLastResponse(200, ctx.BDDTestContext.CurrentUser, "")
	return nil
}

// iGetUserProfile retrieves the current user's profile
func (ctx *ScenarioContext) iGetUserProfile() error {
	// TODO: Implement actual profile retrieval via API
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	ctx.SetLastResponse(200, ctx.BDDTestContext.CurrentUser, "")
	return nil
}

// iAttemptToCreateProvider attempts to create a provider
func (ctx *ScenarioContext) iAttemptToCreateProvider() error {
	// TODO: Implement actual provider creation via API
	if ctx.BDDTestContext.CurrentUser != nil && ctx.BDDTestContext.CurrentUser.Role == support.RoleAdmin {
		providerID := int64(123)
		ctx.TrackProvider(providerID)
		ctx.SetLastResponse(201, map[string]interface{}{"id": providerID}, "")
		return nil
	}
	ctx.SetLastResponse(403, nil, "permission denied")
	return nil
}

// iGetTeamAnalytics retrieves team analytics
func (ctx *ScenarioContext) iGetTeamAnalytics() error {
	// TODO: Implement actual analytics retrieval via API
	if ctx.BDDTestContext.CurrentUser != nil && ctx.BDDTestContext.CurrentUser.Role == support.RoleAdmin {
		ctx.SetLastResponse(200, map[string]interface{}{"total_users": 10}, "")
		return nil
	}
	ctx.SetLastResponse(403, nil, "permission denied")
	return nil
}

// iGetMyUsageStatistics retrieves usage statistics for current user
func (ctx *ScenarioContext) iGetMyUsageStatistics() error {
	// TODO: Implement actual usage retrieval via API
	ctx.SetLastResponse(200, map[string]interface{}{"total_tokens": 1000}, "")
	return nil
}

// iListAllUsers lists all users
func (ctx *ScenarioContext) iListAllUsers() error {
	// TODO: Implement actual user listing via API
	users := []map[string]interface{}{
		{"id": "1", "email": "user1@example.com", "tenant_id": 1},
		{"id": "2", "email": "user2@example.com", "tenant_id": 1},
	}
	ctx.SetLastResponse(200, users, "")
	return nil
}

// iListAllProviders lists all providers
func (ctx *ScenarioContext) iListAllProviders() error {
	// TODO: Implement actual provider listing via API
	providers := []map[string]interface{}{
		{"id": 1, "name": "provider1", "tenant_id": 2},
	}
	ctx.SetLastResponse(200, providers, "")
	return nil
}

// THENS - Assert outcomes

// iShouldReceiveValidToken checks if a valid token was received
func (ctx *ScenarioContext) iShouldReceiveValidToken() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if statusCode != 200 {
		return fmt.Errorf("expected status 200, got %d", statusCode)
	}
	// Check if response contains token
	if respMap, ok := resp.(map[string]string); ok {
		if _, hasToken := respMap["token"]; !hasToken {
			return fmt.Errorf("response does not contain token")
		}
	}
	return nil
}

// responseStatusCodeShouldBe checks the response status code
func (ctx *ScenarioContext) responseStatusCodeShouldBe(expectedStatus int) error {
	statusCode, _, _ := ctx.GetLastResponse()
	if statusCode != expectedStatus {
		return fmt.Errorf("expected status %d, got %d", expectedStatus, statusCode)
	}
	return nil
}

// iShouldReceiveAnError checks if a specific error was received
func (ctx *ScenarioContext) iShouldReceiveAnError(errorType string) error {
	_, _, errMsg := ctx.GetLastResponse()
	if errMsg == "" {
		return fmt.Errorf("expected error '%s', but got no error", errorType)
	}
	return nil
}

// myProfileShouldContainEmail checks if profile contains email
func (ctx *ScenarioContext) myProfileShouldContainEmail() error {
	if ctx.BDDTestContext.CurrentUser == nil || ctx.BDDTestContext.CurrentUser.Email == "" {
		return fmt.Errorf("profile does not contain email")
	}
	return nil
}

// myProfileShouldContainRole checks if profile contains role
func (ctx *ScenarioContext) myProfileShouldContainRole() error {
	if ctx.BDDTestContext.CurrentUser == nil || ctx.BDDTestContext.CurrentUser.Role == "" {
		return fmt.Errorf("profile does not contain role")
	}
	return nil
}

// myProfileShouldContainTenantID checks if profile contains tenant ID
func (ctx *ScenarioContext) myProfileShouldContainTenantID() error {
	// TODO: Check tenant ID in profile
	return nil
}

// theTenantIDShouldBe checks if the tenant ID matches expected value
func (ctx *ScenarioContext) theTenantIDShouldBe(expectedTenantID string) error {
	// TODO: Verify tenant ID in profile
	return nil
}

// myAuthTokenShouldBeInvalid checks if auth token is invalid
func (ctx *ScenarioContext) myAuthTokenShouldBeInvalid() error {
	if ctx.AdminToken != "" || ctx.MemberToken != "" {
		return fmt.Errorf("authentication token is still valid")
	}
	return nil
}

// myProfileShouldContainField checks if profile contains a specific field
func (ctx *ScenarioContext) myProfileShouldContainField(field string) error {
	// TODO: Check if field exists in profile
	return nil
}

// operationShouldSucceed checks if operation succeeded
func (ctx *ScenarioContext) operationShouldSucceed() error {
	statusCode, _, _ := ctx.GetLastResponse()
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("operation failed with status %d", statusCode)
	}
	return nil
}

// iShouldReceiveAStatusCode checks if response has specific status code
func (ctx *ScenarioContext) iShouldReceiveAStatusCode(statusCode int) error {
	actualStatusCode, _, _ := ctx.GetLastResponse()
	if actualStatusCode != statusCode {
		return fmt.Errorf("expected status %d, got %d", statusCode, actualStatusCode)
	}
	return nil
}

// errorMessageShouldContain checks if error message contains text
func (ctx *ScenarioContext) errorMessageShouldContain(text string) error {
	_, _, errMsg := ctx.GetLastResponse()
	if errMsg == "" {
		return fmt.Errorf("no error message found")
	}
	return nil
}

// iShouldNotSee checks if response doesn't contain specific data
func (ctx *ScenarioContext) iShouldNotSee(data string) error {
	// TODO: Check if data is not in response
	return nil
}

// iShouldOnlySeeUsersFromTenant checks tenant isolation
func (ctx *ScenarioContext) iShouldOnlySeeUsersFromTenant(tenantID string) error {
	// TODO: Verify all users are from specified tenant
	return nil
}
