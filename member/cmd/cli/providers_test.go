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
	"testing"

	"codeswitch/services"
)

// TestInMemoryProviderService_LoadProviders_Success tests successful provider loading
func TestInMemoryProviderService_LoadProviders_Success(t *testing.T) {
	providersByType := map[string][]services.Provider{
		"claude": {
			{
				ID:      1,
				Name:    "test-claude",
				APIURL:  "http://test.com",
				Enabled: true,
				Tint:    "rgba(15, 23, 42, 0.12)",
				Accent:  "#0a84ff",
			},
		},
		"codex": {
			{
				ID:      2,
				Name:    "test-codex",
				APIURL:  "http://test.com",
				Enabled: true,
				Tint:    "rgba(236, 72, 153, 0.16)",
				Accent:  "#ec4899",
			},
		},
		"opencode": {
			{
				ID:      3,
				Name:    "test-opencode",
				APIURL:  "http://test.com",
				Enabled: true,
				Tint:    "rgba(139, 92, 246, 0.16)",
				Accent:  "#8b5cf6",
			},
		},
	}

	service := &InMemoryProviderService{
		providers: providersByType,
	}

	tests := []struct {
		name           string
		kind           string
		wantCount      int
		wantFirstName  string
		wantFirstTint  string
		wantFirstAccent string
	}{
		{
			name:           "claude providers",
			kind:           "claude",
			wantCount:      1,
			wantFirstName:  "test-claude",
			wantFirstTint:  "rgba(15, 23, 42, 0.12)",
			wantFirstAccent: "#0a84ff",
		},
		{
			name:           "codex providers",
			kind:           "codex",
			wantCount:      1,
			wantFirstName:  "test-codex",
			wantFirstTint:  "rgba(236, 72, 153, 0.16)",
			wantFirstAccent: "#ec4899",
		},
		{
			name:           "opencode providers",
			kind:           "opencode",
			wantCount:      1,
			wantFirstName:  "test-opencode",
			wantFirstTint:  "rgba(139, 92, 246, 0.16)",
			wantFirstAccent: "#8b5cf6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers, err := service.LoadProviders(tt.kind)
			if err != nil {
				t.Fatalf("LoadProviders() error = %v", err)
			}
			if len(providers) != tt.wantCount {
				t.Errorf("LoadProviders() returned %d providers, want %d", len(providers), tt.wantCount)
			}
			if len(providers) > 0 {
				if providers[0].Name != tt.wantFirstName {
					t.Errorf("LoadProviders()[0].Name = %q, want %q", providers[0].Name, tt.wantFirstName)
				}
				if providers[0].Tint != tt.wantFirstTint {
					t.Errorf("LoadProviders()[0].Tint = %q, want %q", providers[0].Tint, tt.wantFirstTint)
				}
				if providers[0].Accent != tt.wantFirstAccent {
					t.Errorf("LoadProviders()[0].Accent = %q, want %q", providers[0].Accent, tt.wantFirstAccent)
				}
			}
		})
	}
}

// TestInMemoryProviderService_LoadProviders_EmptyKind tests loading empty provider list
func TestInMemoryProviderService_LoadProviders_EmptyKind(t *testing.T) {
	providersByType := map[string][]services.Provider{
		"claude": {
			{ID: 1, Name: "test-claude", APIURL: "http://test.com", Enabled: true},
		},
		// opencode is empty
	}

	service := &InMemoryProviderService{
		providers: providersByType,
	}

	providers, err := service.LoadProviders("opencode")
	if err != nil {
		t.Fatalf("LoadProviders() error = %v", err)
	}
	if len(providers) != 0 {
		t.Errorf("LoadProviders() returned %d providers, want 0", len(providers))
	}
}

// TestInMemoryProviderService_LoadProviders_NotLoaded tests error when providers not loaded
func TestInMemoryProviderService_LoadProviders_NotLoaded(t *testing.T) {
	emptyService := &InMemoryProviderService{}

	_, err := emptyService.LoadProviders("claude")
	if err == nil {
		t.Error("LoadProviders() on empty service should return error")
	}
	if err != nil && err.Error() != "providers not loaded" {
		t.Errorf("LoadProviders() error = %q, want 'providers not loaded'", err.Error())
	}
}

// TestInMemoryProviderService_LoadProviders_MultipleProviders tests loading multiple providers of same type
func TestInMemoryProviderService_LoadProviders_MultipleProviders(t *testing.T) {
	providersByType := map[string][]services.Provider{
		"claude": {
			{ID: 1, Name: "claude-1", APIURL: "http://test1.com", Enabled: true},
			{ID: 2, Name: "claude-2", APIURL: "http://test2.com", Enabled: true},
			{ID: 3, Name: "claude-3", APIURL: "http://test3.com", Enabled: false},
		},
	}

	service := &InMemoryProviderService{
		providers: providersByType,
	}

	providers, err := service.LoadProviders("claude")
	if err != nil {
		t.Fatalf("LoadProviders() error = %v", err)
	}
	if len(providers) != 3 {
		t.Errorf("LoadProviders() returned %d providers, want 3", len(providers))
	}
	// Verify order is preserved
	if providers[0].Name != "claude-1" {
		t.Errorf("providers[0].Name = %q, want 'claude-1'", providers[0].Name)
	}
	if providers[2].Enabled {
		t.Error("providers[2].Enabled should be false")
	}
}

// TestLoadProvidersFromServer_IntegrationNote documents that this function requires integration testing
// TODO: Add integration tests with a test server to test:
// - Successful provider loading from server
// - API client creation failure
// - Server returns non-200 response
// - Server returns nil JSON200
// - Provider with nil optional fields (ModelMapping, SupportedModels, Level)
// - Provider with unknown type (should default to opencode)
// - Multiple providers of different types
// - Visual properties are set correctly based on type
// - API key is empty in loaded providers (security)
//
// To implement these tests, you would need to:
// 1. Set up a test server with mock provider data
// 2. Configure auth token for the test server
// 3. Mock or use httptest.Server for the API endpoint
func TestLoadProvidersFromServer_IntegrationNote(t *testing.T) {
	// This is a placeholder to document integration test requirements
	// The function loadProvidersFromServer makes actual API calls and requires
	// a running server for proper testing.
	//
	// Run integration tests with: make integration-test
	t.Skip("loadProvidersFromServer requires integration test setup with test server")
}
