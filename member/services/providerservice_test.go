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

package services

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
)

// setupProviderTestEnv creates a temporary directory and sets HOME to it.
// This ensures that providerFilePath() creates files in the temp directory
// instead of in the user's actual ~/.code-together folder.
func setupProviderTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})
	return tmpDir
}

// Test helper functions for wildcard matching (local copies since shared package functions are not exported)
func matchWildcard(pattern, text string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == text
	}
	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		prefix, suffix := parts[0], parts[1]
		return strings.HasPrefix(text, prefix) && strings.HasSuffix(text, suffix)
	}
	return false
}

func applyWildcardMapping(pattern, replacement, input string) string {
	if !strings.Contains(pattern, "*") || !strings.Contains(replacement, "*") {
		return replacement
	}
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return replacement
	}
	prefix, suffix := parts[0], parts[1]
	if !strings.HasPrefix(input, prefix) || !strings.HasSuffix(input, suffix) {
		return replacement
	}
	wildcardPart := input[len(prefix) : len(input)-len(suffix)]
	return strings.Replace(replacement, "*", wildcardPart, 1)
}

// ==================== Wildcard Matching Tests ====================

func TestMatchWildcard(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected bool
	}{
		// Exact match
		{
			name:     "Exact match - success",
			pattern:  "claude-sonnet-4",
			text:     "claude-sonnet-4",
			expected: true,
		},
		{
			name:     "Exact match - failure",
			pattern:  "claude-sonnet-4",
			text:     "claude-opus-4",
			expected: false,
		},

		// Prefix wildcard
		{
			name:     "Prefix wildcard - success",
			pattern:  "claude-*",
			text:     "claude-sonnet-4",
			expected: true,
		},
		{
			name:     "Prefix wildcard - multi-segment match",
			pattern:  "claude-*",
			text:     "claude-sonnet-4-latest",
			expected: true,
		},
		{
			name:     "Prefix wildcard - failure",
			pattern:  "claude-*",
			text:     "gpt-4",
			expected: false,
		},

		// Suffix wildcard
		{
			name:     "Suffix wildcard - success",
			pattern:  "*-4",
			text:     "claude-sonnet-4",
			expected: true,
		},
		{
			name:     "Suffix wildcard - failure",
			pattern:  "*-4",
			text:     "claude-sonnet-3.5",
			expected: false,
		},

		// Middle wildcard
		{
			name:     "Middle wildcard - success",
			pattern:  "claude-*-4",
			text:     "claude-sonnet-4",
			expected: true,
		},
		{
			name:     "Middle wildcard - multi-segment match",
			pattern:  "claude-*-4",
			text:     "claude-opus-mini-4",
			expected: true,
		},
		{
			name:     "Middle wildcard - failure prefix",
			pattern:  "claude-*-4",
			text:     "gpt-sonnet-4",
			expected: false,
		},
		{
			name:     "Middle wildcard - failure suffix",
			pattern:  "claude-*-4",
			text:     "claude-sonnet-3",
			expected: false,
		},

		// Edge cases
		{
			name:     "Empty prefix",
			pattern:  "*-sonnet",
			text:     "claude-sonnet",
			expected: true,
		},
		{
			name:     "Empty suffix",
			pattern:  "claude-*",
			text:     "claude-",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchWildcard(tt.pattern, tt.text)
			if result != tt.expected {
				t.Errorf("matchWildcard(%q, %q) = %v, expected %v",
					tt.pattern, tt.text, result, tt.expected)
			}
		})
	}
}

// ==================== Wildcard Mapping Application Tests ====================

