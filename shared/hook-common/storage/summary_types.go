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

import "time"

// PersonalEventsSummary represents a summary report for a single natural day.
// It contains aggregate statistics broken down by tool (claude-code, codex, etc.).
type PersonalEventsSummary struct {
	Date        string       `json:"date"`         // YYYY-MM-DD in local timezone (single day)
	Timezone    string       `json:"timezone"`     // Local timezone used (e.g., "America/New_York" or "EST")
	GeneratedAt time.Time    `json:"generated_at"` // When the report was generated
	Tools       []ToolSummary `json:"tools"`        // Array of tool summaries, one per tool
}

// ToolSummary represents statistics for a single tool on the target day.
// ToolName refers to hook-browser tools (claude-code, codex, etc.).
type ToolSummary struct {
	ToolName string     `json:"tool_name"` // Hook-browser tool name (e.g., "claude-code", "codex")
	Details  DayDetails `json:"details"`   // Statistics for the single day
}

// DayDetails contains all statistics for a single natural day.
type DayDetails struct {
	SessionStats    DailySessionStatistics `json:"session_stats"`    // Session information for the day
	EventTypeCounts map[string]int         `json:"event_type_counts"` // Count of each event type
	ToolUsageStats  []ToolUsageItem        `json:"tool_usage_stats"`  // Tools used within events (Read, Write, Glob, etc.)
	PromptStats     DailyPromptStatistics  `json:"prompt_stats"`      // Prompt statistics for the day
}

// SessionDuration represents a single session's duration information.
type SessionDuration struct {
	SessionID       string         `json:"session_id"`       // Session identifier
	BeginTime       time.Time      `json:"begin_time"`       // First event timestamp
	LastTime        time.Time      `json:"last_time"`        // Last event timestamp
	Duration        int64          `json:"duration_ms"`      // Duration in milliseconds (from first to last event)
	EventTypeCounts map[string]int `json:"event_type_counts"` // Count of each event type in this session
	RemoteURLs      []string       `json:"remote_urls"`      // Git remote URLs (raw)
	Branch          string         `json:"branch"`           // Git branch name
	CommitHash      string         `json:"commit_hash"`      // Git commit hash
}

// DailySessionStatistics contains session-related statistics for a single day.
type DailySessionStatistics struct {
	SessionCount    int                `json:"session_count"`    // Number of sessions active on this day
	EventCount      int                `json:"event_count"`      // Total number of events on this day
	UniqueSessions []SessionDuration  `json:"unique_sessions"`  // Session duration details for each session
}

// ToolUsageItem represents a tool used within events (extracted from PreToolUse/PostToolUse).
// ToolName refers to function tools called by the AI (Read, Write, Glob, etc.).
// This is different from ToolSummary.ToolName which refers to hook-browser tools.
type ToolUsageItem struct {
	ToolName string `json:"tool_name"` // Name of the tool used within events (e.g., "Read", "Write", "Glob")
	Count    int    `json:"count"`     // Number of times this tool was called on the target day
}

// DailyPromptStatistics contains prompt-related statistics for a single day.
type DailyPromptStatistics struct {
	TotalPrompts        int     `json:"total_prompts"`         // Total number of prompts submitted
	AveragePromptLength float64 `json:"average_prompt_length"` // Average character count of prompts
	LongestPrompt       int     `json:"longest_prompt"`        // Character count of longest prompt
	ShortestPrompt      int     `json:"shortest_prompt"`       // Character count of shortest prompt
}
