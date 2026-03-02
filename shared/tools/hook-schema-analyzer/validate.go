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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func validateMain() {
	args := os.Args
	logFile := "../../../docs/hooks/job_done.log"
	if len(args) > 2 {
		logFile = args[2]
	}

	schemaDir := "../../events"
	if len(args) > 3 {
		schemaDir = args[3]
	}

	schemas, err := loadSchemas(schemaDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading schemas: %v\n", err)
		os.Exit(1)
	}

	validationResults, err := validateAgainstLogs(logFile, schemas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating against logs: %v\n", err)
		os.Exit(1)
	}

	for _, result := range validationResults {
		fmt.Printf("\n=== %s ===\n", result.EventName)
		fmt.Printf("Total instances: %d\n", result.TotalInstances)
		fmt.Printf("Schema required fields: %d\n", len(result.SchemaRequiredFields))

		if len(result.SchemaRequiredFields) == 0 {
			fmt.Println("No required fields defined in schema")
			continue
		}

		fmt.Printf("\nTruly required fields (present in all instances):\n")
		trulyRequiredCount := 0
		for _, field := range result.SchemaRequiredFields {
			percent := float64(result.FieldPresence[field]) / float64(result.TotalInstances) * 100
			status := "✓"
			if percent < 100 {
				status = "✗"
			} else {
				trulyRequiredCount++
			}
			fmt.Printf("  %s %s: %d/%d (%.1f%%)\n",
				status, field, result.FieldPresence[field], result.TotalInstances, percent)
		}

		fmt.Printf("\nRecommended required fields for schema:\n")
		for _, field := range result.SchemaRequiredFields {
			percent := float64(result.FieldPresence[field]) / float64(result.TotalInstances) * 100
			if percent == 100 {
				fmt.Printf("  - %s\n", field)
			}
		}

		// Show fields with high presence but not in required
		optionalHighPresence := []string{}
		for field, count := range result.FieldPresence {
			percent := float64(count) / float64(result.TotalInstances) * 100
			if percent >= 95 && percent < 100 {
				isRequired := false
				for _, req := range result.SchemaRequiredFields {
					if req == field {
						isRequired = true
						break
					}
				}
				if !isRequired {
					optionalHighPresence = append(optionalHighPresence, field)
				}
			}
		}

		if len(optionalHighPresence) > 0 {
			sort.Strings(optionalHighPresence)
			fmt.Printf("\nOptional fields with >=95%% presence:\n")
			for _, field := range optionalHighPresence {
				fmt.Printf("  - %s: %d/%d\n", field, result.FieldPresence[field], result.TotalInstances)
			}
		}
	}
}

func init() {
	// This is just to make validate.go compile
	// Run with: go run validate.go [log-file] [schema-dir]
}

type ValidationSchema struct {
	Required []string `json:"required"`
}

type ValidationResult struct {
	EventName            string
	TotalInstances       int
	SchemaRequiredFields []string
	FieldPresence        map[string]int
}

func loadSchemas(schemaDir string) (map[string]*ValidationSchema, error) {
	schemas := make(map[string]*ValidationSchema)

	files, err := filepath.Glob(filepath.Join(schemaDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("globbing schema files: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading schema file %s: %w", file, err)
		}

		var schema ValidationSchema
		if err := json.Unmarshal(data, &schema); err != nil {
			return nil, fmt.Errorf("parsing schema file %s: %w", file, err)
		}

		base := filepath.Base(file)
		eventName := strings.TrimSuffix(base, ".json")
		schemas[eventName] = &schema
	}

	return schemas, nil
}

func validateAgainstLogs(logFile string, schemas map[string]*ValidationSchema) ([]*ValidationResult, error) {
	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("reading log file: %w", err)
	}

	results := make(map[string]*ValidationResult)
	for eventName := range schemas {
		results[eventName] = &ValidationResult{
			EventName:     eventName,
			FieldPresence: make(map[string]int),
		}
		if schema, ok := schemas[eventName]; ok {
			results[eventName].SchemaRequiredFields = schema.Required
		}
	}

	lines := strings.Split(string(data), "\n")
	var currentObject strings.Builder
	inObject := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if inObject {
				processValidationObject(currentObject.String(), results)
				currentObject.Reset()
				inObject = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "{") {
			if inObject {
				processValidationObject(currentObject.String(), results)
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
		processValidationObject(currentObject.String(), results)
	}

	// Convert to sorted slice
	eventNames := make([]string, 0, len(results))
	for name := range results {
		eventNames = append(eventNames, name)
	}
	sort.Strings(eventNames)

	sortedResults := make([]*ValidationResult, 0, len(results))
	for _, name := range eventNames {
		sortedResults = append(sortedResults, results[name])
	}

	return sortedResults, nil
}

func processValidationObject(jsonStr string, results map[string]*ValidationResult) {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return
	}

	eventName, ok := obj["hook_event_name"].(string)
	if !ok {
		return
	}

	result, ok := results[eventName]
	if !ok {
		return
	}

	result.TotalInstances++

	for key := range obj {
		result.FieldPresence[key]++
	}
}
