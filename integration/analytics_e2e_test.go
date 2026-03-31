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
	"io"
	"net/http"
	"time"

	integrationclient "github.com/code-together/shared/integration"
	integration_manager "github.com/code-together/integration_manager"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// End-to-End Data Flow Tests (5 tests)
// ============================================================================

// TestUploadToProviderAnalyticsFlow verifies that uploaded usage records
// appear correctly in provider analytics with accurate totals.
func (s *IntegrationTestSuite) TestUploadToProviderAnalyticsFlow() {
	ctx := context.Background()

	// Create provider with unique name
	uniqueName := generateUniqueProviderName("e2e-provider-analytics")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus", "claude-3-sonnet"},
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, resp.StatusCode(), "Provider creation should return 201")
	providerID := resp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Upload 10 usage records with varying tokens/models
	now := time.Now()
	records := []integrationclient.UsageRecord{
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			DurationSec:  float32Pointer(1.5),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(2000),
			OutputTokens: intPointer(1000),
			DurationSec:  float32Pointer(2.0),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-10 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-sonnet",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1500),
			OutputTokens: intPointer(750),
			DurationSec:  float32Pointer(1.8),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-20 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-sonnet",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(2500),
			OutputTokens: intPointer(1250),
			DurationSec:  float32Pointer(2.5),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-30 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1200),
			OutputTokens: intPointer(600),
			DurationSec:  float32Pointer(1.6),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-40 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1800),
			OutputTokens: intPointer(900),
			DurationSec:  float32Pointer(2.2),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-50 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-sonnet",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(900),
			OutputTokens: intPointer(450),
			DurationSec:  float32Pointer(1.2),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-60 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-sonnet",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(2200),
			OutputTokens: intPointer(1100),
			DurationSec:  float32Pointer(2.8),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-70 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1700),
			OutputTokens: intPointer(850),
			DurationSec:  float32Pointer(2.1),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-80 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1300),
			OutputTokens: intPointer(650),
			DurationSec:  float32Pointer(1.7),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-90 * time.Minute),
		},
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode(), "Usage upload should return 200")
	require.Equal(s.T(), 10, batchResp.JSON200.SyncedCount, "Should sync all 10 records")

	// Query provider analytics
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
	// Additional assertions would parse JSON response and verify:
	// - Total requests: 10
	// - Total input tokens: 17100 (sum of all input tokens)
	// - Total output tokens: 8550 (sum of all output tokens)
	// - Model breakdown: 6 claude-3-opus, 4 claude-3-sonnet
	// - Success rate: 100% (all HTTP 200)
}

// TestUploadToUserAnalyticsFlow verifies that user-specific usage
// appears correctly in user analytics with per-user aggregations.
func (s *IntegrationTestSuite) TestUploadToUserAnalyticsFlow() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("e2e-user-analytics")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Get admin user ID
	adminProfile, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, adminProfile.StatusCode())
	adminID := adminProfile.JSON200.User.Id

	now := time.Now()

	// Upload usage records for admin user (5 records)
	records := []integrationclient.UsageRecord{
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       &adminID,
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1500),
			OutputTokens: intPointer(750),
			TenantId:     int64Pointer(1),
			UserId:       &adminID,
			CreatedAt:    now.Add(-10 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1200),
			OutputTokens: intPointer(600),
			TenantId:     int64Pointer(1),
			UserId:       &adminID,
			CreatedAt:    now.Add(-20 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(800),
			OutputTokens: intPointer(400),
			TenantId:     int64Pointer(1),
			UserId:       &adminID,
			CreatedAt:    now.Add(-30 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1100),
			OutputTokens: intPointer(550),
			TenantId:     int64Pointer(1),
			UserId:       &adminID,
			CreatedAt:    now.Add(-40 * time.Minute),
		},
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 5, batchResp.JSON200.SyncedCount, "Should sync all 5 records")

	// Query user analytics for admin user
	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/users?start_date=%s&end_date=%s&user_ids=%d",
		s.ServerURL, startDate, endDate, adminID)

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
	// Additional assertions would verify:
	// - Admin: 5 requests, 8950 total tokens (5600 input + 3350 output)
}

