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

package services

import (
	"context"
	"fmt"
	"time"

	"switch-server/models"
	"switch-server/repository"
)

type UsageService struct {
	usageRepo repository.UsageRepositoryInterface
}

// NewUsageService creates and returns a new instance of UsageService
func NewUsageService(usageRepo repository.UsageRepositoryInterface) *UsageService {
	return &UsageService{
		usageRepo: usageRepo,
	}
}

func (s *UsageService) GetCurrentUsage(ctx context.Context, userID, tenantID int64) (map[string]interface{}, error) {
	usage, err := s.usageRepo.GetCurrentUsageForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	return usage, nil
}

func (s *UsageService) GetUsageStats(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	stats, err := s.usageRepo.GetUsageStatsForTeam(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	return stats, nil
}

func (s *UsageService) CreateUsageRecord(ctx context.Context, usage models.UsageRecord) error {
	err := s.usageRepo.CreateUsageRecord(ctx, usage)
	if err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}

	return nil
}

func (s *UsageService) GetUsageByTeamIDAndPeriod(ctx context.Context, tenantID int64, startDate, endDate string) ([]models.UsageRecord, error) {
	records, err := s.usageRepo.GetUsageByTeamIDAndPeriod(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by team ID and period: %w", err)
	}

	return records, nil
}

func (s *UsageService) GetUsageByUserIDAndPeriod(ctx context.Context, userID int64, startDate, endDate string) ([]models.UsageRecord, error) {
	records, err := s.usageRepo.GetUsageByUserIDAndPeriod(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by user ID and period: %w", err)
	}

	return records, nil
}

func (s *UsageService) GetProviderStats(ctx context.Context, providerName string, tenantID int64) (map[string]interface{}, error) {
	stats, err := s.usageRepo.GetProviderStats(ctx, providerName, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}

	return stats, nil
}

// GetDashboardMetrics returns time-series metrics for the dashboard
func (s *UsageService) GetDashboardMetrics(ctx context.Context, tenantID, userID int64, role string, startTime, endTime time.Time, interval string) (map[string]interface{}, error) {
	// For members, only show their own data; for managers, show all tenant data
	var dataPoints []map[string]interface{}
	var err error

	if role == "manager" {
		dataPoints, err = s.usageRepo.GetMetricsByTenant(ctx, tenantID, startTime, endTime, interval)
	} else {
		dataPoints, err = s.usageRepo.GetMetricsByUser(ctx, userID, startTime, endTime, interval)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard metrics: %w", err)
	}

	// Calculate summary
	var totalActiveTime float64
	var totalTokens, totalRequests int64
	for _, dp := range dataPoints {
		if v, ok := dp["active_time_seconds"].(float64); ok {
			totalActiveTime += v
		}
		if v, ok := dp["total_tokens"].(int64); ok {
			totalTokens += v
		}
		if v, ok := dp["request_count"].(int64); ok {
			totalRequests += v
		}
	}

	// Cost calculation: $0.001 per 1000 input + $0.002 per 1000 output (simplified: use average ratio)
	estimatedCost := float64(totalTokens) * 0.0015 / 1000

	return map[string]interface{}{
		"period": map[string]interface{}{
			"start": startTime,
			"end":   endTime,
		},
		"interval":    interval,
		"data_points": dataPoints,
		"summary": map[string]interface{}{
			"total_active_time_hours": totalActiveTime / 3600,
			"total_tokens":            totalTokens,
			"total_requests":          totalRequests,
			"estimated_cost":          estimatedCost,
		},
	}, nil
}

// GetProviderRankings returns provider rankings for the last 24 hours
func (s *UsageService) GetProviderRankings(ctx context.Context, tenantID, userID int64, role string) (map[string]interface{}, error) {
	var rankings []map[string]interface{}
	var err error

	if role == "manager" {
		rankings, err = s.usageRepo.GetProviderRankingsByTenant(ctx, tenantID)
	} else {
		rankings, err = s.usageRepo.GetProviderRankingsByUser(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get provider rankings: %w", err)
	}

	return map[string]interface{}{
		"period":   "24h",
		"rankings": rankings,
	}, nil
}

// GetMemberStats returns member statistics for the tenant
func (s *UsageService) GetMemberStats(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	members, err := s.usageRepo.GetMemberStatsByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member stats: %w", err)
	}

	return map[string]interface{}{
		"members": members,
	}, nil
}

// GetProviderAnalytics returns provider analytics with filtering
func (s *UsageService) GetProviderAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, providers, models []string) (map[string]interface{}, error) {
	result, err := s.usageRepo.GetProviderAnalytics(ctx, tenantID, startDate, endDate, providers, models)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider analytics: %w", err)
	}
	return result, nil
}

// GetUserAnalytics returns user analytics with filtering
func (s *UsageService) GetUserAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, userIDs []int64, providers []string) (map[string]interface{}, error) {
	result, err := s.usageRepo.GetUserAnalytics(ctx, tenantID, startDate, endDate, userIDs, providers)
	if err != nil {
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}
	return result, nil
}

// GetHistory returns paginated request logs
func (s *UsageService) GetHistory(ctx context.Context, tenantID int64, startDate, endDate string, page, limit int, userIDs []int64, providers, models []string, sortBy, sortOrder string) (map[string]interface{}, error) {
	result, err := s.usageRepo.GetHistory(ctx, tenantID, startDate, endDate, page, limit, userIDs, providers, models, sortBy, sortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	return result, nil
}

// GetFilterOptions returns available filter options
func (s *UsageService) GetFilterOptions(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	result, err := s.usageRepo.GetFilterOptions(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get filter options: %w", err)
	}
	return result, nil
}
