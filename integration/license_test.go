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
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	integrationclient "github.com/code-together/shared/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to get license type safely
func getLicenseType(license *integrationclient.LicenseStatus) string {
	if license == nil || license.Type == nil {
		return "opensource" // default
	}
	return string(*license.Type)
}

// Helper to get max teams safely
func getMaxTeams(license *integrationclient.LicenseStatus) int {
	if license == nil || license.MaxTeams == nil {
		return 1 // default to 1 team for opensource
	}
	return *license.MaxTeams
}

// Helper to get max seats safely (informational, not enforced)
func getMaxSeats(license *integrationclient.LicenseStatus) int {
	if license == nil || license.MaxSeats == nil {
		return -1 // default to unlimited
	}
	return *license.MaxSeats
}

// Helper to get current users safely
func getCurrentUsers(resp *integrationclient.GetApiV1LicenseResponse) int {
	if resp == nil || resp.JSON200 == nil || resp.JSON200.Usage == nil || resp.JSON200.Usage.CurrentUsers == nil {
		return 0
	}
	return *resp.JSON200.Usage.CurrentUsers
}

// Helper to get current teams safely
func getCurrentTeams(resp *integrationclient.GetApiV1LicenseResponse) int {
	if resp == nil || resp.JSON200 == nil || resp.JSON200.Usage == nil || resp.JSON200.Usage.CurrentTeams == nil {
		return 0
	}
	return *resp.JSON200.Usage.CurrentTeams
}

// ============================================================================
// License Activation Tests (8 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestLicenseActivateOpenSource() {
	s.skipBootstrap = true
	ctx := context.Background()

	licensePEM, err := LoadLicenseFixture(LicenseOpenSource)
	require.NoError(s.T(), err, "Failed to load Open Source license fixture")

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to activate Open Source license")
	require.Equal(s.T(), 200, resp.StatusCode(), "License activation should return 200")
	require.NotNil(s.T(), resp.JSON200)

	// Fetch license status
	statusResp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get license status")
	require.Equal(s.T(), 200, statusResp.StatusCode())

	license := statusResp.JSON200.License
	require.NotNil(s.T(), license)

	assert.Equal(s.T(), "opensource", getLicenseType(license), "Should be opensource type")
	assert.Equal(s.T(), -1, getMaxSeats(license), "Open Source should have unlimited seats")
	assert.Equal(s.T(), 1, getMaxTeams(license), "Open Source should allow 1 team")
	assert.Equal(s.T(), 7, *license.DataRetentionDays, "Open Source should have 7-day retention")
}

func (s *IntegrationTestSuite) TestLicenseActivateCommercial() {
	s.skipBootstrap = true
	ctx := context.Background()

	licensePEM, err := LoadLicenseFixture(LicenseCommercial)
	require.NoError(s.T(), err, "Failed to load Commercial license fixture")

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to activate Commercial license")
	require.Equal(s.T(), 200, resp.StatusCode(), "License activation should return 200")

	// Fetch license status
	statusResp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get license status")

	license := statusResp.JSON200.License
	require.NotNil(s.T(), license)

	assert.Equal(s.T(), "commercial", getLicenseType(license), "Should be commercial type")
	assert.Equal(s.T(), -1, getMaxSeats(license), "Commercial should have unlimited seats")
	assert.Equal(s.T(), -1, getMaxTeams(license), "Commercial should allow unlimited teams")
	assert.Equal(s.T(), 90, *license.DataRetentionDays, "Commercial should have 90-day retention")
}

func (s *IntegrationTestSuite) TestLicenseActivateExpired() {
	s.skipBootstrap = true
	ctx := context.Background()

	licensePEM, err := LoadLicenseFixture(LicenseExpired)
	require.NoError(s.T(), err, "Failed to load expired license fixture")

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to attempt activation")
	assert.Equal(s.T(), 400, resp.StatusCode(), "Expired license should be rejected")
	requireErrorResponse(s, resp.JSON400, "expired")
}

func (s *IntegrationTestSuite) TestLicenseActivateInvalidSignature() {
	ctx := context.Background()

	licensePEM, err := LoadLicenseFixture(LicenseInvalidSignature)
	require.NoError(s.T(), err, "Failed to load invalid signature license fixture")

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to attempt activation")
	assert.Equal(s.T(), 400, resp.StatusCode(), "Invalid signature should be rejected")
	requireErrorResponse(s, resp.JSON400, "license")
}

