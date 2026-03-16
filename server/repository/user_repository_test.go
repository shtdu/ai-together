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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"switch-server/testutil"
)

// TestUserRepository_CreateUser_Success tests successful user creation
func TestUserRepository_CreateUser_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	// Hash password before creating user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo := NewUserRepository(database)
	user, err := repo.CreateUser(ctx, "test@example.com", "Test User", string(hashedPassword), "member", tenantID)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "member", user.Role)
	assert.Equal(t, tenantID, user.TenantID)
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
}

// TestUserRepository_CreateUser_DuplicateEmail tests that duplicate emails are rejected
func TestUserRepository_CreateUser_DuplicateEmail(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Create first user
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "test@example.com", "Test User", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	// Try to create duplicate user
	_, err = repo.CreateUser(ctx, "test@example.com", "Another User", string(hashedPassword), "member", tenantID)

	assert.Error(t, err)
}

// TestUserRepository_GetUserByEmail_Success tests finding existing user by email
func TestUserRepository_GetUserByEmail_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	testutil.CreateTestUser(ctx, t, txDB, tenantID, "test@example.com", "Test User", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	user, err := repo.GetUserByEmail(ctx, "test@example.com")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
}

// TestUserRepository_GetUserByEmail_NotFound tests error for non-existent email
func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewUserRepository(database)
	user, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
}

// TestUserRepository_GetUserByID_Success tests retrieving user by ID
func TestUserRepository_GetUserByID_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "test@example.com", "Test User", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	user, err := repo.GetUserByID(ctx, userID)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
}

// TestUserRepository_GetUserByID_NotFound tests error for invalid ID
func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewUserRepository(database)
	user, err := repo.GetUserByID(ctx, 99999)

	assert.Error(t, err)
	assert.Nil(t, user)
}

// TestUserRepository_UpdateUser_Success tests updating name and role
func TestUserRepository_UpdateUser_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "test@example.com", "Test User", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	err = repo.UpdateUser(ctx, userID, "Updated User", "manager")

	require.NoError(t, err)

	// Verify update
	user, err := repo.GetUserByID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "Updated User", user.Name)
	assert.Equal(t, "manager", user.Role)
}

// TestUserRepository_UpdateUser_NotFound tests update fails for non-existent user
func TestUserRepository_UpdateUser_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewUserRepository(database)
	err := repo.UpdateUser(ctx, 99999, "New Name", "manager")

	assert.Error(t, err)
}

// TestUserRepository_UpdateUserWithPassword_Success tests updating user with password change
func TestUserRepository_UpdateUserWithPassword_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte("newpassword123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "test@example.com", "Test User", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	err = repo.UpdateUserWithPassword(ctx, userID, "Updated User", "manager", string(newHashedPassword))

	require.NoError(t, err)

	// Verify update
	user, err := repo.GetUserByID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "Updated User", user.Name)
	assert.Equal(t, "manager", user.Role)
	// Password should be updated
	assert.NotEqual(t, string(hashedPassword), user.Password)
}

// TestUserRepository_DeleteUser_Success tests deleting user and verifying removal
func TestUserRepository_DeleteUser_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "test@example.com", "Test User", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	err = repo.DeleteUser(ctx, userID)

	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetUserByID(ctx, userID)
	assert.Error(t, err)
}

// TestUserRepository_DeleteUser_NotFound tests error deleting non-existent user
func TestUserRepository_DeleteUser_NotFound(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	repo := NewUserRepository(database)
	err := repo.DeleteUser(ctx, 99999)

	// Should not error even if user doesn't exist (DELETE is idempotent)
	// Based on PostgreSQL behavior, this will succeed with 0 rows affected
	// The repository doesn't check rows affected, so no error
	assert.NoError(t, err)
}

// TestUserRepository_ListUsersByTenant_Success tests returning all users for tenant
func TestUserRepository_ListUsersByTenant_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Create multiple users
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "user1@example.com", "User 1", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "user2@example.com", "User 2", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "user3@example.com", "User 3", "manager", string(hashedPassword))

	repo := NewUserRepository(database)
	users, err := repo.ListUsersByTenant(ctx, tenantID)

	require.NoError(t, err)
	assert.Len(t, users, 3)
}

// TestUserRepository_ListUsersByTenant_Empty tests returning empty array
func TestUserRepository_ListUsersByTenant_Empty(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	repo := NewUserRepository(database)
	users, err := repo.ListUsersByTenant(ctx, tenantID)

	require.NoError(t, err)
	assert.Empty(t, users)
}

// TestUserRepository_ListUsersByTenant_Multiple tests returning sorted results (DESC)
func TestUserRepository_ListUsersByTenant_Multiple(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Create users in order
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "first@example.com", "First User", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "second@example.com", "Second User", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenantID, "third@example.com", "Third User", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	users, err := repo.ListUsersByTenant(ctx, tenantID)

	require.NoError(t, err)
	assert.Len(t, users, 3)
	// Results should be in DESC order by created_at, so third is first
	assert.Equal(t, "third@example.com", users[0].Email)
	assert.Equal(t, "second@example.com", users[1].Email)
	assert.Equal(t, "first@example.com", users[2].Email)
}

// TestUserRepository_TenantIsolation tests users don't leak across tenants
func TestUserRepository_TenantIsolation(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenant1ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-1")
	tenant2ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-2")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Create users in different tenants
	testutil.CreateTestUser(ctx, t, txDB, tenant1ID, "user1@example.com", "User 1", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenant2ID, "user2@example.com", "User 2", "member", string(hashedPassword))

	repo := NewUserRepository(database)
	users1, err := repo.ListUsersByTenant(ctx, tenant1ID)
	require.NoError(t, err)

	users2, err := repo.ListUsersByTenant(ctx, tenant2ID)
	require.NoError(t, err)

	assert.Len(t, users1, 1)
	assert.Len(t, users2, 1)
	assert.Equal(t, "user1@example.com", users1[0].Email)
	assert.Equal(t, "user2@example.com", users2[0].Email)
}

// TestUserRepository_EmailUniqueness tests email uniqueness across tenants
func TestUserRepository_EmailUniqueness(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenant1ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-1")
	tenant2ID := testutil.CreateTestTenant(ctx, t, tx, "tenant-2")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Same email in different tenants should work
	testutil.CreateTestUser(ctx, t, txDB, tenant1ID, "same@example.com", "User 1", "member", string(hashedPassword))
	testutil.CreateTestUser(ctx, t, txDB, tenant2ID, "same@example.com", "User 2", "member", string(hashedPassword))

	// Verify both users exist with same email but different tenants
	repo := NewUserRepository(database)
	users1, _ := repo.ListUsersByTenant(ctx, tenant1ID)
	users2, _ := repo.ListUsersByTenant(ctx, tenant2ID)

	assert.Len(t, users1, 1)
	assert.Len(t, users2, 1)
	assert.Equal(t, "same@example.com", users1[0].Email)
	assert.Equal(t, "same@example.com", users2[0].Email)
	// IDs should be different
	assert.NotEqual(t, users1[0].ID, users2[0].ID)
}
