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
	"encoding/json"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// emptyBox creates an empty box primitive for spacing
func emptyBox() tview.Primitive {
	return tview.NewBox()
}

// SetupEventsPage creates and configures the main events page layout
func (a *App) SetupEventsPage() {
	// Create main flex container for vertical layout
	mainContainer := tview.NewFlex()
	mainContainer.SetDirection(tview.FlexRow)

	// Create the two-pane horizontal flex for content
	contentFlex := tview.NewFlex()
	contentFlex.SetDirection(tview.FlexColumn)

	// Events pane (25% width)
	eventsPane := tview.NewFlex()
	eventsPane.SetDirection(tview.FlexRow)
	eventsPane.SetBorder(true)
	eventsPane.SetBorderColor(a.Styles.BorderColor)
	eventsPane.SetTitle("Events")
	eventsPane.SetTitleColor(a.Styles.HeaderColor)

	// Add events header and list
	eventsPane.AddItem(emptyBox(), 1, 0, false) // Space for header
	eventsPane.AddItem(a.EventList, 0, 1, true)

	// Details pane (75% width)
	detailsPane := tview.NewFlex()
	detailsPane.SetDirection(tview.FlexRow)
	detailsPane.SetBorder(true)
	detailsPane.SetBorderColor(a.Styles.BorderColor)
	detailsPane.SetTitle("Event Details")
	detailsPane.SetTitleColor(a.Styles.DimText)

	// Add details header and content
	detailsPane.AddItem(emptyBox(), 1, 0, false) // Space for header
	detailsPane.AddItem(emptyBox(), 1, 0, false) // Blank line
	detailsPane.AddItem(a.DetailsView, 0, 1, false)

	// Add panes to content flex (25% / 75% split)
	contentFlex.AddItem(eventsPane, 0, 1, true)
	contentFlex.AddItem(detailsPane, 0, 3, false)

	// Add content flex to main container
	mainContainer.AddItem(contentFlex, 0, 1, true)

	// Add status bar at bottom
	mainContainer.AddItem(a.StatusBar, 1, 0, false)

	// Set up event list selection handler
	a.EventList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		a.updateDetailsView(index)
	})

	// Store the main flex reference
	a.MainFlex = mainContainer

	// Add the events page to pages
	a.Pages.AddPage("events", mainContainer, true, false)
}

// SetupTreeViewPage creates and configures the tree view page layout
func (a *App) SetupTreeViewPage() {
	// Create main flex container for vertical layout
	mainContainer := tview.NewFlex()
	mainContainer.SetDirection(tview.FlexRow)

	// Create the two-pane horizontal flex for content
	contentFlex := tview.NewFlex()
	contentFlex.SetDirection(tview.FlexColumn)

	// Tree pane (35% width - wider for indentation)
	treePane := tview.NewFlex()
	treePane.SetDirection(tview.FlexRow)
	treePane.SetBorder(true)
	treePane.SetBorderColor(a.Styles.BorderColor)
	treePane.SetTitle("Tree View")
	treePane.SetTitleColor(a.Styles.HeaderColor)

	// Add tree header and view
	treePane.AddItem(emptyBox(), 1, 0, false) // Space for header
	treePane.AddItem(a.TreeView, 0, 1, true)

	// Details pane (65% width)
	detailsPane := tview.NewFlex()
	detailsPane.SetDirection(tview.FlexRow)
	detailsPane.SetBorder(true)
	detailsPane.SetBorderColor(a.Styles.BorderColor)
	detailsPane.SetTitle("Event Details")
	detailsPane.SetTitleColor(a.Styles.DimText)

	// Add details header and content
	detailsPane.AddItem(emptyBox(), 1, 0, false) // Space for header
	detailsPane.AddItem(emptyBox(), 1, 0, false) // Blank line
	detailsPane.AddItem(a.DetailsView, 0, 1, false)

	// Add panes to content flex (35% / 65% split)
	contentFlex.AddItem(treePane, 0, 35, true)
	contentFlex.AddItem(detailsPane, 0, 65, false)

	// Add content flex to main container
	mainContainer.AddItem(contentFlex, 0, 1, true)

	// Add status bar at bottom
	mainContainer.AddItem(a.StatusBar, 1, 0, false)

	// Add the tree view page to pages
	a.Pages.AddPage("tree", mainContainer, true, false)
}

