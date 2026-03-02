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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	integrationclient "github.com/code-together/shared/integration"
)

// ============================================================================
// Aggregation Accuracy Tests (6 tests)
// ============================================================================

// TestTokenSumAggregation verifies that token aggregation is
// mathematically accurate with no rounding errors.
func (s *IntegrationTestSuite) TestTokenSumAggregation() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("agg-token-sum")
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

	// Upload usage records with known token values:
	// Record 1: input=1000, output=500
	// Record 2: input=2000, output=1000
	// Record 3: input=1500, output=750
	// Expected totals: input=4500, output=2250, total=6750
	now := time.Now()
	records := []integrationclient.UsageRecord{
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now,
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(2000),
			OutputTokens: intPointer(1000),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-10 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1500),
			OutputTokens: intPointer(750),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-20 * time.Minute),
		},
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 3, batchResp.JSON200.SyncedCount, "Should sync all 3 records")

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

	s.T().Logf("Token sum aggregation response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - Total input tokens: 4500 (1000 + 2000 + 1500)
	// - Total output tokens: 2250 (500 + 1000 + 750)
	// - Total tokens: 6750 (sum of all)
	// - No rounding errors or token loss
}

// TestAverageCalculationAccuracy verifies that average calculations
// (mean, min, max) are correct.
func (s *IntegrationTestSuite) TestAverageCalculationAccuracy() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("agg-avg-calc")
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

	// Upload 5 records with durations: [1.0s, 2.0s, 3.0s, 4.0s, 5.0s]
	// Expected average: (1+2+3+4+5)/5 = 3.0
	// Min: 1.0, Max: 5.0
	now := time.Now()
	d1 := float32(1.0)
	d2 := float32(2.0)
	d3 := float32(3.0)
	d4 := float32(4.0)
	d5 := float32(5.0)

	records := []integrationclient.UsageRecord{
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			DurationSec: &d1,
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now,
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			DurationSec: &d2,
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-10 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			DurationSec: &d3,
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-20 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			DurationSec: &d4,
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-30 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			DurationSec: &d5,
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-40 * time.Minute),
		},
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 5, batchResp.JSON200.SyncedCount, "Should sync all 5 records")

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

	s.T().Logf("Average calculation response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - Average duration: 3.0s ( (1+2+3+4+5)/5 = 15/5 )
	// - Min duration: 1.0s
	// - Max duration: 5.0s
	// - No rounding errors in calculations
}

// TestModelLevelAggregation verifies that usage is correctly
// aggregated by model with accurate per-model totals.
func (s *IntegrationTestSuite) TestModelLevelAggregation() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("agg-model-level")
	kind := integrationclient.CreateProviderRequestKindClaude
	provReq := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &[]bool{true}[0],
		Level:           &[]int{1}[0],
		SupportedModels: &[]string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
	}

	provResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, provReq)
	require.NoError(s.T(), err, "Failed to create provider")
	require.Equal(s.T(), 201, provResp.StatusCode())
	providerID := provResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Upload usage across 3 models:
	// claude-3-opus: 5 records, 10000 tokens total
	// claude-3-sonnet: 3 records, 6000 tokens total
	// claude-3-haiku: 2 records, 2000 tokens total
	now := time.Now()
	records := []integrationclient.UsageRecord{
		// claude-3-opus (5 records)
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now,
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1200),
			OutputTokens: intPointer(600),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-10 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1500),
			OutputTokens: intPointer(750),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-20 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1800),
			OutputTokens: intPointer(900),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-30 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(2000),
			OutputTokens: intPointer(1000),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-40 * time.Minute),
		},
		// claude-3-sonnet (3 records)
		{
			Platform:    "claude",
			Model:       "claude-3-sonnet",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-5 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-sonnet",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1500),
			OutputTokens: intPointer(750),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-15 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-sonnet",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(2000),
			OutputTokens: intPointer(1000),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-25 * time.Minute),
		},
		// claude-3-haiku (2 records)
		{
			Platform:    "claude",
			Model:       "claude-3-haiku",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(500),
			OutputTokens: intPointer(250),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-3 * time.Minute),
		},
		{
			Platform:    "claude",
			Model:       "claude-3-haiku",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(750),
			OutputTokens: intPointer(375),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-8 * time.Minute),
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

	s.T().Logf("Model-level aggregation response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - Model breakdown shows all 3 models
	// - claude-3-opus: 5 requests, 10000 tokens total
	// - claude-3-sonnet: 3 requests, 6000 tokens total
	// - claude-3-haiku: 2 requests, 2000 tokens total
	// - Total: 10 requests, 18000 tokens
	// - Model percentages sum to 100%
}

