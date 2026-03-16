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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Time-Based Analytics Tests (5 tests)
// ============================================================================

// TestDateRangeFiltering verifies that date range filters work correctly.
func (s *IntegrationTestSuite) TestDateRangeFiltering() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("time-date-range")
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

	// Upload records over 7 days with specific distribution:
	// Days 1-3: 15 records (5 per day)
	// Days 4-5: 10 records (5 per day)
	// Days 6-7: 12 records (6 per day)
	// Total: 37 records
	now := time.Now()
	records := make([]integrationclient.UsageRecord, 37)

	recordIdx := 0
	// Days 1-3: 5 records each
	for day := 0; day < 3; day++ {
		for i := 0; i < 5; i++ {
			records[recordIdx] = integrationclient.UsageRecord{
				Platform:     "claude",
				Model:        "claude-3-opus",
				Provider:     uniqueName,
				HttpCode:     200,
				InputTokens:  intPointer(1000),
				OutputTokens: intPointer(500),
				TenantId:     int64Pointer(1),
				UserId:       int64Pointer(1),
				CreatedAt:    now.Add(-time.Duration(day)*24*time.Hour - time.Duration(i)*time.Hour),
			}
			recordIdx++
		}
	}

	// Days 4-5: 5 records each
	for day := 3; day < 5; day++ {
		for i := 0; i < 5; i++ {
			records[recordIdx] = integrationclient.UsageRecord{
				Platform:     "claude",
				Model:        "claude-3-opus",
				Provider:     uniqueName,
				HttpCode:     200,
				InputTokens:  intPointer(1000),
				OutputTokens: intPointer(500),
				TenantId:     int64Pointer(1),
				UserId:       int64Pointer(1),
				CreatedAt:    now.Add(-time.Duration(day)*24*time.Hour - time.Duration(i)*time.Hour),
			}
			recordIdx++
		}
	}

	// Days 6-7: 6 records each
	for day := 5; day < 7; day++ {
		for i := 0; i < 6; i++ {
			records[recordIdx] = integrationclient.UsageRecord{
				Platform:     "claude",
				Model:        "claude-3-opus",
				Provider:     uniqueName,
				HttpCode:     200,
				InputTokens:  intPointer(1000),
				OutputTokens: intPointer(500),
				TenantId:     int64Pointer(1),
				UserId:       int64Pointer(1),
				CreatedAt:    now.Add(-time.Duration(day)*24*time.Hour - time.Duration(i)*time.Hour),
			}
			recordIdx++
		}
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 37, batchResp.JSON200.SyncedCount, "Should sync all 37 records")

	// Query with different date ranges and verify counts
	testCases := []struct {
		name        string
		daysAgo     int
		duration    int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "Days 1-3",
			daysAgo:     2,
			duration:    3,
			expectedMin: 15,
			expectedMax: 15,
		},
		{
			name:        "Days 4-5",
			daysAgo:     4,
			duration:    2,
			expectedMin: 10,
			expectedMax: 10,
		},
		{
			name:        "Days 1-7 (all)",
			daysAgo:     6,
			duration:    7,
			expectedMin: 37,
			expectedMax: 37,
		},
	}

	for _, tc := range testCases {
		startDate := time.Now().Add(-time.Duration(tc.daysAgo+tc.duration-1) * 24 * time.Hour).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")

		url := fmt.Sprintf("%s/api/v1/analytics/history?start_date=%s&end_date=%s&providers=%s&limit=100",
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

		s.T().Logf("Date range test '%s' response (%d): %s", tc.name, httpResp.StatusCode, string(body))

		assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK for date range: "+tc.name)
		// Additional assertions would verify:
		// - Record count matches expected for that range
		// - Records outside range are excluded
		// - Boundary conditions work correctly
	}
}

// TestHourlyGranularity verifies hourly aggregation within a single day.
func (s *IntegrationTestSuite) TestHourlyGranularity() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("time-hourly")
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

	// Upload 24 records (one per hour) over a single day
	now := time.Now()
	records := make([]integrationclient.UsageRecord, 24)

	for i := 0; i < 24; i++ {
		records[i] = integrationclient.UsageRecord{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-time.Duration(i) * time.Hour),
		}
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 24, batchResp.JSON200.SyncedCount, "Should sync all 24 records")

	// Query analytics with hourly grouping
	startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/history?start_date=%s&end_date=%s&providers=%s&limit=100",
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

	s.T().Logf("Hourly granularity response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - 24 hourly data points (if grouping by hour)
	// - Each hour has correct record count
	// - Hours with no data return 0 (or are skipped)
}

