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
	"sort"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/code-together/shared/hook-common/storage"
)

// eventData represents the structure within JSONData
type eventData struct {
	Timestamp string `json:"ts,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
}

// TreeNode represents a node in the event hierarchy tree
type TreeNode struct {
	Event       *storage.Event
	Children    []*TreeNode
	Expanded    bool
	Depth       int
	DisplayText string
}

// eventWithTimestamp wraps an event with its extracted timestamp for sorting
type eventWithTimestamp struct {
	event     storage.Event
	timestamp int64
}

// sortEventsByTimestamp sorts events by their timestamp field from JSON data
func sortEventsByTimestamp(events []storage.Event) []storage.Event {
	// Wrap events with their timestamps
	wrapped := make([]eventWithTimestamp, len(events))
	for i, event := range events {
		wrapped[i] = eventWithTimestamp{
			event:     event,
			timestamp: extractTimestamp(event),
		}
	}

	// Sort by timestamp (earliest first)
	sort.Slice(wrapped, func(i, j int) bool {
		return wrapped[i].timestamp < wrapped[j].timestamp
	})

	// Unwrap back to events
	sorted := make([]storage.Event, len(events))
	for i, w := range wrapped {
		sorted[i] = w.event
	}

	return sorted
}

// BuildEventTree constructs a hierarchical tree from flat events
func (a *App) BuildEventTree(events []storage.Event) *TreeNode {
	if len(events) == 0 {
		return nil
	}

	// Sort events by timestamp from JSON data (HTTP handlers may not preserve order)
	sortedEvents := sortEventsByTimestamp(events)

	// Create root node
	root := &TreeNode{
		Event:       nil,
		Children:    []*TreeNode{},
		Expanded:    true,
		Depth:       0,
		DisplayText: "Session",
	}

	// Track state for building hierarchy
	var currentPromptNode *TreeNode
	var agenticLoopNode *TreeNode

	for i, event := range sortedEvents {
		eventName := event.EventName

		switch eventName {
		case "SessionStart":
			// SessionStart is already represented by root
			timestamp := extractTimestamp(event)
			root.DisplayText = fmt.Sprintf("SessionStart (%s)", formatTime(timestamp))
			root.Event = &event

		case "UserPromptSubmit":
			// Create new prompt branch
			promptNum := countPrompts(events[:i])
			promptNode := &TreeNode{
				Event:       &event,
				Children:    []*TreeNode{},
				Expanded:    true,
				Depth:       1,
				DisplayText: fmt.Sprintf("Prompt #%d", promptNum+1),
			}
			root.Children = append(root.Children, promptNode)
			currentPromptNode = promptNode

			// Reset agentic loop for new prompt
			agenticLoopNode = nil

		case "PreToolUse", "PermissionRequest", "PostToolUse", "PostToolUseFailure", "SubagentStart", "SubagentStop":
			// Tool events go under agentic loop
			if currentPromptNode == nil {
				// Shouldn't happen, but handle gracefully
				continue
			}

			// Extract tool_name for display
			toolName := extractToolName(event)
			displayName := eventName
			if toolName != "" && toolName != "ERROR:" {
				displayName = fmt.Sprintf("%s <%s>", eventName, toolName)
			}

			// Create agentic loop node if it doesn't exist
			if agenticLoopNode == nil {
				agenticLoopNode = &TreeNode{
					Event:       nil,
					Children:    []*TreeNode{},
					Expanded:    true,
					Depth:       currentPromptNode.Depth + 1,
					DisplayText: "Agentic Loop",
				}
				currentPromptNode.Children = append(currentPromptNode.Children, agenticLoopNode)
			}

			// Create event node
			eventNode := &TreeNode{
				Event:       &event,
				Children:    []*TreeNode{},
				Expanded:    false,
				Depth:       agenticLoopNode.Depth + 1,
				DisplayText: displayName,
			}
			agenticLoopNode.Children = append(agenticLoopNode.Children, eventNode)

		case "Stop":
			// Stop ends the current prompt branch
			if currentPromptNode != nil {
				eventNode := &TreeNode{
					Event:       &event,
					Children:    []*TreeNode{},
					Expanded:    false,
					Depth:       currentPromptNode.Depth + 1,
					DisplayText: "Stop",
				}
				currentPromptNode.Children = append(currentPromptNode.Children, eventNode)
			}
			// Reset for next prompt
			agenticLoopNode = nil

		case "PreCompact", "SessionEnd":
			// Terminal events go under root
			eventNode := &TreeNode{
				Event:       &event,
				Children:    []*TreeNode{},
				Expanded:    false,
				Depth:       1,
				DisplayText: eventName,
			}
			root.Children = append(root.Children, eventNode)

		case "Notification":
			// Async events - link to root but mark as async
			eventNode := &TreeNode{
				Event:       &event,
				Children:    []*TreeNode{},
				Expanded:    false,
				Depth:       1,
				DisplayText: "Notification (async)",
			}
			root.Children = append(root.Children, eventNode)

		default:
			// Unknown event type - add to root
			eventNode := &TreeNode{
				Event:       &event,
				Children:    []*TreeNode{},
				Expanded:    false,
				Depth:       1,
				DisplayText: eventName,
			}
			root.Children = append(root.Children, eventNode)
		}
	}

	return root
}

// parseTimestamp parses an ISO 8601 timestamp string, trying multiple formats.
// Tries RFC3339Nano first, then falls back to RFC3339.
// Returns the parsed time or zero time if parsing fails.
func parseTimestamp(timestamp string) time.Time {
	if timestamp == "" {
		return time.Time{}
	}

	// Try RFC3339Nano first (most precise)
	ts, err := time.Parse(time.RFC3339Nano, timestamp)
	if err == nil {
		return ts
	}

	// Fall back to RFC3339
	ts, err = time.Parse(time.RFC3339, timestamp)
	if err == nil {
		return ts
	}

	return time.Time{}
}

// extractTimestamp extracts timestamp from event JSONData
func extractTimestamp(event storage.Event) int64 {
	var data eventData
	if err := json.Unmarshal(event.JSONData, &data); err != nil {
		return 0
	}

	return parseTimestamp(data.Timestamp).Unix()
}

// countPrompts counts UserPromptSubmit events in the given slice
func countPrompts(events []storage.Event) int {
	count := 0
	for _, e := range events {
		if e.EventName == "UserPromptSubmit" {
			count++
		}
	}
	return count
}

// formatTime formats timestamp for display
func formatTime(ts int64) string {
	// Simple formatting - could be enhanced
	return fmt.Sprintf("%d", ts)
}

// extractToolName extracts tool_name from event JSONData
func extractToolName(event storage.Event) string {
	var data eventData
	if err := json.Unmarshal(event.JSONData, &data); err != nil {
		return ""
	}
	return data.ToolName
}

// isToolEvent checks if an event is a tool-related event
func isToolEvent(eventName string) bool {
	toolEvents := map[string]bool{
		"PreToolUse":         true,
		"PermissionRequest":  true,
		"PostToolUse":        true,
		"PostToolUseFailure": true,
	}
	return toolEvents[eventName]
}

// BuildTreeViewWidget creates the tview TreeView widget from the tree structure
func (a *App) BuildTreeViewWidget(root *TreeNode) *tview.TreeView {
	tree := tview.NewTreeView()
	tree.SetRoot(tview.NewTreeNode(root.DisplayText).
		SetReference(root))

	// Build tree recursively
	a.buildTreeNode(tree.GetRoot(), root)

	// Set up selection handler (Enter key)
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		a.handleTreeNodeSelected(node)
	})

	return tree
}

// buildTreeNode recursively builds tview tree nodes from our tree structure
func (a *App) buildTreeNode(parentNode *tview.TreeNode, treeNode *TreeNode) {
	for _, child := range treeNode.Children {
		text := child.DisplayText
		if child.Event != nil {
			// Only add ID if DisplayText doesn't already contain angle brackets
			// (Tool events already have <tool_name> in DisplayText, non-tool events need [ID])
			if !strings.ContainsRune(text, '<') && !strings.ContainsRune(text, '>') {
				// No brackets yet, add ID
				text = fmt.Sprintf("%s [%d]", text, child.Event.ID)
			}
		}

		node := tview.NewTreeNode(text).
			SetReference(child).
			SetExpanded(child.Expanded)

		// Set different colors for different depths
		if child.Depth == 0 {
			node.SetColor(a.Styles.HeaderColor)
		} else if child.Depth == 1 {
			node.SetColor(a.Styles.HighlightColor)
		}

		parentNode.AddChild(node)

		// Recursively build children
		if len(child.Children) > 0 {
			a.buildTreeNode(node, child)
		}
	}
}

// handleTreeNodeSelected is called when a tree node is selected (Enter pressed)
func (a *App) handleTreeNodeSelected(node *tview.TreeNode) {
	// Toggle expanded state
	ref := node.GetReference()
	if ref == nil {
		return
	}

	if treeNode, ok := ref.(*TreeNode); ok {
		if len(treeNode.Children) > 0 {
			treeNode.Expanded = !treeNode.Expanded
			node.SetExpanded(treeNode.Expanded)
		} else if treeNode.Event != nil {
			// Leaf node - update details view
			a.updateDetailsViewFromEvent(treeNode.Event)
		}
	}
}

// updateDetailsViewFromEvent updates the details view with a specific event
func (a *App) updateDetailsViewFromEvent(event *storage.Event) {
	jsonStr := formatJSON(event.JSONData)
	a.DetailsView.SetText(jsonStr)
	a.DetailsView.ScrollToBeginning()

	// Find event index in our events list
	index := -1
	for i, e := range a.Events {
		if e.ID == event.ID {
			index = i
			break
		}
	}

	if index >= 0 {
		a.DetailsView.SetTitle(fmt.Sprintf("Event Details (%d/%d)", index+1, len(a.Events)))
	} else {
		a.DetailsView.SetTitle("Event Details")
	}
}

// FlattenTree converts tree back to flat event list (for preserving state)
func (a *App) FlattenTree(root *TreeNode) []storage.Event {
	var events []storage.Event
	a.flattenTreeNode(root, &events)
	return events
}

// flattenTreeNode recursively flattens tree nodes
func (a *App) flattenTreeNode(node *TreeNode, events *[]storage.Event) {
	if node.Event != nil {
		*events = append(*events, *node.Event)
	}
	for _, child := range node.Children {
		a.flattenTreeNode(child, events)
	}
}

// GetTreeLeafCount returns the number of leaf nodes (actual events) in the tree
func (a *App) GetTreeLeafCount(root *TreeNode) int {
	count := 0
	a.countLeafNodes(root, &count)
	return count
}

// countLeafNodes recursively counts leaf nodes
func (a *App) countLeafNodes(node *TreeNode, count *int) {
	if node.Event != nil {
		*count++
	}
	for _, child := range node.Children {
		a.countLeafNodes(child, count)
	}
}
