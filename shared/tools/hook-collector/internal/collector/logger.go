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


package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// logFile is the path to the collector log file
// It's initialized lazily to handle errors gracefully
var logFile string

func init() {
	if homeDir, err := os.UserHomeDir(); err == nil {
		logFile = filepath.Join(homeDir, ".code-together", "collector.log")
	} else {
		// Fallback to current directory if home dir cannot be determined
		logFile = "collector.log"
	}
}

// LogError logs an error with payload to the collector log file
func LogError(originalErr error, payload []byte) {
	if originalErr == nil {
		return
	}

	// Ensure log directory exists
	logDir := filepath.Dir(logFile)
	if mkdirErr := os.MkdirAll(logDir, 0755); mkdirErr != nil {
		// If we can't create the log directory, silently fail
		// We don't want to cause additional errors
		return
	}

	// Open log file in append mode, create if doesn't exist
	f, openErr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		// Can't open log file, silently fail
		return
	}
	defer f.Close()

	// Format log entry
	timestamp := time.Now().UTC().Format(time.RFC3339)
	// if payload is JSON, pretty print it
	payloadStr := string(payload)
	if json.Valid(payload) {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, payload, "", "  "); err == nil {
			payloadStr = pretty.String()
		}
	}
	if payloadStr == "" {
		payloadStr = "<empty or nil payload>"
	}

	logEntry := fmt.Sprintf("\n[%s] ERROR: %s\nPayload:\n%s\n\n",
		timestamp, originalErr.Error(), payloadStr)

	// Write to log file
	if _, writeErr := f.WriteString(logEntry); writeErr != nil {
		// Failed to write, silently fail
		return
	}
}

// SetLogFile sets a custom log file path (useful for testing)
func SetLogFile(path string) {
	logFile = path
}