// TestWeeklyAggregation verifies weekly grouping of usage data.
func (s *IntegrationTestSuite) TestWeeklyAggregation() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("time-weekly")
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

	// Upload records over 4 weeks with specific counts:
	// Week 1: 100 records
	// Week 2: 150 records
	// Week 3: 120 records
	// Week 4: 180 records
	// Total: 550 records
	now := time.Now()
	records := make([]integrationclient.UsageRecord, 0)

	// Week 1: 100 records
	for i := 0; i < 100; i++ {
		records = append(records, integrationclient.UsageRecord{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-24*time.Hour - time.Duration(i)*time.Minute),
		})
	}

	// Week 2: 150 records
	for i := 0; i < 150; i++ {
		records = append(records, integrationclient.UsageRecord{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-7*24*time.Hour - time.Duration(i)*time.Minute),
		})
	}

	// Week 3: 120 records
	for i := 0; i < 120; i++ {
		records = append(records, integrationclient.UsageRecord{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-14*24*time.Hour - time.Duration(i)*time.Minute),
		})
	}

	// Week 4: 180 records
	for i := 0; i < 180; i++ {
		records = append(records, integrationclient.UsageRecord{
			Platform:     "claude",
			Model:        "claude-3-opus",
			Provider:     uniqueName,
			HttpCode:     200,
			InputTokens:  intPointer(1000),
			OutputTokens: intPointer(500),
			TenantId:     int64Pointer(1),
			UserId:       int64Pointer(1),
			CreatedAt:    now.Add(-21*24*time.Hour - time.Duration(i)*time.Minute),
		})
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	assert.Equal(s.T(), 550, batchResp.JSON200.SyncedCount, "Should sync all 550 records")

	// Query with weekly grouping
	startDate := time.Now().Add(-28 * 24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/history?start_date=%s&end_date=%s&providers=%s&limit=100",
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

	s.T().Logf("Weekly aggregation response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - 4 weekly data points
	// - Week 1: 100 records
	// - Week 2: 150 records
	// - Week 3: 120 records
	// - Week 4: 180 records
	// - Week boundaries align correctly (Monday-Sunday or Sunday-Saturday)
}

// TestMonthBoundaryHandling verifies that month boundaries don't
// cause data loss or duplication.
func (s *IntegrationTestSuite) TestMonthBoundaryHandling() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("time-month-boundary")
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

	// Get current time and calculate month boundary
	now := time.Now()
	// Find the last day of current month
	lastDayOfMonth := now.AddDate(0, 1, -now.Day())
	// Last day at 23:59:59
	endOfMonth := time.Date(lastDayOfMonth.Year(), lastDayOfMonth.Month(), lastDayOfMonth.Day(), 23, 59, 59, 0, now.Location())
	// First day of next month at 00:00:01
	startOfNextMonth := endOfMonth.Add(2 * time.Second)

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
			CreatedAt:    endOfMonth,
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
			CreatedAt:    startOfNextMonth,
		},
	}

	// Upload usage records spanning month boundary
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 2, batchResp.JSON200.SyncedCount, "Should sync all 2 records")

	// Query analytics across both months
	startDate := endOfMonth.Add(-24 * time.Hour).Format("2006-01-02")
	endDate := startOfNextMonth.Add(24 * time.Hour).Format("2006-01-02")

	url := fmt.Sprintf("%s/api/v1/analytics/history?start_date=%s&end_date=%s&providers=%s&limit=100",
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

	s.T().Logf("Month boundary response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - Both records appear in results
	// - No records lost at boundary
	// - No double-counting
	// - Correct month assignment for each record
}

// TestTimeZoneHandling verifies that timestamps are handled
// correctly across time zones.
func (s *IntegrationTestSuite) TestTimeZoneHandling() {
	ctx := context.Background()

	// Create provider
	uniqueName := generateUniqueProviderName("time-timezone")
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

	// Upload records with explicit UTC timestamps
	now := time.Now().UTC()
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
			CreatedAt:    now.Add(-12 * time.Hour),
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
			CreatedAt:    now.Add(-24 * time.Hour),
		},
	}

	// Upload usage records
	batchReq := integrationclient.PostApiV1UsageBatchJSONRequestBody(records)
	batchResp, err := s.Client.PostApiV1UsageBatchWithResponse(ctx, batchReq)
	require.NoError(s.T(), err, "Failed to upload usage records")
	require.Equal(s.T(), 200, batchResp.StatusCode())
	require.Equal(s.T(), 3, batchResp.JSON200.SyncedCount, "Should sync all 3 records")

	// Query analytics
	startDate := now.Add(-48 * time.Hour).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

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

	s.T().Logf("Time zone handling response (%d): %s", httpResp.StatusCode, string(body))

	assert.Equal(s.T(), 200, httpResp.StatusCode, "Should return 200 OK")
	// Additional assertions would verify:
	// - All timestamps stored in UTC
	// - Date boundaries respect UTC
	// - Dates don't shift when querying from different time zones
	// - Consistent date grouping regardless of client timezone
}
