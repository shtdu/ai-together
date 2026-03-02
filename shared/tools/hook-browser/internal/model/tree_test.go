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


package model

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/code-together/shared/hook-common/storage"
)

// TestSortEventsByTimestamp tests sorting events by timestamp
func TestSortEventsByTimestamp(t *testing.T) {
	// Create timestamps: 2024-01-25 12:00:00, 12:01:00, 12:02:00
	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        3,
			EventName: "Event3",
			JSONData:  createJSONData(ts2), // Middle timestamp
		},
		{
			ID:        1,
			EventName: "Event1",
			JSONData:  createJSONData(ts1), // Earliest timestamp
		},
		{
			ID:        2,
			EventName: "Event2",
			JSONData:  createJSONData(ts3), // Latest timestamp
		},
	}

	sorted := sortEventsByTimestamp(events)

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(sorted))
	}

	// Should be sorted by timestamp (earliest to latest): IDs 1, 3, 2
	if sorted[0].ID != 1 {
		t.Errorf("Expected first event ID to be 1 (earliest timestamp), got %d", sorted[0].ID)
	}
	if sorted[1].ID != 3 {
		t.Errorf("Expected second event ID to be 3 (middle timestamp), got %d", sorted[1].ID)
	}
	if sorted[2].ID != 2 {
		t.Errorf("Expected third event ID to be 2 (latest timestamp), got %d", sorted[2].ID)
	}
}

// TestSortEventsByTimestamp_Empty tests sorting empty event list
func TestSortEventsByTimestamp_Empty(t *testing.T) {
	events := []storage.Event{}
	sorted := sortEventsByTimestamp(events)

	if len(sorted) != 0 {
		t.Errorf("Expected empty slice, got %d events", len(sorted))
	}
}

// TestSortEventsByTimestamp_SingleEvent tests sorting single event
func TestSortEventsByTimestamp_SingleEvent(t *testing.T) {
	ts := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "Event1",
			JSONData:  createJSONData(ts),
		},
	}

	sorted := sortEventsByTimestamp(events)

	if len(sorted) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(sorted))
	}
	if sorted[0].ID != 1 {
		t.Errorf("Expected event ID to be 1, got %d", sorted[0].ID)
	}
}

// TestSortEventsByTimestamp_MissingTimestamp tests handling events without timestamp
func TestSortEventsByTimestamp_MissingTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "Event1",
			JSONData:  []byte(`{"other": "data"}`), // No timestamp
		},
		{
			ID:        2,
			EventName: "Event2",
			JSONData:  createJSONData(ts),
		},
	}

	sorted := sortEventsByTimestamp(events)

	// Events with timestamp 0 (missing) should come first
	if len(sorted) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(sorted))
	}
	if sorted[0].ID != 1 {
		t.Errorf("Expected first event ID to be 1 (missing timestamp), got %d", sorted[0].ID)
	}
	if sorted[1].ID != 2 {
		t.Errorf("Expected second event ID to be 2, got %d", sorted[1].ID)
	}
}

// TestBuildEventTree_SimpleSession tests building a simple session tree
func TestBuildEventTree_SimpleSession(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()
	ts4 := time.Date(2024, 1, 25, 12, 3, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "Stop",
			JSONData:  createJSONData(ts3),
		},
		{
			ID:        4,
			EventName: "SessionEnd",
			JSONData:  createJSONData(ts4),
		},
	}

	root := app.BuildEventTree(events)

	if root == nil {
		t.Fatal("Expected root node, got nil")
	}

	if len(root.Children) == 0 {
		t.Error("Expected root to have children, got none")
	}

	// Should have 2 children: Prompt branch and SessionEnd
	if len(root.Children) != 2 {
		t.Errorf("Expected 2 children under root, got %d", len(root.Children))
	}
}

