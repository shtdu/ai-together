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

package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

func TestUsageService_GetCurrentUsage_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	usageData := map[string]interface{}{
		"total_tokens":   1000,
		"total_cost":     0.50,
		"total_requests": 10,
	}

	mockRepo.On("GetCurrentUsageForUser", mock.Anything, int64(1)).Return(usageData, nil)

	usage, err := service.GetCurrentUsage(context.Background(), 1, 1)

	require.NoError(t, err)
	assert.Equal(t, 1000, usage["total_tokens"])

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetCurrentUsage_RepositoryError(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	mockRepo.On("GetCurrentUsageForUser", mock.Anything, int64(1)).Return(nil, assert.AnError)

	usage, err := service.GetCurrentUsage(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Nil(t, usage)
	assert.Contains(t, err.Error(), "failed to get current usage")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetUsageStats_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	statsData := map[string]interface{}{
		"total_tokens":   50000,
		"total_cost":     25.00,
		"total_requests": 500,
	}

	mockRepo.On("GetUsageStatsForTeam", mock.Anything, int64(1)).Return(statsData, nil)

	stats, err := service.GetUsageStats(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, 50000, stats["total_tokens"])

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetUsageStats_RepositoryError(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	mockRepo.On("GetUsageStatsForTeam", mock.Anything, int64(1)).Return(nil, assert.AnError)

	stats, err := service.GetUsageStats(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "failed to get usage stats")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_CreateUsageRecord_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	record := models.UsageRecord{
		ID:       1,
		Platform: "openai",
		Model:    "gpt-4",
		TenantID: 1,
		UserID:   1,
	}

	mockRepo.On("CreateUsageRecord", mock.Anything, record).Return(nil)

	err := service.CreateUsageRecord(context.Background(), record)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestUsageService_CreateUsageRecord_RepositoryError(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	record := models.UsageRecord{
		ID:       1,
		Platform: "openai",
		Model:    "gpt-4",
	}

	mockRepo.On("CreateUsageRecord", mock.Anything, record).Return(assert.AnError)

	err := service.CreateUsageRecord(context.Background(), record)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create usage record")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetProviderStats_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	stats := map[string]interface{}{
		"total_requests": 100,
		"total_tokens":   10000,
	}

	mockRepo.On("GetProviderStats", mock.Anything, "claude", int64(1)).Return(stats, nil)

	result, err := service.GetProviderStats(context.Background(), "claude", 1)

	require.NoError(t, err)
	assert.Equal(t, 100, result["total_requests"])

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetProviderStats_RepositoryError(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	mockRepo.On("GetProviderStats", mock.Anything, "claude", int64(1)).Return(nil, assert.AnError)

	stats, err := service.GetProviderStats(context.Background(), "claude", 1)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "failed to get provider stats")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetDashboardMetrics_Manager(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	dataPoints := []map[string]interface{}{
		{
			"active_time_seconds": float64(3600),
			"total_tokens":        int64(1000),
			"request_count":       int64(10),
		},
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	mockRepo.On("GetMetricsByTenant", mock.Anything, int64(1), startTime, endTime, "daily").
		Return(dataPoints, nil)

	metrics, err := service.GetDashboardMetrics(context.Background(), 1, 1, "manager", startTime, endTime, "daily")

	require.NoError(t, err)
	assert.Contains(t, metrics, "summary")
	assert.Contains(t, metrics, "data_points")

	summary := metrics["summary"].(map[string]interface{})
	assert.Equal(t, float64(1), summary["total_active_time_hours"])

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetDashboardMetrics_Member(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	dataPoints := []map[string]interface{}{
		{
			"active_time_seconds": float64(1800),
			"total_tokens":        int64(500),
			"request_count":       int64(5),
		},
	}

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	mockRepo.On("GetMetricsByUser", mock.Anything, int64(1), startTime, endTime, "daily").
		Return(dataPoints, nil)

	metrics, err := service.GetDashboardMetrics(context.Background(), 1, 1, "member", startTime, endTime, "daily")

	require.NoError(t, err)
	assert.Contains(t, metrics, "summary")

	summary := metrics["summary"].(map[string]interface{})
	assert.Equal(t, float64(0.5), summary["total_active_time_hours"])

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetDashboardMetrics_RepositoryError(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	mockRepo.On("GetMetricsByTenant", mock.Anything, int64(1), startTime, endTime, "daily").
		Return([]map[string]interface{}{}, assert.AnError)

	metrics, err := service.GetDashboardMetrics(context.Background(), 1, 1, "manager", startTime, endTime, "daily")

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "failed to get dashboard metrics")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetProviderRankings_Manager(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	rankings := []map[string]interface{}{
		{"provider": "claude", "request_count": 100},
		{"provider": "openai", "request_count": 50},
	}

	mockRepo.On("GetProviderRankingsByTenant", mock.Anything, int64(1)).Return(rankings, nil)

	result, err := service.GetProviderRankings(context.Background(), 1, 1, "manager")

	require.NoError(t, err)
	assert.Equal(t, "24h", result["period"])
	assert.Contains(t, result, "rankings")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetProviderRankings_Member(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	rankings := []map[string]interface{}{
		{"provider": "claude", "request_count": 20},
	}

	mockRepo.On("GetProviderRankingsByUser", mock.Anything, int64(1)).Return(rankings, nil)

	result, err := service.GetProviderRankings(context.Background(), 1, 1, "member")

	require.NoError(t, err)
	assert.Contains(t, result, "rankings")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetMemberStats_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	members := []map[string]interface{}{
		{"user_id": float64(1), "total_requests": float64(100)},
		{"user_id": float64(2), "total_requests": float64(50)},
	}

	mockRepo.On("GetMemberStatsByTenant", mock.Anything, int64(1)).Return(members, nil)

	stats, err := service.GetMemberStats(context.Background(), 1)

	require.NoError(t, err)
	assert.Contains(t, stats, "members")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetMemberStats_RepositoryError(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	mockRepo.On("GetMemberStatsByTenant", mock.Anything, int64(1)).Return([]map[string]interface{}{}, assert.AnError)

	stats, err := service.GetMemberStats(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "failed to get member stats")

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetProviderAnalytics_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	result := map[string]interface{}{
		"providers":      []string{"claude", "openai"},
		"total_requests": float64(150),
	}

	mockRepo.On("GetProviderAnalytics", mock.Anything, int64(1), "2024-01-01", "2024-01-31", []string{"claude"}, []string{"gpt-4"}, []string(nil)).
		Return(result, nil)

	analytics, err := service.GetProviderAnalytics(context.Background(), 1, "2024-01-01", "2024-01-31", []string{"claude"}, []string{"gpt-4"}, nil)

	require.NoError(t, err)
	assert.NotNil(t, analytics)

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetUserAnalytics_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	result := map[string]interface{}{
		"users":          []int64{1, 2},
		"total_requests": float64(200),
	}

	mockRepo.On("GetUserAnalytics", mock.Anything, int64(1), "2024-01-01", "2024-01-31", []int64{1, 2}, []string{"claude"}, []string(nil)).
		Return(result, nil)

	analytics, err := service.GetUserAnalytics(context.Background(), 1, "2024-01-01", "2024-01-31", []int64{1, 2}, []string{"claude"}, nil)

	require.NoError(t, err)
	assert.NotNil(t, analytics)

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetHistory_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	result := map[string]interface{}{
		"records": []models.UsageRecord{},
		"total":   float64(100),
		"page":    float64(1),
		"limit":   float64(10),
	}

	mockRepo.On("GetHistory", mock.Anything, int64(1), "2024-01-01", "2024-01-31", 1, 10, []int64{1}, []string{"claude"}, []string{"gpt-4"}, []string(nil), "created_at", "desc").
		Return(result, nil)

	history, err := service.GetHistory(context.Background(), 1, "2024-01-01", "2024-01-31", 1, 10, []int64{1}, []string{"claude"}, []string{"gpt-4"}, nil, "created_at", "desc")

	require.NoError(t, err)
	assert.NotNil(t, history)

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetFilterOptions_Success(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	result := map[string]interface{}{
		"providers": []string{"claude", "openai"},
		"models":    []string{"gpt-4", "claude-3"},
	}

	mockRepo.On("GetFilterOptions", mock.Anything, int64(1)).Return(result, nil)

	options, err := service.GetFilterOptions(context.Background(), 1)

	require.NoError(t, err)
	assert.NotNil(t, options)

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetProviderAnalytics_WithToolsFilter(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	result := map[string]interface{}{
		"distribution": map[string]interface{}{
			"by_tool": []map[string]interface{}{
				{"name": "claude", "tokens": 5000},
			},
		},
	}

	mockRepo.On("GetProviderAnalytics", mock.Anything, int64(1), "2024-01-01", "2024-01-31", []string(nil), []string(nil), []string{"claude"}).
		Return(result, nil)

	analytics, err := service.GetProviderAnalytics(context.Background(), 1, "2024-01-01", "2024-01-31", nil, nil, []string{"claude"})

	require.NoError(t, err)
	assert.NotNil(t, analytics)

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetUserAnalytics_WithToolsFilter(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	result := map[string]interface{}{
		"leaderboard": []map[string]interface{}{
			{"user_id": int64(1), "total_tokens": 1000},
		},
	}

	mockRepo.On("GetUserAnalytics", mock.Anything, int64(1), "2024-01-01", "2024-01-31", []int64(nil), []string(nil), []string{"codex"}).
		Return(result, nil)

	analytics, err := service.GetUserAnalytics(context.Background(), 1, "2024-01-01", "2024-01-31", nil, nil, []string{"codex"})

	require.NoError(t, err)
	assert.NotNil(t, analytics)

	mockRepo.AssertExpectations(t)
}

func TestUsageService_GetHistory_WithToolsFilter(t *testing.T) {
	mockRepo := new(MockUsageRepository)
	service := NewUsageService(mockRepo)

	result := map[string]interface{}{
		"records":    []models.UsageRecord{},
		"pagination": map[string]interface{}{"total": 0},
	}

	mockRepo.On("GetHistory", mock.Anything, int64(1), "2024-01-01", "2024-01-31", 1, 10, []int64(nil), []string(nil), []string(nil), []string{"opencode"}, "created_at", "desc").
		Return(result, nil)

	history, err := service.GetHistory(context.Background(), 1, "2024-01-01", "2024-01-31", 1, 10, nil, nil, nil, []string{"opencode"}, "created_at", "desc")

	require.NoError(t, err)
	assert.NotNil(t, history)

	mockRepo.AssertExpectations(t)
}
