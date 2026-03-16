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

package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	opencodeSettingsDir      = ".config/opencode"
	opencodeSettingsFileName = "opencode.json"
	opencodeBackupFileName   = "ct-studio.back.opencode.json"
	opencodeAuthTokenValue   = "code-together"
	opencodeSchemaURL        = "https://opencode.ai/config.json"
	codetogetherProviderID   = "codetogether"
	codetogetherProviderName = "Code Together"
	codetogetherProviderNpm  = "@ai-sdk/openai-compatible"
)

type OpenCodeSettingsService struct {
	relayAddr       string
	providerService *ProviderService
}

func NewOpenCodeSettingsService(relayAddr string) *OpenCodeSettingsService {
	return &OpenCodeSettingsService{relayAddr: relayAddr}
}

// SetProviderService sets the provider service (called during initialization)
//
//wails:ignore
func (ocs *OpenCodeSettingsService) SetProviderService(providerService *ProviderService) {
	ocs.providerService = providerService
}

func (ocs *OpenCodeSettingsService) ProxyStatus() (ClaudeProxyStatus, error) {
	status := ClaudeProxyStatus{Enabled: false, BaseURL: ocs.baseURL()}
	settingsPath, _, err := ocs.paths()
	if err != nil {
		return status, err
	}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return status, nil
		}
		return status, err
	}
	var payload opencodeSettingsFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return status, nil
	}
	baseURL := ocs.baseURL()

	// Check for new provider structure first
	if provider, err := getCodeTogetherProvider(payload); err == nil && provider != nil {
		// New format: check provider.codetogether.options.baseURL
		if options, ok := provider["options"].(map[string]interface{}); ok {
			if configBaseURL, ok := options["baseURL"].(string); ok {
				status.Enabled = strings.EqualFold(configBaseURL, baseURL)
				return status, nil
			}
		}
	}

	// Fallback: check old flat format for backward compatibility
	apiKey, _ := payload["apiKey"].(string)
	configBaseURL, _ := payload["baseURL"].(string)
	enabled := strings.EqualFold(apiKey, opencodeAuthTokenValue) &&
		strings.EqualFold(configBaseURL, baseURL)
	status.Enabled = enabled
	return status, nil
}

