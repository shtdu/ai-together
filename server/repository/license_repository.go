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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"switch-server/internal/db"
	"switch-server/models"
)

type LicenseRepository struct {
	db *db.DB
}

func NewLicenseRepository(database *db.DB) *LicenseRepository {
	return &LicenseRepository{
		db: database,
	}
}

// GetTenantLicense retrieves the license for a tenant
func (r *LicenseRepository) GetTenantLicense(ctx context.Context, tenantID int64) (*models.License, error) {
	row, err := r.db.GetTenantLicense(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant license: %w", err)
	}

	license := &models.License{
		TenantID:     tenantID,
		CustomerName: row.Name,
	}

	// Map nullable fields from database
	if row.LicenseID.Valid {
		license.LicenseID = row.LicenseID.String
	}
	if row.LicenseTier.Valid {
		license.Tier = row.LicenseTier.String
	}
	if row.LicenseSeats.Valid {
		license.Seats = int(row.LicenseSeats.Int32)
	}
	if row.LicenseKey.Valid {
		license.LicenseKey = row.LicenseKey.String
	}
	if row.LicenseIssuedAt.Valid {
		license.IssuedAt = row.LicenseIssuedAt.Time
	}
	if row.LicenseExpiresAt.Valid {
		license.ExpiresAt = row.LicenseExpiresAt.Time
	}
	// Map new license type fields
	if row.LicenseType.Valid {
		license.LicenseType = row.LicenseType.String
	}
	if row.LicenseMaxSeats.Valid {
		license.MaxSeats = int(row.LicenseMaxSeats.Int32)
	}
	if row.LicenseMaxTeams.Valid {
		license.MaxTeams = int(row.LicenseMaxTeams.Int32)
	}
	if row.LicenseDataRetentionDays.Valid {
		license.DataRetentionDays = int(row.LicenseDataRetentionDays.Int32)
	}

	return license, nil
}

// GetAllLicenses retrieves all licenses from all tenants
func (r *LicenseRepository) GetAllLicenses(ctx context.Context) ([]*models.License, error) {
	rows, err := r.db.GetAllLicenses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all licenses: %w", err)
	}

	licenses := make([]*models.License, 0, len(rows))
	for _, row := range rows {
		license := &models.License{
			TenantID:     row.ID,
			CustomerName: row.Name,
		}

		// Map nullable fields from database
		if row.LicenseID.Valid {
			license.LicenseID = row.LicenseID.String
		}
		if row.LicenseTier.Valid {
			license.Tier = row.LicenseTier.String
		}
		if row.LicenseSeats.Valid {
			license.Seats = int(row.LicenseSeats.Int32)
		}
		if row.LicenseKey.Valid {
			license.LicenseKey = row.LicenseKey.String
		}
		if row.LicenseIssuedAt.Valid {
			license.IssuedAt = row.LicenseIssuedAt.Time
		}
		if row.LicenseExpiresAt.Valid {
			license.ExpiresAt = row.LicenseExpiresAt.Time
		}
		// Map new license type fields
		if row.LicenseType.Valid {
			license.LicenseType = row.LicenseType.String
		}
		if row.LicenseMaxSeats.Valid {
			license.MaxSeats = int(row.LicenseMaxSeats.Int32)
		}
		if row.LicenseMaxTeams.Valid {
			license.MaxTeams = int(row.LicenseMaxTeams.Int32)
		}
		if row.LicenseDataRetentionDays.Valid {
			license.DataRetentionDays = int(row.LicenseDataRetentionDays.Int32)
		}

		licenses = append(licenses, license)
	}

	return licenses, nil
}

// UpdateTenantLicense updates the license for a tenant
func (r *LicenseRepository) UpdateTenantLicense(ctx context.Context, tenantID int64, license *models.License) error {
	params := db.UpdateTenantLicenseParams{
		ID: tenantID,
	}

	// Set nullable fields
	if license.LicenseID != "" {
		params.LicenseID = pgText(license.LicenseID)
	}
	if license.Tier != "" {
		params.LicenseTier = pgText(license.Tier)
	}
	if license.Seats > 0 {
		params.LicenseSeats = pgInt4(int32(license.Seats))
	}
	if license.LicenseKey != "" {
		params.LicenseKey = pgText(license.LicenseKey)
	}
	if !license.IssuedAt.IsZero() {
		params.LicenseIssuedAt = pgTimestamp(license.IssuedAt)
	}
	if !license.ExpiresAt.IsZero() {
		params.LicenseExpiresAt = pgTimestamp(license.ExpiresAt)
	}
	// Set new license type fields
	if license.LicenseType != "" {
		params.LicenseType = pgText(license.LicenseType)
	}
	if license.MaxSeats != 0 {
		params.LicenseMaxSeats = pgInt4(int32(license.MaxSeats))
	}
	if license.MaxTeams != 0 {
		params.LicenseMaxTeams = pgInt4(int32(license.MaxTeams))
	}
	if license.DataRetentionDays != 0 {
		params.LicenseDataRetentionDays = pgInt4(int32(license.DataRetentionDays))
	}

	err := r.db.UpdateTenantLicense(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update tenant license: %w", err)
	}

	return nil
}

// CountTenantUsers counts the number of users for a tenant
func (r *LicenseRepository) CountTenantUsers(ctx context.Context, tenantID int64) (int, error) {
	count, err := r.db.CountTenantUsers(ctx, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenant users: %w", err)
	}
	return int(count), nil
}

// CountTenantProvidersByKind counts the number of providers of a specific kind for a tenant
func (r *LicenseRepository) CountTenantProvidersByKind(ctx context.Context, tenantID int64, kind string) (int, error) {
	params := db.CountTenantProvidersByKindParams{
		TenantID: tenantID,
		Kind:     kind,
	}
	count, err := r.db.CountTenantProvidersByKind(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenant providers by kind: %w", err)
	}
	return int(count), nil
}

// CountTenantTeams counts the number of teams for a tenant
func (r *LicenseRepository) CountTenantTeams(ctx context.Context, tenantID int64) (int, error) {
	count, err := r.db.CountTenantTeams(ctx, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenant teams: %w", err)
	}
	return int(count), nil
}

// GetProviderCountsByKind returns provider counts grouped by kind for a tenant
func (r *LicenseRepository) GetProviderCountsByKind(ctx context.Context, tenantID int64) (map[string]int, error) {
	rows, err := r.db.GetProviderCountsByKind(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider counts by kind: %w", err)
	}

	counts := make(map[string]int)
	for _, row := range rows {
		counts[row.Kind] = int(row.Count)
	}

	return counts, nil
}

// Helper functions for pgtype conversion
func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func pgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func pgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}