// TestBuildEventTree_AgenticLoop tests building tree with agentic loop
func TestBuildEventTree_AgenticLoop(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 1, 10, 0, time.UTC).Unix()
	ts4 := time.Date(2024, 1, 25, 12, 1, 20, 0, time.UTC).Unix()
	ts5 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "PreToolUse",
			JSONData:  createJSONData(ts3),
		},
		{
			ID:        4,
			EventName: "PostToolUse",
			JSONData:  createJSONData(ts4),
		},
		{
			ID:        5,
			EventName: "Stop",
			JSONData:  createJSONData(ts5),
		},
	}

	root := app.BuildEventTree(events)

	if root == nil {
		t.Fatal("Expected root node, got nil")
	}

	// Should have 1 child: the prompt branch
	if len(root.Children) != 1 {
		t.Fatalf("Expected 1 child under root, got %d", len(root.Children))
	}

	promptNode := root.Children[0]
	if promptNode.DisplayText != "Prompt #1" {
		t.Errorf("Expected prompt branch 'Prompt #1', got '%s'", promptNode.DisplayText)
	}

	// Prompt should have 2 children: Agentic Loop and Stop
	if len(promptNode.Children) != 2 {
		t.Fatalf("Expected 2 children under prompt, got %d", len(promptNode.Children))
	}

	// First child should be Agentic Loop
	agenticLoop := promptNode.Children[0]
	if agenticLoop.DisplayText != "Agentic Loop" {
		t.Errorf("Expected 'Agentic Loop', got '%s'", agenticLoop.DisplayText)
	}

	// Agentic Loop should have 2 children: PreToolUse and PostToolUse
	if len(agenticLoop.Children) != 2 {
		t.Fatalf("Expected 2 children under Agentic Loop, got %d", len(agenticLoop.Children))
	}
}

// TestBuildEventTree_MultiplePrompts tests building tree with multiple prompts
func TestBuildEventTree_MultiplePrompts(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()
	ts4 := time.Date(2024, 1, 25, 12, 3, 0, 0, time.UTC).Unix()
	ts5 := time.Date(2024, 1, 25, 12, 4, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "Stop",
			JSONData:  createJSONData(ts3),
		},
		{
			ID:        4,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts4),
		},
		{
			ID:        5,
			EventName: "Stop",
			JSONData:  createJSONData(ts5),
		},
	}

	root := app.BuildEventTree(events)

	if root == nil {
		t.Fatal("Expected root node, got nil")
	}

	// Should have 2 prompt branches
	if len(root.Children) != 2 {
		t.Fatalf("Expected 2 children under root, got %d", len(root.Children))
	}

	// Verify both are prompt branches
	for i, child := range root.Children {
		expected := fmt.Sprintf("Prompt #%d", i+1)
		if child.DisplayText != expected {
			t.Errorf("Expected '%s', got '%s'", expected, child.DisplayText)
		}
	}
}

// TestBuildEventTree_OutOfOrder tests that events are sorted before building tree
func TestBuildEventTree_OutOfOrder(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 1, 10, 0, time.UTC).Unix()
	ts4 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	// Events delivered out of chronological order
	events := []storage.Event{
		{
			ID:        5,
			EventName: "Stop",
			JSONData:  createJSONData(ts4), // Latest
		},
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1), // Earliest
		},
		{
			ID:        3,
			EventName: "PreToolUse",
			JSONData:  createJSONData(ts3), // Middle
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2), // Earlier middle
		},
	}

	root := app.BuildEventTree(events)

	if root == nil {
		t.Fatal("Expected root node, got nil")
	}

	// Should still build correct tree despite out-of-order input
	if len(root.Children) == 0 {
		t.Error("Expected root to have children, got none")
	}

	// First child should be the prompt branch
	promptNode := root.Children[0]
	if promptNode.DisplayText != "Prompt #1" {
		t.Errorf("Expected 'Prompt #1', got '%s'", promptNode.DisplayText)
	}
}

// TestBuildEventTree_EmptyEvents tests building tree with no events
func TestBuildEventTree_EmptyEvents(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	events := []storage.Event{}
	root := app.BuildEventTree(events)

	if root != nil {
		t.Error("Expected nil root for empty events, got a node")
	}
}

// TestBuildEventTree_WithNotification tests building tree with async notification
func TestBuildEventTree_WithNotification(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "Notification",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "SessionEnd",
			JSONData:  createJSONData(ts3),
		},
	}

	root := app.BuildEventTree(events)

	if root == nil {
		t.Fatal("Expected root node, got nil")
	}

	// Should have 2 children: Notification and SessionEnd
	if len(root.Children) != 2 {
		t.Fatalf("Expected 2 children under root, got %d", len(root.Children))
	}

	// Find the notification node
	foundNotification := false
	for _, child := range root.Children {
		if child.DisplayText == "Notification (async)" {
			foundNotification = true
			break
		}
	}

	if !foundNotification {
		t.Error("Expected to find 'Notification (async)' node")
	}
}

// TestFlattenTree tests flattening tree back to event list
func TestFlattenTree(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "Stop",
			JSONData:  createJSONData(ts3),
		},
	}

	root := app.BuildEventTree(events)
	flattened := app.FlattenTree(root)

	// Should have same number of events
	if len(flattened) != len(events) {
		t.Errorf("Expected %d events, got %d", len(events), len(flattened))
	}

	// All event IDs should be present
	eventIDs := make(map[int]bool)
	for _, event := range flattened {
		eventIDs[event.ID] = true
	}

	for _, original := range events {
		if !eventIDs[original.ID] {
			t.Errorf("Expected to find event ID %d in flattened events", original.ID)
		}
	}
}

