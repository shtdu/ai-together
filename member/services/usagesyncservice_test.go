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

package services

import (
	"os"
	"testing"
	"time"

	"codeswitch/internal/db"
)

func setupUsageSyncTest(t *testing.T) {
	t.Helper()

	// Create temp directory for test database
	tmpDir := t.TempDir()

	// Set HOME to temp dir to avoid using real database
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	// Initialize database in temp directory
	if err := db.Init(); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})
}

func TestUsageSyncService_DeleteOldRecords(t *testing.T) {
	setupUsageSyncTest(t)

	// Create a usage sync service
	us := &UsageSyncService{}

	// Clean up any existing test data (now using temp DB)
	_, _ = db.DB.Exec("DELETE FROM request_log")

	// Create test records with different ages
	now := time.Now()
	oldDate := now.Add(-10 * 24 * time.Hour)   // 10 days ago
	recentDate := now.Add(-3 * 24 * time.Hour) // 3 days ago

	testRecords := []struct {
		age          time.Time
		synced       bool
		shouldDelete bool
	}{
		{oldDate, true, true},      // Old and synced - should be deleted
		{oldDate, false, false},    // Old but not synced - should NOT be deleted
		{recentDate, true, false},  // Recent but synced - should NOT be deleted
		{recentDate, false, false}, // Recent and not synced - should NOT be deleted
		{now, true, false},         // Current and synced - should NOT be deleted
	}

	for i, tc := range testRecords {
		syncedValue := 0
		if tc.synced {
			syncedValue = 1
		}

		result, err := db.DB.Exec(
			`INSERT INTO request_log (platform, model, provider, http_code, created_at, synced_to_server)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			"test", "test-model", "test-provider", 200,
			tc.age.Format(timeLayout), syncedValue,
		)
		if err != nil {
			t.Fatalf("failed to insert test record %d: %v", i, err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("failed to get last insert id for record %d: %v", i, err)
		}
		_ = id // We don't need the ID for this test
	}

	// Verify all records were inserted
	var countBefore int
	err := db.DB.Get(&countBefore, "SELECT COUNT(*) FROM request_log")
	if err != nil {
		t.Fatalf("failed to count records before cleanup: %v", err)
	}
	if countBefore != len(testRecords) {
		t.Errorf("expected %d records before cleanup, got %d", len(testRecords), countBefore)
	}

	// Run cleanup
	deletedCount, err := us.DeleteOldRecords()
	if err != nil {
		t.Fatalf("DeleteOldRecords failed: %v", err)
	}

	// Verify only the correct records were deleted
	var countAfter int
	err = db.DB.Get(&countAfter, "SELECT COUNT(*) FROM request_log")
	if err != nil {
		t.Fatalf("failed to count records after cleanup: %v", err)
	}

	expectedDeleted := 1 // Only the old and synced record
	expectedRemaining := len(testRecords) - expectedDeleted

	if deletedCount != expectedDeleted {
		t.Errorf("expected to delete %d records, got %d", expectedDeleted, deletedCount)
	}

	if countAfter != expectedRemaining {
		t.Errorf("expected %d records after cleanup, got %d", expectedRemaining, countAfter)
	}

	// Verify specific records
	var checkCount int

	// Old and synced should be deleted
	err = db.DB.Get(&checkCount, "SELECT COUNT(*) FROM request_log WHERE created_at < ? AND synced_to_server = 1",
		oldDate.Add(24*time.Hour).Format(timeLayout))
	if err != nil {
		t.Fatalf("failed to check old synced records: %v", err)
	}
	if checkCount != 0 {
		t.Errorf("expected 0 old synced records, got %d", checkCount)
	}

	// Old but not synced should still exist
	err = db.DB.Get(&checkCount, "SELECT COUNT(*) FROM request_log WHERE created_at < ? AND synced_to_server = 0",
		oldDate.Add(24*time.Hour).Format(timeLayout))
	if err != nil {
		t.Fatalf("failed to check old unsynced records: %v", err)
	}
	if checkCount != 1 {
		t.Errorf("expected 1 old unsynced record, got %d", checkCount)
	}

	// Clean up test data
	_, _ = db.DB.Exec("DELETE FROM request_log")
}

func TestUsageSyncService_DeleteOldRecords_EmptyTable(t *testing.T) {
	setupUsageSyncTest(t)

	// Create a usage sync service
	us := &UsageSyncService{}

	// Clean up any existing data (now using temp DB)
	_, _ = db.DB.Exec("DELETE FROM request_log")

	// Run cleanup on empty table
	deletedCount, err := us.DeleteOldRecords()
	if err != nil {
		t.Fatalf("DeleteOldRecords failed on empty table: %v", err)
	}

	if deletedCount != 0 {
		t.Errorf("expected to delete 0 records from empty table, got %d", deletedCount)
	}
}

func TestUsageSyncService_DeleteOldRecords_OnlyRecentRecords(t *testing.T) {
	setupUsageSyncTest(t)

	// Create a usage sync service
	us := &UsageSyncService{}

	// Clean up any existing test data (now using temp DB)
	_, _ = db.DB.Exec("DELETE FROM request_log")

	// Create only recent synced records (within 7 days)
	recentDate := time.Now().Add(-3 * 24 * time.Hour)

	_, err := db.DB.Exec(
		`INSERT INTO request_log (platform, model, provider, http_code, created_at, synced_to_server)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"codex", "test-model", "test-provider", 200,
		recentDate.Format(timeLayout), 1,
	)
	if err != nil {
		t.Fatalf("failed to insert test record: %v", err)
	}

	// Run cleanup
	deletedCount, err := us.DeleteOldRecords()
	if err != nil {
		t.Fatalf("DeleteOldRecords failed: %v", err)
	}

	// Verify no records were deleted
	if deletedCount != 0 {
		t.Errorf("expected to delete 0 records when all are recent, got %d", deletedCount)
	}

	var count int
	err = db.DB.Get(&count, "SELECT COUNT(*) FROM request_log")
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record to remain, got %d", count)
	}

	// Clean up test data
	_, _ = db.DB.Exec("DELETE FROM request_log")
}
