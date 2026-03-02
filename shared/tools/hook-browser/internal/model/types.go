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
	"time"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeEvents Mode = iota
	ModeSessionPicker
)

// EventItem represents an event in the list
type EventItem struct {
	Seq   uint64
	Name  string
	Index int
}

// FormatEventListItem formats an event item for display in the list
func FormatEventListItem(e EventItem) string {
	return fmt.Sprintf("%d - %s", e.Seq, e.Name)
}

// SessionItem represents a session with metadata
type SessionItem struct {
	ID             string
	Count          int
	Index          int
	FirstEventTime string
}

// FormatSessionListItem formats a session item for display in the list
// Returns a single-line format: "abc12345  2h ago  42 events"
func FormatSessionListItem(s SessionItem) string {
	// Extract short ID (first 8 chars of UUID)
	shortID := s.ID
	// if len(s.ID) > 8 {
	// shortID = s.ID[:8]
	// }

	// Format timestamp if available
	timeStr := ""
	if s.FirstEventTime != "" {
		if t := parseTimestamp(s.FirstEventTime); !t.IsZero() {
			// Show relative time for recent sessions, absolute for older ones
			duration := time.Since(t)
			if duration < 24*time.Hour {
				timeStr = formatDuration(duration)
			} else if duration < 7*24*time.Hour {
				timeStr = t.Format("Mon 15:04")
			} else {
				timeStr = t.Format("2006-01-02")
			}
		}
	}

	// Build title with time and count
	if timeStr != "" {
		return fmt.Sprintf("%s  %s  %d events", shortID, timeStr, s.Count)
	}
	return fmt.Sprintf("%s  %d events", shortID, s.Count)
}

// formatDuration creates a human-readable relative time string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// TruncateString truncates a string to a maximum length, adding "..." if truncated
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	if len(s) <= maxLen {
		return s
	}

	// Handle multi-byte characters properly
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	// Need space for "..." so truncate early
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}

	// Ensure we don't slice beyond bounds
	truncateAt := maxLen - 3
	if truncateAt > len(runes) {
		truncateAt = len(runes)
	}
	if truncateAt < 0 {
		truncateAt = 0
	}

	return string(runes[:truncateAt]) + "..."
}
