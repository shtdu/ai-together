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

	integration_manager "github.com/code-together/integration_manager"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Team CRUD Tests (9 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestMgrListTeams() {
	ctx := context.Background()

	// List teams
	resp, err := s.ManagerClient.GetApiV1TeamsWithResponse(ctx)
	require.NoError(s.T(), err, "Failed to list teams")
	require.Equal(s.T(), 200, resp.StatusCode(), "List teams should return 200")
	require.NotNil(s.T(), resp.JSON200, "Response should have JSON200 body")

	// Should at least have the default team
	assert.GreaterOrEqual(s.T(), len(*resp.JSON200), 1, "Should have at least one team")
}

func (s *IntegrationTestSuite) TestMgrCreateTeam() {
	ctx := context.Background()

	teamName := generateUniqueProviderName("test-team")
	teamDesc := "Test team description"
	settings := map[string]string{
		"mcp_server": "http://mcp.example.com",
	}

	req := integration_manager.PostApiV1TeamsJSONRequestBody{
		Name:        teamName,
		Description: &teamDesc,
		Settings:    &settings,
	}

	resp, err := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create team")
	require.Equal(s.T(), 201, resp.StatusCode(), "Team creation should return 201")
	require.NotNil(s.T(), resp.JSON201, "Response should have JSON201 body")

	assert.Equal(s.T(), teamName, resp.JSON201.Name)
	assert.NotNil(s.T(), resp.JSON201.Description)
	assert.Equal(s.T(), teamDesc, *resp.JSON201.Description)
	assert.Greater(s.T(), resp.JSON201.Id, int64(0))

	// Cleanup
	defer s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, resp.JSON201.Id)
}

func (s *IntegrationTestSuite) TestMgrCreateTeamMinimal() {
	ctx := context.Background()

	teamName := generateUniqueProviderName("test-team-minimal")
	req := integration_manager.PostApiV1TeamsJSONRequestBody{
		Name: teamName,
	}

	resp, err := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, req)
	require.NoError(s.T(), err, "Failed to create team")
	require.Equal(s.T(), 201, resp.StatusCode(), "Team creation should return 201")
	require.NotNil(s.T(), resp.JSON201)

	assert.Equal(s.T(), teamName, resp.JSON201.Name)

	// Cleanup
	defer s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, resp.JSON201.Id)
}

func (s *IntegrationTestSuite) TestMgrUpdateTeam() {
	ctx := context.Background()

	// First create a team
	teamName := generateUniqueProviderName("test-team-update")
	createReq := integration_manager.PostApiV1TeamsJSONRequestBody{
		Name: teamName,
	}
	createResp, err := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, createReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Update the team
	updatedName := teamName + "-updated"
	updatedDesc := "Updated description"
	settings := map[string]string{"key": "value"}
	updateReq := integration_manager.PutApiV1TeamsTeamIdJSONRequestBody{
		Name:        &updatedName,
		Description: &updatedDesc,
		Settings:    &settings,
	}

	updateResp, err := s.ManagerClient.PutApiV1TeamsTeamIdWithResponse(ctx, createResp.JSON201.Id, updateReq)
	require.NoError(s.T(), err, "Failed to update team")
	require.Equal(s.T(), 200, updateResp.StatusCode(), "Team update should return 200")
	require.NotNil(s.T(), updateResp.JSON200)

	assert.Equal(s.T(), updatedName, updateResp.JSON200.Name)
	assert.NotNil(s.T(), updateResp.JSON200.Description)
	assert.Equal(s.T(), updatedDesc, *updateResp.JSON200.Description)

	// Cleanup
	defer s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, createResp.JSON201.Id)
}

func (s *IntegrationTestSuite) TestMgrDeleteTeam() {
	ctx := context.Background()

	// First create a team
	teamName := generateUniqueProviderName("test-team-delete")
	createReq := integration_manager.PostApiV1TeamsJSONRequestBody{
		Name: teamName,
	}
	createResp, err := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, createReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Delete the team
	deleteResp, err := s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, createResp.JSON201.Id)
	require.NoError(s.T(), err, "Failed to delete team")
	require.Equal(s.T(), 200, deleteResp.StatusCode(), "Team deletion should return 200")
}

func (s *IntegrationTestSuite) TestMgrDeleteTeamNotFound() {
	ctx := context.Background()

	// Try to delete a non-existent team
	deleteResp, err := s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, 99999)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 404, deleteResp.StatusCode(), "Deleting non-existent team should return 404")
}

func (s *IntegrationTestSuite) TestMgrGetTeamSettings() {
	ctx := context.Background()

	// Get settings for the default team (ID 1)
	resp, err := s.ManagerClient.GetApiV1TeamsTeamIdSettingsWithResponse(ctx, 1)
	require.NoError(s.T(), err, "Failed to get team settings")
	require.Equal(s.T(), 200, resp.StatusCode(), "Get team settings should return 200")
	require.NotNil(s.T(), resp.JSON200)

	assert.NotNil(s.T(), resp.JSON200.TeamId)
	assert.Equal(s.T(), int64(1), *resp.JSON200.TeamId)
}

