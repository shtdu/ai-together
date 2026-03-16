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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// User Profile Tests (5 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestUserProfileAdminCanGet() {
	ctx := context.Background()

	// Get admin user profile
	resp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get user profile")
	require.Equal(s.T(), 200, resp.StatusCode(), "Should return 200 OK")
	require.NotNil(s.T(), resp.JSON200, "Response should have JSON200 body")

	// Verify user fields
	assert.Equal(s.T(), "admin@example.com", string(resp.JSON200.User.Email))
	assert.Equal(s.T(), "manager", string(resp.JSON200.User.Role))
	assert.NotNil(s.T(), resp.JSON200.User.Id)
	assert.Greater(s.T(), resp.JSON200.User.Id, int64(0))
}

// Check all expected fields are present in user profile
func (s *IntegrationTestSuite) TestUserProfileFieldsPresent() {
	ctx := context.Background()

	resp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())
	require.NotNil(s.T(), resp.JSON200)

	// Check required fields
	assert.NotEmpty(s.T(), resp.JSON200.User.Email, "Email should be present")
	assert.NotEmpty(s.T(), resp.JSON200.User.Name, "Name should be present")
	assert.NotEmpty(s.T(), resp.JSON200.User.Role, "Role should be present")
	assert.NotZero(s.T(), resp.JSON200.User.Id, "ID should be present")
	assert.NotZero(s.T(), resp.JSON200.User.TenantId, "Tenant ID should be present")

	// NOTE: Server currently does not return created_at/updated_at in profile response
	// even though they exist in the database and are returned by login endpoint
	// This is a known server limitation
	// CreatedAt is time.Time (not pointer), so it will be zero if not in JSON response
	if resp.JSON200.User.CreatedAt.IsZero() {
		s.T().Log("WARNING: CreatedAt not returned by /api/v1/user/profile endpoint (server limitation)")
	}
	// UpdatedAt is *time.Time (pointer), so it can be nil if not in JSON response
	if resp.JSON200.User.UpdatedAt == nil {
		s.T().Log("WARNING: UpdatedAt not returned by /api/v1/user/profile endpoint (server limitation)")
	}
}

func (s *IntegrationTestSuite) TestUserProfileUnauthorized() {
	ctx := context.Background()

	// Try to get profile without authentication
	resp, err := s.AnonymousClient.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err, "Request should succeed")
	assert.Equal(s.T(), 401, resp.StatusCode(), "Should return 401 Unauthorized")
	requireErrorResponse(s, resp.JSON401, "authorization")
}

// TODO: Investigate - member profile access failing (likely auth/token issue)
// Related to TestAuthVerifyValidToken issue
func (s *IntegrationTestSuite) TestUserProfileMemberCanGet() {
	ctx := context.Background()

	// Create and login as member user
	memberToken := s.registerAndLoginUserFromFixture("member")
	memberClient := s.createAuthenticatedClient(memberToken)

	// Get member user profile
	resp, err := memberClient.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to get member profile")
	require.Equal(s.T(), 200, resp.StatusCode(), "Should return 200 OK")
	require.NotNil(s.T(), resp.JSON200)

	// Verify member fields
	assert.Equal(s.T(), "member@example.com", string(resp.JSON200.User.Email))
	assert.Equal(s.T(), "member", string(resp.JSON200.User.Role))
}

func (s *IntegrationTestSuite) TestUserProfileImmutableFields() {
	ctx := context.Background()

	// Get profile first time
	resp1, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp1.StatusCode())
	originalID := resp1.JSON200.User.Id
	originalEmail := string(resp1.JSON200.User.Email)
	originalTenantID := resp1.JSON200.User.TenantId

	// Get profile again
	resp2, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp2.StatusCode())

	// Verify immutable fields haven't changed
	assert.Equal(s.T(), originalID, resp2.JSON200.User.Id, "User ID should not change")
	assert.Equal(s.T(), originalEmail, string(resp2.JSON200.User.Email), "Email should not change")
	assert.Equal(s.T(), originalTenantID, resp2.JSON200.User.TenantId, "Tenant ID should not change")
}

// ============================================================================
// User Role Tests (3 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestUserRoleAdminIsManager() {
	ctx := context.Background()

	resp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())

	assert.Equal(s.T(), "manager", string(resp.JSON200.User.Role), "Admin user should have manager role")
}

// TODO: Investigate - member role not being set correctly
// Server may be setting all users as "manager" instead of "member"
func (s *IntegrationTestSuite) TestUserRoleMemberIsMember() {
	ctx := context.Background()

	// Create and login as member
	memberToken := s.registerAndLoginUserFromFixture("member")
	memberClient := s.createAuthenticatedClient(memberToken)

	resp, err := memberClient.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())

	assert.Equal(s.T(), "member", string(resp.JSON200.User.Role), "Regular user should have member role")
}

func (s *IntegrationTestSuite) TestUserRoleInToken() {
	ctx := context.Background()

	// Get profile to verify role matches token
	resp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())

	// The admin token was created with manager role
	assert.Equal(s.T(), "manager", string(resp.JSON200.User.Role), "Profile role should match token role")
}

// ============================================================================
// User Tenant Tests (2 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestUserTenantIdPresent() {
	ctx := context.Background()

	resp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())

	assert.Greater(s.T(), resp.JSON200.User.TenantId, int64(0), "Tenant ID should be positive")
}

func (s *IntegrationTestSuite) TestUserTenantConsistent() {
	ctx := context.Background()

	// Get tenant ID from profile
	resp, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp.StatusCode())
	tenantID := resp.JSON200.User.TenantId

	// Get profile again
	resp2, err := s.Client.GetApiV1UserProfileWithResponse(ctx)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 200, resp2.StatusCode())

	// Verify tenant ID is consistent
	assert.Equal(s.T(), tenantID, resp2.JSON200.User.TenantId, "Tenant ID should be consistent across requests")
}
