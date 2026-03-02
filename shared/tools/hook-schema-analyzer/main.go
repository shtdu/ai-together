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

type FieldType string

const (
	TypeString  FieldType = "string"
	TypeNumber  FieldType = "number"
	TypeBoolean FieldType = "boolean"
	TypeObject  FieldType = "object"
	TypeArray   FieldType = "array"
	TypeNull    FieldType = "null"
	TypeAny     FieldType = "any"
)

type FieldInfo struct {
	Types     map[FieldType]int
	Required  int
	Optional  int
	Instances int
}

type EventSchema struct {
	EventName  string
	Fields     map[string]*FieldInfo
	TotalCount int
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "validate" {
		validateMain()
		return
	}

	logFile := "docs/hooks/job_done.1000.log"
	if len(os.Args) > 1 {
		logFile = os.Args[1]
	}

	events, err := parseLogFile(logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing log file: %v\n", err)
		os.Exit(1)
	}

	if len(events) == 0 {
		fmt.Println("No events found in log file")
		return
	}

	outputDir := "../../events"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	for eventName, schema := range events {
		fmt.Printf("Processing event: %s\n", eventName)

		goFile := filepath.Join(outputDir, fmt.Sprintf("%s.go", eventName))
		jsonFile := filepath.Join(outputDir, fmt.Sprintf("%s.json", eventName))

		if err := generateGoStruct(schema, goFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating Go struct for %s: %v\n", eventName, err)
		}

		if err := generateJSONSchema(schema, jsonFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON schema for %s: %v\n", eventName, err)
		}
	}

	fmt.Printf("\nGenerated schemas for %d event types in %s/\n", len(events), outputDir)
}

func parseLogFile(filePath string) (map[string]*EventSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	events := make(map[string]*EventSchema)

	lines := strings.Split(string(data), "\n")
	var currentObject strings.Builder
	inObject := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if inObject {
				if err := processObject(currentObject.String(), events); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
				}
				currentObject.Reset()
				inObject = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "{") {
			if inObject {
				if err := processObject(currentObject.String(), events); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
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
		if err := processObject(currentObject.String(), events); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}

	return events, nil
}

func processObject(jsonStr string, events map[string]*EventSchema) error {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	eventName, ok := obj["hook_event_name"].(string)
	if !ok {
		return fmt.Errorf("missing hook_event_name")
	}

	schema, exists := events[eventName]
	if !exists {
		schema = &EventSchema{
			EventName: eventName,
			Fields:    make(map[string]*FieldInfo),
		}
		events[eventName] = schema
	}

	schema.TotalCount++
	analyzeObject(obj, schema, schema.Fields)
	return nil
}

func analyzeObject(obj map[string]interface{}, schema *EventSchema, fields map[string]*FieldInfo) {
	for key, value := range obj {
		if key == "hook_event_name" {
			continue
		}

		fieldInfo, exists := fields[key]
		if !exists {
			fieldInfo = &FieldInfo{
				Types:     make(map[FieldType]int),
				Instances: 0,
			}
			fields[key] = fieldInfo
		}

		fieldInfo.Instances++
		fieldType := inferType(value)
		fieldInfo.Types[fieldType]++

		// Don't recursively analyze nested objects - keep them as map[string]interface{}
	}
}

func inferType(value interface{}) FieldType {
	if value == nil {
		return TypeNull
	}

	switch value.(type) {
	case string:
		return TypeString
	case float64, float32, int, int64, int32, uint, uint64, uint32:
		return TypeNumber
	case bool:
		return TypeBoolean
	case map[string]interface{}:
		return TypeObject
	case []interface{}:
		return TypeArray
	default:
		return TypeAny
	}
}