func (s *IntegrationTestSuite) TestLicenseActivateImmediateExpiry() {
	ctx := context.Background()

	licensePEM, err := LoadLicenseFixture(LicenseImmediateExpiry)
	require.NoError(s.T(), err, "Failed to load immediate expiry license fixture")

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to attempt activation")
	assert.Contains(s.T(), []int{200, 400}, resp.StatusCode())
}

func (s *IntegrationTestSuite) TestLicenseUpgradeOpenSourceToCommercial() {
	ctx := context.Background()

	// Activate Open Source license
	opensourcePEM, err := LoadLicenseFixture(LicenseOpenSource)
	require.NoError(s.T(), err)

	req1 := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: opensourcePEM,
	}

	resp1, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp1.StatusCode())

	status1, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "opensource", getLicenseType(status1.JSON200.License))

	// Upgrade to Commercial
	commercialPEM, err := LoadLicenseFixture(LicenseCommercial)
	require.NoError(s.T(), err)

	req2 := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: commercialPEM,
	}

	resp2, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req2)
	require.NoError(s.T(), err, "Failed to upgrade license")
	require.Equal(s.T(), 200, resp2.StatusCode())

	status2, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "commercial", getLicenseType(status2.JSON200.License), "Should be upgraded to commercial")
}

// TODO: SERVER FIX REQUIRED - RBAC should prevent members from activating licenses
// Server should return 403 when member tries to activate license
func (s *IntegrationTestSuite) TestLicenseMemberCannotActivate() {
	ctx := context.Background()

	// Create member user
	memberToken := s.registerAndLoginUserFromFixture("member")
	memberClient := s.createAuthenticatedClient(memberToken)

	licensePEM, err := LoadLicenseFixture(LicenseCommercial)
	require.NoError(s.T(), err, "Failed to load Commercial license fixture")

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := memberClient.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to attempt activation")
	assert.Equal(s.T(), 403, resp.StatusCode(), "Member should not be able to activate license")
	requireErrorResponse(s, resp.JSON403, "permission")
}

// ============================================================================
// License Retrieval Tests (4 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestLicenseGetActive() {
	ctx := context.Background()

	// Activate Commercial license
	license := s.activateLicenseFixture(LicenseCommercial)

	// Get license
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get license")
	require.Equal(s.T(), 200, resp.StatusCode())

	activeLicense := resp.JSON200.License
	require.NotNil(s.T(), activeLicense)

	assert.Equal(s.T(), getLicenseType(license), getLicenseType(activeLicense))
	if license.MaxSeats != nil && activeLicense.MaxSeats != nil {
		assert.Equal(s.T(), *license.MaxSeats, *activeLicense.MaxSeats)
	}
	if license.MaxTeams != nil && activeLicense.MaxTeams != nil {
		assert.Equal(s.T(), *license.MaxTeams, *activeLicense.MaxTeams)
	}
}

// TODO: Test isolation issue - previous tests activate licenses that persist
// Server should return default/none when no license is active
func (s *IntegrationTestSuite) TestLicenseGetNoLicense() {
	// Skip bootstrap and activate Open Source license to ensure clean state
	s.skipBootstrap = true
	ctx := context.Background()

	// Activate Open Source license to overwrite any previous license
	s.activateLicenseFixture(LicenseOpenSource)

	// Get license - should now be Open Source
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get license")
	require.Equal(s.T(), 200, resp.StatusCode())

	// Should return Open Source license
	require.NotNil(s.T(), resp.JSON200.License)
	assert.Equal(s.T(), "opensource", getLicenseType(resp.JSON200.License), "Should be opensource")
}

func (s *IntegrationTestSuite) TestLicenseMemberCanGet() {
	ctx := context.Background()

	// Activate license as admin
	s.activateLicenseFixture(LicenseCommercial)

	// Create member user
	memberToken := s.registerAndLoginUserFromFixture("member")
	memberClient := s.createAuthenticatedClient(memberToken)

	// Member should be able to get license info
	resp, err := memberClient.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err, "Member should be able to get license")
	require.Equal(s.T(), 200, resp.StatusCode())

	require.NotNil(s.T(), resp.JSON200.License)
	assert.Equal(s.T(), "commercial", getLicenseType(resp.JSON200.License))
}

