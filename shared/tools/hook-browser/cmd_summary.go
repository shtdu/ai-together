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


package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/code-together/shared/hook-common/storage"
)

// generateSummaryReport generates a summary report and prints it as JSON.
// Returns exit code (0 for success, 1 for error).
func generateSummaryReport(db *sql.DB, toolNameFilter, targetDate string) int {
	// Validate date format
	if targetDate == "" {
		fmt.Fprintf(os.Stderr, "Error: -date flag is required (format: YYYY-MM-DD)\n")
		return 1
	}

	// Validate date format by attempting to parse
	_, _, _, err := storage.GetDayBoundaries(targetDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid date format: %v\n", err)
		return 1
	}

	// Generate summary report
	summary, err := storage.GeneratePersonalEventsSummary(db, toolNameFilter, targetDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating summary: %v\n", err)
		return 1
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return 1
	}

	// Output to stdout
	fmt.Println(string(jsonData))
	return 0
}
