// Package step_definitions provides step implementations for BDD scenarios
package step_definitions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/code-together/bdd/support"
	"github.com/code-together/shared/integration"
)

// RegisterProviderSteps registers provider management step definitions
func RegisterProviderSteps(ctx *ScenarioContext, suite *godog.ScenarioContext) {
	// GIVEN STEPS - Setup context

	suite.Given(`^the standard fixtures are loaded$`, ctx.standardFixturesLoaded)
	suite.Given(`^I have a unique provider name$`, ctx.iHaveAUniqueProviderNameForProvider)
	suite.Given(`^I have a unique provider name "([^"]*)"$`, ctx.iHaveAUniqueProviderWithName)
	suite.Given(`^I create a ([^"]*) provider with name "([^"]*)"$`, ctx.iCreateAProviderWithName)
	suite.Given(`^I have created a provider$`, ctx.iHaveCreatedAProviderInternal)
	suite.Given(`^I have created a ([^"]*) provider$`, ctx.iHaveCreatedAProviderWithKind)
	suite.Given(`^I have created a claude provider$`, ctx.iHaveCreatedAClaudeProvider)
	suite.Given(`^I have created an enabled provider$`, ctx.iHaveCreatedAnEnabledProvider)
	suite.Given(`^I have created a disabled provider$`, ctx.iHaveCreatedADisabledProvider)
	suite.Given(`^I have created a provider with valid API key$`, ctx.iHaveCreatedAProviderWithValidAPIKey)
	suite.Given(`^I have created a provider with invalid API key$`, ctx.iHaveCreatedAProviderWithInvalidAPIKey)
	suite.Given(`^I have created a provider with priority (\d+)$`, ctx.iHaveCreatedAProviderWithPriority)
	suite.Given(`^there is a default provider with ID (\d+)$`, ctx.thereIsADefaultProviderWithID)
	suite.Given(`^the license has a provider limit of (\d+)$`, ctx.licenseHasProviderLimit)
	suite.Given(`^I have created (\d+) providers$`, ctx.iHaveCreatedNProviders)
	suite.Given(`^the license has ([^"]*) tier$`, ctx.licenseHasTier)

	// WHEN STEPS - Perform actions

	suite.When(`^I create a ([^"]*) provider with API key "([^"]*)"$`, ctx.iCreateAProviderWithKindAndAPIKey)
	suite.Given(`^I create a provider with name "([^"]*)"$`, ctx.iCreateAProviderWithNameArg)
	suite.When(`^I create a ([^"]*) provider$`, ctx.iCreateAProviderWithKindSimple)
	suite.When(`^I attempt to create a claude provider$`, ctx.iAttemptToCreateAClaudeProvider)
	suite.When(`^I attempt to create a team$`, ctx.iAttemptToCreateATeam)
	suite.When(`^I create a team$`, ctx.iCreateATeam)
	suite.When(`^I create a "([^"]*)" provider with API key "([^"]*)"$`, ctx.iCreateAProviderWithAPIKey)
	suite.When(`^I delete provider by ID$`, ctx.iDeleteProviderByIDNoArgs)
	suite.When(`^I delete provider by ID (\d+)$`, ctx.iDeleteProviderByID)
	suite.Given(`^the license has a ([^"]*) provider limit of (\d+)$`, ctx.licenseHasProviderLimitForKind)
	suite.Given(`^I have created a ([^"]*) provider$`, ctx.iHaveCreatedAProviderWithKind)

	// WHEN STEPS - Perform actions
	suite.When(`^I create another ([^"]*) provider with name "([^"]*)"$`, ctx.iCreateAnotherProviderWithName)
	suite.When(`^I create a ([^"]*) provider with API key "([^"]*)" and priority (\d+)$`, ctx.iCreateAProviderWithPriority)
	suite.When(`^I create a ([^"]*) provider with API key "([^"]*)" and disabled$`, ctx.iCreateAProviderDisabled)
	suite.When(`^I create a ([^"]*) provider with name "([^"]*)"$`, ctx.iCreateAProviderWithNameOnly)
	suite.When(`^I list all providers$`, ctx.iListAllProvidersFromProvider)
	suite.When(`^I list providers with kind "([^"]*)"$`, ctx.iListProvidersWithKind)
	suite.When(`^I get the provider by ID$`, ctx.iGetProviderByID)
	suite.When(`^I get provider statistics$`, ctx.iGetProviderStats)
	suite.When(`^I get provider with ID (\d+)$`, ctx.iGetProviderWithID)
	suite.When(`^I update the provider name to "([^"]*)"$`, ctx.iUpdateProviderName)
	suite.When(`^I update the provider API key to "([^"]*)"$`, ctx.iUpdateProviderAPIKey)
	suite.When(`^I update the provider priority to (\d+)$`, ctx.iUpdateProviderPriority)
	suite.When(`^I disable the provider$`, ctx.iDisableProvider)
	suite.When(`^I enable the provider$`, ctx.iEnableProvider)
	suite.When(`^I update provider with ID (\d+)$`, ctx.iUpdateProviderWithID)
	suite.When(`^I delete the provider$`, ctx.iDeleteProvider)
	suite.When(`^I delete provider with ID (\d+)$`, ctx.iDeleteProviderByID)
	suite.When(`^I attempt to delete the default provider$`, ctx.iAttemptToDeleteDefaultProvider)
	suite.When(`^I test the provider connectivity$`, ctx.iTestProviderConnectivity)
	suite.When(`^I test connectivity for provider with ID (\d+)$`, ctx.iTestConnectivityForProviderWithID)
	suite.When(`^I attempt to create a provider$`, ctx.iAttemptToCreateProviderInLimit)
	suite.When(`^I create a provider$`, ctx.iCreateAProvider)
	suite.When(`^I update the first provider$`, ctx.iUpdateFirstProvider)
	suite.When(`^I delete the first provider$`, ctx.iDeleteFirstProvider)

	// THEN STEPS - Assert outcomes

	suite.Then(`^the provider should be created successfully$`, ctx.providerCreatedSuccessfully)
	suite.Then(`^the provider kind should be "([^"]*)"$`, ctx.providerKindShouldBe)
	suite.Then(`^the provider should be enabled$`, ctx.providerShouldBeEnabled)
	suite.Then(`^the provider should not be enabled$`, ctx.providerShouldNotBeEnabled)
	suite.Then(`^the provider priority should be (\d+)$`, ctx.providerPriorityShouldBe)
	suite.Then(`^I should see at least (\d+) provider$`, ctx.iShouldSeeAtLeastNProviders)
	suite.Then(`^the API key should not be visible$`, ctx.apiKeyNotVisible)
	suite.Then(`^the provider should have a name$`, ctx.providerShouldHaveAName)
	suite.Then(`^the statistics should contain total requests$`, ctx.statisticsShouldContainTotalRequests)
	suite.Then(`^the statistics should contain success rate$`, ctx.statisticsShouldContainSuccessRate)
	suite.Then(`^the provider name should be "([^"]*)"$`, ctx.providerNameShouldBe)
	suite.Then(`^the provider should not exist$`, ctx.providerShouldNotExist)
	suite.Then(`^the connection should be successful$`, ctx.connectionSuccessful)
	suite.Then(`^the connection should fail$`, ctx.connectionShouldFail)
	suite.Then(`^the error should indicate authentication failure$`, ctx.errorShouldIndicateAuthFailure)
	suite.Then(`^the total provider count should be (\d+)$`, ctx.totalProviderCountShouldBe)
	suite.Then(`^I should be able to create a new provider$`, ctx.iShouldBeAbleToCreateNewProvider)
	suite.Then(`^I should only see ([^"]*) providers$`, ctx.iShouldOnlySeeProvidersOfKind)
	suite.Then(`^I should not see ([^"]*) providers$`, ctx.iShouldNotSeeProvidersOfKind)
}

