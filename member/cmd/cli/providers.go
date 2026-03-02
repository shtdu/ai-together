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
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/code-together/shared/integration"
	"codeswitch/services"
)

// InMemoryProviderService loads providers from server instead of local files
type InMemoryProviderService struct {
	providers map[string][]services.Provider
}

// LoadProviders returns providers for the given kind (claude, codex, opencode)
// Implements the same signature as ProviderService.LoadProviders()
func (s *InMemoryProviderService) LoadProviders(kind string) ([]services.Provider, error) {
	if s.providers == nil {
		return nil, fmt.Errorf("providers not loaded")
	}
	return s.providers[kind], nil
}

// loadProvidersFromServer loads providers directly from the server
// and returns them as in-memory Provider structs organized by type
// Does NOT persist to local files
func loadProvidersFromServer(serverURL, authToken string, verbose bool) (map[string][]services.Provider, error) {
	// Create authenticated API client
	apiClient, err := integration.NewAuthenticatedClient(
		serverURL,
		func() (string, error) { return authToken, nil },
		slog.Default(),
		verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Fetch providers from server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := apiClient.GetApiV1ProvidersWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch providers: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("server returned nil response")
	}

	apiProviders := *resp.JSON200

	// Convert API providers to local format
	providersByType := map[string][]services.Provider{
		"claude":   {},
		"codex":    {},
		"opencode": {},
	}

	for _, apiProvider := range apiProviders {
		// Determine provider type from Kind field
		providerType := "opencode" // default
		if apiProvider.Kind != nil {
			providerType = string(*apiProvider.Kind)
		}

		// Validate the provider type
		if providerType != "claude" && providerType != "codex" && providerType != "opencode" {
			slog.Warn("unknown provider type, defaulting to opencode", "type", providerType, "provider", apiProvider.Name)
			providerType = "opencode"
		}

		// Build the local Provider from API Provider
		localProvider := services.Provider{
			ID:      int(apiProvider.Id),
			Name:    apiProvider.Name,
			APIURL:  apiProvider.ApiUrl,
			Enabled: apiProvider.Enabled,
		}

		// Handle APIKey (pointer field)
		if apiProvider.ApiKey != nil {
			localProvider.APIKey = *apiProvider.ApiKey
		}

		// Handle optional fields
		if apiProvider.ModelMapping != nil {
			localProvider.ModelMapping = *apiProvider.ModelMapping
		} else {
			localProvider.ModelMapping = make(map[string]string)
		}

		if apiProvider.SupportedModels != nil {
			localProvider.SupportedModels = *apiProvider.SupportedModels
		} else {
			localProvider.SupportedModels = []string{}
		}

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

	slog.Info("loaded providers from server",
		"claude", len(providersByType["claude"]),
		"codex", len(providersByType["codex"]),
		"opencode", len(providersByType["opencode"]))

	return providersByType, nil
}
