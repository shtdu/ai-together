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


package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// getTestDBPath returns a unique path for a test database
func getTestDBPath(t *testing.T) string {
	tmpDir := filepath.Join(os.TempDir(), "hook-common-test")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	return filepath.Join(tmpDir, filepath.Base(t.Name())+".db")
}

// cleanupTestDB removes test database and associated files
func cleanupTestDB(t *testing.T, path string) {
	basePath := path[:len(path)-len(filepath.Ext(path))]

	patterns := []string{
		path,
		basePath + "-wal",
		basePath + "-shm",
	}

	for _, p := range patterns {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove %s: %v", p, err)
		}
	}
}

func TestOpenDB_ReadWrite(t *testing.T) {
	path := getTestDBPath(t)
	defer cleanupTestDB(t, path)

	db, err := OpenDB(path, false, "")
	if err != nil {
		t.Fatalf("OpenDB(read-write) failed: %v", err)
	}
	defer db.Close()

	// Verify tables were created by checking if we can query them
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Sessions table was not created: %v", err)
	}
}

func TestOpenDB_ReadOnly(t *testing.T) {
	path := getTestDBPath(t)
	defer cleanupTestDB(t, path)

	// First create a database
	dbRW, err := OpenDB(path, false, "")
	if err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	dbRW.Close()

	// Open in read-only mode (SQLite doesn't enforce read-only in same way as bbolt)
	db, err := OpenDB(path, true, "")
	if err != nil {
		t.Fatalf("OpenDB(read-only) failed: %v", err)
	}
	defer db.Close()

	// Verify we can read
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Failed to verify table: %v", err)
	}
}

func TestOpenDB_ReadOnly_CreateIfNeeded(t *testing.T) {
	path := getTestDBPath(t)
	defer cleanupTestDB(t, path)

	// Try to open non-existent database in read-only mode
	// SQLite will create it anyway
	db, err := OpenDB(path, true, "")
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()
}

func TestOpenDB_NonExistent_ReadWrite(t *testing.T) {
	path := getTestDBPath(t)
	defer cleanupTestDB(t, path)

	// Open non-existent database in read-write mode should create it
	db, err := OpenDB(path, false, "")
	if err != nil {
		t.Fatalf("OpenDB failed to create new database: %v", err)
	}
	defer db.Close()

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}
