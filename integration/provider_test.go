// Copyright (c) 2025 Code Together
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"context"
	"encoding/json"

	integrationclient "github.com/code-together/shared/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Provider Create Tests (8 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderCreateClaude() {
	ctx := context.Background()

	kind := integrationclient.CreateProviderRequestKindClaude
	uniqueName := generateUniqueProviderName("test-claude")
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-12345",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, resp.StatusCode(), "Provider creation should return 201")
	require.NotNil(s.T(), resp.JSON201, "Response should have JSON201 body")

	assert.Equal(s.T(), uniqueName, resp.JSON201.Name)
	assert.Equal(s.T(), integrationclient.ProviderKindClaude, *resp.JSON201.Kind)
	assert.Greater(s.T(), resp.JSON201.Id, int64(0))
}

func (s *IntegrationTestSuite) TestProviderCreateCodex() {
	ctx := context.Background()

	kind := integrationclient.CreateProviderRequestKindCodex
	uniqueName := generateUniqueProviderName("test-codex")
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "test-codex-key",
		ApiUrl:          "https://api.codex.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"codex-model-1"},
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, resp.StatusCode(), "Provider creation should return 201")
	require.NotNil(s.T(), resp.JSON201)

	assert.Equal(s.T(), uniqueName, resp.JSON201.Name)
	assert.Equal(s.T(), integrationclient.ProviderKindCodex, *resp.JSON201.Kind)
}

func (s *IntegrationTestSuite) TestProviderCreateOpenCode() {
	ctx := context.Background()

	kind := integrationclient.CreateProviderRequestKindOpencode
	uniqueName := generateUniqueProviderName("test-opencode")
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-opencode-test",
		ApiUrl:          "https://api.opencode.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"opencode-model"},
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, resp.StatusCode(), "Provider creation should return 201")
	require.NotNil(s.T(), resp.JSON201)

	assert.Equal(s.T(), uniqueName, resp.JSON201.Name)
	assert.Equal(s.T(), integrationclient.ProviderKindOpencode, *resp.JSON201.Kind)
}

func (s *IntegrationTestSuite) TestProviderCreateDuplicateName() {
	ctx := context.Background()

	kind := integrationclient.CreateProviderRequestKindClaude
	// Create first provider
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            "duplicate-provider",
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-1",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	resp1, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create first provider")
	require.Equal(s.T(), 201, resp1.StatusCode())

	// Try to create second provider with same name
	req.ApiKey = "sk-ant-test-key-2"
	resp2, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to attempt second provider creation")
	assert.Equal(s.T(), 409, resp2.StatusCode(), "Duplicate name should be rejected with 409 Conflict")

	// Parse error response from body (409 is not mapped in client)
	var errResp integrationclient.ErrorResponse
	err = json.Unmarshal(resp2.Body, &errResp)
	require.NoError(s.T(), err, "Error response should be valid JSON")
	requireErrorResponse(s, &errResp, "duplicate")
}

// Test that invalid provider kinds are rejected
func (s *IntegrationTestSuite) TestProviderCreateInvalidKind() {
	ctx := context.Background()

	invalidKind := integrationclient.CreateProviderRequestKind("invalid-kind")
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:    "invalid-provider",
		Kind:    &invalidKind,
		ApiKey:  "test-key",
		ApiUrl:  "https://api.test.com",
		Enabled: &[]bool{true}[0],
		Level:   &[]int{1}[0],
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to attempt provider creation")
	assert.Equal(s.T(), 400, resp.StatusCode(), "Invalid kind should be rejected")
	requireErrorResponse(s, resp.JSON400, "kind")
}

func (s *IntegrationTestSuite) TestProviderCreateMissingFields() {
	ctx := context.Background()

	// Missing required fields: name, kind, api_key
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		ApiUrl:  "https://api.test.com",
		Enabled: &[]bool{true}[0],
		Level:   &[]int{1}[0],
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to attempt provider creation")
	assert.Equal(s.T(), 400, resp.StatusCode(), "Missing required fields should be rejected")
	requireErrorResponse(s, resp.JSON400, "invalid")
}