func getGoType(fieldInfo *FieldInfo) string {
	types := fieldInfo.Types

	if len(types) == 1 {
		for t := range types {
			switch t {
			case TypeString:
				return "string"
			case TypeNumber:
				return "float64"
			case TypeBoolean:
				return "bool"
			case TypeObject:
				return "map[string]interface{}"
			case TypeArray:
				return "[]interface{}"
			case TypeNull:
				return "interface{}"
			default:
				return "interface{}"
			}
		}
	}

	if len(types) == 2 {
		if _, hasNull := types[TypeNull]; hasNull {
			for t := range types {
				if t != TypeNull {
					switch t {
					case TypeString:
						return "string"
					case TypeNumber:
						return "float64"
					case TypeBoolean:
						return "bool"
					case TypeObject:
						return "map[string]interface{}"
					case TypeArray:
						return "[]interface{}"
					default:
						return "interface{}"
					}
				}
			}
		}
	}

	return "interface{}"
}

func getJSONType(fieldInfo *FieldInfo) string {
	types := fieldInfo.Types

	if len(types) == 1 {
		for t := range types {
			switch t {
			case TypeString:
				return "string"
			case TypeNumber:
				return "number"
			case TypeBoolean:
				return "boolean"
			case TypeObject:
				return "object"
			case TypeArray:
				return "array"
			default:
				return "null"
			}
		}
	}

	if len(types) == 2 {
		if _, hasNull := types[TypeNull]; hasNull {
			for t := range types {
				if t != TypeNull {
					return string(t)
				}
			}
		}
	}

	return "string"
}

var fieldDescriptions = map[string]string{
	"session_id":            "Unique identifier for session",
	"transcript_path":       "Path to the transcript file",
	"cwd":                   "Current working directory",
	"permission_mode":       "Permission mode (e.g., 'default', 'acceptEdits')",
	"tool_name":             "Name of the tool being called",
	"tool_use_id":           "Unique identifier for this tool use",
	"tool_input":            "Input parameters for the tool",
	"command":               "Command to execute (for Bash tool)",
	"description":           "Description of what the tool does",
	"file_path":             "Path to the file being operated on",
	"content":               "Content to write to file",
	"pattern":               "Search pattern",
	"replace_all":           "Whether to replace all matches",
	"old_string":            "String to replace",
	"new_string":            "Replacement string",
	"offset":                "Offset in the file (for Edit)",
	"limit":                 "Limit for operations",
	"head_limit":            "Limit for head operation",
	"glob":                  "File pattern to match",
	"prompt":                "User prompt submitted",
	"block":                 "Whether to block the tool execution",
	"run_in_background":     "Whether to run the command in background",
	"skill":                 "Name of the skill being used",
	"task_id":               "Task identifier",
	"path":                  "Path parameter",
	"output_mode":           "Output mode",
	"type":                  "Type of the tool or operation",
	"plan":                  "Plan for the operation",
	"query":                 "Query string",
	"todos":                 "List of todos",
	"shell_id":              "Shell identifier (for Bash)",
	"subagent_type":         "Type of the subagent",
	"timeout":               "Timeout for the operation",
	"agent_id":              "Unique identifier for the subagent",
	"agent_transcript_path": "Path to the subagent transcript file",
	"stop_hook_active":      "Whether the stop hook is active",
	"reason":                "Reason for session end",
	"-i":                    "Case-insensitive search flag",
	"-A":                    "Number of lines after match flag",
	"-B":                    "Number of lines before match flag",
	"-C":                    "Number of lines around match flag",
	"-n":                    "Show line numbers flag",
}

func getEventDescription(eventName string) string {
	descriptions := map[string]string{
		"PreToolUse":       "Runs before tool calls (can block them)",
		"PostToolUse":      "Runs after tool calls complete",
		"Stop":             "Runs when Claude Code finishes responding",
		"SubagentStop":     "Runs when subagent tasks complete",
		"SessionEnd":       "Runs when Claude Code session ends",
		"UserPromptSubmit": "Runs when user submits a prompt, before Claude processes it",
		"PreCompact":       "Runs before Claude Code is about to run a compact operation",
	}
	return descriptions[eventName]
}

