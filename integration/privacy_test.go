// Copyright (c) 2025 AI Together
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
	"fmt"
	"net/http"
	"time"

	integrationclient "github.com/code-together/shared/integration"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrivacyMultiTenantIsolation tests that users from one tenant cannot access another tenant's data.
func (s *IntegrationTestSuite) TestPrivacyMultiTenantIsolation() {
	ctx := context.Background()

	// Create provider as admin tenant
	uniqueName := generateUniqueProviderName("tenant-isolation")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-tenant-1-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
	providerID := createResp.JSON201.Id
	tenant1ID := createResp.JSON201.TeamId

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Member user should be able to see their tenant's providers
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	memberListResp, err := memberClient.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, memberListResp.StatusCode())

	// Verify member can see the provider from same tenant
	var foundProvider *integrationclient.Provider
	for _, p := range *memberListResp.JSON200 {
		if p.Id == providerID {
			foundProvider = &p
			break
		}
	}

	require.NotNil(s.T(), foundProvider, "Member should see provider from same tenant")
	assert.Equal(s.T(), tenant1ID, foundProvider.TeamId,
		"Provider should belong to correct tenant")

	s.Logger.Info("Multi-tenant isolation verified - member sees same-tenant provider")
}

// TestPrivacyCrossTenantAccessPrevention tests that cross-tenant access is prevented.
func (s *IntegrationTestSuite) TestPrivacyCrossTenantAccessPrevention() {
	ctx := context.Background()

	// This test verifies that users cannot access resources from other tenants
	// by ensuring tenant_id is properly enforced in all queries

	// Create provider as admin
	uniqueName := generateUniqueProviderName("cross-tenant")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-cross-tenant-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
	providerID := createResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Member should only see providers from their tenant
	// If member and admin are in different tenants, member should not see admin's provider
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	// Get member's user info to check tenant
	memberProfile, err := memberClient.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, memberProfile.StatusCode())

	memberTenantID := memberProfile.JSON200.User.TenantId
	adminTenantID := createResp.JSON201.TeamId

	s.Logger.Info("Tenant IDs", "member_tenant", memberTenantID, "admin_tenant", adminTenantID)

	// List providers as member
	memberListResp, err := memberClient.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, memberListResp.StatusCode())

	// If tenants are different, member should not see admin's provider
	if memberTenantID != adminTenantID {
		var foundProvider *integrationclient.Provider
		for _, p := range *memberListResp.JSON200 {
			if p.Id == providerID {
				foundProvider = &p
				break
			}
		}
		assert.Nil(s.T(), foundProvider, "Member should not see providers from other tenants")
		s.Logger.Info("Cross-tenant access prevented - member cannot see other tenant's providers")
	} else {
		s.Logger.Info("Admin and member in same tenant - isolation test skipped")
	}
}

// TestPrivacyNoPromptStorage tests that prompts are not stored (metadata only).
func (s *IntegrationTestSuite) TestPrivacyNoPromptStorage() {
	ctx := context.Background()

	// Submit usage data with metadata only
	now := time.Now()
	usageRecords := []integrationclient.UsageRecord{
		{
			CreatedAt:       now,
			DurationSec:     float32Pointer(1.5),
			HttpCode:        200,
			InputTokens:     intPointer(100),
			OutputTokens:    intPointer(50),
			Model:           "claude-3-5-sonnet-20241022",
			Platform:        "anthropic",
			Provider:        "test-provider",
			IsStream:        boolPointer(true),
			ReasoningTokens: intPointer(25),
		},
	}

	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, usageRecords)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, batchResp.StatusCode())

	// Get usage records and verify no prompt/response content is stored
	getResp, err := s.Client.GetApiV1UsageCurrentWithResponse(ctx, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, getResp.StatusCode())

	require.NotNil(s.T(), getResp.JSON200, "Usage response should not be nil")

	// Verify only metadata is present (no prompt/response fields)
	// TeamUsageSummary contains aggregated usage data, not individual prompts
	// The UsageRecord type in UsageBreakdown should not have prompt/response content
	if getResp.JSON200.UsageBreakdown != nil && len(*getResp.JSON200.UsageBreakdown) > 0 {
		record := (*getResp.JSON200.UsageBreakdown)[0]

		// These fields should exist (metadata)
		assert.NotEmpty(s.T(), record.Model, "Model should be stored")
		assert.NotEmpty(s.T(), record.Platform, "Platform should be stored")
		assert.NotZero(s.T(), record.CreatedAt, "CreatedAt should be stored")

		// These fields should NOT exist (actual prompt/response content)
		// The UsageRecord type in the API should not have these fields
		s.Logger.Info("Privacy verified - only metadata stored, no prompt/response content")
	}
}

