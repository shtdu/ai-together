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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestDB(t *testing.T) *HookService {
	t.Helper()

	// Create temp database
	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, "test-hook-events.db")

	// Initialize hook service with test path
	// Note: We need to manually set this up for testing
	// In production, HookService.Init() uses the home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	hs := NewHookService()
	if err := hs.Init(); err != nil {
		t.Fatalf("failed to init hook service: %v", err)
	}

	t.Cleanup(func() {
		hs.Close()
	})

	return hs
}

func TestCollectHandler_ValidEvent(t *testing.T) {
	hs := setupTestDB(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/collect/:tool_name", hs.CollectEvent)

	event := map[string]interface{}{
		"session_id":       "test-session-001",
		"hook_event_name":  "UserPromptSubmit",
		"cwd":              "/path/to/workspace",
		"data": map[string]string{"key": "value"},
	}
	eventJSON, _ := json.Marshal(event)

	req := httptest.NewRequest("POST", "/collect/claude", bytes.NewReader(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected success=true, got %v", resp["success"])
	}
}

func TestCollectHandler_MissingSessionID(t *testing.T) {
	hs := setupTestDB(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/collect/:tool_name", hs.CollectEvent)

	event := map[string]interface{}{
		"hook_event_name": "UserPromptSubmit",
	}
	eventJSON, _ := json.Marshal(event)

	req := httptest.NewRequest("POST", "/collect/claude", bytes.NewReader(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["error"] == nil {
		t.Error("expected error in response")
	}
}

func TestCollectHandler_MissingEventName(t *testing.T) {
	hs := setupTestDB(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/collect/:tool_name", hs.CollectEvent)

	event := map[string]interface{}{
		"session_id": "test-session-002",
	}
	eventJSON, _ := json.Marshal(event)

	req := httptest.NewRequest("POST", "/collect/claude", bytes.NewReader(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["error"] == nil {
		t.Error("expected error in response")
	}
}

func TestCollectHandler_InvalidJSON(t *testing.T) {
	hs := setupTestDB(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/collect/:tool_name", hs.CollectEvent)

	req := httptest.NewRequest("POST", "/collect/claude", bytes.NewReader([]byte("{invalid json}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCollectHandler_DatabaseNotInitialized(t *testing.T) {
	// Create a HookService without initializing the database
	hs := NewHookService()
	// Ensure it's not initialized
	hs.initialized = false

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/collect/:tool_name", hs.CollectEvent)

	event := map[string]interface{}{
		"session_id":       "test-session-003",
		"hook_event_name":  "UserPromptSubmit",
	}
	eventJSON, _ := json.Marshal(event)

	req := httptest.NewRequest("POST", "/collect/claude", bytes.NewReader(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["error"] == nil {
		t.Error("expected error in response")
	}
}
