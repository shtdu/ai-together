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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/code-together/shared/integration"
)

// Note: SyncState struct is now defined in config.go
// This service uses ConfigService to manage settings in config.json

// SyncState tracks the last sync time
type SyncState struct {
	LastSyncTime time.Time `json:"last_sync_time"`
}

// ConfigSyncService handles synchronization of provider configurations from server
type ConfigSyncService struct {
	authService     *AuthService
	apiClient       integration.ClientWithResponsesInterface
	providerService *ProviderService
	importService   *ImportService
	configService   *ConfigService
	mu              sync.Mutex
}

// wrapAPIError converts an HTTP response into an error, parsing the response body
// for structured error information from the server
func (cs *ConfigSyncService) wrapAPIError(httpResp *http.Response, action string) error {
	if httpResp == nil {
		return fmt.Errorf("nil HTTP response from server")
	}

	// Successful response
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		return nil
	}

	// Parse the error response
	apiErr := integration.UnwrapJSONResponse(httpResp)
	if apiErr != nil {
		// Add context about what action was being performed
		return fmt.Errorf("%s: %w", action, apiErr)
	}

	// Fallback to status code if parsing failed
	return fmt.Errorf("%s: server returned status %d", action, httpResp.StatusCode)
}

// NewConfigSyncService creates a new config sync service
func NewConfigSyncService(authService *AuthService, apiClient integration.ClientWithResponsesInterface, providerService *ProviderService, importService *ImportService, configService *ConfigService) *ConfigSyncService {
	return &ConfigSyncService{
		authService:     authService,
		apiClient:       apiClient,
		providerService: providerService,
		importService:   importService,
		configService:   configService,
	}
}

// SyncProviders fetches and syncs all providers from the server
// Uses the Provider API endpoint and the shared ConfigBuilder to build configs locally
func (cs *ConfigSyncService) SyncProviders() error {
	slog.Info("syncing providers from Provider API endpoint")
	// All users use the Provider API endpoint
	if err := cs.syncProvidersFromAPI(); err != nil {
		return err
	}

	// Update last sync time
	if err := cs.saveSyncState(); err != nil {
		return fmt.Errorf("failed to save sync state: %w", err)
	}

	return nil
}

// SyncProviderType syncs only a specific provider type (claude, codex, opencode)
func (cs *ConfigSyncService) SyncProviderType(kind string) error {
	// First sync all providers (we filter server-side or locally)
	if err := cs.SyncProviders(); err != nil {
		return err
	}

	// The SyncProviders method already filters by kind when saving
	// So this is just a convenience method for syncing specific types
	return nil
}

// GetLastSyncTime returns the last successful sync time
func (cs *ConfigSyncService) GetLastSyncTime() (time.Time, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	state, err := cs.configService.GetSync()
	if err != nil {
		return time.Time{}, err
	}

	return state.LastSyncTime, nil
}

// saveSyncState saves the current sync state
func (cs *ConfigSyncService) saveSyncState() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	state := SyncState{
		LastSyncTime: time.Now(),
	}

	if err := cs.configService.SetSync(state); err != nil {
		return err
	}

	return nil
}

