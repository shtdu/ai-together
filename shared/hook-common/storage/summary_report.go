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
	"sort"
	"time"
)

// GeneratePersonalEventsSummary generates a summary report for a single natural day.
// Filters at SQL level using timestamp range to avoid loading all events into memory.
// Returns a summary with statistics for each tool on the target day.
func GeneratePersonalEventsSummary(db *sql.DB, toolNameFilter string, targetDate string) (*PersonalEventsSummary, error) {
	// Parse and validate target date
	_, _, tzName, err := GetDayBoundaries(targetDate)
	if err != nil {
		return nil, err
	}

	// Determine which tools to include
	var toolNames []string
	if toolNameFilter != "" {
		toolNames = []string{toolNameFilter}
	} else {
		// Get all tool names
		toolNames, err = GetAllToolNames(db)
		if err != nil {
			return nil, err
		}
	}

	// Generate summaries for each tool
	var toolSummaries []ToolSummary
	for _, toolName := range toolNames {
		// Get all statistics for this tool on the target day
		sessionStats, err := GetDailySessionStatistics(db, toolName, targetDate)
		if err != nil {
			return nil, err
		}

		// Skip tools with no events on the target day
		if sessionStats.EventCount == 0 {
			continue
		}

		eventTypeCounts, err := GetDailyEventTypeCounts(db, toolName, targetDate)
		if err != nil {
			return nil, err
		}

		toolUsageStats, err := GetDailyToolUsageStatistics(db, toolName, targetDate)
		if err != nil {
			return nil, err
		}

		promptStats, err := GetDailyPromptStatistics(db, toolName, targetDate)
		if err != nil {
			return nil, err
		}

		// Combine into ToolSummary
		toolSummaries = append(toolSummaries, ToolSummary{
			ToolName: toolName,
			Details: DayDetails{
				SessionStats:    sessionStats,
				EventTypeCounts: eventTypeCounts,
				ToolUsageStats:  toolUsageStats,
				PromptStats:     promptStats,
			},
		})
	}

	// Sort tools alphabetically
	sort.Slice(toolSummaries, func(i, j int) bool {
		return toolSummaries[i].ToolName < toolSummaries[j].ToolName
	})

	// Create summary
	summary := &PersonalEventsSummary{
		Date:        targetDate,
		Timezone:    tzName,
		GeneratedAt: time.Now(),
		Tools:       toolSummaries,
	}

	return summary, nil
}
