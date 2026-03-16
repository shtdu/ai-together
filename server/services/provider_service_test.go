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

func TestProviderService_CreateProvider_Success(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	testProvider := &models.Provider{
		ID:              1,
		Name:            "Test Provider",
		APIURL:          "https://api.example.com",
		APIKey:          "key123",
		Kind:            "claude",
		TeamID:          1,
		Enabled:         true,
		ModelMapping:    map[string]string{"gpt-4": "claude-3-opus"},
		SupportedModels: []string{"gpt-4", "gpt-3.5-turbo"},
		Level:           1,
	}

	mockRepo.On("CreateProvider", mock.Anything, "Test Provider", "https://api.example.com", "key123", "claude", int64(1), true, mock.AnythingOfType("map[string]string"), []string{"gpt-4", "gpt-3.5-turbo"}, 1).
		Return(testProvider, nil)

	modelMapping := map[string]interface{}{"gpt-4": "claude-3-opus"}
	provider, err := service.CreateProvider("Test Provider", "https://api.example.com", "key123", "claude", 1, true, modelMapping, []string{"gpt-4", "gpt-3.5-turbo"}, 1)

	require.NoError(t, err)
	assert.Equal(t, "Test Provider", provider.Name)
	assert.Equal(t, "claude", provider.Kind)

	mockRepo.AssertExpectations(t)
}

func TestProviderService_CreateProvider_ModelMappingConversion(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	testProvider := &models.Provider{
		ID:      1,
		Name:    "Test Provider",
		APIURL:  "https://api.example.com",
		APIKey:  "key123",
		Kind:    "claude",
		TeamID:  1,
		Enabled: true,
		Level:   0,
	}

	var capturedMapping map[string]string
	mockRepo.On("CreateProvider", mock.Anything, "Test Provider", "https://api.example.com", "key123", "claude", int64(1), true, mock.AnythingOfType("map[string]string"), []string(nil), 0).
		Run(func(args mock.Arguments) {
			capturedMapping = args.Get(7).(map[string]string)
		}).Return(testProvider, nil)

	// Test with map[string]interface{}
	modelMapping := map[string]interface{}{
		"gpt-4":         "claude-3-opus",
		"gpt-3.5-turbo": 123, // Non-string value
	}
	_, err := service.CreateProvider("Test Provider", "https://api.example.com", "key123", "claude", 1, true, modelMapping, nil, 0)

	require.NoError(t, err)
	// Verify conversion happened
	assert.Equal(t, "claude-3-opus", capturedMapping["gpt-4"])
	assert.Equal(t, "123", capturedMapping["gpt-3.5-turbo"]) // Should be converted to string

	mockRepo.AssertExpectations(t)
}

func TestProviderService_CreateProvider_RepositoryError(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("CreateProvider", mock.Anything, "Test Provider", "https://api.example.com", "key123", "claude", int64(1), true, mock.AnythingOfType("map[string]string"), []string(nil), 0).
		Return(nil, assert.AnError)

	provider, err := service.CreateProvider("Test Provider", "https://api.example.com", "key123", "claude", 1, true, nil, nil, 0)

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to create provider")

	mockRepo.AssertExpectations(t)
}

func TestProviderService_GetProviderByID_Success(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	testProvider := &models.Provider{
		ID:     1,
		Name:   "Test Provider",
		TeamID: 1,
	}

	mockRepo.On("GetProviderByID", mock.Anything, int64(1)).Return(testProvider, nil)

	provider, err := service.GetProviderByID(1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), provider.ID)

	mockRepo.AssertExpectations(t)
}

func TestProviderService_GetProviderByID_NotFound(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("GetProviderByID", mock.Anything, int64(999)).Return(nil, assert.AnError)

	provider, err := service.GetProviderByID(999)

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to find provider by ID")

	mockRepo.AssertExpectations(t)
}

func TestProviderService_GetProvidersByTeamID_Success(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	providers := []models.Provider{
		{ID: 1, Name: "Provider 1", TeamID: 1},
		{ID: 2, Name: "Provider 2", TeamID: 1},
	}

	mockRepo.On("GetProvidersByTeamID", mock.Anything, int64(1)).Return(providers, nil)

	result, err := service.GetProvidersByTeamID(1)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	mockRepo.AssertExpectations(t)
}