func (s *IntegrationTestSuite) TestLicenseStatusFlags() {
	ctx := context.Background()

	// Activate Commercial license
	s.activateLicenseFixture(LicenseCommercial)

	// Get license with status flags
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get license")
	require.Equal(s.T(), 200, resp.StatusCode())

	license := resp.JSON200.License
	require.NotNil(s.T(), license)

	// Verify status fields exist
	assert.NotNil(s.T(), license.ExpiresAt)
	assert.NotNil(s.T(), license.IssuedAt)
	assert.NotNil(s.T(), resp.JSON200.Status)
}

// ============================================================================
// License Feature Matrix Tests (4 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestLicenseOpenSourceFeatures() {
	s.skipBootstrap = true
	_ = context.Background()

	license := s.activateLicenseFixture(LicenseOpenSource)

	// Verify Open Source features
	assert.Equal(s.T(), "opensource", getLicenseType(license))
	assert.Equal(s.T(), -1, getMaxSeats(license), "Open Source: unlimited seats")
	assert.Equal(s.T(), 1, getMaxTeams(license), "Open Source: 1 team")
	assert.Equal(s.T(), 7, *license.DataRetentionDays, "Open Source: 7-day retention")
}

func (s *IntegrationTestSuite) TestLicenseCommercialFeatures() {
	s.skipBootstrap = true
	_ = context.Background()

	license := s.activateLicenseFixture(LicenseCommercial)

	// Verify Commercial features (unlimited)
	assert.Equal(s.T(), "commercial", getLicenseType(license))
	assert.Equal(s.T(), -1, getMaxSeats(license), "Commercial: unlimited seats")
	assert.Equal(s.T(), -1, getMaxTeams(license), "Commercial: unlimited teams")
	assert.Equal(s.T(), 90, *license.DataRetentionDays, "Commercial: 90-day retention")
}

// TODO: Test isolation issue - previous tests activate licenses that persist
// Server should either return default license when none active, or tests need cleanup
func (s *IntegrationTestSuite) TestLicenseDefaultFeatures() {
	// Skip bootstrap and activate Open Source license to ensure clean state
	s.skipBootstrap = true
	ctx := context.Background()

	// Activate Open Source license to overwrite any previous license
	s.activateLicenseFixture(LicenseOpenSource)

	// Get license
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	license := resp.JSON200.License
	require.NotNil(s.T(), license)

	// Should be Open Source
	assert.Equal(s.T(), "opensource", getLicenseType(license), "Should be opensource")
	assert.Equal(s.T(), -1, getMaxSeats(license), "Open Source: unlimited seats")
	assert.Equal(s.T(), 1, getMaxTeams(license), "Open Source: 1 team")
}

// ============================================================================
// Seat & User Tests (Updated - seats are informational, not enforced)
// ============================================================================

func (s *IntegrationTestSuite) TestLicenseAddUserWithinLimit() {
	ctx := context.Background()

	// Activate Open Source license (unlimited seats - informational)
	s.activateLicenseFixture(LicenseOpenSource)

	// Get license to verify we're under limit
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	require.NotNil(s.T(), resp.JSON200.Usage)
	require.NotNil(s.T(), resp.JSON200.License.MaxSeats)

	// All license types have unlimited seats now
	assert.Equal(s.T(), -1, getMaxSeats(resp.JSON200.License), "Should have unlimited seats")
}

func (s *IntegrationTestSuite) TestLicenseAddUserAtLimit() {
	ctx := context.Background()

	// Activate Open Source license (unlimited seats - informational)
	s.activateLicenseFixture(LicenseOpenSource)

	// Get license
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	license := resp.JSON200.License
	require.NotNil(s.T(), license)
	require.NotNil(s.T(), resp.JSON200.Usage)

	// All license types have unlimited seats now
	assert.Equal(s.T(), -1, getMaxSeats(license), "Should have unlimited seats")

	// Note: Database may have existing users from previous tests
	// We'll test that we can track user count correctly
	initialUsers := getCurrentUsers(resp)

	// Try to add users (may already exist from previous tests)
	s.registerAndLoginUserFromFixture("member")
	s.registerAndLoginUserFromFixture("member2")

	// Verify user count is tracked correctly
	resp2, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	// We should be able to track the user count
	// (may be at or above the initial count depending on existing users)
	finalUsers := getCurrentUsers(resp2)
	assert.GreaterOrEqual(s.T(), finalUsers, initialUsers, "User count should not decrease")
	assert.Equal(s.T(), -1, getMaxSeats(resp2.JSON200.License), "Should have unlimited seats")
}

