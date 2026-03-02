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


package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"switch-server/internal/db"
	"switch-server/models"
)

type TeamRepository struct {
	db *db.DB
}

func NewTeamRepository(database *db.DB) *TeamRepository {
	return &TeamRepository{
		db: database,
	}
}

func decodeTeamSettings(raw []byte) map[string]string {
	if len(raw) == 0 {
		return map[string]string{}
	}

	var settings map[string]string
	if err := json.Unmarshal(raw, &settings); err != nil {
		return map[string]string{}
	}
	if settings == nil {
		return map[string]string{}
	}
	return settings
}

func (r *TeamRepository) CreateTeam(ctx context.Context, name, description string, ownerID, tenantID int64, settings map[string]string) (*models.Team, error) {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}

	var descriptionPgtype pgtype.Text
	if description != "" {
		descriptionPgtype = pgtype.Text{String: description, Valid: true}
	} else {
		descriptionPgtype = pgtype.Text{Valid: false}
	}

	team, err := r.db.CreateTeam(ctx, db.CreateTeamParams{
		Name:        name,
		Description: descriptionPgtype,
		OwnerID:     ownerID,
		TenantID:    tenantID,
		Settings:    settingsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	descriptionStr := ""
	if team.Description.Valid {
		descriptionStr = team.Description.String
	}

	model := &models.Team{
		ID:          team.ID,
		Name:        team.Name,
		Description: descriptionStr,
		OwnerID:     team.OwnerID,
		TenantID:    team.TenantID,
		Settings:    decodeTeamSettings(team.Settings),
		CreatedAt:   team.CreatedAt.Time,
		UpdatedAt:   team.UpdatedAt.Time,
	}

	return model, nil
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, teamID int64) (*models.Team, error) {
	team, err := r.db.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team by ID: %w", err)
	}

	descriptionStr := ""
	if team.Description.Valid {
		descriptionStr = team.Description.String
	}

	model := &models.Team{
		ID:          team.ID,
		Name:        team.Name,
		Description: descriptionStr,
		OwnerID:     team.OwnerID,
		TenantID:    team.TenantID,
		Settings:    decodeTeamSettings(team.Settings),
		CreatedAt:   team.CreatedAt.Time,
		UpdatedAt:   team.UpdatedAt.Time,
	}

	return model, nil
}

func (r *TeamRepository) GetTeamsByUserID(ctx context.Context, userID int64) ([]models.Team, error) {
	teams, err := r.db.GetTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for user: %w", err)
	}

	result := make([]models.Team, 0)
	for _, team := range teams {
		descriptionStr := ""
		if team.Description.Valid {
			descriptionStr = team.Description.String
		}

		model := models.Team{
			ID:          team.ID,
			Name:        team.Name,
			Description: descriptionStr,
			OwnerID:     team.OwnerID,
			TenantID:    team.TenantID,
			Settings:    decodeTeamSettings(team.Settings),
		}
		result = append(result, model)
	}

	return result, nil
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, teamID int64, name, description string, settings map[string]string) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Only update fields if they are provided (not empty)
	var updateName string
	if name != "" {
		updateName = name
	} else {
		// We need to get the current name first to avoid overwriting with empty values
		currentTeam, err := r.GetTeamByID(ctx, teamID)
		if err != nil {
			return fmt.Errorf("failed to get current team: %w", err)
		}
		updateName = currentTeam.Name
	}

	var descriptionPgtype pgtype.Text
	if description != "" {
		descriptionPgtype = pgtype.Text{String: description, Valid: true}
	} else {
		descriptionPgtype = pgtype.Text{Valid: false}
	}

	err = r.db.UpdateTeam(ctx, db.UpdateTeamParams{
		ID:          teamID,
		Name:        updateName,
		Description: descriptionPgtype,
		Settings:    settingsJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	return nil
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, teamID int64) error {
	err := r.db.DeleteTeam(ctx, teamID)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

func (r *TeamRepository) AddTeamMember(ctx context.Context, teamID, userID int64, role string) error {
	err := r.db.AddTeamMember(ctx, db.AddTeamMemberParams{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	})
	if err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}

	return nil
}

func (r *TeamRepository) RemoveTeamMember(ctx context.Context, teamID, userID int64) error {
	err := r.db.RemoveTeamMember(ctx, db.RemoveTeamMemberParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	return nil
}

func (r *TeamRepository) GetTeamMembers(ctx context.Context, teamID int64) ([]models.User, error) {
	users, err := r.db.GetTeamMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	result := make([]models.User, 0)
	for _, user := range users {
		model := models.User{
			ID:       user.ID,
			Email:    user.Email,
			Name:     user.Name,
			Role:     user.Role,
			TenantID: user.TenantID,
		}
		result = append(result, model)
	}

	return result, nil
}
