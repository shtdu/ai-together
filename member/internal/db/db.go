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


package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var (
	DB        *sqlx.DB
	roDB      *sqlx.DB // Read-only connection for report queries
	mu        sync.Mutex
	roDBOnce  sync.Once
)

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	const sqliteOptions = "?cache=shared&mode=rwc&_busy_timeout=5000&_journal_mode=WAL"

	dbDir := filepath.Join(home, ".code-together")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	dsn := filepath.Join(dbDir, "app.db"+sqliteOptions)

	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if DB != nil {
		db.Close()
		return nil
	}

	if err := ensureRequestLogTable(db); err != nil {
		db.Close()
		return err
	}

	DB = db
	return nil
}

func ensureRequestLogTable(db *sqlx.DB) error {
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return fmt.Errorf("failed to set busy timeout: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to set journal mode: %w", err)
	}

	const createTableSQL = `CREATE TABLE IF NOT EXISTS request_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		platform TEXT,
		model TEXT,
		provider TEXT,
		http_code INTEGER,
		input_tokens INTEGER,
		output_tokens INTEGER,
		cache_create_tokens INTEGER,
		cache_read_tokens INTEGER,
		reasoning_tokens INTEGER,
		is_stream INTEGER DEFAULT 0,
		duration_sec REAL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create request_log table: %w", err)
	}

	// Ensure columns exist (for backward compatibility)
	if err := ensureRequestLogColumn(db, "created_at", "DATETIME DEFAULT CURRENT_TIMESTAMP"); err != nil {
		return fmt.Errorf("failed to ensure created_at column: %w", err)
	}
	if err := ensureRequestLogColumn(db, "is_stream", "INTEGER DEFAULT 0"); err != nil {
		return fmt.Errorf("failed to ensure is_stream column: %w", err)
	}
	if err := ensureRequestLogColumn(db, "duration_sec", "REAL DEFAULT 0"); err != nil {
		return fmt.Errorf("failed to ensure duration_sec column: %w", err)
	}
	if err := ensureRequestLogColumn(db, "synced_to_server", "INTEGER DEFAULT 0"); err != nil {
		return fmt.Errorf("failed to ensure synced_to_server column: %w", err)
	}

	return nil
}

func ensureRequestLogColumn(db *sqlx.DB, column string, definition string) error {
	// Note: Column names are trusted constants, not user input, so string formatting is safe here.
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('request_log') WHERE name = '%s'", column)
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return fmt.Errorf("failed to check column %s: %w", column, err)
	}
	if count == 0 {
		alter := fmt.Sprintf("ALTER TABLE request_log ADD COLUMN %s %s", column, definition)
		if _, err := db.Exec(alter); err != nil {
			return fmt.Errorf("failed to add column %s: %w", column, err)
		}
	}
	return nil
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if DB != nil {
		err := DB.Close()
		DB = nil // Reset to allow re-initialization
		return err
	}
	return nil
}

// GetReadOnlyDB returns a read-only database connection for report queries.
// This prevents read operations from being blocked by write locks.
// The connection is created once and reused for subsequent calls.
func GetReadOnlyDB() (*sqlx.DB, error) {
	var initErr error
	roDBOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			initErr = fmt.Errorf("failed to get home directory: %w", err)
			return
		}

		// Read-only mode: mode=ro, enable shared cache and WAL
		const sqliteOptions = "?cache=shared&mode=ro&_busy_timeout=5000"
		dbPath := filepath.Join(home, ".code-together", "app.db")
		dsn := dbPath + sqliteOptions

		db, err := sqlx.Open("sqlite", dsn)
		if err != nil {
			initErr = fmt.Errorf("failed to open read-only database: %w", err)
			return
		}

		if err := db.Ping(); err != nil {
			db.Close()
			initErr = fmt.Errorf("failed to ping read-only database: %w", err)
			return
		}

		roDB = db
	})

	if initErr != nil {
		return nil, initErr
	}

	if roDB == nil {
		return nil, fmt.Errorf("read-only database not initialized")
	}

	return roDB, nil
}
