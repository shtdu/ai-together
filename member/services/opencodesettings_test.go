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
	"path/filepath"
	"testing"
)

// ==================== Helper Functions Tests ====================

func TestEnsureProviderStructure(t *testing.T) {
	tests := []struct {
		name         string
		settings     opencodeSettingsFile
		wantSchema   string
		wantProvider bool
	}{
		{
			name:         "Empty settings",
			settings:     make(opencodeSettingsFile),
			wantSchema:   opencodeSchemaURL,
			wantProvider: true,
		},
		{
			name: "Settings with existing schema",
			settings: opencodeSettingsFile{
				"$schema": "existing-schema",
			},
			wantSchema:   "existing-schema",
			wantProvider: true,
		},
		{
			name: "Settings with existing provider",
			settings: opencodeSettingsFile{
				"provider": map[string]interface{}{
					"other": "value",
				},
			},
			wantSchema:   opencodeSchemaURL,
			wantProvider: true,
		},
		{
			name: "Settings with both schema and provider",
			settings: opencodeSettingsFile{
				"$schema": "existing-schema",
				"provider": map[string]interface{}{
					"other": "value",
				},
			},
			wantSchema:   "existing-schema",
			wantProvider: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensureProviderStructure(tt.settings)

			// Check $schema
			schema, _ := tt.settings["$schema"].(string)
			if schema != tt.wantSchema {
				t.Errorf("$schema = %v, want %v", schema, tt.wantSchema)
			}

			// Check provider exists
			_, providerExists := tt.settings["provider"]
			if providerExists != tt.wantProvider {
				t.Errorf("provider exists = %v, want %v", providerExists, tt.wantProvider)
			}
		})
	}
}

func TestGetCodeTogetherProvider(t *testing.T) {
	tests := []struct {
		name     string
		settings opencodeSettingsFile
		wantNil  bool
	}{
		{
			name:     "No provider",
			settings: make(opencodeSettingsFile),
			wantNil:  true,
		},
		{
			name: "Provider without codetogether",
			settings: opencodeSettingsFile{
				"provider": map[string]interface{}{
					"other": "value",
				},
			},
			wantNil: true,
		},
		{
			name: "Provider with codetogether",
			settings: opencodeSettingsFile{
				"provider": map[string]interface{}{
					"codetogether": map[string]interface{}{
						"npm":  "@ai-sdk/openai-compatible",
						"name": "AI Together",
					},
				},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := getCodeTogetherProvider(tt.settings)
			if (provider == nil) != tt.wantNil {
				t.Errorf("getCodeTogetherProvider() = %v, wantNil %v", provider, tt.wantNil)
			}
			if err != nil {
				t.Errorf("getCodeTogetherProvider() error = %v", err)
			}
		})
	}
}

func TestSetCodeTogetherProvider(t *testing.T) {
	settings := make(opencodeSettingsFile)
	baseURL := "http://127.0.0.1:18100"

	// Create test providers (model should not affect the result)
	providers := []Provider{
		{
			ID:              1,
			Name:            "Test Provider",
			Enabled:         true,
			SupportedModels: []string{"gpt-4", "gpt-3.5"},
		},
	}

	setCodeTogetherProvider(settings, baseURL, providers)

	// Verify top-level model field is set to "default"
	modelID, ok := settings["model"].(string)
	if !ok {
		t.Fatal("top-level model field not found or not a string")
	}
	if modelID != "default" {
		t.Errorf("top-level model = %v, want 'default'", modelID)
	}

	// Verify provider structure
	provider, ok := settings["provider"].(map[string]interface{})
	if !ok {
		t.Fatal("provider not found or not a map")
	}

	ctProvider, ok := provider["codetogether"].(map[string]interface{})
	if !ok {
		t.Fatal("codetogether provider not found or not a map")
	}

	// Verify fields
	if npm, _ := ctProvider["npm"].(string); npm != codetogetherProviderNpm {
		t.Errorf("npm = %v, want %v", npm, codetogetherProviderNpm)
	}
	if name, _ := ctProvider["name"].(string); name != codetogetherProviderName {
		t.Errorf("name = %v, want %v", name, codetogetherProviderName)
	}

	// Verify options.baseURL and options.apiKey
	options, ok := ctProvider["options"].(map[string]interface{})
	if !ok {
		t.Fatal("options not found or not a map")
	}
	if baseURLValue, _ := options["baseURL"].(string); baseURLValue != baseURL {
		t.Errorf("options.baseURL = %v, want %v", baseURLValue, baseURL)
	}
	if apiKey, _ := options["apiKey"].(string); apiKey != "Fake_Key_Only" {
		t.Errorf("options.apiKey = %v, want Fake_Key_Only", apiKey)
	}

	// Verify models - should always use "default" (relay handles model mapping)
	models, ok := ctProvider["models"].(map[string]interface{})
	if !ok {
		t.Fatal("models not found or not a map")
	}
	if _, ok := models["default"]; !ok {
		t.Error("models.default not found, should always use 'default' model")
	}

	// Verify model name is "Default"
	defaultModel, ok := models["default"].(map[string]interface{})
	if !ok {
		t.Fatal("models.default is not a map")
	}
	if name, _ := defaultModel["name"].(string); name != "Default" {
		t.Errorf("model name = %v, want 'Default'", name)
	}

	// Verify $schema
	schema, _ := settings["$schema"].(string)
	if schema != opencodeSchemaURL {
		t.Errorf("$schema = %v, want %v", schema, opencodeSchemaURL)
	}

	// Verify only codetogether provider exists (no other providers)
	if len(provider) != 1 {
		t.Errorf("provider map should only have codetogether, got %d providers", len(provider))
	}
}

