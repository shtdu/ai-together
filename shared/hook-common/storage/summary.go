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


package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sort"
	"time"
)

// GetAllToolNames retrieves all distinct tool names from the database.
func GetAllToolNames(db *sql.DB) ([]string, error) {
	return GetAllToolNamesSQLite(db)
}

// GetDayBoundaries calculates the start and end Unix timestamps for a target date in local timezone.
// targetDate should be in "YYYY-MM-DD" format.
//
// DST Handling:
// - This function correctly handles DST transitions by calculating day boundaries based on
//   midnight in the local timezone, NOT by adding 24 hours.
// - During DST "spring forward" (23-hour day), the range [2am-3am] is still included in that day.
// - During DST "fall back" (25-hour day), the range [1am-2am] occurs twice and both instances
//   are included in that day's range.
// - Events are queried by their Unix timestamp, which is timezone-agnostic, ensuring consistent
//   results regardless of when the query is run.
//
// Data Retention:
// - Client auto-deletes data older than 14 days
// - This function validates dates are within the retention window (14 days ago to today)
//
// Returns (dayStartUnix, dayEndUnix, timezoneName, error).
func GetDayBoundaries(targetDate string) (int64, int64, string, error) {
	// Parse target date (format: YYYY-MM-DD)
	// time.Parse returns a time in UTC, so we need to convert to local timezone
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, targetDate)
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid date format: %w (expected YYYY-MM-DD)", err)
	}

	// Validate date is within retention window (14 days ago to today)
	now := time.Now()
	minDate := now.AddDate(0, 0, -14) // 14 days ago
	maxDate := now.Truncate(24 * time.Hour) // Today at midnight

	// Convert parsedDate to local timezone for comparison
	parsedInLocal := parsedDate.In(time.Local)

	if parsedInLocal.Before(minDate) {
		return 0, 0, "", fmt.Errorf("date %s is outside retention window (data only available for last 14 days)", targetDate)
	}

	if parsedInLocal.After(maxDate) {
		return 0, 0, "", fmt.Errorf("date %s is in the future (reports only available up to today)", targetDate)
	}

	// Get the local timezone location
	loc := time.Local

	// Calculate day boundaries in local timezone
	// Use Date() to create midnight in the local timezone
	year, month, day := parsedDate.Date()
	dayStart := time.Date(year, month, day, 0, 0, 0, 0, loc)

	// Calculate the NEXT day at midnight in the same timezone
	// This is DST-safe: we find the next date's midnight, not just +24 hours
	// This handles 23-hour days (spring forward) and 25-hour days (fall back) correctly
	nextDate := parsedDate.AddDate(0, 0, 1) // Add one calendar day
	nextYear, nextMonth, nextDayNum := nextDate.Date()
	dayEnd := time.Date(nextYear, nextMonth, nextDayNum, 0, 0, 0, 0, loc)

	// Get timezone name and offset
	tzName, tzOffset := time.Now().In(loc).Zone()

	// Log timezone info for debugging (only in verbose mode)
	if os.Getenv("VERBOSE") == "true" {
		slog.Debug("calculated day boundaries",
			"target", targetDate,
			"day_start", dayStart.Format(time.RFC3339),
			"day_end", dayEnd.Format(time.RFC3339),
			"timezone", tzName,
			"offset_seconds", tzOffset,
			"duration_hours", dayEnd.Sub(dayStart).Hours())
	}

	return dayStart.Unix(), dayEnd.Unix(), tzName, nil
}

