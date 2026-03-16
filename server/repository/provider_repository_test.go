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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"switch-server/testutil"
)

// TestProviderRepository_CreateProvider_Success tests successful provider creation
func TestProviderRepository_CreateProvider_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	modelMapping := map[string]string{"model1": "alias1"}
	supportedModels := []string{"model1", "model2"}

	repo := NewProviderRepository(database)
	provider, err := repo.CreateProvider(ctx, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true, modelMapping, supportedModels, 1)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "Test Provider", provider.Name)
	assert.Equal(t, "https://api.example.com", provider.APIURL)
	assert.Equal(t, "test-key", provider.APIKey)
	assert.Equal(t, teamID, provider.TeamID)
	assert.Equal(t, "claude", provider.Kind)
	assert.True(t, provider.Enabled)
	assert.Equal(t, 1, provider.Level)
}

// TestProviderRepository_CreateProvider_WithEmptyFields tests provider with optional fields
func TestProviderRepository_CreateProvider_WithEmptyFields(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewProviderRepository(database)
	provider, err := repo.CreateProvider(ctx, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, false, map[string]string{}, []string{}, 0)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.False(t, provider.Enabled)
	assert.Equal(t, 0, provider.Level)
	assert.Empty(t, provider.ModelMapping)
	assert.Empty(t, provider.SupportedModels)
}

// TestProviderRepository_GetProviderByID_Success tests retrieving provider by ID
func TestProviderRepository_GetProviderByID_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	providerID := testutil.CreateTestProvider(ctx, t, txDB, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true)

	repo := NewProviderRepository(database)
	provider, err := repo.GetProviderByID(ctx, providerID)

	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, providerID, provider.ID)
	assert.Equal(t, "Test Provider", provider.Name)
}

// TestProviderRepository_GetProviderByID_NotFound tests error for non-existent provider
func TestProviderRepository_GetProviderByID_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewProviderRepository(database)
	provider, err := repo.GetProviderByID(ctx, 99999)

	assert.Error(t, err)
	assert.Nil(t, provider)
}

// TestProviderRepository_UpdateProvider_Success tests updating provider
func TestProviderRepository_UpdateProvider_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	providerID := testutil.CreateTestProvider(ctx, t, txDB, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true)

	newModelMapping := map[string]string{"new-model": "new-alias"}
	newSupportedModels := []string{"new-model"}

	repo := NewProviderRepository(database)
	err := repo.UpdateProvider(ctx, providerID, "Updated Provider", "https://new-api.example.com", "new-key", "codex", true, newModelMapping, newSupportedModels, 2)

	require.NoError(t, err)

	// Verify update
	provider, _ := repo.GetProviderByID(ctx, providerID)
	assert.Equal(t, "Updated Provider", provider.Name)
	assert.Equal(t, "https://new-api.example.com", provider.APIURL)
	assert.Equal(t, "codex", provider.Kind)
	assert.Equal(t, 2, provider.Level)
}

// TestProviderRepository_UpdateProvider_NotFound tests updating non-existent provider
func TestProviderRepository_UpdateProvider_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewProviderRepository(database)
	err := repo.UpdateProvider(ctx, 99999, "New Name", "https://api.example.com", "key", "claude", true, map[string]string{}, []string{}, 1)

	assert.Error(t, err)
}

// TestProviderRepository_DeleteProvider_Success tests deleting provider
func TestProviderRepository_DeleteProvider_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	providerID := testutil.CreateTestProvider(ctx, t, txDB, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true)

	repo := NewProviderRepository(database)
	err := repo.DeleteProvider(ctx, providerID)

	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetProviderByID(ctx, providerID)
	assert.Error(t, err)
}

// TestProviderRepository_DeleteProvider_NotFound tests deleting non-existent provider
func TestProviderRepository_DeleteProvider_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewProviderRepository(database)
	err := repo.DeleteProvider(ctx, 99999)

	assert.Error(t, err)
}

// TestProviderRepository_GetProvidersByTeamID_Success tests getting providers by team
func TestProviderRepository_GetProvidersByTeamID_Success(t *testing.T) {
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

	repo := NewProviderRepository(database)
	providers, err := repo.GetProvidersByTeamID(ctx, teamID)

	require.NoError(t, err)
	assert.Len(t, providers, 2)
}

// TestProviderRepository_GetProvidersByTeamID_Empty tests getting providers for team with no providers
func TestProviderRepository_GetProvidersByTeamID_Empty(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewProviderRepository(database)
	providers, err := repo.GetProvidersByTeamID(ctx, teamID)

	require.NoError(t, err)
	assert.Empty(t, providers)
}

// TestProviderRepository_EnableProvider_Success tests enabling provider
func TestProviderRepository_EnableProvider_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	providerID := testutil.CreateTestProvider(ctx, t, txDB, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, false)

	repo := NewProviderRepository(database)
	err := repo.EnableProvider(ctx, providerID)

	require.NoError(t, err)

	// Verify enabled
	provider, _ := repo.GetProviderByID(ctx, providerID)
	assert.True(t, provider.Enabled)
}

// TestProviderRepository_EnableProvider_NotFound tests enabling non-existent provider
func TestProviderRepository_EnableProvider_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewProviderRepository(database)
	err := repo.EnableProvider(ctx, 99999)

	assert.Error(t, err)
}