func (s *IntegrationTestSuite) TestMgrUpdateTeamSettings() {
	ctx := context.Background()

	// First create a team
	teamName := generateUniqueProviderName("test-team-settings")
	createReq := integration_manager.PostApiV1TeamsJSONRequestBody{
		Name: teamName,
	}
	createResp, err := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, createReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Update team settings
	settings := map[string]string{
		"mcp_server":  "http://mcp.example.com",
		"setting_key": "setting_value",
	}
	updateResp, err := s.ManagerClient.PutApiV1TeamsTeamIdSettingsWithResponse(ctx, createResp.JSON201.Id, settings)
	require.NoError(s.T(), err, "Failed to update team settings")
	require.Equal(s.T(), 200, updateResp.StatusCode(), "Update team settings should return 200")
	require.NotNil(s.T(), updateResp.JSON200)

	assert.NotNil(s.T(), updateResp.JSON200.TeamId)
	assert.Equal(s.T(), createResp.JSON201.Id, *updateResp.JSON200.TeamId)

	// Cleanup
	defer s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, createResp.JSON201.Id)
}

func (s *IntegrationTestSuite) TestMgrGetTeamNotFound() {
	ctx := context.Background()

	// Try to get a non-existent team
	resp, err := s.ManagerClient.GetApiV1TeamsTeamIdWithResponse(ctx, 99999)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 404, resp.StatusCode(), "Getting non-existent team should return 404")
}

// ============================================================================
// Team Member Tests (3 tests)
// ============================================================================

func (s *IntegrationTestSuite) TestMgrListTeamMembers() {
	ctx := context.Background()

	// List members for the default team (ID 1)
	resp, err := s.ManagerClient.GetApiV1TeamsTeamIdMembersWithResponse(ctx, 1)
	require.NoError(s.T(), err, "Failed to list team members")
	require.Equal(s.T(), 200, resp.StatusCode(), "List team members should return 200")
	require.NotNil(s.T(), resp.JSON200)

	assert.NotNil(s.T(), resp.JSON200.TeamId)
	assert.Equal(s.T(), int64(1), *resp.JSON200.TeamId)
}

func (s *IntegrationTestSuite) TestMgrAddTeamMember() {
	ctx := context.Background()

	// First create a team
	teamName := generateUniqueProviderName("test-team-add-member")
	createReq := integration_manager.PostApiV1TeamsJSONRequestBody{
		Name: teamName,
	}
	createResp, err := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, createReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Add a member (member user already exists from bootstrap)
	memberEmail := openapi_types.Email("member@example.com")
	role := integration_manager.PostApiV1TeamsTeamIdMembersJSONBodyRoleMember
	addReq := integration_manager.PostApiV1TeamsTeamIdMembersJSONRequestBody{
		Email: memberEmail,
		Role:  role,
	}

	addResp, err := s.ManagerClient.PostApiV1TeamsTeamIdMembersWithResponse(ctx, createResp.JSON201.Id, addReq)
	require.NoError(s.T(), err, "Failed to add team member")
	require.Equal(s.T(), 201, addResp.StatusCode(), "Add team member should return 201")
	require.NotNil(s.T(), addResp.JSON201)

	assert.Equal(s.T(), string(memberEmail), string(addResp.JSON201.User.Email))
	// role is PostApiV1TeamsTeamIdMembersJSONBodyRole, but response.Role is *string
	assert.NotNil(s.T(), addResp.JSON201.Role)
	assert.Equal(s.T(), string(role), *addResp.JSON201.Role)

	// Cleanup
	defer s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, createResp.JSON201.Id)
}

func (s *IntegrationTestSuite) TestMgrRemoveTeamMember() {
	ctx := context.Background()

	// First create a team
	teamName := generateUniqueProviderName("test-team-remove-member")
	createReq := integration_manager.PostApiV1TeamsJSONRequestBody{
		Name: teamName,
	}
	createResp, err := s.ManagerClient.PostApiV1TeamsWithResponse(ctx, createReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, createResp.StatusCode())

	// Add a member first
	memberEmail := openapi_types.Email("member@example.com")
	role := integration_manager.PostApiV1TeamsTeamIdMembersJSONBodyRoleMember
	addReq := integration_manager.PostApiV1TeamsTeamIdMembersJSONRequestBody{
		Email: memberEmail,
		Role:  role,
	}
	addResp, err := s.ManagerClient.PostApiV1TeamsTeamIdMembersWithResponse(ctx, createResp.JSON201.Id, addReq)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 201, addResp.StatusCode())

	// Remove the member
	removeResp, err := s.ManagerClient.DeleteApiV1TeamsTeamIdMembersMemberIdWithResponse(ctx, createResp.JSON201.Id, addResp.JSON201.User.Id)
	require.NoError(s.T(), err, "Failed to remove team member")
	require.Equal(s.T(), 200, removeResp.StatusCode(), "Remove team member should return 200")

	// Cleanup
	defer s.ManagerClient.DeleteApiV1TeamsTeamIdWithResponse(ctx, createResp.JSON201.Id)
}

func (s *IntegrationTestSuite) TestMgrRemoveTeamMemberNotFound() {
	ctx := context.Background()

	// Try to remove a non-existent member from the default team
	removeResp, err := s.ManagerClient.DeleteApiV1TeamsTeamIdMembersMemberIdWithResponse(ctx, 1, 99999)
	require.NoError(s.T(), err)
	// The endpoint returns success even if member doesn't exist (idempotent)
	assert.Equal(s.T(), 200, removeResp.StatusCode())
}
