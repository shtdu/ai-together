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
	"encoding/json"
	"fmt"
	"strings"
)

// Provider represents a generic provider interface that can be converted to ccSwitchConfig format.
// Both server and member modules should have their Provider types implement this interface.
type Provider interface {
	GetID() int64
	GetName() string
	GetAPIURL() string
	GetAPIKey() string
	GetKind() string
	IsEnabled() bool
	GetModelMapping() map[string]string
	GetSupportedModels() []string
	GetLevel() int
}

// ConfigBuilder builds ccSwitchConfig from a list of providers
type ConfigBuilder struct{}

// NewConfigBuilder creates a new ConfigBuilder instance
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// BuildCCSwitchConfig transforms a list of providers to CCSwitchConfig format
func (b *ConfigBuilder) BuildCCSwitchConfig(providers []Provider) CCSwitchConfig {
	config := CCSwitchConfig{
		Claude:   CCProviderSection{Providers: make(map[string]CCProviderEntry)},
		Codex:    CCProviderSection{Providers: make(map[string]CCProviderEntry)},
		OpenCode: CCProviderSection{Providers: make(map[string]CCProviderEntry)},
		MCP: CCMCPSection{
			Claude:   CCMCPPlatform{Servers: make(map[string]CCMCPServerEntry)},
			Codex:    CCMCPPlatform{Servers: make(map[string]CCMCPServerEntry)},
			OpenCode: CCMCPPlatform{Servers: make(map[string]CCMCPServerEntry)},
		},
	}

	for _, provider := range providers {
		// Skip disabled providers
		if !provider.IsEnabled() {
			continue
		}

		kind := b.DetermineProviderKind(provider)
		key, entry := b.TransformProviderToCCEntry(provider, kind)

		switch kind {
		case "claude":
			config.Claude.Providers[key] = entry
		case "codex":
			config.Codex.Providers[key] = entry
		case "opencode":
			config.OpenCode.Providers[key] = entry
		}
	}

	return config
}

// DetermineProviderKind determines the provider kind from the Provider interface
func (b *ConfigBuilder) DetermineProviderKind(provider Provider) string {
	// Use Kind field if available
	if provider.GetKind() != "" {
		return strings.ToLower(provider.GetKind())
	}

	// Fallback: infer from API URL or name
	apiURL := strings.ToLower(provider.GetAPIURL())
	name := strings.ToLower(provider.GetName())

	// Check for Claude/Anthropic
	if strings.Contains(apiURL, "anthropic") || strings.Contains(apiURL, "claude") ||
		strings.Contains(name, "anthropic") || strings.Contains(name, "claude") {
		return "claude"
	}

	// Check for Codex/GitHub
	if strings.Contains(apiURL, "github") || strings.Contains(name, "github") ||
		strings.Contains(name, "codex") {
		return "codex"
	}

	// Check for OpenCode
	if strings.Contains(apiURL, "opencode") || strings.Contains(name, "opencode") {
		return "opencode"
	}

	// Default to codex
	return "codex"
}

// TransformProviderToCCEntry converts a Provider to CCProviderEntry format
func (b *ConfigBuilder) TransformProviderToCCEntry(provider Provider, kind string) (string, CCProviderEntry) {
	// Generate a unique key for the provider
	key := b.NormalizeProviderKey(provider.GetName(), provider.GetID())

	entry := CCProviderEntry{
		ID:         fmt.Sprintf("%d", provider.GetID()),
		Name:       provider.GetName(),
		WebsiteURL: "", // Can be added later if provider has website field
		Level:      provider.GetLevel(),
		Settings: CCProviderSetting{
			Env:    make(map[string]string),
			Auth:   make(map[string]string),
			Config: "",
		},
		ModelMapping:    provider.GetModelMapping(),
		SupportedModels: provider.GetSupportedModels(),
	}

	switch kind {
	case "claude":
		entry.Settings.Env["ANTHROPIC_BASE_URL"] = provider.GetAPIURL()
		entry.Settings.Env["ANTHROPIC_AUTH_TOKEN"] = provider.GetAPIKey()
	case "codex":
		entry.Settings.Auth["OPENAI_API_KEY"] = provider.GetAPIKey()
		// Build TOML config for Codex
		entry.Settings.Config = b.BuildCodexConfig(provider)
	case "opencode":
		// OpenCode uses OpenAI-compatible API format
		entry.Settings.Env["OPENAI_BASE_URL"] = provider.GetAPIURL()
		entry.Settings.Env["OPENAI_API_KEY"] = provider.GetAPIKey()
	}

	return key, entry
}

// BuildCodexConfig creates TOML configuration for Codex providers
func (b *ConfigBuilder) BuildCodexConfig(provider Provider) string {
	providerKey := b.NormalizeProviderKey(provider.GetName(), provider.GetID())
	return fmt.Sprintf(`model_provider = "%s"

[model_providers.%s]
name = "%s"
base_url = "%s"
`, providerKey, providerKey, provider.GetName(), provider.GetAPIURL())
}

// NormalizeProviderKey creates a normalized key for provider maps
func (b *ConfigBuilder) NormalizeProviderKey(name string, id int64) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return fmt.Sprintf("%s-%d", normalized, id)
}

// ParseProviderJSON parses JSON strings into map[string]string and []string
// This is useful for providers that store ModelMapping and SupportedModels as JSON strings
func ParseProviderJSON(modelMappingJSON, supportedModelsJSON string) (map[string]string, []string, error) {
	var modelMapping map[string]string
	if modelMappingJSON != "" {
		if err := json.Unmarshal([]byte(modelMappingJSON), &modelMapping); err != nil {
			return nil, nil, fmt.Errorf("failed to parse model_mapping: %w", err)
		}
	}

	var supportedModels []string
	if supportedModelsJSON != "" {
		if err := json.Unmarshal([]byte(supportedModelsJSON), &supportedModels); err != nil {
			return nil, nil, fmt.Errorf("failed to parse supported_models: %w", err)
		}
	}

	if modelMapping == nil {
		modelMapping = make(map[string]string)
	}
	if supportedModels == nil {
		supportedModels = []string{}
	}

	return modelMapping, supportedModels, nil
}
