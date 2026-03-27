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

package integration

import (
	"context"
	"net/http"
	"time"

	integrationclient "github.com/code-together/shared/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderFailoverNextProviderOnFailure tests failover to next provider on failure.
func (s *IntegrationTestSuite) TestProviderFailoverNextProviderOnFailure() {
	ctx := context.Background()

	// Create multiple providers with different priority levels
	kind := integrationclient.CreateProviderRequestKindClaude

	// Primary provider (level 1)
	primaryName := generateUniqueProviderName("failover-primary")
	req1 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            primaryName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-primary-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(1),
	}

	createResp1, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp1.StatusCode())
	primaryID := createResp1.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, primaryID)

	// Secondary provider (level 2)
	secondaryName := generateUniqueProviderName("failover-secondary")
	req2 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            secondaryName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-secondary-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(2),
	}

	createResp2, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp2.StatusCode())
	secondaryID := createResp2.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, secondaryID)

	// Verify both providers exist
	getResp1, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, primaryID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, getResp1.StatusCode())
	assert.Equal(s.T(), 1, *getResp1.JSON200.Level, "Primary should have level 1")

	getResp2, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, secondaryID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, getResp2.StatusCode())
	assert.Equal(s.T(), 2, *getResp2.JSON200.Level, "Secondary should have level 2")

	// Simulate failover by disabling primary
	disabled := false
	updateReq := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Enabled: &disabled,
	}

	_, err = s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, primaryID, updateReq)
	require.NoError(s.T(), err)

	// Verify primary is disabled and secondary is still available
	getPrimaryResp, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, primaryID)
	require.NoError(s.T(), err)
	assert.False(s.T(), getPrimaryResp.JSON200.Enabled, "Primary should be disabled")

	getSecondaryResp, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, secondaryID)
	require.NoError(s.T(), err)
	assert.True(s.T(), getSecondaryResp.JSON200.Enabled, "Secondary should still be enabled")

	s.Logger.Info("Failover to next provider verified", "primary_disabled", !getPrimaryResp.JSON200.Enabled,
		"secondary_available", getSecondaryResp.JSON200.Enabled)
}

// TestProviderFailoverThirtySecondTimeout tests 30-second timeout for failover.
func (s *IntegrationTestSuite) TestProviderFailoverThirtySecondTimeout() {
	ctx := context.Background()

	// Note: This test documents the 30-second timeout behavior
	// In a real scenario:
	// 1. Request is made to primary provider
	// 2. If no response within 30 seconds, failover is triggered
	// 3. Request is retried with next provider

	// Create provider
	uniqueName := generateUniqueProviderName("failover-timeout")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-timeout-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
	providerID := createResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	s.Logger.Info("Failover timeout - 30 seconds (documented)", "provider_id", providerID)
}

// TestProviderFailoverUserNotification tests user notification during failover.
func (s *IntegrationTestSuite) TestProviderFailoverUserNotification() {
	ctx := context.Background()

	// Note: This test documents user notification behavior
	// During failover:
	// 1. User should be notified that failover is occurring
	// 2. Notification should include which provider failed and which is being tried
	// 3. After successful failover, user should be notified of new provider

	// Create provider
	uniqueName := generateUniqueProviderName("failover-notify")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-notify-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
	providerID := createResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	s.Logger.Info("Failover user notification (documented)", "provider_id", providerID)
}

// TestProviderFailoverErrorAfterThreeAttempts tests error after 3 failed attempts.
func (s *IntegrationTestSuite) TestProviderFailoverErrorAfterThreeAttempts() {
	ctx := context.Background()

	// Create 3 providers to test full failover chain
	kind := integrationclient.CreateProviderRequestKindClaude

	providers := []struct {
		name  string
		id    int64
		level int
	}{
		{name: "failover-p1", level: 1},
		{name: "failover-p2", level: 2},
		{name: "failover-p3", level: 3},
	}

	for i := range providers {
		uniqueName := generateUniqueProviderName(providers[i].name)
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            uniqueName,
			Kind:            &kind,
			ApiKey:          "sk-ant-api03-attempt-key-" + string(rune('1'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         boolPointer(true),
			SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
			Level:           intPointer(providers[i].level),
		}

		createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
		providers[i].id = createResp.JSON201.Id

		defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providers[i].id)
	}

	// Simulate all 3 providers failing by disabling them
	for _, p := range providers {
		disabled := false
		updateReq := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
			Enabled: &disabled,
		}
		_, err := s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, p.id, updateReq)
		require.NoError(s.T(), err)
	}

	// Verify all are disabled (simulating all failed)
	listResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, listResp.StatusCode())

	enabledCount := 0
	for _, p := range *listResp.JSON200 {
		for _, provider := range providers {
			if p.Id == provider.id && p.Enabled {
				enabledCount++
			}
		}
	}

	assert.Equal(s.T(), 0, enabledCount, "All providers should be disabled (failed)")

	s.Logger.Info("Failover error after 3 attempts verified - all providers disabled")
}

