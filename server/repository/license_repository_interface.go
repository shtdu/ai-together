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


package repository

import (
	"context"

	"switch-server/models"
)

// LicenseRepositoryInterface defines the contract for license repository operations
// This interface enables mocking for unit tests
type LicenseRepositoryInterface interface {
	GetTenantLicense(ctx context.Context, tenantID int64) (*models.License, error)
	GetAllLicenses(ctx context.Context) ([]*models.License, error)
	UpdateTenantLicense(ctx context.Context, tenantID int64, license *models.License) error
	CountTenantUsers(ctx context.Context, tenantID int64) (int, error)
	CountTenantProvidersByKind(ctx context.Context, tenantID int64, kind string) (int, error)
	CountTenantTeams(ctx context.Context, tenantID int64) (int, error)
	GetProviderCountsByKind(ctx context.Context, tenantID int64) (map[string]int, error)
}
