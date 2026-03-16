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
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"switch-server/internal/db"
)

// SetupTestDB creates a test database connection and returns cleanup function
// Usage: cleanup, database := SetupTestDB(t); defer cleanup()
func SetupTestDB(t *testing.T) (cleanup func(), database *db.DB) {
	t.Helper()

	// Get test database URL from environment or use default
	testDSN := os.Getenv("TEST_DATABASE_URL")
	if testDSN == "" {
		testDSN = "postgres://postgres:postgres@localhost:5432/codetogether_test?sslmode=disable"
	}

	// Create connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	cleanup = func() {
		pool.Close()
	}

	// Create DB using the same pattern as ConnectDB
	queries := db.New(pool)
	database = &db.DB{}
	database.Queries = queries

	return cleanup, database
}

// SetupTestDBWithPool creates a test database with a custom pool
// Usage: cleanup, database := SetupTestDBWithPool(t, pool); defer cleanup()
func SetupTestDBWithPool(t *testing.T, pool *pgxpool.Pool) (cleanup func(), database *db.DB) {
	t.Helper()

	queries := db.New(pool)
	database = &db.DB{
		Queries: queries,
	}

	cleanup = func() {
		pool.Close()
	}

	return cleanup, database
}

// BeginTestTx begins a test transaction and returns context, transaction, and cleanup function
// The cleanup function will rollback the transaction automatically
// Usage: ctx, tx, txCleanup := BeginTestTx(t, db); defer txCleanup()
func BeginTestTx(t *testing.T, database *db.DB) (context.Context, pgx.Tx, func()) {
	t.Helper()

	ctx := context.Background()
	tx, err := database.Pool().Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin test transaction: %v", err)
	}

	cleanup := func() {
		err := tx.Rollback(ctx)
		if err != nil {
			t.Logf("Warning: Failed to rollback test transaction: %v", err)
		}
	}

	return ctx, tx, cleanup
}

// TruncateTable truncates a table and resets serial sequence
// Useful for cleanup between tests when not using transactions
func TruncateTable(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tableName string) {
	t.Helper()

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
	_, err := pool.Exec(ctx, query)
	if err != nil {
		t.Fatalf("Failed to truncate table %s: %v", tableName, err)
	}
}

// TruncateAllTables truncates all main tables in order
// Useful for cleanup between tests when not using transactions
func TruncateAllTables(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	tables := []string{
		"request_log",
		"team_usage_summary",
		"team_members",
		"providers",
		"provider_templates",
		"teams",
		"users",
		"tenants",
	}

	for _, table := range tables {
		TruncateTable(t, ctx, pool, table)
	}
}