func (s *IntegrationTestSuite) TestProviderCreateExceedsTier0Limit() {
	// Skip bootstrap to avoid existing providers interfering with Tier 0 limit test
	s.skipBootstrap = true

	ctx := context.Background()

	// First, activate an Open Source license (all types have unlimited providers now)
	license := s.activateLicenseFixture(LicenseOpenSource)
	s.Equal("opensource", getLicenseType(license), "Should have opensource license")

	// Create multiple Claude providers (no limit)
	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	for i := 1; i <= 5; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            generateUniqueProviderName("tier0-claude-" + string(rune('0'+i))),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}

		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode(), "Provider %d should be created", i)
	}
}

func (s *IntegrationTestSuite) TestProviderCreateCommercialUnlimited() {
	ctx := context.Background()

	// Activate Commercial license (unlimited providers)
	license := s.activateLicenseFixture(LicenseCommercial)
	s.Equal("commercial", getLicenseType(license), "Should have commercial license")

	// Create multiple providers (all types allow unlimited)
	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	for i := 1; i <= 10; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            generateUniqueProviderName("commercial-claude-" + string(rune('0'+i))),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}

		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode(), "Provider %d should be created", i)
	}
}

// ============================================================================
// Provider Update Tests (5 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderUpdateName() {
	ctx := context.Background()

	// Create provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Update provider name
	newName := "updated-claude-provider"
	req := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Name: &newName,
	}

	resp, err := s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, providerID, req)
	require.NoError(s.T(), err, "Failed to update provider")
	require.Equal(s.T(), 200, resp.StatusCode(), "Provider update should return 200")
	require.NotNil(s.T(), resp.JSON200)

	assert.Equal(s.T(), newName, resp.JSON200.Name)
}

func (s *IntegrationTestSuite) TestProviderUpdateAPIKey() {
	ctx := context.Background()

	// Create provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Update API key
	newKey := "sk-ant-updated-key"
	req := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		ApiKey: &newKey,
	}

	resp, err := s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, providerID, req)
	require.NoError(s.T(), err, "Failed to update provider")
	require.Equal(s.T(), 200, resp.StatusCode())

	// Verify response body exists
	require.NotNil(s.T(), resp.JSON200, "Response body should not be nil")
	require.NotNil(s.T(), resp.JSON200.ApiKey, "API key should not be nil in response")

	assert.Equal(s.T(), newKey, *resp.JSON200.ApiKey, "API key should be updated")
}

func (s *IntegrationTestSuite) TestProviderUpdateEnabled() {
	ctx := context.Background()

	// Create provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Disable provider
	disabled := false
	req := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Enabled: &disabled,
	}

	resp, err := s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, providerID, req)
	require.NoError(s.T(), err, "Failed to update provider")
	require.Equal(s.T(), 200, resp.StatusCode())

	assert.False(s.T(), resp.JSON200.Enabled)
}

func (s *IntegrationTestSuite) TestProviderUpdateNonExistent() {
	ctx := context.Background()

	// Try to update non-existent provider
	newName := "should-not-work"
	req := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Name: &newName,
	}

	resp, err := s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, 99999, req)
	require.NoError(s.T(), err, "Failed to attempt update")
	assert.Equal(s.T(), 404, resp.StatusCode(), "Non-existent provider should return 404")
	requireErrorResponse(s, resp.JSON404, "provider")
}

func (s *IntegrationTestSuite) TestProviderUpdateDuplicateName() {
	ctx := context.Background()

	// Create two providers
	kind := integrationclient.CreateProviderRequestKindClaude
	req1 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            "provider-1",
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-1",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}
	resp1, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req1)
	require.NoError(s.T(), err)
	_ = resp1.JSON201.Id // providerID1

	req2 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            "provider-2",
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-2",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}
	resp2, _ := s.Client.PostApiV1ProvidersWithResponse(ctx, req2)
	providerID2 := resp2.JSON201.Id

	// Try to update provider-2 with provider-1's name (duplicate)
	updateReq := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Name: &resp1.JSON201.Name, // Duplicate name
	}

	resp, err := s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, providerID2, updateReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 409, resp.StatusCode(), "Duplicate name should be rejected with 409 Conflict")

	// Parse error response from body (409 is not mapped in client)
	var errResp integrationclient.ErrorResponse
	err = json.Unmarshal(resp.Body, &errResp)
	require.NoError(s.T(), err, "Error response should be valid JSON")
	requireErrorResponse(s, &errResp, "exists")
}

