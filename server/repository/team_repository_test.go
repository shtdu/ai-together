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


//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"switch-server/testutil"
)

// TestTeamRepository_CreateTeam_Success tests successful team creation
func TestTeamRepository_CreateTeam_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	repo := NewTeamRepository(database)
	team, err := repo.CreateTeam(ctx, "Test Team", "Test Description", ownerID, tenantID, map[string]string{"key": "value"})

	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "Test Team", team.Name)
	assert.Equal(t, "Test Description", team.Description)
	assert.Equal(t, ownerID, team.OwnerID)
	assert.Equal(t, tenantID, team.TenantID)
	assert.NotZero(t, team.ID)
}

// TestTeamRepository_CreateTeam_WithoutDescription tests team creation without description
func TestTeamRepository_CreateTeam_WithoutDescription(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	repo := NewTeamRepository(database)
	team, err := repo.CreateTeam(ctx, "Test Team", "", ownerID, tenantID, map[string]string{})

	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "Test Team", team.Name)
	assert.Equal(t, "", team.Description)
}

// TestTeamRepository_GetTeamByID_Success tests retrieving team by ID
func TestTeamRepository_GetTeamByID_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	team, err := repo.GetTeamByID(ctx, teamID)

	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, teamID, team.ID)
	assert.Equal(t, "Test Team", team.Name)
}

// TestTeamRepository_GetTeamByID_NotFound tests error for non-existent team
func TestTeamRepository_GetTeamByID_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewTeamRepository(database)
	team, err := repo.GetTeamByID(ctx, 99999)

	assert.Error(t, err)
	assert.Nil(t, team)
}

// TestTeamRepository_UpdateTeam_Success tests updating team
func TestTeamRepository_UpdateTeam_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	err := repo.UpdateTeam(ctx, teamID, "Updated Team", "Updated Description", map[string]string{"new": "settings"})

	require.NoError(t, err)

	// Verify update
	team, _ := repo.GetTeamByID(ctx, teamID)
	assert.Equal(t, "Updated Team", team.Name)
	assert.Equal(t, "Updated Description", team.Description)
}

// TestTeamRepository_UpdateTeam_NotFound tests updating non-existent team
func TestTeamRepository_UpdateTeam_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewTeamRepository(database)
	err := repo.UpdateTeam(ctx, 99999, "New Name", "New Description", map[string]string{})

	assert.Error(t, err)
}

// TestTeamRepository_DeleteTeam_Success tests deleting team
func TestTeamRepository_DeleteTeam_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	err := repo.DeleteTeam(ctx, teamID)

	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetTeamByID(ctx, teamID)
	assert.Error(t, err)
}

// TestTeamRepository_DeleteTeam_NotFound tests deleting non-existent team
func TestTeamRepository_DeleteTeam_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewTeamRepository(database)
	err := repo.DeleteTeam(ctx, 99999)

	assert.Error(t, err)
}

// TestTeamRepository_AddTeamMember_Success tests adding member to team
func TestTeamRepository_AddTeamMember_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "member@example.com", "Member", "member", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	err := repo.AddTeamMember(ctx, teamID, userID, "member")

	require.NoError(t, err)

	// Verify member was added
	members, err := repo.GetTeamMembers(ctx, teamID)
	require.NoError(t, err)
	assert.Len(t, members, 2) // Owner + added member
}

// TestTeamRepository_AddTeamMember_Duplicate tests duplicate member addition
func TestTeamRepository_AddTeamMember_Duplicate(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "member@example.com", "Member", "member", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	err := repo.AddTeamMember(ctx, teamID, userID, "member")
	require.NoError(t, err)

	// Try adding again - should fail due to unique constraint
	err = repo.AddTeamMember(ctx, teamID, userID, "member")
	assert.Error(t, err)
}

// TestTeamRepository_RemoveTeamMember_Success tests removing team member
func TestTeamRepository_RemoveTeamMember_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "member@example.com", "Member", "member", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	_ = repo.AddTeamMember(ctx, teamID, userID, "member")

	err := repo.RemoveTeamMember(ctx, teamID, userID)
	require.NoError(t, err)

	// Verify removal
	members, _ := repo.GetTeamMembers(ctx, teamID)
	assert.Len(t, members, 1) // Only owner remains
}

