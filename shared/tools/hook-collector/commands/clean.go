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


package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/code-together/shared/hook-common/storage"
)

// CleanOldEvents removes events older than the specified time
func CleanOldEvents(db *sql.DB, toolName string, olderThan time.Time, dryRun bool) (int, int, error) {
	var eventsDeleted int
	var sessionsDeleted int

	// Track sessions to delete
	sessionsToDelete := make(map[string]bool)

	// Get all sessions
	sessions, err := storage.GetAllSessionIDs(db, toolName)
	if err != nil {
		return 0, 0, err
	}

	for _, sessionID := range sessions {
		// Get events for this session
		sessionEvents, err := storage.GetSessionEvents(db, sessionID, toolName)
		if err != nil {
			continue
		}

		// Find old events
		var toDelete []int
		var hasRecentEvent bool

		for _, event := range sessionEvents {
			// Parse event to check timestamp
			var eventData map[string]interface{}
			if err := json.Unmarshal(event.JSONData, &eventData); err != nil {
				continue // Skip malformed events
			}

			tsStr, ok := eventData["ts"].(string)
			if !ok {
				continue // Skip events without timestamp
			}

			ts, err := time.Parse(time.RFC3339Nano, tsStr)
			if err != nil {
				continue // Skip events with invalid timestamp
			}

			if ts.Before(olderThan) {
				toDelete = append(toDelete, event.ID)
			} else {
				hasRecentEvent = true
			}
		}

		if dryRun {
			if len(toDelete) > 0 {
				fmt.Printf("Would delete %d events from session %s\n", len(toDelete), sessionID)
				eventsDeleted += len(toDelete)
				if !hasRecentEvent {
					fmt.Printf("Would delete entire session %s (all events old)\n", sessionID)
					sessionsToDelete[sessionID] = true
				}
			}
			continue
		}

		// Delete old events
		for _, seq := range toDelete {
			_, err := storage.DeleteEventRange(db, sessionID, toolName, seq, seq)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to delete event %d: %v\n", seq, err)
				return eventsDeleted, sessionsDeleted, fmt.Errorf("failed to delete event %d: %w", seq, err)
			}
			eventsDeleted++
		}

		// Mark empty sessions for deletion
		if !hasRecentEvent {
			sessionsToDelete[sessionID] = true
		}
	}

	if dryRun {
		for range sessionsToDelete {
			sessionsDeleted++
		}
		return eventsDeleted, sessionsDeleted, nil
	}

	// Delete empty sessions
	for sessionID := range sessionsToDelete {
		if err := storage.DeleteSession(db, sessionID, toolName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to delete session %s: %v\n", sessionID, err)
			return eventsDeleted, sessionsDeleted, fmt.Errorf("failed to delete session %s: %w", sessionID, err)
		}
		sessionsDeleted++
	}

	return eventsDeleted, sessionsDeleted, nil
}

// CompactDatabase compacts the SQLite database using VACUUM
func CompactDatabase(dbPath string) error {
	db, err := storage.OpenDB(dbPath, false, "")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	return nil
}

// RunCleanCommand executes the clean command
func RunCleanCommand(dbPath, toolName, olderThan string, dryRun, compact bool, retentionDays int) {
	// Default to retention days if not specified
	if olderThan == "" {
		olderThan = fmt.Sprintf("%dd", retentionDays)
	}

	// Parse time filter
	cutoff, err := parseTimeFilter(olderThan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid older-than time '%s': %v\n", olderThan, err)
		os.Exit(1)
	}

	// Open database with specific tool
	db, err := storage.OpenDB(dbPath, false, toolName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Clean old events
	eventsDeleted, sessionsDeleted, err := CleanOldEvents(db, toolName, cutoff, dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to clean events: %v\n", err)
		os.Exit(1)
	}

	if dryRun {
		fmt.Printf("Dry run complete: %d events would be deleted, %d sessions would be deleted\n",
			eventsDeleted, sessionsDeleted)
	} else {
		fmt.Printf("Clean complete: %d events deleted, %d sessions deleted\n",
			eventsDeleted, sessionsDeleted)

		// Compact database if requested
		if compact && eventsDeleted > 0 {
			fmt.Println("Compacting database...")
			db.Close()
			if err := CompactDatabase(dbPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to compact database: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Compaction complete")
		}
	}
}