// ShowTreeViewPage switches to the tree view page
func (a *App) ShowTreeViewPage() {
	a.CurrentPage = "tree"

	// Rebuild tree to ensure it's current (in case events changed)
	if len(a.Events) > 0 {
		a.TreeRoot = a.BuildEventTree(a.Events)
		a.TreeView = a.BuildTreeViewWidget(a.TreeRoot)
		a.RebuildTreeViewPage()
	}

	a.Pages.SwitchToPage("tree")
	a.Application.SetFocus(a.TreeView)
	a.UpdateStatusBar()
}

// RebuildTreeViewPage rebuilds the tree view page with current events
func (a *App) RebuildTreeViewPage() {
	// Remove the old tree page if it exists
	if a.Pages.HasPage("tree") {
		a.Pages.RemovePage("tree")
	}

	// Rebuild the tree view page with current events
	a.SetupTreeViewPage()
}

// SetupSessionPickerPage creates and configures the session picker modal
func (a *App) SetupSessionPickerPage() {
	// Create modal form
	modal := tview.NewForm()
	modal.SetButtonsAlign(tview.AlignCenter)

	// Create a flex container for the modal content
	modalFlex := tview.NewFlex()
	modalFlex.SetDirection(tview.FlexRow)

	// Add title
	titleText := tview.NewTextView()
	titleText.SetText("Select a Session")
	titleText.SetTextColor(a.Styles.HighlightColor)
	titleText.SetTextAlign(tview.AlignCenter)

	// Add help text
	helpText := tview.NewTextView()
	helpText.SetText("↑/↓: Navigate • Enter: Select • Esc: Cancel • q: Quit")
	helpText.SetTextColor(a.Styles.DimText)
	helpText.SetTextAlign(tview.AlignCenter)

	// Combine all elements
	modalFlex.AddItem(emptyBox(), 0, 1, false) // Spacer
	modalFlex.AddItem(titleText, 1, 0, false)
	modalFlex.AddItem(emptyBox(), 1, 0, false)
	modalFlex.AddItem(a.SessionList, 0, 4, true)
	modalFlex.AddItem(emptyBox(), 1, 0, false)
	modalFlex.AddItem(helpText, 1, 0, false)
	modalFlex.AddItem(emptyBox(), 0, 1, false) // Spacer

	// Wrap in a centered panel
	modalPanel := tview.NewFlex()
	modalPanel.SetDirection(tview.FlexColumn)
	modalPanel.AddItem(emptyBox(), 0, 1, false)
	modalPanel.AddItem(modalFlex, 0, 2, true)
	modalPanel.AddItem(emptyBox(), 0, 1, false)

	horizontalCenter := tview.NewFlex()
	horizontalCenter.SetDirection(tview.FlexRow)
	horizontalCenter.AddItem(emptyBox(), 0, 1, false)
	horizontalCenter.AddItem(modalPanel, 0, 5, true)
	horizontalCenter.AddItem(emptyBox(), 0, 1, false)

	// Add background
	modalPanel.SetBorder(true)
	modalPanel.SetBorderColor(a.Styles.HighlightColor)
	modalPanel.SetBackgroundColor(tcell.ColorBlack)

	// Add the picker page to pages
	a.Pages.AddPage("picker", horizontalCenter, true, false)
}

// ShowEventsPage switches to the events page
func (a *App) ShowEventsPage() {
	a.CurrentPage = "events"
	a.Pages.SwitchToPage("events")
	a.Application.SetFocus(a.EventList)
	a.UpdateStatusBar()
}

// ShowSessionPickerPage switches to the session picker page
func (a *App) ShowSessionPickerPage() {
	a.CurrentPage = "picker"
	a.Pages.SwitchToPage("picker")
	a.Application.SetFocus(a.SessionList)
}