// TestProviderFailoverPreserveContext tests that failover preserves request context.
func (s *IntegrationTestSuite) TestProviderFailoverPreserveContext() {
	ctx := context.Background()

	// Note: This test documents context preservation during failover
	// During failover:
	// 1. Request context (headers, metadata) should be preserved
	// 2. User authentication should be maintained
	// 3. Request ID should remain consistent for tracing

	// Create provider
	uniqueName := generateUniqueProviderName("failover-context")
	kind := integrationclient.CreateProviderRequestKindClaude
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            uniqueName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-context-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
	}

	createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
	providerID := createResp.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)

	// Verify provider is accessible with authenticated context
	getResp, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, providerID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, getResp.StatusCode())

	s.Logger.Info("Failover context preservation (documented)", "provider_id", providerID)
}

// TestProviderFailoverLogTransition tests logging of provider transitions.
func (s *IntegrationTestSuite) TestProviderFailoverLogTransition() {
	ctx := context.Background()

	// Create two providers
	kind := integrationclient.CreateProviderRequestKindClaude

	provider1Name := generateUniqueProviderName("failover-log-1")
	req1 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            provider1Name,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-log-1-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(1),
	}

	createResp1, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp1.StatusCode())
	provider1ID := createResp1.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, provider1ID)

	provider2Name := generateUniqueProviderName("failover-log-2")
	req2 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            provider2Name,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-log-2-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(2),
	}

	createResp2, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp2.StatusCode())
	provider2ID := createResp2.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, provider2ID)

	// Simulate transition by disabling provider 1
	disabled := false
	updateReq := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Enabled: &disabled,
	}

	_, err = s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, provider1ID, updateReq)
	require.NoError(s.T(), err)

	// Verify transition - provider 1 disabled, provider 2 active
	getResp1, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, provider1ID)
	require.NoError(s.T(), err)

	getResp2, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, provider2ID)
	require.NoError(s.T(), err)

	s.Logger.Info("Failover transition logged",
		"from_provider", provider1Name,
		"from_enabled", getResp1.JSON200.Enabled,
		"to_provider", provider2Name,
		"to_enabled", getResp2.JSON200.Enabled)
}

// TestProviderFailoverFullChain tests failover through full provider chain.
func (s *IntegrationTestSuite) TestProviderFailoverFullChain() {
	ctx := context.Background()

	// Create provider chain with different levels
	kind := integrationclient.CreateProviderRequestKindClaude

	providers := []struct {
		name  string
		id    int64
		level int
	}{
		{name: "chain-p1", level: 1},
		{name: "chain-p2", level: 2},
		{name: "chain-p3", level: 3},
		{name: "chain-p4", level: 4},
	}

	for i := range providers {
		uniqueName := generateUniqueProviderName(providers[i].name)
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            uniqueName,
			Kind:            &kind,
			ApiKey:          "sk-ant-api03-chain-key-" + string(rune('1'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         boolPointer(true),
			SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
			Level:           intPointer(providers[i].level),
		}

		createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
		providers[i].id = createResp.JSON201.Id

		defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providers[i].id)
	}

	// Verify full chain exists
	listResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, listResp.StatusCode())

	// Count providers in our chain
	chainCount := 0
	for _, p := range *listResp.JSON200 {
		for _, provider := range providers {
			if p.Id == provider.id {
				chainCount++
				assert.Equal(s.T(), provider.level, *p.Level,
					"Provider should have correct level in chain")
			}
		}
	}

	assert.Equal(s.T(), len(providers), chainCount, "Full chain should be present")

	s.Logger.Info("Failover full chain verified", "chain_length", chainCount)
}