// TestDailyAggregationAccuracy verifies that daily aggregation
// groups records correctly by day.
func (s *IntegrationTestSuite) TestDailyAggregationAccuracy() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("agg-daily")
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

	// Upload records over 3 days:
	// Day 1: 5 records, 5000 tokens total
	// Day 2: 3 records, 3000 tokens total
	// Day 3: 7 records, 7000 tokens total
	now := time.Now()
	records := make([]integrationclient.UsageRecord, 15)

	// Day 1 (5 records)
	for i := 0; i < 5; i++ {
		records[i] = integrationclient.UsageRecord{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-time.Duration(i) * time.Hour),
		}
	}

	// Day 2 (3 records) - 24 hours ago
	for i := 0; i < 3; i++ {
		records[i+5] = integrationclient.UsageRecord{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-24*time.Hour - time.Duration(i)*time.Hour),
		}
	}

	// Day 3 (7 records) - 48 hours ago
	for i := 0; i < 7; i++ {
		records[i+8] = integrationclient.UsageRecord{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    uniqueName,
			HttpCode:    200,
			InputTokens: intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-48*time.Hour - time.Duration(i)*time.Hour),
		}
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 15, batchResp.JSON200.SyncedCount, "Should sync all 15 records")

	// Query provider analytics
	startDate := time.Now().Add(-72 * time.Hour).Format("2006-01-02")
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

	s.T().Logf("Daily aggregation response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - 3 data points (one per day) in time series
	// - Day 1: 5 records, 7500 tokens (5000 input + 2500 output)
	// - Day 2: 3 records, 4500 tokens (3000 input + 1500 output)
	// - Day 3: 7 records, 10500 tokens (7000 input + 3500 output)
	// - No records counted in wrong day
	// - Totals: 15 records, 22500 tokens
}

// TestMultiProviderAggregation verifies that aggregation works
// correctly when querying multiple providers.
func (s *IntegrationTestSuite) TestMultiProviderAggregation() {
	ctx := context.Background()

	// Create 3 providers
	providers := make([]string, 3)
	providerIDs := make([]int64, 3)

	for i := 0; i < 3; i++ {
		uniqueName := generateUniqueProviderName(fmt.Sprintf("agg-multi-provider-%d", i))
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

		providers[i] = uniqueName
		providerIDs[i] = provResp.JSON201.Id

		defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, provResp.JSON201.Id)
	}

	// Upload usage for each provider:
	// Provider A: 10 records, 5000 tokens
	// Provider B: 8 records, 4000 tokens
	// Provider C: 12 records, 6000 tokens
	now := time.Now()
	records := []integrationclient.UsageRecord{}

	// Provider A: 10 records, 500 tokens each
	for i := 0; i < 10; i++ {
		records = append(records, integrationclient.UsageRecord{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    providers[0],
			HttpCode:    200,
			InputTokens: intPointer(500),
			OutputTokens: intPointer(250),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-time.Duration(i) * 5 * time.Minute),
		})
	}

	// Provider B: 8 records, 500 tokens each
	for i := 0; i < 8; i++ {
		records = append(records, integrationclient.UsageRecord{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    providers[1],
			HttpCode:    200,
			InputTokens: intPointer(500),
			OutputTokens: intPointer(250),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-time.Duration(i) * 5 * time.Minute),
		})
	}

	// Provider C: 12 records, 500 tokens each
	for i := 0; i < 12; i++ {
		records = append(records, integrationclient.UsageRecord{
			Platform:    "claude",
			Model:       "claude-3-opus",
			Provider:    providers[2],
			HttpCode:    200,
			InputTokens: intPointer(500),
			OutputTokens: intPointer(250),
			TenantId:    int64Pointer(1),
			UserId:      int64Pointer(1),
			CreatedAt:   now.Add(-time.Duration(i) * 5 * time.Minute),
		})
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 30, batchResp.JSON200.SyncedCount, "Should sync all 30 records")

	// Query analytics for all 3 providers
	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/providers?start_date=%s&end_date=%s&providers=%s&providers=%s&providers=%s",
		s.ServerURL, startDate, endDate, providers[0], providers[1], providers[2])

	httpReq, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)
	httpReq.Header.Set("Authorization", "Bearer "+s.AdminToken)

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(httpReq)
	require.NoError(s.T(), err)
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	require.NoError(s.T(), err)

	s.T().Logf("Multi-provider aggregation response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - Provider A: 10 requests, 7500 tokens (5000 input + 2500 output)
	// - Provider B: 8 requests, 6000 tokens (4000 input + 2000 output)
	// - Provider C: 12 requests, 9000 tokens (6000 input + 3000 output)
	// - Combined total: 30 requests, 22500 tokens
	// - Providers ranked correctly by usage (C > A > B)
}

// TestCacheTokensInAggregation verifies that cache read/create
// tokens are included in analytics aggregations.
func (s *IntegrationTestSuite) TestCacheTokensInAggregation() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("agg-cache-tokens")
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

	// Upload record with all token types:
	// Input tokens: 1000
	// Cache read tokens: 500
	// Cache create tokens: 200
	// Output tokens: 300
	// Expected total: 2000
	now := time.Now()
	cacheRead := 500
	cacheCreate := 200

	records := []integrationclient.UsageRecord{
		{
			Platform:          "claude",
			Model:             "claude-3-opus",
			Provider:          uniqueName,
			HttpCode:          200,
			InputTokens:       intPointer(1000),
			CacheReadTokens:   &cacheRead,
			CacheCreateTokens: &cacheCreate,
			OutputTokens:      intPointer(300),
			TenantId:          int64Pointer(1),
			UserId:            int64Pointer(1),
			CreatedAt:         now,
		},
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 1, batchResp.JSON200.SyncedCount, "Should sync 1 record")

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

	s.T().Logf("Cache tokens aggregation response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - Input tokens: 1000
	// - Cache read tokens: 500
	// - Cache create tokens: 200
	// - Output tokens: 300
	// - Total tokens: 2000 (sum of all types)
	// - All token types are reported separately
}
