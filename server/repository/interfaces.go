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

package repository

import (
	"context"
	"time"

	"switch-server/models"
)

// UserRepositoryInterface defines the contract for user repository operations
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, email, hashedPassword, name, role string, tenantID int64) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)
	ListUsersByTenant(ctx context.Context, tenantID int64) ([]models.User, error)
	UpdateUser(ctx context.Context, userID int64, name, role string) error
	UpdateUserWithPassword(ctx context.Context, userID int64, name, role, hashedPassword string) error
	DeleteUser(ctx context.Context, userID int64) error
}

// ProviderRepositoryInterface defines the contract for provider repository operations
type ProviderRepositoryInterface interface {
	CreateProvider(ctx context.Context, name, apiURL, apiKey, kind string, teamID int64, enabled bool, modelMapping map[string]string, supportedModels []string, level int) (*models.Provider, error)
	GetProviderByID(ctx context.Context, providerID int64) (*models.Provider, error)
	GetProvidersByTeamID(ctx context.Context, teamID int64) ([]models.Provider, error)
	UpdateProvider(ctx context.Context, providerID int64, name, apiURL, apiKey, kind string, enabled bool, modelMapping map[string]string, supportedModels []string, level int) error
	DeleteProvider(ctx context.Context, providerID int64) error
	EnableProvider(ctx context.Context, providerID int64) error
	DisableProvider(ctx context.Context, providerID int64) error
	CountProvidersByNameAndTeam(ctx context.Context, name string, teamID int64) (int64, error)
	CountProvidersByNameAndTeamExcludingID(ctx context.Context, name string, teamID, excludeID int64) (int64, error)
}

// UsageRepositoryInterface defines the contract for usage repository operations
type UsageRepositoryInterface interface {
	CreateUsageRecord(ctx context.Context, record models.UsageRecord) error
	GetCurrentUsageForUser(ctx context.Context, userID int64) (map[string]interface{}, error)
	GetUsageStatsForTeam(ctx context.Context, tenantID int64) (map[string]interface{}, error)
	GetUsageByTeamIDAndPeriod(ctx context.Context, tenantID int64, startDate, endDate string) ([]models.UsageRecord, error)
	GetUsageByUserIDAndPeriod(ctx context.Context, userID int64, startDate, endDate string) ([]models.UsageRecord, error)
	GetProviderStats(ctx context.Context, providerName string, tenantID int64) (map[string]interface{}, error)
	GetMetricsByTenant(ctx context.Context, tenantID int64, startTime, endTime time.Time, interval string) ([]map[string]interface{}, error)
	GetMetricsByUser(ctx context.Context, userID int64, startTime, endTime time.Time, interval string) ([]map[string]interface{}, error)
	GetProviderRankingsByTenant(ctx context.Context, tenantID int64) ([]map[string]interface{}, error)
	GetProviderRankingsByUser(ctx context.Context, userID int64) ([]map[string]interface{}, error)
	GetMemberStatsByTenant(ctx context.Context, tenantID int64) ([]map[string]interface{}, error)
	GetProviderAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, providers, models []string) (map[string]interface{}, error)
	GetUserAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, userIDs []int64, providers []string) (map[string]interface{}, error)
	GetHistory(ctx context.Context, tenantID int64, startDate, endDate string, page, limit int, userIDs []int64, providers, models []string, sortBy, sortOrder string) (map[string]interface{}, error)
	GetFilterOptions(ctx context.Context, tenantID int64) (map[string]interface{}, error)
}

// TeamRepositoryInterface defines the contract for team repository operations
type TeamRepositoryInterface interface {
	CreateTeam(ctx context.Context, name, description string, ownerID, tenantID int64, settings map[string]string) (*models.Team, error)
	GetTeamByID(ctx context.Context, teamID int64) (*models.Team, error)
	GetTeamsByUserID(ctx context.Context, userID int64) ([]models.Team, error)
	AddTeamMember(ctx context.Context, teamID, userID int64, role string) error
	RemoveTeamMember(ctx context.Context, teamID, userID int64) error
	GetTeamMembers(ctx context.Context, teamID int64) ([]models.User, error)
	UpdateTeam(ctx context.Context, teamID int64, name, description string, settings map[string]string) error
	DeleteTeam(ctx context.Context, teamID int64) error
}
