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

	integrationclient "github.com/code-together/shared/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPermissionAdminCanUpdateProvider tests that admin can update providers.
func (s *IntegrationTestSuite) TestPermissionAdminCanUpdateProvider() {
	ctx := context.Background()

	// Create a provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Update the provider
	kind := integrationclient.UpdateProviderRequestKind("claude")
	name := "updated-claude"
	apiKey := "new-api-key"
	enabled := false
	req := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Name:    &name,
		Kind:    &kind,
		ApiKey:  &apiKey,
		Enabled: &enabled,
	}

	resp, err := s.Client.PutApiV1ProvidersProviderIdWithResponse(ctx, providerID, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode())
	assert.NotNil(s.T(), resp.JSON200)
	assert.Equal(s.T(), "updated-claude", resp.JSON200.Name)
}

// TestPermissionAdminCanDeleteProvider tests that admin can delete providers.
// NOTE: Server returns 200 (not 204) for successful delete
func (s *IntegrationTestSuite) TestPermissionAdminCanDeleteProvider() {
	ctx := context.Background()

	// Create a provider
	providerID := s.createProviderFixtureFromFixture("claude")

	// Delete the provider
	resp, err := s.Client.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode(), "Server returns 200 for delete")

	// Verify provider is deleted
	getResp, err := s.Client.GetApiV1ProvidersProviderIdWithResponse(ctx, providerID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, getResp.StatusCode())
}

// TestPermissionMemberCanReadProviders tests that member can read providers.
// TODO: SERVER FIX REQUIRED - RBAC not working for member read access
// Server should allow members to read providers, but currently returns 403
func (s *IntegrationTestSuite) TestPermissionMemberCanReadProviders() {
	ctx := context.Background()

	// Create a provider as admin
	s.createProviderFixtureFromFixture("claude")

	// Create member client
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	// Member can list providers
	resp, err := memberClient.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode())
	assert.NotNil(s.T(), resp.JSON200)
	assert.GreaterOrEqual(s.T(), len(*resp.JSON200), 1)
}

// TestPermissionMemberCannotCreateProvider tests that member cannot create providers.
// TODO: SERVER FIX REQUIRED - RBAC not blocking member write operations
// Server should return 403 when member tries to create provider
func (s *IntegrationTestSuite) TestPermissionMemberCannotCreateProvider() {
	ctx := context.Background()

	// Create member client
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	// Try to create provider
	kind := integrationclient.CreateProviderRequestKind("claude")
	enabled := true
	req := integrationclient.PostApiV1ProvidersJSONRequestBody{
		Name:    "member-provider",
		Kind:    &kind,
		ApiKey:  "test-key",
		Enabled: &enabled,
	}

	resp, err := memberClient.PostApiV1ProvidersWithResponse(ctx, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode())
	requireErrorResponse(s, resp.JSON403, "permission")
}

// TestPermissionMemberCannotUpdateProvider tests that member cannot update providers.
// TODO: Server RBAC not properly rejecting member write operations on providers
// TODO: SERVER FIX REQUIRED - RBAC not blocking member write operations
// Server should return 403 when member tries to update provider
func (s *IntegrationTestSuite) TestPermissionMemberCannotUpdateProvider() {
	ctx := context.Background()

	// Create a provider as admin
	providerID := s.createProviderFixtureFromFixture("claude")

	// Create member client
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	// Try to update provider
	kind := integrationclient.UpdateProviderRequestKind("claude")
	name := "updated-by-member"
	enabled := false
	req := integrationclient.PutApiV1ProvidersProviderIdJSONRequestBody{
		Name:    &name,
		Kind:    &kind,
		Enabled: &enabled,
	}

	resp, err := memberClient.PutApiV1ProvidersProviderIdWithResponse(ctx, providerID, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode())
	requireErrorResponse(s, resp.JSON403, "permission")
}

// TestPermissionMemberCannotDeleteProvider tests that member cannot delete providers.
// TODO: SERVER FIX REQUIRED - RBAC not blocking member write operations
// Server should return 403 when member tries to delete provider
func (s *IntegrationTestSuite) TestPermissionMemberCannotDeleteProvider() {
	ctx := context.Background()

	// Create a provider as admin
	providerID := s.createProviderFixtureFromFixture("claude")

	// Create member client
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	// Try to delete provider
	resp, err := memberClient.DeleteApiV1ProvidersProviderIdWithResponse(ctx, providerID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode())
	requireErrorResponse(s, resp.JSON403, "permission")

	// Verify provider still exists by listing all providers
	listResp, err := s.Client.GetApiV1ProvidersWithResponse(ctx)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, listResp.StatusCode())
	assert.NotNil(s.T(), listResp.JSON200)

	// Find the provider in the list
	found := false
	for _, p := range *listResp.JSON200 {
		if p.Id == providerID {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Provider should still exist after failed delete")
}

// TestPermissionMemberCannotAccessLicenseWrite tests that member cannot write to license.
// TODO: SERVER FIX REQUIRED - RBAC not blocking member license write operations
// Server should return 403 when member tries to activate license
func (s *IntegrationTestSuite) TestPermissionMemberCannotAccessLicenseWrite() {
	ctx := context.Background()

	// Create member client
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	// Try to activate license (requires write permission)
	licensePEM, err := LoadLicenseFixture(LicenseCommercial)
	require.NoError(s.T(), err)

	req := integrationclient.PostApiV1LicenseActivateJSONRequestBody{
		LicenseKey: licensePEM,
	}

	resp, err := memberClient.PostApiV1LicenseActivateWithResponse(ctx, req)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusForbidden, resp.StatusCode())
	requireErrorResponse(s, resp.JSON403, "permission")
}

// TestPermissionMemberCanGetLicense tests that member can read license info.
// TODO: SERVER FIX REQUIRED - RBAC not allowing member to read license
// Server should allow members to read license information
func (s *IntegrationTestSuite) TestPermissionMemberCanGetLicense() {
	ctx := context.Background()

	// Activate license as admin
	s.activateLicenseFixture(LicenseCommercial)

	// Create member client
	memberClient := s.createAuthenticatedClient(s.MemberToken)

	// Member can read license
	resp, err := memberClient.GetApiV1LicenseWithResponse(ctx)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode())
	assert.NotNil(s.T(), resp.JSON200)
}
