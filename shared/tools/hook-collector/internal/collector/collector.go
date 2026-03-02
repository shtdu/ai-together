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


package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/code-together/shared/hook-common/storage"
	hookutils "github.com/code-together/shared/hook-utils"
	"github.com/code-together/shared/tools/hook-collector/internal/utils"
)

const maxEventSize = 10 * 1024 * 1024 // 10MB limit per event

// ReadEventFromStdin reads a JSON event from stdin
func ReadEventFromStdin() ([]byte, error) {
	// Check available size before reading (on platforms that support Stat)
	stat, err := os.Stdin.Stat()
	if err == nil && stat.Size() > maxEventSize {
		return nil, fmt.Errorf("event too large: %d bytes (max: %d bytes). "+
			"This may indicate the event includes large file contents. "+
			"Hooks should avoid embedding entire file contents in events. "+
			"Check your hook configuration to ensure it's not including full file dumps.",
			stat.Size(), maxEventSize)
	}

	// Read all input from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Validate size after read (in case Stat was unavailable or inaccurate)
	if len(data) > maxEventSize {
		return nil, fmt.Errorf("event too large: %d bytes (max: %d bytes). "+
			"This may indicate the event includes large file contents. "+
			"Consider using file references instead of embedding content in events.",
			len(data), maxEventSize)
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, fmt.Errorf("no input provided")
	}

	return []byte(trimmed), nil
}

// ExtractSessionID extracts the session_id from JSON event
// This performs minimal JSON parsing - only extracts session_id for lookup
func ExtractSessionID(jsonData []byte) (string, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	sessionID, ok := event["session_id"]
	if !ok {
		return "", fmt.Errorf("missing session_id in event")
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		return "", fmt.Errorf("session_id is not a string")
	}

	if sessionIDStr == "" {
		return "", fmt.Errorf("session_id is empty")
	}

	return sessionIDStr, nil
}

// AddTimestamp adds a ts field to the JSON event
func AddTimestamp(jsonData []byte, timestamp time.Time) ([]byte, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Add timestamp in ISO 8601 format
	event["ts"] = timestamp.UTC().Format(time.RFC3339Nano)

	return json.Marshal(event)
}

// CollectEvent collects an event from stdin and stores it in the database
func CollectEvent(dbPath string, toolName string, overwrite bool, verbose bool) error {
	// Read event from stdin
	eventJSON, err := ReadEventFromStdin()
	if err != nil {
		LogError(err, eventJSON)
		return err
	}

	// Extract session_id
	sessionID, err := ExtractSessionID(eventJSON)
	if err != nil {
		LogError(err, eventJSON)
		return err
	}

	// Extract cwd if present (for debugging/logging only)
	cwd := hookutils.ExtractCWD(eventJSON)

	// Add timestamp
	eventWithTS, err := AddTimestamp(eventJSON, time.Now())
	if err != nil {
		LogError(err, eventJSON)
		return err
	}

	// Check disk space before attempting write (fail fast)
	if err := utils.CheckDiskSpace(dbPath); err != nil {
		LogError(err, eventJSON)
		return err
	}

	if verbose {
		fmt.Printf("Opening database %s for tool %s\n", dbPath, toolName)
	}
	// Open database
	db, err := storage.OpenDB(dbPath, false, toolName)
	if err != nil {
		LogError(err, eventJSON)
		return err
	}
	defer db.Close()

	// Store event - id will be auto-incremented by SQLite
	if err := storage.StoreEvent(db, sessionID, toolName, eventWithTS); err != nil {
		LogError(err, eventJSON)
		return fmt.Errorf("failed to store event: %w", err)
	}

	if verbose {
		if cwd != "" {
			fmt.Printf("Stored event session=%s cwd=%s\n", sessionID, cwd)
		} else {
			fmt.Printf("Stored event session=%s\n", sessionID)
		}
	}

	return nil
}