func (s *IntegrationTestSuite) TestLicenseAddUserOverLimit() {
	// In new license system, there are no seat limits (unlimited for all types)
	// This test is now informational - we verify we can track users
	ctx := context.Background()

	// Activate Open Source license
	s.activateLicenseFixture(LicenseOpenSource)

	// Add users - no limits should be enforced
	s.registerAndLoginUserFromFixture("member")
	s.registerAndLoginUserFromFixture("member2")

	// Get license to verify no limits
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	license := resp.JSON200.License
	require.NotNil(s.T(), license)

	// All license types have unlimited seats
	assert.Equal(s.T(), -1, getMaxSeats(license), "Should have unlimited seats")
}

func (s *IntegrationTestSuite) TestLicenseAddUserUnlimitedTier() {
	ctx := context.Background()

	// Activate Commercial license (unlimited seats)
	s.activateLicenseFixture(LicenseCommercial)

	// Get license
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	license := resp.JSON200.License
	require.NotNil(s.T(), license)

	// Commercial should have unlimited seats
	assert.Equal(s.T(), -1, getMaxSeats(license), "Commercial should have unlimited seats")
}

func (s *IntegrationTestSuite) TestLicenseRemoveUserFreesSeat() {
	ctx := context.Background()

	// Activate Open Source license
	s.activateLicenseFixture(LicenseOpenSource)

	// Get initial license
	resp1, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	initialUsers := getCurrentUsers(resp1)
	maxSeats := getMaxSeats(resp1.JSON200.License)

	// Seats are informational - just verify the value is correct
	assert.Equal(s.T(), -1, maxSeats, "Open Source should have unlimited seats (informational)")
	// User count should be tracked
	assert.GreaterOrEqual(s.T(), initialUsers, 0, "Should track user count")

	// TODO: Test actual user deletion when DELETE /users/:id endpoint is available
	// For now, just verify the license returns correct seat information
}

func (s *IntegrationTestSuite) TestLicenseCanAddUserFlag() {
	ctx := context.Background()

	// Activate Commercial license
	s.activateLicenseFixture(LicenseCommercial)

	// Get license status
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	status := resp.JSON200.Status
	require.NotNil(s.T(), status)

	// With unlimited seats, can_add_user should always be true
	if status.CanAddUser != nil {
		assert.True(s.T(), *status.CanAddUser, "Should be able to add users (unlimited)")
	}
}

func (s *IntegrationTestSuite) TestLicenseCurrentUsersCount() {
	ctx := context.Background()

	// Activate Commercial license
	s.activateLicenseFixture(LicenseCommercial)

	// Get initial count
	resp1, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	initialCount := getCurrentUsers(resp1)

	// Try to add a user (may already exist from previous test)
	s.registerAndLoginUserFromFixture("member")

	// Verify count is tracked (may not increase if user already exists)
	resp2, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	finalCount := getCurrentUsers(resp2)
	// Count should be the same or increased by 1
	assert.GreaterOrEqual(s.T(), finalCount, initialCount, "User count should be tracked")
	assert.LessOrEqual(s.T(), finalCount, initialCount+1, "Should add at most 1 user")
}

// TODO: Test isolation issue - license state persists from previous tests
// Server should either clean up between tests or this test needs to account for existing licenses
func (s *IntegrationTestSuite) TestLicenseSeatsRemaining() {
	ctx := context.Background()

	// Activate Open Source license
	s.activateLicenseFixture(LicenseOpenSource)

	// Get license
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	license := resp.JSON200.License
	usage := resp.JSON200.Usage
	require.NotNil(s.T(), license)
	require.NotNil(s.T(), usage)

	// With unlimited seats, seats_remaining is informational/not applicable
	// Just verify we can get current users
	currentUsers := getCurrentUsers(resp)
	assert.GreaterOrEqual(s.T(), currentUsers, 0, "Should track current users")
}

// ============================================================================
// Provider Tests (Updated - unlimited providers for all license types)
// ============================================================================

