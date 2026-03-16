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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/code-together/shared/provider"
)

type Provider struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	APIURL  string `json:"apiUrl"`
	APIKey  string `json:"apiKey"`
	Tint    string `json:"tint"`
	Accent  string `json:"accent"`
	Enabled bool   `json:"enabled"`

	// Model whitelist - Models natively supported by Provider (exact or wildcard patterns)
	SupportedModels []string `json:"supportedModels,omitempty"`

	// Model mapping - External model name -> Provider internal model name
	// Supports exact match and wildcard (e.g., "claude-*" -> "anthropic/claude-*")
	ModelMapping map[string]string `json:"modelMapping,omitempty"`

	// Priority grouping - Smaller number = higher priority (1-10, default 1)
	// Use omitempty to ensure zero values are not serialized, backward compatible
	Level int `json:"level,omitempty"`

	// Team ID from server (used when pushing to server)
	// Use omitempty to ensure zero values are not serialized, backward compatible
	TeamId int64 `json:"teamId,omitempty"`

	// Internal field: Configuration validation errors (not persisted)
	configErrors []string `json:"-"`
}

// GetProviderAdapter converts a Provider to implement the shared.Provider interface
func (p *Provider) GetProviderAdapter() provider.Provider {
	return &providerAdapter{p: p}
}

// providerAdapter implements the shared.Provider interface for services.Provider
type providerAdapter struct {
	p *Provider
}

func (a *providerAdapter) GetID() int64      { return int64(a.p.ID) }
func (a *providerAdapter) GetName() string   { return a.p.Name }
func (a *providerAdapter) GetAPIURL() string { return a.p.APIURL }
func (a *providerAdapter) GetAPIKey() string { return a.p.APIKey }
func (a *providerAdapter) GetKind() string   { return a.p.Tint } // Tint maps to Kind
func (a *providerAdapter) IsEnabled() bool   { return a.p.Enabled }
func (a *providerAdapter) GetLevel() int     { return a.p.Level }

func (a *providerAdapter) GetModelMapping() map[string]string {
	if a.p.ModelMapping == nil {
		return map[string]string{}
	}
	return a.p.ModelMapping
}

func (a *providerAdapter) GetSupportedModels() []string {
	if a.p.SupportedModels == nil {
		return []string{}
	}
	return a.p.SupportedModels
}

type providerEnvelope struct {
	Providers []Provider `json:"providers"`
}

type ProviderService struct {
	mu                sync.Mutex
	permissionService *PermissionService
	localChanges      map[string]bool // Track per provider type: "claude", "codex", "opencode"
}

func NewProviderService() *ProviderService {
	return &ProviderService{
		localChanges: make(map[string]bool),
	}
}

// SetPermissionService sets the permission service (called during initialization)
//
//wails:ignore
func (ps *ProviderService) SetPermissionService(permissionService *PermissionService) {
	ps.permissionService = permissionService
}

func (ps *ProviderService) Start() error { return nil }
func (ps *ProviderService) Stop() error  { return nil }

func providerFilePath(kind string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".code-together")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	var filename string
	switch strings.ToLower(kind) {
	case "claude", "claude-code", "claude_code":
		filename = "claude-code.json"
	case "codex":
		filename = "codex.json"
	case "opencode":
		filename = "opencode.json"
	default:
		return "", fmt.Errorf("unknown provider type: %s", kind)
	}
	return filepath.Join(dir, filename), nil
}