// SetupHelpPage creates and configures the help screen
func (a *App) SetupHelpPage() {
	// Create help content
	helpText := tview.NewTextView()
	helpText.SetDynamicColors(true)
	helpText.SetTextColor(a.Styles.NormalText)
	helpText.SetTextAlign(tview.AlignLeft)

	// Build help content
	helpContent := fmt.Sprintf(`[%s]
  ╔═══════════════════════════════════════════════════════════╗
  ║                    Keyboard Shortcuts                     ║
  ╚═══════════════════════════════════════════════════════════╝
`,
		colorToHex(a.Styles.HighlightColor))

	helpContent += fmt.Sprintf(`
[%s]Navigation[%s]
  ↑/↓ or j/k     Move up/down through events
  J/K            Page up/down (10 items at a time)
  PageUp/PageDn  Page up/down
  :              Jump to event number

[%s]Actions[%s]
  c              Copy current event to clipboard
  r              Refresh events from database
  ?              Show this help screen
  q              Quit application

[%s]Session Management[%s]
  s              Show session picker
  Esc            Return to events / close session picker
  Enter          Select session

[%s]Search[%s]
  /              Start search
  n              Next search match
  N              Previous search match

Press any key to close help
`,
		colorToHex(a.Styles.HeaderColor), colorToHex(a.Styles.NormalText),
		colorToHex(a.Styles.HeaderColor), colorToHex(a.Styles.NormalText),
		colorToHex(a.Styles.HeaderColor), colorToHex(a.Styles.NormalText),
		colorToHex(a.Styles.HeaderColor), colorToHex(a.Styles.NormalText),
	)

	helpText.SetText(helpContent)

	// Create modal wrapper
	modalFlex := tview.NewFlex()
	modalFlex.SetDirection(tview.FlexRow)
	modalFlex.AddItem(emptyBox(), 0, 1, false)
	modalFlex.AddItem(helpText, 0, 20, false)
	modalFlex.AddItem(emptyBox(), 0, 1, false)

	horizontalFlex := tview.NewFlex()
	horizontalFlex.SetDirection(tview.FlexColumn)
	horizontalFlex.AddItem(emptyBox(), 0, 1, false)
	horizontalFlex.AddItem(modalFlex, 0, 3, false)
	horizontalFlex.AddItem(emptyBox(), 0, 1, false)

	// Add border
	horizontalFlex.SetBorder(true)
	horizontalFlex.SetBorderColor(a.Styles.HighlightColor)
	horizontalFlex.SetTitle("Help")
	horizontalFlex.SetTitleColor(a.Styles.HighlightColor)

	// Add to pages
	a.Pages.AddPage("help", horizontalFlex, true, false)
}

// ShowHelpPage switches to the help page
func (a *App) ShowHelpPage() {
	a.CurrentPage = "help"
	a.Pages.SwitchToPage("help")
}

// updateDetailsView updates the details view with the selected event's JSON
func (a *App) updateDetailsView(index int) {
	if index >= 0 && index < len(a.Events) {
		jsonStr := formatJSON(a.Events[index].JSONData)
		a.DetailsView.SetText(jsonStr)
		a.DetailsView.ScrollToBeginning()
		a.DetailsView.SetTitle(fmt.Sprintf("Event Details (%d/%d)", index+1, len(a.Events)))
	} else {
		a.DetailsView.SetText("Select an event to view details")
		a.DetailsView.SetTitle("Event Details")
	}
}

// formatJSON formats JSON bytes with indentation
func formatJSON(jsonData []byte) string {
	var prettyJSON interface{}
	if err := json.Unmarshal(jsonData, &prettyJSON); err != nil {
		return string(jsonData)
	}
	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	return string(formatted)
}

// colorToHex converts tcell.Color to hex string for tview
func colorToHex(c tcell.Color) string {
	return fmt.Sprintf("#%06x", c.Hex())
}
