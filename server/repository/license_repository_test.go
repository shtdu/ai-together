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

//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"switch-server/models"
	"switch-server/testutil"
)

// TestLicenseRepository_GetTenantLicense_Success tests retrieving tenant license
func TestLicenseRepository_GetTenantLicense_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	testutil.CreateTestLicense(ctx, t, database.WithTx(tx), tenantID, "LICENSE-123", "professional", 100)

	repo := NewLicenseRepository(database)
	license, err := repo.GetTenantLicense(ctx, tenantID)

	require.NoError(t, err)
	assert.NotNil(t, license)
	assert.Equal(t, tenantID, license.TenantID)
	assert.Equal(t, "test-tenant", license.CustomerName)
	assert.Equal(t, "LICENSE-123", license.LicenseID)
	assert.Equal(t, "professional", license.Tier)
	assert.Equal(t, 100, license.Seats)
}

// TestLicenseRepository_GetTenantLicense_NotFound tests getting license for non-existent tenant
func TestLicenseRepository_GetTenantLicense_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewLicenseRepository(database)
	license, err := repo.GetTenantLicense(ctx, 99999)

	require.NoError(t, err)
	assert.NotNil(t, license)
	// Should return empty license for non-existent tenant
	assert.Empty(t, license.LicenseID)
}

// TestLicenseRepository_GetAllLicenses_Success tests getting all licenses
func TestLicenseRepository_GetAllLicenses_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenant1ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-1")
	tenant2ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-2")

	testutil.CreateTestLicense(ctx, t, database.WithTx(tx), tenant1ID, "LICENSE-1", "professional", 100)
	testutil.CreateTestLicense(ctx, t, database.WithTx(tx), tenant2ID, "LICENSE-2", "starter", 10)

	repo := NewLicenseRepository(database)
	licenses, err := repo.GetAllLicenses(ctx)

	require.NoError(t, err)
	assert.Len(t, licenses, 2)
}

// TestLicenseRepository_GetAllLicenses_Empty tests getting licenses when none exist
func TestLicenseRepository_GetAllLicenses_Empty(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewLicenseRepository(database)
	licenses, err := repo.GetAllLicenses(ctx)

	require.NoError(t, err)
	assert.Empty(t, licenses)
}

// TestLicenseRepository_UpdateTenantLicense_AllFields tests updating all license fields
func TestLicenseRepository_UpdateTenantLicense_AllFields(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	txDB := database.WithTx(tx)

	repo := NewLicenseRepository(database)
	license := &models.License{
		TenantID:     tenantID,
		CustomerName: "test-tenant",
		LicenseID:    "LICENSE-456",
		Tier:         "enterprise",
		Seats:        500,
		LicenseKey:   "test-key-456",
		IssuedAt:     time.Now(),
		ExpiresAt:    time.Now().AddDate(1, 0, 0),
	}

	err := repo.UpdateTenantLicense(ctx, tenantID, license)

	require.NoError(t, err)

	// Verify update
	updated, _ := repo.GetTenantLicense(ctx, tenantID)
	assert.Equal(t, "LICENSE-456", updated.LicenseID)
	assert.Equal(t, "enterprise", updated.Tier)
	assert.Equal(t, 500, updated.Seats)
}

// TestLicenseRepository_UpdateTenantLicense_PartialFields tests updating partial license fields
func TestLicenseRepository_UpdateTenantLicense_PartialFields(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	txDB := database.WithTx(tx)

	repo := NewLicenseRepository(database)
	license := &models.License{
		TenantID:     tenantID,
		CustomerName: "test-tenant",
		LicenseID:    "LICENSE-789",
		Tier:         "business",
	}

	err := repo.UpdateTenantLicense(ctx, tenantID, license)

	require.NoError(t, err)

	// Verify partial update
	updated, _ := repo.GetTenantLicense(ctx, tenantID)
	assert.Equal(t, "LICENSE-789", updated.LicenseID)
	assert.Equal(t, "business", updated.Tier)
}

