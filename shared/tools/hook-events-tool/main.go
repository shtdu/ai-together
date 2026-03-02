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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	outputFolder := flag.String("o", ".", "Output folder")
	flag.Parse()

	logPath := filepath.Join(os.Getenv("HOME"), ".claude", "job_done.log")

	events, err := parseLogFile(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing log file: %v\n", err)
		os.Exit(1)
	}

	if len(events) == 0 {
		fmt.Println("No events found in log file")
		return
	}

	sequenceBySession := make(map[string]int)
	var processedCount int

	for _, event := range events {
		sessionID, ok := event["session_id"].(string)
		if !ok {
			continue
		}

		eventName, ok := event["hook_event_name"].(string)
		if !ok {
			continue
		}

		sequenceBySession[sessionID]++
		sequenceID := sequenceBySession[sessionID]

		sessionDir := filepath.Join(*outputFolder, sessionID)
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			continue
		}

		fileName := fmt.Sprintf("%02d_%s.json", sequenceID, eventName)
		filePath := filepath.Join(sessionDir, fileName)
		data, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling event: %v\n", err)
			continue
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			continue
		}

		processedCount++
	}

	fmt.Printf("Processed %d events\n", processedCount)
}

func parseLogFile(filePath string) ([]map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var events []map[string]interface{}
	lines := strings.Split(string(data), "\n")
	var currentObject strings.Builder
	inObject := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if inObject {
				if obj, err := parseObject(currentObject.String()); err == nil {
					events = append(events, obj)
				}
				currentObject.Reset()
				inObject = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "{") {
			if inObject {
				if obj, err := parseObject(currentObject.String()); err == nil {
					events = append(events, obj)
				}
				currentObject.Reset()
			}
			inObject = true
			currentObject.WriteString(line)
			currentObject.WriteString("\n")
		} else if inObject {
			currentObject.WriteString(line)
			currentObject.WriteString("\n")
		}
	}

	if currentObject.Len() > 0 && inObject {
		if obj, err := parseObject(currentObject.String()); err == nil {
			events = append(events, obj)
		}
	}

	return events, nil
}

func parseObject(jsonStr string) (map[string]interface{}, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil, fmt.Errorf("empty json string")
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return obj, nil
}