// getSessionGitMeta extracts git metadata from the first event of a session.
// Returns (remoteURLs, branch, commitHash). Returns nil/empty strings if unavailable.
func getSessionGitMeta(db *sql.DB, sessionID string) ([]string, string, string) {
	// Query the first event for this session (regardless of tool or time)
	query := `
		SELECT event_data
		FROM events
		WHERE session_id = ?
		ORDER BY id ASC
		LIMIT 1
	`

	var jsonData []byte
	err := db.QueryRow(query, sessionID).Scan(&jsonData)
	if err != nil {
		// No events or error - return empty values
		return nil, "", ""
	}

	// Parse JSON to extract git metadata
	var event struct {
		Git struct {
			RemoteURLs []string `json:"remote_urls"`
			Branch     string   `json:"branch"`
			CommitHash string   `json:"commit_hash"`
		} `json:"git"`
	}

	if err := json.Unmarshal(jsonData, &event); err != nil {
		// JSON parse error - return empty values
		return nil, "", ""
	}

	// Return raw git metadata (URL parsing happens in service layer)
	return event.Git.RemoteURLs, event.Git.Branch, event.Git.CommitHash
}

// GetDailyEventTypeCounts retrieves event type counts for a specific tool on the target day.
func GetDailyEventTypeCounts(db *sql.DB, toolName string, targetDate string) (map[string]int, error) {
	dayStart, dayEnd, _, err := GetDayBoundaries(targetDate)
	if err != nil {
		return nil, err
	}

	events, err := GetEventsForToolOnDaySQLite(db, toolName, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	// Count events by type
	counts := make(map[string]int)
	for _, event := range events {
		counts[event.EventName]++
	}

	return counts, nil
}

// GetDailySessionStatistics retrieves session statistics for a specific tool on the target day.
func GetDailySessionStatistics(db *sql.DB, toolName string, targetDate string) (DailySessionStatistics, error) {
	dayStart, dayEnd, _, err := GetDayBoundaries(targetDate)
	if err != nil {
		return DailySessionStatistics{}, err
	}

	events, err := GetEventsForToolOnDaySQLite(db, toolName, dayStart, dayEnd)
	if err != nil {
		return DailySessionStatistics{}, err
	}

	// Track session timestamps: sessionID -> (firstTimestamp, lastTimestamp)
	type sessionTimestamps struct {
		firstTs int64
		lastTs  int64
	}
	timestamps := make(map[string]sessionTimestamps)

	// Track event types per session: sessionID -> map[eventType]count
	eventTypesBySession := make(map[string]map[string]int)

	for _, event := range events {
		ts := event.Timestamp
		if existing, ok := timestamps[event.SessionID]; ok {
			// Update last timestamp if current event is later
			if ts > existing.lastTs {
				existing.lastTs = ts
				timestamps[event.SessionID] = existing
			}
		} else {
			// First time seeing this session
			timestamps[event.SessionID] = sessionTimestamps{
				firstTs: ts,
				lastTs:  ts,
			}
			// Initialize event type map for this session
			eventTypesBySession[event.SessionID] = make(map[string]int)
		}

		// Count event type for this session
		eventTypesBySession[event.SessionID][event.EventName]++
	}

	// Convert to SessionDuration slice and sort by session ID
	var uniqueSessions []SessionDuration
	for sessionID, ts := range timestamps {
		beginTime := time.Unix(ts.firstTs, 0).In(time.Local)
		lastTime := time.Unix(ts.lastTs, 0).In(time.Local)
		durationMs := (ts.lastTs - ts.firstTs) * 1000 // Convert seconds to milliseconds

		// Extract git metadata from first event (regardless of tool or time)
		remoteURLs, branch, commitHash := getSessionGitMeta(db, sessionID)

		uniqueSessions = append(uniqueSessions, SessionDuration{
			SessionID:      sessionID,
			BeginTime:      beginTime,
			LastTime:       lastTime,
			Duration:       durationMs,
			EventTypeCounts: eventTypesBySession[sessionID],
			RemoteURLs:     remoteURLs,
			Branch:         branch,
			CommitHash:     commitHash,
		})
	}

	// Sort by session ID
	sort.Slice(uniqueSessions, func(i, j int) bool {
		return uniqueSessions[i].SessionID < uniqueSessions[j].SessionID
	})

	return DailySessionStatistics{
		SessionCount:    len(uniqueSessions),
		EventCount:      len(events),
		UniqueSessions: uniqueSessions,
	}, nil
}

// GetTimeRangeStatistics retrieves time range statistics for a specific tool.
// Note: Not used for single-day report, but kept for potential future use.
func GetTimeRangeStatistics(db *sql.DB, toolName string) (earliest, latest time.Time, duration int64, days int, err error) {
	earliestTs, latestTs, err := GetTimeRangeSQLite(db, toolName)
	if err != nil {
		return time.Time{}, time.Time{}, 0, 0, err
	}

	if earliestTs == 0 || latestTs == 0 {
		return time.Time{}, time.Time{}, 0, 0, nil
	}

	earliest = time.Unix(earliestTs, 0).In(time.Local)
	latest = time.Unix(latestTs, 0).In(time.Local)
	duration = latestTs - earliestTs
	days = int(duration / 86400) // 86400 seconds per day

	return earliest, latest, duration, days, nil
}

// GetDailyToolUsageStatistics retrieves tool usage statistics for a specific tool on the target day.
// Uses PostToolUse events to extract the tool_name field (e.g., Read, Write, Glob).
// These are tools used WITHIN events, not hook-browser tools.
func GetDailyToolUsageStatistics(db *sql.DB, toolName string, targetDate string) ([]ToolUsageItem, error) {
	dayStart, dayEnd, _, err := GetDayBoundaries(targetDate)
	if err != nil {
		return nil, err
	}

	events, err := GetEventsForToolOnDaySQLite(db, toolName, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	// Count tool usage from PostToolUse events
	toolCounts := make(map[string]int)
	for _, event := range events {
		if event.EventName == "PostToolUse" {
			// Parse JSON to extract tool_name
			var toolUseData struct {
				ToolName string `json:"tool_name"`
			}
			if err := json.Unmarshal(event.JSONData, &toolUseData); err == nil {
				if toolUseData.ToolName != "" {
					toolCounts[toolUseData.ToolName]++
				}
			}
		}
	}

	// Convert map to slice and sort by count (descending) then by tool name
	var items []ToolUsageItem
	for toolName, count := range toolCounts {
		items = append(items, ToolUsageItem{
			ToolName: toolName,
			Count:    count,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count // Higher count first
		}
		return items[i].ToolName < items[j].ToolName
	})

	return items, nil
}

// GetDailyPromptStatistics retrieves prompt statistics for a specific tool on the target day.
// Uses UserPromptSubmit events to extract the prompt field.
// Prompt length is calculated as UTF-8 character count.
func GetDailyPromptStatistics(db *sql.DB, toolName string, targetDate string) (DailyPromptStatistics, error) {
	dayStart, dayEnd, _, err := GetDayBoundaries(targetDate)
	if err != nil {
		return DailyPromptStatistics{}, err
	}

	events, err := GetEventsForToolOnDaySQLite(db, toolName, dayStart, dayEnd)
	if err != nil {
		return DailyPromptStatistics{}, err
	}

	var promptLengths []int
	totalLength := 0

	for _, event := range events {
		if event.EventName == "UserPromptSubmit" {
			// Parse JSON to extract prompt
			var promptData struct {
				Prompt string `json:"prompt"`
			}
			if err := json.Unmarshal(event.JSONData, &promptData); err == nil {
				if promptData.Prompt != "" {
					length := len([]rune(promptData.Prompt)) // UTF-8 character count
					promptLengths = append(promptLengths, length)
					totalLength += length
				}
			}
		}
	}

	if len(promptLengths) == 0 {
		return DailyPromptStatistics{}, nil
	}

	// Find min and max
	minLength := promptLengths[0]
	maxLength := promptLengths[0]
	for _, length := range promptLengths {
		if length < minLength {
			minLength = length
		}
		if length > maxLength {
			maxLength = length
		}
	}

	avgLength := float64(totalLength) / float64(len(promptLengths))
	// Round to 2 decimal places for cleaner output
	avgLength = math.Round(avgLength*100) / 100

	return DailyPromptStatistics{
		TotalPrompts:       len(promptLengths),
		AveragePromptLength: avgLength,
		LongestPrompt:      maxLength,
		ShortestPrompt:     minLength,
	}, nil
}
