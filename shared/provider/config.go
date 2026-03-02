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

// CCSwitchConfig represents the configuration format for cc-switch provider configuration.
// It is used for both importing configurations and syncing between server and member clients.
type CCSwitchConfig struct {
	Claude   CCProviderSection `json:"claude"`
	Codex    CCProviderSection `json:"codex"`
	OpenCode CCProviderSection `json:"opencode"`
	MCP      CCMCPSection      `json:"mcp"`
}

// CCProviderSection contains providers for a specific platform (Claude, Codex, or OpenCode)
type CCProviderSection struct {
	Providers map[string]CCProviderEntry `json:"providers"`
}

// CCProviderEntry represents a single provider configuration
type CCProviderEntry struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	WebsiteURL      string              `json:"websiteUrl"`
	Settings        CCProviderSetting   `json:"settingsConfig"`
	Level           int                 `json:"level,omitempty"`
	ModelMapping    map[string]string   `json:"modelMapping,omitempty"`
	SupportedModels []string            `json:"supportedModels,omitempty"`
}

// CCProviderSetting contains the environment variables, authentication, and config for a provider
type CCProviderSetting struct {
	Env    map[string]string `json:"env"`
	Auth   map[string]string `json:"auth"`
	Config string            `json:"config"`
}

// CCMCPSection contains MCP server configurations for all platforms
type CCMCPSection struct {
	Claude   CCMCPPlatform `json:"claude"`
	Codex    CCMCPPlatform `json:"codex"`
	OpenCode CCMCPPlatform `json:"opencode"`
}

// CCMCPPlatform contains MCP servers for a specific platform
type CCMCPPlatform struct {
	Servers map[string]CCMCPServerEntry `json:"servers"`
}

// CCMCPServerEntry represents a single MCP server configuration
type CCMCPServerEntry struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Enabled     bool              `json:"enabled"`
	Homepage    string            `json:"homepage"`
	Description string            `json:"description"`
	Server      CCMCPServerConfig `json:"server"`
}

// CCMCPServerConfig contains the configuration for an MCP server
type CCMCPServerConfig struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	URL     string            `json:"url"`
}
