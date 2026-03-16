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
	"time"

	"github.com/stretchr/testify/mock"
	"switch-server/models"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email, hashedPassword, name, role string, tenantID int64) (*models.User, error) {
	args := m.Called(ctx, email, hashedPassword, name, role, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) ListUsersByTenant(ctx context.Context, tenantID int64) ([]models.User, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, userID int64, name, role string) error {
	args := m.Called(ctx, userID, name, role)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserWithPassword(ctx context.Context, userID int64, name, role, hashedPassword string) error {
	args := m.Called(ctx, userID, name, role, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockProviderRepository is a mock implementation of ProviderRepository
type MockProviderRepository struct {
	mock.Mock
}

func (m *MockProviderRepository) CreateProvider(ctx context.Context, name, apiURL, apiKey, kind string, teamID int64, enabled bool, modelMapping map[string]string, supportedModels []string, level int) (*models.Provider, error) {
	args := m.Called(ctx, name, apiURL, apiKey, kind, teamID, enabled, modelMapping, supportedModels, level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Provider), args.Error(1)
}

func (m *MockProviderRepository) GetProviderByID(ctx context.Context, providerID int64) (*models.Provider, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Provider), args.Error(1)
}

func (m *MockProviderRepository) GetProvidersByTeamID(ctx context.Context, teamID int64) ([]models.Provider, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Provider), args.Error(1)
}

func (m *MockProviderRepository) UpdateProvider(ctx context.Context, providerID int64, name, apiURL, apiKey, kind string, enabled bool, modelMapping map[string]string, supportedModels []string, level int) error {
	args := m.Called(ctx, providerID, name, apiURL, apiKey, kind, enabled, modelMapping, supportedModels, level)
	return args.Error(0)
}

func (m *MockProviderRepository) DeleteProvider(ctx context.Context, providerID int64) error {
	args := m.Called(ctx, providerID)
	return args.Error(0)
}

func (m *MockProviderRepository) EnableProvider(ctx context.Context, providerID int64) error {
	args := m.Called(ctx, providerID)
	return args.Error(0)
}

func (m *MockProviderRepository) DisableProvider(ctx context.Context, providerID int64) error {
	args := m.Called(ctx, providerID)
	return args.Error(0)
}

func (m *MockProviderRepository) CountProvidersByNameAndTeam(ctx context.Context, name string, teamID int64) (int64, error) {
	args := m.Called(ctx, name, teamID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProviderRepository) CountProvidersByNameAndTeamExcludingID(ctx context.Context, name string, teamID, excludeID int64) (int64, error) {
	args := m.Called(ctx, name, teamID, excludeID)
	return args.Get(0).(int64), args.Error(1)
}

// MockUsageRepository is a mock implementation of UsageRepository
type MockUsageRepository struct {
	mock.Mock
}

func (m *MockUsageRepository) CreateUsageRecord(ctx context.Context, record models.UsageRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockUsageRepository) GetCurrentUsageForUser(ctx context.Context, userID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetUsageStatsForTeam(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetUsageByTeamIDAndPeriod(ctx context.Context, tenantID int64, startDate, endDate string) ([]models.UsageRecord, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UsageRecord), args.Error(1)
}

func (m *MockUsageRepository) GetUsageByUserIDAndPeriod(ctx context.Context, userID int64, startDate, endDate string) ([]models.UsageRecord, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UsageRecord), args.Error(1)
}

func (m *MockUsageRepository) GetProviderStats(ctx context.Context, providerName string, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, providerName, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetMetricsByTenant(ctx context.Context, tenantID int64, startTime, endTime time.Time, interval string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startTime, endTime, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetMetricsByUser(ctx context.Context, userID int64, startTime, endTime time.Time, interval string) ([]map[string]interface{}, error) {
	args := m.Called(ctx, userID, startTime, endTime, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetProviderRankingsByTenant(ctx context.Context, tenantID int64) ([]map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetProviderRankingsByUser(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetMemberStatsByTenant(ctx context.Context, tenantID int64) ([]map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetProviderAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, providers, models []string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, providers, models)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetUserAnalytics(ctx context.Context, tenantID int64, startDate, endDate string, userIDs []int64, providers []string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, userIDs, providers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetHistory(ctx context.Context, tenantID int64, startDate, endDate string, page, limit int, userIDs []int64, providers, models []string, sortBy, sortOrder string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, page, limit, userIDs, providers, models, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUsageRepository) GetFilterOptions(ctx context.Context, tenantID int64) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockTeamRepository is a mock implementation of TeamRepository
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) CreateTeam(ctx context.Context, name, description string, ownerID, tenantID int64, settings map[string]string) (*models.Team, error) {
	args := m.Called(ctx, name, description, ownerID, tenantID, settings)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetTeamByID(ctx context.Context, teamID int64) (*models.Team, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetTeamsByUserID(ctx context.Context, userID int64) ([]models.Team, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamRepository) AddTeamMember(ctx context.Context, teamID, userID int64, role string) error {
	args := m.Called(ctx, teamID, userID, role)
	return args.Error(0)
}

func (m *MockTeamRepository) RemoveTeamMember(ctx context.Context, teamID, userID int64) error {
	args := m.Called(ctx, teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamMembers(ctx context.Context, teamID int64) ([]models.User, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockTeamRepository) UpdateTeam(ctx context.Context, teamID int64, name, description string, settings map[string]string) error {
	args := m.Called(ctx, teamID, name, description, settings)
	return args.Error(0)
}

func (m *MockTeamRepository) DeleteTeam(ctx context.Context, teamID int64) error {
	args := m.Called(ctx, teamID)
	return args.Error(0)
}

// MockLicenseService is a mock implementation of LicenseService
type MockLicenseService struct {
	mock.Mock
}

// Ensure MockLicenseService implements LicenseServiceInterface
var _ LicenseServiceInterface = (*MockLicenseService)(nil)

func (m *MockLicenseService) GetLicense(ctx context.Context, tenantID int64) (*models.License, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.License), args.Error(1)
}

func (m *MockLicenseService) GetLicenseUsage(ctx context.Context, tenantID int64) (*models.LicenseUsage, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LicenseUsage), args.Error(1)
}

func (m *MockLicenseService) CanCreateTeam(ctx context.Context, tenantID int64) (bool, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLicenseService) IsLicenseValid(ctx context.Context, tenantID int64) (bool, string, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockLicenseService) ActivateLicense(ctx context.Context, tenantID int64, licenseKeyPEM string) error {
	args := m.Called(ctx, tenantID, licenseKeyPEM)
	return args.Error(0)
}

func (m *MockLicenseService) GetEffectiveLicense(ctx context.Context, tenantID int64) *models.License {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.License)
}

func (m *MockLicenseService) HasActiveLicense(ctx context.Context, tenantID int64) bool {
	args := m.Called(ctx, tenantID)
	return args.Bool(0)
}

func (m *MockLicenseService) GetTiers() map[string]interface{} {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[string]interface{})
}

func (m *MockLicenseService) VerifyLicenseIntegrity(ctx context.Context, tenantID int64) (bool, string, error) {
	args := m.Called(ctx, tenantID)
	return args.Bool(0), args.String(1), args.Error(2)
}