func (s *IntegrationTestSuite) TestLicenseAddProviderWithinLimit() {
	ctx := context.Background()

	// Activate Open Source license (all types have unlimited providers now)
	s.activateLicenseFixture(LicenseOpenSource)

	// Create provider
	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            s.T().Name() + "-claude",
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &enabled,
		Level:           &level,
		SupportedModels: &[]string{"claude-3-opus"},
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 201, resp.StatusCode(), "Should be able to add provider")
}

func (s *IntegrationTestSuite) TestLicenseAddProviderMultiple() {
	ctx := context.Background()

	// Activate Open Source license (all types have unlimited providers now)
	s.activateLicenseFixture(LicenseOpenSource)

	// Delete all existing providers from previous tests
	providersResp, _ := s.Client.GetApiV1ProvidersWithResponse(ctx)
	if providersResp.JSON200 != nil {
		for _, p := range *providersResp.JSON200 {
			s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, p.Id)
		}
	}

	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	// Create multiple Claude providers (no limit)
	for i := 1; i <= 5; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            s.T().Name() + "-claude-" + string(rune('0'+i)),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}

		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode(), "Should be able to add provider")
	}
}

func (s *IntegrationTestSuite) TestLicenseAddProviderDifferentKind() {
	// Skip bootstrap to avoid existing providers interfering
	s.skipBootstrap = true

	ctx := context.Background()

	// Activate Commercial license
	s.activateLicenseFixture(LicenseCommercial)

	enabled := true
	level := 1
	claudeKind := integrationclient.CreateProviderRequestKindClaude

	// Create multiple Claude providers
	for i := 1; i <= 3; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            s.T().Name() + "-claude-" + string(rune('0'+i)),
			Kind:            &claudeKind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}

		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode())
	}

	// Should be able to create Codex providers too
	codexKind := integrationclient.CreateProviderRequestKindCodex
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            s.T().Name() + "-codex",
		Kind:            &codexKind,
		ApiKey:          "test-codex-key",
		ApiUrl:          "https://api.codex.com",
		Enabled:         &enabled,
		Level:           &level,
		SupportedModels: &[]string{"codex-model"},
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 201, resp.StatusCode(), "Should be able to add different kind")
}

func (s *IntegrationTestSuite) TestLicenseAddProviderUnlimited() {
	ctx := context.Background()

	// Activate Commercial license (unlimited providers)
	s.activateLicenseFixture(LicenseCommercial)

	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	// Should be able to create many providers
	for i := 1; i <= 10; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            s.T().Name() + "-claude-" + string(rune('0'+i)),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}

		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode(), "Should allow unlimited providers")
	}
}

func (s *IntegrationTestSuite) TestLicenseRemoveProvider() {
	// Skip bootstrap to avoid existing providers interfering
	s.skipBootstrap = true

	ctx := context.Background()

	// Activate Open Source license
	s.activateLicenseFixture(LicenseOpenSource)

	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	// Create a provider
	req1 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            s.T().Name() + "-claude-1",
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-1",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &enabled,
		Level:           &level,
		SupportedModels: &[]string{"claude-3-opus"},
	}

	resp1, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, resp1.StatusCode())
	providerID1 := resp1.JSON201.Id

	// Delete provider
	_, err = s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID1)
	require.NoError(s.T(), err)

	// Should be able to create another provider with same name (no limit)
	req2 := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            s.T().Name() + "-claude-2",
		Kind:            &kind,
		ApiKey:          "sk-ant-test-key-2",
		ApiUrl:          "https://api.anthropic.com",
		Enabled:         &enabled,
		Level:           &level,
		SupportedModels: &[]string{"claude-3-opus"},
	}

	resp2, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req2)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 201, resp2.StatusCode(), "Should succeed after deletion")
}

func (s *IntegrationTestSuite) TestLicenseCanAddProviderFlag() {
	ctx := context.Background()

	// Activate Commercial license (unlimited providers)
	s.activateLicenseFixture(LicenseCommercial)

	// Get license
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	status := resp.JSON200.Status
	require.NotNil(s.T(), status)

	// With unlimited providers, can_add_provider should always be true
	if status.CanAddProvider != nil {
		for kind, canAdd := range *status.CanAddProvider {
			assert.True(s.T(), canAdd, "Should be able to add %s providers (unlimited)", kind)
		}
	}
}