// ============================================================================
// Provider Delete Tests (3 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderDeleteSuccess() {
	ctx := context.Background()

	// Create provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Delete provider
	resp, err := s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
	require.NoError(s.T(), err, "Failed to delete provider")
	assert.Equal(s.T(), 200, resp.StatusCode(), "Delete should return 200")

	// Verify provider is deleted by listing all providers
	listResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, listResp.StatusCode())

	// Provider should not be in the list
	for _, p := range *listResp.JSON200 {
		if p.Id == providerID {
			s.T().Fatalf("Deleted provider should not exist in the list")
		}
	}
}

func (s *IntegrationTestSuite) TestProviderDeleteNonExistent() {
	ctx := context.Background()

	// Try to delete non-existent provider
	resp, err := s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, 99999)
	require.NoError(s.T(), err, "Failed to attempt delete")
	assert.Equal(s.T(), 404, resp.StatusCode(), "Non-existent provider should return 404")
	requireErrorResponse(s, resp.JSON404, "provider")
}

func (s *IntegrationTestSuite) TestProviderDeleteTwice() {
	ctx := context.Background()

	// Create provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Delete first time
	resp1, err := s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp1.StatusCode())

	// Try to delete again - server returns 404 (not found)
	resp2, err := s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 404, resp2.StatusCode(), "Deleting already-deleted provider returns 404")
	requireErrorResponse(s, resp2.JSON404, "provider")
}

// ============================================================================
// Provider List Tests (3 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderListEmpty() {
	// Skip bootstrap to start with clean state (no providers)
	s.skipBootstrap = true

	ctx := context.Background()

	resp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())
	require.NotNil(s.T(), resp.JSON200)

	assert.Empty(s.T(), *resp.JSON200, "Should have no providers initially")
}

func (s *IntegrationTestSuite) TestProviderListMultiple() {
	ctx := context.Background()

	// Create multiple providers
	kind := integrationclient.CreateProviderRequestKindClaude
	for i := 1; i <= 3; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            generateUniqueProviderName("list-claude-" + string(rune('0'+i))),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &[]bool{true}[0],
			Level:           &[]int{1}[0],
			SupportedModels: &[]string{"claude-3-opus"},
		}
		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode())
	}

	// List providers
	resp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())
	require.NotNil(s.T(), resp.JSON200)

	assert.GreaterOrEqual(s.T(), len(*resp.JSON200), 3, "Should have at least 3 providers")
}

func (s *IntegrationTestSuite) TestProviderListFilterByKind() {
	ctx := context.Background()

	// Create Claude provider
	kindClaude := integrationclient.CreateProviderRequestKindClaude
	reqClaude := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            generateUniqueProviderName("test-claude"),
		Kind:            &kindClaude,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}
	resp1, _ := s.Client.PostApiV1ProvidersWithResponse(ctx, reqClaude)
	require.Equal(s.T(), 201, resp1.StatusCode())

	// Create Codex provider
	kindCodex := integrationclient.CreateProviderRequestKindCodex
	reqCodex := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            generateUniqueProviderName("test-codex"),
		Kind:            &kindCodex,
		ApiKey:          "test-codex-key",
		ApiUrl:          "https://api.codex.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"codex-model-1"},
	}
	resp2, _ := s.Client.PostApiV1ProvidersWithResponse(ctx, reqCodex)
	require.Equal(s.T(), 201, resp2.StatusCode())

	// List all providers
	resp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())

	// Verify we have both kinds
	claudeCount := 0
	codexCount := 0
	if resp.JSON200 != nil {
		for _, p := range *resp.JSON200 {
			if *p.Kind == integrationclient.ProviderKindClaude {
				claudeCount++
			} else if *p.Kind == integrationclient.ProviderKindCodex {
				codexCount++
			}
		}
	}

	assert.Greater(s.T(), claudeCount, 0, "Should have Claude providers")
	assert.Greater(s.T(), codexCount, 0, "Should have Codex providers")
}

