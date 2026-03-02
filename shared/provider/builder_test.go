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
	"sort"
	"testing"
)

// mockProvider is a mock implementation of the Provider interface for testing
type mockProvider struct {
	id              int64
	name            string
	apiURL          string
	apiKey          string
	kind            string
	enabled         bool
	modelMapping    map[string]string
	supportedModels []string
	level           int
}

func (m *mockProvider) GetID() int64           { return m.id }
func (m *mockProvider) GetName() string        { return m.name }
func (m *mockProvider) GetAPIURL() string       { return m.apiURL }
func (m *mockProvider) GetAPIKey() string       { return m.apiKey }
func (m *mockProvider) GetKind() string         { return m.kind }
func (m *mockProvider) IsEnabled() bool         { return m.enabled }
func (m *mockProvider) GetModelMapping() map[string]string {
	if m.modelMapping == nil {
		return make(map[string]string)
	}
	return m.modelMapping
}
func (m *mockProvider) GetSupportedModels() []string {
	if m.supportedModels == nil {
		return []string{}
	}
	return m.supportedModels
}
func (m *mockProvider) GetLevel() int { return m.level }

func TestConfigBuilder_BuildCCSwitchConfig(t *testing.T) {
	builder := NewConfigBuilder()

	tests := []struct {
		name      string
		providers []Provider
		want      CCSwitchConfig
	}{
		{
			name:      "empty providers",
			providers: []Provider{},
			want: CCSwitchConfig{
				Claude:   CCProviderSection{Providers: map[string]CCProviderEntry{}},
				Codex:    CCProviderSection{Providers: map[string]CCProviderEntry{}},
				OpenCode: CCProviderSection{Providers: map[string]CCProviderEntry{}},
				MCP: CCMCPSection{
					Claude:   CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					Codex:    CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					OpenCode: CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
				},
			},
		},
		{
			name: "all disabled providers",
			providers: []Provider{
				&mockProvider{id: 1, name: "Test Provider", enabled: false},
			},
			want: CCSwitchConfig{
				Claude:   CCProviderSection{Providers: map[string]CCProviderEntry{}},
				Codex:    CCProviderSection{Providers: map[string]CCProviderEntry{}},
				OpenCode: CCProviderSection{Providers: map[string]CCProviderEntry{}},
				MCP: CCMCPSection{
					Claude:   CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					Codex:    CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					OpenCode: CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
				},
			},
		},
		{
			name: "claude provider",
			providers: []Provider{
				&mockProvider{
					id:              1,
					name:            "Anthropic",
					apiURL:          "https://api.anthropic.com",
					apiKey:          "test-key",
					kind:            "claude",
					enabled:         true,
					level:           1,
					modelMapping:    map[string]string{"claude-3": "claude-3-5-sonnet-20241022"},
					supportedModels: []string{"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"},
				},
			},
			want: CCSwitchConfig{
				Claude: CCProviderSection{Providers: map[string]CCProviderEntry{
					"anthropic-1": {
						ID:         "1",
						Name:       "Anthropic",
						WebsiteURL: "",
						Level:      1,
						Settings: CCProviderSetting{
							Env: map[string]string{
								"ANTHROPIC_BASE_URL":     "https://api.anthropic.com",
								"ANTHROPIC_AUTH_TOKEN":   "test-key",
							},
							Auth:   map[string]string{},
							Config: "",
						},
						ModelMapping:    map[string]string{"claude-3": "claude-3-5-sonnet-20241022"},
						SupportedModels: []string{"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"},
					},
				}},
				Codex:    CCProviderSection{Providers: map[string]CCProviderEntry{}},
				OpenCode: CCProviderSection{Providers: map[string]CCProviderEntry{}},
				MCP: CCMCPSection{
					Claude:   CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					Codex:    CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					OpenCode: CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
				},
			},
		},
		{
			name: "codex provider",
			providers: []Provider{
				&mockProvider{
					id:              2,
					name:            "GitHub Codex",
					apiURL:          "https://github.com/api",
					apiKey:          "github-key",
					kind:            "codex",
					enabled:         true,
					level:           2,
					modelMapping:    map[string]string{"gpt-4": "gpt-4-turbo"},
					supportedModels: []string{"gpt-4-turbo", "gpt-3.5-turbo"},
				},
			},
			want: CCSwitchConfig{
				Claude:   CCProviderSection{Providers: map[string]CCProviderEntry{}},
				Codex: CCProviderSection{Providers: map[string]CCProviderEntry{
					"github-codex-2": {
						ID:         "2",
						Name:       "GitHub Codex",
						WebsiteURL: "",
						Level:      2,
						Settings: CCProviderSetting{
							Env:  map[string]string{},
							Auth: map[string]string{"OPENAI_API_KEY": "github-key"},
							Config: `model_provider = "github-codex-2"

[model_providers.github-codex-2]
name = "GitHub Codex"
base_url = "https://github.com/api"
`,
						},
						ModelMapping:    map[string]string{"gpt-4": "gpt-4-turbo"},
						SupportedModels: []string{"gpt-4-turbo", "gpt-3.5-turbo"},
					},
				}},
				OpenCode: CCProviderSection{Providers: map[string]CCProviderEntry{}},
				MCP: CCMCPSection{
					Claude:   CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					Codex:    CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					OpenCode: CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
				},
			},
		},
		{
			name: "opencode provider",
			providers: []Provider{
				&mockProvider{
					id:              3,
					name:            "OpenCode AI",
					apiURL:          "https://opencode.example.com",
					apiKey:          "opencode-key",
					kind:            "opencode",
					enabled:         true,
					level:           3,
					modelMapping:    map[string]string{"open-model": "open-model-v2"},
					supportedModels: []string{"open-model-v2"},
				},
			},
			want: CCSwitchConfig{
				Claude:   CCProviderSection{Providers: map[string]CCProviderEntry{}},
				Codex:    CCProviderSection{Providers: map[string]CCProviderEntry{}},
				OpenCode: CCProviderSection{Providers: map[string]CCProviderEntry{
					"opencode-ai-3": {
						ID:         "3",
						Name:       "OpenCode AI",
						WebsiteURL: "",
						Level:      3,
						Settings: CCProviderSetting{
							Env: map[string]string{
								"OPENAI_BASE_URL": "https://opencode.example.com",
								"OPENAI_API_KEY":  "opencode-key",
							},
							Auth:   map[string]string{},
							Config: "",
						},
						ModelMapping:    map[string]string{"open-model": "open-model-v2"},
						SupportedModels: []string{"open-model-v2"},
					},
				}},
				MCP: CCMCPSection{
					Claude:   CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					Codex:    CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
					OpenCode: CCMCPPlatform{Servers: map[string]CCMCPServerEntry{}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := builder.BuildCCSwitchConfig(tt.providers)

			// Compare Claude providers
			if len(got.Claude.Providers) != len(tt.want.Claude.Providers) {
				t.Errorf("BuildCCSwitchConfig() Claude providers count = %v, want %v", len(got.Claude.Providers), len(tt.want.Claude.Providers))
			}
			for key, wantEntry := range tt.want.Claude.Providers {
				gotEntry, ok := got.Claude.Providers[key]
				if !ok {
					t.Errorf("BuildCCSwitchConfig() missing Claude provider key %v", key)
					continue
				}
				if !compareProviderEntries(gotEntry, wantEntry) {
					t.Errorf("BuildCCSwitchConfig() Claude provider %v mismatch, got %+v, want %+v", key, gotEntry, wantEntry)
				}
			}

			// Compare Codex providers
			if len(got.Codex.Providers) != len(tt.want.Codex.Providers) {
				t.Errorf("BuildCCSwitchConfig() Codex providers count = %v, want %v", len(got.Codex.Providers), len(tt.want.Codex.Providers))
			}
			for key, wantEntry := range tt.want.Codex.Providers {
				gotEntry, ok := got.Codex.Providers[key]
				if !ok {
					t.Errorf("BuildCCSwitchConfig() missing Codex provider key %v", key)
					continue
				}
				if !compareProviderEntries(gotEntry, wantEntry) {
					t.Errorf("BuildCCSwitchConfig() Codex provider %v mismatch, got %+v, want %+v", key, gotEntry, wantEntry)
				}
			}

			// Compare OpenCode providers
			if len(got.OpenCode.Providers) != len(tt.want.OpenCode.Providers) {
				t.Errorf("BuildCCSwitchConfig() OpenCode providers count = %v, want %v", len(got.OpenCode.Providers), len(tt.want.OpenCode.Providers))
			}
			for key, wantEntry := range tt.want.OpenCode.Providers {
				gotEntry, ok := got.OpenCode.Providers[key]
				if !ok {
					t.Errorf("BuildCCSwitchConfig() missing OpenCode provider key %v", key)
					continue
				}
				if !compareProviderEntries(gotEntry, wantEntry) {
					t.Errorf("BuildCCSwitchConfig() OpenCode provider %v mismatch, got %+v, want %+v", key, gotEntry, wantEntry)
				}
			}
		})
	}
}

func compareProviderEntries(got, want CCProviderEntry) bool {
	if got.ID != want.ID || got.Name != want.Name || got.Level != want.Level {
		return false
	}

	// Compare Env
	if len(got.Settings.Env) != len(want.Settings.Env) {
		return false
	}
	for k, v := range want.Settings.Env {
		if got.Settings.Env[k] != v {
			return false
		}
	}

	// Compare Auth
	if len(got.Settings.Auth) != len(want.Settings.Auth) {
		return false
	}
	for k, v := range want.Settings.Auth {
		if got.Settings.Auth[k] != v {
			return false
		}
	}

	// Compare Config (only check if non-empty in want)
	if want.Settings.Config != "" && got.Settings.Config != want.Settings.Config {
		return false
	}

	// Compare ModelMapping
	if len(got.ModelMapping) != len(want.ModelMapping) {
		return false
	}
	for k, v := range want.ModelMapping {
		if got.ModelMapping[k] != v {
			return false
		}
	}

	// Compare SupportedModels
	if len(got.SupportedModels) != len(want.SupportedModels) {
		return false
	}
	gotModels := make([]string, len(got.SupportedModels))
	copy(gotModels, got.SupportedModels)
	sort.Strings(gotModels)
	wantModels := make([]string, len(want.SupportedModels))
	copy(wantModels, want.SupportedModels)
	sort.Strings(wantModels)
	for i := range gotModels {
		if gotModels[i] != wantModels[i] {
			return false
		}
	}

	return true
}

func TestConfigBuilder_DetermineProviderKind(t *testing.T) {
	builder := NewConfigBuilder()

	tests := []struct {
		name     string
		provider Provider
		want     string
	}{
		{
			name: "explicit claude kind",
			provider: &mockProvider{
				name: "Test",
				kind: "claude",
			},
			want: "claude",
		},
		{
			name: "anthropic url",
			provider: &mockProvider{
				name:  "Test",
				apiURL: "https://api.anthropic.com",
			},
			want: "claude",
		},
		{
			name: "claude in url",
			provider: &mockProvider{
				name:  "Test",
				apiURL: "https://api.claude.example.com",
			},
			want: "claude",
		},
		{
			name: "anthropic in name",
			provider: &mockProvider{
				name: "Anthropic Provider",
			},
			want: "claude",
		},
		{
			name: "explicit codex kind",
			provider: &mockProvider{
				name: "Test",
				kind: "codex",
			},
			want: "codex",
		},
		{
			name: "github url",
			provider: &mockProvider{
				name:  "Test",
				apiURL: "https://github.com/api",
			},
			want: "codex",
		},
		{
			name: "codex in name",
			provider: &mockProvider{
				name: "GitHub Codex",
			},
			want: "codex",
		},
		{
			name: "explicit opencode kind",
			provider: &mockProvider{
				name: "Test",
				kind: "opencode",
			},
			want: "opencode",
		},
		{
			name: "opencode in url",
			provider: &mockProvider{
				name:  "Test",
				apiURL: "https://opencode.example.com",
			},
			want: "opencode",
		},
		{
			name: "default to codex",
			provider: &mockProvider{
				name:  "Unknown Provider",
				apiURL: "https://unknown.example.com",
			},
			want: "codex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.DetermineProviderKind(tt.provider); got != tt.want {
				t.Errorf("DetermineProviderKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigBuilder_NormalizeProviderKey(t *testing.T) {
	builder := NewConfigBuilder()

	tests := []struct {
		name string
		id   int64
		want string
	}{
		{name: "Test Provider", id: 1, want: "test-provider-1"},
		{name: "  Spaces  ", id: 2, want: "spaces-2"},
		{name: "Underscores_Test", id: 3, want: "underscores-test-3"},
		{name: "MiXeD CaSe", id: 4, want: "mixed-case-4"},
		{name: "Multiple   Spaces", id: 5, want: "multiple---spaces-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.NormalizeProviderKey(tt.name, tt.id); got != tt.want {
				t.Errorf("NormalizeProviderKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseProviderJSON(t *testing.T) {
	tests := []struct {
		name                string
		modelMappingJSON    string
		supportedModelsJSON string
		wantModelMapping    map[string]string
		wantSupportedModels []string
		wantErr             bool
	}{
		{
			name:                "empty strings",
			modelMappingJSON:    "",
			supportedModelsJSON: "",
			wantModelMapping:    map[string]string{},
			wantSupportedModels: []string{},
			wantErr:             false,
		},
		{
			name:                "valid model mapping",
			modelMappingJSON:    `{"claude-3":"claude-3-5-sonnet-20241022"}`,
			supportedModelsJSON: "",
			wantModelMapping:    map[string]string{"claude-3": "claude-3-5-sonnet-20241022"},
			wantSupportedModels: []string{},
			wantErr:             false,
		},
		{
			name:                "valid supported models",
			modelMappingJSON:    "",
			supportedModelsJSON: `["model1","model2"]`,
			wantModelMapping:    map[string]string{},
			wantSupportedModels: []string{"model1", "model2"},
			wantErr:             false,
		},
		{
			name:                "both valid",
			modelMappingJSON:    `{"key":"value"}`,
			supportedModelsJSON: `["model1"]`,
			wantModelMapping:    map[string]string{"key": "value"},
			wantSupportedModels: []string{"model1"},
			wantErr:             false,
		},
		{
			name:                "invalid model mapping",
			modelMappingJSON:    `{invalid json}`,
			supportedModelsJSON: "",
			wantModelMapping:    nil,
			wantSupportedModels: nil,
			wantErr:             true,
		},
		{
			name:                "invalid supported models",
			modelMappingJSON:    "",
			supportedModelsJSON: `[invalid json]`,
			wantModelMapping:    nil,
			wantSupportedModels: nil,
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModelMapping, gotSupportedModels, err := ParseProviderJSON(tt.modelMappingJSON, tt.supportedModelsJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProviderJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !compareStringMaps(gotModelMapping, tt.wantModelMapping) {
					t.Errorf("ParseProviderJSON() modelMapping = %v, want %v", gotModelMapping, tt.wantModelMapping)
				}
				if !compareStringSlices(gotSupportedModels, tt.wantSupportedModels) {
					t.Errorf("ParseProviderJSON() supportedModels = %v, want %v", gotSupportedModels, tt.wantSupportedModels)
				}
			}
		})
	}
}

func compareStringMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]string, len(a))
	copy(aCopy, a)
	sort.Strings(aCopy)
	bCopy := make([]string, len(b))
	copy(bCopy, b)
	sort.Strings(bCopy)
	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}
	return true
}
