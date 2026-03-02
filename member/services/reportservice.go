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
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"codeswitch/internal/hookdb"
	"github.com/code-together/shared/hook-common/storage"
	"github.com/code-together/shared/hook-utils"
)

// ReportService generates daily activity reports from hook events database
type ReportService struct {
	mu sync.Mutex
}

// DailyReport represents the complete daily activity report
type DailyReport struct {
	Date        string       `json:"date"`         // YYYY-MM-DD
	Timezone    string       `json:"timezone"`     // Local timezone name
	GeneratedAt string       `json:"generated_at"` // ISO timestamp
	Tools       []ToolReport `json:"tools"`        // Array of tool summaries
}

// ToolReport represents statistics for a single tool on the target day
type ToolReport struct {
	ToolName string     `json:"tool_name"` // Hook-browser tool name (e.g., "claude", "codex", "opencode")
	Details  DayDetails `json:"details"`   // Statistics for the single day
}

// DayDetails contains all statistics for a single natural day
type DayDetails struct {
	SessionStats    SessionStats       `json:"session_stats"`
	EventTypeCounts map[string]int     `json:"event_type_counts"`
	ToolUsageStats  []ToolUsageItem    `json:"tool_usage_stats"`
	PromptStats     PromptStats        `json:"prompt_stats"`
}

// SessionStats contains session-related statistics
type SessionStats struct {
	SessionCount    int           `json:"session_count"`
	EventCount      int           `json:"event_count"`
	UniqueSessions []SessionInfo  `json:"unique_sessions"`
}

// SessionInfo represents a single session's duration information
type SessionInfo struct {
	SessionID      string         `json:"session_id"`
	BeginTime      string         `json:"begin_time"`       // ISO timestamp
	DurationMs     int64          `json:"duration_ms"`      // Duration in milliseconds
	EventTypeCounts map[string]int `json:"event_type_counts"` // Count of each event type in this session
	RepositoryName string         `json:"repository_name"`  // Git repository name (e.g., "owner/repo")
	Branch         string         `json:"branch"`           // Git branch name
	CommitHash     string         `json:"commit_hash"`      // Git commit hash
}

// extractRepositoryName extracts the "owner/repo" short name from the first remote URL.
// Returns empty string if no URLs available or parsing fails.
func extractRepositoryName(remoteURLs []string) string {
	if len(remoteURLs) == 0 {
		return ""
	}
	return hookutils.ExtractRepoShortName(remoteURLs[0])
}

// ToolUsageItem represents a tool used within events
type ToolUsageItem struct {
	ToolName string `json:"tool_name"`
	Count    int    `json:"count"`
}

// PromptStats contains prompt-related statistics
type PromptStats struct {
	TotalPrompts        int     `json:"total_prompts"`
	AveragePromptLength float64 `json:"average_prompt_length"`
	LongestPrompt       int     `json:"longest_prompt"`
	ShortestPrompt      int     `json:"shortest_prompt"`
}

// reportCacheFilePath returns the cache file path for a given date
// Creates the cache directory if it doesn't exist
func reportCacheFilePath(dateStr string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(home, ".code-together", "reports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return filepath.Join(dir, dateStr+".json"), nil
}

// isToday checks if the given date string (YYYY-MM-DD) represents today
func isToday(dateStr string) bool {
	// Parse the date string
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// On parse error, treat as past date to avoid caching issues
		return false
	}

	// Compare with today's date in local timezone
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// loadReportFromCache loads a daily report from the cache file
// Returns error if cache file doesn't exist or is corrupted
func loadReportFromCache(cachePath string) (*DailyReport, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var report DailyReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached report: %w", err)
	}

	return &report, nil
}

