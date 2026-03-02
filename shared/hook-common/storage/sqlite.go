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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	hookutils "github.com/code-together/shared/hook-utils"
	_ "modernc.org/sqlite"
)

// OpenSQLite opens a SQLite database at the given path
func OpenSQLite(path string) (*sql.DB, error) {
	// Ensure directory exists
	if err := createDirForFile(path); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Open database with WAL mode for better concurrency
	// Using modernc.org/sqlite (pure Go, no CGO)
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_timeout=5000&_synchronous=NORMAL", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for better concurrency
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create schema
	if err := CreateSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createDirForFile(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

// StoreEventSQLite stores an event in SQLite with tool isolation
// Returns the auto-incremented id
func StoreEventSQLite(db *sql.DB, sessionID string, toolName string, eventJSON []byte, eventName string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Extract timestamp from JSON if present
	var eventData map[string]interface{}
	var ts int64
	if err := json.Unmarshal(eventJSON, &eventData); err == nil {
		if tsVal, ok := eventData["ts"].(string); ok {
			// Parse ISO 8601 timestamp
			if t, err := time.Parse(time.RFC3339Nano, tsVal); err == nil {
				ts = t.Unix()
			}
		} else if tsVal, ok := eventData["ts"].(float64); ok {
			ts = int64(tsVal)
		}
	}
	if ts == 0 {
		ts = time.Now().Unix()
	}

	// Check if this is the first event for this session
	// We query before inserting to know if we need to add git metadata
	var eventCount int
	err = tx.QueryRow(`
        SELECT COUNT(*) FROM events
        WHERE session_id = ? AND tool_name = ?
    `, sessionID, toolName).Scan(&eventCount)

	isFirstEvent := (err == nil && eventCount == 0)

	if isFirstEvent {
		// This is the first event - extract CWD and get git metadata
		cwd := hookutils.ExtractCWD(eventJSON)

		if cwd != "" {
			gitMeta, err := hookutils.GetGitMetaFast(cwd)
			if err != nil {
				// Check if this is the expected "not a git repo" error
				if errors.Is(err, hookutils.ErrNotGitRepo) {
					// Not in a git repo - this is expected, skip silently
					slog.Debug("not in a git repository - skipping git metadata extraction",
						"session", sessionID,
						"tool", toolName,
						"cwd", cwd)
				} else {
					// Unexpected error - log warning but don't fail
					// We may have partial metadata in gitMeta
					slog.Warn("failed to get some git metadata for first event",
						"session", sessionID,
						"tool", toolName,
						"cwd", cwd,
						"error", err)
				}
			}

			// Merge metadata if we got any (even partial)
			if gitMeta != nil {
				// Merge git metadata into event JSON
				eventJSON, err = mergeGitMeta(eventJSON, gitMeta)
				if err != nil {
					slog.Error("failed to merge git metadata into event",
						"session", sessionID,
						"error", err)
					// Continue with original eventJSON
				} else {
					slog.Info("merged git metadata into first event",
						"session", sessionID,
						"tool", toolName,
						"branch", gitMeta.Branch,
						"commit", gitMeta.CommitHash)
				}
			}
		}
	}

	// Insert event - id will be auto-incremented
	result, err := tx.Exec(`
		INSERT INTO events (session_id, tool_name, event_name, event_data, ts)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, toolName, eventName, string(eventJSON), ts)

	if err != nil {
		return 0, fmt.Errorf("failed to insert event: %w", err)
	}

	// Get the auto-incremented id
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Update session (insert if not exists)
	_, err = tx.Exec(`
		INSERT INTO sessions (session_id, tool_name, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(session_id, tool_name) DO NOTHING
	`, sessionID, toolName, ts)

	if err != nil {
		return 0, fmt.Errorf("failed to update session: %w", err)
	}

	return id, tx.Commit()
}

// GetSessionInfoSQLite retrieves all sessions with event counts (tool-specific)
func GetSessionInfoSQLite(db *sql.DB, toolName string) ([]SessionInfo, error) {
	rows, err := db.Query(`
		SELECT s.session_id, COUNT(e.id) as event_count
		FROM sessions s
		LEFT JOIN events e ON s.session_id = e.session_id AND s.tool_name = e.tool_name
		WHERE s.tool_name = ?
		GROUP BY s.session_id
		ORDER BY s.created_at DESC
	`, toolName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionInfo
	for rows.Next() {
		var s SessionInfo
		if err := rows.Scan(&s.ID, &s.Count); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetSessionEventsSQLite retrieves all events for a session (tool-specific)
func GetSessionEventsSQLite(db *sql.DB, sessionID string, toolName string) ([]Event, error) {
	rows, err := db.Query(`
		SELECT id, event_name, event_data FROM events
		WHERE session_id = ? AND tool_name = ?
		ORDER BY id ASC
	`, sessionID, toolName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.EventName, &e.JSONData); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetFirstEventTimeSQLite gets the timestamp of the first event in a session (tool-specific)
func GetFirstEventTimeSQLite(db *sql.DB, sessionID string, toolName string) (int64, error) {
	var ts int64
	err := db.QueryRow(`
		SELECT ts FROM events
		WHERE session_id = ? AND tool_name = ?
		ORDER BY id ASC LIMIT 1
	`, sessionID, toolName).Scan(&ts)
	return ts, err
}

// DeleteSessionSQLite removes a session and all its events (tool-specific)
func DeleteSessionSQLite(db *sql.DB, sessionID string, toolName string) error {
	_, err := db.Exec(`
		DELETE FROM events WHERE session_id = ? AND tool_name = ?
	`, sessionID, toolName)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		DELETE FROM sessions WHERE session_id = ? AND tool_name = ?
	`, sessionID, toolName)
	return err
}

// DeleteEventRangeSQLite removes events by id range (tool-specific)
func DeleteEventRangeSQLite(db *sql.DB, sessionID string, toolName string, startID, endID int64) (int, error) {
	result, err := db.Exec(`
		DELETE FROM events
		WHERE session_id = ? AND tool_name = ? AND id BETWEEN ? AND ?
	`, sessionID, toolName, startID, endID)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()

	// If session is now empty, remove the session record
	var remainingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM events WHERE session_id = ? AND tool_name = ?
	`, sessionID, toolName).Scan(&remainingCount)
	if err == nil && remainingCount == 0 {
		DeleteSessionSQLite(db, sessionID, toolName)
	}

	return int(count), nil
}

// RawEvent represents a raw event row from the database.
type RawEvent struct {
	ID        int64  // Auto-incremented event ID
	SessionID string // Session identifier
	EventName string // Event type name
	JSONData  []byte // Event data as JSON
	Timestamp int64  // Unix timestamp
}

// GetAllToolNamesSQLite retrieves all distinct tool_name values from the events table.
func GetAllToolNamesSQLite(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
		SELECT DISTINCT tool_name FROM events ORDER BY tool_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var toolNames []string
	for rows.Next() {
		var toolName string
		if err := rows.Scan(&toolName); err != nil {
			return nil, err
		}
		toolNames = append(toolNames, toolName)
	}
	return toolNames, rows.Err()
}

// GetEventsForToolOnDaySQLite retrieves events for a specific tool within a timestamp range.
// Filters at SQL level to avoid loading all events into memory.
// dayStartUnix and dayEndUnix are Unix timestamps defining the day boundaries (inclusive start, exclusive end).
func GetEventsForToolOnDaySQLite(db *sql.DB, toolName string, dayStartUnix, dayEndUnix int64) ([]RawEvent, error) {
	rows, err := db.Query(`
		SELECT id, session_id, event_name, event_data, ts
		FROM events
		WHERE tool_name = ? AND ts >= ? AND ts < ?
		ORDER BY ts ASC
	`, toolName, dayStartUnix, dayEndUnix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []RawEvent
	for rows.Next() {
		var e RawEvent
		if err := rows.Scan(&e.ID, &e.SessionID, &e.EventName, &e.JSONData, &e.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetTimeRangeSQLite retrieves the earliest and latest timestamps for a specific tool.
func GetTimeRangeSQLite(db *sql.DB, toolName string) (earliestTs, latestTs int64, err error) {
	err = db.QueryRow(`
		SELECT MIN(ts) as earliest_ts, MAX(ts) as latest_ts
		FROM events
		WHERE tool_name = ?
	`, toolName).Scan(&earliestTs, &latestTs)
	return earliestTs, latestTs, err
}
