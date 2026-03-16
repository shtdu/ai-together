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

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const maxEventSize = 10 * 1024 * 1024 // 10MB limit per event

// ValidateAndReadEvent reads and validates event from request body
func ValidateAndReadEvent(body io.ReadCloser) ([]byte, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	if len(data) > maxEventSize {
		return nil, fmt.Errorf("event too large: %d bytes (max: %d bytes)",
			len(data), maxEventSize)
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, fmt.Errorf("empty event data")
	}

	if !json.Valid([]byte(trimmed)) {
		return nil, fmt.Errorf("invalid JSON format")
	}

	return []byte(trimmed), nil
}

// ValidateSessionID checks if session_id exists in event data
// (Session ID is now extracted from body, not validated against URL)
func ValidateSessionID(eventJSON []byte) (string, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(eventJSON, &event); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	sessionID, ok := event["session_id"].(string)
	if !ok || sessionID == "" {
		return "", fmt.Errorf("missing session_id in event data")
	}

	return sessionID, nil
}

// AddTimestamp adds ts field to event JSON
func AddTimestamp(eventJSON []byte) ([]byte, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(eventJSON, &event); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	event["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	return json.Marshal(event)
}

// FormatErrorPayload pretty-prints JSON for logging
func FormatErrorPayload(payload []byte) string {
	if !json.Valid(payload) {
		return string(payload)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, payload, "", "  "); err == nil {
		return pretty.String()
	}

	return string(payload)
}
