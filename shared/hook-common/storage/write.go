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
)

// StoreEvent stores an event in the database (tool-specific)
func StoreEvent(db *sql.DB, sessionID string, toolName string, eventJSON []byte) error {
	eventName := extractEventName(eventJSON)
	_, err := StoreEventSQLite(db, sessionID, toolName, eventJSON, eventName)
	return err
}

// DeleteSession removes a session and all its events (tool-specific)
func DeleteSession(db *sql.DB, sessionID string, toolName string) error {
	return DeleteSessionSQLite(db, sessionID, toolName)
}

// DeleteEventRange removes events by id range (tool-specific)
func DeleteEventRange(db *sql.DB, sessionID string, toolName string, startSeq, endSeq int) (int, error) {
	// Convert int parameters to int64 for id ranges
	return DeleteEventRangeSQLite(db, sessionID, toolName, int64(startSeq), int64(endSeq))
}