// PushProvidersToServer pushes local provider changes to the server
// This includes creating new providers, updating existing ones, and deleting removed ones
func (cs *ConfigSyncService) PushProvidersToServer() error {
	slog.Info("pushing providers to server")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Fetch existing server providers
	serverProviders, err := cs.getServerProviders(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch server providers: %w", err)
	}

	// Build name -> ID map for server providers
	serverProviderIDs := make(map[string]int64)
	for _, p := range serverProviders {
		serverProviderIDs[p.Name] = p.Id
	}

	// Only push provider types that have local changes
	totalSuccess := 0
	totalFailure := 0
	var allErrors []string
	successfulKinds := []string{}
	skippedKinds := []string{}

	for _, kind := range []string{"claude", "codex", "opencode"} {
		// Skip if no local changes for this kind
		if !cs.providerService.HasLocalChanges(kind) {
			skippedKinds = append(skippedKinds, kind)
			continue
		}

		success, failure, errs := cs.pushProviderKind(ctx, kind, serverProviderIDs, serverProviders)
		totalSuccess += success
		totalFailure += failure
		allErrors = append(allErrors, errs...)

		if failure == 0 && success > 0 {
			cs.providerService.ClearLocalChanges(kind)
			successfulKinds = append(successfulKinds, kind)
		}
	}

	if len(skippedKinds) > 0 {
		slog.Info("skipped provider types with no changes", "types", strings.Join(skippedKinds, ", "))
	}

	if totalFailure > 0 {
		return fmt.Errorf("push completed: %d succeeded, %d failed\n%s",
			totalSuccess, totalFailure, strings.Join(allErrors, "\n"))
	}

	if len(successfulKinds) > 0 {
		slog.Info("cleared local changes", "types", strings.Join(successfulKinds, ", "))
		slog.Info("successfully pushed providers (including creates, updates, and deletes)", "count", totalSuccess)

		// Trigger immediate sync from server to get server-side settings
		slog.Info("syncing from server after push")
		if err := cs.SyncProviders(); err != nil {
			slog.Warn("failed to sync from server after push", "error", err)
			// Don't fail the push operation if sync fails
		}
	}

	return nil
}

// pushProviderKind pushes all providers of a specific kind to the server
// This includes creating new providers, updating existing ones, and deleting removed ones
func (cs *ConfigSyncService) pushProviderKind(ctx context.Context, kind string, serverProviderIDs map[string]int64, serverProviders []integration.Provider) (success, failure int, errors []string) {
	localProviders, err := cs.providerService.LoadProviders(kind)
	if err != nil {
		return 0, 0, []string{fmt.Sprintf("[%s] load failed: %v", kind, err)}
	}

	// Build map of local provider names for quick lookup
	localProviderNames := make(map[string]bool)
	for _, p := range localProviders {
		localProviderNames[p.Name] = true
	}

	// Find server providers of this kind that no longer exist locally
	// and delete them from the server
	for _, sp := range serverProviders {
		if sp.Kind != nil && string(*sp.Kind) == kind {
			// This server provider is of the current kind
			// Check if it still exists locally
			if !localProviderNames[sp.Name] {
				// Provider exists on server but not locally - delete it
				if err := cs.deleteProvider(ctx, sp.Id, sp.Name); err != nil {
					failure++
					errors = append(errors, fmt.Sprintf("[%s] delete %s (id=%d) failed: %v", kind, sp.Name, sp.Id, err))
				} else {
					success++
					slog.Info("deleted provider from server", "kind", kind, "name", sp.Name, "id", sp.Id)
				}
			}
		}
	}

	// Now create/update local providers
	for _, p := range localProviders {
		serverID, exists := serverProviderIDs[p.Name]
		isUpdate := exists && p.ID > 0

		var err error
		if isUpdate {
			updateReq := cs.buildUpdateProviderRequest(p, kind)
			err = cs.updateProvider(ctx, serverID, updateReq)
		} else {
			createReq := cs.buildCreateProviderRequest(p, kind)
			err = cs.createProvider(ctx, createReq)
		}

		if err != nil {
			failure++
			action := "update"
			if !isUpdate {
				action = "create"
			}
			errors = append(errors, fmt.Sprintf("[%s] %s %s failed: %v", kind, p.Name, action, err))
		} else {
			success++
		}
	}

	return success, failure, errors
}

// buildCreateProviderRequest converts a local Provider to CreateProviderRequest format
func (cs *ConfigSyncService) buildCreateProviderRequest(p Provider, kind string) integration.CreateProviderRequest {
	providerKind := integration.CreateProviderRequestKind(kind)
	req := integration.CreateProviderRequest{
		Name:            p.Name,
		ApiUrl:          p.APIURL,
		ApiKey:          p.APIKey, // Required field, use empty string if not set
		Enabled:         &p.Enabled,
		Kind:            &providerKind,
		SupportedModels: &p.SupportedModels,
		ModelMapping:    &p.ModelMapping,
	}
	if p.Level > 0 {
		req.Level = &p.Level
	}
	return req
}