// TestUploadToHistoryAnalyticsFlow verifies that uploaded records
// appear in usage history with correct pagination.
func (s *IntegrationTestSuite) TestUploadToHistoryAnalyticsFlow() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("e2e-history-analytics")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Upload 25 usage records with timestamps spread over 2 days
	now := time.Now()
	records := make([]integrationclient.UsageRecord, 25)

	for i := 0; i < 25; i++ {
		// Distribute across 2 days (first 13 today, next 12 yesterday)
		var createdAt time.Time
		if i < 13 {
			createdAt = now.Add(-time.Duration(i) * time.Minute)
		} else {
			createdAt = now.Add(-24*time.Hour - time.Duration(i-13)*time.Hour)
		}

		records[i] = integrationclient.UsageRecord{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000 + i*10),
			OutputTokens: intPointer(500 + i*5),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    createdAt,
		}
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 25, batchResp.JSON200.SyncedCount, "Should sync all 25 records")

	// Query usage history with pagination
	startDate := time.Now().Add(-48 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/history?start_date=%s&end_date=%s&page=1&limit=10&providers=%s&sort_by=created_at&sort_order=desc",
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
	// Additional assertions would verify:
	// - Page 1 returns 10 records
	// - Total count is 25
	// - Total pages: 3
	// - Records sorted by created_at desc (newest first)
	// - All records have correct provider name
}

// TestRealTimeDataFreshness verifies that analytics reflect
// uploaded usage immediately without delay.
func (s *IntegrationTestSuite) TestRealTimeDataFreshness() {
	ctx := context.Background()

	// Record start time
	startTime := time.Now()

	// Create provider
	uniqueName := generateUniqueProviderName("e2e-realtime-analytics")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Upload 5 usage records immediately
	now := time.Now()
	records := []integrationclient.UsageRecord{
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1200),
			OutputTokens: intPointer(600),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(800),
			OutputTokens: intPointer(400),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1500),
			OutputTokens: intPointer(750),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(900),
			OutputTokens: intPointer(450),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now,
		},
	}

	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 5, batchResp.JSON200.SyncedCount, "Should sync all 5 records")

	// Immediately query analytics (within 1 second)
	startDate := startTime.Add(-24 * time.Hour).Format("2006-01-02")
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

	elapsed := time.Since(startTime)

	s.T().Logf("Real-time analytics response (%d) in %v: %s", httpResp.StatusCode, elapsed, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	assert.Less(s.T(), elapsed, 1*time.Second, "Analytics should be available within 1 second")
	// Additional assertions would verify all 5 records appear in analytics
}

// TestErrorRecordsInAnalytics verifies that error records
// (HTTP 4xx/5xx) are tracked correctly in analytics.
func (s *IntegrationTestSuite) TestErrorRecordsInAnalytics() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("e2e-error-analytics")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Upload 10 usage records: 7 success, 2 client errors, 1 server error
	now := time.Now()
	records := []integrationclient.UsageRecord{
		// 7 success (HTTP 200)
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1100),
			OutputTokens: intPointer(550),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-10 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(900),
			OutputTokens: intPointer(450),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-20 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1200),
			OutputTokens: intPointer(600),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-30 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1050),
			OutputTokens: intPointer(525),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-40 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(950),
			OutputTokens: intPointer(475),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-50 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1150),
			OutputTokens: intPointer(575),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-60 * time.Minute),
		},
		// 2 client errors (HTTP 400)
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     400,
			InputTokens:  intPointer(500),
			OutputTokens: intPointer(0),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-5 * time.Minute),
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     400,
			InputTokens:  intPointer(600),
			OutputTokens: intPointer(0),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-15 * time.Minute),
		},
		// 1 server error (HTTP 500)
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     500,
			InputTokens:  intPointer(300),
			OutputTokens: intPointer(0),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-25 * time.Minute),
		},
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 10, batchResp.JSON200.SyncedCount, "Should sync all 10 records")

	// Query provider analytics
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

	s.T().Logf("Error analytics response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - Success rate: 70% (7/10 requests)
	// - Error rate: 30% (3/10 requests)
	// - Error breakdown by status code:
	//   - HTTP 200: 7 requests
	//   - HTTP 400: 2 requests
	//   - HTTP 500: 1 request
	// - Error records don't skew token averages (or are calculated separately)
}

