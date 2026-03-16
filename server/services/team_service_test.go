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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"switch-server/models"
)

func TestTeamService_CreateTeam_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)

	// Note: We pass nil for userService since CreateTeam doesn't use it
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	testTeam := &models.Team{
		ID:          1,
		Name:        "Test Team",
		Description: "A test team",
		OwnerID:     1,
		TenantID:    1,
	}

	mockTeamRepo.On("CreateTeam", mock.Anything, "Test Team", "A test team", int64(1), int64(1), map[string]string(nil)).
		Return(testTeam, nil)

	team, err := service.CreateTeam("Test Team", "A test team", 1, 1, nil)

	require.NoError(t, err)
	assert.Equal(t, "Test Team", team.Name)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_CreateTeam_WithSettings(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	testTeam := &models.Team{
		ID:          1,
		Name:        "Test Team",
		Description: "A test team",
		OwnerID:     1,
		TenantID:    1,
		Settings:    map[string]string{"key": "value"},
	}

	settings := map[string]string{"key": "value"}
	mockTeamRepo.On("CreateTeam", mock.Anything, "Test Team", "A test team", int64(1), int64(1), settings).
		Return(testTeam, nil)

	team, err := service.CreateTeam("Test Team", "A test team", 1, 1, settings)

	require.NoError(t, err)
	assert.Equal(t, "Test Team", team.Name)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_CreateTeam_RepositoryError(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("CreateTeam", mock.Anything, "Test Team", "A test team", int64(1), int64(1), mock.Anything).
		Return(nil, assert.AnError)

	team, err := service.CreateTeam("Test Team", "A test team", 1, 1, nil)

	assert.Error(t, err)
	assert.Nil(t, team)
	assert.Contains(t, err.Error(), "failed to create team")

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeamByID_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	testTeam := &models.Team{
		ID:      1,
		Name:    "Test Team",
		OwnerID: 1,
	}

	mockTeamRepo.On("GetTeamByID", mock.Anything, int64(1)).Return(testTeam, nil)

	team, err := service.GetTeamByID(1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), team.ID)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeamByID_NotFound(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("GetTeamByID", mock.Anything, int64(999)).Return(nil, assert.AnError)

	team, err := service.GetTeamByID(999)

	assert.Error(t, err)
	assert.Nil(t, team)
	assert.Contains(t, err.Error(), "failed to find team by ID")

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeamsByUserID_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	teams := []models.Team{
		{ID: 1, Name: "Team 1", OwnerID: 1},
		{ID: 2, Name: "Team 2", OwnerID: 1},
	}

	mockTeamRepo.On("GetTeamsByUserID", mock.Anything, int64(1)).Return(teams, nil)

	result, err := service.GetTeamsByUserID(1)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeamsByUserID_RepositoryError(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("GetTeamsByUserID", mock.Anything, int64(1)).Return([]models.Team{}, assert.AnError)

	teams, err := service.GetTeamsByUserID(1)

	assert.Error(t, err)
	assert.Nil(t, teams)
	assert.Contains(t, err.Error(), "failed to get teams for user")

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_AddTeamMember_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("AddTeamMember", mock.Anything, int64(1), int64(2), "member").Return(nil)

	err := service.AddTeamMember(1, 2, "member")

	assert.NoError(t, err)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_AddTeamMember_RepositoryError(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("AddTeamMember", mock.Anything, int64(1), int64(2), "member").Return(assert.AnError)

	err := service.AddTeamMember(1, 2, "member")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add team member")

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_RemoveTeamMember_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("RemoveTeamMember", mock.Anything, int64(1), int64(2)).Return(nil)

	err := service.RemoveTeamMember(1, 2)

	assert.NoError(t, err)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_RemoveTeamMember_RepositoryError(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("RemoveTeamMember", mock.Anything, int64(1), int64(2)).Return(assert.AnError)

	err := service.RemoveTeamMember(1, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove team member")

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeamMembers_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	members := []models.User{
		{ID: 2, Email: "member1@example.com", Name: "Member 1"},
		{ID: 3, Email: "member2@example.com", Name: "Member 2"},
	}

	mockTeamRepo.On("GetTeamMembers", mock.Anything, int64(1)).Return(members, nil)

	result, err := service.GetTeamMembers(1)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeamMembers_RepositoryError(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("GetTeamMembers", mock.Anything, int64(1)).Return([]models.User{}, assert.AnError)

	members, err := service.GetTeamMembers(1)

	assert.Error(t, err)
	assert.Nil(t, members)
	assert.Contains(t, err.Error(), "failed to get team members")

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("UpdateTeam", mock.Anything, int64(1), "Updated Name", "Updated Description", map[string]string(nil)).
		Return(nil)

	err := service.UpdateTeam(1, "Updated Name", "Updated Description", nil)

	assert.NoError(t, err)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_WithSettings(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	settings := map[string]string{"key": "value"}
	mockTeamRepo.On("UpdateTeam", mock.Anything, int64(1), "", "", settings).
		Return(nil)

	err := service.UpdateTeam(1, "", "", settings)

	assert.NoError(t, err)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_UpdateTeam_RepositoryError(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("UpdateTeam", mock.Anything, int64(1), "Updated Name", "Updated Description", mock.Anything).
		Return(assert.AnError)

	err := service.UpdateTeam(1, "Updated Name", "Updated Description", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update team")

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_DeleteTeam_Success(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("DeleteTeam", mock.Anything, int64(1)).Return(nil)

	err := service.DeleteTeam(1)

	assert.NoError(t, err)

	mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_DeleteTeam_RepositoryError(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewTeamService(mockTeamRepo, mockUserRepo, nil)

	mockTeamRepo.On("DeleteTeam", mock.Anything, int64(1)).Return(assert.AnError)

	err := service.DeleteTeam(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete team")

	mockTeamRepo.AssertExpectations(t)
}