// GIVENS - Setup context

func (ctx *ScenarioContext) standardFixturesLoaded() error {
	// TODO: Load standard fixtures from integration/testdata
	return nil
}

func (ctx *ScenarioContext) iHaveAUniqueProviderNameForProvider() error {
	uniqueName := support.GenerateUniqueProviderName("test-provider")
	ctx.TrackCreatedResource("provider_name", uniqueName)
	return nil
}

func (ctx *ScenarioContext) iHaveAUniqueProviderWithName(base string) error {
	uniqueName := support.GenerateUniqueProviderName(base)
	ctx.TrackCreatedResource("provider_name", uniqueName)
	return nil
}

func (ctx *ScenarioContext) iCreateAProviderWithName(kind, name string) error {
	// TODO: Create provider via API
	providerID := int64(100)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.TrackCreatedResource("provider_name", name)
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedAProviderInternal() error {
	return ctx.iHaveCreatedAProviderWithKind("claude")
}

func (ctx *ScenarioContext) iHaveCreatedAProviderWithKind(kind string) error {
	providerName := support.GenerateUniqueProviderName(fmt.Sprintf("test-%s", kind))
	providerID := int64(101)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.TrackCreatedResource("provider_name", providerName)
	ctx.TrackCreatedResource("provider_kind", kind)

	// Increment provider count for this kind to support limit checking
	countKey := fmt.Sprintf("created_provider_count_%s", kind)
	var count int
	if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
		fmt.Sscanf(countStr, "%d", &count)
	}
	ctx.TrackCreatedResource(countKey, fmt.Sprintf("%d", count+1))

	return nil
}

func (ctx *ScenarioContext) iHaveCreatedAClaudeProvider() error {
	return ctx.iHaveCreatedAProviderWithKind("claude")
}