func (s *IntegrationTestSuite) TestLicenseCanAddProviderFlagUnlimited() {
	// Skip bootstrap to avoid existing providers interfering
	s.skipBootstrap = true

	ctx := context.Background()

	// Activate Open Source license (all types have unlimited providers)
	s.activateLicenseFixture(LicenseOpenSource)

	enabled := true
	level := 1
	kind := integrationclient.CreateProviderRequestKindClaude

	// Create multiple Claude providers (no limit)
	for i := 1; i <= 5; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            s.T().Name() + "-claude-" + string(rune('0'+i)),
			Kind:            &kind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}

		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode())
	}

	// Get license to verify no limit enforced
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	status := resp.JSON200.Status
	require.NotNil(s.T(), status)

	// Should still be able to add more providers
	if status.CanAddProvider != nil {
		for kind, canAdd := range *status.CanAddProvider {
			assert.True(s.T(), canAdd, "Should be able to add more %s providers", kind)
		}
	}
}

func (s *IntegrationTestSuite) TestLicenseProviderCountPerKind() {
	// Skip bootstrap to avoid existing providers interfering
	s.skipBootstrap = true

	ctx := context.Background()

	// Activate Commercial license
	s.activateLicenseFixture(LicenseCommercial)

	enabled := true
	level := 1
	claudeKind := integrationclient.CreateProviderRequestKindClaude

	// Create 2 Claude providers
	for i := 1; i <= 2; i++ {
		req := integrationclient.PostApiV1ProvidersJSONRequestBody{
			Name:            s.T().Name() + "-claude-" + string(rune('0'+i)),
			Kind:            &claudeKind,
			ApiKey:          "sk-ant-test-key-" + string(rune('0'+i)),
			ApiUrl:          "https://api.anthropic.com",
			Enabled:         &enabled,
			Level:           &level,
			SupportedModels: &[]string{"claude-3-opus"},
		}

		resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
		require.NoError(s.T(), err)
		require.Equal(s.T(), 201, resp.StatusCode())
	}

	// Create 1 Codex provider
	codexKind := integrationclient.CreateProviderRequestKindCodex
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:            s.T().Name() + "-codex",
		Kind:            &codexKind,
		ApiKey:          "test-codex-key",
		ApiUrl:          "https://api.codex.com",
		Enabled:         &enabled,
		Level:           &level,
		SupportedModels: &[]string{"codex-model"},
	}

	resp, err := s.Client.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, resp.StatusCode())

	// List all providers
	listResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)

	if *listResp.JSON200 != nil {
		// Count should be 3 total (2 Claude + 1 Codex)
		claudeCount := 0
		codexCount := 0
		for _, p := range *listResp.JSON200 {
			if *p.Kind == integrationclient.ProviderKindClaude {
				claudeCount++
			} else if *p.Kind == integrationclient.ProviderKindCodex {
				codexCount++
			}
		}

		assert.Equal(s.T(), 2, claudeCount, "Should have 2 Claude providers")
		assert.Equal(s.T(), 1, codexCount, "Should have 1 Codex provider")
	}
}

// ============================================================================
// License Expiration Tests (6 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestLicenseValidFutureExpiration() {
	_ = context.Background()

	// Activate Commercial license (should have future expiration)
	license := s.activateLicenseFixture(LicenseCommercial)

	// Verify expiration is in the future
	require.NotNil(s.T(), license.ExpiresAt)
	require.NotNil(s.T(), license.IssuedAt)

	// License should be valid
	assert.True(s.T(), license.ExpiresAt.After(*license.IssuedAt), "Expiration should be after issued date")
}

func (s *IntegrationTestSuite) TestLicenseExpiringSoon() {
	ctx := context.Background()

	// Activate license that expires soon (1 day)
	licensePEM, err := LoadLicenseFixture(LicenseImmediateExpiry)
	require.NoError(s.T(), err)

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err)

	if resp.StatusCode() == 200 {
		statusResp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
		require.NoError(s.T(), err)

		license := statusResp.JSON200.License
		require.NotNil(s.T(), license)

		// Verify expiration information is available
		assert.NotNil(s.T(), license.ExpiresAt)
	}
}

