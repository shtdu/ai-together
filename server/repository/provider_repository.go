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

type ProviderRepository struct {
	db *db.DB
}

func NewProviderRepository(database *db.DB) *ProviderRepository {
	return &ProviderRepository{
		db: database,
	}
}

func decodeModelMapping(raw []byte) map[string]string {
	if len(raw) == 0 {
		return map[string]string{}
	}

	var modelMapping map[string]string
	if err := json.Unmarshal(raw, &modelMapping); err != nil {
		return map[string]string{}
	}
	if modelMapping == nil {
		return map[string]string{}
	}
	return modelMapping
}

func decodeSupportedModels(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}

	var supportedModels []string
	if err := json.Unmarshal(raw, &supportedModels); err != nil {
		return []string{}
	}
	if supportedModels == nil {
		return []string{}
	}
	return supportedModels
}

func (r *ProviderRepository) CreateProvider(ctx context.Context, name, apiURL, apiKey, kind string, teamID int64, enabled bool, modelMapping map[string]string, supportedModels []string, level int) (*models.Provider, error) {
	modelMappingJSON, err := json.Marshal(modelMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model mapping: %w", err)
	}

	supportedModelsJSON, err := json.Marshal(supportedModels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal supported models: %w", err)
	}

	provider, err := r.db.CreateProvider(ctx, db.CreateProviderParams{
		Name:            name,
		ApiUrl:          apiURL,
		ApiKey:          apiKey,
		TeamID:          teamID,
		Enabled:         pgtype.Bool{Bool: enabled, Valid: true},
		Kind:            kind,
		ModelMapping:    modelMappingJSON,
		SupportedModels: supportedModelsJSON,
		Level:           pgtype.Int4{Int32: int32(level), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	enabledBool := false
	if provider.Enabled.Valid {
		enabledBool = provider.Enabled.Bool
	}

	levelInt := 0
	if provider.Level.Valid {
		levelInt = int(provider.Level.Int32)
	}

	model := &models.Provider{
		ID:              provider.ID,
		Name:            provider.Name,
		APIURL:          provider.ApiUrl,
		APIKey:          provider.ApiKey,
		TeamID:          provider.TeamID,
		Kind:            provider.Kind,
		Enabled:         enabledBool,
		ModelMapping:    decodeModelMapping(provider.ModelMapping),
		SupportedModels: decodeSupportedModels(provider.SupportedModels),
		Level:           levelInt,
		CreatedAt:       provider.CreatedAt.Time,
		UpdatedAt:       provider.UpdatedAt.Time,
	}

	return model, nil
}

func (r *ProviderRepository) GetProviderByID(ctx context.Context, providerID int64) (*models.Provider, error) {
	provider, err := r.db.GetProviderByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider by ID: %w", err)
	}

	enabledBool := false
	if provider.Enabled.Valid {
		enabledBool = provider.Enabled.Bool
	}

	levelInt := 0
	if provider.Level.Valid {
		levelInt = int(provider.Level.Int32)
	}

	model := &models.Provider{
		ID:              provider.ID,
		Name:            provider.Name,
		APIURL:          provider.ApiUrl,
		APIKey:          provider.ApiKey,
		TeamID:          provider.TeamID,
		Kind:            provider.Kind,
		Enabled:         enabledBool,
		ModelMapping:    decodeModelMapping(provider.ModelMapping),
		SupportedModels: decodeSupportedModels(provider.SupportedModels),
		Level:           levelInt,
		CreatedAt:       provider.CreatedAt.Time,
		UpdatedAt:       provider.UpdatedAt.Time,
	}

	return model, nil
}

func (r *ProviderRepository) UpdateProvider(ctx context.Context, providerID int64, name, apiURL, apiKey, kind string, enabled bool, modelMapping map[string]string, supportedModels []string, level int) error {
	modelMappingJSON, err := json.Marshal(modelMapping)
	if err != nil {
		return fmt.Errorf("failed to marshal model mapping: %w", err)
	}

	supportedModelsJSON, err := json.Marshal(supportedModels)
	if err != nil {
		return fmt.Errorf("failed to marshal supported models: %w", err)
	}

	err = r.db.UpdateProvider(ctx, db.UpdateProviderParams{
		ID:              providerID,
		Name:            name,
		ApiUrl:          apiURL,
		ApiKey:          apiKey,
		Enabled:         pgtype.Bool{Bool: enabled, Valid: true},
		Kind:            kind,
		ModelMapping:    modelMappingJSON,
		SupportedModels: supportedModelsJSON,
		Level:           pgtype.Int4{Int32: int32(level), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	return nil
}

func (r *ProviderRepository) DeleteProvider(ctx context.Context, providerID int64) error {
	err := r.db.DeleteProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	return nil
}

func (r *ProviderRepository) GetProvidersByTeamID(ctx context.Context, teamID int64) ([]models.Provider, error) {
	providers, err := r.db.GetProvidersByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers by team ID: %w", err)
	}

	var result []models.Provider
	for _, provider := range providers {
		enabledBool := false
		if provider.Enabled.Valid {
			enabledBool = provider.Enabled.Bool
		}

		levelInt := 0
		if provider.Level.Valid {
			levelInt = int(provider.Level.Int32)
		}

		model := models.Provider{
			ID:              provider.ID,
			Name:            provider.Name,
			APIURL:          provider.ApiUrl,
			APIKey:          provider.ApiKey,
			TeamID:          provider.TeamID,
			Kind:            provider.Kind,
			Enabled:         enabledBool,
			ModelMapping:    decodeModelMapping(provider.ModelMapping),
			SupportedModels: decodeSupportedModels(provider.SupportedModels),
			Level:           levelInt,
			CreatedAt:       provider.CreatedAt.Time,
			UpdatedAt:       provider.UpdatedAt.Time,
		}
		result = append(result, model)
	}

	return result, nil
}

func (r *ProviderRepository) EnableProvider(ctx context.Context, providerID int64) error {
	err := r.db.EnableProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to enable provider: %w", err)
	}
	return nil
}

func (r *ProviderRepository) DisableProvider(ctx context.Context, providerID int64) error {
	err := r.db.DisableProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to disable provider: %w", err)
	}
	return nil
}

// CountProvidersByNameAndTeam counts providers with the same name in a team
func (r *ProviderRepository) CountProvidersByNameAndTeam(ctx context.Context, name string, teamID int64) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM providers
		WHERE name = $1 AND team_id = $2`

	var count int64
	err := r.db.Pool().QueryRow(ctx, query, name, teamID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count providers by name: %w", err)
	}

	return count, nil
}

// CountProvidersByNameAndTeamExcludingID counts providers with the same name in a team, excluding a specific provider ID
func (r *ProviderRepository) CountProvidersByNameAndTeamExcludingID(ctx context.Context, name string, teamID, excludeID int64) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM providers
		WHERE name = $1 AND team_id = $2 AND id != $3`

	var count int64
	err := r.db.Pool().QueryRow(ctx, query, name, teamID, excludeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count providers by name excluding ID: %w", err)
	}

	return count, nil
}
