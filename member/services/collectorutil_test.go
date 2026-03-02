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
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestValidateAndReadEvent_ValidEvent(t *testing.T) {
	event := `{"session_id":"test-001","hook_event_name":"UserPromptSubmit"}`
	body := io.NopCloser(strings.NewReader(event))

	data, err := ValidateAndReadEvent(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !json.Valid(data) {
		t.Error("expected valid JSON")
	}
}

func TestValidateAndReadEvent_EmptyEvent(t *testing.T) {
	body := io.NopCloser(strings.NewReader("   "))

	_, err := ValidateAndReadEvent(body)
	if err == nil {
		t.Fatal("expected error for empty event")
	}

	if !strings.Contains(err.Error(), "empty event data") {
		t.Errorf("expected 'empty event data' error, got %v", err)
	}
}

func TestValidateAndReadEvent_InvalidJSON(t *testing.T) {
	body := io.NopCloser(strings.NewReader("{invalid json}"))

	_, err := ValidateAndReadEvent(body)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' error, got %v", err)
	}
}

func TestValidateAndReadEvent_EventTooLarge(t *testing.T) {
	largeData := strings.Repeat("x", maxEventSize+1)
	body := io.NopCloser(strings.NewReader(largeData))

	_, err := ValidateAndReadEvent(body)
	if err == nil {
		t.Fatal("expected error for event too large")
	}

	if !strings.Contains(err.Error(), "event too large") {
		t.Errorf("expected 'event too large' error, got %v", err)
	}
}

func TestValidateSessionID_Present(t *testing.T) {
	event := []byte(`{"session_id":"test-001","hook_event_name":"UserPromptSubmit"}`)

	sessionID, err := ValidateSessionID(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sessionID != "test-001" {
		t.Errorf("expected session_id 'test-001', got '%s'", sessionID)
	}
}

func TestValidateSessionID_Missing(t *testing.T) {
	event := []byte(`{"hook_event_name":"UserPromptSubmit"}`)

	_, err := ValidateSessionID(event)
	if err == nil {
		t.Fatal("expected error for missing session_id")
	}

	if !strings.Contains(err.Error(), "missing session_id") {
		t.Errorf("expected 'missing session_id' error, got %v", err)
	}
}

func TestValidateSessionID_EmptyString(t *testing.T) {
	event := []byte(`{"session_id":"","hook_event_name":"UserPromptSubmit"}`)

	_, err := ValidateSessionID(event)
	if err == nil {
		t.Fatal("expected error for empty session_id")
	}

	if !strings.Contains(err.Error(), "missing session_id") {
		t.Errorf("expected 'missing session_id' error, got %v", err)
	}
}

func TestAddTimestamp(t *testing.T) {
	event := []byte(`{"session_id":"test-001","hook_event_name":"UserPromptSubmit"}`)

	withTS, err := AddTimestamp(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(withTS, &result); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if ts, ok := result["ts"].(string); !ok || ts == "" {
		t.Error("expected ts field to be added")
	}

	// Check original fields are preserved
	if result["session_id"] != "test-001" {
		t.Error("session_id was modified")
	}
	if result["hook_event_name"] != "UserPromptSubmit" {
		t.Error("hook_event_name was modified")
	}
}

func TestFormatErrorPayload_ValidJSON(t *testing.T) {
	payload := []byte(`{"session_id":"test-001","hook_event_name":"UserPromptSubmit"}`)

	formatted := FormatErrorPayload(payload)
	if !strings.Contains(formatted, "\n") {
		t.Error("expected formatted JSON to contain newlines")
	}

	if !strings.Contains(formatted, "  ") {
		t.Error("expected formatted JSON to contain indentation")
	}
}

func TestFormatErrorPayload_InvalidJSON(t *testing.T) {
	payload := []byte(`not json`)

	formatted := FormatErrorPayload(payload)
	if formatted != "not json" {
		t.Errorf("expected original string for invalid JSON, got '%s'", formatted)
	}
}
