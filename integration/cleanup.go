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

package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
)

// setupTestDB creates the test database if it doesn't exist and returns the connection URL.
// It reads TEST_DATABASE_URL from environment (default: postgres://postgres:postgres@localhost:5432/codetogether_test?sslmode=disable).
func setupTestDB(logger *slog.Logger) string {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://postgres:postgres@localhost:5432/codetogether_test?sslmode=disable"
	}

	// Parse connection string to get database name
	connConfig, err := pgx.ParseConfig(testDBURL)
	if err != nil {
		logger.Error("Failed to parse database URL", "error", err)
		panic(fmt.Sprintf("Failed to parse database URL: %v", err))
	}

	// Connect to postgres database to create test database
	adminDBURL := fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
		connConfig.User, connConfig.Host, connConfig.Port, "postgres")

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, adminDBURL)
	if err != nil {
		logger.Error("Failed to connect to postgres database", "error", err)
		panic(fmt.Sprintf("Failed to connect to postgres: %v (ensure PostgreSQL is running)", err))
	}
	defer conn.Close(ctx)

	// Check if database exists
	var dbExists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT datname FROM pg_database WHERE datname = $1)", "codetogether_test").Scan(&dbExists)
	if err != nil {
		logger.Error("Failed to check if database exists", "error", err)
		panic(fmt.Sprintf("Failed to check database: %v", err))
	}

	// Create database if it doesn't exist
	if !dbExists {
		_, err = conn.Exec(ctx, "CREATE DATABASE codetogether_test")
		if err != nil {
			logger.Error("Failed to create test database", "error", err)
			panic(fmt.Sprintf("Failed to create database: %v", err))
		}
		logger.Info("Created test database", "database", "codetogether_test")

		// Run migrations to create schema
		// Note: The server should handle this on startup, but for a fresh database
		// we need to ensure the schema exists
		logger.Info("Note: Please ensure the server has run migrations on the test database")
	}

	// Connect to test database and ensure at least one tenant exists
	testConn, err := pgx.Connect(ctx, testDBURL)
	if err != nil {
		logger.Error("Failed to connect to test database", "error", err)
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}
	defer testConn.Close(ctx)

	// Check if tenants table has any rows
	var tenantCount int
	err = testConn.QueryRow(ctx, "SELECT COUNT(*) FROM tenants").Scan(&tenantCount)
	if err == nil && tenantCount == 0 {
		// No tenants exist, need to create one for registration to work
		// This is a workaround since we can't use the setup endpoint from the API client
		logger.Warn("No tenants found. Tests will fail. Please run initial server setup via /api/v1/setup/admin endpoint")
	}

	return testDBURL
}

// cleanupDatabase truncates all tables except one tenant to preserve setup.
// This should only be used for cleanup between tests, NEVER for fixture creation.
func cleanupDatabase(dbURL string, logger *slog.Logger) error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	// Check if tenants exist
	var tenantCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM tenants").Scan(&tenantCount)
	if err != nil {
		logger.Debug("No tenants table yet or error checking", "error", err)
		// Table might not exist, which is fine for first run
		return nil
	}

	// Truncate tables in correct order (respecting FKs)
	// Only truncate tenants if there are multiple (preserve at least one for setup)
	tables := []string{
		"request_log",
		"team_usage_summary",
		"team_members",
		"teams",
		"providers",
		"users",
		"licenses",
	}

	for _, table := range tables {
		_, err := conn.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Table might not exist yet, which is fine
			logger.Debug("Failed to truncate table (might not exist yet)", "table", table, "error", err)
		}
	}

	// Only truncate tenants if there are multiple (preserve initial setup)
	if tenantCount > 1 {
		_, err = conn.Exec(ctx, "TRUNCATE TABLE tenants CASCADE")
		if err != nil {
			logger.Debug("Failed to truncate tenants", "error", err)
		}
	}

	return nil
}