// saveReportToCache saves a daily report to the cache file
// Uses atomic write pattern (write to .tmp then rename)
func saveReportToCache(report *DailyReport, cachePath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	tmp := cachePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tmp, cachePath); err != nil {
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// filterTools returns a new DailyReport with only the specified tool
// If toolName is empty, returns the full report
func filterTools(report *DailyReport, toolName string) *DailyReport {
	if toolName == "" {
		return report
	}

	filtered := &DailyReport{
		Date:        report.Date,
		Timezone:    report.Timezone,
		GeneratedAt: report.GeneratedAt,
		Tools:       make([]ToolReport, 0),
	}

	for _, tool := range report.Tools {
		if tool.ToolName == toolName {
			filtered.Tools = append(filtered.Tools, tool)
			break
		}
	}

	return filtered
}

// GenerateDailyReport creates a summary report for a specific date
// For past dates, attempts to load from cache before generating from database.
// Today's reports are always generated fresh from the database.
func (s *ReportService) GenerateDailyReport(dateStr, toolName string) (*DailyReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Step 1: Get cache file path
	cachePath, err := reportCacheFilePath(dateStr)
	if err != nil {
		// If we can't get cache path, continue without caching
		cachePath = ""
	}

	// Step 2: Try loading from cache (only for past dates)
	if cachePath != "" && !isToday(dateStr) {
		if cached, err := loadReportFromCache(cachePath); err == nil {
			// Cache hit - filter and return
			return filterTools(cached, toolName), nil
		}
		// Cache miss or error - continue to generate from DB
	}

	// Step 3: Generate from database (always get ALL tools)
	db := hookdb.HookDB
	if db == nil {
		return nil, fmt.Errorf("hook database not initialized")
	}

	// Use empty string to get all tools, then filter in-memory
	summary, err := storage.GeneratePersonalEventsSummary(db, "", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// Get timezone info
	_, _, tzName, _ := storage.GetDayBoundaries(dateStr)

	// Convert storage types to report types
	tools := make([]ToolReport, len(summary.Tools))
	for i, tool := range summary.Tools {
		// Convert session durations
		sessions := make([]SessionInfo, len(tool.Details.SessionStats.UniqueSessions))
		for j, sess := range tool.Details.SessionStats.UniqueSessions {
			// Extract repository name from raw remote URLs (service layer logic)
			repoName := extractRepositoryName(sess.RemoteURLs)

			sessions[j] = SessionInfo{
				SessionID:      sess.SessionID,
				BeginTime:      sess.BeginTime.Format("2006-01-02T15:04:05-07:00"),
				DurationMs:     sess.Duration,
				EventTypeCounts: sess.EventTypeCounts,
				RepositoryName: repoName,
				Branch:         sess.Branch,
				CommitHash:     sess.CommitHash,
			}
		}

		// Convert tool usage items
		usage := make([]ToolUsageItem, len(tool.Details.ToolUsageStats))
		for k, item := range tool.Details.ToolUsageStats {
			usage[k] = ToolUsageItem{
				ToolName: item.ToolName,
				Count:    item.Count,
			}
		}

		tools[i] = ToolReport{
			ToolName: tool.ToolName,
			Details: DayDetails{
				SessionStats: SessionStats{
					SessionCount:    tool.Details.SessionStats.SessionCount,
					EventCount:      tool.Details.SessionStats.EventCount,
					UniqueSessions: sessions,
				},
				EventTypeCounts: tool.Details.EventTypeCounts,
				ToolUsageStats:  usage,
				PromptStats: PromptStats{
					TotalPrompts:        tool.Details.PromptStats.TotalPrompts,
					AveragePromptLength: tool.Details.PromptStats.AveragePromptLength,
					LongestPrompt:       tool.Details.PromptStats.LongestPrompt,
					ShortestPrompt:      tool.Details.PromptStats.ShortestPrompt,
				},
			},
		}
	}

	report := &DailyReport{
		Date:        summary.Date,
		Timezone:    tzName,
		GeneratedAt: summary.GeneratedAt.Format("2006-01-02T15:04:05-07:00"),
		Tools:       tools,
	}

	// Step 4: Save to cache (only for past dates)
	if cachePath != "" && !isToday(dateStr) {
		// Best effort - ignore save errors since we have the data
		_ = saveReportToCache(report, cachePath)
	}

	// Step 5: Filter tools in-memory if requested
	return filterTools(report, toolName), nil
}