// validateOpenCodeProviders validates that all enabled OpenCode providers have proper configuration
// Returns an error if any enabled provider has no whitelisted models or has wildcard models
func (ocs *OpenCodeSettingsService) validateOpenCodeProviders() error {
	if ocs.providerService == nil {
		// Provider service not available, skip validation (shouldn't happen in normal operation)
		return nil
	}

	providers, err := ocs.providerService.LoadProviders("opencode")
	if err != nil {
		return fmt.Errorf("failed to load OpenCode providers: %w", err)
	}

	// Filter to only enabled providers
	var enabledProviders []Provider
	for _, provider := range providers {
		if provider.Enabled {
			enabledProviders = append(enabledProviders, provider)
		}
	}

	// If no enabled providers, that's an issue
	if len(enabledProviders) == 0 {
		return errors.New("no enabled OpenCode providers configured")
	}

	var validationErrors []string

	for _, provider := range enabledProviders {
		// Check 1: Provider must have whitelisted models configured
		if len(provider.SupportedModels) == 0 {
			validationErrors = append(validationErrors,
				fmt.Sprintf("provider \"%s\" has no whitelisted models configured", provider.Name))
			continue
		}

		// Check 2: No wildcards allowed in whitelisted models
		var wildcardModels []string
		for _, model := range provider.SupportedModels {
			if strings.Contains(model, "*") {
				wildcardModels = append(wildcardModels, model)
			}
		}

		if len(wildcardModels) > 0 {
			validationErrors = append(validationErrors,
				fmt.Sprintf("provider \"%s\" has wildcard models in whitelist: %v", provider.Name, wildcardModels))
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

func (ocs *OpenCodeSettingsService) EnableProxy() error {
	// Validate providers before enabling
	if err := ocs.validateOpenCodeProviders(); err != nil {
		return fmt.Errorf("cannot enable OpenCode: %s", err.Error())
	}

	// Load enabled providers to get the model name
	providers, err := ocs.providerService.LoadProviders("opencode")
	if err != nil {
		return fmt.Errorf("failed to load OpenCode providers: %w", err)
	}

	settingsPath, backupPath, err := ocs.paths()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	// Read existing config if it exists, otherwise start with empty map
	settings := make(opencodeSettingsFile)
	if _, err := os.Stat(settingsPath); err == nil {
		content, readErr := os.ReadFile(settingsPath)
		if readErr != nil {
			return readErr
		}
		// Backup existing config before modifying
		if err := os.WriteFile(backupPath, content, 0o600); err != nil {
			return err
		}
		// Unmarshal existing config to preserve all fields
		if err := json.Unmarshal(content, &settings); err != nil {
			// If unmarshal fails, start with empty map (will overwrite corrupted file)
			settings = make(opencodeSettingsFile)
		}
	}

	// Remove old flat format fields if they exist (migration cleanup)
	delete(settings, "apiKey")
	delete(settings, "baseURL")

	// Set proxy settings using new provider structure, using the first model from enabled providers
	setCodeTogetherProvider(settings, ocs.baseURL(), providers)

	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, payload, 0o600)
}

func (ocs *OpenCodeSettingsService) DisableProxy() error {
	settingsPath, backupPath, err := ocs.paths()
	if err != nil {
		return err
	}

	// First, try to restore from backup
	if _, err := os.Stat(backupPath); err == nil {
		// Backup exists, restore it
		if err := os.Remove(settingsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.Rename(backupPath, settingsPath); err != nil {
			return err
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		// Error checking backup file
		return err
	}

	// No backup exists, remove provider entry from current config
	if _, err := os.Stat(settingsPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// No config file exists, nothing to do
			return nil
		}
		return err
	}

	// Read current config
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	var settings opencodeSettingsFile
	if err := json.Unmarshal(content, &settings); err != nil {
		// If unmarshal fails, just delete the file
		return os.Remove(settingsPath)
	}

	// Remove the provider.codetogether entry
	if provider, ok := settings["provider"].(map[string]interface{}); ok {
		delete(provider, codetogetherProviderID)
		// If provider object is now empty, remove it entirely
		if len(provider) == 0 {
			delete(settings, "provider")
		}
	}

	// Remove the top-level model field
	delete(settings, "model")

	// If settings are now empty (except possibly $schema), remove the file
	if len(settings) == 0 || (len(settings) == 1 && settings["$schema"] != nil) {
		return os.Remove(settingsPath)
	}

	// Write updated config
	payload, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, payload, 0o600)
}

func (ocs *OpenCodeSettingsService) paths() (settingsPath string, backupPath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	dir := filepath.Join(home, opencodeSettingsDir)
	return filepath.Join(dir, opencodeSettingsFileName), filepath.Join(dir, opencodeBackupFileName), nil
}

func (ocs *OpenCodeSettingsService) baseURL() string {
	addr := strings.TrimSpace(ocs.relayAddr)
	if addr == "" {
		addr = ":18100"
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	host := addr
	if strings.HasPrefix(host, ":") {
		host = "127.0.0.1" + host
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	return host
}

// ensureProviderStructure ensures that the provider structure exists in settings
func ensureProviderStructure(settings opencodeSettingsFile) {
	if settings == nil {
		settings = make(opencodeSettingsFile)
	}

	// Ensure $schema is set
	if _, exists := settings["$schema"]; !exists {
		settings["$schema"] = opencodeSchemaURL
	}

	// Ensure provider object exists
	if _, exists := settings["provider"]; !exists {
		settings["provider"] = make(map[string]interface{})
	}
}

// getCodeTogetherProvider retrieves the Code Together provider configuration
// Returns the provider config map, or nil if not found
func getCodeTogetherProvider(settings opencodeSettingsFile) (map[string]interface{}, error) {
	if settings == nil {
		return nil, nil
	}

	provider, ok := settings["provider"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	ctProvider, ok := provider[codetogetherProviderID].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	return ctProvider, nil
}

// setCodeTogetherProvider sets or updates the Code Together provider configuration
// Always uses "default" as the model ID - the relay service handles model mapping
func setCodeTogetherProvider(settings opencodeSettingsFile, baseURL string, providers []Provider) {
	ensureProviderStructure(settings)

	// Clear all existing providers and only keep codetogether
	provider := make(map[string]interface{})
	settings["provider"] = provider

	// Always use "default" as the model - relay handles model mapping
	const modelID = "default"
	const modelName = "Default"

	// Set the top-level model field
	settings["model"] = modelID

	// Create the Code Together provider config
	ctProvider := map[string]interface{}{
		"npm":  codetogetherProviderNpm,
		"name": codetogetherProviderName,
		"options": map[string]interface{}{
			"baseURL": baseURL,
			"apiKey":  "Fake_Key_Only",
		},
		"models": map[string]interface{}{
			modelID: map[string]interface{}{
				"name": modelName,
			},
		},
	}

	provider[codetogetherProviderID] = ctProvider
}

// opencodeSettingsFile uses map[string]interface{} to preserve all OpenCode config fields
// while allowing us to read/write apiKey and baseURL for proxy management
type opencodeSettingsFile map[string]interface{}
