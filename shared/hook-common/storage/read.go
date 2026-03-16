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
	"database/sql"
	"encoding/json"
)

// GetAllSessionIDs returns all session IDs in the database (tool-specific)
func GetAllSessionIDs(db *sql.DB, toolName string) ([]string, error) {
	rows, err := db.Query(`
		SELECT DISTINCT session_id FROM sessions WHERE tool_name = ?
		ORDER BY created_at DESC
	`, toolName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessionIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		sessionIDs = append(sessionIDs, id)
	}
	return sessionIDs, rows.Err()
}

// GetSessionInfo returns information about all sessions including event counts (tool-specific)
func GetSessionInfo(db *sql.DB, toolName string) ([]SessionInfo, error) {
	return GetSessionInfoSQLite(db, toolName)
}

// GetSessionEvents retrieves all events for a given session (tool-specific)
func GetSessionEvents(db *sql.DB, sessionID string, toolName string) ([]Event, error) {
	return GetSessionEventsSQLite(db, sessionID, toolName)
}

// extractEventName extracts the hook_event_name field from JSON data
func extractEventName(jsonData []byte) string {
	var event map[string]interface{}
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return "Unknown"
	}

	name, ok := event["hook_event_name"].(string)
	if !ok {
		return "Unknown"
	}

	return name
}