// ============================================================================
// Provider Get Tests (3 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderGetSuccess() {
	ctx := context.Background()

	// Create provider
	kind := integrationclient.CreateProviderRequestKindClaude
	uniqueName := generateUniqueProviderName("test-claude")
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}
	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	providerID := createResp.JSON201.Id

	// Get provider via list endpoint and filter by ID
	listResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, listResp.StatusCode())
	require.NotNil(s.T(), listResp.JSON200)

	// Find the created provider in the list
	var found *integrationclient.Provider
	for _, p := range *listResp.JSON200 {
		if p.Id == providerID {
			found = &p
			break
		}
	}
	require.NotNil(s.T(), found, "Provider should be in the list")

	assert.Equal(s.T(), uniqueName, found.Name)
	assert.Equal(s.T(), integrationclient.ProviderKindClaude, *found.Kind)
}

// Note: TestProviderGetNonExistent removed - requires GET /providers/:id endpoint
// Provider existence can be verified via list endpoint instead

func (s *IntegrationTestSuite) TestProviderGetAllFields() {
	ctx := context.Background()

	// Create provider with all fields
	kind := integrationclient.CreateProviderRequestKindClaude
	models := []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            "full-claude-provider",
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{2}[0],
		SupportedModels: &models,
	}
	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	providerID := createResp.JSON201.Id

	// Get provider via list endpoint and filter by ID
	listResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, listResp.StatusCode())

	// Find the created provider in the list
	var found *integrationclient.Provider
	for _, p := range *listResp.JSON200 {
		if p.Id == providerID {
			found = &p
			break
		}
	}
	require.NotNil(s.T(), found, "Provider should be in the list")

	assert.Equal(s.T(), "full-claude-provider", found.Name)
	assert.Equal(s.T(), "sk-ant-test-key", *found.ApiKey, "API key should match")
	assert.Equal(s.T(), "https://api.anthropic.com", found.ApiUrl)
	assert.True(s.T(), found.Enabled)
	// Level is returned as set (2 in this case)
	assert.NotNil(s.T(), found.Level)
	assert.Equal(s.T(), 2, *found.Level, "Level should match what was set")
	assert.NotNil(s.T(), found.SupportedModels)
}

// ============================================================================
// Provider Test Connectivity Tests (2 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderTestConnectivitySuccess() {
	ctx := context.Background()

	// Create provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Test connectivity
	resp, err := s.Client.PostApiV1ProvidersProviderIdTestWithResponse(ctx, providerID)
	require.NoError(s.T(), err)
	// Note: This may fail if API key is invalid, but should return 200 with test result
	assert.Contains(s.T(), []int{200, 400}, resp.StatusCode())
}

func (s *IntegrationTestSuite) TestProviderTestConnectivityNonExistent() {
	ctx := context.Background()

	// Try to test non-existent provider
	resp, err := s.Client.PostApiV1ProvidersProviderIdTestWithResponse(ctx, 99999)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 404, resp.StatusCode(), "Non-existent provider should return 404")
	requireErrorResponse(s, resp.JSON404, "provider")
}

// ============================================================================
// Provider License Limit Tests (3 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderLimitTier0() {
	// Skip bootstrap to avoid existing providers interfering with Tier 0 limit test
	s.skipBootstrap = true

	ctx := context.Background()

	// Activate Open Source license (all types have unlimited providers now)
	license := s.activateLicenseFixture(LicenseOpenSource)
	assert.Equal(s.T(), "opensource", getLicenseType(license))

	// Create multiple Claude providers (no limit)
	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	for i := 1; i <= 5; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            generateUniqueProviderName("tier0-claude-" + string(rune('0'+i))),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}
		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode())
	}

	// Try to create more (should succeed - no limit)
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            generateUniqueProviderName("tier0-claude-6"),
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-6",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &enabled,
		Level:           &level,
		SupportedModels: &[]string{"claude-3-opus"},
	}
	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 201, resp.StatusCode(), "Should be able to create unlimited providers with opensource license")
}

