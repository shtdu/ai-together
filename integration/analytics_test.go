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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	integrationclient "github.com/code-together/shared/integration"
)

// ============================================================================
// Provider Analytics Tests (2 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestAnalyticsProviderGetSuccess() {
	ctx := context.Background()

	// Create a provider with unique name
	uniqueName := generateUniqueProviderName("analytics-provider")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-12345",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
	}

	// Create provider
	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, resp.StatusCode(), "Provider creation should return 201")
	providerID := resp.JSON201.Id

	// Upload some usage records for this provider
	records := []integrationclient.UsageRecord{
		{
			Platform:      "claude",
			Model:         "claude-3-opus",
			Provider:      uniqueName,
			HttpCode:      200,
			InputTokens:   intPointer(1000),
			OutputTokens:  intPointer(500),
			TenantId:      int64Pointer(1),
			UserId:        int64Pointer(1),
			CreatedAt:     time.Now(),
		},
		{
			Platform:      "claude",
			Model:         "claude-3-sonnet",
			Provider:      uniqueName,
			HttpCode:      200,
			InputTokens:   intPointer(500),
			OutputTokens:  intPointer(300),
			TenantId:      int64Pointer(1),
			UserId:        int64Pointer(1),
			CreatedAt:     time.Now().Add(-1 * time.Hour),
		},
	}

	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode(), "Usage upload should return 200")
	require.Equal(s.T(), 2, batchResp.JSON200.SyncedCount, "Should sync 2 records")

	// Get provider analytics via direct HTTP call
	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/providers?start_date=%s&end_date=%s&providers=%s",
		s.ServerURL, startDate, endDate, uniqueName)

	httpReq, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)
	httpReq.Header.Set("Authorization", "Bearer "+s.AdminToken)

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(s.T(), err)
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	require.NoError(s.T(), err)

	s.T().Logf("Provider analytics response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")

	// Cleanup
	s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
}

func (s *IntegrationTestSuite) TestAnalyticsProviderUnauthorized() {
	// Create and login as member user
	memberToken := s.registerAndLoginUserFromFixture("member")

	// Try to access provider analytics as member via direct HTTP
	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/providers?start_date=%s&end_date=%s",
		s.ServerURL, startDate, endDate)

	httpReq, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)
	httpReq.Header.Set("Authorization", "Bearer "+memberToken)

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(s.T(), err)
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	require.NoError(s.T(), err)

	s.T().Logf("Unauthorized analytics response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 403, httpResp.StatusCode, "Member should be forbidden from accessing provider analytics")

	// Parse and validate error response
	var errResp integrationclient.ErrorResponse
	err = json.Unmarshal(body, &errResp)
	require.NoError(s.T(), err, "Error response should be valid JSON")
	requireErrorResponse(s, &errResp, "permission")
}

// ============================================================================
// User Analytics Tests (1 test)
// ============================================================================

func (s *IntegrationTestSuite) TestAnalyticsUserGetSuccess() {
	ctx := context.Background()

	// Create a member user
	memberToken := s.registerAndLoginUserFromFixture("member")
	memberClient := s.createAuthenticatedClient(memberToken)

	// Get the user ID from member's profile
	profileResp, err := memberClient.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get member profile")
	require.Equal(s.T(), 200, profileResp.StatusCode(), "Profile fetch should return 200")
	memberUserID := profileResp.JSON200.User.Id

	// Create provider and upload usage for the member
	uniqueName := generateUniqueProviderName("analytics-user-provider")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-12345",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id

	// Upload usage records for the member
	records := []integrationclient.UsageRecord{
		{
			Platform:      "claude",
			Model:         "claude-3-opus",
			Provider:      uniqueName,
			HttpCode:      200,
			InputTokens:   intPointer(2000),
			OutputTokens:  intPointer(1000),
			TenantId:      int64Pointer(1),
			UserId:        &memberUserID,
			CreatedAt:     time.Now().Add(-2 * time.Hour),
		},
	}

	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())

	// Get user analytics as admin via direct HTTP
	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/users?start_date=%s&end_date=%s&user_ids=%d",
		s.ServerURL, startDate, endDate, memberUserID)

	httpReq, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)
	httpReq.Header.Set("Authorization", "Bearer "+s.AdminToken)

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(s.T(), err)
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	require.NoError(s.T(), err)

	s.T().Logf("User analytics response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")

	// Cleanup
	s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
}

// ============================================================================
// Historical Analytics Tests (1 test)
// ============================================================================

func (s *IntegrationTestSuite) TestAnalyticsHistoryGetSuccess() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("analytics-history-provider")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-12345",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus", "claude-3-sonnet"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id

	// Upload historical usage records
	records := []integrationclient.UsageRecord{
		{
			Platform:      "claude",
			Model:         "claude-3-opus",
			Provider:      uniqueName,
			HttpCode:      200,
			InputTokens:   intPointer(500),
			OutputTokens:  intPointer(200),
			TenantId:      int64Pointer(1),
			UserId:        int64Pointer(1),
			CreatedAt:     time.Now().Add(-24 * time.Hour),
		},
		{
			Platform:      "claude",
			Model:         "claude-3-sonnet",
			Provider:      uniqueName,
			HttpCode:      200,
			InputTokens:   intPointer(300),
			OutputTokens:  intPointer(150),
			TenantId:      int64Pointer(1),
			UserId:        int64Pointer(1),
			CreatedAt:     time.Now().Add(-12 * time.Hour),
		},
	}

	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())

	// Get usage history with pagination via direct HTTP
	startDate := time.Now().Add(-48 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/history?start_date=%s&end_date=%s&page=1&limit=50&providers=%s&sort_by=created_at&sort_order=desc",
		s.ServerURL, startDate, endDate, uniqueName)

	httpReq, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)
	httpReq.Header.Set("Authorization", "Bearer "+s.AdminToken)

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(s.T(), err)
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	require.NoError(s.T(), err)

	s.T().Logf("Usage history response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")

	// Cleanup
	s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
}

// Helper function for creating int pointers
func intPointer(i int) *int {
	return &i
}

// Helper function for creating int64 pointers
func int64Pointer(i int64) *int64 {
	return &i
}
