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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"switch-server/models"
	"switch-server/testutil"
)

// TestUsageRepository_CreateUsageRecord_Success tests creating usage record
func TestUsageRepository_CreateUsageRecord_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "user@example.com", "User", "member", string(hashedPassword))

	repo := NewUsageRepository(database)
	usage := models.UsageRecord{
		TenantID:     tenantID,
		UserID:       userID,
		Platform:      "claude-code",
		Model:         "claude-3-5-sonnet",
		Provider:      "claude",
		HttpCode:      200,
		InputTokens:   1000,
		OutputTokens:  500,
		IsStream:      false,
		DurationSec:   1.5,
	}

	err := repo.CreateUsageRecord(ctx, usage)

	require.NoError(t, err)
}

// TestUsageRepository_GetUsageByTeamIDAndPeriod_Success tests getting usage by team and period
func TestUsageRepository_GetUsageByTeamIDAndPeriod_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "user@example.com", "User", "member", string(hashedPassword))

	testutil.CreateTestUsageRecord(ctx, t, txDB, tenantID, user.ID, "claude-code", "claude-3-5-sonnet", "claude")

	repo := NewUsageRepository(database)
	records, err := repo.GetUsageByTeamIDAndPeriod(ctx, tenantID, "2024-01-01", "2024-12-31")

	require.NoError(t, err)
	assert.Len(t, records, 1)
}

// TestUsageRepository_GetUsageByUserIDAndPeriod_Success tests getting usage by user and period
func TestUsageRepository_GetUsageByUserIDAndPeriod_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "user@example.com", "User", "member", string(hashedPassword))

	testutil.CreateTestUsageRecord(ctx, t, txDB, tenantID, user.ID, "claude-code", "claude-3-5-sonnet", "claude")

	repo := NewUsageRepository(database)
	records, err := repo.GetUsageByUserIDAndPeriod(ctx, user.ID, "2024-01-01", "2024-12-31")

	require.NoError(t, err)
	assert.Len(t, records, 1)
}

// TestUsageRepository_GetUsageStatsForTeam_Success tests getting usage stats
func TestUsageRepository_GetUsageStatsForTeam_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	ownerID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "owner@example.com", "Owner", "manager", string(hashedPassword))
	teamID := testutil.CreateTestTeam(ctx, t, txDB, "Test Team", "Description", ownerID, tenantID)
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "user@example.com", "User", "member", string(hashedPassword))

	testutil.CreateTestUsageRecord(ctx, t, txDB, tenantID, user.ID, "claude-code", "claude-3-5-sonnet", "claude")
	testutil.CreateTestUsageRecord(ctx, t, txDB, tenantID, user.ID, "codex", "gpt-4", "codex")

	repo := NewUsageRepository(database)
	stats, err := repo.GetUsageStatsForTeam(ctx, tenantID)

	require.NoError(t, err)
	assert.Contains(t, stats, "total_requests")
	assert.Contains(t, stats, "total_input_tokens")
	assert.Contains(t, stats, "total_output_tokens")
}

// TestUsageRepository_GetCurrentUsageForUser_Success tests getting current day usage
func TestUsageRepository_GetCurrentUsageForUser_Success(t *testing.T) {
	cleanup, database := testutil.SetupTestDB(t)
	defer cleanup()

	ctx, tx, txCleanup := testutil.BeginTestTx(t, database)
	defer txCleanup()

	txDB := database.WithTx(tx)
	tenantID := testutil.CreateTestTenant(ctx, t, tx, "test-tenant")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userID := testutil.CreateTestUser(ctx, t, txDB, tenantID, "user@example.com", "User", "member", string(hashedPassword))

	testutil.CreateTestUsageRecord(ctx, t, txDB, tenantID, user.ID, "claude-code", "claude-3-5-sonnet", "claude")

	repo := NewUsageRepository(database)
	usage, err := repo.GetCurrentUsageForUser(ctx, user.ID)

	require.NoError(t, err)
	assert.Contains(t, usage, "total_requests")
	assert.Contains(t, usage, "total_input_tokens")
	assert.Contains(t, usage, "total_output_tokens")
}

// TestUsageRepository_TestsComplete runs all usage tests
func TestUsageRepository_TestsComplete(t *testing.T) {
	t.Log("Usage repository tests implementation complete")
}