// TestGetTreeLeafCount tests counting leaf nodes (events) in tree
func TestGetTreeLeafCount(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "Stop",
			JSONData:  createJSONData(ts3),
		},
	}

	root := app.BuildEventTree(events)
	leafCount := app.GetTreeLeafCount(root)

	// All 3 events are leaf nodes (have Event != nil)
	if leafCount != 3 {
		t.Errorf("Expected 3 leaf nodes, got %d", leafCount)
	}
}

// TestExtractTimestamp tests extracting timestamp from event JSON
func TestExtractTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()

	event := storage.Event{
		ID:        1,
		EventName: "TestEvent",
		JSONData:  createJSONData(ts),
	}

	timestamp := extractTimestamp(event)
	if timestamp != ts {
		t.Errorf("Expected timestamp %d, got %d", ts, timestamp)
	}
}

// TestExtractTimestamp_Missing tests extracting from event without timestamp
func TestExtractTimestamp_Missing(t *testing.T) {
	event := storage.Event{
		ID:        1,
		EventName: "TestEvent",
		JSONData:  []byte(`{"other": "data"}`),
	}

	timestamp := extractTimestamp(event)
	if timestamp != 0 {
		t.Errorf("Expected timestamp 0 for missing field, got %d", timestamp)
	}
}

// TestBuildEventTree_SessionIsolation tests that tree only contains events from that session
func TestBuildEventTree_SessionIsolation(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	// Session 1 events
	session1Events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "Stop",
			JSONData:  createJSONData(ts3),
		},
	}

	// Build tree for session 1
	root1 := app.BuildEventTree(session1Events)
	leafCount1 := app.GetTreeLeafCount(root1)

	if leafCount1 != 3 {
		t.Errorf("Expected 3 events in session 1 tree, got %d", leafCount1)
	}

	// Session 2 events (different IDs, same structure)
	ts4 := time.Date(2024, 1, 26, 14, 0, 0, 0, time.UTC).Unix()
	ts5 := time.Date(2024, 1, 26, 14, 1, 0, 0, time.UTC).Unix()
	ts6 := time.Date(2024, 1, 26, 14, 2, 0, 0, time.UTC).Unix()

	session2Events := []storage.Event{
		{
			ID:        4,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts4),
		},
		{
			ID:        5,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts5),
		},
		{
			ID:        6,
			EventName: "Stop",
			JSONData:  createJSONData(ts6),
		},
	}

	// Build tree for session 2
	root2 := app.BuildEventTree(session2Events)
	leafCount2 := app.GetTreeLeafCount(root2)

	if leafCount2 != 3 {
		t.Errorf("Expected 3 events in session 2 tree, got %d", leafCount2)
	}

	// Verify the trees are independent (different event IDs)
	flattened1 := app.FlattenTree(root1)
	flattened2 := app.FlattenTree(root2)

	// Session 1 should only have IDs 1, 2, 3
	for _, event := range flattened1 {
		if event.ID < 1 || event.ID > 3 {
			t.Errorf("Session 1 tree has unexpected event ID: %d", event.ID)
		}
	}

	// Session 2 should only have IDs 4, 5, 6
	for _, event := range flattened2 {
		if event.ID < 4 || event.ID > 6 {
			t.Errorf("Session 2 tree has unexpected event ID: %d", event.ID)
		}
	}
}