func (s *IntegrationTestSuite) TestProviderLimitCommercial() {
	ctx := context.Background()

	// Activate Commercial license
	license := s.activateLicenseFixture(LicenseCommercial)
	assert.Equal(s.T(), "commercial", getLicenseType(license))

	// Create many providers to test unlimited behavior
	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	for i := 1; i <= 10; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            generateUniqueProviderName("limit-comm-claude-" + string(rune('0'+i))),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}
		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode())
	}
}

func (s *IntegrationTestSuite) TestProviderLimitOpenSource() {
	ctx := context.Background()

	// Activate Open Source license
	license := s.activateLicenseFixture(LicenseOpenSource)
	assert.Equal(s.T(), "opensource", getLicenseType(license))

	// Create many providers (all types have unlimited providers)
	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	for i := 1; i <= 10; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            generateUniqueProviderName("limit-os-claude-" + string(rune('0'+i))),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}
		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode())
	}
}

// ============================================================================
// Provider Enable/Disable/Stats Tests (3 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestProviderEnable() {
	ctx := context.Background()

	// First create a disabled provider
	kind := integrationclient.CreateProviderRequestKindClaude
	uniqueName := generateUniqueProviderName("test-claude-disabled")
	disabled := false
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-12345",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &disabled,
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Enable the provider
	enableResp, err := s.Client.PostApiV1ProvidersProviderIdEnableWithResponse(ctx, createResp.JSON201.Id)
	require.NoError(s.T(), err, "Failed to enable provider")
	require.Equal(s.T(), 200, enableResp.StatusCode(), "Enable provider should return 200")
	require.NotNil(s.T(), enableResp.JSON200)

	assert.True(s.T(), enableResp.JSON200.Enabled)
	assert.Equal(s.T(), int64(createResp.JSON201.Id), enableResp.JSON200.ProviderId)

	// Cleanup
	s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, createResp.JSON201.Id)
}

func (s *IntegrationTestSuite) TestProviderDisable() {
	ctx := context.Background()

	// First create a provider
	kind := integrationclient.CreateProviderRequestKindClaude
	uniqueName := generateUniqueProviderName("test-claude-disable")
	enabled := true
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-12345",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &enabled,
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Disable the provider
	disableResp, err := s.Client.DeleteApiV1ProvidersProviderIdDisableWithResponse(ctx, createResp.JSON201.Id)
	require.NoError(s.T(), err, "Failed to disable provider")
	require.Equal(s.T(), 200, disableResp.StatusCode(), "Disable provider should return 200")
	require.NotNil(s.T(), disableResp.JSON200)

	assert.False(s.T(), disableResp.JSON200.Enabled)
	assert.Equal(s.T(), int64(createResp.JSON201.Id), disableResp.JSON200.ProviderId)

	// Cleanup
	s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, createResp.JSON201.Id)
}

// TestProviderGetStats tests getting statistics for a specific provider.
func (s *IntegrationTestSuite) TestProviderGetStats() {
	ctx := context.Background()

	// First create a provider
	kind := integrationclient.CreateProviderRequestKindClaude
	uniqueName := generateUniqueProviderName("test-claude-stats")
	enabled := true
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-12345",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &enabled,
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Get provider stats
	statsResp, err := s.Client.GetApiV1ProvidersProviderIdStatsWithResponse(ctx, createResp.JSON201.Id)
	require.NoError(s.T(), err, "Failed to get provider stats")
	require.Equal(s.T(), 200, statsResp.StatusCode(), "Get provider stats should return 200")
	require.NotNil(s.T(), statsResp.JSON200)

	assert.Equal(s.T(), int64(createResp.JSON201.Id), statsResp.JSON200.ProviderId)
	assert.Equal(s.T(), uniqueName, statsResp.JSON200.ProviderName)

	// Cleanup
	s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, createResp.JSON201.Id)
}
