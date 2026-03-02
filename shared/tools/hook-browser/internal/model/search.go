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
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SetupSearchModal creates and configures the search modal
func (a *App) SetupSearchModal() {
	// Create input field for search
	a.SearchField = tview.NewInputField()
	a.SearchField.SetLabel("Search: ")
	a.SearchField.SetPlaceholder("Type to filter events...")
	a.SearchField.SetFieldWidth(50)
	a.SearchField.SetBackgroundColor(a.Styles.StatusBarBg)
	a.SearchField.SetFieldTextColor(a.Styles.StatusBarText)

	// Wrap in a flex for centering
	searchFlex := tview.NewFlex()
	searchFlex.SetDirection(tview.FlexRow)
	searchFlex.AddItem(tview.NewBox(), 0, 1, false)
	searchFlex.AddItem(tview.NewBox(), 0, 1, false)
	searchFlex.AddItem(tview.NewBox(), 0, 1, false)

	horizontalFlex := tview.NewFlex()
	horizontalFlex.SetDirection(tview.FlexColumn)
	horizontalFlex.AddItem(tview.NewBox(), 0, 1, false)
	horizontalFlex.AddItem(a.SearchField, 0, 2, false)
	horizontalFlex.AddItem(tview.NewBox(), 0, 1, false)

	searchFlex.AddItem(horizontalFlex, 0, 1, false)

	// Add border
	searchFlex.SetBorder(true)
	searchFlex.SetBorderColor(a.Styles.HighlightColor)
	searchFlex.SetTitle("Search Events")
	searchFlex.SetTitleColor(a.Styles.HighlightColor)

	// Add to pages
	a.Pages.AddPage("search", searchFlex, true, false)
}

// ShowSearchModal shows the search modal
func (a *App) ShowSearchModal() {
	a.SearchField.SetText("")
	a.CurrentPage = "search"
	a.Pages.SwitchToPage("search")
	a.Application.SetFocus(a.SearchField)

	// Set up search handler
	a.SearchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			// Perform search
			query := a.SearchField.GetText()
			a.PerformSearch(query)
			a.ShowEventsPage()
			return nil
		case tcell.KeyEscape:
			// Cancel search
			a.ClearSearch()
			a.ShowEventsPage()
			return nil
		}
		return event
	})

	// Set up changed func for real-time search
	a.SearchField.SetChangedFunc(func(text string) {
		// Real-time search as user types
		if len(text) > 0 {
			a.PerformSearch(text)
		} else {
			a.ClearSearch()
		}
	})
}

// PerformSearch searches for events matching the query
func (a *App) PerformSearch(query string) {
	if query == "" {
		a.ClearSearch()
		return
	}

	a.SearchQuery = query
	a.SearchMatches = []int{}
	a.SearchActive = true

	// Search through all events
	for i, event := range a.Events {
		if strings.Contains(strings.ToLower(event.EventName), strings.ToLower(query)) {
			a.SearchMatches = append(a.SearchMatches, i)
		}
	}

	// Update event list to show only matches
	a.FilterEventList(a.SearchMatches)

	// Update status bar with match count
	if len(a.SearchMatches) > 0 {
		a.CurrentMatchIndex = 0
		a.StatusBar.SetText(formatText("Search: '%s' | %d matches | 1/%d | n:next N:prev Esc:clear",
			query, len(a.SearchMatches), len(a.SearchMatches)))
	} else {
		a.StatusBar.SetText(formatText("Search: '%s' | No matches found | Esc:clear", query))
	}

	// Select first match if available
	if len(a.SearchMatches) > 0 {
		a.EventList.SetCurrentItem(0)
	}
}

// FilterEventList filters the event list to show only specific indices
func (a *App) FilterEventList(indices []int) {
	a.EventList.Clear()

	for _, idx := range indices {
		if idx >= 0 && idx < len(a.Events) {
			event := a.Events[idx]
			item := EventItem{
				Seq:   uint64(event.ID),
				Name:  event.EventName,
				Index: idx,
			}
			// Highlight matching text
			displayText := FormatEventListItem(item)
			if a.SearchQuery != "" {
				displayText = a.highlightMatch(displayText, a.SearchQuery)
			}
			a.EventList.AddItem(displayText, "", 0, nil)
		}
	}
}

// highlightMatch highlights the search query in the text
func (a *App) highlightMatch(text, query string) string {
	if query == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	var result strings.Builder
	lastIndex := 0

	for {
		idx := strings.Index(lowerText[lastIndex:], lowerQuery)
		if idx == -1 {
			result.WriteString(text[lastIndex:])
			break
		}

		idx += lastIndex

		// Add text before match
		result.WriteString(text[lastIndex:idx])

		// Add highlighted match
		matchText := text[idx : idx+len(query)]
		result.WriteString(fmt.Sprintf("[%s::b]%s[-]", colorToHex(a.Styles.HighlightColor), matchText))

		lastIndex = idx + len(query)
	}

	return result.String()
}

// NextMatch navigates to the next search match
func (a *App) NextMatch() {
	if !a.SearchActive || len(a.SearchMatches) == 0 {
		return
	}

	a.CurrentMatchIndex = (a.CurrentMatchIndex + 1) % len(a.SearchMatches)

	// In filtered list, the current match index corresponds directly to list index
	listIdx := a.CurrentMatchIndex

	a.EventList.SetCurrentItem(listIdx)
	a.StatusBar.SetText(formatText("Search: '%s' | %d matches | %d/%d | n:next N:prev Esc:clear",
		a.SearchQuery, len(a.SearchMatches), a.CurrentMatchIndex+1, len(a.SearchMatches)))
}

// PrevMatch navigates to the previous search match
func (a *App) PrevMatch() {
	if !a.SearchActive || len(a.SearchMatches) == 0 {
		return
	}

	a.CurrentMatchIndex = (a.CurrentMatchIndex - 1 + len(a.SearchMatches)) % len(a.SearchMatches)

	// In filtered list, the current match index corresponds directly to list index
	listIdx := a.CurrentMatchIndex

	a.EventList.SetCurrentItem(listIdx)
	a.StatusBar.SetText(formatText("Search: '%s' | %d matches | %d/%d | n:next N:prev Esc:clear",
		a.SearchQuery, len(a.SearchMatches), a.CurrentMatchIndex+1, len(a.SearchMatches)))
}

// ClearSearch clears the current search and shows all events
func (a *App) ClearSearch() {
	a.SearchActive = false
	a.SearchQuery = ""
	a.SearchMatches = []int{}
	a.CurrentMatchIndex = 0

	// Repopulate with all events
	a.PopulateEventList(a.Events)
	a.UpdateStatusBar()
}
