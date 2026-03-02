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


package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/code-together/shared/hook-common/config"
	"github.com/code-together/shared/hook-common/storage"
	"github.com/code-together/shared/tools/hook-browser/internal/model"
)

var (
	dbPath   string
	toolName string
	summary  bool
	date     string
)

func init() {
	defaultDBPath, err := config.GetDefaultDatabasePath()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&dbPath, "db-path", defaultDBPath, "Path to hook events database (SQLite)")
	flag.StringVar(&toolName, "tool-name", "", "Tool name for isolation (e.g., claude-code, codex). Uses default if not specified.")
	flag.BoolVar(&summary, "summary", false, "Generate summary report as JSON (requires -date flag)")
	flag.StringVar(&date, "date", "", "Target date for summary report (format: YYYY-MM-DD, required with -summary)")
}

// SessionWithTime wraps SessionInfo with first event timestamp for sorting
type SessionWithTime struct {
	Info           storage.SessionInfo
	FirstEventTime string
}

// getSortedSessions retrieves sessions and sorts them by first event time (newest first)
func getSortedSessions(db *sql.DB, toolName string) ([]SessionWithTime, error) {
	// Get all sessions
	sessions, err := storage.GetSessionInfo(db, toolName)
	if err != nil {
		return nil, err
	}

	// Create slice with timestamps
	var sessionsWithTime []SessionWithTime

	for _, session := range sessions {
		// Get first event timestamp
		var firstEventTime string
		firstEventTs, err := storage.GetFirstEventTimeSQLite(db, session.ID, toolName)
		if err == nil && firstEventTs > 0 {
			firstEventTime = time.Unix(firstEventTs, 0).Format(time.RFC3339)
		}

		sessionsWithTime = append(sessionsWithTime, SessionWithTime{
			Info:           session,
			FirstEventTime: firstEventTime,
		})
	}

	// Sort by FirstEventTime (newest first)
	sort.Slice(sessionsWithTime, func(i, j int) bool {
		ti, tj := sessionsWithTime[i].FirstEventTime, sessionsWithTime[j].FirstEventTime

		if ti != "" && tj != "" {
			return ti > tj // Newer first
		}
		if ti != "" {
			return true
		}
		if tj != "" {
			return false
		}
		return sessionsWithTime[i].Info.ID < sessionsWithTime[j].Info.ID
	})

	return sessionsWithTime, nil
}

func main() {
	flag.Parse()

	dbHandle, err := storage.OpenDB(dbPath, true, toolName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer dbHandle.Close()

	// If summary flag is set, generate summary report and exit
	if summary {
		exitCode := generateSummaryReport(dbHandle, toolName, date)
		os.Exit(exitCode)
	}

	sessionsWithTime, err := getSortedSessions(dbHandle, toolName)
	if err != nil || len(sessionsWithTime) == 0 {
		fmt.Fprintf(os.Stderr, "No sessions found in database\n")
		os.Exit(1)
	}

	sessions := make([]string, len(sessionsWithTime))
	for i, swt := range sessionsWithTime {
		sessions[i] = swt.Info.ID
	}

	events, err := storage.GetSessionEvents(dbHandle, sessions[0], toolName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading events: %v\n", err)
		os.Exit(1)
	}

	// Convert SessionWithTime to model.SessionWithTime
	modelSessionWithTime := make([]model.SessionWithTime, len(sessionsWithTime))
	for i, swt := range sessionsWithTime {
		modelSessionWithTime[i] = model.SessionWithTime{
			Info:           swt.Info,
			FirstEventTime: swt.FirstEventTime,
		}
	}

	// Create the app
	app := model.NewApp(dbHandle, toolName, sessions, events, sessions[0], modelSessionWithTime)

	// Build and set up tree view (must be done before SetupTreeViewPage)
	treeRoot := app.BuildEventTree(events)
	app.TreeRoot = treeRoot
	app.TreeView = app.BuildTreeViewWidget(treeRoot)

	// Setup UI
	app.SetupEventsPage()
	app.SetupTreeViewPage()
	app.SetupSessionPickerPage()
	app.SetupHelpPage()
	app.SetupSearchModal()
	app.SetupInputHandlers()

	// Populate data
	app.PopulateEventList(events)
	app.PopulateSessionList(modelSessionWithTime)

	// Determine initial mode
	initialMode := app.SetInitialMode()
	if initialMode == model.ModeSessionPicker {
		app.ShowSessionPickerPage()
	} else {
		app.ShowEventsPage()
	}

	// Run the application
	if err := app.Application.SetRoot(app.Pages, true).EnableMouse(true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
