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
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"codeswitch/internal/hookdb"
	"github.com/code-together/shared/hook-common/storage"
)

func setupHookServiceTest(t *testing.T) *HookService {
	t.Helper()

	// Create temp database
	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, "test-hook-events.db")

	// Set HOME to temp dir
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

func TestHookService_IsInitialized(t *testing.T) {
	hs := setupHookServiceTest(t)

	if !hs.IsInitialized() {
		t.Error("expected service to be initialized")
	}
}

func TestHookService_StoreEvent_ValidEvent(t *testing.T) {
	hs := setupHookServiceTest(t)

	event := map[string]interface{}{
		"session_id":      "test-session-001",
		"hook_event_name": "UserPromptSubmit",
		"cwd":             "/path/to/workspace",
		"data":            map[string]string{"key": "value"},
	}
	eventJSON, _ := json.Marshal(event)

	err := hs.StoreEvent("test-session-001", "claude", eventJSON)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHookService_StoreEvent_MissingSessionID(t *testing.T) {
	hs := setupHookServiceTest(t)

	event := map[string]interface{}{
		"hook_event_name": "UserPromptSubmit",
	}
	eventJSON, _ := json.Marshal(event)

	err := hs.StoreEvent("test-session-002", "claude", eventJSON)
	if err == nil {
		t.Fatal("expected error for missing session_id")
	}

	if err.Error() != "missing session_id in event data" {
		t.Errorf("expected 'missing session_id' error, got %v", err)
	}
}

func TestHookService_StoreEvent_MissingEventName(t *testing.T) {
	hs := setupHookServiceTest(t)

	event := map[string]interface{}{
		"session_id": "test-session-003",
	}
	eventJSON, _ := json.Marshal(event)

	err := hs.StoreEvent("test-session-003", "claude", eventJSON)
	if err == nil {
		t.Fatal("expected error for missing hook_event_name")
	}

	if err.Error() != "missing hook_event_name in event data" {
		t.Errorf("expected 'missing hook_event_name' error, got %v", err)
	}
}

func TestHookService_GetHealth(t *testing.T) {
	// Test with service created by NewHookService (may be initialized or not)
	hs := NewHookService()
	health := hs.GetHealth()

	// NewHookService calls Init() internally, so it should be initialized if no error
	if hs.IsInitialized() {
		if health["enabled"] != true {
			t.Error("expected enabled=true when database initialized")
		}
		if health["database"] == nil {
			t.Error("expected database path when initialized")
		}
	}

	// Test with explicitly initialized database
	hs = setupHookServiceTest(t)
	health = hs.GetHealth()

	if health["enabled"] != true {
		t.Error("expected enabled=true when database initialized")
	}

	if health["database"] == nil {
		t.Error("expected database path when initialized")
	}
}

// TestHookService_WriteQueue_ConcurrentWrites tests that concurrent writes succeed
func TestHookService_WriteQueue_ConcurrentWrites(t *testing.T) {
	hs := setupHookServiceTest(t)
	defer hs.Close()

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Send 50 events concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()

			event := map[string]interface{}{
				"session_id":      "test-concurrent",
				"hook_event_name": "TestEvent",
				"num":             num,
			}
			eventJSON, _ := json.Marshal(event)

			resultChan := make(chan error, 1)
			req := writeRequest{
				sessionID: "test-concurrent",
				toolName:  "test",
				eventJSON: eventJSON,
				result:    resultChan,
			}

			select {
			case hs.writeQueue <- req:
				if err := <-resultChan; err != nil {
					t.Errorf("write failed: %v", err)
					return
				}
				mu.Lock()
				successCount++
				mu.Unlock()
			default:
				t.Error("queue was full")
			}
		}(i)
	}

	wg.Wait()

	// All writes should succeed
	if successCount != 50 {
		t.Errorf("expected 50 successful writes, got %d", successCount)
	}

	// Verify all events in database
	events, err := storage.GetSessionEvents(hookdb.HookDB, "test-concurrent", "test")
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	if len(events) != 50 {
		t.Errorf("expected 50 events in database, got %d", len(events))
	}
}

// TestHookService_WriteQueue_GracefulShutdown tests that queue drains on shutdown
func TestHookService_WriteQueue_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize hook DB
	if err := hookdb.Init(); err != nil {
		t.Fatalf("failed to init hook db: %v", err)
	}
	defer hookdb.Close()

	hs := NewHookService()

	// Queue 10 events without waiting for results
	eventCount := 10
	for i := 0; i < eventCount; i++ {
		event := map[string]interface{}{
			"session_id":      "test-shutdown",
			"hook_event_name": "TestEvent",
			"num":             i,
		}
		eventJSON, _ := json.Marshal(event)

		resultChan := make(chan error, 1)
		req := writeRequest{
			sessionID: "test-shutdown",
			toolName:  "test",
			eventJSON: eventJSON,
			result:    resultChan,
		}

		select {
		case hs.writeQueue <- req:
			// Don't wait for result - testing shutdown while queue has items
		default:
			t.Fatal("failed to queue event")
		}
	}

	// Wait for queue to drain (writer goroutine processes events)
	time.Sleep(50 * time.Millisecond)

	// Verify all events were stored BEFORE closing
	// Note: hookdb.HookDB is still valid here
	events, err := storage.GetSessionEvents(hookdb.HookDB, "test-shutdown", "test")
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	if len(events) != eventCount {
		t.Errorf("expected %d events, got %d BEFORE shutdown", eventCount, len(events))
	}

	// Now close service (should drain any remaining queue)
	if err := hs.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Note: After Close(), hookdb.HookDB is nil, so we can't query anymore
	// But we've already verified the events were stored
}

// TestHookService_GetHealth_QueueMetrics tests health endpoint includes queue metrics
func TestHookService_GetHealth_QueueMetrics(t *testing.T) {
	hs := setupHookServiceTest(t)
	defer hs.Close()

	health := hs.GetHealth()

	// Check queue metrics exist
	if depth, ok := health["queue_depth"]; !ok {
		t.Error("health should include queue_depth")
	} else if depthInt, ok := depth.(int); !ok {
		t.Errorf("queue_depth should be int, got %T", depth)
	} else if depthInt < 0 || depthInt > 100 {
		t.Errorf("queue_depth out of range: %d", depthInt)
	}

	if capacity, ok := health["queue_capacity"]; !ok {
		t.Error("health should include queue_capacity")
	} else if capacity != 100 {
		t.Errorf("expected queue_capacity 100, got %v", capacity)
	}
}
