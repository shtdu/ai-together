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


package commands

import (
	"compress/gzip"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/code-together/shared/hook-common/storage"
)

// ExportFilters holds filter criteria for export
type ExportFilters struct {
	Since     time.Time
	Until     time.Time
	SessionID string
	EventType string
}

// ExportFormat defines the output format
type ExportFormat string

const (
	FormatJSON  ExportFormat = "json"
	FormatJSONL ExportFormat = "jsonl"
	FormatCSV   ExportFormat = "csv"
)

// ExportEvents exports events from the database to a file
func ExportEvents(db *sql.DB, toolName string, filters ExportFilters, format ExportFormat, outputPath string, compress bool) error {
	// Open output file
	var out io.Writer
	var file *os.File
	var err error

	if outputPath == "-" || outputPath == "" {
		// Write to stdout
		out = os.Stdout
	} else {
		// Write to file
		if compress && strings.HasSuffix(outputPath, ".gz") {
			file, err = os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			gzWriter := gzip.NewWriter(file)
			defer gzWriter.Close()
			out = gzWriter
		} else {
			file, err = os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			out = file
		}
	}

	// Export based on format
	switch format {
	case FormatJSON:
		if err := exportJSON(db, toolName, filters, out); err != nil {
			return err
		}
	case FormatJSONL:
		if err := exportJSONL(db, toolName, filters, out); err != nil {
			return err
		}
	case FormatCSV:
		if err := exportCSV(db, toolName, filters, out); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// exportJSON exports events as a JSON array
func exportJSON(db *sql.DB, toolName string, filters ExportFilters, out io.Writer) error {
	var events []map[string]interface{}

	// Get all sessions
	sessions, err := storage.GetAllSessionIDs(db, toolName)
	if err != nil {
		return err
	}

	// Iterate through sessions
	for _, sessionID := range sessions {
		// Skip if session filter is set and doesn't match
		if filters.SessionID != "" && sessionID != filters.SessionID {
			continue
		}

		// Get events for this session
		sessionEvents, err := storage.GetSessionEvents(db, sessionID, toolName)
		if err != nil {
			continue
		}

		// Process events
		for _, event := range sessionEvents {
			// Parse event to check filters
			var eventData map[string]interface{}
			if err := json.Unmarshal(event.JSONData, &eventData); err != nil {
				continue // Skip malformed events
			}

			// Apply filters
			if !matchesFilters(eventData, sessionID, event.JSONData, filters) {
				continue
			}

			events = append(events, eventData)
		}
	}

	// Write JSON array
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(events)
}

// exportJSONL exports events as JSONL (one JSON object per line)
func exportJSONL(db *sql.DB, toolName string, filters ExportFilters, out io.Writer) error {
	// Get all sessions
	sessions, err := storage.GetAllSessionIDs(db, toolName)
	if err != nil {
		return err
	}

	// Iterate through sessions
	for _, sessionID := range sessions {
		// Skip if session filter is set and doesn't match
		if filters.SessionID != "" && sessionID != filters.SessionID {
			continue
		}

		// Get events for this session
		sessionEvents, err := storage.GetSessionEvents(db, sessionID, toolName)
		if err != nil {
			continue
		}

		// Process events
		for _, event := range sessionEvents {
			// Parse event to check filters
			var eventData map[string]interface{}
			if err := json.Unmarshal(event.JSONData, &eventData); err != nil {
				continue // Skip malformed events
			}

			// Apply filters
			if !matchesFilters(eventData, sessionID, event.JSONData, filters) {
				continue
			}

			// Write event as JSON line
			_, err := out.Write(event.JSONData)
			if err != nil {
				return err
			}
			_, err = out.Write([]byte("\n"))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// exportCSV exports events as CSV
func exportCSV(db *sql.DB, toolName string, filters ExportFilters, out io.Writer) error {
	writer := csv.NewWriter(out)
	defer writer.Flush()

	// Collect all unique keys across all events for CSV headers
	keysMap := make(map[string]bool)
	var rows [][]string

	// Get all sessions
	sessions, err := storage.GetAllSessionIDs(db, toolName)
	if err != nil {
		return err
	}

	// First pass: collect all unique keys
	for _, sessionID := range sessions {
		if filters.SessionID != "" && sessionID != filters.SessionID {
			continue
		}

		sessionEvents, err := storage.GetSessionEvents(db, sessionID, toolName)
		if err != nil {
			continue
		}

		for _, event := range sessionEvents {
			var eventData map[string]interface{}
			if err := json.Unmarshal(event.JSONData, &eventData); err != nil {
				continue
			}

			if !matchesFilters(eventData, sessionID, event.JSONData, filters) {
				continue
			}

			for k := range eventData {
				keysMap[k] = true
			}
		}
	}

	// Second pass: build rows
	for _, sessionID := range sessions {
		if filters.SessionID != "" && sessionID != filters.SessionID {
			continue
		}

		sessionEvents, err := storage.GetSessionEvents(db, sessionID, toolName)
		if err != nil {
			continue
		}

		for _, event := range sessionEvents {
			var eventData map[string]interface{}
			if err := json.Unmarshal(event.JSONData, &eventData); err != nil {
				continue
			}

			if !matchesFilters(eventData, sessionID, event.JSONData, filters) {
				continue
			}

			// Convert event to CSV row
			row := make([]string, len(keysMap))
			for i, key := range getSortedKeys(keysMap) {
				val := eventData[key]
				row[i] = fmt.Sprintf("%v", val)
			}
			rows = append(rows, row)
		}
	}

	// Write header
	keys := getSortedKeys(keysMap)
	if err := writer.Write(keys); err != nil {
		return err
	}

	// Write rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// matchesFilters checks if an event matches the export filters
func matchesFilters(event map[string]interface{}, sessionID string, eventJSON []byte, filters ExportFilters) bool {
	// Filter by time
	if !filters.Since.IsZero() || !filters.Until.IsZero() {
		tsStr, ok := event["ts"].(string)
		if !ok {
			return false
		}
		ts, err := time.Parse(time.RFC3339Nano, tsStr)
		if err != nil {
			return false
		}
		if !filters.Since.IsZero() && ts.Before(filters.Since) {
			return false
		}
		if !filters.Until.IsZero() && ts.After(filters.Until) {
			return false
		}
	}

	// Filter by event type
	if filters.EventType != "" {
		eventType, ok := event["hook_event_name"].(string)
		if !ok || eventType != filters.EventType {
			return false
		}
	}

	return true
}

// getSortedKeys returns sorted keys from a map
func getSortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple sort (for production use, sort.Sort)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

// parseTimeFilter parses a time filter string (e.g., "7d", "2024-01-01")
func parseTimeFilter(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try duration format (e.g., "7d", "30d")
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		var d int
		if _, err := fmt.Sscanf(days, "%d", &d); err == nil {
			return time.Now().AddDate(0, 0, -d), nil
		}
	}

	// Try ISO date format
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", s)
}

// RunExportCommand executes the export command
func RunExportCommand(dbPath, toolName, outputFile, format, since, until, sessionID, eventType string, compress bool) {
	// Parse time filters
	sinceTime, err := parseTimeFilter(since)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid since time '%s': %v\n", since, err)
		os.Exit(1)
	}

	untilTime, err := parseTimeFilter(until)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid until time '%s': %v\n", until, err)
		os.Exit(1)
	}

	// Validate format
	var exportFormat ExportFormat
	switch strings.ToLower(format) {
	case "json":
		exportFormat = FormatJSON
	case "jsonl":
		exportFormat = FormatJSONL
	case "csv":
		exportFormat = FormatCSV
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s'\n", format)
		os.Exit(1)
	}

	// Open database in read-only mode for export
	db, err := storage.OpenDB(dbPath, true, toolName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create filters
	filters := ExportFilters{
		Since:     sinceTime,
		Until:     untilTime,
		SessionID: sessionID,
		EventType: eventType,
	}

	// Export events
	if err := ExportEvents(db, toolName, filters, exportFormat, outputFile, compress); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to export events: %v\n", err)
		os.Exit(1)
	}

	// Only print success message if not writing to stdout
	if outputFile != "" {
		fmt.Printf("Export complete: %s (format: %s)\n", outputFile, exportFormat)
	}
}