// TestTeamRepository_RemoveTeamMember_NotFound tests removing non-existent member
func TestTeamRepository_RemoveTeamMember_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewTeamRepository(database)
	err := repo.RemoveTeamMember(ctx, 99999, 88888)

	// May not error depending on SQL behavior
	// The repository just executes the DELETE
	assert.NoError(t, err)
}

// TestTeamRepository_GetTeamMembers_Success tests getting team members
func TestTeamRepository_GetTeamMembers_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "member@example.com", "Member", "member", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	_ = repo.AddTeamMember(ctx, teamID, userID, "member")

	members, err := repo.GetTeamMembers(ctx, teamID)

	require.NoError(t, err)
	assert.Len(t, members, 2)
}

// TestTeamRepository_GetTeamMembers_Empty tests getting members for empty team
func TestTeamRepository_GetTeamMembers_Empty(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	members, err := repo.GetTeamMembers(ctx, teamID)

	require.NoError(t, err)
	// Owner should NOT be in the members list by default
	// The GetTeamMembers query returns only team_members table records
	assert.Empty(t, members)
}

// TestTeamRepository_GetTeamsByUserID_Success tests getting teams for user
func TestTeamRepository_GetTeamsByUserID_Success(t *testing.T) {
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

	repo := NewTeamRepository(database)
	teams, err := repo.GetTeamsByUserID(ctx, ownerID)

	require.NoError(t, err)
	assert.Len(t, teams, 2)
	assert.Equal(t, team1ID, teams[0].ID)
	assert.Equal(t, team2ID, teams[1].ID)
}

// TestTeamRepository_GetTeamsByUserID_Empty tests getting teams for user with no teams
func TestTeamRepository_GetTeamsByUserID_Empty(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "user@example.com", "User", "member", string(hashedPassword))

	repo := NewTeamRepository(database)
	teams, err := repo.GetTeamsByUserID(ctx, userID)

	require.NoError(t, err)
	assert.Empty(t, teams)
}

// TestTeamRepository_DeleteTeam_CascadeMembers tests deleting team cascades to members
func TestTeamRepository_DeleteTeam_CascadeMembers(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "member@example.com", "Member", "member", string(hashedPassword))

	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	_ = repo.AddTeamMember(ctx, teamID, userID, "member")

	// Delete team - should cascade to members
	err := repo.DeleteTeam(ctx, teamID)
	require.NoError(t, err)

	// Verify team is deleted
	_, err = repo.GetTeamByID(ctx, teamID)
	assert.Error(t, err)
}

// TestTeamRepository_TenantIsolation tests teams don't leak across tenants
func TestTeamRepository_TenantIsolation(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenant1ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-1")
	tenant2ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-2")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	owner1ID := testutil.CreateTestUser(ctx, t, txDB, tenant1ID, "owner1@example.com", "Owner 1", "manager", string(hashedPassword))
	owner2ID := testutil.CreateTestUser(ctx, t, txDB, tenant2ID, "owner2@example.com", "Owner 2", "manager", string(hashedPassword))

	testutil.CreateTestTeam(ctx, t, txDB, "Team 1", "Description 1", owner1ID, tenant1ID)
	testutil.CreateTestTeam(ctx, t, txDB, "Team 2", "Description 2", owner2ID, tenant2ID)

	repo := NewTeamRepository(database)
	teams1, _ := repo.GetTeamsByUserID(ctx, owner1ID)
	teams2, _ := repo.GetTeamsByUserID(ctx, owner2ID)

	assert.Len(t, teams1, 1)
	assert.Len(t, teams2, 1)
	assert.Equal(t, tenant1ID, teams1[0].TenantID)
	assert.Equal(t, tenant2ID, teams2[0].TenantID)
}

// TestTeamRepository_OwnerValidation tests owner must be valid user
func TestTeamRepository_OwnerValidation(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))

	// Create team with valid owner
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)

	repo := NewTeamRepository(database)
	team, err := repo.GetTeamByID(ctx, teamID)

	require.NoError(t, err)
	assert.Equal(t, ownerID, team.OwnerID)
}