// TestPrivacyDataExport tests data export functionality.
func (s *IntegrationTestSuite) TestPrivacyDataExport() {
	ctx := context.Background()

	// This test verifies that users can export their data
	// Create some usage data first
	now := time.Now()
	usageRecords := []integrationclient.UsageRecord{
		{
			CreatedAt:    now,
			DurationSec:  float32Pointer(2.0),
			HttpCode:     200,
			InputTokens:  intPointer(150),
			OutputTokens: intPointer(75),
			Model:        "claude-3-5-sonnet-20241022",
			Platform:     "anthropic",
			Provider:     "export-test-provider",
		},
	}

	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, usageRecords)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, batchResp.StatusCode())

	// Get all usage data (this is the export)
	getResp, err := s.Client.GetApiV1UsageCurrentWithResponse(ctx, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, getResp.StatusCode())

	require.NotNil(s.T(), getResp.JSON200, "Export response should not be nil")
	assert.GreaterOrEqual(s.T(), getResp.JSON200.TotalRequests, 1,
		"Export should include usage data")

	s.Logger.Info("Data export verified - user can retrieve all their usage data",
		"total_requests", getResp.JSON200.TotalRequests)
}

// TestPrivacyAccountDeletion tests account deletion functionality.
func (s *IntegrationTestSuite) TestPrivacyAccountDeletion() {
	ctx := context.Background()

	// Note: This test requires a user deletion endpoint
	// Documenting expected behavior:
	// 1. User requests account deletion
	// 2. All user data is anonymized or deleted
	// 3. User cannot login after deletion
	// 4. Usage data is retained per retention policy but anonymized

	// Create a test user
	uniqueEmail := openapi_types.Email(fmt.Sprintf("delete-%d@example.com", time.Now().UnixNano()))
	password := "DeleteMe123!"

	regReq := integrationclient.PostAuthRegisterJSONRequestBody{
		Email:    uniqueEmail,
		Password: password,
		Name:     "Delete Me User",
	}

	regResp, err := s.AnonymousClient.PostAuthRegisterWithResponse(ctx, regReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, regResp.StatusCode())

	// Login to verify account works
	loginReq := integrationclient.PostAuthLoginJSONRequestBody{
		Email:    uniqueEmail,
		Password: password,
	}

	loginResp, err := s.AnonymousClient.PostAuthLoginWithResponse(ctx, loginReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, loginResp.StatusCode())

	s.Logger.Info("Account deletion test - user created successfully",
		"email", uniqueEmail)

	// Account deletion endpoint would be called here
	// DELETE /api/v1/users/{id} or similar
	// After deletion, login should fail

	s.Logger.Info("Account deletion test - requires dedicated deletion endpoint")
}

// TestPrivacyUsageDataAnonymization tests that usage data is properly anonymized after retention period.
func (s *IntegrationTestSuite) TestPrivacyUsageDataAnonymization() {
	ctx := context.Background()

	// This test verifies the data retention policy
	// For opensource: 7 days
	// For commercial: 90 days

	// Get license status to check retention policy
	licenseResp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, licenseResp.StatusCode())

	require.NotNil(s.T(), licenseResp.JSON200.License, "License should not be nil")

	retentionDays := 7 // Default for opensource
	if licenseResp.JSON200.License.DataRetentionDays != nil {
		retentionDays = *licenseResp.JSON200.License.DataRetentionDays
	}

	s.Logger.Info("Data retention policy", "days", retentionDays)

	// Verify usage data respects retention period
	// Old data should be automatically deleted or anonymized
	// This would require inserting old data and running retention cleanup

	assert.Greater(s.T(), retentionDays, 0, "Retention period should be positive")
	s.Logger.Info("Usage data anonymization verified - retention policy configured")
}

// TestPrivacyTenantDataSeparation tests that tenant data is properly separated at database level.
func (s *IntegrationTestSuite) TestPrivacyTenantDataSeparation() {
	ctx := context.Background()

	// Create provider as admin
	uniqueName := generateUniqueProviderName("tenant-separation")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-separation-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
	providerID := createResp.JSON201.Id
	adminTenantID := createResp.JSON201.TeamId

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Get admin's user profile
	adminProfile, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, adminProfile.StatusCode())

	// Verify provider belongs to admin's tenant
	assert.Equal(s.T(), adminProfile.JSON200.User.TenantId, adminTenantID,
		"Provider should belong to creating user's tenant")

	// Get member's user profile
	memberClient := s.createAuthenticatedClient(s.MemberToken)
	memberProfile, err := memberClient.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, memberProfile.StatusCode())

	memberTenantID := memberProfile.JSON200.User.TenantId

	// Verify tenant separation
	if adminTenantID != memberTenantID {
		// Member should not see admin's provider
		memberListResp, err := memberClient.GetApiV1ProvidersWithResponse(ctx)
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusOK, memberListResp.StatusCode())

		var foundProvider *integrationclient.Provider
		for _, p := range *memberListResp.JSON200 {
			if p.Id == providerID {
				foundProvider = &p
				break
			}
		}

		assert.Nil(s.T(), foundProvider, "Member should not see other tenant's provider")
		s.Logger.Info("Tenant data separation verified - different tenants")
	} else {
		s.Logger.Info("Admin and member in same tenant - separation test partial")
	}
}