func TestProviderService_GetProvidersByTeamID_RepositoryError(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("GetProvidersByTeamID", mock.Anything, int64(1)).Return([]models.Provider{}, assert.AnError)

	providers, err := service.GetProvidersByTeamID(1)

	assert.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "failed to get providers for team")

	mockRepo.AssertExpectations(t)
}

func TestProviderService_UpdateProvider_Success(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("UpdateProvider", mock.Anything, int64(1), "Updated Name", "https://new.example.com", "newkey", "claude", true, mock.AnythingOfType("map[string]string"), []string(nil), 0).
		Return(nil)

	err := service.UpdateProvider(1, "Updated Name", "https://new.example.com", "newkey", "claude", true, nil, nil, 0)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestProviderService_UpdateProvider_ModelMappingConversion(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	var capturedMapping map[string]string
	mockRepo.On("UpdateProvider", mock.Anything, int64(1), "Updated Name", "https://new.example.com", "newkey", "claude", true, mock.AnythingOfType("map[string]string"), []string(nil), 0).
		Run(func(args mock.Arguments) {
			capturedMapping = args.Get(7).(map[string]string)
		}).Return(nil)

	modelMapping := map[string]interface{}{
		"gpt-4": "claude-3-opus",
		"test":  456, // Non-string value
	}
	err := service.UpdateProvider(1, "Updated Name", "https://new.example.com", "newkey", "claude", true, modelMapping, nil, 0)

	require.NoError(t, err)
	// Verify conversion
	assert.Equal(t, "claude-3-opus", capturedMapping["gpt-4"])
	assert.Equal(t, "456", capturedMapping["test"])

	mockRepo.AssertExpectations(t)
}

func TestProviderService_UpdateProvider_RepositoryError(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("UpdateProvider", mock.Anything, int64(1), "Updated Name", "https://new.example.com", "newkey", "claude", true, mock.AnythingOfType("map[string]string"), []string(nil), 0).
		Return(assert.AnError)

	err := service.UpdateProvider(1, "Updated Name", "https://new.example.com", "newkey", "claude", true, nil, nil, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update provider")

	mockRepo.AssertExpectations(t)
}

func TestProviderService_DeleteProvider_Success(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("DeleteProvider", mock.Anything, int64(1)).Return(nil)

	err := service.DeleteProvider(1)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestProviderService_DeleteProvider_RepositoryError(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("DeleteProvider", mock.Anything, int64(1)).Return(assert.AnError)

	err := service.DeleteProvider(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete provider")

	mockRepo.AssertExpectations(t)
}

func TestProviderService_EnableProvider_Success(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("EnableProvider", mock.Anything, int64(1)).Return(nil)

	err := service.EnableProvider(1)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestProviderService_EnableProvider_RepositoryError(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("EnableProvider", mock.Anything, int64(1)).Return(assert.AnError)

	err := service.EnableProvider(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to enable provider")

	mockRepo.AssertExpectations(t)
}

func TestProviderService_DisableProvider_Success(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("DisableProvider", mock.Anything, int64(1)).Return(nil)

	err := service.DisableProvider(1)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestProviderService_DisableProvider_RepositoryError(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("DisableProvider", mock.Anything, int64(1)).Return(assert.AnError)

	err := service.DisableProvider(1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to disable provider")

	mockRepo.AssertExpectations(t)
}

func TestProviderService_UpdateProvider_NilModelMapping(t *testing.T) {
	mockRepo := new(MockProviderRepository)
	service := NewProviderService(mockRepo)

	mockRepo.On("UpdateProvider", mock.Anything, int64(1), "Updated Name", "https://new.example.com", "newkey", "claude", true, mock.AnythingOfType("map[string]string"), []string(nil), 0).
		Return(nil)

	err := service.UpdateProvider(1, "Updated Name", "https://new.example.com", "newkey", "claude", true, nil, nil, 0)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
