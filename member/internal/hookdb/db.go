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

package hookdb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/code-together/shared/hook-common/storage"
	_ "modernc.org/sqlite"
)

var (
	HookDB *sql.DB
	mu     sync.Mutex
)

// Init initializes the hook events database
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	dbDir := filepath.Join(home, ".code-together")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "hook-events.db")

	mu.Lock()
	defer mu.Unlock()

	if HookDB != nil {
		return nil // Already initialized
	}

	// Use shared storage to open DB (creates schema automatically)
	db, err := storage.OpenSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open hook events database: %w", err)
	}

	HookDB = db
	return nil
}

// Close closes the database connection
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if HookDB != nil {
		err := HookDB.Close()
		HookDB = nil // Reset to allow re-initialization
		return err
	}
	return nil
}
