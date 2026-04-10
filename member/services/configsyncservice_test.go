package services

import (
	"testing"

	"github.com/code-together/shared/integration"
)

func TestConfigSyncService_ConvertAPIProvidersToLocal_AppliesRelayTokenForRelayProvider(t *testing.T) {
	service := &ConfigSyncService{}
	kind := integration.ProviderKindClaude
	emptyKey := ""

	providers := []integration.Provider{
		{
			Id:      1,
			Name:    "Claude Relay",
			ApiUrl:  "https://server.example.com/api/v1/relay/claude",
			ApiKey:  &emptyKey,
			Kind:    &kind,
			Enabled: true,
			TeamId:  1,
		},
	}

	converted := service.convertAPIProvidersToLocal(providers, "relay-token-123")
	got := converted["claude"][0]

	if got.APIKey != "relay-token-123" {
		t.Fatalf("APIKey = %q, want relay token", got.APIKey)
	}
	if got.APIURL != "https://server.example.com/api/v1/relay/claude" {
		t.Fatalf("APIURL = %q", got.APIURL)
	}
}

func TestConfigSyncService_ConvertAPIProvidersToLocal_PreservesDirectProviderKey(t *testing.T) {
	service := &ConfigSyncService{}
	kind := integration.ProviderKindClaude
	providerKey := "provider-secret"

	providers := []integration.Provider{
		{
			Id:      1,
			Name:    "Claude Direct",
			ApiUrl:  "https://api.anthropic.com",
			ApiKey:  &providerKey,
			Kind:    &kind,
			Enabled: true,
			TeamId:  1,
		},
	}

	converted := service.convertAPIProvidersToLocal(providers, "relay-token-123")
	got := converted["claude"][0]

	if got.APIKey != providerKey {
		t.Fatalf("APIKey = %q, want direct provider key", got.APIKey)
	}
}