// TestBuildTreeViewWidget_ToolNameDisplay tests that tool events show tool_name
func TestBuildTreeViewWidget_ToolNameDisplay(t *testing.T) {
	app := &App{
		Styles: DefaultStyles(),
	}

	ts1 := time.Date(2024, 1, 25, 12, 0, 0, 0, time.UTC).Unix()
	ts2 := time.Date(2024, 1, 25, 12, 1, 0, 0, time.UTC).Unix()
	ts3 := time.Date(2024, 1, 25, 12, 1, 10, 0, time.UTC).Unix()
	ts4 := time.Date(2024, 1, 25, 12, 1, 20, 0, time.UTC).Unix()
	ts5 := time.Date(2024, 1, 25, 12, 2, 0, 0, time.UTC).Unix()

	events := []storage.Event{
		{
			ID:        1,
			EventName: "SessionStart",
			JSONData:  createJSONData(ts1),
		},
		{
			ID:        2,
			EventName: "UserPromptSubmit",
			JSONData:  createJSONData(ts2),
		},
		{
			ID:        3,
			EventName: "PreToolUse",
			JSONData:  createJSONDataWithTool(ts3, "Task"),
		},
		{
			ID:        4,
			EventName: "PostToolUse",
			JSONData:  createJSONDataWithTool(ts4, "ReadFile"),
		},
		{
			ID:        5,
			EventName: "Stop",
			JSONData:  createJSONData(ts5),
		},
	}

	root := app.BuildEventTree(events)
	treeView := app.BuildTreeViewWidget(root)

	if treeView == nil {
		t.Fatal("Expected tree view, got nil")
	}

	// Verify the tree was built
	if root == nil {
		t.Fatal("Expected root node, got nil")
	}

	// Verify tool events are in the tree with correct tool names
	foundPreToolUse := false
	foundPostToolUse := false
	preToolUseText := ""
	postToolUseText := ""

	for _, child := range root.Children {
		if child.DisplayText == "Prompt #1" {
			for _, grandchild := range child.Children {
				if grandchild.DisplayText == "Agentic Loop" {
					for _, toolEvent := range grandchild.Children {
						if toolEvent.Event != nil {
							if toolEvent.Event.EventName == "PreToolUse" {
								foundPreToolUse = true
								preToolUseText = toolEvent.DisplayText
							}
							if toolEvent.Event.EventName == "PostToolUse" {
								foundPostToolUse = true
								postToolUseText = toolEvent.DisplayText
							}
						}
					}
				}
			}
		}
	}

	if !foundPreToolUse {
		t.Error("Expected to find PreToolUse in tree")
	}
	if !foundPostToolUse {
		t.Error("Expected to find PostToolUse in tree")
	}

	// Verify tool names are in DisplayText
	if preToolUseText != "PreToolUse <Task>" {
		t.Errorf("Expected PreToolUse DisplayText to be 'PreToolUse <Task>', got '%s'", preToolUseText)
	}
	if postToolUseText != "PostToolUse <ReadFile>" {
		t.Errorf("Expected PostToolUse DisplayText to be 'PostToolUse <ReadFile>', got '%s'", postToolUseText)
	}
}

// TestExtractToolName tests extracting tool_name from event JSON
func TestExtractToolName(t *testing.T) {
	// Test with actual JSON structure from the user's example
	jsonData := []byte(`{
		"cwd": "/Users/jian/workspaces/github/ct_member",
		"hook_event_name": "PreToolUse",
		"permission_mode": "plan",
		"session_id": "39f9b856-7f8d-47cb-9215-fb7de384c229",
		"tool_name": "Task",
		"tool_use_id": "call_bfae9e5b4eb14017bd81b302",
		"transcript_path": "/Users/jian/.claude/projects/-Users-jian-workspaces-github-ct-member/39f9b856-7f8d-47cb-9215-fb7de384c229.jsonl",
		"ts": "2026-01-30T08:08:28.922288Z"
	}`)

	event := storage.Event{
		ID:        1,
		EventName: "PreToolUse",
		JSONData:  jsonData,
	}

	toolName := extractToolName(event)
	if toolName != "Task" {
		t.Errorf("Expected tool_name 'Task', got '%s'", toolName)
	}
}

// TestExtractToolName_Missing tests extracting when tool_name is absent
func TestExtractToolName_Missing(t *testing.T) {
	jsonData := []byte(`{
		"cwd": "/Users/jian/workspaces/github/ct_member",
		"hook_event_name": "SessionStart",
		"session_id": "test-session",
		"ts": "2026-01-30T08:08:28.922288Z"
	}`)

	event := storage.Event{
		ID:        1,
		EventName: "SessionStart",
		JSONData:  jsonData,
	}

	toolName := extractToolName(event)
	if toolName != "" {
		t.Errorf("Expected empty tool_name, got '%s'", toolName)
	}
}

// Helper function to create JSON data with ISO 8601 timestamp
func createJSONData(ts int64) []byte {
	data := map[string]interface{}{
		"ts":   time.Unix(ts, 0).Format(time.RFC3339),
		"test": "data",
	}
	jsonBytes, _ := json.Marshal(data)
	return jsonBytes
}

// Helper function to create JSON data with tool_name
func createJSONDataWithTool(ts int64, toolName string) []byte {
	data := map[string]interface{}{
		"ts":        time.Unix(ts, 0).Format(time.RFC3339),
		"tool_name": toolName,
	}
	jsonBytes, _ := json.Marshal(data)
	return jsonBytes
}
