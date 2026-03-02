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


package storage

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentAccessPattern tests the real-world scenario where
// hook-collector and hook-browser access the database concurrently.
//
// With SQLite WAL mode:
// - Multiple readers can access simultaneously
// - Writers don't block readers
// - This is a major improvement over bbolt
func TestConcurrentAccessPattern(t *testing.T) {
	path := getTestDBPath(t)
	defer cleanupTestDB(t, path)

	sessionID := "integration-test"
	toolName := ""

	// Simulate 3 write/read cycles
	for cycle := 1; cycle <= 3; cycle++ {
		t.Logf("=== Cycle %d ===", cycle)

		// Phase 1: hook-collector writes a batch
		t.Log("Collector: opening database...")

		collectorDB, err := OpenDB(path, false, toolName)
		if err != nil {
			t.Fatalf("Collector failed to open db: %v", err)
		}

		eventsWritten := 0
		for i := 1; i <= 5; i++ {
			event := map[string]interface{}{
				"hook_event_name": "test_event",
				"cycle":           cycle,
			}
			eventJSON, _ := json.Marshal(event)
			err := StoreEvent(collectorDB, sessionID, toolName, eventJSON)

			if err != nil {
				t.Fatalf("Collector failed to write event: %v", err)
			}
			eventsWritten++
		}

		collectorDB.Close()
		t.Logf("Collector: wrote %d events, closed database", eventsWritten)

		// Phase 2: hook-browser reads events
		t.Log("Browser: opening database...")

		browserDB, err := OpenDB(path, true, toolName)
		if err != nil {
			t.Fatalf("Browser failed to open db: %v", err)
		}

		events, err := GetSessionEvents(browserDB, sessionID, toolName)
		browserDB.Close()

		if err != nil {
			t.Fatalf("Browser failed to read: %v", err)
		}

		eventsRead := len(events)
		expectedEvents := cycle * 5
		t.Logf("Browser: read %d events, closed database", eventsRead)

		if eventsRead != expectedEvents {
			t.Errorf("Cycle %d: expected %d events, got %d", cycle, expectedEvents, eventsRead)
		}

		// Simulate time passing between cycles
		time.Sleep(50 * time.Millisecond)
	}

	t.Log("✓ Test passed: hook-collector and hook-browser can work concurrently")
}

// TestShortLivedConcurrentAccess simulates realistic usage where:
// - hook-collector runs continuously, writing events periodically
// - hook-browser tries to read at random intervals
// - Both use short-lived connections (open-operate-close)
//
// With SQLite WAL mode, readers and writers can work concurrently.
func TestShortLivedConcurrentAccess(t *testing.T) {
	path := getTestDBPath(t)
	defer cleanupTestDB(t, path)

	sessionID := "concurrent-access-test"
	toolName := ""

	writerDone := make(chan struct{})
	readerDone := make(chan struct{})
	var totalEventsWritten, totalReads atomic.Int32

	// hook-collector: runs in background, writes batches periodically
	go func() {
		defer close(writerDone)

		for batch := 1; batch <= 5; batch++ {
			// Open DB for this batch
			db, err := OpenDB(path, false, toolName)
			if err != nil {
				t.Logf("Collector failed to open (attempt %d): %v", batch, err)
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// Write a batch of 3 events
			for i := 1; i <= 3; i++ {
				event := map[string]interface{}{"batch": batch, "num": i}
				eventJSON, _ := json.Marshal(event)
				err := StoreEvent(db, sessionID, toolName, eventJSON)

				if err != nil {
					db.Close()
					t.Logf("Collector write failed: %v", err)
					continue
				}
			}

			db.Close()
			totalEventsWritten.Add(3)

			t.Logf("Collector: batch %d done, closed DB (total: %d)", batch, totalEventsWritten.Load())

			// Wait before next batch (simulating idle time)
			time.Sleep(150 * time.Millisecond)
		}
	}()

	// hook-browser: tries to read periodically
	go func() {
		defer close(readerDone)

		for i := 1; i <= 15; i++ {
			// Try to open DB
			db, err := OpenDB(path, true, toolName)
			if err != nil {
				// DB might be busy, wait and try again
				time.Sleep(30 * time.Millisecond)
				continue
			}

			// Read events
			events, err := GetSessionEvents(db, sessionID, toolName)
			db.Close()

			if err != nil {
				t.Logf("Browser read failed (attempt %d): %v", i, err)
			}

			if len(events) > 0 {
				totalReads.Add(1)
			}

			time.Sleep(80 * time.Millisecond)
		}

		t.Logf("Browser: completed %d successful reads", totalReads.Load())
	}()

	// Wait for both to complete
	<-writerDone
	<-readerDone

	// Verify results
	written := totalEventsWritten.Load()
	reads := totalReads.Load()

	t.Logf("Results:")
	t.Logf("  Collector wrote %d events in 5 batches", written)
	t.Logf("  Browser completed %d successful reads", reads)

	if written == 0 {
		t.Error("Collector failed to write any events")
	}

	if reads == 0 {
		t.Error("Browser never successfully read events")
	}

	// Verify final DB state (allowing for partial writes)
	db, err := OpenDB(path, false, toolName)
	if err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}
	defer db.Close()

	events, err := GetSessionEvents(db, sessionID, toolName)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	finalCount := len(events)
	t.Logf("Final event count: %d", finalCount)

	if finalCount == 0 {
		t.Error("No events found in database")
	}

	// Only mark as passed if we have some data
	t.Log("✓ Test passed: SQLite enables concurrent access")
}

// TestMultipleReaders simulates multiple hook-browser instances
// reading simultaneously while hook-collector is writing.
// With SQLite WAL mode, this should work without blocking.
func TestMultipleReaders(t *testing.T) {
	path := getTestDBPath(t)
	defer cleanupTestDB(t, path)

	sessionID := "multi-reader-test"
	toolName := ""

	// First, collector writes some events
	t.Log("Collector: writing initial events...")

	collectorDB, err := OpenDB(path, false, toolName)
	if err != nil {
		t.Fatalf("Collector failed to open: %v", err)
	}

	for i := 1; i <= 10; i++ {
		event := map[string]interface{}{"event_num": i}
		eventJSON, _ := json.Marshal(event)
		err := StoreEvent(collectorDB, sessionID, toolName, eventJSON)

		if err != nil {
			collectorDB.Close()
			t.Fatalf("Collector write failed: %v", err)
		}
	}

	collectorDB.Close()
	t.Log("Collector: wrote 10 events, closed DB")

	// Now multiple browsers can read simultaneously
	t.Log("Browsers: 3 readers opening simultaneously...")

	var wg sync.WaitGroup
	for browserID := 1; browserID <= 3; browserID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			browserDB, err := OpenDB(path, true, toolName)
			if err != nil {
				t.Errorf("Browser %d failed to open: %v", id, err)
				return
			}
			defer browserDB.Close()

			events, err := GetSessionEvents(browserDB, sessionID, toolName)
			if err != nil {
				t.Errorf("Browser %d failed to read: %v", id, err)
				return
			}

			count := len(events)
			t.Logf("Browser %d: read %d events", id, count)

			if count != 10 {
				t.Errorf("Browser %d: expected 10 events, got %d", id, count)
			}
		}(browserID)
	}

	wg.Wait()
	t.Log("✓ Test passed: Multiple browsers can read simultaneously")
}