// ==================== Migration Tests ====================

func TestProxyStatusOldFormat(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", homeDir) }()
	os.Setenv("HOME", tmpDir)

	// Create config with old flat format
	configDir := filepath.Join(tmpDir, opencodeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, opencodeSettingsFileName)
	oldSettings := opencodeSettingsFile{
		"apiKey":  "code-together",
		"baseURL": "http://127.0.0.1:18100",
	}
	data, _ := json.MarshalIndent(oldSettings, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Test ProxyStatus
	service := NewOpenCodeSettingsService(":18100")
	status, err := service.ProxyStatus()
	if err != nil {
		t.Fatalf("ProxyStatus() error = %v", err)
	}

	if !status.Enabled {
		t.Errorf("ProxyStatus() Enabled = %v, want true", status.Enabled)
	}
	if status.BaseURL != "http://127.0.0.1:18100" {
		t.Errorf("ProxyStatus() BaseURL = %v, want http://127.0.0.1:18100", status.BaseURL)
	}
}

func TestProxyStatusNewFormat(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", homeDir) }()
	os.Setenv("HOME", tmpDir)

	// Create config with new nested format
	configDir := filepath.Join(tmpDir, opencodeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, opencodeSettingsFileName)
	newSettings := opencodeSettingsFile{
		"$schema": opencodeSchemaURL,
		"provider": map[string]interface{}{
			"codetogether": map[string]interface{}{
				"npm":  codetogetherProviderNpm,
				"name": codetogetherProviderName,
				"options": map[string]interface{}{
					"baseURL": "http://127.0.0.1:18100",
				},
				"models": map[string]interface{}{
					"default": map[string]interface{}{
						"name": "Default Model",
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(newSettings, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Test ProxyStatus
	service := NewOpenCodeSettingsService(":18100")
	status, err := service.ProxyStatus()
	if err != nil {
		t.Fatalf("ProxyStatus() error = %v", err)
	}

	if !status.Enabled {
		t.Errorf("ProxyStatus() Enabled = %v, want true", status.Enabled)
	}
}

func TestEnableProxyMigratesOldFormat(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", homeDir) }()
	os.Setenv("HOME", tmpDir)

	// Create config with old flat format
	configDir := filepath.Join(tmpDir, opencodeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, opencodeSettingsFileName)
	oldSettings := opencodeSettingsFile{
		"apiKey":  "code-together",
		"baseURL": "http://127.0.0.1:18100",
		"mcp": map[string]interface{}{
			"enabled": true,
		},
	}
	data, _ := json.MarshalIndent(oldSettings, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Create provider service and add test providers
	providerService := NewProviderService()
	testProviders := []Provider{
		{
			ID:              1,
			Name:            "Test Provider",
			Enabled:         true,
			SupportedModels: []string{"gpt-4"},
		},
	}
	if err := providerService.SaveProviders("opencode", testProviders); err != nil {
		t.Fatal(err)
	}

	// Enable proxy (should migrate to new format)
	service := NewOpenCodeSettingsService(":18100")
	service.SetProviderService(providerService)
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	// Read back the config
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var newSettings opencodeSettingsFile
	if err := json.Unmarshal(content, &newSettings); err != nil {
		t.Fatal(err)
	}

	// Verify old fields are removed
	if _, exists := newSettings["apiKey"]; exists {
		t.Error("apiKey field should be removed after migration")
	}
	if _, exists := newSettings["baseURL"]; exists {
		t.Error("baseURL field should be removed after migration")
	}

	// Verify new provider structure exists
	provider, ok := newSettings["provider"].(map[string]interface{})
	if !ok {
		t.Fatal("provider not found after migration")
	}

	ctProvider, ok := provider["codetogether"].(map[string]interface{})
	if !ok {
		t.Fatal("codetogether provider not found after migration")
	}

	// Verify provider fields
	if npm, _ := ctProvider["npm"].(string); npm != codetogetherProviderNpm {
		t.Errorf("npm = %v, want %v", npm, codetogetherProviderNpm)
	}

	// Verify mcp field is preserved
	if _, exists := newSettings["mcp"]; !exists {
		t.Error("mcp field should be preserved after migration")
	}
}

func TestEnableProxyPreservesExistingFields(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", homeDir) }()
	os.Setenv("HOME", tmpDir)

	// Create config with existing fields
	configDir := filepath.Join(tmpDir, opencodeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, opencodeSettingsFileName)
	existingSettings := opencodeSettingsFile{
		"$schema": opencodeSchemaURL,
		"mcp": map[string]interface{}{
			"enabled": true,
			"servers": []string{"server1", "server2"},
		},
		"customField": "customValue",
	}
	data, _ := json.MarshalIndent(existingSettings, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Create provider service and add test providers
	providerService := NewProviderService()
	testProviders := []Provider{
		{
			ID:              1,
			Name:            "Test Provider",
			Enabled:         true,
			SupportedModels: []string{"gpt-4"},
		},
	}
	if err := providerService.SaveProviders("opencode", testProviders); err != nil {
		t.Fatal(err)
	}

	// Enable proxy
	service := NewOpenCodeSettingsService(":18100")
	service.SetProviderService(providerService)
	if err := service.EnableProxy(); err != nil {
		t.Fatalf("EnableProxy() error = %v", err)
	}

	// Read back the config
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var newSettings opencodeSettingsFile
	if err := json.Unmarshal(content, &newSettings); err != nil {
		t.Fatal(err)
	}

	// Verify existing fields are preserved
	if _, exists := newSettings["mcp"]; !exists {
		t.Error("mcp field should be preserved")
	}
	if customField, _ := newSettings["customField"].(string); customField != "customValue" {
		t.Errorf("customField = %v, want customValue", customField)
	}

	// Verify provider structure exists
	if _, exists := newSettings["provider"]; !exists {
		t.Error("provider field should exist")
	}
}

func TestDisableProxyRemovesOnlyCodeTogetherProvider(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", homeDir) }()
	os.Setenv("HOME", tmpDir)

	// Create config with provider and other fields
	configDir := filepath.Join(tmpDir, opencodeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, opencodeSettingsFileName)
	settings := opencodeSettingsFile{
		"$schema": opencodeSchemaURL,
		"provider": map[string]interface{}{
			"codetogether": map[string]interface{}{
				"npm":  codetogetherProviderNpm,
				"name": codetogetherProviderName,
				"options": map[string]interface{}{
					"baseURL": "http://127.0.0.1:18100",
				},
			},
			"otherProvider": map[string]interface{}{
				"name": "Other Provider",
			},
		},
		"mcp": map[string]interface{}{
			"enabled": true,
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Disable proxy
	service := NewOpenCodeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	// Read back the config
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var newSettings opencodeSettingsFile
	if err := json.Unmarshal(content, &newSettings); err != nil {
		t.Fatal(err)
	}

	// Verify codetogether provider is removed
	provider, _ := newSettings["provider"].(map[string]interface{})
	if _, exists := provider["codetogether"]; exists {
		t.Error("codetogether provider should be removed")
	}

	// Verify other provider is preserved
	if _, exists := provider["otherProvider"]; !exists {
		t.Error("otherProvider should be preserved")
	}

	// Verify mcp field is preserved
	if _, exists := newSettings["mcp"]; !exists {
		t.Error("mcp field should be preserved")
	}
}

func TestDisableProxyRemovesEmptyProvider(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", homeDir) }()
	os.Setenv("HOME", tmpDir)

	// Create config with only codetogether provider
	configDir := filepath.Join(tmpDir, opencodeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, opencodeSettingsFileName)
	settings := opencodeSettingsFile{
		"$schema": opencodeSchemaURL,
		"provider": map[string]interface{}{
			"codetogether": map[string]interface{}{
				"npm":  codetogetherProviderNpm,
				"name": codetogetherProviderName,
				"options": map[string]interface{}{
					"baseURL": "http://127.0.0.1:18100",
				},
			},
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	// Disable proxy
	service := NewOpenCodeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	// File should be removed since only schema would remain
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		// File exists, read it to verify it doesn't have provider
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatal(err)
		}

		var newSettings opencodeSettingsFile
		if err := json.Unmarshal(content, &newSettings); err != nil {
			t.Fatal(err)
		}

		// Verify provider object is removed (not just codetogether)
		if _, exists := newSettings["provider"]; exists {
			t.Error("provider object should be removed when empty")
		}
	}
	// If file doesn't exist, that's also correct (it was removed because only schema remained)
}

func TestBaseURL(t *testing.T) {
	tests := []struct {
		name      string
		relayAddr string
		want      string
	}{
		{
			name:      "Empty address",
			relayAddr: "",
			want:      "http://127.0.0.1:18100",
		},
		{
			name:      "Port only",
			relayAddr: ":18100",
			want:      "http://127.0.0.1:18100",
		},
		{
			name:      "Host only",
			relayAddr: "localhost:18100",
			want:      "http://localhost:18100",
		},
		{
			name:      "Full URL with http",
			relayAddr: "http://example.com:18100",
			want:      "http://example.com:18100",
		},
		{
			name:      "Full URL with https",
			relayAddr: "https://example.com:18100",
			want:      "https://example.com:18100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOpenCodeSettingsService(tt.relayAddr)
			got := service.baseURL()
			if got != tt.want {
				t.Errorf("baseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDisableProxyRestoresBackup(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	defer func() { os.Setenv("HOME", homeDir) }()
	os.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, opencodeSettingsDir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(configDir, opencodeSettingsFileName)
	backupPath := filepath.Join(configDir, opencodeBackupFileName)

	// Create backup file with original settings
	originalSettings := opencodeSettingsFile{
		"$schema": opencodeSchemaURL,
		"mcp": map[string]interface{}{
			"enabled": true,
		},
		"originalField": "originalValue",
	}
	backupData, _ := json.MarshalIndent(originalSettings, "", "  ")
	if err := os.WriteFile(backupPath, backupData, 0o600); err != nil {
		t.Fatal(err)
	}

	// Create current settings with provider
	currentSettings := opencodeSettingsFile{
		"$schema": opencodeSchemaURL,
		"provider": map[string]interface{}{
			"codetogether": map[string]interface{}{
				"npm":  codetogetherProviderNpm,
				"name": codetogetherProviderName,
				"options": map[string]interface{}{
					"baseURL": "http://127.0.0.1:18100",
				},
			},
		},
	}
	currentData, _ := json.MarshalIndent(currentSettings, "", "  ")
	if err := os.WriteFile(settingsPath, currentData, 0o600); err != nil {
		t.Fatal(err)
	}

	// Disable proxy (should restore from backup)
	service := NewOpenCodeSettingsService(":18100")
	if err := service.DisableProxy(); err != nil {
		t.Fatalf("DisableProxy() error = %v", err)
	}

	// Verify backup was restored
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var restoredSettings opencodeSettingsFile
	if err := json.Unmarshal(content, &restoredSettings); err != nil {
		t.Fatal(err)
	}

	// Verify original settings are restored
	if _, exists := restoredSettings["originalField"]; !exists {
		t.Error("originalField should be restored from backup")
	}

	// Verify provider is not present
	if _, exists := restoredSettings["provider"]; exists {
		t.Error("provider should not exist after backup restore")
	}

	// Verify backup file was removed
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("backup file should be removed after restore")
	}
}
