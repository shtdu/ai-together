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

	_ "modernc.org/sqlite"
)

// CreateSchema creates the necessary tables if they don't exist
func CreateSchema(db *sql.DB) error {
	queries := []string{
		// Sessions table with tool_name for isolation
		`CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT NOT NULL,
			tool_name TEXT NOT NULL,
			created_at INTEGER,
			PRIMARY KEY (session_id, tool_name)
		)`,
		// Events table with tool_name for isolation
		// id is auto-incremented and used as the sequence
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			tool_name TEXT NOT NULL,
			event_name TEXT,
			event_data TEXT,
			ts INTEGER,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		)`,
		// Indexes for common queries
		`CREATE INDEX IF NOT EXISTS idx_sessions_tool_created ON sessions(tool_name, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_events_session_tool ON events(session_id, tool_name)`,
		`CREATE INDEX IF NOT EXISTS idx_events_tool_ts ON events(tool_name, ts)`,
		`CREATE INDEX IF NOT EXISTS idx_events_id ON events(id)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
