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


package provider

import (
	"testing"
)

// MockProvider implements Provider interface for testing
type MockProvider struct {
	id               int64
	name             string
	apiURL           string
	apiKey           string
	kind             string
	enabled          bool
	modelMapping     map[string]string
	supportedModels  []string
	level            int
}

func NewMockProvider(name string, enabled bool, supportedModels []string, modelMapping map[string]string) *MockProvider {
	return &MockProvider{
		id:              1,
		name:            name,
		apiURL:          "https://api.example.com",
		apiKey:          "test-key",
		kind:            "claude",
		enabled:         enabled,
		supportedModels: supportedModels,
		modelMapping:    modelMapping,
		level:           1,
	}
}

func (m *MockProvider) GetID() int64                 { return m.id }
func (m *MockProvider) GetName() string              { return m.name }
func (m *MockProvider) GetAPIURL() string            { return m.apiURL }
func (m *MockProvider) GetAPIKey() string            { return m.apiKey }
func (m *MockProvider) GetKind() string              { return m.kind }
func (m *MockProvider) IsEnabled() bool              { return m.enabled }
func (m *MockProvider) GetModelMapping() map[string]string {
	if m.modelMapping == nil {
		return map[string]string{}
	}
	return m.modelMapping
}
func (m *MockProvider) GetSupportedModels() []string {
	if m.supportedModels == nil {
		return []string{}
	}
	return m.supportedModels
}
func (m *MockProvider) GetLevel() int                { return m.level }

// ==================== IsModelSupported Tests ====================

func TestIsModelSupported(t *testing.T) {
	tests := []struct {
		name             string
		provider         Provider
		modelName        string
		expectedSupported bool
	}{
		{
			name: "empty config - all models supported",
			provider: NewMockProvider("test", true, nil, nil),
			modelName:        "any-model",
			expectedSupported: true,
		},
		{
			name: "exact match in supported models",
			provider: NewMockProvider("test", true,
				[]string{"claude-sonnet-4", "claude-opus-4"}, nil),
			modelName:        "claude-sonnet-4",
			expectedSupported: true,
		},
		{
			name: "wildcard match in supported models",
			provider: NewMockProvider("test", true,
				[]string{"claude-*", "gpt-*"}, nil),
			modelName:        "claude-sonnet-4",
			expectedSupported: true,
		},
		{
			name: "exact match in model mapping",
			provider: NewMockProvider("test", true, nil,
				map[string]string{"claude-sonnet": "claude-sonnet-4"}),
			modelName:        "claude-sonnet",
			expectedSupported: true,
		},
		{
			name: "wildcard match in model mapping",
			provider: NewMockProvider("test", true, nil,
				map[string]string{"claude-*": "anthropic/claude-*"}),
			modelName:        "claude-sonnet-4",
			expectedSupported: true,
		},
		{
			name: "model not supported",
			provider: NewMockProvider("test", true,
				[]string{"claude-sonnet-4"}, nil),
			modelName:        "gpt-4",
			expectedSupported: false,
		},
		{
			name:             "disabled provider",
			provider:         NewMockProvider("test", false, []string{"*"}, nil),
			modelName:        "any-model",
			expectedSupported: true, // IsModelSupported doesn't check enabled status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsModelSupported(tt.provider, tt.modelName)
			if result != tt.expectedSupported {
				t.Errorf("IsModelSupported() = %v, want %v", result, tt.expectedSupported)
			}
		})
	}
}

// ==================== GetEffectiveModel Tests ====================

func TestGetEffectiveModel(t *testing.T) {
	tests := []struct {
		name           string
		provider       Provider
		requestedModel string
		expectedModel  string
	}{
		{
			name:           "no mapping - return original",
			provider:       NewMockProvider("test", true, nil, nil),
			requestedModel: "claude-sonnet-4",
			expectedModel:  "claude-sonnet-4",
		},
		{
			name: "exact mapping",
			provider: NewMockProvider("test", true, nil,
				map[string]string{"claude-sonnet": "claude-sonnet-4"}),
			requestedModel: "claude-sonnet",
			expectedModel:  "claude-sonnet-4",
		},
		{
			name: "wildcard mapping - prefix",
			provider: NewMockProvider("test", true, nil,
				map[string]string{"claude-*": "anthropic/claude-*"}),
			requestedModel: "claude-sonnet-4",
			expectedModel:  "anthropic/claude-sonnet-4",
		},
		{
			name: "wildcard mapping - middle",
			provider: NewMockProvider("test", true, nil,
				map[string]string{"claude-*-4": "anthropic/claude-*-v4"}),
			requestedModel: "claude-sonnet-4",
			expectedModel:  "anthropic/claude-sonnet-v4",
		},
		{
			name: "wildcard mapping - no match returns original",
			provider: NewMockProvider("test", true, nil,
				map[string]string{"gpt-*": "openai/gpt-*"}),
			requestedModel: "claude-sonnet-4",
			expectedModel:  "claude-sonnet-4",
		},
		{
			name:           "empty mapping - return original",
			provider:       NewMockProvider("test", true, nil, map[string]string{}),
			requestedModel: "claude-sonnet-4",
			expectedModel:  "claude-sonnet-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEffectiveModel(tt.provider, tt.requestedModel)
			if result != tt.expectedModel {
				t.Errorf("GetEffectiveModel() = %q, want %q", result, tt.expectedModel)
			}
		})
	}
}