// buildUpdateProviderRequest converts a local Provider to UpdateProviderRequest format
func (cs *ConfigSyncService) buildUpdateProviderRequest(p Provider, kind string) integration.UpdateProviderRequest {
	providerKind := integration.UpdateProviderRequestKind(kind)
	req := integration.UpdateProviderRequest{
		ApiUrl:          &p.APIURL,
		Enabled:         &p.Enabled,
		Kind:            &providerKind,
		SupportedModels: &p.SupportedModels,
		ModelMapping:    &p.ModelMapping,
	}
	if p.APIKey != "" {
		req.ApiKey = &p.APIKey
	}
	if p.Level > 0 {
		req.Level = &p.Level
	}
	if p.Name != "" {
		req.Name = &p.Name
	}
	return req
}

// pushSingleProvider is no longer needed, removed

// updateProvider updates an existing provider on the server
func (cs *ConfigSyncService) updateProvider(ctx context.Context, serverID int64, req integration.UpdateProviderRequest) error {
	resp, err := cs.apiClient.PutApiV1ProvidersProviderIdWithResponse(ctx, integration.ProviderId(serverID), req)
	if err != nil {
		return err
	}
	if err := cs.wrapAPIError(resp.HTTPResponse, "update provider"); err != nil {
		return err
	}
	var name string
	if req.Name != nil {
		name = *req.Name
	}
	slog.Info("updated provider", "name", name, "id", serverID)
	return nil
}

// createProvider creates a new provider on the server
func (cs *ConfigSyncService) createProvider(ctx context.Context, req integration.CreateProviderRequest) error {
	resp, err := cs.apiClient.PostApiV1ProvidersWithResponse(ctx, req)
	if err != nil {
		return err
	}
	if err := cs.wrapAPIError(resp.HTTPResponse, "create provider"); err != nil {
		return err
	}
	if resp.JSON201 == nil {
		return fmt.Errorf("server returned nil response")
	}
	slog.Info("created provider", "name", req.Name, "id", resp.JSON201.Id)
	return nil
}

// deleteProvider deletes a provider from the server
func (cs *ConfigSyncService) deleteProvider(ctx context.Context, providerID int64, name string) error {
	resp, err := cs.apiClient.DeleteApiV1ProvidersProviderIdWithResponse(ctx, integration.ProviderId(providerID))
	if err != nil {
		return err
	}
	// Accept both 204 No Content and 200 OK
	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusOK {
		if err := cs.wrapAPIError(resp.HTTPResponse, "delete provider"); err != nil {
			return err
		}
	}
	return nil
}

// getServerProviders fetches all providers from the server
func (cs *ConfigSyncService) getServerProviders(ctx context.Context) ([]integration.Provider, error) {
	resp, err := cs.apiClient.GetApiV1ProvidersWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers: %w", err)
	}

	if err := cs.wrapAPIError(resp.HTTPResponse, "fetch providers"); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("server returned nil response")
	}

	return *resp.JSON200, nil
}

// detectProviderType determines the provider type (claude/codex/opencode) from the API URL
func detectProviderType(apiURL string) string {
	url := strings.ToLower(apiURL)

	// Claude providers typically use anthropic.com or contain "anthropic" in the URL
	if strings.Contains(url, "anthropic") || strings.Contains(url, "integration.anthropic.com") {
		return "claude"
	}

	// Codex providers typically use openai.com or contain "openai" in the URL
	if strings.Contains(url, "openai") || strings.Contains(url, "integration.openai.com") {
		return "codex"
	}

	// Default to opencode for everything else
	return "opencode"
}