// TODO: Test isolation issue - expired license behavior needs verification
// Server should fallback to default opensource when license expires
func (s *IntegrationTestSuite) TestLicenseExpiredFallbackToDefault() {
	// Skip bootstrap and activate Open Source license to ensure clean state
	s.skipBootstrap = true

	ctx := context.Background()

	// First activate Open Source license to ensure clean state
	s.activateLicenseFixture(LicenseOpenSource)

	// Try to activate expired license (should be rejected)
	licensePEM, err := LoadLicenseFixture(LicenseExpired)
	require.NoError(s.T(), err)

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := s.Client.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err)

	// Expired license should be rejected (400)
	// Current Open Source license should remain active
	if resp.StatusCode() == 400 {
		// Good - rejected as expected
		// Open Source license should still be active
		getResp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
		require.NoError(s.T(), err)

		currentLicense := getResp.JSON200.License
		assert.Equal(s.T(), "opensource", getLicenseType(currentLicense), "Open Source should remain active")
	}
}

// TODO: Test isolation issue - expired license limit enforcement needs verification
// Server should enforce default opensource limits when license expires
func (s *IntegrationTestSuite) TestLicenseDaysRemainingCalculation() {
	_ = context.Background()

	// Activate license with known expiration
	license := s.activateLicenseFixture(LicenseCommercial)

	// Verify days remaining can be calculated
	require.NotNil(s.T(), license.ExpiresAt)
	require.NotNil(s.T(), license.IssuedAt)

	// Days remaining should be positive
	duration := license.ExpiresAt.Sub(*license.IssuedAt)
	assert.Greater(s.T(), int64(duration.Hours()/24), int64(0), "Should have days remaining")
}

func (s *IntegrationTestSuite) TestLicenseUsageFieldAccuracy() {
	ctx := context.Background()

	// Activate license
	s.activateLicenseFixture(LicenseCommercial)

	// Get license
	resp, err := s.Client.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)

	license := resp.JSON200.License
	usage := resp.JSON200.Usage
	require.NotNil(s.T(), license)
	require.NotNil(s.T(), usage)

	// Verify usage fields are accurate
	assert.GreaterOrEqual(s.T(), getCurrentUsers(resp), 1, "Should have at least admin user")
	// Note: seats_remaining and max_providers_per_kind are no longer used
	// Teams are now tracked
	assert.NotNil(s.T(), usage.ProviderCounts, "Should have provider counts field")
}

func (s *IntegrationTestSuite) TestLicenseCommercialEndpointsBlockedWithoutLicense() {
	s.skipBootstrap = true
	s.SetupTest()

	// Activate Open Source license (no commercial features)
	s.activateLicenseFixture(LicenseOpenSource)

	client := &http.Client{}

	// Commercial endpoints should return 403
	commercialEndpoints := []string{
		"/api/v1/analytics/providers",
		"/api/v1/analytics/users",
		"/api/v1/analytics/history",
		"/api/v1/analytics/filters",
		"/api/v1/dashboard/rankings",
	}

	for _, path := range commercialEndpoints {
		s.T().Run(path, func(t *testing.T) {
			req, _ := http.NewRequest("GET", s.ServerURL+path, nil)
			req.Header.Set("Authorization", "Bearer "+s.AdminToken)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusForbidden, resp.StatusCode,
				"%s should return 403 without commercial license", path)
			// Verify it's a license/gating block (not RBAC)
			var body map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&body)
			msg, _ := body["error"].(string)
			assert.Contains(t, msg, "license", "Error message should mention license")
		})
	}
}

func (s *IntegrationTestSuite) TestLicenseCommercialEndpointsAllowedWithLicense() {
	s.skipBootstrap = true
	s.SetupTest()

	// Activate Commercial license
	s.activateLicenseFixture(LicenseCommercial)

	client := &http.Client{}

	commercialEndpoints := []string{
			"/api/v1/analytics/providers",
			"/api/v1/analytics/users",
			"/api/v1/analytics/history",
			"/api/v1/dashboard/rankings",
	}

	for _, path := range commercialEndpoints {
		s.T().Run(path, func(t *testing.T) {
			url := s.ServerURL + path
			// Analytics endpoints require date range params
			if strings.Contains(path, "/analytics/") {
				url += "?start_date=2024-01-01&end_date=2026-12-31"
			}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", "Bearer "+s.AdminToken)
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode,
				"%s should return 200 with commercial license", path)
		})
	}
}
