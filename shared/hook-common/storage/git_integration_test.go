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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tidwall/gjson"
)

// setupTestGitRepo creates a temporary git repository for testing
func setupTestGitRepo(t *testing.T) string {
	tmpDir := t.TempDir()

	// Initialize git repo
	runCmd(t, tmpDir, "git", "init")
	runCmd(t, tmpDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, tmpDir, "git", "config", "user.name", "Test User")

	// Create a test file and commit
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	runCmd(t, tmpDir, "git", "add", "README.md")
	runCmd(t, tmpDir, "git", "commit", "-m", "Initial commit")

	// Add a remote
	runCmd(t, tmpDir, "git", "remote", "add", "origin", "git@github.com:test/repo.git")

	return tmpDir
}

// runCmd executes a command in the given directory
func runCmd(t *testing.T, dir string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command %s %v failed: %v\nOutput: %s", name, args, err, string(output))
	}
}

func TestStoreEventSQLite_FirstEventGetsGitMeta(t *testing.T) {
	// Setup: Create temp DB and temp git repo
	tmpDB := getTestDBPath(t)
	defer cleanupTestDB(t, tmpDB)

	tmpDir := setupTestGitRepo(t)

	db, err := OpenSQLite(tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer db.Close()

	sessionID := "test-session-git"
	toolName := "claude"
	eventName := "TestEvent"

	// Create event with CWD pointing to test git repo
	eventJSON := []byte(fmt.Sprintf(`{
        "session_id": "%s",
        "hook_event_name": "%s",
        "cwd": "%s",
        "data": "test"
    }`, sessionID, eventName, tmpDir))

	// Act: Store first event
	id1, err := StoreEventSQLite(db, sessionID, toolName, eventJSON, eventName)
	if err != nil {
		t.Fatalf("StoreEventSQLite() error = %v", err)
	}

	// Assert: Event was stored with git metadata
	var eventData string
	err = db.QueryRow(`
        SELECT event_data FROM events WHERE id = ?
    `, id1).Scan(&eventData)
	if err != nil {
		t.Fatalf("failed to query event: %v", err)
	}

	// Check git field exists
	result := gjson.Get(eventData, "git.remote_urls")
	if !result.Exists() {
		t.Error("git metadata not found in first event")
	}

	// Verify git metadata content
	branch := gjson.Get(eventData, "git.branch").String()
	if branch == "" {
		t.Error("git.branch is empty")
	}

	commit := gjson.Get(eventData, "git.commit_hash").String()
	if commit == "" {
		t.Error("git.commit_hash is empty")
	}

	// Verify remote URL
	urls := gjson.Get(eventData, "git.remote_urls").Array()
	if len(urls) == 0 {
		t.Error("git.remote_urls is empty")
	}

	// Act: Store second event for same session
	id2, err := StoreEventSQLite(db, sessionID, toolName, eventJSON, eventName)
	if err != nil {
		t.Fatalf("StoreEventSQLite() 2nd error = %v", err)
	}

	// Assert: Second event does NOT have git metadata
	err = db.QueryRow(`
        SELECT event_data FROM events WHERE id = ?
    `, id2).Scan(&eventData)
	if err != nil {
		t.Fatalf("failed to query 2nd event: %v", err)
	}

	result = gjson.Get(eventData, "git.remote_urls")
	if result.Exists() {
		t.Error("second event should not have git metadata")
	}
}

func TestStoreEventSQLite_NonGitRepo(t *testing.T) {
	// Test that events in non-git directories don't fail
	tmpDB := getTestDBPath(t)
	defer cleanupTestDB(t, tmpDB)

	tmpDir := t.TempDir() // Not a git repo

	db, err := OpenSQLite(tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer db.Close()

	eventJSON := []byte(fmt.Sprintf(`{
        "session_id": "non-git-session",
        "hook_event_name": "TestEvent",
        "cwd": "%s"
    }`, tmpDir))

	// Should not fail
	_, err = StoreEventSQLite(db, "non-git-session", "claude", eventJSON, "TestEvent")
	if err != nil {
		t.Errorf("StoreEventSQLite() should not fail for non-git repo, got error = %v", err)
	}

	// Verify event was stored without git metadata
	var eventData string
	err = db.QueryRow(`
        SELECT event_data FROM events WHERE session_id = ?
    `, "non-git-session").Scan(&eventData)
	if err != nil {
		t.Fatalf("failed to query event: %v", err)
	}

	result := gjson.Get(eventData, "git.remote_urls")
	if result.Exists() {
		t.Error("non-git event should not have git metadata")
	}
}

func TestStoreEventSQLite_NoCWD(t *testing.T) {
	// Test that events without CWD don't fail
	tmpDB := getTestDBPath(t)
	defer cleanupTestDB(t, tmpDB)

	db, err := OpenSQLite(tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer db.Close()

	eventJSON := []byte(`{
        "session_id": "no-cwd-session",
        "hook_event_name": "TestEvent"
    }`)

	// Should not fail
	_, err = StoreEventSQLite(db, "no-cwd-session", "claude", eventJSON, "TestEvent")
	if err != nil {
		t.Errorf("StoreEventSQLite() should not fail for event without CWD, got error = %v", err)
	}
}

func TestStoreEventSQLite_MultipleSessions(t *testing.T) {
	// Test that different sessions get their own git metadata
	tmpDB := getTestDBPath(t)
	defer cleanupTestDB(t, tmpDB)

	// Create two different git repos
	tmpDir1 := setupTestGitRepo(t)
	tmpDir2 := setupTestGitRepo(t)

	// Change branch in second repo
	runCmd(t, tmpDir2, "git", "checkout", "-b", "feature-branch")

	db, err := OpenSQLite(tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer db.Close()

	// Store event for session 1
	eventJSON1 := []byte(fmt.Sprintf(`{
        "session_id": "session-1",
        "hook_event_name": "TestEvent",
        "cwd": "%s"
    }`, tmpDir1))

	_, err = StoreEventSQLite(db, "session-1", "claude", eventJSON1, "TestEvent")
	if err != nil {
		t.Fatalf("StoreEventSQLite() session 1 error = %v", err)
	}

	// Store event for session 2
	eventJSON2 := []byte(fmt.Sprintf(`{
        "session_id": "session-2",
        "hook_event_name": "TestEvent",
        "cwd": "%s"
    }`, tmpDir2))

	_, err = StoreEventSQLite(db, "session-2", "claude", eventJSON2, "TestEvent")
	if err != nil {
		t.Fatalf("StoreEventSQLite() session 2 error = %v", err)
	}

	// Verify session 1 has main/master branch
	var eventData1 string
	err = db.QueryRow(`
        SELECT event_data FROM events WHERE session_id = ?
    `, "session-1").Scan(&eventData1)
	if err != nil {
		t.Fatalf("failed to query session 1: %v", err)
	}

	branch1 := gjson.Get(eventData1, "git.branch").String()

	// Verify session 2 has feature-branch
	var eventData2 string
	err = db.QueryRow(`
        SELECT event_data FROM events WHERE session_id = ?
    `, "session-2").Scan(&eventData2)
	if err != nil {
		t.Fatalf("failed to query session 2: %v", err)
	}

	branch2 := gjson.Get(eventData2, "git.branch").String()

	if branch2 != "feature-branch" {
		t.Errorf("session 2 should have branch 'feature-branch', got '%s'", branch2)
	}

	// Branches should be different (or at least session 2 should be feature-branch)
	if branch1 == branch2 && branch2 == "feature-branch" {
		t.Error("session 1 should not have feature-branch")
	}
}

func TestStoreEventSQLite_PreservesOriginalFields(t *testing.T) {
	// Test that merging git metadata preserves original event fields
	tmpDB := getTestDBPath(t)
	defer cleanupTestDB(t, tmpDB)

	tmpDir := setupTestGitRepo(t)

	db, err := OpenSQLite(tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer db.Close()

	sessionID := "preserve-test"
	toolName := "claude"

	// Create event with multiple fields
	originalEvent := map[string]interface{}{
		"session_id":      sessionID,
		"hook_event_name": "UserPromptSubmit",
		"cwd":             tmpDir,
		"prompt":          "hello world",
		"count":           42,
		"flag":            true,
	}
	eventJSON, _ := json.Marshal(originalEvent)

	_, err = StoreEventSQLite(db, sessionID, toolName, eventJSON, "UserPromptSubmit")
	if err != nil {
		t.Fatalf("StoreEventSQLite() error = %v", err)
	}

	// Verify all original fields are preserved
	var eventData string
	err = db.QueryRow(`
        SELECT event_data FROM events WHERE session_id = ?
    `, sessionID).Scan(&eventData)
	if err != nil {
		t.Fatalf("failed to query event: %v", err)
	}

	// Check original fields
	result := gjson.Parse(eventData)

	if !result.Get("session_id").Exists() {
		t.Error("session_id was lost")
	}
	if !result.Get("hook_event_name").Exists() {
		t.Error("hook_event_name was lost")
	}
	if !result.Get("cwd").Exists() {
		t.Error("cwd was lost")
	}
	if !result.Get("prompt").Exists() {
		t.Error("prompt was lost")
	}
	if result.Get("prompt").String() != "hello world" {
		t.Error("prompt value was changed")
	}
	if result.Get("count").Int() != 42 {
		t.Error("count value was changed")
	}
	if !result.Get("flag").Bool() {
		t.Error("flag value was changed")
	}

	// Check git field was added
	if !result.Get("git.remote_urls").Exists() {
		t.Error("git metadata was not added")
	}
}