// convertAPIProvidersToLocal converts API Provider format to local Provider format
func (cs *ConfigSyncService) convertAPIProvidersToLocal(apiProviders []integration.Provider) map[string][]Provider {
	providersByType := map[string][]Provider{
		"claude":   {},
		"codex":    {},
		"opencode": {},
	}

	for _, apiProvider := range apiProviders {
		// Use the Kind field from the server response, fallback to detection if not set
		providerType := ""
		if apiProvider.Kind != nil {
			providerType = string(*apiProvider.Kind)
		}
		if providerType == "" {
			providerType = detectProviderType(apiProvider.ApiUrl)
		}
		// Validate the provider type
		if providerType != "claude" && providerType != "codex" && providerType != "opencode" {
			slog.Warn("unknown provider type, defaulting to opencode", "type", providerType, "provider", apiProvider.Name)
			providerType = "opencode"
		}

		// Build the local Provider from API Provider
		localProvider := Provider{
			ID:      int(apiProvider.Id),
			Name:    apiProvider.Name,
			APIURL:  apiProvider.ApiUrl,
			APIKey:  "",
			Enabled: apiProvider.Enabled,
		}

		// Handle optional fields
		if apiProvider.ApiKey != nil {
			localProvider.APIKey = *apiProvider.ApiKey
		}

		// Handle ModelMapping
		if apiProvider.ModelMapping != nil {
			localProvider.ModelMapping = *apiProvider.ModelMapping
		} else {
			localProvider.ModelMapping = make(map[string]string)
		}

		// Handle SupportedModels
		if apiProvider.SupportedModels != nil {
			localProvider.SupportedModels = *apiProvider.SupportedModels
		} else {
			localProvider.SupportedModels = []string{}
		}

		// Handle Level
		if apiProvider.Level != nil {
			localProvider.Level = *apiProvider.Level
		}

		// Handle TeamId
		localProvider.TeamId = apiProvider.TeamId

		// Set default visual properties based on type
		switch providerType {
		case "claude":
			localProvider.Tint = "rgba(15, 23, 42, 0.12)"
			localProvider.Accent = "#0a84ff"
		case "codex":
			localProvider.Tint = "rgba(236, 72, 153, 0.16)"
			localProvider.Accent = "#ec4899"
		case "opencode":
			localProvider.Tint = "rgba(139, 92, 246, 0.16)"
			localProvider.Accent = "#8b5cf6"
		}

		providersByType[providerType] = append(providersByType[providerType], localProvider)
	}

	return providersByType
}

// syncProvidersFromAPI fetches providers from the Provider API endpoint and saves them
func (cs *ConfigSyncService) syncProvidersFromAPI() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch providers from server using API client
	resp, err := cs.apiClient.GetApiV1ProvidersWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch providers: %w", err)
	}

	if err := cs.wrapAPIError(resp.HTTPResponse, "fetch providers"); err != nil {
		return err
	}

	if resp.JSON200 == nil {
		return fmt.Errorf("server returned nil response")
	}

	apiProviders := *resp.JSON200
	if len(apiProviders) == 0 {
		slog.Info("no providers found from API")
		return nil
	}

	// Convert API providers to local format
	providersByType := cs.convertAPIProvidersToLocal(apiProviders)

	// Save providers by type
	totalImported := 0
	skippedTypes := []string{}
	for providerType, providers := range providersByType {
		if len(providers) == 0 {
			continue
		}

		// Check if there are local changes for this provider type
		// If yes, skip syncing to preserve manager's local edits
		if cs.providerService.HasLocalChanges(providerType) {
			slog.Info("skipping provider sync due to local changes",
				"type", providerType,
				"reason", "manager has local edits pending push",
			)
			skippedTypes = append(skippedTypes, providerType)
			continue
		}

		// Use replaceProvidersInternal to replace all providers of this type
		// This bypasses permission checks since syncing from server should be allowed for all users
		if err := cs.providerService.replaceProvidersInternal(providerType, providers); err != nil {
			return fmt.Errorf("failed to save %s providers: %w", providerType, err)
		}

		enabledCount := 0
		for _, p := range providers {
			if p.Enabled {
				enabledCount++
			}
		}
		totalImported += len(providers)
		slog.Info("synced providers from API",
			"type", providerType,
			"total", len(providers),
			"enabled", enabledCount,
		)
	}

	if totalImported > 0 {
		fmt.Printf("Sync completed from API: %d providers imported (including disabled)\n", totalImported)
	}

	if len(skippedTypes) > 0 {
		slog.Info("skipped sync for provider types with local changes",
			"types", strings.Join(skippedTypes, ", "),
			"action", "push local changes to server to enable future syncs",
		)
	}

	return nil
}