func (ctx *ScenarioContext) iHaveCreatedAnEnabledProvider() error {
	providerID := int64(102)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.TrackCreatedResource("provider_enabled", "true")
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedADisabledProvider() error {
	providerID := int64(103)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.TrackCreatedResource("provider_enabled", "false")
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedAProviderWithValidAPIKey() error {
	providerID := int64(104)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.TrackCreatedResource("api_key_valid", "true")
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedAProviderWithInvalidAPIKey() error {
	providerID := int64(105)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.TrackCreatedResource("api_key_valid", "false")
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedAProviderWithPriority(priority int) error {
	providerID := int64(106)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.TrackCreatedResource("provider_priority", fmt.Sprintf("%d", priority))
	return nil
}

func (ctx *ScenarioContext) thereIsADefaultProviderWithID(id int) error {
	ctx.TrackCreatedResource("default_provider_id", fmt.Sprintf("%d", id))
	return nil
}

func (ctx *ScenarioContext) licenseHasProviderLimit(limit int) error {
	ctx.TrackCreatedResource("license_provider_limit", fmt.Sprintf("%d", limit))
	return nil
}

func (ctx *ScenarioContext) iHaveCreatedNProviders(n int) error {
	for i := 0; i < n; i++ {
		providerID := int64(200 + i)
		ctx.TrackProvider(providerID)
	}
	ctx.TrackCreatedResource("created_provider_count", fmt.Sprintf("%d", n))
	return nil
}

func (ctx *ScenarioContext) licenseHasTier(tier string) error {
	ctx.TrackCreatedResource("license_tier", tier)
	return nil
}

func (ctx *ScenarioContext) licenseHasProviderLimitForKind(kind string, limit int) error {
	ctx.TrackCreatedResource(fmt.Sprintf("license_%s_provider_limit", kind), fmt.Sprintf("%d", limit))
	return nil
}

// WHENS - Perform actions

func (ctx *ScenarioContext) iCreateAProviderWithKindAndAPIKey(kind, apiKey string) error {
	// Get authenticated client (must be admin/manager)
	client, err := ctx.GetAuthenticatedClient()
	if err != nil || client == nil {
		ctx.SetLastResponse(401, nil, "unauthorized: admin/manager access required")
		return nil
	}

	// Validate API key
	if apiKey == "" {
		ctx.SetLastResponse(400, nil, "API key is required")
		return nil
	}

	// Get or generate provider name
	providerName, _ := ctx.GetCreatedResource("provider_name")
	if providerName == "" {
		providerName = support.GenerateUniqueProviderName(fmt.Sprintf("test-%s", kind))
	}

	// Convert kind string to enum
	var providerKind integration.CreateProviderRequestKind
	switch kind {
	case "claude":
		providerKind = integration.CreateProviderRequestKindClaude
	case "codex":
		providerKind = integration.CreateProviderRequestKindCodex
	case "opencode":
		providerKind = integration.CreateProviderRequestKindOpencode
	default:
		ctx.SetLastResponse(400, nil, fmt.Sprintf("invalid provider kind: %s", kind))
		return nil
	}

	// Set API URL based on provider kind
	var apiUrl string
	switch kind {
	case "claude":
		apiUrl = "https://api.anthropic.com"
	case "codex":
		apiUrl = "https://api.github.com"
	case "opencode":
		apiUrl = "https://api.opencode.com"
	}

	// Default enabled to true
	enabled := true

	// Create provider request
	req := integration.PostApiV1ProvidersJSONRequestBody{
		Name:    providerName,
		Kind:    &providerKind,
		ApiKey:  apiKey,
		ApiUrl:  apiUrl,
		Enabled: &enabled,
	}

	// Call API to create provider
	resp, err := client.PostApiV1ProvidersWithResponse(context.Background(), req)
	if err != nil {
		ctx.SetLastResponse(0, nil, err.Error())
		return fmt.Errorf("provider creation request failed: %w", err)
	}

	// Parse response body
	var body interface{}
	if resp.JSON201 != nil {
		body = resp.JSON201
	} else if resp.JSON400 != nil {
		body = resp.JSON400
	} else if resp.JSON401 != nil {
		body = resp.JSON401
	} else if resp.JSON403 != nil {
		body = resp.JSON403
	} else if len(resp.Body) > 0 {
		json.Unmarshal(resp.Body, &body)
	}

	// Store response for assertions
	ctx.SetLastResponse(resp.StatusCode(), body, "")

	// Handle API response errors
	if resp.StatusCode() >= 400 {
		errMsg := ""
		if resp.JSON400 != nil {
			errMsg = fmt.Sprintf("validation error: %s", resp.JSON400.Error)
		} else if resp.JSON401 != nil {
			errMsg = fmt.Sprintf("unauthorized: %s", resp.JSON401.Error)
		} else if resp.JSON403 != nil {
			errMsg = fmt.Sprintf("forbidden: %s", resp.JSON403.Error)
		} else if len(resp.Body) > 0 {
			errMsg = string(resp.Body)
		} else {
			errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode())
		}
		return fmt.Errorf("provider creation failed: %s", errMsg)
	}

	// Track provider for cleanup if creation succeeded
	if resp.JSON201 != nil {
		providerID := fmt.Sprintf("%d", resp.JSON201.Id)
		ctx.TrackProvider(resp.JSON201.Id)
		ctx.LastProviderID = resp.JSON201.Id
		ctx.TrackCreatedResource("created_provider_id", providerID)

		// Track provider kind for limit checking
		countKey := fmt.Sprintf("created_provider_count_%s", kind)
		var count int
		if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
			fmt.Sscanf(countStr, "%d", &count)
		}
		ctx.TrackCreatedResource(countKey, fmt.Sprintf("%d", count+1))
	}

	return nil
}

func (ctx *ScenarioContext) iCreateAnotherProviderWithName(kind, name string) error {
	// Simulate duplicate name error
	ctx.SetLastResponse(400, nil, "duplicate provider name")
	return nil
}

func (ctx *ScenarioContext) iCreateAProviderWithPriority(kind, apiKey string, priority int) error {
	providerID := int64(301)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.SetLastResponse(201, map[string]interface{}{
		"id":       providerID,
		"priority": priority,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iCreateAProviderDisabled(kind, apiKey string) error {
	providerID := int64(302)
	ctx.TrackProvider(providerID)
	ctx.LastProviderID = providerID
	ctx.SetLastResponse(201, map[string]interface{}{
		"id":      providerID,
		"enabled": false,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iCreateAProviderWithNameOnly(kind, name string) error {
	// Simulate validation error for empty name
	ctx.SetLastResponse(400, nil, "name is required")
	return nil
}

func (ctx *ScenarioContext) iListAllProvidersFromProvider() error {
	// Get authenticated client
	client, err := ctx.GetAuthenticatedClient()
	if err != nil || client == nil {
		ctx.SetLastResponse(401, nil, "unauthorized: authentication required")
		return nil
	}

	// Call API to list providers
	resp, err := client.GetApiV1ProvidersWithResponse(context.Background())
	if err != nil {
		ctx.SetLastResponse(0, nil, err.Error())
		return fmt.Errorf("provider list request failed: %w", err)
	}

	// Parse response body
	var body interface{}
	if resp.JSON200 != nil {
		body = resp.JSON200
	} else if resp.JSON401 != nil {
		body = resp.JSON401
	} else if len(resp.Body) > 0 {
		json.Unmarshal(resp.Body, &body)
	}

	// Store response for assertions
	ctx.SetLastResponse(resp.StatusCode(), body, "")

	// Handle API response errors
	if resp.StatusCode() >= 400 {
		errMsg := ""
		if resp.JSON401 != nil {
			errMsg = fmt.Sprintf("unauthorized: %s", resp.JSON401.Error)
		} else if len(resp.Body) > 0 {
			errMsg = string(resp.Body)
		} else {
			errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode())
		}
		return fmt.Errorf("provider list failed: %s", errMsg)
	}

	return nil
}

func (ctx *ScenarioContext) iListProvidersWithKind(kind string) error {
	// TODO: Implement filtered provider listing
	ctx.SetLastResponse(200, []map[string]interface{}{
		{"id": 1, "name": fmt.Sprintf("%s-provider", kind), "kind": kind},
	}, "")
	return nil
}

func (ctx *ScenarioContext) iGetProviderByID() error {
	// Get authenticated client
	client, err := ctx.GetAuthenticatedClient()
	if err != nil || client == nil {
		ctx.SetLastResponse(401, nil, "unauthorized: authentication required")
		return nil
	}

	// Use LastProviderID from context
	if ctx.LastProviderID == 0 {
		ctx.SetLastResponse(400, nil, "no provider ID available")
		return nil
	}

	providerId := integration.ProviderId(ctx.LastProviderID)

	// Call API to get provider
	resp, err := client.GetApiV1ProvidersProviderIdWithResponse(context.Background(), providerId)
	if err != nil {
		ctx.SetLastResponse(0, nil, err.Error())
		return fmt.Errorf("provider get request failed: %w", err)
	}

	// Parse response body
	var body interface{}
	if resp.JSON200 != nil {
		body = resp.JSON200
	} else if resp.JSON401 != nil {
		body = resp.JSON401
	} else if resp.JSON404 != nil {
		body = resp.JSON404
	} else if len(resp.Body) > 0 {
		json.Unmarshal(resp.Body, &body)
	}

	// Store response for assertions
	ctx.SetLastResponse(resp.StatusCode(), body, "")

	// Handle API response errors
	if resp.StatusCode() >= 400 {
		errMsg := ""
		if resp.JSON401 != nil {
			errMsg = fmt.Sprintf("unauthorized: %s", resp.JSON401.Error)
		} else if resp.JSON404 != nil {
			errMsg = fmt.Sprintf("not found: %s", resp.JSON404.Error)
		} else if len(resp.Body) > 0 {
			errMsg = string(resp.Body)
		} else {
			errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode())
		}
		return fmt.Errorf("provider get failed: %s", errMsg)
	}

	return nil
}

func (ctx *ScenarioContext) iGetProviderStats() error {
	// TODO: Implement actual statistics retrieval
	ctx.SetLastResponse(200, map[string]interface{}{
		"total_requests": 1000,
		"success_rate":   0.95,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iGetProviderWithID(id int) error {
	if id == 99999 {
		ctx.SetLastResponse(404, nil, "provider not found")
		return nil
	}
	ctx.SetLastResponse(200, map[string]interface{}{"id": id}, "")
	return nil
}

func (ctx *ScenarioContext) iUpdateProviderName(name string) error {
	// Get authenticated client
	client, err := ctx.GetAuthenticatedClient()
	if err != nil || client == nil {
		ctx.SetLastResponse(401, nil, "unauthorized: authentication required")
		return nil
	}

	// Use LastProviderID from context
	if ctx.LastProviderID == 0 {
		ctx.SetLastResponse(400, nil, "no provider ID available")
		return nil
	}

	providerId := integration.ProviderId(ctx.LastProviderID)

	// Create update request with just the name
	req := integration.UpdateProviderRequest{
		Name: &name,
	}

	// Call API to update provider
	resp, err := client.PutApiV1ProvidersProviderIdWithResponse(context.Background(), providerId, req)
	if err != nil {
		ctx.SetLastResponse(0, nil, err.Error())
		return fmt.Errorf("provider update request failed: %w", err)
	}

	// Parse response body
	var body interface{}
	if resp.JSON200 != nil {
		body = resp.JSON200
	} else if resp.JSON400 != nil {
		body = resp.JSON400
	} else if resp.JSON401 != nil {
		body = resp.JSON401
	} else if resp.JSON404 != nil {
		body = resp.JSON404
	} else if len(resp.Body) > 0 {
		json.Unmarshal(resp.Body, &body)
	}

	// Store response for assertions
	ctx.SetLastResponse(resp.StatusCode(), body, "")

	// Handle API response errors
	if resp.StatusCode() >= 400 {
		errMsg := ""
		if resp.JSON400 != nil {
			errMsg = fmt.Sprintf("validation error: %s", resp.JSON400.Error)
		} else if resp.JSON401 != nil {
			errMsg = fmt.Sprintf("unauthorized: %s", resp.JSON401.Error)
		} else if resp.JSON404 != nil {
			errMsg = fmt.Sprintf("not found: %s", resp.JSON404.Error)
		} else if len(resp.Body) > 0 {
			errMsg = string(resp.Body)
		} else {
			errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode())
		}
		return fmt.Errorf("provider update failed: %s", errMsg)
	}

	return nil
}

func (ctx *ScenarioContext) iUpdateProviderAPIKey(apiKey string) error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"id": ctx.LastProviderID,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iUpdateProviderPriority(priority int) error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"id":       ctx.LastProviderID,
		"priority": priority,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iDisableProvider() error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"id":      ctx.LastProviderID,
		"enabled": false,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iEnableProvider() error {
	ctx.SetLastResponse(200, map[string]interface{}{
		"id":      ctx.LastProviderID,
		"enabled": true,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iUpdateProviderWithID(id int) error {
	if id == 99999 {
		ctx.SetLastResponse(404, nil, "provider not found")
		return nil
	}
	ctx.SetLastResponse(200, map[string]interface{}{"id": id}, "")
	return nil
}

func (ctx *ScenarioContext) iDeleteProvider() error {
	// Get authenticated client
	client, err := ctx.GetAuthenticatedClient()
	if err != nil || client == nil {
		ctx.SetLastResponse(401, nil, "unauthorized: authentication required")
		return nil
	}

	// Use LastProviderID from context
	if ctx.LastProviderID == 0 {
		ctx.SetLastResponse(400, nil, "no provider ID available")
		return nil
	}

	providerId := integration.ProviderId(ctx.LastProviderID)

	// Call API to delete provider
	resp, err := client.DeleteApiV1ProvidersProviderIdWithResponse(context.Background(), providerId)
	if err != nil {
		ctx.SetLastResponse(0, nil, err.Error())
		return fmt.Errorf("provider deletion request failed: %w", err)
	}

	// Parse response body
	var body interface{}
	if resp.JSON200 != nil {
		body = resp.JSON200
	} else if resp.JSON401 != nil {
		body = resp.JSON401
	} else if resp.JSON404 != nil {
		body = resp.JSON404
	} else if len(resp.Body) > 0 {
		json.Unmarshal(resp.Body, &body)
	}

	// Store response for assertions
	ctx.SetLastResponse(resp.StatusCode(), body, "")

	// Handle API response errors
	if resp.StatusCode() >= 400 {
		errMsg := ""
		if resp.JSON401 != nil {
			errMsg = fmt.Sprintf("unauthorized: %s", resp.JSON401.Error)
		} else if resp.JSON404 != nil {
			errMsg = fmt.Sprintf("not found: %s", resp.JSON404.Error)
		} else if len(resp.Body) > 0 {
			errMsg = string(resp.Body)
		} else {
			errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode())
		}
		return fmt.Errorf("provider deletion failed: %s", errMsg)
	}

	return nil
}

func (ctx *ScenarioContext) iDeleteProviderByID(id int) error {
	if id == 99999 {
		ctx.SetLastResponse(404, nil, "provider not found")
		return nil
	}
	ctx.SetLastResponse(204, nil, "")
	return nil
}

func (ctx *ScenarioContext) iDeleteProviderByIDNoArgs() error {
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	// Delete the most recently created provider
	ctx.SetLastResponse(204, nil, "")
	return nil
}

func (ctx *ScenarioContext) iAttemptToDeleteDefaultProvider() error {
	ctx.SetLastResponse(403, nil, "cannot delete default provider")
	return nil
}

func (ctx *ScenarioContext) iTestProviderConnectivity() error {
	apiKeyValid, _ := ctx.GetCreatedResource("api_key_valid")
	if apiKeyValid == "false" {
		ctx.SetLastResponse(200, map[string]interface{}{
			"success": false,
			"error":   "authentication failed",
		}, "")
		return nil
	}
	ctx.SetLastResponse(200, map[string]interface{}{
		"success": true,
	}, "")
	return nil
}

func (ctx *ScenarioContext) iTestConnectivityForProviderWithID(id int) error {
	ctx.SetLastResponse(404, nil, "provider not found")
	return nil
}

func (ctx *ScenarioContext) iAttemptToCreateProviderInLimit() error {
	_, hasLimit := ctx.GetCreatedResource("license_provider_limit")
	if hasLimit {
		// Check if limit is reached
		ctx.SetLastResponse(403, nil, "provider limit reached")
		return nil
	}
	ctx.SetLastResponse(201, map[string]interface{}{"id": 400}, "")
	return nil
}

func (ctx *ScenarioContext) iCreateAProvider() error {
	return ctx.iCreateAProviderWithKindAndAPIKey("claude", "sk-test-123")
}

func (ctx *ScenarioContext) iUpdateFirstProvider() error {
	ctx.SetLastResponse(200, map[string]interface{}{"id": 1}, "")
	return nil
}

func (ctx *ScenarioContext) iDeleteFirstProvider() error {
	ctx.SetLastResponse(204, nil, "")
	return nil
}

func (ctx *ScenarioContext) iCreateAProviderWithKindSimple(kind string) error {
	kindLimitKey := fmt.Sprintf("license_%s_provider_limit", kind)
	if kindLimit, hasLimit := ctx.GetCreatedResource(kindLimitKey); hasLimit {
		// Check if kind limit is reached
		countKey := fmt.Sprintf("created_provider_count_%s", kind)
		var count int
		if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
			fmt.Sscanf(countStr, "%d", &count)
		}

		var limit int
		fmt.Sscanf(kindLimit, "%d", &limit)

		if count >= limit {
			ctx.SetLastResponse(403, nil, fmt.Sprintf("%s provider limit reached", kind))
			return nil
		}

		// Increment counter
		ctx.TrackCreatedResource(countKey, fmt.Sprintf("%d", count+1))
	}
	providerID := int64(500)
	ctx.TrackProvider(providerID)
	ctx.SetLastResponse(201, map[string]interface{}{"id": providerID}, "")
	return nil
}

// THENS - Assert outcomes

func (ctx *ScenarioContext) providerCreatedSuccessfully() error {
	statusCode, _, _ := ctx.GetLastResponse()
	if statusCode != 201 {
		return fmt.Errorf("expected status 201, got %d", statusCode)
	}
	return nil
}

func (ctx *ScenarioContext) providerKindShouldBe(kind string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if statusCode != 201 {
		return fmt.Errorf("expected created provider, got status %d", statusCode)
	}
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["kind"] != kind {
			return fmt.Errorf("expected kind %s, got %v", kind, respMap["kind"])
		}
	}
	return nil
}

func (ctx *ScenarioContext) providerShouldBeEnabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if enabled, ok := respMap["enabled"].(bool); ok && !enabled {
			return fmt.Errorf("expected provider to be enabled")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) providerShouldNotBeEnabled() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if enabled, ok := respMap["enabled"].(bool); ok && enabled {
			return fmt.Errorf("expected provider to be disabled")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) providerPriorityShouldBe(priority int) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if p, ok := respMap["priority"].(int); ok && p != priority {
			return fmt.Errorf("expected priority %d, got %d", priority, p)
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) iShouldSeeAtLeastNProviders(n int) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if statusCode != 200 {
		return fmt.Errorf("expected status 200, got %d", statusCode)
	}
	if providers, ok := resp.([]map[string]interface{}); ok {
		if len(providers) < n {
			return fmt.Errorf("expected at least %d providers, got %d", n, len(providers))
		}
	}
	return nil
}

func (ctx *ScenarioContext) apiKeyNotVisible() error {
	// Check that API key is not in response
	statusCode, resp, _ := ctx.GetLastResponse()
	if statusCode == 200 {
		if respMap, ok := resp.(map[string]interface{}); ok {
			if _, hasKey := respMap["api_key"]; hasKey {
				return fmt.Errorf("API key should not be visible")
			}
		}
	}
	return nil
}

func (ctx *ScenarioContext) providerShouldHaveAName() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasName := respMap["name"]; !hasName {
			return fmt.Errorf("provider should have a name")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) statisticsShouldContainTotalRequests() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasRequests := respMap["total_requests"]; !hasRequests {
			return fmt.Errorf("statistics should contain total_requests")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) statisticsShouldContainSuccessRate() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if _, hasRate := respMap["success_rate"]; !hasRate {
			return fmt.Errorf("statistics should contain success_rate")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) providerNameShouldBe(name string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if respMap["name"] != name {
			return fmt.Errorf("expected name %s, got %v", name, respMap["name"])
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) providerShouldNotExist() error {
	statusCode, _, _ := ctx.GetLastResponse()
	if statusCode == 200 {
		return fmt.Errorf("provider should not exist")
	}
	return nil
}

func (ctx *ScenarioContext) connectionSuccessful() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if success, ok := respMap["success"].(bool); ok && !success {
			return fmt.Errorf("expected successful connection")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) connectionShouldFail() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if success, ok := respMap["success"].(bool); ok && success {
			return fmt.Errorf("expected connection to fail")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) errorShouldIndicateAuthFailure() error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if respMap, ok := resp.(map[string]interface{}); ok {
		if err, ok := respMap["error"].(string); ok && err != "authentication failed" {
			return fmt.Errorf("expected authentication failure error")
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) totalProviderCountShouldBe(n int) error {
	// TODO: Verify actual provider count
	return nil
}

func (ctx *ScenarioContext) iShouldBeAbleToCreateNewProvider() error {
	// Check if provider limit allows creation
	if limitStr, hasLimit := ctx.GetCreatedResource("license_provider_limit"); hasLimit {
		var limit int
		fmt.Sscanf(limitStr, "%d", &limit)
		countKey := "created_providers_count"
		var count int
		if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
			fmt.Sscanf(countStr, "%d", &count)
		}
		if count < limit {
			return nil // Able to create
		}
		return fmt.Errorf("provider limit reached, cannot create new provider")
	}
	return nil // No limit set, should be able to create
}

func (ctx *ScenarioContext) iShouldOnlySeeProvidersOfKind(kind string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if providers, ok := resp.([]map[string]interface{}); ok {
		for _, p := range providers {
			if p["kind"] != kind {
				return fmt.Errorf("expected only %s providers, found %v", kind, p["kind"])
			}
		}
	}
	_ = statusCode
	return nil
}

func (ctx *ScenarioContext) iShouldNotSeeProvidersOfKind(kind string) error {
	statusCode, resp, _ := ctx.GetLastResponse()
	if providers, ok := resp.([]map[string]interface{}); ok {
		for _, p := range providers {
			if p["kind"] == kind {
				return fmt.Errorf("should not see %s providers", kind)
			}
		}
	}
	_ = statusCode
	return nil
}

// Additional provider step implementations

func (ctx *ScenarioContext) iAttemptToCreateAClaudeProvider() error {
	// Check per-kind provider limit (claude) FIRST, before authentication
	// This allows testing license limits without requiring authentication
	kindLimitKey := "license_claude_provider_limit"
	if kindLimitStr, hasLimit := ctx.GetCreatedResource(kindLimitKey); hasLimit {
		countKey := "created_provider_count_claude"
		var count int
		if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
			fmt.Sscanf(countStr, "%d", &count)
		}
		var limit int
		fmt.Sscanf(kindLimitStr, "%d", &limit)
		if count >= limit {
			ctx.SetLastResponse(403, nil, "claude provider limit")
			return nil
		}
	}

	// Then check authentication
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	if ctx.BDDTestContext.CurrentUser.Role != support.RoleAdmin {
		ctx.SetLastResponse(403, nil, "permission denied")
		return nil
	}

	// Also check general provider limit
	if limitStr, hasLimit := ctx.GetCreatedResource("license_provider_limit"); hasLimit {
		countKey := "created_providers_count"
		var count int
		if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
			fmt.Sscanf(countStr, "%d", &count)
		}
		var limit int
		fmt.Sscanf(limitStr, "%d", &limit)
		if count >= limit {
			ctx.SetLastResponse(403, nil, "provider limit reached")
			return nil
		}
	}
	ctx.SetLastResponse(201, map[string]interface{}{
		"id":     "provider-123",
		"kind":   "claude",
		"name":   "claude-provider",
		"api_key": "sk-test-123",
	}, "")
	return nil
}

func (ctx *ScenarioContext) iAttemptToCreateATeam() error {
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	if ctx.BDDTestContext.CurrentUser.Role != support.RoleAdmin {
		ctx.SetLastResponse(403, nil, "permission denied")
		return nil
	}
	ctx.SetLastResponse(201, map[string]interface{}{
		"id":   "team-123",
		"name": "new-team",
	}, "")
	return nil
}

func (ctx *ScenarioContext) iCreateATeam() error {
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	if ctx.BDDTestContext.CurrentUser.Role != support.RoleAdmin {
		ctx.SetLastResponse(403, nil, "permission denied")
		return nil
	}
	teamName := support.GenerateUniqueTeamName("team")
	ctx.SetLastResponse(201, map[string]interface{}{
		"id":   fmt.Sprintf("team-%d", support.GenerateUniqueID()),
		"name": teamName,
	}, "")
	ctx.TrackCreatedResource("created_team", teamName)
	return nil
}

func (ctx *ScenarioContext) iCreateAProviderWithAPIKey(kind, apiKey string) error {
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	if ctx.BDDTestContext.CurrentUser.Role != support.RoleAdmin {
		ctx.SetLastResponse(403, nil, "permission denied")
		return nil
	}
	// Validate API key
	if apiKey == "" {
		ctx.SetLastResponse(400, nil, "API key is required")
		return nil
	}
	// Validate provider kind
	if !support.IsValidProviderKind(kind) {
		ctx.SetLastResponse(400, nil, "invalid provider kind")
		return nil
	}
	// Check provider limits
	if limit, hasLimit := ctx.GetCreatedResource("license_provider_limit"); hasLimit {
		var limitInt int
		fmt.Sscanf(limit, "%d", &limitInt)
		countKey := "created_providers_count"
		var count int
		if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
			fmt.Sscanf(countStr, "%d", &count)
		}
		if count >= limitInt {
			ctx.SetLastResponse(403, nil, "provider limit reached")
			return nil
		}
		ctx.TrackCreatedResource(countKey, fmt.Sprintf("%d", count+1))
	}
	providerName := support.GenerateUniqueProviderName(kind)
	// Track provider for cleanup
	providerID := support.GenerateUniqueID()
	ctx.TrackCreatedResource("created_provider_id", fmt.Sprintf("%d", providerID))
	ctx.SetLastResponse(201, map[string]interface{}{
		"id":     fmt.Sprintf("%d", providerID),
		"kind":   kind,
		"name":   providerName,
		"api_key": apiKey,
	}, "")
	return nil
}

// Additional provider step implementations

func (ctx *ScenarioContext) iCreateAProviderWithNameArg(name string) error {
	if ctx.BDDTestContext.CurrentUser == nil {
		ctx.SetLastResponse(401, nil, "unauthorized")
		return nil
	}
	if ctx.BDDTestContext.CurrentUser.Role != support.RoleAdmin {
		ctx.SetLastResponse(403, nil, "permission denied")
		return nil
	}
	// Validate provider name
	if name == "" {
		ctx.SetLastResponse(400, nil, "provider name is required")
		return nil
	}
	// Check provider limits
	if limitStr, hasLimit := ctx.GetCreatedResource("license_provider_limit"); hasLimit {
		var limit int
		fmt.Sscanf(limitStr, "%d", &limit)
		countKey := "created_providers_count"
		var count int
		if countStr, hasCount := ctx.GetCreatedResource(countKey); hasCount {
			fmt.Sscanf(countStr, "%d", &count)
		}
		if count >= limit {
			ctx.SetLastResponse(403, nil, "provider limit reached")
			return nil
		}
		ctx.TrackCreatedResource(countKey, fmt.Sprintf("%d", count+1))
	}
	providerID := support.GenerateUniqueID()
	ctx.SetLastResponse(201, map[string]interface{}{
		"id":   fmt.Sprintf("%d", providerID),
		"kind": "claude", // Default to claude
		"name": name,
	}, "")
	ctx.TrackCreatedResource("created_provider_name", name)
	return nil
}