func TestApplyWildcardMapping(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		replacement string
		input       string
		expected    string
	}{
		// Prefix wildcard mapping
		{
			name:        "Prefix wildcard mapping",
			pattern:     "claude-*",
			replacement: "anthropic/claude-*",
			input:       "claude-sonnet-4",
			expected:    "anthropic/claude-sonnet-4",
		},
		{
			name:        "Prefix wildcard mapping - multi-segment",
			pattern:     "claude-*",
			replacement: "anthropic/claude-*",
			input:       "claude-opus-4-latest",
			expected:    "anthropic/claude-opus-4-latest",
		},

		// Middle wildcard mapping
		{
			name:        "Middle wildcard mapping",
			pattern:     "claude-*-4",
			replacement: "anthropic/claude-*-v4",
			input:       "claude-sonnet-4",
			expected:    "anthropic/claude-sonnet-v4",
		},

		// No wildcard (return replacement directly)
		{
			name:        "No wildcard - pattern",
			pattern:     "claude-sonnet-4",
			replacement: "anthropic/claude-sonnet-4",
			input:       "claude-sonnet-4",
			expected:    "anthropic/claude-sonnet-4",
		},
		{
			name:        "No wildcard - replacement",
			pattern:     "claude-*",
			replacement: "fixed-model",
			input:       "claude-sonnet-4",
			expected:    "fixed-model",
		},

		// Edge cases
		{
			name:        "Empty match part",
			pattern:     "claude-*",
			replacement: "anthropic/claude-*",
			input:       "claude-",
			expected:    "anthropic/claude-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyWildcardMapping(tt.pattern, tt.replacement, tt.input)
			if result != tt.expected {
				t.Errorf("applyWildcardMapping(%q, %q, %q) = %q, expected %q",
					tt.pattern, tt.replacement, tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== IsModelSupported Tests ====================

func TestProvider_IsModelSupported(t *testing.T) {
	tests := []struct {
		name      string
		provider  Provider
		modelName string
		expected  bool
	}{
		// Backward compatibility: No whitelist and mapping configured
		{
			name:      "Backward compatible - not configured",
			provider:  Provider{},
			modelName: "any-model",
			expected:  true,
		},

		// Scenario A: Native support (exact match)
		{
			name: "Native support - exact match - success",
			provider: Provider{
				SupportedModels: []string{
					"claude-sonnet-4",
					"claude-opus-4",
				},
			},
			modelName: "claude-sonnet-4",
			expected:  true,
		},
		{
			name: "Native support - exact match - failure",
			provider: Provider{
				SupportedModels: []string{
					"claude-sonnet-4",
				},
			},
			modelName: "gpt-4",
			expected:  false,
		},

		// Scenario A+: Native support (wildcard match)
		{
			name: "Native support - wildcard match - success",
			provider: Provider{
				SupportedModels: []string{
					"claude-*",
				},
			},
			modelName: "claude-sonnet-4",
			expected:  true,
		},
		{
			name: "Native support - wildcard match - failure",
			provider: Provider{
				SupportedModels: []string{
					"claude-*",
				},
			},
			modelName: "gpt-4",
			expected:  false,
		},

		// Scenario B: Mapping support (exact match)
		{
			name: "Mapping support - exact match - success",
			provider: Provider{
				SupportedModels: []string{
					"anthropic/claude-sonnet-4",
				},
				ModelMapping: map[string]string{
					"claude-sonnet-4": "anthropic/claude-sonnet-4",
				},
			},
			modelName: "claude-sonnet-4",
			expected:  true,
		},

		// Scenario B+: Mapping support (wildcard match)
		{
			name: "Mapping support - wildcard match - success",
			provider: Provider{
				SupportedModels: []string{
					"anthropic/claude-*",
				},
				ModelMapping: map[string]string{
					"claude-*": "anthropic/claude-*",
				},
			},
			modelName: "claude-sonnet-4",
			expected:  true,
		},

		// Mixed mode
		{
			name: "Mixed mode - native + mapping",
			provider: Provider{
				SupportedModels: []string{
					"native-model",
					"vendor/external",
				},
				ModelMapping: map[string]string{
					"external": "vendor/external",
				},
			},
			modelName: "external",
			expected:  true,
		},
		{
			name: "Mixed mode - native only",
			provider: Provider{
				SupportedModels: []string{
					"native-model",
				},
				ModelMapping: map[string]string{
					"external": "vendor/external",
				},
			},
			modelName: "native-model",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.IsModelSupported(tt.modelName)
			if result != tt.expected {
				t.Errorf("IsModelSupported(%q) = %v, expected %v",
					tt.modelName, result, tt.expected)
			}
		})
	}
}

// ==================== GetEffectiveModel Tests ====================

func TestProvider_GetEffectiveModel(t *testing.T) {
	tests := []struct {
		name           string
		provider       Provider
		requestedModel string
		expected       string
	}{
		// No mapping
		{
			name:           "No mapping - return original name",
			provider:       Provider{},
			requestedModel: "claude-sonnet-4",
			expected:       "claude-sonnet-4",
		},

		// Exact mapping
		{
			name: "Exact mapping - success",
			provider: Provider{
				ModelMapping: map[string]string{
					"claude-sonnet-4": "anthropic/claude-sonnet-4",
				},
			},
			requestedModel: "claude-sonnet-4",
			expected:       "anthropic/claude-sonnet-4",
		},
		{
			name: "Exact mapping - no match",
			provider: Provider{
				ModelMapping: map[string]string{
					"claude-sonnet-4": "anthropic/claude-sonnet-4",
				},
			},
			requestedModel: "gpt-4",
			expected:       "gpt-4",
		},

		// Wildcard mapping
		{
			name: "Wildcard mapping - prefix",
			provider: Provider{
				ModelMapping: map[string]string{
					"claude-*": "anthropic/claude-*",
				},
			},
			requestedModel: "claude-sonnet-4",
			expected:       "anthropic/claude-sonnet-4",
		},
		{
			name: "Wildcard mapping - middle",
			provider: Provider{
				ModelMapping: map[string]string{
					"claude-*-4": "anthropic/claude-*-v4",
				},
			},
			requestedModel: "claude-sonnet-4",
			expected:       "anthropic/claude-sonnet-v4",
		},

		// Exact takes precedence over wildcard
		{
			name: "Exact mapping takes precedence",
			provider: Provider{
				ModelMapping: map[string]string{
					"claude-sonnet-4": "exact-match",
					"claude-*":        "wildcard-match",
				},
			},
			requestedModel: "claude-sonnet-4",
			expected:       "exact-match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.GetEffectiveModel(tt.requestedModel)
			if result != tt.expected {
				t.Errorf("GetEffectiveModel(%q) = %q, expected %q",
					tt.requestedModel, result, tt.expected)
			}
		})
	}
}

// ==================== ValidateConfiguration Tests ====================

func TestProvider_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		provider      Provider
		expectErrors  bool
		errorContains string
	}{
		// Valid configuration
		{
			name: "Valid configuration - complete",
			provider: Provider{
				Name: "test-provider",
				SupportedModels: []string{
					"model-a",
					"internal-model-b",
				},
				ModelMapping: map[string]string{
					"external-model-b": "internal-model-b",
				},
			},
			expectErrors: false,
		},

		// Invalid mapping: Target model not in whitelist
		{
			name: "Invalid mapping - target not in whitelist",
			provider: Provider{
				Name: "test-provider",
				SupportedModels: []string{
					"model-a",
				},
				ModelMapping: map[string]string{
					"external": "model-b",
				},
			},
			expectErrors:  true,
			errorContains: "not in supportedModels",
		},

		// Warning: Only mapping configured, whitelist not configured
		{
			name: "Warning - no whitelist",
			provider: Provider{
				Name: "test-provider",
				ModelMapping: map[string]string{
					"external": "internal",
				},
			},
			expectErrors:  true,
			errorContains: "supportedModels is not",
		},

		// Warning: Self-mapping
		{
			name: "Warning - self-mapping",
			provider: Provider{
				Name: "test-provider",
				SupportedModels: []string{
					"model-a",
				},
				ModelMapping: map[string]string{
					"model-a": "model-a",
				},
			},
			expectErrors:  true,
			errorContains: "maps to itself",
		},

		// Wildcard mapping (skip validation)
		{
			name: "Wildcard mapping - skip validation",
			provider: Provider{
				Name: "test-provider",
				SupportedModels: []string{
					"anthropic/claude-*",
				},
				ModelMapping: map[string]string{
					"claude-*": "anthropic/claude-*",
				},
			},
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.provider.ValidateConfiguration()

			if tt.expectErrors && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}

			if !tt.expectErrors && len(errors) > 0 {
				t.Errorf("Unexpected validation errors: %v", errors)
			}

			if tt.expectErrors && tt.errorContains != "" {
				found := false
				for _, err := range errors {
					if containsString(err, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error to contain %q, but actual errors were: %v", tt.errorContains, errors)
				}
			}
		})
	}
}