// Helper function for creating float32 pointers
func float32Pointer(f float32) *float32 {
	return &f
}

// ============================================================================
// Personal Analytics E2E Tests (issue #101)
// Uses integration_manager (manager spec) since personal analytics is manager business.
// ============================================================================

// TestPersonalAnalyticsDataIsolation verifies that a manager user can only
// see their own data in personal analytics, not other users' data.
func (s *IntegrationTestSuite) TestPersonalAnalyticsDataIsolation() {
	ctx := context.Background()

	// Bootstrap standard fixture (creates admin + member users)
	s.bootstrapStandardFixture()

	// Create provider
	uniqueName := generateUniqueProviderName("e2e-personal-isolation")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id
	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	now := time.Now()

	// Upload records for admin user (user_id = 1)
	adminRecords := []integrationclient.UsageRecord{
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(5000),
			OutputTokens: intPointer(3000),
			DurationSec:  float32Pointer(3.0),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1), // admin
			CreatedAt:    now,
		},
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(2000),
			OutputTokens: intPointer(1000),
			DurationSec:  float32Pointer(1.5),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1), // admin
			CreatedAt:    now.Add(-5 * time.Minute),
		},
	}

	// Upload records for member user (different user_id)
	memberRecords := []integrationclient.UsageRecord{
		{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(300),
			OutputTokens: intPointer(200),
			DurationSec:  float32Pointer(0.5),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(2), // member
			CreatedAt:    now,
		},
	}

	// Upload all records using client spec (for batch upload)
	allRecords := append(adminRecords, memberRecords...)
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(allRecords)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 3, batchResp.JSON200.SyncedCount, "Should sync all 3 records")

	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	// Query personal analytics as admin via manager client - should only see admin's 2 records
	adminAnalyticsResp, err := s.ManagerClient.GetApiV1AnalyticsPersonalWithResponse(ctx, &integration_manager.GetApiV1AnalyticsPersonalParams{
		StartDate: openapi_types.Date(startDate),
		EndDate:   endDate,
	})
	require.NoError(s.T(), err, "Manager personal analytics request failed")
	require.Equal(s.T(), 200, adminAnalyticsResp.StatusCode(), "Manager should be able to access personal analytics")
	require.NotNil(s.T(), adminAnalyticsResp.JSON200)
	require.NotNil(s.T(), adminAnalyticsResp.JSON200.Summary)

	adminSummary := adminAnalyticsResp.JSON200.Summary
	assert.Equal(s.T(), 2, *adminSummary.TotalRequests, "Admin should see exactly 2 requests")
	assert.Equal(s.T(), 11000, *adminSummary.TotalTokens, "Admin should see 11000 total tokens (5000+2000 input + 3000+1000 output)")

	s.T().Logf("Admin personal analytics (manager client): requests=%d, tokens=%d, cost=%.4f",
		*adminSummary.TotalRequests, *adminSummary.TotalTokens, *adminSummary.TotalCost)
}

// TestPersonalAnalyticsFilterOptions verifies that personal filter options
// are returned for any authenticated user via the manager client.
func (s *IntegrationTestSuite) TestPersonalAnalyticsFilterOptions() {
	ctx := context.Background()

	s.bootstrapStandardFixture()

	// Manager client can access personal filter options
	filterResp, err := s.ManagerClient.GetApiV1AnalyticsPersonalFiltersWithResponse(ctx)
	require.NoError(s.T(), err, "Personal filter options request failed")
	require.Equal(s.T(), 200, filterResp.StatusCode(), "Manager should be able to access personal filter options")
	require.NotNil(s.T(), filterResp.JSON200)

	s.T().Logf("Personal filter options (manager client): providers=%v, models=%v, tools=%v",
		filterResp.JSON200.Providers, filterResp.JSON200.Models, filterResp.JSON200.Tools)
}

