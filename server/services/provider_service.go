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

type ProviderService struct {
	providerRepo repository.ProviderRepositoryInterface
}

// NewProviderService creates and returns a new instance of ProviderService
func NewProviderService(providerRepo repository.ProviderRepositoryInterface) *ProviderService {
	return &ProviderService{
		providerRepo: providerRepo,
	}
}

// CreateProvider creates a new provider configuration for a team
// It stores the provider details including API credentials, model mappings, and supported models
func (s *ProviderService) CreateProvider(name, apiURL, apiKey, kind string, teamID int64, enabled bool, modelMapping map[string]interface{}, supportedModels []string, level int) (*models.Provider, error) {
	// Convert map[string]interface{} to map[string]string for storage
	convertedModelMapping := make(map[string]string)
	for k, v := range modelMapping {
		if str, ok := v.(string); ok {
			convertedModelMapping[k] = str
		} else {
			convertedModelMapping[k] = fmt.Sprintf("%v", v)
		}
	}

	// Pass []string directly to repository (will be marshaled as JSON array)
	provider, err := s.providerRepo.CreateProvider(context.Background(), name, apiURL, apiKey, kind, teamID, enabled, convertedModelMapping, supportedModels, level)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
}

// GetProviderByID retrieves a specific provider by its unique identifier
// It returns the provider details or an error if the provider doesn't exist
func (s *ProviderService) GetProviderByID(providerID int64) (*models.Provider, error) {
	provider, err := s.providerRepo.GetProviderByID(context.Background(), providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider by ID: %w", err)
	}

	return provider, nil
}

// GetProvidersByTeamID retrieves all providers associated with a specific team
// It returns a slice of providers or an error if the operation fails
func (s *ProviderService) GetProvidersByTeamID(teamID int64) ([]models.Provider, error) {
	providers, err := s.providerRepo.GetProvidersByTeamID(context.Background(), teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers for team: %w", err)
	}

	return providers, nil
}

// UpdateProvider updates specific fields of a provider configuration
// It accepts individual parameters for each field to update
func (s *ProviderService) UpdateProvider(providerID int64, name, apiURL, apiKey, kind string, enabled bool, modelMapping map[string]interface{}, supportedModels []string, level int) error {
	// Convert map[string]interface{} to map[string]string for storage
	convertedModelMapping := make(map[string]string)
	if modelMapping != nil {
		for k, v := range modelMapping {
			if str, ok := v.(string); ok {
				convertedModelMapping[k] = str
			} else {
				convertedModelMapping[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Pass []string directly to repository (will be marshaled as JSON array)
	err := s.providerRepo.UpdateProvider(context.Background(), providerID, name, apiURL, apiKey, kind, enabled, convertedModelMapping, supportedModels, level)
	if err != nil {
		return fmt.Errorf("failed to update provider: %w", err)
	}

	return nil
}

// DeleteProvider removes a provider configuration from the database by its ID
// It returns an error if the operation fails
func (s *ProviderService) DeleteProvider(providerID int64) error {
	err := s.providerRepo.DeleteProvider(context.Background(), providerID)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	return nil
}

// EnableProvider updates a provider's status to enabled
// It returns an error if the operation fails
func (s *ProviderService) EnableProvider(providerID int64) error {
	err := s.providerRepo.EnableProvider(context.Background(), providerID)
	if err != nil {
		return fmt.Errorf("failed to enable provider: %w", err)
	}
	return nil
}

// DisableProvider updates a provider's status to disabled
// It returns an error if the operation fails
func (s *ProviderService) DisableProvider(providerID int64) error {
	err := s.providerRepo.DisableProvider(context.Background(), providerID)
	if err != nil {
		return fmt.Errorf("failed to disable provider: %w", err)
	}
	return nil
}

// CountProvidersByNameAndTeam counts providers with the same name in a team
func (s *ProviderService) CountProvidersByNameAndTeam(ctx context.Context, name string, teamID int64) (int64, error) {
	return s.providerRepo.CountProvidersByNameAndTeam(ctx, name, teamID)
}

// CountProvidersByNameAndTeamExcludingID counts providers with the same name in a team, excluding a specific provider ID
func (s *ProviderService) CountProvidersByNameAndTeamExcludingID(ctx context.Context, name string, teamID, excludeID int64) (int64, error) {
	return s.providerRepo.CountProvidersByNameAndTeamExcludingID(ctx, name, teamID, excludeID)
}