// ==================== Level Grouping Tests ====================

func TestProviderLevelGrouping(t *testing.T) {
	tests := []struct {
		name      string
		providers []Provider
		expected  map[int][]string // level -> provider names
	}{
		{
			name: "Default Level (not set)",
			providers: []Provider{
				{ID: 1, Name: "Provider-A", Level: 0}, // 0 should default to 1
				{ID: 2, Name: "Provider-B"},           // Not set should default to 1
			},
			expected: map[int][]string{
				1: {"Provider-A", "Provider-B"},
			},
		},
		{
			name: "Multiple Level grouping",
			providers: []Provider{
				{ID: 1, Name: "Provider-L1-A", Level: 1},
				{ID: 2, Name: "Provider-L2-A", Level: 2},
				{ID: 3, Name: "Provider-L1-B", Level: 1},
				{ID: 4, Name: "Provider-L3-A", Level: 3},
			},
			expected: map[int][]string{
				1: {"Provider-L1-A", "Provider-L1-B"},
				2: {"Provider-L2-A"},
				3: {"Provider-L3-A"},
			},
		},
		{
			name: "Preserve order within same Level",
			providers: []Provider{
				{ID: 1, Name: "First", Level: 1},
				{ID: 2, Name: "Second", Level: 1},
				{ID: 3, Name: "Third", Level: 1},
			},
			expected: map[int][]string{
				1: {"First", "Second", "Third"},
			},
		},
		{
			name: "Level 10 to Level 1 mixed",
			providers: []Provider{
				{ID: 1, Name: "L10", Level: 10},
				{ID: 2, Name: "L1", Level: 1},
				{ID: 3, Name: "L5", Level: 5},
			},
			expected: map[int][]string{
				1:  {"L1"},
				5:  {"L5"},
				10: {"L10"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate grouping logic
			levelGroups := make(map[int][]Provider)
			for _, provider := range tt.providers {
				level := provider.Level
				if level <= 0 {
					level = 1 // Default Level 1
				}
				levelGroups[level] = append(levelGroups[level], provider)
			}

			// Verify grouping results
			for expectedLevel, expectedNames := range tt.expected {
				actualProviders, exists := levelGroups[expectedLevel]
				if !exists {
					t.Errorf("Level %d does not exist, expected %d providers", expectedLevel, len(expectedNames))
					continue
				}

				if len(actualProviders) != len(expectedNames) {
					t.Errorf("Level %d provider count mismatch: actual %d, expected %d",
						expectedLevel, len(actualProviders), len(expectedNames))
					continue
				}

				// Verify order
				for i, expectedName := range expectedNames {
					if actualProviders[i].Name != expectedName {
						t.Errorf("Level %d position %d: actual %q, expected %q",
							expectedLevel, i, actualProviders[i].Name, expectedName)
					}
				}
			}

			// Verify no extra Levels
			if len(levelGroups) != len(tt.expected) {
				t.Errorf("Level group count mismatch: actual %d, expected %d",
					len(levelGroups), len(tt.expected))
			}
		})
	}
}

func TestProviderLevelOrdering(t *testing.T) {
	tests := []struct {
		name     string
		levels   []int
		expected []int
	}{
		{
			name:     "Ascending sort",
			levels:   []int{3, 1, 2},
			expected: []int{1, 2, 3},
		},
		{
			name:     "Already sorted",
			levels:   []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "Reverse order",
			levels:   []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:     "Duplicate Levels (should not occur in practice, but algorithm should handle)",
			levels:   []int{2, 1, 2, 3, 1},
			expected: []int{1, 1, 2, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use Go's sort package (consistent with actual code)
			levels := make([]int, len(tt.levels))
			copy(levels, tt.levels)
			sort.Ints(levels)

			for i, expected := range tt.expected {
				if levels[i] != expected {
					t.Errorf("Position %d: actual %d, expected %d", i, levels[i], expected)
				}
			}
		})
	}
}

func TestProviderLevelJSON(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected string
	}{
		{
			name: "Level set to 2",
			provider: Provider{
				ID:    1,
				Name:  "Test",
				Level: 2,
			},
			expected: `"level":2`,
		},
		{
			name: "Level not set (zero value, should omitempty)",
			provider: Provider{
				ID:    1,
				Name:  "Test",
				Level: 0,
			},
			expected: "", // omitempty should not serialize level field
		},
		{
			name: "Level set to 1",
			provider: Provider{
				ID:    1,
				Name:  "Test",
				Level: 1,
			},
			expected: `"level":1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.provider)
			if err != nil {
				t.Fatalf("JSON serialization failed: %v", err)
			}

			jsonStr := string(data)
			if tt.expected == "" {
				// Verify level field does not exist
				if containsString(jsonStr, `"level"`) {
					t.Errorf("Expected level field to be omitted, but found in JSON: %s", jsonStr)
				}
			} else {
				// Verify level field exists and is correct
				if !containsString(jsonStr, tt.expected) {
					t.Errorf("Expected JSON to contain %q, but actual was: %s", tt.expected, jsonStr)
				}
			}
		})
	}
}

// Helper functions
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==================== IsProviderNameAvailable Tests ====================

func TestProviderService_IsProviderNameAvailable(t *testing.T) {
	tests := []struct {
		name              string
		existingProviders map[string][]Provider // kind -> providers
		checkName         string
		excludeID         int
		expectedAvailable bool
		expectedErr       bool
	}{
		{
			name: "Name available - no existing providers",
			existingProviders: map[string][]Provider{
				"claude":   {},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "NewProvider",
			excludeID:         0,
			expectedAvailable: true,
			expectedErr:       false,
		},
		{
			name: "Name taken - duplicate in same kind",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "TestProvider"},
				},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "TestProvider",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       false,
		},
		{
			name: "Name taken - duplicate across kinds (claude -> codex)",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "MyProvider"},
				},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "MyProvider",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       false,
		},
		{
			name: "Name taken - duplicate across kinds (codex -> opencode)",
			existingProviders: map[string][]Provider{
				"claude": {},
				"codex": {
					{ID: 1, Name: "SharedProvider"},
				},
				"opencode": {},
			},
			checkName:         "SharedProvider",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       false,
		},
		{
			name: "Name available - editing same provider (excludeID works)",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "TestProvider"},
				},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "TestProvider",
			excludeID:         1,
			expectedAvailable: true,
			expectedErr:       false,
		},
		{
			name: "Case insensitive - exact match",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "MyProvider"},
				},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "myprovider",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       false,
		},
		{
			name: "Case insensitive - mixed case",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "MyProvider"},
				},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "MYPROVIDER",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       false,
		},
		{
			name: "Whitespace trimmed - leading/trailing spaces",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "MyProvider"},
				},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "  MyProvider  ",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       false,
		},
		{
			name: "Empty name - should error",
			existingProviders: map[string][]Provider{
				"claude":   {},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       true,
		},
		{
			name: "Whitespace only name - should error",
			existingProviders: map[string][]Provider{
				"claude":   {},
				"codex":    {},
				"opencode": {},
			},
			checkName:         "   ",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       true,
		},
		{
			name: "Name available - different provider exists",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "ProviderA"},
					{ID: 2, Name: "ProviderB"},
				},
				"codex": {
					{ID: 3, Name: "ProviderC"},
				},
				"opencode": {},
			},
			checkName:         "ProviderD",
			excludeID:         0,
			expectedAvailable: true,
			expectedErr:       false,
		},
		{
			name: "Name taken - check all three kinds",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "ClaudeProvider"},
				},
				"codex": {
					{ID: 2, Name: "CodexProvider"},
				},
				"opencode": {
					{ID: 3, Name: "OpenCodeProvider"},
				},
			},
			checkName:         "ClaudeProvider",
			excludeID:         0,
			expectedAvailable: false,
			expectedErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupProviderTestEnv(t)
			ps := NewProviderService()

			// Set up existing providers by mocking the LoadProviders behavior
			// In a real scenario, these would be saved to disk
			for kind, providers := range tt.existingProviders {
				path, err := providerFilePath(kind)
				if err != nil {
					t.Fatalf("Failed to get provider file path: %v", err)
				}
				data, err := json.MarshalIndent(providerEnvelope{Providers: providers}, "", "  ")
				if err != nil {
					t.Fatalf("Failed to marshal providers: %v", err)
				}
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatalf("Failed to write providers: %v", err)
				}
				defer os.Remove(path)
			}

			// Test IsProviderNameAvailable
			available, err := ps.IsProviderNameAvailable(tt.checkName, tt.excludeID)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if available != tt.expectedAvailable {
				t.Errorf("IsProviderNameAvailable(%q, excludeID=%d) = %v, expected %v",
					tt.checkName, tt.excludeID, available, tt.expectedAvailable)
			}
		})
	}
}

func TestProviderService_findProviderNameConflict(t *testing.T) {
	tests := []struct {
		name              string
		existingProviders map[string][]Provider
		checkName         string
		excludeID         int
		expectedKind      string
		expectedName      string
	}{
		{
			name: "Conflict found in claude kind",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "TestProvider"},
				},
				"codex":    {},
				"opencode": {},
			},
			checkName:    "TestProvider",
			excludeID:    0,
			expectedKind: "claude",
			expectedName: "TestProvider",
		},
		{
			name: "Conflict found in codex kind",
			existingProviders: map[string][]Provider{
				"claude": {},
				"codex": {
					{ID: 1, Name: "MyProvider"},
				},
				"opencode": {},
			},
			checkName:    "MyProvider",
			excludeID:    0,
			expectedKind: "codex",
			expectedName: "MyProvider",
		},
		{
			name: "Conflict found in opencode kind",
			existingProviders: map[string][]Provider{
				"claude": {},
				"codex":  {},
				"opencode": {
					{ID: 1, Name: "SharedProvider"},
				},
			},
			checkName:    "SharedProvider",
			excludeID:    0,
			expectedKind: "opencode",
			expectedName: "SharedProvider",
		},
		{
			name: "No conflict - unknown kind",
			existingProviders: map[string][]Provider{
				"claude":   {},
				"codex":    {},
				"opencode": {},
			},
			checkName:    "NonExistentProvider",
			excludeID:    0,
			expectedKind: "unknown",
			expectedName: "NonExistentProvider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupProviderTestEnv(t)
			ps := NewProviderService()

			// Set up existing providers
			for kind, providers := range tt.existingProviders {
				path, err := providerFilePath(kind)
				if err != nil {
					t.Fatalf("Failed to get provider file path: %v", err)
				}
				data, err := json.MarshalIndent(providerEnvelope{Providers: providers}, "", "  ")
				if err != nil {
					t.Fatalf("Failed to marshal providers: %v", err)
				}
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatalf("Failed to write providers: %v", err)
				}
				defer os.Remove(path)
			}

			// Test findProviderNameConflict
			conflict := ps.findProviderNameConflict(tt.checkName, tt.excludeID)

			if conflict.Kind != tt.expectedKind {
				t.Errorf("findProviderNameConflict(%q, excludeID=%d).Kind = %q, expected %q",
					tt.checkName, tt.excludeID, conflict.Kind, tt.expectedKind)
			}

			if conflict.Name != tt.expectedName {
				t.Errorf("findProviderNameConflict(%q, excludeID=%d).Name = %q, expected %q",
					tt.checkName, tt.excludeID, conflict.Name, tt.expectedName)
			}
		})
	}
}

func TestProviderService_SaveProviders_DuplicateNameValidation(t *testing.T) {
	tests := []struct {
		name              string
		existingProviders map[string][]Provider
		newProviders      []Provider
		saveKind          string
		expectedErr       bool
		errorContains     string
	}{
		{
			name: "Save successful - no duplicates",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "ProviderA"},
				},
			},
			newProviders: []Provider{
				{ID: 1, Name: "ProviderA"},
				{ID: 2, Name: "ProviderB"},
			},
			saveKind:    "claude",
			expectedErr: false,
		},
		{
			name: "Save fails - duplicate within same kind",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "ProviderA"},
				},
			},
			newProviders: []Provider{
				{ID: 1, Name: "ProviderA"},
				{ID: 2, Name: "ProviderA"}, // Duplicate name
			},
			saveKind:      "claude",
			expectedErr:   true,
			errorContains: "already used",
		},
		{
			name: "Save fails - duplicate across kinds (claude vs codex)",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "SharedProvider"},
				},
				"codex": {
					{ID: 2, Name: "CodexProvider"},
				},
			},
			newProviders: []Provider{
				{ID: 2, Name: "CodexProvider"},
				{ID: 3, Name: "SharedProvider"}, // Duplicate with claude
			},
			saveKind:      "codex",
			expectedErr:   true,
			errorContains: "already used",
		},
		{
			name: "Save successful - edit keeps same name",
			existingProviders: map[string][]Provider{
				"claude": {
					{ID: 1, Name: "MyProvider"},
				},
			},
			newProviders: []Provider{
				{ID: 1, Name: "MyProvider"}, // Same name, same ID - allowed
			},
			saveKind:    "claude",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupProviderTestEnv(t)
			ps := NewProviderService()

			// Set up existing providers
			for kind, providers := range tt.existingProviders {
				path, err := providerFilePath(kind)
				if err != nil {
					t.Fatalf("Failed to get provider file path: %v", err)
				}
				data, err := json.MarshalIndent(providerEnvelope{Providers: providers}, "", "  ")
				if err != nil {
					t.Fatalf("Failed to marshal providers: %v", err)
				}
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatalf("Failed to write providers: %v", err)
				}
				defer os.Remove(path)
			}

			// Attempt to save new providers
			err := ps.SaveProviders(tt.saveKind, tt.newProviders)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