func toPascalCase(fieldName string) string {
	if fieldName == "" {
		return fieldName
	}

	fieldNameMap := map[string]string{
		"-i":   "CaseInsensitive",
		"-A":   "AfterContext",
		"-B":   "BeforeContext",
		"-C":   "ContextLines",
		"-n":   "ShowLineNumbers",
		"type": "TypeField",
	}

	if mappedName, ok := fieldNameMap[fieldName]; ok {
		return mappedName
	}

	parts := strings.Split(fieldName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			runes := []rune(part)
			if runes[0] >= 'a' && runes[0] <= 'z' {
				runes[0] = runes[0] - 32
			}
			parts[i] = string(runes)
		}
	}

	return strings.Join(parts, "")
}

func generateGoStruct(schema *EventSchema, filePath string) error {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("package events\n\n"))
	description := getEventDescription(schema.EventName)
	if description != "" {
		builder.WriteString(fmt.Sprintf("// %s\n", description))
	}
	builder.WriteString(fmt.Sprintf("type %s struct {\n", schema.EventName))

	usedGoNames := make(map[string]bool)

	fieldNames := make([]string, 0, len(schema.Fields))
	for fieldName := range schema.Fields {
		fieldNames = append(fieldNames, fieldName)
	}

	sort.Strings(fieldNames)

	for _, fieldName := range fieldNames {
		fieldInfo := schema.Fields[fieldName]
		goFieldName := toPascalCase(fieldName)

		// Handle duplicate Go field names by adding suffix
		suffix := 0
		baseName := goFieldName
		for usedGoNames[goFieldName] {
			suffix++
			goFieldName = fmt.Sprintf("%s%d", baseName, suffix)
		}
		usedGoNames[goFieldName] = true

		jsonTag := fieldName

		goType := getGoType(fieldInfo)
		comment := ""
		if desc, ok := fieldDescriptions[fieldName]; ok {
			comment = fmt.Sprintf("// %s\n\t", desc)
		}
		line := fmt.Sprintf("\t%s%s %s `json:\"%s\"`", comment, goFieldName, goType, jsonTag)
		builder.WriteString(line + "\n")
	}

	builder.WriteString("}\n")

	content := builder.String()
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

func generateJSONSchema(schema *EventSchema, filePath string) error {
	eventDesc := getEventDescription(schema.EventName)
	jsonSchema := map[string]interface{}{
		"$schema":    "https://json-schema.org/draft/2020-12/schema",
		"title":      schema.EventName,
		"type":       "object",
		"required":   []string{},
		"properties": map[string]interface{}{},
	}

	if eventDesc != "" {
		jsonSchema["description"] = eventDesc
	}

	properties := jsonSchema["properties"].(map[string]interface{})
	requiredSet := make(map[string]bool)

	fieldNames := make([]string, 0, len(schema.Fields))
	for fieldName := range schema.Fields {
		fieldNames = append(fieldNames, fieldName)
	}

	sort.Strings(fieldNames)

	for _, fieldName := range fieldNames {
		fieldInfo := schema.Fields[fieldName]

		// Only mark as required if field is present in all instances
		if fieldInfo.Instances > 0 && fieldInfo.Instances == schema.TotalCount {
			requiredSet[fieldName] = true
		}
		jsonType := getJSONType(fieldInfo)

		propSchema := map[string]interface{}{
			"type": jsonType,
		}

		if desc, ok := fieldDescriptions[fieldName]; ok {
			propSchema["description"] = desc
		}

		if len(fieldInfo.Types) > 1 {
			if _, hasNull := fieldInfo.Types[TypeNull]; hasNull {
				types := []string{}
				for t := range fieldInfo.Types {
					if t == TypeString {
						types = append(types, "string")
					} else if t == TypeNumber {
						types = append(types, "number")
					} else if t == TypeBoolean {
						types = append(types, "boolean")
					} else if t == TypeObject {
						types = append(types, "object")
					} else if t == TypeArray {
						types = append(types, "array")
					}
				}
				if len(types) > 0 {
					propSchema["type"] = types
				}
			}
		}

		properties[fieldName] = propSchema
	}

	// Build sorted required list from set
	var required []string
	for name := range requiredSet {
		required = append(required, name)
	}
	sort.Strings(required)

	jsonSchema["required"] = required

	data, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON schema: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}