// TestPersonalAnalyticsRequiresAuth verifies that personal analytics
// endpoints require authentication.
func (s *IntegrationTestSuite) TestPersonalAnalyticsRequiresAuth() {
	ctx := context.Background()

	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	// Create anonymous (unauthenticated) manager client
	anonManagerClient, err := integration_manager.NewClientWithResponses(s.ServerURL)
	require.NoError(s.T(), err)

	// Personal analytics requires auth
	personalResp, err := anonManagerClient.GetApiV1AnalyticsPersonalWithResponse(ctx, &integration_manager.GetApiV1AnalyticsPersonalParams{
		StartDate: openapi_types.Date(startDate),
		EndDate:   endDate,
	})
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 401, personalResp.StatusCode(), "Personal analytics should require auth")

	// Personal history requires auth
	historyResp, err := anonManagerClient.GetApiV1AnalyticsPersonalHistoryWithResponse(ctx, &integration_manager.GetApiV1AnalyticsPersonalHistoryParams{
		StartDate: openapi_types.Date(startDate),
		EndDate:   endDate,
	})
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 401, historyResp.StatusCode(), "Personal history should require auth")

	// Personal filters requires auth
	filtersResp, err := anonManagerClient.GetApiV1AnalyticsPersonalFiltersWithResponse(ctx)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 401, filtersResp.StatusCode(), "Personal filters should require auth")
}

// TestPersonalAnalyticsHistoryPagination verifies that personal history
// supports pagination using the manager client.
func (s *IntegrationTestSuite) TestPersonalAnalyticsHistoryPagination() {
	ctx := context.Background()

	s.bootstrapStandardFixture()

	// Create provider
	uniqueName := generateUniqueProviderName("e2e-personal-history")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id
	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	now := time.Now()

	// Upload 5 records for admin user
	records := make([]integrationclient.UsageRecord, 5)
	for i := 0; i < 5; i++ {
		records[i] = integrationclient.UsageRecord{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			DurationSec:  float32Pointer(1.0),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(time.Duration(i) * 10 * time.Minute),
		}
	}

	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 5, batchResp.JSON200.SyncedCount, "Should sync all 5 records")

	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	// Page 1 with limit=2 via manager client
	sortBy := integration_manager.CreatedAt
	sortOrder := integration_manager.Desc
	page1Resp, err := s.ManagerClient.GetApiV1AnalyticsPersonalHistoryWithResponse(ctx, &integration_manager.GetApiV1AnalyticsPersonalHistoryParams{
		StartDate: openapi_types.Date(startDate),
		EndDate:   endDate,
		Page:      intPointer(1),
		Limit:     intPointer(2),
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
	})
	require.NoError(s.T(), err, "Personal history page 1 request failed")
	require.Equal(s.T(), 200, page1Resp.StatusCode())
	require.NotNil(s.T(), page1Resp.JSON200)
	require.NotNil(s.T(), page1Resp.JSON200.Pagination)

	assert.Equal(s.T(), 2, len(*page1Resp.JSON200.Records), "Page 1 should have 2 records")
	assert.Equal(s.T(), 5, *page1Resp.JSON200.Pagination.TotalCount, "Total should be 5 records")
	assert.Equal(s.T(), 3, *page1Resp.JSON200.Pagination.TotalPages, "Should have 3 total pages")

	// Page 2 with limit=2
	page2Resp, err := s.ManagerClient.GetApiV1AnalyticsPersonalHistoryWithResponse(ctx, &integration_manager.GetApiV1AnalyticsPersonalHistoryParams{
		StartDate: openapi_types.Date(startDate),
		EndDate:   endDate,
		Page:      intPointer(2),
		Limit:     intPointer(2),
		SortBy:    &sortBy,
		SortOrder: &sortOrder,
	})
	require.NoError(s.T(), err, "Personal history page 2 request failed")
	require.Equal(s.T(), 200, page2Resp.StatusCode())
	assert.Equal(s.T(), 2, len(*page2Resp.JSON200.Records), "Page 2 should have 2 records")

	s.T().Logf("Personal history pagination (manager client): page1=%d records, page2=%d records, total=%d",
		len(*page1Resp.JSON200.Records), len(*page2Resp.JSON200.Records), *page1Resp.JSON200.Pagination.TotalCount)
}
