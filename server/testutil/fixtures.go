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


package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"switch-server/internal/db"
)

// CreateTestTenant creates a test tenant in the database
// Returns the tenant ID
func CreateTestTenant(ctx context.Context, t *testing.T, tx pgx.Tx, name string) int64 {
	t.Helper()

	// Insert tenant directly using SQL
	query := `INSERT INTO tenants (name) VALUES ($1) RETURNING id`
	var tenantID int64
	err := tx.QueryRow(ctx, query, name).Scan(&tenantID)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	return tenantID
}

// CreateTestUser creates a test user in the database
// Returns the user ID
func CreateTestUser(ctx context.Context, t *testing.T, txDB *db.Queries, tenantID int64, email, name, role, password string) int64 {
	t.Helper()

	err := txDB.CreateUser(ctx, db.CreateUserParams{
		Email:    email,
		Name:     name,
		Password: password,
		Role:     role,
		TenantID: tenantID,
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Get the user by email to retrieve ID
	user, err := txDB.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to retrieve test user: %v", err)
	}

	return user.ID
}

// CreateTestTeam creates a test team in the database
// Returns the team ID
func CreateTestTeam(ctx context.Context, t *testing.T, txDB *db.Queries, name, description string, ownerID, tenantID int64) int64 {
	t.Helper()

	var descriptionPg pgtype.Text
	if description != "" {
		descriptionPg = pgtype.Text{String: description, Valid: true}
	} else {
		descriptionPg = pgtype.Text{Valid: false}
	}

	team, err := txDB.CreateTeam(ctx, db.CreateTeamParams{
		Name:        name,
		Description: descriptionPg,
		OwnerID:     ownerID,
		TenantID:    tenantID,
		Settings:    []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Failed to create test team: %v", err)
	}

	return team.ID
}

// CreateTestProvider creates a test provider in the database
// Returns the provider ID
func CreateTestProvider(ctx context.Context, t *testing.T, txDB *db.Queries, name, apiURL, apiKey, kind string, teamID int64, enabled bool) int64 {
	t.Helper()

	provider, err := txDB.CreateProvider(ctx, db.CreateProviderParams{
		Name:            name,
		ApiUrl:          apiURL,
		ApiKey:          apiKey,
		TeamID:          teamID,
		Enabled:         pgtype.Bool{Bool: enabled, Valid: true},
		Kind:            kind,
		ModelMapping:    []byte("{}"),
		SupportedModels: []byte("[]"),
		Level:           pgtype.Int4{Int32: 1, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create test provider: %v", err)
	}

	return provider.ID
}

// CreateTestUsageRecord creates a test usage record in the database
// Returns the record ID
func CreateTestUsageRecord(ctx context.Context, t *testing.T, txDB *db.Queries, tenantID, userID int64, platform, model, provider string) {
	t.Helper()

	err := txDB.CreateUsageRecord(ctx, db.CreateUsageRecordParams{
		Platform:          pgtype.Text{String: platform, Valid: true},
		Model:             pgtype.Text{String: model, Valid: true},
		Provider:          pgtype.Text{String: provider, Valid: true},
		HttpCode:          pgtype.Int4{Int32: 200, Valid: true},
		InputTokens:       pgtype.Int4{Int32: 100, Valid: true},
		OutputTokens:      pgtype.Int4{Int32: 50, Valid: true},
		CacheCreateTokens: pgtype.Int4{Int32: 0, Valid: true},
		CacheReadTokens:   pgtype.Int4{Int32: 0, Valid: true},
		ReasoningTokens:   pgtype.Int4{Int32: 0, Valid: true},
		IsStream:          pgtype.Bool{Bool: false, Valid: true},
		DurationSec:       pgtype.Float4{Float32: 1.5, Valid: true},
		TenantID:          tenantID,
		UserID:            userID,
	})
	if err != nil {
		t.Fatalf("Failed to create test usage record: %v", err)
	}
}

// CreateTestLicense creates a test license for a tenant
func CreateTestLicense(ctx context.Context, t *testing.T, txDB *db.Queries, tenantID int64, licenseID, tier string, seats int) {
	t.Helper()

	err := txDB.UpdateTenantLicense(ctx, db.UpdateTenantLicenseParams{
		ID:               tenantID,
		LicenseID:        pgtype.Text{String: licenseID, Valid: true},
		LicenseTier:      pgtype.Text{String: tier, Valid: true},
		LicenseSeats:     pgtype.Int4{Int32: int32(seats), Valid: true},
		LicenseKey:       pgtype.Text{String: "test-key-" + licenseID, Valid: true},
		LicenseIssuedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		LicenseExpiresAt: pgtype.Timestamp{Time: time.Now().AddDate(1, 0, 0), Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create test license: %v", err)
	}
}

// WaitForTestWait is a helper for testing delays/contexts
func WaitForTestWait() {
	time.Sleep(10 * time.Millisecond)
}
