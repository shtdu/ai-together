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
	"time"

	"switch-server/models"
)

// UserServiceInterface defines the contract for user service operations
// This interface enables mocking for testing
type UserServiceInterface interface {
	CreateUser(email, password, name, role string, tenantID int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(userID int64) (*models.User, error)
	ValidatePassword(user *models.User, password string) bool
	HashPassword(password string) (string, error)
	ListUsersByTenant(tenantID int64) ([]models.User, error)
	UpdateUser(userID int64, name, role, password string) (*models.User, error)
	DeleteUser(userID int64) error
	UpdateProfileName(userID int64, name string) (*models.User, error)
	ChangePassword(userID int64, currentPassword, newPassword string) error
}

// ProviderServiceInterface defines the contract for provider service operations
type ProviderServiceInterface interface {
	CreateProvider(name, apiURL, apiKey, kind string, teamID int64, enabled bool, modelMapping map[string]interface{}, supportedModels []string, level int) (*models.Provider, error)
	GetProviderByID(providerID int64) (*models.Provider, error)
	GetProvidersByTeamID(teamID int64) ([]models.Provider, error)
	UpdateProvider(providerID int64, name, apiURL, apiKey, kind string, enabled bool, modelMapping map[string]interface{}, supportedModels []string, level int) error
	DeleteProvider(providerID int64) error
	EnableProvider(providerID int64) error
	DisableProvider(providerID int64) error
	CountProvidersByNameAndTeam(ctx context.Context, name string, teamID int64) (int64, error)
	CountProvidersByNameAndTeamExcludingID(ctx context.Context, name string, teamID, excludeID int64) (int64, error)
}

// UsageServiceInterface defines the contract for usage service operations
type UsageServiceInterface interface {
	GetCurrentUsage(ctx context.Context, userID, tenantID int64) (map[string]interface{}, error)
	GetUsageStats(ctx context.Context, tenantID int64) (map[string]interface{}, error)
	CreateUsageRecord(ctx context.Context, record models.UsageRecord) error
	GetProviderStats(ctx context.Context, providerName string, tenantID int64) (map[string]interface{}, error)
	// Dashboard methods
	GetDashboardMetrics(ctx context.Context, tenantID, userID int64, role string, startTime, endTime time.Time, interval string) (map[string]interface{}, error)
	GetProviderRankings(ctx context.Context, tenantID, userID int64, role string) (map[string]interface{}, error)
	GetMemberStats(ctx context.Context, tenantID int64) (map[string]interface{}, error)
	// Analytics methods
	GetProviderAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, providers, models, tools []string) (map[string]interface{}, error)
	GetUserAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, userIDs []int64, providers, tools []string) (map[string]interface{}, error)
	GetHistory(ctx context.Context, tenantID int64, startDate, endDate string, page, limit int, userIDs []int64, providers, models, tools []string, sortBy, sortOrder string) (map[string]interface{}, error)
	GetFilterOptions(ctx context.Context, tenantID int64) (map[string]interface{}, error)
}

// TeamServiceInterface defines the contract for team service operations
type TeamServiceInterface interface {
	CreateTeam(name, description string, ownerID, tenantID int64, settings map[string]string) (*models.Team, error)
	GetTeamByID(teamID int64) (*models.Team, error)
	GetTeamsByUserID(userID int64) ([]models.Team, error)
	AddTeamMember(teamID, userID int64, role string) error
	RemoveTeamMember(teamID, userID int64) error
	GetTeamMembers(teamID int64) ([]models.User, error)
	UpdateTeam(teamID int64, name, description string, settings map[string]string) error
	DeleteTeam(teamID int64) error
}

// LicenseServiceInterface defines the contract for license service operations
type LicenseServiceInterface interface {
	GetLicense(ctx context.Context, tenantID int64) (*models.License, error)
	GetLicenseUsage(ctx context.Context, tenantID int64) (*models.LicenseUsage, error)
	CanCreateTeam(ctx context.Context, tenantID int64) (bool, error)
	IsLicenseValid(ctx context.Context, tenantID int64) (bool, string, error)
	ActivateLicense(ctx context.Context, tenantID int64, licenseKeyPEM string) error
	GetEffectiveLicense(ctx context.Context, tenantID int64) *models.License
	HasActiveLicense(ctx context.Context, tenantID int64) bool
	GetTiers() map[string]interface{}
	VerifyLicenseIntegrity(ctx context.Context, tenantID int64) (bool, string, error)
}