// TestLicenseRepository_UpdateTenantLicense_NotFound tests updating non-existent tenant
func TestLicenseRepository_UpdateTenantLicense_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewLicenseRepository(database)
	license := &models.License{
		TenantID:     99999,
		CustomerName: "non-existent",
		LicenseID:    "LICENSE-999",
		Tier:         "enterprise",
	}

	err := repo.UpdateTenantLicense(ctx, 99999, license)

	assert.Error(t, err)
}

// TestLicenseRepository_CountTenantUsers_Success tests counting users in tenant
func TestLicenseRepository_CountTenantUsers_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "user1@example.com", "User 1", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "user2@example.com", "User 2", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "user3@example.com", "User 3", "manager", string(hashedPassword))

	repo := NewLicenseRepository(database)
	count, err := repo.CountTenantUsers(ctx, tenantID)

	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

// TestLicenseRepository_CountTenantUsers_Zero tests counting users in empty tenant
func TestLicenseRepository_CountTenantUsers_Zero(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	repo := NewLicenseRepository(database)
	count, err := repo.CountTenantUsers(ctx, tenantID)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestLicenseRepository_CountTenantProvidersByKind_Success tests counting providers by kind
func TestLicenseRepository_CountTenantProvidersByKind_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	testutil.CreateTestProvider(ctx, t, txDB, "Provider 1", "https://api1.example.com", "key1", "claude", teamID, true)
	testutil.CreateTestProvider(ctx, t, txDB, "Provider 2", "https://api2.example.com", "key2", "claude", teamID, true)
	testutil.CreateTestProvider(ctx, t, txDB, "Provider 3", "https://api3.example.com", "key3", "codex", teamID, true)

	repo := NewLicenseRepository(database)
	count, err := repo.CountTenantProvidersByKind(ctx, tenantID, "claude")

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestLicenseRepository_GetProviderCountsByKind_Success tests getting provider counts by kind
func TestLicenseRepository_GetProviderCountsByKind_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	testutil.CreateTestProvider(ctx, t, txDB, "Provider 1", "https://api1.example.com", "key1", "claude", teamID, true)
	testutil.CreateTestProvider(ctx, t, txDB, "Provider 2", "https://api2.example.com", "key2", "codex", teamID, true)
	testutil.CreateTestProvider(ctx, t, txDB, "Provider 3", "https://api3.example.com", "key3", "openai", teamID, true)

	repo := NewLicenseRepository(database)
	counts, err := repo.GetProviderCountsByKind(ctx, tenantID)

	require.NoError(t, err)
	assert.Equal(t, 2, counts["claude"])
	assert.Equal(t, 1, counts["codex"])
	assert.Equal(t, 1, counts["openai"])
}

// TestLicenseRepository_GetProviderCountsByKind_Empty tests getting provider counts for tenant with no providers
func TestLicenseRepository_GetProviderCountsByKind_Empty(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	repo := NewLicenseRepository(database)
	counts, err := repo.GetProviderCountsByKind(ctx, tenantID)

	require.NoError(t, err)
	assert.Empty(t, counts)
}

// TestLicenseRepository_NULLFields tests handling of NULL license fields
func TestLicenseRepository_NULLFields(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	repo := NewLicenseRepository(database)
	license, err := repo.GetTenantLicense(ctx, tenantID)

	require.NoError(t, err)
	// Tenant created without license should have NULL fields
	assert.Empty(t, license.LicenseID)
	assert.Empty(t, license.Tier)
	assert.Equal(t, 0, license.Seats)
}

// TestLicenseRepository_TenantIsolation tests licenses don't leak across tenants
func TestLicenseRepository_TenantIsolation(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenant1ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-1")
	tenant2ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-2")

	testutil.CreateTestLicense(ctx, t, database.WithTx(tx), tenant1ID, "LICENSE-1", "professional", 100)
	testutil.CreateTestLicense(ctx, t, database.WithTx(tx), tenant2ID, "LICENSE-2", "starter", 10)

	repo := NewLicenseRepository(database)
	licenses, _ := repo.GetAllLicenses(ctx)

	// Verify tenant isolation
	for _, lic := range licenses {
		if lic.TenantID == tenant1ID {
			assert.Equal(t, "LICENSE-1", lic.LicenseID)
		}
		if lic.TenantID == tenant2ID {
			assert.Equal(t, "LICENSE-2", lic.LicenseID)
		}
	}
}