// TestProviderRepository_DisableProvider_Success tests disabling provider
func TestProviderRepository_DisableProvider_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	providerID := testutil.CreateTestProvider(ctx, t, txDB, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true)

	repo := NewProviderRepository(database)
	err := repo.DisableProvider(ctx, providerID)

	require.NoError(t, err)

	// Verify disabled
	provider, _ := repo.GetProviderByID(ctx, providerID)
	assert.False(t, provider.Enabled)
}

// TestProviderRepository_DisableProvider_NotFound tests disabling non-existent provider
func TestProviderRepository_DisableProvider_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewProviderRepository(database)
	err := repo.DisableProvider(ctx, 99999)

	assert.Error(t, err)
}

// TestProviderRepository_CountProvidersByNameAndTeam_Success tests counting providers
func TestProviderRepository_CountProvidersByNameAndTeam_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	testutil.CreateTestProvider(ctx, t, txDB, "Provider", "https://api.example.com", "key", "claude", teamID, true)
	testutil.CreateTestProvider(ctx, t, txDB, "Provider", "https://api2.example.com", "key2", "codex", teamID, true)

	repo := NewProviderRepository(database)
	count, err := repo.CountProvidersByNameAndTeam(ctx, "Provider", teamID)

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// TestProviderRepository_CountProvidersByNameAndTeam_Zero tests counting with no matches
func TestProviderRepository_CountProvidersByNameAndTeam_Zero(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewProviderRepository(database)
	count, err := repo.CountProvidersByNameAndTeam(ctx, "NonExistent", teamID)

	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestProviderRepository_CountProvidersByNameAndTeamExcludingID_Success tests counting excluding ID
func TestProviderRepository_CountProvidersByNameAndTeamExcludingID_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	providerID := testutil.CreateTestProvider(ctx, t, txDB, "Provider", "https://api.example.com", "key", "claude", teamID, true)
	testutil.CreateTestProvider(ctx, t, txDB, "Provider", "https://api2.example.com", "key2", "codex", teamID, true)

	repo := NewProviderRepository(database)
	count, err := repo.CountProvidersByNameAndTeamExcludingID(ctx, "Provider", teamID, providerID)

	require.NoError(t, err)
	// Should count only the second provider (first one is excluded)
	assert.Equal(t, int64(1), count)
}

// TestProviderRepository_TeamIsolation tests providers don't leak across teams
func TestProviderRepository_TeamIsolation(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	team1ID := testutil.CreateTestTeam(ctx, t, txDB, "Team 1", "Description 1", ownerID, tenantID)
	team2ID := testutil.CreateTestTeam(ctx, t, txDB, "Team 2", "Description 2", ownerID, tenantID)

	testutil.CreateTestProvider(ctx, t, txDB, "Provider", "https://api1.example.com", "key1", "claude", team1ID, true)
	testutil.CreateTestProvider(ctx, t, txDB, "Provider", "https://api2.example.com", "key2", "codex", team2ID, true)

	repo := NewProviderRepository(database)
	providers1, _ := repo.GetProvidersByTeamID(ctx, team1ID)
	providers2, _ := repo.GetProvidersByTeamID(ctx, team2ID)

	assert.Len(t, providers1, 1)
	assert.Len(t, providers2, 1)
	assert.NotEqual(t, providers1[0].ID, providers2[0].ID)
}

// TestProviderRepository_ModelMapping_JSON tests JSON marshaling of model mapping
func TestProviderRepository_ModelMapping_JSON(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	modelMapping := map[string]string{"model1": "alias1", "model2": "alias2"}
	supportedModels := []string{"model1", "model2", "model3"}

	repo := NewProviderRepository(database)
	provider, err := repo.CreateProvider(ctx, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true, modelMapping, supportedModels, 1)

	require.NoError(t, err)
	assert.Equal(t, modelMapping, provider.ModelMapping)
	assert.Equal(t, supportedModels, provider.SupportedModels)
}

// TestProviderRepository_ModelMapping_EmptyJSON tests empty JSON handling
func TestProviderRepository_ModelMapping_EmptyJSON(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	// Create provider with empty JSON
	providerID := testutil.CreateTestProvider(ctx, t, txDB, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true)

	repo := NewProviderRepository(database)
	provider, err := repo.GetProviderByID(ctx, providerID)

	require.NoError(t, err)
	// Empty JSON should be decoded to empty map
	assert.NotNil(t, provider.ModelMapping)
	assert.NotNil(t, provider.SupportedModels)
}

// TestProviderRepository_SupportedModels_JSON tests JSON marshaling of supported models
func TestProviderRepository_SupportedModels_JSON(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	supportedModels := []string{"claude-3-5-sonnet", "claude-3-opus"}
	modelMapping := map[string]string{"claude-3-5-sonnet": "sonnet"}

	repo := NewProviderRepository(database)
	provider, err := repo.CreateProvider(ctx, "Test Provider", "https://api.example.com", "test-key", "claude", teamID, true, modelMapping, supportedModels, 1)

	require.NoError(t, err)
	assert.Equal(t, supportedModels, provider.SupportedModels)
}
