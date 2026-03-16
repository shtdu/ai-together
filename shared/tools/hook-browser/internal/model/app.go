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
	"database/sql"
	"fmt"
	"time"

	"github.com/rivo/tview"

	"github.com/code-together/shared/hook-common/storage"
)

// App represents the application state and UI components
type App struct {
	// tview application
	Application *tview.Application

	// UI containers
	Pages    *tview.Pages
	MainFlex *tview.Flex

	// Widgets
	EventList   *tview.List
	DetailsView *tview.TextView
	SessionList *tview.List
	StatusBar   *tview.TextView
	TreeView    *tview.TreeView
	TreeRoot    *TreeNode

	// Notification system
	NotificationView *tview.TextView
	NotificationTimer *time.Timer
	NotificationActive bool

	// Search system
	SearchField      *tview.InputField
	SearchQuery      string
	SearchMatches    []int // Indices of matching events
	CurrentMatchIndex int
	SearchActive     bool

	// Data
	DB             *sql.DB
	ToolName       string
	Sessions       []string
	Events         []storage.Event
	CurrentSession string
	CurrentPage    string // Track current page name

	// Session metadata for display
	SessionWithTime []SessionWithTime

	// Styling
	Styles *Styles
}

// SessionWithTime wraps SessionInfo with first event timestamp for sorting
type SessionWithTime struct {
	Info           storage.SessionInfo
	FirstEventTime string
}

// NewApp creates and initializes a new App instance
func NewApp(db *sql.DB, toolName string, sessions []string, events []storage.Event, firstSession string, sessionWithTime []SessionWithTime) *App {
	styles := DefaultStyles()

	app := &App{
		Application:     tview.NewApplication(),
		DB:              db,
		ToolName:        toolName,
		Sessions:        sessions,
		Events:          events,
		CurrentSession:  firstSession,
		SessionWithTime: sessionWithTime,
		Styles:          styles,
	}

	// Initialize widgets
	app.EventList = tview.NewList()
	app.EventList.ShowSecondaryText(false)
	app.EventList.SetSelectedBackgroundColor(styles.StatusBarBg)
	app.EventList.SetSelectedTextColor(styles.HighlightColor)
	app.EventList.SetMainTextColor(styles.NormalText)

	app.DetailsView = tview.NewTextView()
	app.DetailsView.SetDynamicColors(true)
	app.DetailsView.SetScrollable(true)
	app.DetailsView.SetWrap(true)
	app.DetailsView.SetTextColor(styles.NormalText)

	app.SessionList = tview.NewList()
	app.SessionList.ShowSecondaryText(false)
	app.SessionList.SetSelectedBackgroundColor(styles.StatusBarBg)
	app.SessionList.SetSelectedTextColor(styles.HighlightColor)
	app.SessionList.SetMainTextColor(styles.NormalText)

	app.StatusBar = tview.NewTextView()
	app.StatusBar.SetTextColor(styles.StatusBarText)
	app.StatusBar.SetBackgroundColor(styles.StatusBarBg)

	// Create pages container
	app.Pages = tview.NewPages()

	return app
}

// SetInitialMode sets the initial UI mode based on session count
func (a *App) SetInitialMode() Mode {
	if len(a.Sessions) > 1 {
		return ModeSessionPicker
	}
	return ModeEvents
}

// PopulateEventList populates the event list with events
func (a *App) PopulateEventList(events []storage.Event) {
	a.EventList.Clear()
	a.Events = events

	for i, event := range events {
		item := EventItem{
			Seq:   uint64(event.ID),
			Name:  event.EventName,
			Index: i,
		}
		a.EventList.AddItem(FormatEventListItem(item), "", 0, nil)
	}
}

// PopulateSessionList populates the session list with sessions
func (a *App) PopulateSessionList(sessionWithTime []SessionWithTime) {
	a.SessionList.Clear()
	a.SessionWithTime = sessionWithTime

	for i, swt := range sessionWithTime {
		item := SessionItem{
			ID:             swt.Info.ID,
			Count:          swt.Info.Count,
			Index:          i,
			FirstEventTime: swt.FirstEventTime,
		}
		a.SessionList.AddItem(FormatSessionListItem(item), "", 0, nil)
	}
}

// UpdateStatusBar updates the status bar text
func (a *App) UpdateStatusBar() {
	mode := "List"
	if a.CurrentPage == "tree" {
		mode = "Tree"
	}

	a.StatusBar.SetText(
		formatText("Session: %s | Events: %d | Mode: %s | t:toggle | ↑/↓:nav | /:search | c:copy | Esc/s:switch | r:refresh | q:quit",
			a.CurrentSession, len(a.Events), mode),
	)
}

// formatText is a simple text formatting helper
func formatText(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