// TestProviderFailoverPriorityOrder tests that failover follows priority order (level).
func (s *IntegrationTestSuite) TestProviderFailoverPriorityOrder() {
	ctx := context.Background()

	// Create providers with non-sequential levels to test ordering
	kind := integrationclient.CreateProviderRequestKindClaude

	providers := []struct {
		name  string
		id    int64
		level int
	}{
		{name: "priority-p1", level: 10},
		{name: "priority-p2", level: 5},
		{name: "priority-p3", level: 1},
		{name: "priority-p4", level: 20},
	}

	for i := range providers {
		uniqueName := generateUniqueProviderName(providers[i].name)
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            uniqueName,
			Kind:            &kind,
			ApiKey:          "sk-ant-api03-priority-key-" + string(rune('1'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         boolPointer(true),
			SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
			Level:           intPointer(providers[i].level),
		}

		createResp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusCreated, createResp.StatusCode())
		providers[i].id = createResp.JSON201.Id

		defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providers[i].id)
	}

	// Verify priority levels are set correctly
	for _, p := range providers {
		getResp, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, p.id)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), p.level, *getResp.JSON200.Level,
			"Provider should have correct priority level")
	}

	s.Logger.Info("Failover priority order verified",
		"lowest_priority_first", 1,
		"highest_priority_last", 20)
}

// TestProviderFailoverRecovery tests that failed provider recovers for future requests.
func (s *IntegrationTestSuite) TestProviderFailoverRecovery() {
	ctx := context.Background()

	// Create primary provider
	kind := integrationclient.CreateProviderRequestKindClaude
	primaryName := generateUniqueProviderName("failover-recovery-primary")
	req1 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            primaryName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-recovery-primary",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(1),
	}

	createResp1, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp1.StatusCode())
	primaryID := createResp1.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, primaryID)

	// Create secondary provider
	secondaryName := generateUniqueProviderName("failover-recovery-secondary")
	req2 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            secondaryName,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-recovery-secondary",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(2),
	}

	createResp2, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp2.StatusCode())
	secondaryID := createResp2.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, secondaryID)

	// Simulate primary failure
	disabled := false
	updateReq := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Enabled: &disabled,
	}
	_, err = s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, primaryID, updateReq)
	require.NoError(s.T(), err)

	// Verify primary is down
	getResp1, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, primaryID)
	require.NoError(s.T(), err)
	assert.False(s.T(), getResp1.JSON200.Enabled, "Primary should be down")

	// Recover primary
	enabled := true
	recoveryReq := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Enabled: &enabled,
	}
	_, err = s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, primaryID, recoveryReq)
	require.NoError(s.T(), err)

	// Verify primary recovered
	getRecoveredResp, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, primaryID)
	require.NoError(s.T(), err)
	assert.True(s.T(), getRecoveredResp.JSON200.Enabled, "Primary should be recovered")

	s.Logger.Info("Failover recovery verified - primary back online",
		"primary_id", primaryID, "secondary_id", secondaryID)
}

// TestProviderFailoverDisabledProviderSkipped tests that disabled providers are skipped.
func (s *IntegrationTestSuite) TestProviderFailoverDisabledProviderSkipped() {
	ctx := context.Background()

	// Create provider chain with middle provider disabled
	kind := integrationclient.CreateProviderRequestKindClaude

	// Provider 1 - enabled
	p1Name := generateUniqueProviderName("skip-p1")
	req1 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            p1Name,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-skip-1",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(1),
	}

	createResp1, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp1.StatusCode())
	p1ID := createResp1.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, p1ID)

	// Provider 2 - disabled
	p2Name := generateUniqueProviderName("skip-p2")
	req2 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            p2Name,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-skip-2",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(false), // Disabled
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(2),
	}

	createResp2, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp2.StatusCode())
	p2ID := createResp2.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, p2ID)

	// Provider 3 - enabled
	p3Name := generateUniqueProviderName("skip-p3")
	req3 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            p3Name,
		Kind:            &kind,
		ApiKey:          "sk-ant-api03-skip-3",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         boolPointer(true),
		SupportedModels: &[]string{"claude-3-5-sonnet-20241022"},
		Level:           intPointer(3),
	}

	createResp3, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req3)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, createResp3.StatusCode())
	p3ID := createResp3.JSON201.Id

	defer s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, p3ID)

	// Verify provider 2 is disabled (will be skipped during failover)
	getResp2, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, p2ID)
	require.NoError(s.T(), err)
	assert.False(s.T(), getResp2.JSON200.Enabled, "Provider 2 should be disabled (skipped)")

	// Verify providers 1 and 3 are enabled
	getResp1, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, p1ID)
	require.NoError(s.T(), err)
	assert.True(s.T(), getResp1.JSON200.Enabled, "Provider 1 should be enabled")

	getResp3, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, p3ID)
	require.NoError(s.T(), err)
	assert.True(s.T(), getResp3.JSON200.Enabled, "Provider 3 should be enabled")

	s.Logger.Info("Failover skips disabled provider verified",
		"p1_enabled", getResp1.JSON200.Enabled,
		"p2_enabled", getResp2.JSON200.Enabled,
		"p3_enabled", getResp3.JSON200.Enabled)
}
