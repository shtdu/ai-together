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


package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/gdamore/tcell/v2"

	"github.com/code-together/shared/hook-common/storage"
)

// SetupInputHandlers configures keyboard input handling for the application
func (a *App) SetupInputHandlers() {
	// Global input capture for both pages
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			a.Stop()
			return nil
		}

		// Check which page is visible using tracked current page
		switch a.CurrentPage {
		case "events":
			return a.handleEventsInput(event)
		case "tree":
			return a.handleTreeInput(event)
		case "picker":
			return a.handlePickerInput(event)
		case "help":
			// Any key closes help
			a.ShowEventsPage()
			return nil
		}

		return event
	})

	// Event list input handler
	a.EventList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return a.handleEventsListInput(event)
	})

	// Session list input handler
	a.SessionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return a.handleSessionListInput(event)
	})
}

// handleEventsInput handles global input when in events mode
func (a *App) handleEventsInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		a.Stop()
		return nil
	case 's':
		a.ShowSessionPickerPage()
		return nil
	case 'r':
		a.RefreshEvents()
		return nil
	case 'c':
		a.CopyEventContent()
		return nil
	case '?':
		a.ShowHelpPage()
		return nil
	case '/':
		a.ShowSearchModal()
		return nil
	case 'n':
		a.NextMatch()
		return nil
	case 'N':
		a.PrevMatch()
		return nil
	case 't':
		// Toggle to tree view
		a.ShowTreeViewPage()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEscape:
		if a.SearchActive {
			a.ClearSearch()
			return nil
		}
		a.ShowSessionPickerPage()
		return nil
	}

	return event
}

// handleEventsListInput handles keyboard input for the events list
func (a *App) handleEventsListInput(event *tcell.EventKey) *tcell.EventKey {
	currentIndex := a.EventList.GetCurrentItem()
	itemCount := a.EventList.GetItemCount()

	switch event.Rune() {
	case 'J':
		// Move down 10 items
		newIndex := currentIndex + 10
		if newIndex >= itemCount {
			newIndex = itemCount - 1
		}
		if newIndex != currentIndex {
			a.EventList.SetCurrentItem(newIndex)
		}
		return nil

	case 'K':
		// Move up 10 items
		newIndex := currentIndex - 10
		if newIndex < 0 {
			newIndex = 0
		}
		if newIndex != currentIndex {
			a.EventList.SetCurrentItem(newIndex)
		}
		return nil
	}

	switch event.Key() {
	case tcell.KeyPgUp, tcell.KeyPgDn:
		// Let TextView handle pgup/pgdn natively
		return event
	}

	return event
}

// handlePickerInput handles global input when in session picker mode
func (a *App) handlePickerInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		a.Stop()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEscape:
		// Close picker and return to events
		a.ShowEventsPage()
		return nil
	}

	return event
}

// handleSessionListInput handles keyboard input for the session list
func (a *App) handleSessionListInput(event *tcell.EventKey) *tcell.EventKey {
	currentIndex := a.SessionList.GetCurrentItem()

	switch event.Key() {
	case tcell.KeyEnter:
		// Select the session
		if currentIndex >= 0 && currentIndex < len(a.SessionWithTime) {
			sessionID := a.SessionWithTime[currentIndex].Info.ID
			a.SwitchSession(sessionID)
			a.ShowEventsPage()
		}
		return nil
	}

	return event
}

// SwitchSession switches to a different session
func (a *App) SwitchSession(sessionID string) {
	events, err := storage.GetSessionEvents(a.DB, sessionID, a.ToolName)
	if err != nil {
		a.ShowNotification(fmt.Sprintf("Failed to load session: %v", err), SeverityError)
		return
	}

	a.CurrentSession = sessionID
	a.Events = events // Update events for tree view
	a.PopulateEventList(events)
	a.DetailsView.SetText("")

	// Rebuild tree for new session
	a.TreeRoot = a.BuildEventTree(events)
	a.TreeView = a.BuildTreeViewWidget(a.TreeRoot)
	a.RebuildTreeViewPage()
}

// RefreshEvents reloads events from the database for the current session
func (a *App) RefreshEvents() {
	events, err := storage.GetSessionEvents(a.DB, a.CurrentSession, a.ToolName)
	if err != nil {
		a.ShowNotification(fmt.Sprintf("Failed to refresh events: %v", err), SeverityError)
		return
	}

	a.Events = events // Update events for tree view
	a.PopulateEventList(events)

	// Rebuild tree with refreshed events
	a.TreeRoot = a.BuildEventTree(events)
	a.TreeView = a.BuildTreeViewWidget(a.TreeRoot)
	a.RebuildTreeViewPage()

	a.ShowNotification("Events refreshed", SeveritySuccess)
}

// Stop stops the application
func (a *App) Stop() {
	a.Application.Stop()
}

// CopyEventContent copies the currently selected event's JSON content to clipboard
func (a *App) CopyEventContent() {
	currentIndex := a.EventList.GetCurrentItem()
	if currentIndex < 0 || currentIndex >= len(a.Events) {
		a.ShowNotification("No event selected", SeverityWarning)
		return
	}

	event := a.Events[currentIndex]
	formatted := formatJSONEvent(event.JSONData)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("clip")
	default:
		a.ShowNotification("Clipboard not supported on this platform", SeverityError)
		return
	}

	cmd.Stdin = bytes.NewReader([]byte(formatted))
	if err := cmd.Run(); err != nil {
		a.ShowNotification(fmt.Sprintf("Failed to copy: %v", err), SeverityError)
		return
	}

	a.ShowNotification("Event content copied to clipboard", SeveritySuccess)
}

// formatJSONEvent formats JSON bytes with indentation
func formatJSONEvent(jsonData []byte) string {
	var prettyJSON interface{}
	if err := json.Unmarshal(jsonData, &prettyJSON); err != nil {
		return string(jsonData)
	}
	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	return string(formatted)
}

// handleTreeInput handles global input when in tree view mode
func (a *App) handleTreeInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		a.Stop()
		return nil
	case 't':
		// Toggle back to list view
		a.ShowEventsPage()
		return nil
	case 's':
		a.ShowSessionPickerPage()
		return nil
	case '?':
		a.ShowHelpPage()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEscape:
		a.ShowEventsPage()
		return nil
	}

	return event
}
