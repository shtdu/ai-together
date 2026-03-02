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

	"switch-server/models"
	"switch-server/repository"
)

type TeamService struct {
	teamRepo    repository.TeamRepositoryInterface
	userRepo    repository.UserRepositoryInterface
	userService *UserService
}

// NewTeamService creates and returns a new instance of TeamService
// It requires a UserService for handling user-related operations within teams
func NewTeamService(teamRepo repository.TeamRepositoryInterface, userRepo repository.UserRepositoryInterface, userService *UserService) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		userRepo:    userRepo,
		userService: userService,
	}
}

// CreateTeam creates a new team with the specified name, description, owner, tenant, and settings
// It returns the created team object or an error if the operation fails
func (s *TeamService) CreateTeam(name, description string, ownerID, tenantID int64, settings map[string]string) (*models.Team, error) {
	// Insert the team into the database
	team, err := s.teamRepo.CreateTeam(context.Background(), name, description, ownerID, tenantID, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

// GetTeamByID retrieves a specific team by its unique identifier
// It returns the team object or an error if the team doesn't exist
func (s *TeamService) GetTeamByID(teamID int64) (*models.Team, error) {
	team, err := s.teamRepo.GetTeamByID(context.Background(), teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team by ID: %w", err)
	}

	return team, nil
}

// GetTeamsByUserID retrieves all teams that the specified user is a member of
// It returns a slice of teams or an error if the operation fails
func (s *TeamService) GetTeamsByUserID(userID int64) ([]models.Team, error) {
	teams, err := s.teamRepo.GetTeamsByUserID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for user: %w", err)
	}

	return teams, nil
}

// AddTeamMember adds a user to a team with the specified role
// It returns an error if the operation fails
func (s *TeamService) AddTeamMember(teamID, userID int64, role string) error {
	err := s.teamRepo.AddTeamMember(context.Background(), teamID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}

	return nil
}

// RemoveTeamMember removes a user from a team
// It returns an error if the operation fails
func (s *TeamService) RemoveTeamMember(teamID, userID int64) error {
	err := s.teamRepo.RemoveTeamMember(context.Background(), teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	return nil
}

// GetTeamMembers retrieves all members of a specific team
// It returns a slice of users or an error if the operation fails
func (s *TeamService) GetTeamMembers(teamID int64) ([]models.User, error) {
	members, err := s.teamRepo.GetTeamMembers(context.Background(), teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	return members, nil
}

// UpdateTeam updates specific fields of a team (name, description, or settings)
// Only non-empty name/description values and non-nil settings will be updated
func (s *TeamService) UpdateTeam(teamID int64, name, description string, settings map[string]string) error {
	err := s.teamRepo.UpdateTeam(context.Background(), teamID, name, description, settings)
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	return nil
}

// DeleteTeam removes a team from the database by its ID
// It returns an error if the operation fails
func (s *TeamService) DeleteTeam(teamID int64) error {
	err := s.teamRepo.DeleteTeam(context.Background(), teamID)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}