// ==================== ValidateConfiguration Tests ====================

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		provider       Provider
		expectedErrors int
		hasWarning     bool
	}{
		{
			name: "valid configuration",
			provider: NewMockProvider("test", true,
				[]string{"claude-sonnet-4", "claude-opus-4"},
				map[string]string{"claude-sonnet": "claude-sonnet-4", "claude-opus": "claude-opus-4"}),
			expectedErrors: 0,
			hasWarning:     false,
		},
		{
			name: "valid wildcard configuration",
			provider: NewMockProvider("test", true,
				[]string{"claude-*"},
				map[string]string{"claude-*": "anthropic/claude-*"}),
			expectedErrors: 0,
			hasWarning:     false,
		},
		{
			name: "invalid mapping target",
			provider: NewMockProvider("test", true,
				[]string{"claude-sonnet-4"},
				map[string]string{"external": "non-existent-model"}),
			expectedErrors: 1,
			hasWarning:     true,
		},
		{
			name: "mapping without supported models",
			provider: NewMockProvider("test", true,
				nil,
				map[string]string{"external": "non-existent-model"}),
			expectedErrors: 2, // Warning about no supported models + invalid mapping
			hasWarning:     true,
		},
		{
			name: "self-mapping warning",
			provider: NewMockProvider("test", true,
				[]string{"claude-sonnet-4"},
				map[string]string{"claude-sonnet-4": "claude-sonnet-4"}),
			expectedErrors: 1, // warning about self-mapping
			hasWarning:     true,
		},
		{
			name:           "empty configuration",
			provider:       NewMockProvider("test", true, nil, nil),
			expectedErrors: 0,
			hasWarning:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateConfiguration(tt.provider)
			if len(errors) != tt.expectedErrors {
				t.Errorf("ValidateConfiguration() returned %d errors, want %d", len(errors), tt.expectedErrors)
				if len(errors) > 0 {
					t.Logf("Errors: %v", errors)
				}
			}
			if tt.hasWarning && len(errors) == 0 {
				t.Errorf("ValidateConfiguration() expected warnings but got none")
			}
		})
	}
}

// ==================== matchWildcard Tests ====================

func TestMatchWildcard(t *testing.T) {
	tests := []struct {
		pattern  string
		text     string
		expected bool
	}{
		// Exact match
		{"claude-sonnet-4", "claude-sonnet-4", true},
		{"gpt-4", "gpt-4", true},

		// No wildcard - exact match required
		{"claude-sonnet-4", "claude-opus-4", false},
		{"gpt-4", "gpt-4-turbo", false},

		// Prefix wildcard
		{"claude-*", "claude-sonnet-4", true},
		{"claude-*", "claude-opus-4", true},
		{"claude-*", "gpt-4", false},

		// Suffix wildcard
		{"*-4", "claude-sonnet-4", true},
		{"*-4", "claude-sonnet-3.5", false},

		// Middle wildcard
		{"claude-*-4", "claude-sonnet-4", true},
		{"claude-*-4", "claude-opus-4", true},
		{"claude-*-4", "claude-sonnet-3.5", false},

		// Multiple wildcards (not supported)
		{"*-*", "claude-sonnet-4", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.text, func(t *testing.T) {
			result := matchWildcard(tt.pattern, tt.text)
			if result != tt.expected {
				t.Errorf("matchWildcard(%q, %q) = %v, want %v",
					tt.pattern, tt.text, result, tt.expected)
			}
		})
	}
}

// ==================== applyWildcardMapping Tests ====================

func TestApplyWildcardMapping(t *testing.T) {
	tests := []struct {
		pattern     string
		replacement string
		input       string
		expected    string
	}{
		// Prefix wildcard
		{
			pattern:     "claude-*",
			replacement: "anthropic/claude-*",
			input:       "claude-sonnet-4",
			expected:    "anthropic/claude-sonnet-4",
		},
		// Suffix wildcard
		{
			pattern:     "*-4",
			replacement: "model-*-v4",
			input:       "claude-sonnet-4",
			expected:    "model-claude-sonnet-v4",
		},
		// Middle wildcard
		{
			pattern:     "claude-*-4",
			replacement: "anthropic/claude-*-v4",
			input:       "claude-sonnet-4",
			expected:    "anthropic/claude-sonnet-v4",
		},
		// No wildcard in pattern
		{
			pattern:     "claude-sonnet",
			replacement: "claude-sonnet-4",
			input:       "claude-sonnet",
			expected:    "claude-sonnet-4",
		},
		// No match - return replacement
		{
			pattern:     "claude-*",
			replacement: "anthropic/claude-*",
			input:       "gpt-4",
			expected:    "anthropic/claude-*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := applyWildcardMapping(tt.pattern, tt.replacement, tt.input)
			if result != tt.expected {
				t.Errorf("applyWildcardMapping(%q, %q, %q) = %q, want %q",
					tt.pattern, tt.replacement, tt.input, result, tt.expected)
			}
		})
	}
}