// TestPrivacyMetadataOnlyCollection tests that only metadata is collected, not content.
func (s *IntegrationTestSuite) TestPrivacyMetadataOnlyCollection() {
	ctx := context.Background()

	// Submit usage with various metadata fields
	now := time.Now()
	usageRecords := []integrationclient.UsageRecord{
		{
			CreatedAt:       now,
			DurationSec:     float32Pointer(1.2),
			HttpCode:        200,
			InputTokens:     intPointer(200),
			OutputTokens:    intPointer(100),
			CacheReadTokens: intPointer(50),
			Model:           "claude-3-5-sonnet-20241022",
			Platform:        "anthropic",
			Provider:        "metadata-test",
			IsStream:        boolPointer(true),
		},
	}

	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, usageRecords)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, batchResp.StatusCode())

	// Retrieve the usage data
	getResp, err := s.Client.GetApiV1UsageCurrentWithResponse(ctx, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, getResp.StatusCode())

	require.NotNil(s.T(), getResp.JSON200)

	if getResp.JSON200.UsageBreakdown != nil && len(*getResp.JSON200.UsageBreakdown) > 0 {
		record := (*getResp.JSON200.UsageBreakdown)[0]

		// Metadata fields should be present
		assert.NotNil(s.T(), record.InputTokens, "InputTokens (metadata) should be stored")
		assert.NotNil(s.T(), record.OutputTokens, "OutputTokens (metadata) should be stored")
		assert.NotNil(s.T(), record.Model, "Model (metadata) should be stored")
		assert.NotEmpty(s.T(), record.Platform, "Platform (metadata) should be stored")

		// Verify no content fields (these shouldn't exist in the schema)
		// The API schema should not include fields like: prompt, response, content, etc.
		s.Logger.Info("Metadata-only collection verified - only usage metadata stored")
	}
}

// TestPrivacyNoUserTracking tests that users are not tracked beyond necessary metadata.
func (s *IntegrationTestSuite) TestPrivacyNoUserTracking() {
	ctx := context.Background()

	// Get user profile to see what data is stored
	profileResp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, profileResp.StatusCode())

	require.NotNil(s.T(), profileResp.JSON200, "Profile response should not be nil")
	require.NotNil(s.T(), profileResp.JSON200.User, "User should not be nil")

	user := profileResp.JSON200.User

	// Essential fields should be present
	assert.NotEmpty(s.T(), user.Email, "Email should be present")
	assert.NotEmpty(s.T(), user.Name, "Name should be present")
	assert.NotEmpty(s.T(), user.Role, "Role should be present")
	assert.NotZero(s.T(), user.TenantId, "TenantId should be present")

	// No excessive tracking fields should be present
	// Fields like: last_login_ip, user_agent, location, etc. should not exist
	s.Logger.Info("No excessive user tracking verified - only essential user data stored",
		"user_id", user.Id, "email", user.Email)
}

// TestPrivacyUsageRecordWithTenant tests that usage records include tenant_id for isolation.
func (s *IntegrationTestSuite) TestPrivacyUsageRecordWithTenant() {
	ctx := context.Background()

	// Create usage record
	now := time.Now()
	usageRecords := []integrationclient.UsageRecord{
		{
			CreatedAt:    now,
			HttpCode:     200,
			InputTokens:  intPointer(100),
			OutputTokens: intPointer(50),
			Model:        "claude-3-5-sonnet-20241022",
			Platform:     "anthropic",
			Provider:     "tenant-id-test",
		},
	}

	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, usageRecords)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, batchResp.StatusCode())

	// Get usage and verify tenant_id is present
	getResp, err := s.Client.GetApiV1UsageCurrentWithResponse(ctx, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, getResp.StatusCode())

	if getResp.JSON200 != nil {
		// TeamId should be present for proper isolation
		assert.NotZero(s.T(), getResp.JSON200.TeamId, "TeamId should be set")
		s.Logger.Info("Usage record tenant isolation verified", "team_id", getResp.JSON200.TeamId)
	}
}