func (ps *ProviderService) SaveProviders(kind string, providers []Provider) error {
	// Check permissions before saving
	if ps.permissionService != nil {
		if err := ps.permissionService.EnsureAdmin(); err != nil {
			return err
		}
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Mark this provider type as having local changes
	normalizedKind := normalizeProviderKind(kind)
	ps.localChanges[normalizedKind] = true

	path, err := providerFilePath(kind)
	if err != nil {
		return err
	}

	// Defensive logging: track providers with model mappings
	mappingCount := 0
	for _, p := range providers {
		if len(p.ModelMapping) > 0 {
			mappingCount++
		}
	}

	existingProviders, err := ps.LoadProviders(kind)
	if err != nil {
		return err
	}
	nameByID := make(map[int]string, len(existingProviders))
	for _, p := range existingProviders {
		nameByID[p.ID] = p.Name
	}

	// Validate each provider's configuration
	validationErrors := make([]string, 0)
	for _, p := range providers {
		// Rule 1: name cannot be modified
		if oldName, ok := nameByID[p.ID]; ok && oldName != p.Name {
			return fmt.Errorf("provider id %d name cannot be modified", p.ID)
		}

		// Rule 1.5: name must be unique across all provider kinds
		available, err := ps.IsProviderNameAvailable(p.Name, p.ID)
		if err != nil {
			return fmt.Errorf("failed to check provider name availability: %w", err)
		}
		if !available {
			conflictInfo := ps.findProviderNameConflict(p.Name, p.ID)
			return fmt.Errorf("provider name '%s' is already used by %s provider '%s'",
				p.Name, conflictInfo.Kind, conflictInfo.Name)
		}

		// Rule 2: Validate model configuration
		if errs := p.ValidateConfiguration(); len(errs) > 0 {
			for _, errMsg := range errs {
				validationErrors = append(validationErrors, fmt.Sprintf("[%s] %s", p.Name, errMsg))
			}
		}
	}

	// If there are validation errors, return aggregated error
	if len(validationErrors) > 0 {
		return fmt.Errorf("Configuration validation failed:\n  - %s", strings.Join(validationErrors, "\n  - "))
	}

	data, err := json.MarshalIndent(providerEnvelope{Providers: providers}, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// replaceProvidersInternal is the internal implementation that replaces all providers of a kind
// This does NOT check permissions - it's used by sync operations
//
//wails:ignore
func (ps *ProviderService) replaceProvidersInternal(kind string, providers []Provider) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	path, err := providerFilePath(kind)
	if err != nil {
		return err
	}

	validationErrors := make([]string, 0)
	for _, p := range providers {
		if errs := p.ValidateConfiguration(); len(errs) > 0 {
			for _, errMsg := range errs {
				validationErrors = append(validationErrors, fmt.Sprintf("[%s] %s", p.Name, errMsg))
			}
		}
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("Configuration validation failed:\n  - %s", strings.Join(validationErrors, "\n  - "))
	}

	data, err := json.MarshalIndent(providerEnvelope{Providers: providers}, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (ps *ProviderService) ReplaceProviders(kind string, providers []Provider) error {
	// Check permissions before saving
	if ps.permissionService != nil {
		if err := ps.permissionService.EnsureAdmin(); err != nil {
			return err
		}
	}

	return ps.replaceProvidersInternal(kind, providers)
}

func (ps *ProviderService) LoadProviders(kind string) ([]Provider, error) {
	path, err := providerFilePath(kind)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var envelope providerEnvelope
	if len(data) == 0 {
		return []Provider{}, nil
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	// Initialize nil maps/slices to empty for consistent handling
	// This ensures that modelMapping and supportedModels are always non-nil,
	// preventing data loss during save/load cycles
	nilMappingCount := 0
	nilSupportedCount := 0
	for i := range envelope.Providers {
		if envelope.Providers[i].ModelMapping == nil {
			envelope.Providers[i].ModelMapping = make(map[string]string)
			nilMappingCount++
		}
		if envelope.Providers[i].SupportedModels == nil {
			envelope.Providers[i].SupportedModels = []string{}
			nilSupportedCount++
		}
	}

	// Defensive logging: track nil map initialization
	if nilMappingCount > 0 || nilSupportedCount > 0 {
		slog.Debug("initialized nil maps to empty maps",
			"kind", kind,
			"providers", len(envelope.Providers),
			"nil_model_mapping", nilMappingCount,
			"nil_supported_models", nilSupportedCount)
	}

	return envelope.Providers, nil
}

// IsModelSupported checks if provider supports specified model
// Support conditions: 1) Model is in SupportedModels (exact or wildcard match)
//  2. Model is in ModelMapping keys (exact or wildcard match)
func (p *Provider) IsModelSupported(modelName string) bool {
	return provider.IsModelSupported(p.GetProviderAdapter(), modelName)
}

// GetEffectiveModel gets the actual model name that should be used
// If mapping exists (exact or wildcard), returns mapped model name; otherwise returns original model name
func (p *Provider) GetEffectiveModel(requestedModel string) string {
	return provider.GetEffectiveModel(p.GetProviderAdapter(), requestedModel)
}

// ValidateConfiguration validates provider's model configuration
// Returns list of validation errors (empty means validation passed)
func (p *Provider) ValidateConfiguration() []string {
	errs := provider.ValidateConfiguration(p.GetProviderAdapter())
	p.configErrors = errs
	return errs
}

// normalizeProviderKind normalizes provider kind to a consistent key format
func normalizeProviderKind(kind string) string {
	switch strings.ToLower(kind) {
	case "claude", "claude-code", "claude_code":
		return "claude"
	case "codex":
		return "codex"
	case "opencode":
		return "opencode"
	default:
		return strings.ToLower(kind)
	}
}

// HasLocalChanges checks if the specified provider type has local changes pending push
//
//wails:ignore
func (ps *ProviderService) HasLocalChanges(kind string) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	normalizedKind := normalizeProviderKind(kind)
	return ps.localChanges[normalizedKind]
}

// HasAnyLocalChanges checks if any provider type has local changes pending push
//
//wails:ignore
func (ps *ProviderService) HasAnyLocalChanges() bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, hasChanges := range ps.localChanges {
		if hasChanges {
			return true
		}
	}
	return false
}

// ClearLocalChanges clears the local change tracking for the specified provider type
//
//wails:ignore
func (ps *ProviderService) ClearLocalChanges(kind string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	normalizedKind := normalizeProviderKind(kind)
	delete(ps.localChanges, normalizedKind)
}

// DeleteAllProviders deletes all provider configuration files and clears local changes tracking
// This is used during logout to clean up all provider settings
//
//wails:ignore
func (ps *ProviderService) DeleteAllProviders() error {
	slog.Info("deleting all provider configurations")

	// Delete all provider files
	for _, kind := range []string{"claude", "codex", "opencode"} {
		path, err := providerFilePath(kind)
		if err != nil {
			slog.Warn("failed to get provider file path", "kind", kind, "error", err)
			continue
		}

		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				slog.Warn("failed to delete provider file", "kind", kind, "path", path, "error", err)
			}
		} else {
			slog.Info("deleted provider file", "kind", kind, "path", path)
		}
	}

	// Clear local changes tracking
	ps.mu.Lock()
	ps.localChanges = make(map[string]bool)
	ps.mu.Unlock()

	slog.Info("deleted all provider configurations and cleared local changes tracking")
	return nil
}

// IsProviderNameAvailable checks if a provider name is available across all kinds
// excludeProviderID is used to allow providers to keep their existing name during edits
// Returns true if the name is available, false if it's already in use
//
//wails:ignore
func (ps *ProviderService) IsProviderNameAvailable(name string, excludeProviderID int) (bool, error) {
	normalizedName := strings.TrimSpace(strings.ToLower(name))

	if normalizedName == "" {
		return false, fmt.Errorf("provider name cannot be empty")
	}

	for _, kind := range []string{"claude", "codex", "opencode"} {
		providers, err := ps.LoadProviders(kind)
		if err != nil {
			return false, err
		}

		for _, provider := range providers {
			if provider.ID == excludeProviderID {
				continue
			}

			existingName := strings.TrimSpace(strings.ToLower(provider.Name))
			if existingName == normalizedName {
				return false, nil
			}
		}
	}

	return true, nil
}

// findProviderNameConflict finds details of a provider name conflict
//
//wails:ignore
func (ps *ProviderService) findProviderNameConflict(name string, excludeProviderID int) providerNameConflict {
	normalizedName := strings.TrimSpace(strings.ToLower(name))

	for _, kind := range []string{"claude", "codex", "opencode"} {
		providers, _ := ps.LoadProviders(kind)
		for _, provider := range providers {
			if provider.ID == excludeProviderID {
				continue
			}

			existingName := strings.TrimSpace(strings.ToLower(provider.Name))
			if existingName == normalizedName {
				return providerNameConflict{Kind: kind, Name: provider.Name}
			}
		}
	}

	return providerNameConflict{Kind: "unknown", Name: name}
}

type providerNameConflict struct {
	Kind string
	Name string
}

// ListEnabledProviderNames returns names of enabled providers for a given platform
// If platform is empty, returns enabled providers from all platforms
func (ps *ProviderService) ListEnabledProviderNames(platform string) ([]string, error) {
	platforms := []string{}
	if platform == "" {
		platforms = []string{"claude", "codex", "opencode"}
	} else {
		platforms = []string{normalizeProviderKind(platform)}
	}

	allProviders := make([]string, 0)
	for _, kind := range platforms {
		providers, err := ps.LoadProviders(kind)
		if err != nil {
			return []string{}, err
		}
		if providers == nil {
			continue
		}
		for _, p := range providers {
			if p.Enabled {
				allProviders = append(allProviders, p.Name)
			}
		}
	}

	return allProviders, nil
}
